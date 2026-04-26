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

// Package registry implements the Capability Registry for the GAIA Kernel.
// It acts as the DNS and Load Balancer, tracking agent lifecycles and routing
// steps to the most reliable agents based on real-time health metrics.
package registry

import (
	"gaia/kernel/pkg/types"
	"sync"
)

// CapabilityRegistry defines the core interface for agent and capability management.
// Defining this as an interface ensures the Kernel is decoupled from the storage
// layer, allowing us to swap the in-memory map for a distributed KV store (e.g., Redis)
// in a multi-node production deployment.
type CapabilityRegistry interface {
	Register(manifest *types.AgentManifest) error
	Deregister(agentID string) error
	UpdateHealth(agentID string, success bool, latencyMS int) error
	Heartbeat(agentID string) error
	SelectAgent(capability string) (*types.AgentRecord, error)
	ListAgents() []*types.AgentRecord
	ListCapabilities() []string
}

// InMemoryRegistry is the Foundation Phase implementation of the CapabilityRegistry.
// It relies on concurrent maps protected by RWMutexes for safe parallel access.
type InMemoryRegistry struct {
	mu           sync.RWMutex
	agents       map[string]*types.AgentRecord
	capabilities map[string][]string // capability name -> list of agent_ids
}

// NewInMemoryRegistry initializes a new capability registry.
// Maps are allocated empty.
func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		agents:       make(map[string]*types.AgentRecord),
		capabilities: make(map[string][]string),
	}
}

// ListAgents returns all registered agents.
func (r *InMemoryRegistry) ListAgents() []*types.AgentRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	list := make([]*types.AgentRecord, 0, len(r.agents))
	for _, a := range r.agents {
		list = append(list, a)
	}
	return list
}

// ListCapabilities returns all registered capabilities.
func (r *InMemoryRegistry) ListCapabilities() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]string, 0, len(r.capabilities))
	for c := range r.capabilities {
		list = append(list, c)
	}
	return list
}
