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

// Package state implements the multi-tiered state management architecture for the GAIA kernel.
package state

import (
	"fmt"
	"gaia/kernel/pkg/types"
	"sync"
)

// AgentStateStore defines the Interface for the Managed Agent State (Tier 4).
// Using an interface allows us to swap the in-memory scaffold for SQLite/PostgreSQL in Phase 3.
type AgentStateStore interface {
	Put(agentID string, key string, data interface{}, manifest *types.AgentManifest) error
	Get(agentID string, key string) (interface{}, error)
	Delete(agentID string, key string) error
	DeleteNamespace(agentID string) error // The "Kill Switch"
}

// InMemoryAgentStore is the Phase 2 (Foundation) scaffold adapter for Tier 4.
type InMemoryAgentStore struct {
	mu    sync.RWMutex
	store map[string]map[string]interface{}
}

// NewInMemoryAgentStore initializes the scaffold database for Tier 4.
// This is strictly used for Phase 2 development and testing. It creates
// the root map that will house the individual agent namespaces.
func NewInMemoryAgentStore() *InMemoryAgentStore {
	return &InMemoryAgentStore{
		store: make(map[string]map[string]interface{}),
	}
}

// Put saves a key-value document into the Managed Agent State.
// It enforces strict multi-tenancy by partitioning data by agentID.
// Before allowing the write, it validates the agent's manifest to ensure
// the agent explicitly requested `state_requirements.required = true`.
// If the namespace does not exist, it is created securely.
func (s *InMemoryAgentStore) Put(agentID string, key string, data interface{}, manifest *types.AgentManifest) error {
	// Verify the agent explicitly requested state storage during Handshake
	if manifest.StateRequirements == nil || !manifest.StateRequirements.Required {
		return fmt.Errorf("state: %w: agent did not request state_requirements in manifest", fmt.Errorf(string(types.ErrorCodePolicyDenied)))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Initialize namespace if it doesn't exist
	if s.store[agentID] == nil {
		s.store[agentID] = make(map[string]interface{})
	}

	// NOTE: Quota enforcement (max_bytes) is omitted in this in-memory scaffold,
	// but will be implemented in the Phase 3 SQLite adapter.
	s.store[agentID][key] = data
	return nil
}

// Get retrieves a document strictly from the agent's partitioned namespace.
// It is fully thread-safe via RWMutex. If the agent namespace or the specific
// key does not exist, it returns nil rather than an error, allowing agents to
// check for existence without triggering failure flows.
func (s *InMemoryAgentStore) Get(agentID string, key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	namespace, exists := s.store[agentID]
	if !exists {
		return nil, nil // Return nil if namespace doesn't exist (not an error)
	}

	data, ok := namespace[key]
	if !ok {
		return nil, nil // Return nil if key doesn't exist
	}
	return data, nil
}

// Delete removes a specific key from an agent's namespace.
// If the key or the namespace does not exist, it performs a no-op silently
// to ensure idempotency. It is fully locked during the map deletion.
func (s *InMemoryAgentStore) Delete(agentID string, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if namespace, exists := s.store[agentID]; exists {
		delete(namespace, key)
	}
	return nil
}

// DeleteNamespace is the administrative "Kill Switch".
// It completely drops the database partition for a specific agent.
// This is typically invoked by the Capability Registry when an agent
// is blacklisted or permanently ejected from the system.
func (s *InMemoryAgentStore) DeleteNamespace(agentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.store, agentID)
	return nil
}
