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
You must output ONLY valid JSON conforming to the requested schema. Do not include markdown formatting like "X json" blocks. Just the raw JSON object.

RULES:
1. Generate between 1 to 3 steps maximum. We use incremental planning to avoid context drift.
2. If the goal requires more steps after these, set "has_more" to true.
3. Every step must use exactly one capability from the PROVIDED CAPABILITIES list. Do not hallucinate capabilities.
4. You may interpolate data into step inputs using "{{step_id.output.field}}" or "{{state.field}}".
5. Provide a valid DAG. Steps can list prerequisites in "depends_on".

OUTPUT SCHEMA (JSON):
{
  "steps": [
    {
      "step_id": "string (unique identifier)",
      "capability": "string (must match a provided capability name)",
      "input": { ... JSON object matching capability input_schema ... },
      "depends_on": ["string"] 
    }
  ],
  "has_more": boolean
}`

// BuildUserPrompt constructs the highly structured context window for the LLM.
// It bundles the Goal, Active State, and Capability Manifest into a single prompt.
func BuildUserPrompt(goal string, state map[string]interface{}, capabilities []types.Capability) (string, error) {
	stateBytes, err := json.MarshalIndent(state, "", "  ")
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
