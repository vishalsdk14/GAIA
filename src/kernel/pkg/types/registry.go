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
// This file defines the schemas used by the Capability Registry and Planner.
package types

import "time"

// AgentStatus defines the lifecycle status of an agent in the registry.
type AgentStatus string

const (
	AgentStatusConnecting   AgentStatus = "connecting"
	AgentStatusActive       AgentStatus = "active"
	AgentStatusDegraded     AgentStatus = "degraded"
	AgentStatusQuarantined  AgentStatus = "quarantined"
	AgentStatusBlacklisted  AgentStatus = "blacklisted"
	AgentStatusDisconnected AgentStatus = "disconnected"
	AgentStatusRejected     AgentStatus = "rejected"
)

// AgentRecord is the internal tracking object for a registered agent.
type AgentRecord struct {
	AgentID          string           `json:"agent_id"`
	Status           AgentStatus      `json:"status"`
	Manifest         AgentManifest    `json:"manifest"`
	TrustScore       float64          `json:"trust_score"`
	RegisteredAt     time.Time        `json:"registered_at"`
	LastHealthCheck  time.Time        `json:"last_health_check"`
	RollingMetrics   *RollingMetrics  `json:"rolling_metrics,omitempty"`
}

// RollingMetrics tracks agent health and performance.
type RollingMetrics struct {
	SuccessRate  float64        `json:"success_rate"`
	P95LatencyMS int            `json:"p95_latency_ms"`
	ErrorCounts  map[string]int `json:"error_counts,omitempty"`
}

// BackoffType defines the retry delay strategy.
type BackoffType string

const (
	BackoffTypeNone        BackoffType = "none"
	BackoffTypeLinear      BackoffType = "linear"
	BackoffTypeExponential BackoffType = "exponential"
)

// RetryPolicy defines the per-step retry configuration.
type RetryPolicy struct {
	MaxAttempts int         `json:"max_attempts"`
	Backoff     BackoffType `json:"backoff"`
	BaseDelayMS int         `json:"base_delay_ms"`
	MaxDelayMS  int         `json:"max_delay_ms"`
}

// PlanStatus defines the lifecycle status of a PlanRecord.
type PlanStatus string

const (
	PlanStatusGenerating PlanStatus = "generating"
	PlanStatusValid      PlanStatus = "valid"
	PlanStatusRejected   PlanStatus = "rejected"
	PlanStatusExecuting  PlanStatus = "executing"
	PlanStatusCompleted  PlanStatus = "completed"
	PlanStatusFailed     PlanStatus = "failed"
	PlanStatusReplanning PlanStatus = "replanning"
)

// UsageMetrics tracks resource consumption for an LLM interaction.
type UsageMetrics struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// PlanRecord is the internal tracking object for a plan segment.
type PlanRecord struct {
	PlanID     string       `json:"plan_id"`
	TaskID     string       `json:"task_id"`
	Status     PlanStatus   `json:"status"`
	Steps      []Step       `json:"steps"`
	HasMore    bool         `json:"has_more"`
	Generation int          `json:"generation"`
	Usage      UsageMetrics `json:"usage,omitempty"` // Captured from LLM response (BUG-002)
	CreatedAt  time.Time    `json:"created_at"`
}
