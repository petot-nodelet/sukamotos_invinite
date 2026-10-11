package handlers

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sukamoto/wedding-invitation/internal/config"
	"github.com/sukamoto/wedding-invitation/internal/services"
	"github.com/sukamoto/wedding-invitation/internal/storage"
	_ "modernc.org/sqlite"
)

func TestReadTemplateSourceLoadsAllRegisteredOriginalTemplates(t *testing.T) {
	sources := []string{
		"templates/SILVER/template1_classic_matcha.html",
		"templates/SILVER/template4-burgundy-minimalist.html",
		"templates/SILVER/template8_islamic_traditional.html",
		"templates/GOLD/dusty_rose_wedding.html",
		"templates/GOLD/black-champagne-gold-ivory.html",
		"templates/PLATINUM/3_Imperial-Modern.html",
	}

	for _, source := range sources {
		t.Run(source, func(t *testing.T) {
			page, err := readTemplateSource(source)
			if err != nil {
				t.Fatalf("readTemplateSource(%q): %v", source, err)
			}
			if !strings.Contains(page, "<html") || !strings.Contains(page, "/templates/personalize.js") {
				t.Fatalf("template %q is missing HTML or its personalization runtime", source)
			}
		})
	}
}

func TestReadTemplateSourceRejectsPathTraversal(t *testing.T) {
	if _, err := readTemplateSource("templates/../../etc/passwd"); err == nil {
		t.Fatal("readTemplateSource accepted a path outside the template directory")
	}
}

func TestGalleryUploadAndMediaRequirePublishedInvitation(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "gallery.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY,
			reference TEXT NOT NULL,
			status TEXT NOT NULL,
			invitation_data_json TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE gallery_photos (
			id INTEGER PRIMARY KEY,
			storage_key TEXT NOT NULL,
			order_id INTEGER NOT NULL,
			expires_at TEXT,
			created_at TEXT NOT NULL
		);
		CREATE TABLE invitations (
			order_id INTEGER NOT NULL,
			published INTEGER NOT NULL,
			published_at TEXT NOT NULL
		);
		INSERT INTO orders VALUES (1, 'SKM-TEST', 'pending_payment', '{}', '2026-01-01T00:00:00Z');
	`)
	if err != nil {
		t.Fatal(err)
	}
	galleryPath := filepath.Join(t.TempDir(), "photos")
	galleryStore, err := storage.NewGalleryStore(galleryPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{InvitationRetentionDays: 365}
	svc := services.New(db, cfg, nil, galleryStore)
	handler := New(svc, cfg)
	router := chi.NewRouter()
	router.Post("/api/v1/orders/{reference}/photos", handler.UploadGalleryPhotos)
	router.Get("/media/{key}", handler.ServeInvitationPhoto)

	imageBytes, err := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p7sAAAAASUVORK5CYII=",
	)
	if err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	filePart, err := writer.CreateFormFile("photos", "gallery.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := filePart.Write(imageBytes); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/SKM-TEST/photos", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("upload returned %d: %s", recorder.Code, recorder.Body.String())
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			Photos []string `json:"photos"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if !envelope.Success || len(envelope.Data.Photos) != 1 {
		t.Fatalf("unexpected upload response: %+v", envelope)
	}
	key := filepath.Base(envelope.Data.Photos[0])

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/media/"+key, nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("draft media returned %d; expected 404", recorder.Code)
	}

	_, err = db.Exec(`INSERT INTO invitations VALUES (1, 1, ?)`, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatal(err)
	}
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/media/"+key, nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("published media returned %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("media Content-Type = %q, want image/png", got)
	}
	if got := recorder.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("media nosniff header = %q", got)
	}
	if !bytes.Equal(recorder.Body.Bytes(), imageBytes) {
		t.Fatalf("served media does not match uploaded file: %s", fmt.Sprintf("%x", recorder.Body.Bytes()))
	}

	if _, err := db.Exec(`UPDATE gallery_photos SET expires_at = '2000-01-01T00:00:00Z' WHERE storage_key = ?`, key); err != nil {
		t.Fatal(err)
	}
	if err := svc.CleanupExpiredPhotos(); err != nil {
		t.Fatalf("clean expired gallery photos: %v", err)
	}
	if _, err := os.Stat(filepath.Join(galleryPath, key)); !os.IsNotExist(err) {
		t.Fatalf("expired photo still exists or stat failed: %v", err)
	}
	var remaining int
	if err := db.QueryRow(`SELECT COUNT(1) FROM gallery_photos WHERE storage_key = ?`, key).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatalf("expired photo database record remains: %d", remaining)
	}
}
