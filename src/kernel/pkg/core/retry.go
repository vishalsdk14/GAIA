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

package core

import (
	"gaia/kernel/pkg/types"
	"math"
	"math/rand"
	"time"
)

// BackoffCalculator implements the formal backoff algorithms from docs/specs/failure-handling.md.
type BackoffCalculator struct {
	rand *rand.Rand
}

func NewBackoffCalculator() *BackoffCalculator {
	return &BackoffCalculator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetDelay calculates the sleep duration for a specific retry attempt.
func (c *BackoffCalculator) GetDelay(policy *types.RetryPolicy, attempt int) time.Duration {
	if policy == nil || attempt <= 0 {
		return 0
	}

	baseDelay := time.Duration(policy.BaseDelayMS) * time.Millisecond
	maxDelay := time.Duration(policy.MaxDelayMS) * time.Millisecond

	var delay time.Duration

	switch policy.Backoff {
	case "none":
		delay = 0
	case "linear":
		delay = baseDelay * time.Duration(attempt)
	case "exponential":
		// Formula: base_delay * 2^(attempt-1)
		pow := math.Pow(2, float64(attempt-1))
		delay = time.Duration(float64(baseDelay) * pow)
		
		// Apply +/- 20% Jitter
		jitter := float64(delay) * 0.20
		variation := (c.rand.Float64() * 2 * jitter) - jitter
		delay = time.Duration(float64(delay) + variation)
	default:
		delay = baseDelay
	}

	if delay > maxDelay {
		return maxDelay
	}
	return delay
}

// IsRetryable checks if a step can be retried based on failure type and idempotency.
func IsRetryable(step *types.Step, err *types.Error, agent *types.AgentManifest) bool {
	// 1. Check if error code is retryable (Soft Failure)
	switch err.Code {
	case types.ErrorCodeTimeout, types.ErrorCodeAgentUnavailable, types.ErrorCodeInternalError, types.ErrorCodeUnknown:
		// Retryable codes
	default:
		return false // Hard failures or policy violations are not retryable
	}

	// 2. Check Idempotency Constraints (Failure Handling Spec 2.2)
	// We check if the agent capability is explicitly marked as idempotent.
	var cap *types.Capability
	for i := range agent.Capabilities {
		if agent.Capabilities[i].Name == step.Capability {
			cap = &agent.Capabilities[i]
			break
		}
	}

	if cap == nil {
		return false
	}

	// Idempotent OR doesn't mutate state -> Safe to retry
	if cap.Idempotent || (cap.Constraints != nil && !cap.Constraints.MutatesState) {
		return true
	}

	// Non-idempotent AND mutates state -> Unsafe to retry
	return false
}
