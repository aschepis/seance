package api

import (
	"encoding/json"
	"net/http"

	"github.com/aschepis/seance/internal/conversations"
)

// Handler provides HTTP handlers for the conversation API.
type Handler struct {
	parser *conversations.Parser
}

// NewHandler creates a new API handler.
func NewHandler(parser *conversations.Parser) *Handler {
	return &Handler{parser: parser}
}

// RegisterRoutes registers all API routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/conversations", h.listConversations)
	mux.HandleFunc("GET /api/conversations/{id}", h.getConversation)
	mux.HandleFunc("GET /api/search", h.search)
}

func (h *Handler) listConversations(w http.ResponseWriter, r *http.Request) {
	groups, err := h.parser.ListProjects()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, groups)
}

func (h *Handler) getConversation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing conversation id")
		return
	}

	conv, err := h.parser.GetConversation(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, conv)
}

func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "missing search query parameter 'q'")
		return
	}

	results, err := h.parser.Search(query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
