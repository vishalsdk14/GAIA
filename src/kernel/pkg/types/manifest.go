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
// This file specifically implements the schemas related to Agent Identity and Capabilities,
// mapping the AgentManifest JSON schema to strictly typed Go structs.
package types

// Transport defines the underlying network transport used by an agent.
type Transport string

const (
	TransportHTTP      Transport = "http"
	TransportGRPC      Transport = "grpc"
	TransportIPC       Transport = "ipc"
	TransportWebSocket Transport = "websocket"
)

// Protocol defines the communication protocol dialect.
type Protocol string

const (
	ProtocolNative Protocol = "native"
	ProtocolA2A    Protocol = "a2a"
	ProtocolMCP    Protocol = "mcp"
)

// AgentManifest represents the authoritative registration record for a GAIA agent.
type AgentManifest struct {
	AgentID            string              `json:"agent_id"`
	Version            string              `json:"version"`
	Transport          Transport           `json:"transport"`
	Protocol           Protocol            `json:"protocol"`
	Endpoint           string              `json:"endpoint"`
	HealthEndpoint     string              `json:"health_endpoint,omitempty"`
	HealthEndpointSpec *HealthEndpointSpec `json:"health_endpoint_spec,omitempty"`
	Invoke             InvokeContract      `json:"invoke"`
	Capabilities       []Capability        `json:"capabilities"`
	Auth               *AuthConfig         `json:"auth,omitempty"`
	StateRequirements  *StateRequirements  `json:"state_requirements,omitempty"`
}

// HealthEndpointSpec defines the protocol expected for health checks.
type HealthEndpointSpec struct {
	Method           string                 `json:"method"`
	ExpectedResponse map[string]interface{} `json:"expected_response,omitempty"`
	TimeoutMS        int                    `json:"timeout_ms"`
}

// InvokeContract defines default timeout and async support for agent invocations.
type InvokeContract struct {
	TimeoutMS      int  `json:"timeout_ms"`
	AsyncSupported bool `json:"async_supported"`
}

// StateRequirements defines Managed Agent State requirements (Tier 4).
type StateRequirements struct {
	Required bool `json:"required"`
	MaxBytes int  `json:"max_bytes,omitempty"`
}

// AuthConfig defines authentication and authorization configuration.
type AuthConfig struct {
	Type      string   `json:"type"`
	SecretRef string   `json:"secret_ref,omitempty"`
	Scopes    []string `json:"scopes,omitempty"`
}

// Capability defines a specific skill or tool offered by an agent.
type Capability struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	InputSchema  map[string]interface{} `json:"input_schema"`
	OutputSchema map[string]interface{} `json:"output_schema"`
	Idempotent   bool                   `json:"idempotent"`
	Constraints  *Constraints           `json:"constraints,omitempty"`
}

// Constraints defines behavioral constraints declared by the agent.
type Constraints struct {
	ReadOnly     bool `json:"read_only"`
	MutatesState bool `json:"mutates_state"`
	ExternalIO   bool `json:"external_io"`
}

