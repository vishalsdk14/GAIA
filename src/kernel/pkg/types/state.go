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
