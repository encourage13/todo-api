package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/encourage13/todo-api/internal/model"
	"github.com/google/uuid"
)

type userService interface {
	Register(ctx context.Context, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (*model.User, error)
}

type AuthHandler struct {
	users  userService
	logger *slog.Logger
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func NewAuthHandler(u userService, l *slog.Logger) *AuthHandler {
	return &AuthHandler{users: u, logger: l}
}

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	err := decodeJSON(w, r, &req)

	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.users.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		handleServiceError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{ID: user.ID, Email: user.Email})
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	err := decodeJSON(w, r, &req)

	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.users.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		handleServiceError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusOK, authResponse{ID: user.ID, Email: user.Email})
}

func (h *AuthHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", h.register)
	mux.HandleFunc("POST /auth/login", h.login)
	return mux
}
