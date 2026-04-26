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
)

// AgentTransport defines the network layer abstraction for dispatching requests to agents.
// This decouples the Control Loop from the underlying protocols (MCP, A2A, Native).
type AgentTransport interface {
	// Dispatch sends a request to the agent and waits for a response or ACK.
	Dispatch(req *types.Request, agent *types.AgentManifest) (*types.Response, error)
}

// ProtocolDispatcher is a switchboard that routes requests to the correct adapter.
type ProtocolDispatcher struct {
	native *NativeTransport
	mcp    *MCPTransport
	a2a    *A2ATransport
	ws     *WSTransport
	grpc   *GRPCTransport
	ipc    *IPCTransport
}

func NewProtocolDispatcher() *ProtocolDispatcher {
	return &ProtocolDispatcher{
		native: &NativeTransport{},
		mcp:    &MCPTransport{},
		a2a:    &A2ATransport{},
		ws:     NewWSTransport(),
		grpc:   &GRPCTransport{},
		ipc:    &IPCTransport{},
	}
}

func (d *ProtocolDispatcher) Dispatch(req *types.Request, agent *types.AgentManifest) (*types.Response, error) {
	// Phase 11: Hybrid Routing
	// If transport is IPC or Endpoint starts with ipc://, use IPC path regardless of protocol adapter
	if agent.Transport == types.TransportIPC {
		return d.ipc.Dispatch(req, agent)
	}

	// Route based on protocol
	switch agent.Protocol {
	case types.ProtocolNative:
		if agent.Transport == types.TransportGRPC {
			return d.grpc.Dispatch(req, agent)
		}
		return d.native.Dispatch(req, agent)
	case types.ProtocolMCP:
		return d.mcp.Dispatch(req, agent)
	case types.ProtocolA2A:
		return d.a2a.Dispatch(req, agent)
	case types.ProtocolWebSocket:
		return d.ws.Dispatch(req, agent)
	default:
		return nil, fmt.Errorf("transport: unsupported protocol: %s", agent.Protocol)
	}
}

// GRPCTransport implements high-performance remote agent communication (Phase 11).
type GRPCTransport struct{}

func (g *GRPCTransport) Dispatch(req *types.Request, agent *types.AgentManifest) (*types.Response, error) {
	// TODO: Implement gRPC client logic
	return nil, fmt.Errorf("transport: gRPC support is in implementation")
}

// IPCTransport implements zero-latency local agent communication (Phase 11).
type IPCTransport struct{}

func (i *IPCTransport) Dispatch(req *types.Request, agent *types.AgentManifest) (*types.Response, error) {
	// TODO: Implement Unix Domain Socket / Shared Memory logic
	return nil, fmt.Errorf("transport: IPC support is in implementation")
}

// MockTransport provides a Foundation phase scaffold for testing.
type MockTransport struct{}

func (m *MockTransport) Dispatch(req *types.Request, agent *types.AgentManifest) (*types.Response, error) {
	return &types.Response{
		RequestID: req.RequestID,
		Success:   true,
		Output: map[string]interface{}{"status": "mock_success"},
	}, nil
}
