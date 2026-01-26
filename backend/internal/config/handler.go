package config

import (
	"encoding/json"
	"net/http"
	"os"
)

type config struct {
	ChatbotEnabled  bool   `json:"chatbot_enabled"`
	ChatbotApiUrl   string `json:"chatbot_api_url"`
	ChatbotApiToken string `json:"chatbot_api_token"`
}

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) getConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	enabled := os.Getenv("CHATBOT_ENABLED") == "true"
	apiUrl := os.Getenv("CHATBOT_API_URL")
	token := os.Getenv("CHATBOT_API_TOKEN")

	// Ensure it's only enabled if credentials are actually provided
	if enabled && (apiUrl == "" || token == "") {
		enabled = false
	}

	cfg := config{
		ChatbotEnabled:  enabled,
		ChatbotApiUrl:   apiUrl,
		ChatbotApiToken: token,
	}

	json.NewEncoder(w).Encode(cfg)
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, mw func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/config", mw(h.getConfig))
}
