package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/encourage13/todo-api/internal/model"
)

func handleServiceError(w http.ResponseWriter, logger *slog.Logger, err error) {
	switch {
	case errors.Is(err, model.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, model.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, model.ErrEmailTaken):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, model.ErrTodoNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		logger.Error("internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
