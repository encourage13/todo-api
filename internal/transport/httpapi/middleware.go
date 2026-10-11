package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type tokenParser interface {
	Parse(token string) (uuid.UUID, error)
}

type meResponse struct {
	ID uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
}

type ctxKey struct{}
type requestIDKey struct{}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	value := ctx.Value(ctxKey{})
	v, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}
	return v, true
}

func RequireAuth(parser tokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			headerWithoutPrefix, ok := strings.CutPrefix(header, "Bearer ")

			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			userID, err := parser.Parse(headerWithoutPrefix)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})

	}
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestIDFromContext(ctx context.Context) string {
	value, ok := ctx.Value(requestIDKey{}).(string)
	if !ok {
		return ""
	}
	return value
}

// Me godoc
//
//	@Summary	Get current user
//	@Tags		auth
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	meResponse
//	@Failure	401	{object}	errorResponse
//	@Failure	500	{object}	errorResponse
//	@Router		/me [get]
func Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, meResponse{ID: userID})
}
