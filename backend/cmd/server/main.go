package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sukamoto/wedding-invitation/internal/config"
	"github.com/sukamoto/wedding-invitation/internal/database"
	"github.com/sukamoto/wedding-invitation/internal/handlers"
	"github.com/sukamoto/wedding-invitation/internal/middleware"
	"github.com/sukamoto/wedding-invitation/internal/services"
	"github.com/sukamoto/wedding-invitation/internal/storage"
)

func main() {
	cfg := config.Load()

	// Ensure proof storage exists
	if err := os.MkdirAll(cfg.ProofStoragePath, 0o750); err != nil {
		log.Fatalf("proof storage: %v", err)
	}

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// Locate migrations relative to working directory or binary
	migrationsPath := findMigrations()
	if err := database.Migrate(db, migrationsPath); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(db, cfg); err != nil {
		log.Fatalf("seed: %v", err)
	}

	proofStore, err := storage.NewProofStore(cfg.ProofStoragePath, cfg.MaxUploadBytes)
	if err != nil {
		log.Fatalf("proof store: %v", err)
	}
	galleryStore, err := storage.NewGalleryStore(cfg.GalleryStoragePath, cfg.MaxUploadBytes)
	if err != nil {
		log.Fatalf("gallery store: %v", err)
	}

	svc := services.New(db, cfg, proofStore, galleryStore)
	h := handlers.New(svc, cfg)
	if err := svc.CleanupExpiredPhotos(); err != nil {
		log.Printf("gallery cleanup: %v", err)
	}
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := svc.CleanupExpiredPhotos(); err != nil {
				log.Printf("gallery cleanup: %v", err)
			}
		}
	}()

	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// CORS — only if frontend is on a different origin during development
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiters
	publicRL := middleware.NewRateLimiter(60, time.Minute) // 60 req/min
	authRL := middleware.NewRateLimiter(10, time.Minute)   // stricter for login
	orderRL := middleware.NewRateLimiter(20, time.Minute)  // order creation / status

	// Health (no rate limit needed)
	r.Get("/health", h.Health)

	// Public API
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(publicRL.Middleware)

		r.Get("/packages", h.ListPackages)
		r.Get("/templates", h.ListTemplates)
		r.Get("/templates/{slug}", h.GetTemplate)

		r.With(orderRL.Middleware).Post("/orders", h.CreateOrder)
		r.With(orderRL.Middleware).Get("/orders/{reference}", h.GetOrderStatus)
		r.With(orderRL.Middleware).Post("/orders/{reference}/payment", h.SubmitPayment)
		r.With(orderRL.Middleware).Post("/orders/{reference}/photos", h.UploadGalleryPhotos)

		r.Get("/invitations/{slug}", h.GetInvitation)

		// Admin auth (stricter rate limit)
		r.Route("/admin/auth", func(r chi.Router) {
			r.Use(authRL.Middleware)
			r.Use(middleware.CheckOrigin(cfg))
			r.Post("/login", h.AdminLogin)
			r.With(middleware.RequireAdmin(svc, cfg), middleware.RequireCSRF(svc, cfg)).
				Post("/logout", h.AdminLogout)
		})

		// Admin protected routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.RequireAdmin(svc, cfg))
			r.Use(middleware.RequireCSRF(svc, cfg))

			r.Get("/orders", h.AdminListOrders)
			r.Get("/orders/{id}", h.AdminGetOrder)
			r.Post("/orders/{id}/verify-payment", h.AdminVerifyPayment)
			r.Post("/orders/{id}/publish", h.AdminPublishOrder)
			r.Post("/orders/{id}/reject", h.AdminRejectOrder)

			r.Get("/templates", h.AdminListTemplates)
			r.Patch("/templates/{id}", h.AdminUpdateTemplate)
		})
	})

	// Public invitation HTML page
	r.Get("/i/{slug}", h.RenderInvitation)
	r.With(orderRL.Middleware).Get("/media/{key}", h.ServeInvitationPhoto)

	// Serve the original static HTML and the shared personalization runtime.
	templateDir, err := findTemplateRoot()
	if err != nil {
		log.Fatalf("template files: %v", err)
	}
	r.Handle("/demo/*", http.StripPrefix("/demo/", http.FileServer(http.Dir(templateDir))))
	r.Handle("/templates/*", http.StripPrefix("/templates/", http.FileServer(http.Dir(templateDir))))

	log.Printf("SUKAMOTO backend listening on %s", cfg.Addr)
	log.Printf("Database: %s", cfg.DatabasePath)
	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func findMigrations() string {
	candidates := []string{
		"migrations",
		"./migrations",
		filepath.Join("..", "..", "migrations"),
		filepath.Join("backend", "migrations"),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	// Fallback: relative to this source file location when running via go run
	return "migrations"
}

func findTemplateRoot() (string, error) {
	candidates := []string{
		filepath.Join("frontend", "templates"),
		filepath.Join("..", "frontend", "templates"),
	}
	if _, sourceFile, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Join(filepath.Dir(sourceFile), "..", "..", "..", "frontend", "templates"))
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			absolute, absErr := filepath.Abs(candidate)
			if absErr != nil {
				return "", absErr
			}
			return absolute, nil
		}
	}
	return "", os.ErrNotExist
}
