package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/pdrhlik/deliberix/server/config"
	"github.com/pdrhlik/deliberix/server/handler"
	"github.com/pdrhlik/deliberix/server/middleware"
	"github.com/pdrhlik/deliberix/server/notify"
	"github.com/pdrhlik/deliberix/server/store"
)

func main() {
	cfg := config.Load()

	s, err := store.New(cfg.DBDSN)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer s.DB.Close()

	notifyService := notify.NewService(s.DB)

	// Start background email sender if SMTP is configured
	if notify.SMTPConfigured(cfg.SMTPHost) {
		port := notify.ParseSMTPPort(cfg.SMTPPort)
		dialer := notify.NewDialer(cfg.SMTPHost, port, cfg.SMTPUser, cfg.SMTPPassword)
		sender := notify.NewGomailSender(dialer, cfg.SMTPFrom, "Deliberix")

		go func() {
			for {
				err := notifyService.SendEmails(context.Background(), sender)
				log.Printf("email sender stopped: %v, restarting...", err)
			}
		}()
		log.Println("email sender started")
	} else {
		log.Println("SMTP not configured, emails will be auto-verified")
		notifyService = nil
	}

	h := &handler.Handler{
		Store:  s,
		Config: cfg,
		Notify: notifyService,
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
	r.Post("/api/v1/auth/verify-email", handler.ErrorHandler(h.VerifyEmail()))
	r.Post("/api/v1/auth/magic-link", handler.ErrorHandler(h.RequestMagicLink()))
	r.Post("/api/v1/auth/magic-link/verify", handler.ErrorHandler(h.VerifyMagicLink()))
	r.Post("/api/v1/auth/forgot-password", handler.ErrorHandler(h.ForgotPassword()))
	r.Post("/api/v1/auth/reset-password", handler.ErrorHandler(h.ResetPassword()))

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret, s))
		r.Get("/api/v1/auth/me", handler.ErrorHandler(h.Me()))
		r.Patch("/api/v1/auth/me", handler.ErrorHandler(h.UpdateProfile()))
		r.Post("/api/v1/auth/change-password", handler.ErrorHandler(h.ChangePassword()))
		r.Post("/api/v1/auth/resend-verification", handler.ErrorHandler(h.ResendVerification()))

		// Survey routes
		r.Get("/api/v1/survey", handler.ErrorHandler(h.ListSurveys()))
		r.Post("/api/v1/survey", handler.ErrorHandler(h.CreateSurvey()))
		r.Get("/api/v1/survey/public", handler.ErrorHandler(h.ListPublicSurveys()))
		r.Get("/api/v1/survey/{slug}", handler.ErrorHandler(h.GetSurvey()))
		r.Patch("/api/v1/survey/{slug}", handler.ErrorHandler(h.UpdateSurvey()))
		r.Post("/api/v1/survey/{slug}/join", handler.ErrorHandler(h.JoinSurvey()))
		r.Get("/api/v1/survey/{slug}/participant/me", handler.ErrorHandler(h.GetMyParticipation()))

		// Participant management routes
		r.Get("/api/v1/survey/{slug}/participants", handler.ErrorHandler(h.ListParticipants()))
		r.Patch("/api/v1/survey/{slug}/participant/{userId}/role", handler.ErrorHandler(h.UpdateParticipantRole()))
		r.Delete("/api/v1/survey/{slug}/participant/{userId}", handler.ErrorHandler(h.RemoveParticipant()))

		// Statement routes
		r.Get("/api/v1/survey/{slug}/statement", handler.ErrorHandler(h.ListStatements()))
		r.Post("/api/v1/survey/{slug}/statement", handler.ErrorHandler(h.SubmitStatement()))
		r.Post("/api/v1/survey/{slug}/statement/seed", handler.ErrorHandler(h.AddSeedStatement()))
		r.Get("/api/v1/survey/{slug}/statement/next", handler.ErrorHandler(h.GetNextStatement()))

		// Moderation routes
		r.Get("/api/v1/survey/{slug}/moderation", handler.ErrorHandler(h.GetModerationQueue()))
		r.Patch("/api/v1/statement/{id}/moderate", handler.ErrorHandler(h.ModerateStatement()))

		// Results routes
		r.Get("/api/v1/survey/{slug}/results", handler.ErrorHandler(h.GetResults()))
		r.Get("/api/v1/survey/{slug}/stats", handler.ErrorHandler(h.GetSurveyStats()))

		// Response routes
		r.Post("/api/v1/statement/{id}/response", handler.ErrorHandler(h.SubmitResponse()))
		r.Get("/api/v1/survey/{slug}/progress", handler.ErrorHandler(h.GetVoteProgress()))
	})

	// Serve website static files if enabled
	if cfg.ServeWebsite {
		websiteDir := cfg.WebsitePath
		if _, err := os.Stat(websiteDir); err == nil {
			// Serve static assets (CSS, JS, images)
			fileServer := http.FileServer(http.Dir(websiteDir))

			// Website routes — serve pre-built HTML files
			websitePages := []string{"/", "/about", "/contact", "/cs", "/cs/about", "/cs/contact"}
			for _, page := range websitePages {
				pagePath := page
				r.Get(pagePath, func(w http.ResponseWriter, r *http.Request) {
					htmlFile := filepath.Join(websiteDir, pagePath, "index.html")
					if pagePath == "/" {
						htmlFile = filepath.Join(websiteDir, "index.html")
					}
					http.ServeFile(w, r, htmlFile)
				})
			}

			// Serve static assets under /_astro/ etc
			r.Handle("/*", fileServer)

			log.Printf("serving website from %s", websiteDir)
		} else {
			log.Printf("SERVE_WEBSITE=true but website directory not found: %s", websiteDir)
		}
	}

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
