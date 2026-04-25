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
// This file defines the wire-level message structures (Request, Response, Event) 
// used for agent communication and the internal Event Bus.
package types

import "time"

// RequestMode defines whether the request is synchronous or asynchronous.
type RequestMode string

const (
	RequestModeSync  RequestMode = "sync"
	RequestModeAsync RequestMode = "async"
)

// Request is the message sent from the Kernel to an Agent to trigger a capability.
type Request struct {
	Type       string      `json:"type"` // Must be "REQUEST"
	RequestID  string      `json:"request_id"`
	From       string      `json:"from"`
	TaskID     string      `json:"task_id"`
	StepID     string      `json:"step_id"`
	Capability string      `json:"capability"`
	Input      interface{} `json:"input"`
	Mode       RequestMode `json:"mode"`
	TimeoutMS  int         `json:"timeout_ms,omitempty"`
}

// RequestMetrics tracks execution telemetry for a request.
type RequestMetrics struct {
	DurationMS   int     `json:"duration_ms,omitempty"`
	CostEstimate float64 `json:"cost_estimate,omitempty"`
	TokensUsed   int     `json:"tokens_used,omitempty"`
}

// Response is the standardized output returned by an Agent after processing a Request.
type Response struct {
	RequestID string          `json:"request_id"`
	Success   bool            `json:"success"`
	Output    interface{}     `json:"output,omitempty"`
	Error     *Error          `json:"error,omitempty"`
	JobID     string          `json:"job_id,omitempty"`
	Metrics   *RequestMetrics `json:"metrics,omitempty"`
}

// AsyncCompletion is the callback payload sent by an Agent to signal final completion.
type AsyncCompletion struct {
	Type      string      `json:"type"` // Must be "ASYNC_COMPLETION"
	JobID     string      `json:"job_id"`
	RequestID string      `json:"request_id"`
	Success   bool        `json:"success"`
	Output    interface{} `json:"output,omitempty"`
	Error     *Error      `json:"error,omitempty"`
}

// EventName defines the catalog of system events.
type EventName string

const (
	EventTaskCreated        EventName = "TASK_CREATED"
	EventTaskPlanning       EventName = "TASK_PLANNING"
	EventTaskExecuting      EventName = "TASK_EXECUTING"
	EventTaskCompleted      EventName = "TASK_COMPLETED"
	EventTaskFailed         EventName = "TASK_FAILED"
	EventTaskCancelled      EventName = "TASK_CANCELLED"
	EventStepStarted        EventName = "STEP_STARTED"
	EventStepApprovalRequired EventName = "STEP_APPROVAL_REQUIRED"
	EventStepCompleted      EventName = "STEP_COMPLETED"
	EventStepFailed         EventName = "STEP_FAILED"
	EventPlanGenerated      EventName = "PLAN_GENERATED"
	EventPlanRejected       EventName = "PLAN_REJECTED"
	EventReplanTriggered    EventName = "REPLAN_TRIGGERED"
	EventPlanExecuting      EventName = "PLAN_EXECUTING"
	EventPlanCompleted      EventName = "PLAN_COMPLETED"
	EventPlanFailed         EventName = "PLAN_FAILED"
	EventAgentRegistered    EventName = "AGENT_REGISTERED"
	EventAgentDegraded      EventName = "AGENT_DEGRADED"
	EventAgentQuarantined   EventName = "AGENT_QUARANTINED"
	EventAgentBlacklisted   EventName = "AGENT_BLACKLISTED"
	EventAgentEjected       EventName = "AGENT_EJECTED"
	EventAgentDisconnected  EventName = "AGENT_DISCONNECTED"
)

// Event is an asynchronous event emitted by the Kernel via the Event Bus.
type Event struct {
	Type            string                 `json:"type"` // Must be "EVENT"
	Name            EventName              `json:"name"`
	Payload         map[string]interface{} `json:"payload,omitempty"`
	TaskID          string                 `json:"task_id,omitempty"`
	StepID          string                 `json:"step_id,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
	SequenceNumber  int                    `json:"sequence_number,omitempty"`
	PreviousEventID string                 `json:"previous_event_id,omitempty"`
}
