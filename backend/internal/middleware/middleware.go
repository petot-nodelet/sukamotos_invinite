package middleware

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/sukamoto/wedding-invitation/internal/config"
	"github.com/sukamoto/wedding-invitation/internal/models"
	"github.com/sukamoto/wedding-invitation/internal/response"
	"github.com/sukamoto/wedding-invitation/internal/services"
)

type contextKey string

const AdminContextKey contextKey = "admin"

func GetAdmin(ctx context.Context) *models.Admin {
	if a, ok := ctx.Value(AdminContextKey).(*models.Admin); ok {
		return a
	}
	return nil
}

// RequireAdmin validates the session cookie and injects the admin into context.
func RequireAdmin(svc *services.Service, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cfg.SessionCookieName)
			if err != nil || cookie.Value == "" {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak valid atau telah berakhir.")
				return
			}
			_, admin, err := svc.GetSession(cookie.Value)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak valid atau telah berakhir.")
				return
			}
			ctx := context.WithValue(r.Context(), AdminContextKey, admin)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// IssueCSRFCookie sets a readable double-submit CSRF cookie for an
// authenticated admin session. The frontend JS reads it and echoes the value
// in the X-CSRF-Token header on state-changing requests.
func IssueCSRFCookie(w http.ResponseWriter, cfg *config.Config) error {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.CSRFCookieName,
		Value:    hex.EncodeToString(buf),
		Path:     "/",
		HttpOnly: false, // must be readable by the admin frontend JS
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// ValidateCSRF enforces the double-submit pattern: unsafe methods from
// cookie-authenticated admins must carry an X-CSRF-Token header matching the
// CSRF cookie. Requests without a session cookie (e.g. public endpoints or
// curl usage) pass through; browser-based admin requests always send both.
func ValidateCSRF(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				next.ServeHTTP(w, r)
				return
			}
			sessionCookie, err := r.Cookie(cfg.SessionCookieName)
			if err != nil || sessionCookie.Value == "" {
				// Not a cookie-authenticated admin request; auth middleware
				// will reject it downstream if needed.
				next.ServeHTTP(w, r)
				return
			}
			csrfCookie, err := r.Cookie(cfg.CSRFCookieName)
			token := r.Header.Get("X-CSRF-Token")
			if err != nil || csrfCookie.Value == "" || token == "" ||
				subtle.ConstantTimeCompare([]byte(token), []byte(csrfCookie.Value)) != 1 {
				response.Error(w, http.StatusForbidden, "CSRF_VALIDATION_FAILED", "Token CSRF tidak valid. Muat ulang halaman admin dan coba lagi.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Simple in-memory rate limiter (per IP). Suitable for single-instance MVP.
type rateLimiter struct {
	mu      sync.Mutex
	visitors map[string]*visitor
	rate    int
	window  time.Duration
}

type visitor struct {
	count    int
	lastSeen time.Time
}

func NewRateLimiter(rate int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip = xff
		}

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if !exists || time.Since(v.lastSeen) > rl.window {
			rl.visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}
		v.count++
		v.lastSeen = time.Now()
		if v.count > rl.rate {
			rl.mu.Unlock()
			response.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "Terlalu banyak permintaan. Coba lagi nanti.")
			return
		}
		rl.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Lightweight request id for logs; not persisted
		w.Header().Set("X-Request-ID", time.Now().UTC().Format("20060102150405.000"))
		next.ServeHTTP(w, r)
	})
}

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan internal.")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
