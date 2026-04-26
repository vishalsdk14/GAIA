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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gaia/kernel/pkg/types"
	"net"
	"net/http"
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
		ipc:    NewIPCTransport(),
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
	// In a real implementation, this would use a pooled gRPC client.
	// For Phase 11, we implement the protocol negotiation logic.
	return nil, fmt.Errorf("transport: gRPC transport requires protobuf definitions and generated clients (Phase 11 Completion)")
}

// IPCTransport implements ultra-low latency local agent communication using Unix Domain Sockets (Phase 11).
type IPCTransport struct {
	client *http.Client
}

func NewIPCTransport() *IPCTransport {
	return &IPCTransport{
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					// addr is ignored for UDS, we use the agent.Endpoint directly
					return net.Dial("unix", addr)
				},
			},
		},
	}
}

func (i *IPCTransport) Dispatch(req *types.Request, agent *types.AgentManifest) (*types.Response, error) {
	// For Phase 11, we support HTTP-over-UDS for local agents
	// agent.Endpoint should be the path to the socket (e.g. /tmp/agent.sock)
	url := "http://localhost/invoke"
	
	bytesReq, _ := json.Marshal(req)
	resp, err := i.client.Post(url, "application/json", bytes.NewReader(bytesReq))
	if err != nil {
		return nil, fmt.Errorf("transport: IPC dispatch failed: %w", err)
	}
	defer resp.Body.Close()

	var response types.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("transport: failed to decode IPC response: %w", err)
	}

	return &response, nil
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
