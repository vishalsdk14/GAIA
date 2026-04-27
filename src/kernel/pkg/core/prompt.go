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
	"gaia/kernel/pkg/types"
	"strings"
)

// SystemPrompt defines the strict instructions for the LLM to act as the GAIA Planner.
// It enforces JSON output and the incremental planning rules.
const SystemPrompt = `You are the GAIA Orchestration Kernel Planner.
Your job is to break down a high-level user GOAL into an executable Directed Acyclic Graph (DAG) of steps.
You must output ONLY valid JSON conforming to the requested schema.

WEB NAVIGATION STRATEGY (PANDA/TuriX):
1. INITIAL STEP: For any new web task, your first step MUST be "navigate" to the target URL.
2. URL STICKINESS: Once on the target site, DO NOT navigate again unless you need to go to a completely different domain. Check the "current_url" in the state before deciding to navigate.
3. ELEMENT IDENTIFICATION: Use "get_map" to see interactive elements. It draws numeric labels on the screen. Refer to elements as "{{step_id.map[N].id}}".
4. MANDATORY MAPPING: You MUST call "get_map" again after ANY action that might change the page (e.g., "navigate", "click_id" on a link, or "press_key: Enter"). IDs from a previous map are strictly STALE after navigation.
5. RESILIENCE: If a step fails with "Selector not found" or "ID stale", your NEXT step must be "get_map" to synchronize your state with the current DOM.
6. **DAG DEPENDENCIES (CRITICAL)**: The Kernel executes all steps in a plan IN PARALLEL by default. If Step 2 needs data from Step 1 (e.g., {{step_1.price}}), you **MUST** add "depends_on": ["step_1"] to Step 2. If you forget this, the task will fail immediately.
7. **Selector Warning**: Numeric GAIA IDs (e.g., 57) are NOT CSS IDs. Never use #57. Use the "_id" capabilities (like click_id) with the raw number.
9. **SEARCH DISCOVERY**: To search for a product, first use "find_element" or "get_map" to locate the search bar (look for "search" or "input"). Then use "type_id" and "press_key" (Enter). Do not try to find the product on the home page.
10. **PRICE EXTRACTION**: Prices often contain currency symbols (₹, $, £). When using "get_map", look for the "text" field that contains these symbols or numeric values near the product title. Use "scrape_id" on that specific ID.
11. SEARCH HINT: After typing into a search box, use "press_key" with "Enter" as it is more reliable than finding and clicking a search icon.
11. **UNIQUE STEP IDS (CRITICAL)**: You MUST use unique IDs for EVERY step throughout the task lifecycle. If you already finished steps 1-5, you MUST start your next plan with "step_6". Never reuse "step_1" as it will overwrite your previous results and cause interpolation failures.
12. **SEARCH RECOVERY**: If you just searched and don't see the results, call "get_map" to see the NEW IDs. Do not try to search again with the same query unless the page failed to load.
14. **GREEDY PLANNING**: Do not stop after every step. If you know the next 3-5 steps (e.g., Type -> Enter -> Wait -> Scrape), include them ALL in a single plan to save time and resources.

INTERPOLATION CHEATSHEET:
- To use an ID from a previous "get_map" step (e.g. step_1): {{step_1.map[0].id}}
- To use an ID from a "find_element" step: {{step_1.id}}
- To use text from a "scrape" step: {{step_1.results[0]}}
- DO NOT use {{step_id.output.map}} - the 'output' field is already unwrapped by the kernel.

RULES:
1. Generate between 1 to 3 steps maximum. We use incremental planning.
2. If the goal requires more steps after these, set "has_more" to true.
3. Every step must use exactly one capability from the PROVIDED CAPABILITIES list.
4. You may interpolate data into step inputs using "{{step_id.field}}" or "{{state.field}}".
5. Provide a valid DAG. Steps can list prerequisites in "depends_on".

OUTPUT SCHEMA (JSON):
{
  "steps": [
    {
      "step_id": "string",
      "capability": "string",
      "input": { ... },
      "depends_on": ["string"] 
    }
  ],
  "has_more": boolean
}`

// BuildUserPrompt constructs the highly structured context window for the LLM.
// It bundles the Goal, Active State, and Capability Manifest into a single prompt.
func BuildUserPrompt(goal string, state map[string]interface{}, capabilities []types.Capability) (string, error) {
	// Phase 18: [COMPRESSION] Prune state to save tokens and cost (BUG-002)
	compressedState := compressState(state)
	
	// Track progress for the Planner
	stepCount := 0
	if meta, ok := state["metadata"].(map[string]interface{}); ok {
		if sc, ok := meta["step_count"].(float64); ok {
			stepCount = int(sc)
		}
	}

	stateBytes, err := json.MarshalIndent(compressedState, "", "  ")
	if err != nil {
		return "", fmt.Errorf("core: failed to serialize state: %w", err)
	}

	// Filter down the capabilities to only what the LLM needs
	capBytes, err := json.MarshalIndent(capabilities, "", "  ")
	if err != nil {
		return "", fmt.Errorf("core: failed to serialize capabilities: %w", err)
	}

	prompt := fmt.Sprintf(`USER GOAL:
%s

CURRENT MISSION PROGRESS: %d steps already executed.

ACTIVE STATE:
%s

PROVIDED CAPABILITIES:
%s

Generate the raw JSON plan now.`, goal, stepCount, string(stateBytes), string(capBytes))

	return prompt, nil
}

