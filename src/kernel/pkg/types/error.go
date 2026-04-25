package types

// ErrorCode defines the machine-readable error classification.
type ErrorCode string

const (
	ErrorCodeSchemaViolation     ErrorCode = "SCHEMA_VIOLATION"
	ErrorCodeTimeout             ErrorCode = "TIMEOUT"
	ErrorCodePolicyDenied        ErrorCode = "POLICY_DENIED"
	ErrorCodeCapabilityNotFound ErrorCode = "CAPABILITY_NOT_FOUND"
	ErrorCodeAgentUnavailable    ErrorCode = "AGENT_UNAVAILABLE"
	ErrorCodeExecutionFailed     ErrorCode = "EXECUTION_FAILED"
	ErrorCodeInternalError       ErrorCode = "INTERNAL_ERROR"
	ErrorCodeUnknown             ErrorCode = "UNKNOWN"
)

// Error represents a standardized error object in the GAIA kernel.
type Error struct {
	Code      ErrorCode              `json:"code"`
	Message   string                 `json:"message"`
	Retryable bool                   `json:"retryable"`
	Details   map[string]interface{} `json:"details,omitempty"`
}
