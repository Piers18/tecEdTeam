package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"movie-tracker/internal/chat"
	"movie-tracker/internal/respond"
)

type ChatHandler struct {
	agent *chat.Agent
}

func NewChatHandler(agent *chat.Agent) *ChatHandler {
	return &ChatHandler{agent: agent}
}

func (h *ChatHandler) Send(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID string `json:"session_id"`
		Message   string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Message == "" {
		respond.Error(w, http.StatusBadRequest, "message is required")
		return
	}

	sessionID := body.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	reply, err := h.agent.Chat(r.Context(), sessionID, body.Message)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "agent error: "+err.Error())
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{
		"session_id": sessionID,
		"reply":      reply,
	})
}

func (h *ChatHandler) ClearSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		respond.Error(w, http.StatusBadRequest, "sessionId is required")
		return
	}
	h.agent.ClearSession(sessionID)
	w.WriteHeader(http.StatusNoContent)
}
