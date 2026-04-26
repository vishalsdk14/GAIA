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

// Package core implements the central orchestration logic of the GAIA Kernel.
// This includes the 10-phase control loop, task coordination, and planner integration.
package core

import (
	"os"
)

// LLMProvider defines the supported types of LLM backends.
type LLMProvider string

const (
	// LLMProviderLocal refers to a locally hosted LLM (e.g., via Ollama or Llama.cpp).
	LLMProviderLocal LLMProvider = "local"
	// LLMProviderCloud refers to a cloud-based LLM provider (e.g., OpenAI, Anthropic).
	LLMProviderCloud LLMProvider = "cloud"
)

// KernelConfig houses the dynamic configuration parameters for the GAIA Kernel.
// These values are typically passed from the GUI or CLI on startup and can be
// updated at runtime to adjust the kernel's behavior without code changes.
type KernelConfig struct {
	// PlannerProvider specifies whether to use a local or cloud LLM for planning.
	PlannerProvider LLMProvider `json:"planner_provider"`
	
	// PlannerEndpoint is the URL for the LLM API.
	PlannerEndpoint string `json:"planner_endpoint"`
	
	// PlannerModel is the specific model to use.
	PlannerModel string `json:"planner_model"`
	
	// PlannerAPIKey is the secret used for authenticating with cloud LLM providers.
	PlannerAPIKey string `json:"planner_api_key,omitempty"`
	
	// MaxReplans defines the circuit breaker limit for how many times a task can be replanned.
	MaxReplans int `json:"max_replans"`
	
	// MaxConcurrentTasks defines the maximum number of parallel tasks the kernel will process.
	MaxConcurrentTasks int `json:"max_concurrent_tasks"`

	// MaxConcurrentPerAgent defines the maximum number of parallel steps a single agent can handle.
	MaxConcurrentPerAgent int `json:"max_concurrent_per_agent"`

	// DBPath is the location of the persistent SQLite database for the kernel.
	DBPath string `json:"db_path"`

	// LogLevel defines the verbosity of the kernel logs (e.g., 'DEBUG', 'INFO', 'ERROR').
	LogLevel string `json:"log_level"`

	// Environment specifies the deployment context (e.g., 'development', 'production').
	// This is used by the Policy Engine to enforce environment-specific rules.
	Environment string `json:"environment"`

	// RetryMaxAttempts defines the default maximum number of retry attempts for a step.
	RetryMaxAttempts int `json:"retry_max_attempts"`

	// RetryBaseDelayMS is the initial delay before the first retry.
	RetryBaseDelayMS int `json:"retry_base_delay_ms"`

	// RetryMaxDelayMS is the maximum bound for exponential backoff.
	RetryMaxDelayMS int `json:"retry_max_delay_ms"`

	// AuditLogPath is the location of the immutable, tamper-proof audit trail.
	AuditLogPath string `json:"audit_log_path"`

	// GlobalPolicies are system-wide CEL rules evaluated for every task (Phase 10).
	GlobalPolicies []string `json:"global_policies,omitempty"`

	// EnablePerformanceProfiling starts the pprof server if true (Phase 11).
	EnablePerformanceProfiling bool `json:"enable_performance_profiling"`

	// InterpolationEngine specifies which engine to use ('legacy' or 'fast').
	InterpolationEngine string `json:"interpolation_engine"`
}

// DefaultConfig returns a sane set of defaults for the GAIA Kernel.
func DefaultConfig() *KernelConfig {
	return &KernelConfig{
		PlannerProvider:    LLMProviderLocal,
		PlannerEndpoint:    "http://localhost:11434/api/generate", // Default Ollama endpoint
		PlannerModel:       "llama3",
		MaxReplans:         2,
		MaxConcurrentTasks: 10,
		MaxConcurrentPerAgent: 3,
		DBPath:             "./data/gaia_state.db",
		LogLevel:           "INFO",
		Environment:        "development",
		RetryMaxAttempts:   3,
		RetryBaseDelayMS:   500,
		RetryMaxDelayMS:    10000,
		AuditLogPath:       "./data/audit.log",
		GlobalPolicies: []string{
			"usage.tokens < 100000",
			"cost.usd < 10.0",
		},
		EnablePerformanceProfiling: false,
		InterpolationEngine:        "fast",
	}
}

// LoadConfigFromEnv overrides the default configuration with values from environment variables.
// This follows the 12-factor app methodology and improves monorepo developer experience.
func (c *KernelConfig) LoadConfigFromEnv() {
	if val := os.Getenv("GAIA_PLANNER_PROVIDER"); val != "" {
		c.PlannerProvider = LLMProvider(val)
	}
	if val := os.Getenv("GAIA_PLANNER_ENDPOINT"); val != "" {
		c.PlannerEndpoint = val
	}
	if val := os.Getenv("GAIA_PLANNER_MODEL"); val != "" {
		c.PlannerModel = val
	}
	if val := os.Getenv("GAIA_PLANNER_API_KEY"); val != "" {
		c.PlannerAPIKey = val
	}
	if val := os.Getenv("GAIA_DB_PATH"); val != "" {
		c.DBPath = val
	}
	if val := os.Getenv("GAIA_LOG_LEVEL"); val != "" {
		c.LogLevel = val
	}
	if val := os.Getenv("GAIA_AUDIT_LOG_PATH"); val != "" {
		c.AuditLogPath = val
	}
}
