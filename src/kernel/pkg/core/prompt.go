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
11. **Unique Step IDs**: Use unique IDs for EVERY step throughout the task lifecycle (e.g., if you finished step 1-3, start your next plan with step 4). Never reuse "step_1" in a later plan as it will overwrite your previous data.
12. **Mapping Source**: IDs for {{step_id.map[N].id}} come ONLY from "get_map" steps. You cannot extract IDs from "navigate" or "type" steps.

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
	stateBytes, err := json.Marshal(state)
	if err != nil {
		return "", fmt.Errorf("core: failed to serialize state: %w", err)
	}

	// Filter down the capabilities to only what the LLM needs (name, description, schemas)
	// to save tokens, though here we serialize the whole structs for simplicity.
	capBytes, err := json.MarshalIndent(capabilities, "", "  ")
	if err != nil {
		return "", fmt.Errorf("core: failed to serialize capabilities: %w", err)
	}

	prompt := fmt.Sprintf(`USER GOAL:
%s

ACTIVE STATE:
%s

PROVIDED CAPABILITIES:
%s

Generate the raw JSON plan now.`, goal, string(stateBytes), string(capBytes))

	return prompt, nil
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
