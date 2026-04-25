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

// Package api implements the external and internal HTTP interfaces for the GAIA Kernel.
// This file specifically implements the Managed Agent State (Tier 4) API, allowing
// connected agents to persist and retrieve state documents securely.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// getAgentID extracts the agent identity from the request headers.
// Note: In the foundation phase, this relies on X-Agent-ID; Phase 8 will transition
// this to verified mTLS / JWT identity extraction.
func getAgentID(r *http.Request) string {
	return r.Header.Get("X-Agent-ID")
}

// handleGetState retrieves a JSON document for a specific key within an agent's namespace.
// It enforces strict isolation: agents can only access their own partitioned data.
func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
	agentID := getAgentID(r)
	if agentID == "" {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Agent ID required"})
		return
	}

	key := chi.URLParam(r, "key")
	data, err := s.agentStore.Get(agentID, key)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if data == nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Key not found"})
		return
	}

	jsonResponse(w, http.StatusOK, data)
}

// handlePutState stores or overwrites a JSON document for a specific key.
// It performs a policy check against the agent's manifest to verify state access
// permissions and storage quota (max_bytes) before committing the write.
func (s *Server) handlePutState(w http.ResponseWriter, r *http.Request) {
	agentID := getAgentID(r)
	if agentID == "" {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Agent ID required"})
		return
	}

	// Fetch manifest to check quota (real impl would pull from registry)
	agent, _ := s.registry.SelectAgent("any") // Mocking for now
	if agent == nil {
		jsonResponse(w, http.StatusForbidden, map[string]string{"error": "Agent not registered"})
		return
	}

	var data interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	key := chi.URLParam(r, "key")
	if err := s.agentStore.Put(agentID, key, data, &agent.Manifest); err != nil {
		jsonResponse(w, http.StatusRequestEntityTooLarge, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "stored"})
}

// handleDeleteState removes a specific key-value pair from the agent's namespace.
// This operation is idempotent and returns 204 No Content on success.
func (s *Server) handleDeleteState(w http.ResponseWriter, r *http.Request) {
	agentID := getAgentID(r)
	if agentID == "" {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Agent ID required"})
		return
	}

	key := chi.URLParam(r, "key")
	if err := s.agentStore.Delete(agentID, key); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleListStateKeys retrieves a paginated list of all keys currently stored by the agent.
// It also returns the current storage usage in bytes to help agents manage their quotas.
func (s *Server) handleListStateKeys(w http.ResponseWriter, r *http.Request) {
	agentID := getAgentID(r)
	if agentID == "" {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "Agent ID required"})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 { limit = 100 }
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	keys, err := s.agentStore.ListKeys(agentID, limit, offset)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	usage, _ := s.agentStore.GetUsage(agentID)
	
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"keys":       keys,
		"total_keys": len(keys),
		"bytes_used": usage,
	})
}
