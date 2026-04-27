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
	"log/slog"
	"time"
)

// Register processes an incoming AgentManifest, validates it, and binds its capabilities
// into the global routing table.
// If the agent already exists, this acts as a reconnect and resets their trust score.
func (r *InMemoryRegistry) Register(manifest *types.AgentManifest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Initialize the Agent Record
	record := &types.AgentRecord{
		AgentID:         manifest.AgentID,
		Status:          types.AgentStatusActive, // Initial status is Active after handshake
		Manifest:        *manifest,
		TrustScore:      1.0,                     // Start with perfect trust
		RegisteredAt:    time.Now().UTC(),
		LastHealthCheck: time.Now().UTC(),
		RollingMetrics: &types.RollingMetrics{
			SuccessRate:  1.0,
			P95LatencyMS: 0,
			ErrorCounts:  make(map[string]int),
		},
	}

	r.agents[manifest.AgentID] = record

	// 2. Bind Capabilities to the Routing Table
	for _, cap := range manifest.Capabilities {
		// Prevent duplicates in the slice
		exists := false
		for _, id := range r.capabilities[cap.Name] {
			if id == manifest.AgentID {
				exists = true
				break
			}
		}
		if !exists {
			r.capabilities[cap.Name] = append(r.capabilities[cap.Name], manifest.AgentID)
		}
	}

	slog.Info("Agent registered successfully", "agent_id", manifest.AgentID, "capabilities", len(manifest.Capabilities))
	return nil
}

// Deregister removes an agent from the registry and unbinds its capabilities.
// This implements the "Graceful Disconnect" (DRAIN) flow defined in lifecycles.md Section 3.6.
// It ensures no new steps are routed to this agent.
func (r *InMemoryRegistry) Deregister(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[agentID]; !exists {
		return fmt.Errorf("registry: agent %s not found", agentID)
	}

	// Unbind all capabilities associated with this agent
	for capName, agentList := range r.capabilities {
		filtered := make([]string, 0, len(agentList))
		for _, id := range agentList {
			if id != agentID {
				filtered = append(filtered, id)
			}
		}
		r.capabilities[capName] = filtered
	}

	// Remove the agent record
	delete(r.agents, agentID)
	return nil
}

// UpdateHealth adjusts the agent's rolling metrics and triggers lifecycle transitions
// (e.g., active -> degraded -> quarantined) if thresholds are breached.
// It uses an Exponential Moving Average (EMA) for latency to maintain O(1) memory overhead.
func (r *InMemoryRegistry) UpdateHealth(agentID string, success bool, latencyMS int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	record, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("registry: agent %s not found", agentID)
	}

	// 1. Update Success Rate (EMA approximation)
	alpha := 0.1 // Smoothing factor for 10-period window
	successVal := 0.0
	if success {
		successVal = 1.0
	}
	record.RollingMetrics.SuccessRate = (alpha * successVal) + ((1 - alpha) * record.RollingMetrics.SuccessRate)

	// 2. Update Latency (EMA approximation of P95)
	record.RollingMetrics.P95LatencyMS = int((alpha * float64(latencyMS)) + ((1 - alpha) * float64(record.RollingMetrics.P95LatencyMS)))
	record.LastHealthCheck = time.Now().UTC()

	// 3. Recalculate Trust Score: (Success Rate * 0.6) + (Latency Score * 0.3) + (Availability * 0.1)
	// For Foundation phase, Latency Score is a simple inverse normalized to 1000ms SLA, and Availability is assumed 1.0 if not timed out.
	latencyScore := 1.0 - (float64(record.RollingMetrics.P95LatencyMS) / 1000.0)
	if latencyScore < 0 {
		latencyScore = 0
	}
	availability := 1.0 // Assumed 1.0 since health check reached here
	
	record.TrustScore = (record.RollingMetrics.SuccessRate * 0.6) + (latencyScore * 0.3) + (availability * 0.1)

	// 4. Lifecycle Enforcement (Thresholds from failure-handling.md)
	if record.TrustScore < 0.70 && record.Status != types.AgentStatusDegraded {
		// If trust drops below 0.70, degrade priority
		record.Status = types.AgentStatusDegraded
		slog.Warn("Agent health degraded", "agent_id", agentID, "trust_score", record.TrustScore)
	} else if record.TrustScore > 0.85 && record.Status == types.AgentStatusDegraded {
		// If trust recovers above 0.85, restore to active
		record.Status = types.AgentStatusActive
		slog.Info("Agent health restored", "agent_id", agentID, "trust_score", record.TrustScore)
	}

	return nil
}

// Heartbeat refreshes the agent's 'LastHealthCheck' timestamp and maintains its status.
// This is used for Phase 11 availability tracking to prevent agents from being
// marked as 'Zombie' or 'Disconnected' during periods of inactivity.
func (r *InMemoryRegistry) Heartbeat(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	record, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("registry: agent %s not found", agentID)
	}

	record.LastHealthCheck = time.Now().UTC()
	
	// If the agent was degraded due to a timeout but is now heartbeating,
	// we keep it as Active to ensure it stays in the pool.
	if record.Status == types.AgentStatusDegraded && record.TrustScore > 0.85 {
		record.Status = types.AgentStatusActive
	}

	return nil
}
