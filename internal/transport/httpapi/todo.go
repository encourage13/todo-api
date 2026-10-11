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
	Title       string `json:"title" example:"Buy milk" minLength:"1" maxLength:"200"`
	Description string `json:"description" example:"2 liters" maxLength:"2000"`
}

type updateTodoRequest struct {
	Title       *string `json:"title" example:"Buy milk" minLength:"1" maxLength:"200"`
	Description *string `json:"description" example:"2 liters" maxLength:"2000"`
	Completed   *bool   `json:"completed" example:"true"`
}

type todoResponse struct {
	ID          uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
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

// create godoc
//
//	@Summary	Create a todo
//	@Tags		todos
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		createTodoRequest	true	"Todo to create"
//	@Success	201		{object}	todoResponse
//	@Failure	400		{object}	errorResponse	"Invalid request body or validation error"
//	@Failure	401		{object}	errorResponse
//	@Failure	500		{object}	errorResponse
//	@Router		/todos [post]
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

// list godoc
//
//	@Summary		List todos
//	@Description	Returns all todos of the current user.
//	@Tags			todos
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		todoResponse
//	@Failure		401	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Router			/todos [get]
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

// get godoc
//
//	@Summary	Get a todo
//	@Tags		todos
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Todo ID"	Format(uuid)
//	@Success	200	{object}	todoResponse
//	@Failure	400	{object}	errorResponse	"Invalid todo id"
//	@Failure	401	{object}	errorResponse
//	@Failure	404	{object}	errorResponse	"Todo not found"
//	@Failure	500	{object}	errorResponse
//	@Router		/todos/{id} [get]
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

// delete godoc
//
//	@Summary	Delete a todo
//	@Tags		todos
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id	path	string	true	"Todo ID"	Format(uuid)
//	@Success	204
//	@Failure	400	{object}	errorResponse	"Invalid todo id"
//	@Failure	401	{object}	errorResponse
//	@Failure	404	{object}	errorResponse	"Todo not found"
//	@Failure	500	{object}	errorResponse
//	@Router		/todos/{id} [delete]
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

// update godoc
//
//	@Summary		Update a todo
//	@Description	Partially updates a todo. Only provided fields are changed; at least one field is required.
//	@Tags			todos
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Todo ID"	Format(uuid)
//	@Param			request	body		updateTodoRequest	true	"Fields to update"
//	@Success		200		{object}	todoResponse
//	@Failure		400		{object}	errorResponse	"Invalid todo id, request body or validation error"
//	@Failure		401		{object}	errorResponse
//	@Failure		404		{object}	errorResponse	"Todo not found"
//	@Failure		500		{object}	errorResponse
//	@Router			/todos/{id} [patch]
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
