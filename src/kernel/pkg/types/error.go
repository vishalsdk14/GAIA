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
// This file standardizes Error classifications and messages across the system.
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
