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
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"gaia/kernel/pkg/types"
)

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
			if depID == "" {
				continue
			}
			// Phase 15: [ROOT CAUSE FIX] Isolation of Current Plan vs History
			// If the dependency ID exists in the current plan, we MUST wait for it to be 'Done'.
			// We only fall back to checking history if the dependency is NOT in the current plan.
			if status, inCurrentPlan := statusMap[depID]; inCurrentPlan {
				if status != types.StepStatusDone {
					allDepsMet = false
					break
				}
			} else {
				// Dependency is not in the current plan, so it must exist in history.
				if _, inHistory := history[depID]; !inHistory {
					allDepsMet = false
					break
				}
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
		
		slog.Warn("Scheduler could not resolve interpolation key", "key", keyPath)
		return match
	})
	
	var resolvedInput interface{}
	if err := json.Unmarshal([]byte(strInput), &resolvedInput); err != nil {
		return nil, fmt.Errorf("scheduler: failed to unmarshal interpolated input: %w", err)
	}
	
	return resolvedInput, nil
}
