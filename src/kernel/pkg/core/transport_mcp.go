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
	"encoding/json"
	"fmt"
	"gaia/kernel/pkg/types"
	"net/http"
	"time"
)

// MCPTransport implements the Model Context Protocol (MCP) adapter.
// It translates GAIA requests into MCP 'tools/call' JSON-RPC requests.
type MCPTransport struct{}

type mcpRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type mcpCallToolParams struct {
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments"`
}

type mcpResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result"`
	Error   *mcpError   `json:"error"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Dispatch translates a GAIA request to an MCP tool call and transmits it via HTTP.
func (t *MCPTransport) Dispatch(req *types.Request, agent *types.AgentManifest) (*types.Response, error) {
	if agent.Endpoint == "" {
		return nil, fmt.Errorf("transport: mcp agent has no endpoint")
	}

	// 1. Wrap GAIA request into MCP CallTool JSON-RPC
	mcpReq := mcpRequest{
		JSONRPC: "2.0",
		ID:      req.RequestID,
		Method:  "tools/call",
		Params: mcpCallToolParams{
			Name:      req.Capability,
			Arguments: req.Input,
		},
	}

	body, _ := json.Marshal(mcpReq)
	httpReq, _ := http.NewRequest("POST", agent.Endpoint, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	startTime := time.Now()
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("transport: mcp request failed: %w", err)
	}
	defer resp.Body.Close()

	var mcpResp mcpResponse
	if err := json.NewDecoder(resp.Body).Decode(&mcpResp); err != nil {
		return nil, fmt.Errorf("transport: failed to decode mcp response: %w", err)
	}

	if mcpResp.Error != nil {
		return &types.Response{
			RequestID: req.RequestID,
			Success:   false,
			Error: &types.Error{
				Code:    types.ErrorCodeExecutionFailed,
				Message: mcpResp.Error.Message,
			},
		}, nil
	}

	// 2. Map MCP Result back to GAIA Response
	return &types.Response{
		RequestID: req.RequestID,
		Success:   true,
		Output:    mcpResp.Result,
		Metrics: &types.RequestMetrics{
			DurationMS: int(time.Since(startTime).Milliseconds()),
		},
	}, nil
}
