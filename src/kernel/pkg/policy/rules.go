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

package policy

// DefaultPolicies contains a set of canonical CEL rules used for the Foundation Phase.
// Because the engine is dynamic, these are simply raw strings that the kernel
// will compile at runtime. In the future, these will be stored in a database
// and managed via the GUI.
var DefaultPolicies = map[string]string{
	// CostControl ensures that the cumulative cost of the task plus the estimated
	// cost of the current step does not exceed the task's total budget.
	// Variables expected: task.budget, task.accumulated_cost, step.cost_estimate
	"CostControl": `task.accumulated_cost + step.cost_estimate <= task.budget`,

	// SandboxEnforcement verifies that an agent is legally allowed to mutate global state.
	// If the capability declares mutates_state == true, the agent's auth scopes
	// MUST include "state:write". Otherwise, it defaults to true (safe).
	// Variables expected: capability.constraints.mutates_state, agent.auth.scopes
	"SandboxEnforcement": `capability.constraints.mutates_state ? ("state:write" in agent.auth.scopes) : true`,

	// ApprovalGate is a strict environmental policy.
	// If the deployment environment is "production" and the capability requires
	// external network I/O, the policy fails (forcing human approval).
	// Variables expected: capability.constraints.external_io, env
	"ApprovalGate": `(env == "production" && capability.constraints.external_io) ? false : true`,
}
