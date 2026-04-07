package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/pdrhlik/edemos/server/config"
	"github.com/pdrhlik/edemos/server/handler"
	"github.com/pdrhlik/edemos/server/middleware"
	"github.com/pdrhlik/edemos/server/store"
)

func main() {
	cfg := config.Load()

	s, err := store.New(cfg.DBDSN)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer s.DB.Close()

	h := &handler.Handler{
		Store:  s,
		Config: cfg,
	}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Public auth routes
	r.Post("/api/v1/auth/register", handler.ErrorHandler(h.Register()))
	r.Post("/api/v1/auth/login", handler.ErrorHandler(h.Login()))

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret, s))
		r.Get("/api/v1/auth/me", handler.ErrorHandler(h.Me()))

		// Survey routes
		r.Get("/api/v1/survey", handler.ErrorHandler(h.ListSurveys()))
		r.Post("/api/v1/survey", handler.ErrorHandler(h.CreateSurvey()))
		r.Get("/api/v1/survey/public", handler.ErrorHandler(h.ListPublicSurveys()))
		r.Get("/api/v1/survey/{id}", handler.ErrorHandler(h.GetSurvey()))
		r.Patch("/api/v1/survey/{id}", handler.ErrorHandler(h.UpdateSurvey()))
		r.Post("/api/v1/survey/{id}/join", handler.ErrorHandler(h.JoinSurvey()))
		r.Get("/api/v1/survey/{id}/participant/me", handler.ErrorHandler(h.GetMyParticipation()))

		// Statement routes
		r.Get("/api/v1/survey/{id}/statement", handler.ErrorHandler(h.ListStatements()))
		r.Post("/api/v1/survey/{id}/statement", handler.ErrorHandler(h.SubmitStatement()))
		r.Post("/api/v1/survey/{id}/statement/seed", handler.ErrorHandler(h.AddSeedStatement()))
		r.Get("/api/v1/survey/{id}/statement/next", handler.ErrorHandler(h.GetNextStatement()))

		// Moderation routes
		r.Get("/api/v1/survey/{id}/moderation", handler.ErrorHandler(h.GetModerationQueue()))
		r.Patch("/api/v1/statement/{id}/moderate", handler.ErrorHandler(h.ModerateStatement()))

		// Results routes
		r.Get("/api/v1/survey/{id}/results", handler.ErrorHandler(h.GetResults()))
		r.Get("/api/v1/survey/{id}/stats", handler.ErrorHandler(h.GetSurveyStats()))

		// Response routes
		r.Post("/api/v1/statement/{id}/response", handler.ErrorHandler(h.SubmitResponse()))
		r.Get("/api/v1/survey/{id}/progress", handler.ErrorHandler(h.GetVoteProgress()))
	})

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
