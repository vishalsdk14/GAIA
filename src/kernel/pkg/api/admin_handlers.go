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
	"bufio"
	"encoding/json"
	"gaia/kernel/pkg/logger"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

/**
 * handleListAuditLogs returns the full tamper-proof audit trail.
 * In a production environment, this would support pagination and date filtering.
 */
func (s *Server) handleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("audit.log") // Should use config.AuditLogPath
	if err != nil {
		if os.IsNotExist(err) {
			jsonResponse(w, http.StatusOK, []interface{}{})
			return
		}
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer f.Close()

	var logs []logger.AuditEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry logger.AuditEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			logs = append(logs, entry)
		}
	}

	jsonResponse(w, http.StatusOK, logs)
}

/**
 * handleVerifyAuditIntegrity manually triggers a full chain validation.
 */
func (s *Server) handleVerifyAuditIntegrity(w http.ResponseWriter, r *http.Request) {
	al := logger.GetAuditLogger()
	if al == nil {
		jsonResponse(w, http.StatusServiceUnavailable, map[string]string{"error": "Audit logger not initialized"})
		return
	}

	if err := al.VerifyChain(); err != nil {
		jsonResponse(w, http.StatusConflict, map[string]string{
			"status": "TAMPERED",
			"error":  err.Error(),
		})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "OK", "message": "Audit chain integrity verified"})
}

/**
 * handleRestoreAgentState rolls back an agent's managed state to a target timestamp.
 */
func (s *Server) handleRestoreAgentState(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	targetTime := r.URL.Query().Get("target_time") // Format: 2006-01-02 15:04:05

	if targetTime == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "target_time query parameter is required"})
		return
	}

	if err := s.agentStore.RestoreAgentState(agentID, targetTime); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{
		"status":     "RESTORED",
		"agent_id":   agentID,
		"target_time": targetTime,
	})
}
