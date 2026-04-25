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
	"net/http"

	"github.com/go-chi/chi/v5"
)

type createTaskRequest struct {
	Goal string `json:"goal"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Goal == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Goal is required"})
		return
	}

	task, err := s.orchestrator.SubmitTask(req.Goal)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusCreated, task)
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	task, err := s.orchestrator.GetTask(taskID)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if task == nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Task not found"})
		return
	}

	jsonResponse(w, http.StatusOK, task)
}

func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	agents := s.registry.ListAgents()
	jsonResponse(w, http.StatusOK, agents)
}

func (s *Server) handleListCapabilities(w http.ResponseWriter, r *http.Request) {
	caps := s.registry.ListCapabilities()
	jsonResponse(w, http.StatusOK, caps)
}

func (s *Server) handleApproveStep(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	stepID := chi.URLParam(r, "stepID")
	
	if err := s.orchestrator.ApproveStep(taskID, stepID); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "approved"})
}
