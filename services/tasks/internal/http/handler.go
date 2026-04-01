package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"tip2_pr2/services/tasks/internal/client/authclient"
	"tip2_pr2/services/tasks/internal/service"
)

type Handler struct {
	taskService *service.TaskService
	authClient  *authclient.Client
}

func New(taskService *service.TaskService, authClient *authclient.Client) *Handler {
	return &Handler{
		taskService: taskService,
		authClient:  authClient,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/tasks", h.withAuth(h.CreateTask))
	mux.HandleFunc("GET /v1/tasks", h.withAuth(h.ListTasks))
	mux.HandleFunc("GET /v1/tasks/{id}", h.withAuth(h.GetTask))
	mux.HandleFunc("PATCH /v1/tasks/{id}", h.withAuth(h.UpdateTask))
	mux.HandleFunc("DELETE /v1/tasks/{id}", h.withAuth(h.DeleteTask))
}

func (h *Handler) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token, ok := extractBearerToken(authHeader)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "unauthorized",
			})
			return
		}

		ok, status, err := h.authClient.Verify(r.Context(), token)
		if err != nil {
			writeJSON(w, status, map[string]string{
				"error": "auth service unavailable",
			})
			return
		}

		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "unauthorized",
			})
			return
		}

		next(w, r)
	}
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req service.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json",
		})
		return
	}

	task, err := h.taskService.Create(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.taskService.List())
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	task, err := h.taskService.Get(id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "task not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal error",
		})
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req service.UpdateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json",
		})
		return
	}

	task, err := h.taskService.Update(id, req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "task not found",
			})
			return
		}

		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.taskService.Delete(id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "task not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal error",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
