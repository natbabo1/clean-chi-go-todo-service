package todo

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/user/todo-list/internal/middleware"
	"github.com/user/todo-list/internal/platform/response"
	"github.com/user/todo-list/internal/platform/validator"
)

type Handler struct {
	svc       Service
	validator *validator.Validator
}

func NewHandler(svc Service, v *validator.Validator) *Handler {
	return &Handler{svc: svc, validator: v}
}

// List returns a paginated list of todos for the authenticated user.
//
//	@Summary		List todos
//	@Description	Get a paginated list of todos belonging to the authenticated user
//	@Tags			todos
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			limit		query		int		false	"Items per page"	default(20)
//	@Param			completed	query		bool	false	"Filter by completed status"
//	@Success		200			{object}	response.listEnvelope{data=[]todo.Response}
//	@Failure		401			{object}	response.errorEnvelope
//	@Router			/todos [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 20)
	if limit > 100 {
		limit = 100
	}

	params := ListParams{UserID: userID, Page: page, Limit: limit}
	if c := r.URL.Query().Get("completed"); c != "" {
		b, err := strconv.ParseBool(c)
		if err == nil {
			params.Completed = &b
		}
	}

	todos, total, err := h.svc.List(r.Context(), params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}

	out := make([]Response, len(todos))
	for i := range todos {
		out[i] = ToResponse(&todos[i])
	}
	response.JSONList(w, http.StatusOK, out, response.Pagination{Page: page, Limit: limit, Total: total})
}

// GetByID returns a single todo.
//
//	@Summary		Get todo
//	@Description	Get a single todo by ID
//	@Tags			todos
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Todo ID"
//	@Success		200	{object}	response.envelope{data=todo.Response}
//	@Failure		401	{object}	response.errorEnvelope
//	@Failure		404	{object}	response.errorEnvelope
//	@Router			/todos/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, todoID, ok := extractIDs(w, r)
	if !ok {
		return
	}
	t, err := h.svc.GetByID(r.Context(), userID, todoID)
	if errors.Is(err, ErrNotFound) {
		response.Error(w, http.StatusNotFound, "todo_not_found", "todo not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, ToResponse(t))
}

// Create creates a new todo.
//
//	@Summary		Create todo
//	@Description	Create a new todo for the authenticated user
//	@Tags			todos
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		todo.CreateInput	true	"Todo payload"
//	@Success		201		{object}	response.envelope{data=todo.Response}
//	@Failure		401		{object}	response.errorEnvelope
//	@Failure		422		{object}	response.validationErrorEnvelope
//	@Router			/todos [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	var in CreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if errs := h.validator.Validate(in); errs != nil {
		response.ValidationErrors(w, errs)
		return
	}

	t, err := h.svc.Create(r.Context(), userID, in)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	response.JSON(w, http.StatusCreated, ToResponse(t))
}

// Update partially updates a todo.
//
//	@Summary		Update todo
//	@Description	Partially update a todo
//	@Tags			todos
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Todo ID"
//	@Param			body	body		todo.UpdateInput	true	"Update payload"
//	@Success		200		{object}	response.envelope{data=todo.Response}
//	@Failure		401		{object}	response.errorEnvelope
//	@Failure		404		{object}	response.errorEnvelope
//	@Router			/todos/{id} [patch]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, todoID, ok := extractIDs(w, r)
	if !ok {
		return
	}

	var in UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}

	t, err := h.svc.Update(r.Context(), userID, todoID, in)
	if errors.Is(err, ErrNotFound) {
		response.Error(w, http.StatusNotFound, "todo_not_found", "todo not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, ToResponse(t))
}

// Delete hard-deletes a todo.
//
//	@Summary		Delete todo
//	@Description	Permanently delete a todo
//	@Tags			todos
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Todo ID"
//	@Success		204	{object}	nil
//	@Failure		401	{object}	response.errorEnvelope
//	@Failure		404	{object}	response.errorEnvelope
//	@Router			/todos/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, todoID, ok := extractIDs(w, r)
	if !ok {
		return
	}

	err := h.svc.Delete(r.Context(), userID, todoID)
	if errors.Is(err, ErrNotFound) {
		response.Error(w, http.StatusNotFound, "todo_not_found", "todo not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func extractIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return uuid.Nil, uuid.Nil, false
	}
	todoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid todo id")
		return uuid.Nil, uuid.Nil, false
	}
	return userID, todoID, true
}

func queryInt(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}
