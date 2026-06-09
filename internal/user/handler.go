package user

import (
	"errors"
	"net/http"

	"github.com/user/todo-list/internal/middleware"
	"github.com/user/todo-list/internal/platform/response"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// GetMe returns the currently authenticated user.
//
//	@Summary		Get current user
//	@Description	Returns the profile of the authenticated user
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.envelope{data=user.Response}
//	@Failure		401	{object}	response.errorEnvelope
//	@Failure		404	{object}	response.errorEnvelope
//	@Router			/me [get]
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}
	u, err := h.svc.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "user_not_found", "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, ToResponse(u))
}