// compressState recursively prunes large data structures in the state to save tokens.
// It specifically targets historical browser maps and long strings that are no longer
// required for the current planning phase.
func compressState(state map[string]interface{}) map[string]interface{} {
	// pruned initialized as the cleaned output map
	pruned := make(map[string]interface{})
	
	// maxStringLen defines the cutoff for truncating long text fields (e.g., debug logs)
	const maxStringLen = 200
	// maxArrayLen defines the maximum number of items to keep in generic lists
	const maxArrayLen = 100

	// Phase 20: [CONTEXT_GC] Identify the latest step that contains a "map" to preserve it.
	// In web navigation, only the current page's interactive map is critical. 
	// Storing maps from every previous step in history causes exponential token growth.
	latestMapStep := ""
	for k, v := range state {
		// Only look at keys starting with "step_"
		if strings.HasPrefix(k, "step_") {
			if m, ok := v.(map[string]interface{}); ok {
				// Check if this step result contains a browser map
				if _, hasMap := m["map"]; hasMap {
					latestMapStep = k 
				}
			}
		}
	}

	// Iterate through the global state to apply pruning rules
	for k, v := range state {
		// Rule 1: Skip internal binary paths or large media references that the LLM cannot process.
		// These paths are used by the Kernel for local storage but bloat the prompt.
		if strings.Contains(k, "screenshot_path") || strings.Contains(k, "media") {
			continue
		}

		// Rule 2: Prune old heavy data (Context GC).
		// We preserve the structure of previous steps (so the LLM knows they happened)
		// but we strip the "map" and large "results" from all but the most recent step.
		if strings.HasPrefix(k, "step_") && k != latestMapStep {
			if m, ok := v.(map[string]interface{}); ok {
				smallStep := make(map[string]interface{})
				for subK, subV := range m {
					// Drop the map and large results for historical steps
					if subK == "map" || subK == "results" {
						continue
					}
					smallStep[subK] = subV
				}
				pruned[k] = smallStep
				continue
			}
		}

		pruned[k] = compressValue(v, maxStringLen, maxArrayLen)
	}

	return pruned
}

// compressValue is a recursive helper to trim strings, arrays, and maps.
func compressValue(v interface{}, maxStringLen, maxArrayLen int) interface{} {
	switch val := v.(type) {
	case string:
		if len(val) > maxStringLen {
			// Optimization: Never prune strings containing currency or keywords
			if strings.Contains(val, "₹") || strings.Contains(strings.ToLower(val), "coca") {
				return val
			}
			return val[:maxStringLen] + "...[TRUNCATED]"
		}
		return val
	case []interface{}:
		if len(val) > maxArrayLen {
			newArr := make([]interface{}, 0, maxArrayLen)
			// Priority 1: Items with keywords
			for _, item := range val {
				if m, ok := item.(map[string]interface{}); ok {
					txt, _ := m["text"].(string)
					sLower := strings.ToLower(txt)
					if strings.Contains(txt, "₹") || strings.Contains(sLower, "coca") || strings.Contains(sLower, "add") {
						newArr = append(newArr, item)
					}
				}
				if len(newArr) >= maxArrayLen {
					break
				}
			}
			// Priority 2: Fill remaining slots with the first items
			for _, item := range val {
				if len(newArr) >= maxArrayLen {
					break
				}
				// Avoid duplicates if already added by priority
				alreadyAdded := false
				for _, added := range newArr {
					if added == item {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					newArr = append(newArr, item)
				}
			}
			return newArr
		}
		// Recursively compress elements in the array
		newArr := make([]interface{}, len(val))
		for i, item := range val {
			newArr[i] = compressValue(item, maxStringLen, maxArrayLen)
		}
		return newArr
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for mk, mv := range val {
			// Rule 3: Strip non-essential visual metadata from elements (X/Y/W/H)
			if mk == "x" || mk == "y" || mk == "width" || mk == "height" {
				continue
			}
			newMap[mk] = compressValue(mv, maxStringLen, maxArrayLen)
		}
		return newMap
	default:
		return v
	}
}

// BuildCorrectionPrompt is used for Phase 2.3 (Planner Failure Recovery).
// It instructs the LLM to fix a specific schema violation or malformed JSON.
func BuildCorrectionPrompt(failedResponse string, errorDetail string) string {
	return fmt.Sprintf(`Your previous response was invalid.
ERROR: %s

FAILED RESPONSE:
%s

Please correct the error and return ONLY valid JSON matching the requested schema. Do not include extra text.`, errorDetail, failedResponse)
}
