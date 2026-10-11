package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sukamoto/wedding-invitation/internal/config"
	"github.com/sukamoto/wedding-invitation/internal/middleware"
	"github.com/sukamoto/wedding-invitation/internal/models"
	"github.com/sukamoto/wedding-invitation/internal/response"
	"github.com/sukamoto/wedding-invitation/internal/services"
)

type Handler struct {
	svc *services.Service
	cfg *config.Config
}

func New(svc *services.Service, cfg *config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

// GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /api/v1/packages
func (h *Handler) ListPackages(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListActivePackages()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil daftar paket.")
		return
	}
	if list == nil {
		list = []models.Package{}
	}
	response.JSON(w, http.StatusOK, list)
}

// GET /api/v1/templates
func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	pkg := r.URL.Query().Get("package")
	list, err := h.svc.ListTemplates(pkg)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil daftar template.")
		return
	}
	if list == nil {
		list = []models.Template{}
	}
	// Public shape: hide internal paths already handled in model tags
	response.JSON(w, http.StatusOK, list)
}

// GET /api/v1/templates/{slug}
func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	t, err := h.svc.GetTemplateBySlug(slug)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "TEMPLATE_NOT_AVAILABLE", "Template tidak ditemukan.")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil template.")
		return
	}
	response.JSON(w, http.StatusOK, t)
}

// POST /api/v1/orders
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "JSON tidak valid.")
		return
	}

	resp, err := h.svc.CreateOrder(req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrPackageUnavailable):
			response.Error(w, http.StatusUnprocessableEntity, "PACKAGE_NOT_AVAILABLE", "Paket tidak tersedia.")
		case errors.Is(err, services.ErrTemplateUnavailable):
			response.Error(w, http.StatusUnprocessableEntity, "TEMPLATE_NOT_AVAILABLE", "Template tidak tersedia.")
		case errors.Is(err, services.ErrValidation):
			response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal membuat pesanan.")
		}
		return
	}
	response.JSON(w, http.StatusCreated, resp)
}

// GET /api/v1/orders/{reference}
func (h *Handler) GetOrderStatus(w http.ResponseWriter, r *http.Request) {
	ref := chi.URLParam(r, "reference")
	resp, err := h.svc.GetOrderByReference(ref)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Pesanan tidak ditemukan.")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil status pesanan.")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

// POST /api/v1/orders/{reference}/payment
func (h *Handler) SubmitPayment(w http.ResponseWriter, r *http.Request) {
	ref := chi.URLParam(r, "reference")

	if err := r.ParseMultipartForm(h.cfg.MaxUploadBytes + 1024); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "Form tidak valid atau file terlalu besar.")
		return
	}

	method := r.FormValue("method")
	if method == "" {
		method = h.cfg.PaymentMethod
	}

	file, header, err := r.FormFile("proof")
	if err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "PAYMENT_PROOF_INVALID", "Bukti pembayaran wajib diunggah.")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	err = h.svc.SubmitPaymentProof(ref, method, file, contentType)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotFound):
			response.Error(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Pesanan tidak ditemukan.")
		case errors.Is(err, services.ErrInvalidTransition):
			response.Error(w, http.StatusConflict, "INVALID_STATUS_TRANSITION", "Status pesanan tidak memungkinkan unggah bukti.")
		case errors.Is(err, services.ErrValidation):
			response.Error(w, http.StatusUnprocessableEntity, "PAYMENT_PROOF_INVALID", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal menyimpan bukti pembayaran.")
		}
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "payment_submitted",
	})
}

// GET /api/v1/invitations/{slug}
func (h *Handler) GetInvitation(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	inv, err := h.svc.GetPublishedInvitation(slug)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Undangan tidak ditemukan.")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil undangan.")
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"slug":          inv.Slug,
		"template_slug": inv.TemplateSlug,
		"published":     inv.Published,
		"invitation":    inv.Data,
	})
}

// GET /i/{slug}
func (h *Handler) RenderInvitation(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	inv, err := h.svc.GetPublishedInvitation(slug)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	template, err := h.svc.GetTemplateBySlug(inv.TemplateSlug)
	if err != nil {
		http.Error(w, "Invitation template unavailable", http.StatusInternalServerError)
		return
	}
	page, err := readTemplateSource(template.SourcePath)
	if err != nil {
		http.Error(w, "Invitation template unavailable", http.StatusInternalServerError)
		return
	}
	bodyStart := strings.Index(strings.ToLower(page), "<body")
	if bodyStart < 0 {
		http.Error(w, "Invitation template is invalid", http.StatusInternalServerError)
		return
	}
	bodyEnd := strings.Index(page[bodyStart:], ">")
	if bodyEnd < 0 {
		http.Error(w, "Invitation template is invalid", http.StatusInternalServerError)
		return
	}
	bodyEnd += bodyStart + 1
	// RawMessage is validated through a decode/re-encode cycle so user data cannot close
	// the script element or inject markup into the published page.
	var invitationData interface{}
	if err := json.Unmarshal(inv.Data, &invitationData); err != nil {
		http.Error(w, "Invitation data is invalid", http.StatusInternalServerError)
		return
	}
	invitationJSON, err := json.Marshal(invitationData)
	if err != nil {
		http.Error(w, "Invitation data is invalid", http.StatusInternalServerError)
		return
	}
	payload := `<script id="sukamoto-invitation-data" type="application/json" data-template-slug="` +
		htmlEscape(inv.TemplateSlug) + `">` + string(invitationJSON) + `</script>`
	page = page[:bodyEnd] + payload + page[bodyEnd:]

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(page))
}

