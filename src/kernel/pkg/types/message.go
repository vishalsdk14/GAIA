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
	EventStepCompleted      EventName = "STEP_COMPLETED"
	EventStepFailed         EventName = "STEP_FAILED"
	EventPlanGenerated      EventName = "PLAN_GENERATED"
	EventPlanRejected       EventName = "PLAN_REJECTED"
	EventReplanTriggered    EventName = "REPLAN_TRIGGERED"
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
