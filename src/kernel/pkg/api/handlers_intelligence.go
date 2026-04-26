// Copyright 2026 GAIA Contributors
//
// Intelligence Proxy Handlers
// These endpoints allow agents to request LLM services (text/vision) from the Kernel.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type intelligenceCompleteRequest struct {
	Prompt string `json:"prompt"`
}

type intelligenceVisionRequest struct {
	Prompt      string `json:"prompt"`
	ImageBase64 string `json:"image_base64"`
}

type intelligenceResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (s *Server) handleIntelligenceComplete(w http.ResponseWriter, r *http.Request) {
	agentID := r.Header.Get("X-GAIA-Agent-ID")
	slog.Info("Intelligence request: Complete", "agent_id", agentID)

	var req intelligenceCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, intelligenceResponse{Success: false, Error: "Invalid request body"})
		return
	}

	result, err := s.orchestrator.Complete(req.Prompt)
	if err != nil {
		slog.Error("Intelligence failed: Complete", "agent_id", agentID, "error", err)
		jsonResponse(w, http.StatusInternalServerError, intelligenceResponse{Success: false, Error: err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, intelligenceResponse{Success: true, Result: result})
}

func (s *Server) handleIntelligenceVision(w http.ResponseWriter, r *http.Request) {
	agentID := r.Header.Get("X-GAIA-Agent-ID")
	slog.Info("Intelligence request: Vision", "agent_id", agentID)

	var req intelligenceVisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, intelligenceResponse{Success: false, Error: "Invalid request body"})
		return
	}

	result, err := s.orchestrator.Vision(req.Prompt, req.ImageBase64)
	if err != nil {
		slog.Error("Intelligence failed: Vision", "agent_id", agentID, "error", err)
		jsonResponse(w, http.StatusInternalServerError, intelligenceResponse{Success: false, Error: err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, intelligenceResponse{Success: true, Result: result})
}
