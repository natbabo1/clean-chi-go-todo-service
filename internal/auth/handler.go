package auth

import (
	"encoding/json"
	"errors"
	"net/http"

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

// Register creates a new user account and returns an access token.
//
//	@Summary		Register
//	@Description	Create a new user account
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		auth.RegisterInput	true	"Register payload"
//	@Success		201		{object}	response.envelope{data=auth.TokenResponse}
//	@Failure		409		{object}	response.errorEnvelope
//	@Failure		422		{object}	response.validationErrorEnvelope
//	@Router			/auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if errs := h.validator.Validate(in); errs != nil {
		response.ValidationErrors(w, errs)
		return
	}

	res, err := h.svc.Register(r.Context(), in)
	if errors.Is(err, ErrEmailAlreadyExists) {
		response.Error(w, http.StatusConflict, "email_already_exists", "email already in use")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	response.JSON(w, http.StatusCreated, res)
}

// Login authenticates a user and returns an access token.
//
//	@Summary		Login
//	@Description	Authenticate with email and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		auth.LoginInput	true	"Login payload"
//	@Success		200		{object}	response.envelope{data=auth.TokenResponse}
//	@Failure		401		{object}	response.errorEnvelope
//	@Failure		422		{object}	response.validationErrorEnvelope
//	@Router			/auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in LoginInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if errs := h.validator.Validate(in); errs != nil {
		response.ValidationErrors(w, errs)
		return
	}

	res, err := h.svc.Login(r.Context(), in)
	if errors.Is(err, ErrInvalidCredentials) {
		response.Error(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, res)
}
