package types

import "time"

// TaskStatus defines the possible states of a GAIA task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusWaiting   TaskStatus = "waiting"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusAborted   TaskStatus = "aborted"
)

// Task represents a high-level goal being managed by the GAIA kernel.
type Task struct {
	TaskID      string                 `json:"task_id"`
	Goal        string                 `json:"goal"`
	Status      TaskStatus             `json:"status"`
	Plan        []Step                 `json:"plan,omitempty"`
	StepIndex   int                    `json:"step_index"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	FinishedAt *time.Time              `json:"finished_at,omitempty"`
}

// StepStatus defines the possible states of a single unit of work (Step).
type StepStatus string

const (
	StepStatusPending      StepStatus = "pending"
	StepStatusRunning      StepStatus = "running"
	StepStatusPendingAsync StepStatus = "pending_async"
	StepStatusDone         StepStatus = "done"
	StepStatusFailed       StepStatus = "failed"
)

// Step represents an individual unit of work within a task plan.
type Step struct {
	StepID         string                 `json:"step_id"`
	Capability     string                 `json:"capability"`
	Input          interface{}            `json:"input"`
	DependsOn      []string               `json:"depends_on,omitempty"`
	Status         StepStatus             `json:"status"`
	JobID          string                 `json:"job_id,omitempty"`
	AsyncTimeoutMS int                    `json:"async_timeout_ms,omitempty"`
	AssignedAgent  string                 `json:"assigned_agent,omitempty"`
	Output         interface{}            `json:"output,omitempty"`
	OutputSchema   map[string]interface{} `json:"output_schema,omitempty"`
	Error          *Error                 `json:"error,omitempty"`
	RetryCount     int                    `json:"retry_count"`
}
