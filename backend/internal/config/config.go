package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr              string
	DatabasePath      string
	ProofStoragePath  string
	GalleryStoragePath string
	SessionCookieName string
	SessionDuration   time.Duration
	CSRFCookieName    string
	AdminUsername     string
	AdminPassword     string // only used for bootstrap; hashed on first run
	PaymentMethod     string
	PaymentDesc       string
	MaxUploadBytes    int64
	InvitationRetentionDays int
	AllowedOrigins    []string
	BaseURL           string // used for invitation_url construction
}

func Load() *Config {
	return &Config{
		Addr:              getEnv("ADDR", ":8080"),
		DatabasePath:      getEnv("DATABASE_PATH", "./data/sukamoto.db"),
		ProofStoragePath:  getEnv("PROOF_STORAGE_PATH", "./storage/proofs"),
		GalleryStoragePath: getEnv("GALLERY_STORAGE_PATH", "./storage/gallery"),
		SessionCookieName: getEnv("SESSION_COOKIE_NAME", "sukamoto_session"),
		SessionDuration:   time.Duration(getEnvInt("SESSION_HOURS", 24)) * time.Hour,
		CSRFCookieName:    getEnv("CSRF_COOKIE_NAME", "sukamoto_csrf"),
		AdminUsername:     getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:     getEnv("ADMIN_PASSWORD", "changeme"),
		PaymentMethod:     getEnv("PAYMENT_METHOD", "bank_transfer"),
		PaymentDesc:       getEnv("PAYMENT_DESC", "Transfer ke BCA 1234567890 a.n. SUKAMOTO. Cantumkan kode pesanan di berita transfer."),
		MaxUploadBytes:    int64(getEnvInt("MAX_UPLOAD_MB", 5)) * 1024 * 1024,
		InvitationRetentionDays: getEnvInt("INVITATION_RETENTION_DAYS", 365),
		AllowedOrigins:    []string{getEnv("ALLOWED_ORIGIN", "http://localhost:3000")},
		BaseURL:           getEnv("BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
