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
	tokens tokenIssuer
	logger *slog.Logger
}

type authRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret123" minLength:"8" maxLength:"72"`
}

type authResponse struct {
	ID    uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
	Email string    `json:"email" example:"user@example.com"`
}

type tokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type tokenIssuer interface {
	Issue(userID uuid.UUID) (string, error)
}

func NewAuthHandler(u userService, l *slog.Logger, t tokenIssuer) *AuthHandler {
	return &AuthHandler{users: u, tokens: t, logger: l}
}

// register godoc
//
//	@Summary	Register a new user
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		authRequest	true	"User credentials"
//	@Success	201		{object}	authResponse
//	@Failure	400		{object}	errorResponse	"Invalid request body, email or password"
//	@Failure	409		{object}	errorResponse	"Email already registered"
//	@Failure	500		{object}	errorResponse
//	@Router		/auth/register [post]
func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	err := decodeJSON(w, r, &req)

	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.users.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		handleServiceError(w, r, h.logger, err)
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{ID: user.ID, Email: user.Email})
}

// login godoc
//
//	@Summary		Log in
//	@Description	Returns a JWT to pass in the Authorization header as "Bearer <token>".
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		authRequest	true	"User credentials"
//	@Success		200		{object}	tokenResponse
//	@Failure		400		{object}	errorResponse	"Invalid request body"
//	@Failure		401		{object}	errorResponse	"Invalid email or password"
//	@Failure		500		{object}	errorResponse
//	@Router			/auth/login [post]
func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	err := decodeJSON(w, r, &req)

	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.users.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		handleServiceError(w, r, h.logger, err)
		return
	}

	token, err := h.tokens.Issue(user.ID)
	if err != nil {
		handleServiceError(w, r, h.logger, err)
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{Token: token})
}

func (h *AuthHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", h.register)
	mux.HandleFunc("POST /auth/login", h.login)
	return mux
}
