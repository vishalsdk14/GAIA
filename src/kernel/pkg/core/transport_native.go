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
	"log/slog"
	"net/http"
	"time"
)

// NativeTransport implements the standard GAIA HTTP/REST protocol.
type NativeTransport struct{}

// Dispatch sends a standard GAIA JSON request to the agent's endpoint.
func (t *NativeTransport) Dispatch(req *types.Request, agent *types.AgentManifest) (*types.Response, error) {
	if agent.Endpoint == "" {
		return nil, fmt.Errorf("transport: native agent has no endpoint")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("transport: failed to marshal request: %w", err)
	}

	url := agent.Endpoint
	if url[len(url)-1] == '/' {
		url += "invoke"
	} else {
		url += "/invoke"
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("transport: failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if agent.Auth != nil && agent.Auth.Type == "bearer" {
		httpReq.Header.Set("Authorization", "Bearer "+agent.Auth.SecretRef)
	}

	startTime := time.Now()
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		slog.Error("NativeTransport: agent unavailable", "endpoint", agent.Endpoint, "error", err)
		return nil, fmt.Errorf("transport: %w: %v", fmt.Errorf(string(types.ErrorCodeAgentUnavailable)), err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	slog.Info("NativeTransport: received response", "endpoint", agent.Endpoint, "status", resp.StatusCode, "duration_ms", duration.Milliseconds())

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("transport: agent returned status %d", resp.StatusCode)
	}

	var gaiaResp types.Response
	if err := json.NewDecoder(resp.Body).Decode(&gaiaResp); err != nil {
		return nil, fmt.Errorf("transport: failed to decode agent response: %w", err)
	}

	if gaiaResp.Metrics == nil {
		gaiaResp.Metrics = &types.RequestMetrics{}
	}
	gaiaResp.Metrics.DurationMS = int(time.Since(startTime).Milliseconds())

	// Phase 11: Developer Experience - Log full response in DEBUG mode
	respBytes, _ := json.Marshal(gaiaResp)
	slog.Debug("NativeTransport: decoded response body", "endpoint", agent.Endpoint, "body", string(respBytes))

	return &gaiaResp, nil
}
