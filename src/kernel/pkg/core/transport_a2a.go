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

// A2ATransport implements the Agent-to-Agent (A2A) protocol adapter.
// It maps GAIA requests to the A2A 'Agent Card' invocation format.
type A2ATransport struct{}

type a2aInvocation struct {
	Action     string      `json:"action"`
	Parameters interface{} `json:"parameters"`
	Context    interface{} `json:"context,omitempty"`
}

type a2aResult struct {
	Status    string      `json:"status"` // "success", "error"
	Artifact  interface{} `json:"artifact"`
	ErrorType string      `json:"error_type,omitempty"`
}

// Dispatch translates a GAIA request to an A2A invocation and transmits it.
func (t *A2ATransport) Dispatch(req *types.Request, agent *types.AgentManifest) (*types.Response, error) {
	if agent.Endpoint == "" {
		return nil, fmt.Errorf("transport: a2a agent has no endpoint")
	}

	// 1. Map GAIA -> A2A
	invoc := a2aInvocation{
		Action:     req.Capability,
		Parameters: req.Input,
	}

	body, _ := json.Marshal(invoc)
	httpReq, _ := http.NewRequest("POST", agent.Endpoint, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	startTime := time.Now()
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("transport: a2a request failed: %w", err)
	}
	defer resp.Body.Close()

	var result a2aResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("transport: failed to decode a2a result: %w", err)
	}

	// 2. Map A2A -> GAIA
	gaiaResp := &types.Response{
		RequestID: req.RequestID,
		Metrics: &types.RequestMetrics{
			DurationMS: int(time.Since(startTime).Milliseconds()),
		},
	}

	if result.Status == "success" {
		gaiaResp.Success = true
		gaiaResp.Output = result.Artifact
	} else {
		gaiaResp.Success = false
		gaiaResp.Error = &types.Error{
			Code:    types.ErrorCodeExecutionFailed,
			Message: "A2A Error: " + result.ErrorType,
		}
	}

	return gaiaResp, nil
}
