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

// Package types defines the canonical Go structures for the GAIA Orchestration Kernel.
// This file defines the core data schemas for Tier 1 and Tier 2 state management.
package types

import "time"

// ActiveState represents the Tier 1 hot state injected into the Planner.
type ActiveState struct {
	TaskID             string                 `json:"task_id"`
	AccumulatedOutputs map[string]interface{} `json:"accumulated_outputs"`
	DeltaLog           []DeltaLogEntry        `json:"delta_log"`
	Metadata           ActiveStateMetadata    `json:"metadata"`
}

// DeltaLogEntry represents an append-only log of step completions to prevent write races.
type DeltaLogEntry struct {
	StepID    string      `json:"step_id"`
	Output    interface{} `json:"output"`
	Timestamp time.Time   `json:"timestamp"`
}

// ActiveStateMetadata tracks size and limits for snapshot triggering.
type ActiveStateMetadata struct {
	StateSizeBytes        int `json:"state_size_bytes"`
	StepCount             int `json:"step_count"`
	LastSnapshotGeneration int `json:"last_snapshot_generation"`
}

// Snapshot represents a Tier 2 state checkpoint.
type Snapshot struct {
	Summary        string                 `json:"summary"`
	KeyState       map[string]interface{} `json:"key_state"`
	CheckpointStep int                    `json:"checkpoint_step"`
	CreatedAt      time.Time              `json:"created_at"`
}
