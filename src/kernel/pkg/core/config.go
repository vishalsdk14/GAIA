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
	
	// PlannerAPIKey is the secret used for authenticating with cloud LLM providers.
	PlannerAPIKey string `json:"planner_api_key,omitempty"`
	
	// MaxReplans defines the circuit breaker limit for how many times a task can be replanned.
	MaxReplans int `json:"max_replans"`
	
	// MaxConcurrentTasks defines the maximum number of parallel tasks the kernel will process.
	MaxConcurrentTasks int `json:"max_concurrent_tasks"`
}

// DefaultConfig returns a sane set of defaults for the GAIA Kernel.
func DefaultConfig() *KernelConfig {
	return &KernelConfig{
		PlannerProvider:    LLMProviderLocal,
		PlannerEndpoint:    "http://localhost:11434/api/generate", // Default Ollama endpoint
		MaxReplans:         2,
		MaxConcurrentTasks: 10,
	}
}
