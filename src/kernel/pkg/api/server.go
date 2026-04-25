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

package api

import (
	"encoding/json"
	"gaia/kernel/pkg/core"
	"gaia/kernel/pkg/registry"
	"gaia/kernel/pkg/state"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server implements the GAIA Kernel HTTP Gateway.
type Server struct {
	router       *chi.Mux
	orchestrator *core.Orchestrator
	registry     registry.CapabilityRegistry
	agentStore   *state.SQLiteStore
}

// NewServer initializes the HTTP router and wires up the kernel subsystems.
func NewServer(o *core.Orchestrator, r registry.CapabilityRegistry, as *state.SQLiteStore) *Server {
	s := &Server{
		router:       chi.NewRouter(),
		orchestrator: o,
		registry:     r,
		agentStore:   as,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))
}

func (s *Server) setupRoutes() {
	s.router.Route("/api/v1", func(r chi.Router) {
		// Task Management
		r.Post("/tasks", s.handleCreateTask)
		r.Get("/tasks/{taskID}", s.handleGetTask)
		
		// Registry
		r.Get("/registry/agents", s.handleListAgents)
		r.Get("/registry/capabilities", s.handleListCapabilities)

		// Real-time Streaming
		r.Get("/stream", s.handleStream)
	})

	s.router.Route("/internal/v1", func(r chi.Router) {
		// Managed Agent State (Tier 4)
		r.Route("/state", func(r chi.Router) {
			r.Get("/", s.handleListStateKeys)
			r.Get("/{key}", s.handleGetState)
			r.Put("/{key}", s.handlePutState)
			r.Delete("/{key}", s.handleDeleteState)
		})
	})
}

// Start launches the HTTP server on the specified address.
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

// jsonResponse is a helper to write JSON objects to the response writer.
func jsonResponse(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
