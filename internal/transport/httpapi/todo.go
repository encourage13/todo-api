package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/encourage13/todo-api/internal/model"
	"github.com/encourage13/todo-api/internal/service"
	"github.com/google/uuid"
)

type todoService interface {
	Get(ctx context.Context, userID, id uuid.UUID) (*model.Todo, error)
	List(ctx context.Context, userID uuid.UUID) ([]model.Todo, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	Create(ctx context.Context, userID uuid.UUID, title, description string) (*model.Todo, error)
	Update(ctx context.Context, userID, id uuid.UUID, in service.UpdateTodoInput) (*model.Todo, error)
}

type TodoHandler struct {
	todos  todoService
	logger *slog.Logger
}

type createTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateTodoRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}

type todoResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewTodoHandler(t todoService, l *slog.Logger) *TodoHandler {
	return &TodoHandler{todos: t, logger: l}
}

func toTodoResponse(t model.Todo) todoResponse {
	return todoResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Completed:   t.Completed,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func requireUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal error")
		return uuid.Nil, false
	}
	return userID, true
}

func parseIDParam(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	value := r.PathValue("id")
	id, err := uuid.Parse(value)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid todo id")
		return uuid.Nil, false
	}
	return id, true
}

func (h *TodoHandler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req createTodoRequest
	err := decodeJSON(w, r, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	todo, err := h.todos.Create(r.Context(), userID, req.Title, req.Description)
	if err != nil {
		handleServiceError(w, r, h.logger, err)
		return
	}
	writeJSON(w, http.StatusCreated, toTodoResponse(*todo))
}

func (h *TodoHandler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	todos, err := h.todos.List(r.Context(), userID)
	if err != nil {
		handleServiceError(w, r, h.logger, err)
		return
	}

	items := make([]todoResponse, 0)
	for _, todo := range todos {
		items = append(items, toTodoResponse(todo))
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *TodoHandler) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	todo, err := h.todos.Get(r.Context(), userID, id)
	if err != nil {
		handleServiceError(w, r, h.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, toTodoResponse(*todo))
}

func (h *TodoHandler) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	if err := h.todos.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TodoHandler) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	var req updateTodoRequest
	err := decodeJSON(w, r, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updateReq := service.UpdateTodoInput{
		Title:       req.Title,
		Description: req.Description,
		Completed:   req.Completed,
	}

	todo, err := h.todos.Update(r.Context(), userID, id, updateReq)
	if err != nil {
		handleServiceError(w, r, h.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, toTodoResponse(*todo))
}

func (h *TodoHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /todos", h.create)
	mux.HandleFunc("GET /todos", h.list)
	mux.HandleFunc("GET /todos/{id}", h.get)
	mux.HandleFunc("DELETE /todos/{id}", h.delete)
	mux.HandleFunc("PATCH /todos/{id}", h.update)
	return mux
}
