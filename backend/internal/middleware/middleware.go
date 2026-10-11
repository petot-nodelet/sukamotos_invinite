package middleware

import (
	"context"
	"crypto/subtle"
	"net/http"
	"net/url"
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

// CheckOrigin rejects browser requests from origins outside the configured frontend
// origins and the request's own host. Requests without an Origin header remain usable
// by non-browser clients; state-changing admin requests are additionally token-checked.
func CheckOrigin(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !originAllowed(r, cfg.AllowedOrigins) {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "Origin permintaan tidak diizinkan.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireCSRF requires a session-bound double-submit token on unsafe methods.
func RequireCSRF(svc *services.Service, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isSafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			if !originAllowed(r, cfg.AllowedOrigins) {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "Origin permintaan tidak diizinkan.")
				return
			}
			sessionCookie, err := r.Cookie(cfg.SessionCookieName)
			if err != nil || sessionCookie.Value == "" {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak valid atau telah berakhir.")
				return
			}
			session, _, err := svc.GetSession(sessionCookie.Value)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak valid atau telah berakhir.")
				return
			}

			csrfCookie, err := r.Cookie(cfg.CSRFCookieName)
			headerToken := r.Header.Get("X-CSRF-Token")
			if err != nil || csrfCookie.Value == "" || headerToken == "" ||
				len(headerToken) != len(csrfCookie.Value) ||
				subtle.ConstantTimeCompare([]byte(headerToken), []byte(csrfCookie.Value)) != 1 ||
				len(session.CSRFToken) != len(headerToken) ||
				subtle.ConstantTimeCompare([]byte(headerToken), []byte(session.CSRFToken)) != 1 {
				response.Error(w, http.StatusForbidden, "CSRF_INVALID", "Token keamanan permintaan tidak valid.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func originAllowed(r *http.Request, allowedOrigins []string) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil ||
		parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	if parsed.Host == r.Host {
		return parsed.Scheme == "http" || parsed.Scheme == "https"
	}
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// Simple in-memory rate limiter (per IP). Suitable for single-instance MVP.
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int
	window   time.Duration
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
