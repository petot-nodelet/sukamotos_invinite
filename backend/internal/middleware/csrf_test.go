package middleware

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/sukamoto/wedding-invitation/internal/config"
	"github.com/sukamoto/wedding-invitation/internal/services"
	_ "modernc.org/sqlite"
)

func TestRequireCSRFValidatesSessionBoundToken(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "csrf.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`
		CREATE TABLE admins (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL,
			active INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			admin_id INTEGER NOT NULL,
			csrf_token TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		INSERT INTO admins VALUES (1, 'admin', 1, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z');
		INSERT INTO sessions VALUES ('session-id', 1, 'session-csrf-token', '2099-01-01T00:00:00Z', '2026-01-01T00:00:00Z');
	`)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		SessionCookieName: "session",
		CSRFCookieName:    "csrf",
		AllowedOrigins:    []string{"http://localhost:3000"},
	}
	svc := services.New(db, cfg, nil, nil)
	nextCalled := false
	handler := RequireCSRF(svc, cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	}))

	validRequest := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/admin/action", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.AddCookie(&http.Cookie{Name: "session", Value: "session-id"})
		req.AddCookie(&http.Cookie{Name: "csrf", Value: "session-csrf-token"})
		req.Header.Set("X-CSRF-Token", "session-csrf-token")
		return req
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, validRequest())
	if recorder.Code != http.StatusNoContent || !nextCalled {
		t.Fatalf("valid token was rejected: status=%d nextCalled=%v", recorder.Code, nextCalled)
	}

	nextCalled = false
	recorder = httptest.NewRecorder()
	req := validRequest()
	req.Header.Set("X-CSRF-Token", "attacker-token")
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusForbidden || nextCalled {
		t.Fatalf("invalid token was accepted: status=%d nextCalled=%v", recorder.Code, nextCalled)
	}

	nextCalled = false
	recorder = httptest.NewRecorder()
	req = validRequest()
	req.Header.Set("Origin", "https://attacker.example")
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusForbidden || nextCalled {
		t.Fatalf("untrusted origin was accepted: status=%d nextCalled=%v", recorder.Code, nextCalled)
	}
}
