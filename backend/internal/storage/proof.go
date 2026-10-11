package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Allowed MIME types for payment proofs
var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"application/pdf": ".pdf",
}

type ProofStore struct {
	basePath string
	maxBytes int64
}

func NewProofStore(basePath string, maxBytes int64) (*ProofStore, error) {
	if err := os.MkdirAll(basePath, 0o750); err != nil {
		return nil, fmt.Errorf("create proof dir: %w", err)
	}
	return &ProofStore{basePath: basePath, maxBytes: maxBytes}, nil
}

// Save validates content type by sniffing and stores the file under a random name.
// Returns the storage key (relative filename).
func (s *ProofStore) Save(r io.Reader, declaredMIME string) (string, error) {
	// Read limited content
	limited := io.LimitReader(r, s.maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("read upload: %w", err)
	}
	if int64(len(data)) > s.maxBytes {
		return "", fmt.Errorf("file too large")
	}
	if len(data) == 0 {
		return "", fmt.Errorf("empty file")
	}

	// Sniff real content type
	sniffed := http.DetectContentType(data)
	// Normalize
	if strings.HasPrefix(sniffed, "image/jpeg") {
		sniffed = "image/jpeg"
	}

	ext, ok := allowedMIME[sniffed]
	if !ok {
		// Fallback: try declared if sniff failed to a known type
		if declaredMIME != "" {
			if e, ok2 := allowedMIME[declaredMIME]; ok2 {
				ext = e
				sniffed = declaredMIME
			} else {
				return "", fmt.Errorf("unsupported file type: %s", sniffed)
			}
		} else {
			return "", fmt.Errorf("unsupported file type: %s", sniffed)
		}
	}

	// Random filename
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return "", fmt.Errorf("generate name: %w", err)
	}
	name := hex.EncodeToString(randBytes) + ext
	fullPath := filepath.Join(s.basePath, name)

	if err := os.WriteFile(fullPath, data, 0o640); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return name, nil
}

// Open returns a read-only handle for admin download. Caller must close.
func (s *ProofStore) Open(key string) (*os.File, error) {
	// Prevent path traversal
	clean := filepath.Base(key)
	if clean != key || strings.Contains(key, "..") {
		return nil, fmt.Errorf("invalid key")
	}
	full := filepath.Join(s.basePath, clean)
	return os.Open(full)
}

// ContentTypeFromKey returns a reasonable Content-Type for the stored file.
func ContentTypeFromKey(key string) string {
	ext := filepath.Ext(key)
	ct := mime.TypeByExtension(ext)
	if ct == "" {
		return "application/octet-stream"
	}
	return ct
}
