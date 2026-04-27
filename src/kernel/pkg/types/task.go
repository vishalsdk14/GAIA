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
// This file implements the Task and Step schemas, defining the DAG-based execution structures.
package types

import "time"

// TaskStatus defines the possible states of a GAIA task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusPlanning  TaskStatus = "planning"
	TaskStatusExecuting TaskStatus = "executing"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// Task represents a high-level goal being managed by the GAIA kernel.
type Task struct {
	TaskID      string                 `json:"task_id"`
	Goal        string                 `json:"goal"`
	Status      TaskStatus             `json:"status"`
	Plan        []Step                 `json:"plan,omitempty"`
	HasMore     bool                   `json:"has_more"`
	CurrentStep int                    `json:"current_step"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	
	// Telemetry (BUG-002)
	TotalSteps       int      `json:"total_steps"`
	TokensPrompt     int      `json:"tokens_prompt"`
	TokensCompletion int      `json:"tokens_completion"`
	TotalDurationMS  int64    `json:"total_duration_ms"`
	EstimatedCostUSD float64  `json:"estimated_cost_usd"`
	AgentsInvolved   []string `json:"agents_involved,omitempty"`

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
	StepStatusAwaitingApproval StepStatus = "awaiting_approval"
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
	DurationMS     int64                  `json:"duration_ms"`
}