func readTemplateSource(sourcePath string) (string, error) {
	sourcePath = filepath.ToSlash(filepath.Clean(sourcePath))
	sourcePath = strings.TrimPrefix(sourcePath, "templates/")
	if sourcePath == "" || sourcePath == "." || strings.HasPrefix(sourcePath, "../") || filepath.IsAbs(sourcePath) {
		return "", os.ErrInvalid
	}

	candidates := []string{
		filepath.Join("frontend", "templates", filepath.FromSlash(sourcePath)),
		filepath.Join("..", "frontend", "templates", filepath.FromSlash(sourcePath)),
	}
	if _, sourceFile, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Join(filepath.Dir(sourceFile), "..", "..", "..", "frontend", "templates", filepath.FromSlash(sourcePath)))
	}
	for _, candidate := range candidates {
		data, err := os.ReadFile(candidate)
		if err == nil {
			return string(data), nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	return "", os.ErrNotExist
}

func htmlEscape(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(s)
}

// ---------- Admin auth ----------

// POST /api/v1/admin/auth/login
func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "JSON tidak valid.")
		return
	}
	if req.Username == "" || req.Password == "" {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Username dan password wajib diisi.")
		return
	}

	admin, err := h.svc.AuthenticateAdmin(req.Username, req.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Username atau password salah.")
		return
	}

	sessionID, expires, err := h.svc.CreateSession(admin.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal membuat sesi.")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   r.TLS != nil, // true behind HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"username": admin.Username,
		"expires":  expires.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// POST /api/v1/admin/auth/logout
func (h *Handler) AdminLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.cfg.SessionCookieName)
	if err == nil && cookie.Value != "" {
		_ = h.svc.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	response.NoContent(w)
}

// GET /api/v1/admin/orders
func (h *Handler) AdminListOrders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	list, total, err := h.svc.ListOrders(status, page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil daftar pesanan.")
		return
	}
	if list == nil {
		list = []models.Order{}
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}
	response.JSONWithMeta(w, http.StatusOK, list, models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GET /api/v1/admin/orders/{id}
func (h *Handler) AdminGetOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID tidak valid.")
		return
	}
	order, payment, err := h.svc.GetOrderByID(id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Pesanan tidak ditemukan.")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil pesanan.")
		return
	}
	data := map[string]interface{}{
		"order": order,
	}
	if payment != nil {
		data["payment"] = payment
	}
	response.JSON(w, http.StatusOK, data)
}

// POST /api/v1/admin/orders/{id}/verify-payment
func (h *Handler) AdminVerifyPayment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID tidak valid.")
		return
	}
	var req models.VerifyPaymentRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "JSON tidak valid.")
		return
	}

	admin := middleware.GetAdmin(r.Context())
	err = h.svc.VerifyPayment(id, admin.ID, req.Decision, req.Note)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotFound):
			response.Error(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Pesanan tidak ditemukan.")
		case errors.Is(err, services.ErrInvalidTransition):
			response.Error(w, http.StatusConflict, "INVALID_STATUS_TRANSITION", err.Error())
		case errors.Is(err, services.ErrAlreadyVerified):
			response.Error(w, http.StatusConflict, "PAYMENT_ALREADY_VERIFIED", "Pembayaran sudah diverifikasi.")
		case errors.Is(err, services.ErrValidation):
			response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memverifikasi pembayaran.")
		}
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /api/v1/admin/orders/{id}/publish
func (h *Handler) AdminPublishOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID tidak valid.")
		return
	}
	var req models.PublishRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "JSON tidak valid.")
		return
	}

	admin := middleware.GetAdmin(r.Context())
	err = h.svc.PublishOrder(id, admin.ID, req.Slug)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotFound):
			response.Error(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Pesanan tidak ditemukan.")
		case errors.Is(err, services.ErrInvalidTransition):
			response.Error(w, http.StatusConflict, "INVALID_STATUS_TRANSITION", err.Error())
		case errors.Is(err, services.ErrSlugExists):
			response.Error(w, http.StatusConflict, "SLUG_ALREADY_EXISTS", "Slug sudah digunakan.")
		case errors.Is(err, services.ErrValidation):
			response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal menerbitkan undangan.")
		}
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "published",
		"slug":   req.Slug,
	})
}

// POST /api/v1/admin/orders/{id}/reject
func (h *Handler) AdminRejectOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID tidak valid.")
		return
	}
	var req models.RejectOrderRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&req); err != nil {
		// note is optional
		req.Note = ""
	}

	admin := middleware.GetAdmin(r.Context())
	err = h.svc.RejectOrder(id, admin.ID, req.Note)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotFound):
			response.Error(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Pesanan tidak ditemukan.")
		case errors.Is(err, services.ErrInvalidTransition):
			response.Error(w, http.StatusConflict, "INVALID_STATUS_TRANSITION", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal menolak pesanan.")
		}
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

// GET /api/v1/admin/templates
func (h *Handler) AdminListTemplates(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListAllTemplates()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil template.")
		return
	}
	if list == nil {
		list = []models.Template{}
	}
	response.JSON(w, http.StatusOK, list)
}

// PATCH /api/v1/admin/templates/{id}
func (h *Handler) AdminUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID tidak valid.")
		return
	}
	var body struct {
		Available *bool `json:"available"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "JSON tidak valid.")
		return
	}
	if body.Available == nil {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Field available wajib diisi.")
		return
	}
	if err := h.svc.UpdateTemplateAvailability(id, *body.Available); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "TEMPLATE_NOT_AVAILABLE", "Template tidak ditemukan.")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memperbarui template.")
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"id":        id,
		"available": *body.Available,
	})
}
