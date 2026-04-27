// Copyright 2026 GAIA Contributors
//
// Licensed under the MIT License.
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://opensource.org/licenses/MIT
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gaia/kernel/pkg/types"
)

// AddImplicitDependencies scans each step for interpolation tags (e.g., {{step_1...}})
// and automatically adds the referenced step IDs to the 'DependsOn' slice if they
// are part of the current plan. This prevents race conditions caused by LLM forgetfulness.
func AddImplicitDependencies(plan []types.Step) {
	currentPlanIDs := make(map[string]bool)
	for _, s := range plan {
		currentPlanIDs[s.StepID] = true
	}

	for i := range plan {
		step := &plan[i]
		
		// Marshal input to scan for tags in the raw JSON
		bytes, _ := json.Marshal(step.Input)
		inputStr := string(bytes)
		
		// Use a permissive regex to catch all potential tags
		matches := interpolationRegex.FindAllStringSubmatch(inputStr, -1)
		for _, match := range matches {
			if len(match) < 2 { continue }
			
			content := strings.TrimSpace(match[1])
			
			// Extract root ID: split by any non-alphanumeric/underscore char
			var referencedID string
			firstSeparator := strings.IndexFunc(content, func(r rune) bool {
				return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-')
			})
			
			if firstSeparator == -1 {
				referencedID = content
			} else {
				referencedID = content[:firstSeparator]
			}
			
			// If this ID is in the current plan, it must be a dependency
			if currentPlanIDs[referencedID] && referencedID != step.StepID {
				alreadyPresent := false
				for _, d := range step.DependsOn {
					if d == referencedID {
						alreadyPresent = true
						break
					}
				}
				if !alreadyPresent {
					fmt.Printf("[KERNEL] DAG: Adding implicit dependency: %s -> %s\n", step.StepID, referencedID)
					step.DependsOn = append(step.DependsOn, referencedID)
				}
			}
		}
	}
}

// GetReadySteps implements Phase 3: DAG Resolution.
// It scans the plan and returns all steps that are pending and whose dependencies are fulfilled.
// A dependency is fulfilled if it is either 'done' in the current plan OR exists in the historical state.
func GetReadySteps(plan []types.Step, history map[string]interface{}) []*types.Step {
	// 1. Build a map of step statuses for O(1) dependency checking
	statusMap := make(map[string]types.StepStatus)
	for _, step := range plan {
		statusMap[step.StepID] = step.Status
	}

	var readySteps []*types.Step

	// 2. Scan for ready steps
	for i := range plan {
		step := &plan[i]
		if step.Status != types.StepStatusPending {
			// We only consider Pending steps.
			// If it's already running, done, or failed, skip it.
			continue
		}

		// Check if all dependencies are "done" or in history
		allDepsMet := true
		for _, depID := range step.DependsOn {
			_, inHistory := history[depID]
			if statusMap[depID] != types.StepStatusDone && !inHistory {
				allDepsMet = false
				break
			}
		}

		if allDepsMet {
			readySteps = append(readySteps, step)
		}
	}

	return readySteps
}

// interpolationRegex matches {{state.field}} or {{step_id.output.field}}
var interpolationRegex = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// ResolveInterpolation implements Phase 4: Data Binding.
func ResolveInterpolation(input interface{}, hotState map[string]interface{}) (interface{}, error) {
	bytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("scheduler: failed to marshal input for interpolation: %w", err)
	}
	
	strInput := string(bytes)
	
	strInput = interpolationRegex.ReplaceAllStringFunc(strInput, func(match string) string {
		keyPath := strings.TrimSpace(match[2 : len(match)-2])
		
		// Normalize [0] to .0
		keyPath = strings.ReplaceAll(keyPath, "[", ".")
		keyPath = strings.ReplaceAll(keyPath, "]", "")
		
		parts := strings.Split(keyPath, ".")
		
		var current interface{} = hotState
		found := false
		
		for i, part := range parts {
			if part == "" { continue }
			
			// Special handling for the common LLM mistake: adding ".output."
			if i == 1 && part == "output" {
				if m, ok := current.(map[string]interface{}); ok {
					if _, exists := m["output"]; !exists {
						continue // Skip "output" layer
					}
				}
			}

			if m, ok := current.(map[string]interface{}); ok {
				if val, exists := m[part]; exists {
					current = val
					found = true
					continue
				}
			}
			
			if a, ok := current.([]interface{}); ok {
				if idx, err := strconv.Atoi(part); err == nil {
					if idx >= 0 && idx < len(a) {
						current = a[idx]
						found = true
						continue
					}
				}
			}

			found = false
			break
		}
		
		if found {
			valBytes, _ := json.Marshal(current)
			if s, ok := current.(string); ok {
				// Escape for JSON but strip the marshaled quotes since the tag is inside quotes
				if len(valBytes) >= 2 {
					return string(valBytes[1 : len(valBytes)-1])
				}
				return s
			}
			return string(valBytes)
		}
		
		fmt.Printf("WARNING: Scheduler could not resolve interpolation key: %s\n", keyPath)
		return match
	})
	
	var resolvedInput interface{}
	if err := json.Unmarshal([]byte(strInput), &resolvedInput); err != nil {
		return nil, fmt.Errorf("scheduler: failed to unmarshal interpolated input: %w", err)
	}
	
	return resolvedInput, nil
}
