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
	"strings"

	"gaia/kernel/pkg/types"
)

// GetReadySteps implements Phase 3: DAG Resolution.
// It scans the plan and returns all steps that are pending and whose dependencies are fulfilled.
func GetReadySteps(plan []types.Step) []*types.Step {
	// 1. Build a map of step statuses for O(1) dependency checking
	statusMap := make(map[string]types.StepStatus)
	for _, step := range plan {
		statusMap[step.StepID] = step.Status
	}

	var readySteps []*types.Step

	// 2. Scan for ready steps
	for i := range plan {
		step := &plan[i]
		if step.Status != types.StepStatusPending && step.Status != types.StepStatusFailed {
			// We only consider Pending steps (or Failed steps that have been reset to Pending during retry).
			// If it's already running or done, skip it.
			continue
		}

		// Check if all dependencies are "done"
		allDepsMet := true
		for _, depID := range step.DependsOn {
			if statusMap[depID] != types.StepStatusDone {
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
// It recursively resolves variables injected by the planner into the step input.
func ResolveInterpolation(input interface{}, hotState map[string]interface{}) (interface{}, error) {
	// The most robust way to handle arbitrary JSON interpolation in Go without complex
	// recursive reflection is to marshal to JSON, do string replacement, and unmarshal.
	
	bytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("scheduler: failed to marshal input for interpolation: %w", err)
	}
	
	strInput := string(bytes)
	
	// Perform regex replacement
	strInput = interpolationRegex.ReplaceAllStringFunc(strInput, func(match string) string {
		// match is like "{{state.user_email}}"
		key := strings.TrimSpace(match[2 : len(match)-2])
		
		// Attempt to resolve the key from the hot state
		// Note: A full implementation would use a JSONPath library to traverse nested state objects.
		// For the foundation phase, we do a flat map lookup.
		if val, exists := hotState[key]; exists {
			// Serialize the resolved value so it safely injects back into the JSON string
			valBytes, _ := json.Marshal(val)
			// Strip the surrounding quotes if the original JSON expected a raw string interpolation inside quotes?
			// Actually, if they wrote `"email": "{{state.email}}"`, then json.Marshal adds quotes: `"foo@bar"`.
			// So it becomes `"email": ""foo@bar""` which is invalid JSON.
			// Let's assume the planner writes `"email": "{{state.email}}"`.
			// If val is string, we just inject the raw string because it's already wrapped in quotes in the template.
			if strVal, ok := val.(string); ok {
				return strVal
			}
			return string(valBytes)
		}
		
		// If unresolvable, leave it alone (it will fail later in Phase 5/6)
		return match
	})
	
	var resolvedInput interface{}
	if err := json.Unmarshal([]byte(strInput), &resolvedInput); err != nil {
		return nil, fmt.Errorf("scheduler: failed to unmarshal interpolated input: %w", err)
	}
	
	return resolvedInput, nil
}
