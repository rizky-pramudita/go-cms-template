package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/keeps-dev/go-cms-template/internal/handlers"
	"github.com/keeps-dev/go-cms-template/internal/middleware"
	"github.com/keeps-dev/go-cms-template/internal/repository"
	"github.com/keeps-dev/go-cms-template/internal/response"
)

func New(db *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Initialize repositories
	contentTypeRepo := repository.NewContentTypeRepository(db)
	contentPostRepo := repository.NewContentPostRepository(db)
	mediaRepo := repository.NewMediaRepository(db)
	tagRepo := repository.NewTagRepository(db)
	contactRepo := repository.NewContactRepository(db)
	settingRepo := repository.NewSettingRepository(db)

	// Initialize handlers
	contentTypeHandler := handlers.NewContentTypeHandler(contentTypeRepo)
	contentPostHandler := handlers.NewContentPostHandler(contentPostRepo)
	mediaHandler := handlers.NewMediaHandler(mediaRepo)
	tagHandler := handlers.NewTagHandler(tagRepo)
	contactHandler := handlers.NewContactHandler(contactRepo)
	settingHandler := handlers.NewSettingHandler(settingRepo)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, map[string]string{"status": "healthy"})
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Content Types
		r.Route("/content-types", func(r chi.Router) {
			r.Get("/", contentTypeHandler.List)
			r.Post("/", contentTypeHandler.Create)
			r.Get("/slug/{slug}", contentTypeHandler.GetBySlug)
			r.Get("/{id}", contentTypeHandler.Get)
			r.Put("/{id}", contentTypeHandler.Update)
			r.Delete("/{id}", contentTypeHandler.Delete)
		})

		// Posts
		r.Route("/posts", func(r chi.Router) {
			r.Get("/", contentPostHandler.List)
			r.Post("/", contentPostHandler.Create)
			r.Get("/slug/{slug}", contentPostHandler.GetBySlug)
			r.Get("/{id}", contentPostHandler.Get)
			r.Put("/{id}", contentPostHandler.Update)
			r.Delete("/{id}", contentPostHandler.Delete)
			// Post media management
			r.Post("/{id}/media", contentPostHandler.AttachMedia)
			r.Delete("/{id}/media/{mediaId}", contentPostHandler.DetachMedia)
		})

		// Media
		r.Route("/media", func(r chi.Router) {
			r.Get("/", mediaHandler.List)
			r.Post("/", mediaHandler.Create)
			r.Get("/{id}", mediaHandler.Get)
			r.Put("/{id}", mediaHandler.Update)
			r.Delete("/{id}", mediaHandler.Delete)
		})

		// Tags
		r.Route("/tags", func(r chi.Router) {
			r.Get("/", tagHandler.List)
			r.Post("/", tagHandler.Create)
			r.Get("/slug/{slug}", tagHandler.GetBySlug)
			r.Get("/{id}", tagHandler.Get)
			r.Put("/{id}", tagHandler.Update)
			r.Delete("/{id}", tagHandler.Delete)
		})

		// Contact Submissions
		r.Route("/contacts", func(r chi.Router) {
			r.Get("/", contactHandler.List)
			r.Post("/", contactHandler.Create)
			r.Get("/unread-count", contactHandler.GetUnreadCount)
			r.Get("/{id}", contactHandler.Get)
			r.Put("/{id}", contactHandler.Update)
			r.Delete("/{id}", contactHandler.Delete)
		})

		// Settings
		r.Route("/settings", func(r chi.Router) {
			r.Get("/", settingHandler.List)
			r.Post("/", settingHandler.Create)
			r.Post("/upsert", settingHandler.Upsert)
			r.Post("/bulk", settingHandler.GetMultiple)
			r.Get("/{key}", settingHandler.Get)
			r.Put("/{key}", settingHandler.Update)
			r.Delete("/{key}", settingHandler.Delete)
		})
	})

	// 404 handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.NotFound(w, "Endpoint not found")
	})

	// 405 handler
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	})

	return r
}
