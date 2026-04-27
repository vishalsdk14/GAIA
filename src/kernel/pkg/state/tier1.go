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

// Package state implements the multi-tiered state management architecture for the GAIA kernel.
package state

import (
	"fmt"
	"gaia/kernel/pkg/types"
	"sync"
	"time"
)

// ActiveStateManager wraps the Tier 1 ActiveState with a RWMutex to prevent race conditions during parallel step execution.
type ActiveStateManager struct {
	mu    sync.RWMutex
	state *types.ActiveState
}

// NewActiveStateManager initializes a new Tier 1 state manager for a given task.
// It pre-allocates memory for the DeltaLog slice to prevent expensive array re-allocations
// during the hot path of step executions. The AccumulatedOutputs map is initialized empty.
func NewActiveStateManager(taskID string) *ActiveStateManager {
	return &ActiveStateManager{
		state: &types.ActiveState{
			TaskID:             taskID,
			AccumulatedOutputs: make(map[string]interface{}),
			DeltaLog:           make([]types.DeltaLogEntry, 0, 64), // Pre-allocate capacity
		},
	}
}

// AppendResult records the output of a completed step.
// This is an O(1) operation that simply appends the result to the DeltaLog.
// By avoiding immediate map insertions and using the Event Sourcing pattern,
// we prevent concurrent write races when multiple parallel steps complete simultaneously.
// The RWMutex ensures that the append operation is thread-safe.
func (m *ActiveStateManager) AppendResult(stepID string, output interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Append directly to the pre-allocated slice
	m.state.DeltaLog = append(m.state.DeltaLog, types.DeltaLogEntry{
		StepID:    stepID,
		Output:    output,
		Timestamp: time.Now().UTC(),
	})

	m.state.Metadata.StepCount++
}

// GetSnapshot is used by the Planner during Phase 6 (Interpolation).
// It collapses the linear DeltaLog into the primary AccumulatedOutputs map, 
// resolving any pending states. It then returns a deep copy of the map to ensure
// that the caller (Planner) cannot mutate the internal kernel state.
// The DeltaLog slice is cleared (length set to 0) while retaining its underlying capacity.
func (m *ActiveStateManager) GetSnapshot() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Collapse the DeltaLog into the main map
	for _, entry := range m.state.DeltaLog {
		// Phase 18: [COLLISION PROTECTION] (BUG-002)
		// If the StepID already exists, preserve the old one by renaming it
		// before storing the new one. This prevents "state amnesia" if the
		// Planner reuses IDs.
		if oldData, exists := m.state.AccumulatedOutputs[entry.StepID]; exists {
			// Find a unique version name
			for v := 1; v < 100; v++ {
				versionedKey := fmt.Sprintf("%s_v%d", entry.StepID, v)
				if _, taken := m.state.AccumulatedOutputs[versionedKey]; !taken {
					m.state.AccumulatedOutputs[versionedKey] = oldData
					break
				}
			}
		}
		m.state.AccumulatedOutputs[entry.StepID] = entry.Output
	}
	
	// Clear the DeltaLog while retaining underlying array capacity
	m.state.DeltaLog = m.state.DeltaLog[:0]

	// Return a deep copy to prevent the caller (e.g. Interpolator) from mutating internal state
	snapshot := make(map[string]interface{}, len(m.state.AccumulatedOutputs)+1)
	for k, v := range m.state.AccumulatedOutputs {
		snapshot[k] = v
	}
	
	// Inject metadata for Phase Awareness (BUG-002)
	snapshot["metadata"] = m.state.Metadata
	
	return snapshot
}

const MAX_TIER1_STEPS = 50

// RequiresTier2Snapshot evaluates the Snapshot Pruning rules.
// It determines if the Tier 1 Active State has grown too large and requires
// eviction/checkpointing to Tier 2 storage to maintain O(1) lookup performance.
// Rule 1: Triggers if step count exceeds the defined maximum.
func (m *ActiveStateManager) RequiresTier2Snapshot() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Rule 1: Step count exceeds the maximum allowed in Tier 1
	if m.state.Metadata.StepCount > MAX_TIER1_STEPS {
		return true
	}
	
	// Rule 2: State size exceeds limit (simplified for this foundation phase)
	// In a full implementation, we'd calculate JSON byte size here.
	return false
}
