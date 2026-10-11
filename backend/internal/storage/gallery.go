package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"bytes"
)

var allowedGalleryMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
}

type GalleryStore struct {
	basePath string
	maxBytes int64
}

func NewGalleryStore(basePath string, maxBytes int64) (*GalleryStore, error) {
	if err := os.MkdirAll(basePath, 0o750); err != nil {
		return nil, fmt.Errorf("create gallery dir: %w", err)
	}
	return &GalleryStore{basePath: basePath, maxBytes: maxBytes}, nil
}

func (s *GalleryStore) Save(r io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(r, s.maxBytes+1))
	if err != nil {
		return "", fmt.Errorf("read gallery image: %w", err)
	}
	if int64(len(data)) > s.maxBytes {
		return "", fmt.Errorf("gallery image too large")
	}
	if len(data) == 0 {
		return "", fmt.Errorf("gallery image is empty")
	}

	mimeType := http.DetectContentType(data)
	if strings.HasPrefix(mimeType, "image/jpeg") {
		mimeType = "image/jpeg"
	}
	ext, ok := allowedGalleryMIME[mimeType]
	if !ok {
		return "", fmt.Errorf("unsupported gallery image type")
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width < 1 || config.Height < 1 || int64(config.Width)*int64(config.Height) > 40000000 {
		return "", fmt.Errorf("invalid or oversized gallery image dimensions")
	}

	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate gallery image name: %w", err)
	}
	key := hex.EncodeToString(random) + ext
	if err := os.WriteFile(filepath.Join(s.basePath, key), data, 0o640); err != nil {
		return "", fmt.Errorf("write gallery image: %w", err)
	}
	return key, nil
}

func (s *GalleryStore) Open(key string) (*os.File, error) {
	if filepath.Base(key) != key || strings.Contains(key, "..") {
		return nil, fmt.Errorf("invalid gallery image key")
	}
	return os.Open(filepath.Join(s.basePath, key))
}

func (s *GalleryStore) Remove(key string) error {
	if filepath.Base(key) != key || strings.Contains(key, "..") {
		return fmt.Errorf("invalid gallery image key")
	}
	err := os.Remove(filepath.Join(s.basePath, key))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
