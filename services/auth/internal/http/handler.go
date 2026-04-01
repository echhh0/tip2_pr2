package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"tip2_pr2/services/auth/internal/service"
)

type Handler struct {
	authService *service.AuthService
}

func New(authService *service.AuthService) *Handler {
	return &Handler{authService: authService}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/auth/login", h.Login)
	mux.HandleFunc("GET /v1/auth/verify", h.Verify)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json",
		})
		return
	}

	token, ok := h.authService.Login(req.Username, req.Password)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "invalid credentials",
		})
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	})
}

func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	token, ok := extractBearerToken(authHeader)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"valid": false,
			"error": "unauthorized",
		})
		return
	}

	result := h.authService.VerifyHTTP(token)
	if !result.Valid {
		writeJSON(w, http.StatusUnauthorized, result)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func extractBearerToken(authHeader string) (string, bool) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return "", false
	}
	if parts[0] != "Bearer" {
		return "", false
	}
	if strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return parts[1], true
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
