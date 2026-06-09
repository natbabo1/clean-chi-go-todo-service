package app

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/user/todo-list/internal/auth"
	"github.com/user/todo-list/internal/middleware"
	"github.com/user/todo-list/internal/platform/jwt"
	"github.com/user/todo-list/internal/platform/response"
	"github.com/user/todo-list/internal/todo"
	"github.com/user/todo-list/internal/user"
)

type handlers struct {
	auth *auth.Handler
	user *user.Handler
	todo *todo.Handler
}

func buildRouter(
	log *slog.Logger,
	jwtManager *jwt.Manager,
	corsOrigins []string,
	h handlers,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recover(log))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Swagger UI — served at /swagger/index.html, doc.json at /swagger/doc.json
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", h.auth.Register)
		r.Post("/auth/login", h.auth.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(jwtManager))
			r.Get("/me", h.user.GetMe)
			r.Route("/todos", func(r chi.Router) {
				r.Get("/", h.todo.List)
				r.Post("/", h.todo.Create)
				r.Get("/{id}", h.todo.GetByID)
				r.Patch("/{id}", h.todo.Update)
				r.Delete("/{id}", h.todo.Delete)
			})
		})
	})

	return r
}
