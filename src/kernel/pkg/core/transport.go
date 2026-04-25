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
	"fmt"
	"gaia/kernel/pkg/types"
	"time"
)

// AgentTransport defines the network layer abstraction for dispatching requests to agents.
// This decouples the Control Loop from the underlying protocols (MCP, A2A, Native).
type AgentTransport interface {
	// Dispatch sends a request to the agent and waits for a response or ACK.
	Dispatch(req *types.Request, agent *types.AgentManifest) (*types.Response, error)
}

// MockTransport provides a Foundation phase scaffold until real network protocols
// (like MCP/A2A) are implemented in Task 4.
type MockTransport struct{}

// Dispatch simulates network latency and returns a mock success response.
func (m *MockTransport) Dispatch(req *types.Request, agent *types.AgentManifest) (*types.Response, error) {
	fmt.Printf("Transport: Dispatching Step=%s to Agent=%s (Capability=%s)\n", req.StepID, agent.AgentID, req.Capability)

	// Simulate network latency
	time.Sleep(500 * time.Millisecond)

	// Return a dummy successful response
	return &types.Response{
		RequestID: req.RequestID,
		Success:   true,
		Output: map[string]interface{}{
			"status": "mock_success",
			"note":   "Task 4 will implement real MCP/A2A network calls.",
		},
		Metrics: &types.RequestMetrics{
			DurationMS: 500,
		},
	}, nil
}
