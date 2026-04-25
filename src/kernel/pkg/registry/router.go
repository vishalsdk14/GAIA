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

package registry

import (
	"fmt"
	"gaia/kernel/pkg/types"
	"sort"
)

// SelectAgent finds the optimal agent to execute a specific capability.
// It filters out unavailable agents and ranks the remaining agents by their Trust Score.
// This is the core routing logic called by the Dispatcher during Phase 3 of the control loop.
func (r *InMemoryRegistry) SelectAgent(capability string) (*types.AgentRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentIDs, exists := r.capabilities[capability]
	if !exists || len(agentIDs) == 0 {
		return nil, fmt.Errorf("registry: %w: capability '%s' has no registered agents", 
			fmt.Errorf(string(types.ErrorCodeCapabilityNotFound)), capability)
	}

	// 1. Filter eligible agents
	var eligible []*types.AgentRecord
	for _, id := range agentIDs {
		agent, ok := r.agents[id]
		if !ok {
			continue // Should not happen if Deregister maintains consistency
		}

		// Strictly reject agents in terminal or isolated states
		if agent.Status == types.AgentStatusQuarantined ||
			agent.Status == types.AgentStatusBlacklisted ||
			agent.Status == types.AgentStatusDisconnected ||
			agent.Status == types.AgentStatusRejected {
			continue
		}

		eligible = append(eligible, agent)
	}

	if len(eligible) == 0 {
		return nil, fmt.Errorf("registry: %w: all agents for '%s' are currently quarantined or disconnected", 
			fmt.Errorf(string(types.ErrorCodeAgentUnavailable)), capability)
	}

	// 2. Rank by Trust Score
	// Sort descending (highest trust score first)
	sort.Slice(eligible, func(i, j int) bool {
		// If trust scores are identical, prefer Active over Degraded
		if eligible[i].TrustScore == eligible[j].TrustScore {
			if eligible[i].Status == types.AgentStatusActive && eligible[j].Status != types.AgentStatusActive {
				return true
			}
			return false
		}
		return eligible[i].TrustScore > eligible[j].TrustScore
	})

	// 3. Return the highest ranked agent
	return eligible[0], nil
}
