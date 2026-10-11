package services

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sukamoto/wedding-invitation/internal/config"
	"github.com/sukamoto/wedding-invitation/internal/models"
	"github.com/sukamoto/wedding-invitation/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrInvalidTransition   = errors.New("invalid status transition")
	ErrValidation          = errors.New("validation error")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrPackageUnavailable  = errors.New("package not available")
	ErrTemplateUnavailable = errors.New("template not available")
	ErrSlugExists          = errors.New("slug already exists")
	ErrAlreadyVerified     = errors.New("payment already verified")
)

type Service struct {
	db     *sql.DB
	cfg    *config.Config
	proofs *storage.ProofStore
	gallery *storage.GalleryStore
}

func New(db *sql.DB, cfg *config.Config, proofs *storage.ProofStore, gallery *storage.GalleryStore) *Service {
	return &Service{db: db, cfg: cfg, proofs: proofs, gallery: gallery}
}

type GalleryUpload struct {
	Reader io.Reader
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// ---------- Packages & Templates ----------

func (s *Service) ListActivePackages() ([]models.Package, error) {
	rows, err := s.db.Query(`
		SELECT id, slug, name, price, currency, features_json, active, created_at, updated_at
		FROM packages WHERE active = 1 ORDER BY price ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Package
	for rows.Next() {
		var p models.Package
		var featuresRaw string
		var activeInt int
		var created, updated string
		if err := rows.Scan(&p.ID, &p.Slug, &p.Name, &p.Price, &p.Currency, &featuresRaw, &activeInt, &created, &updated); err != nil {
			return nil, err
		}
		p.Active = activeInt == 1
		p.CreatedAt = parseTime(created)
		p.UpdatedAt = parseTime(updated)
		_ = json.Unmarshal([]byte(featuresRaw), &p.Features)
		if p.Features == nil {
			p.Features = []string{}
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (s *Service) ListTemplates(packageSlug string) ([]models.Template, error) {
	query := `
		SELECT id, slug, name, package_slug, source_path, thumbnail_path, demo_path, metadata_json, available, created_at, updated_at
		FROM templates WHERE available = 1`
	args := []interface{}{}
	if packageSlug != "" {
		query += ` AND package_slug = ?`
		args = append(args, packageSlug)
	}
	query += ` ORDER BY name ASC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Template
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *t)
	}
	return list, rows.Err()
}

func (s *Service) GetTemplateBySlug(slug string) (*models.Template, error) {
	row := s.db.QueryRow(`
		SELECT id, slug, name, package_slug, source_path, thumbnail_path, demo_path, metadata_json, available, created_at, updated_at
		FROM templates WHERE slug = ?`, slug)
	t, err := scanTemplate(row)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	// Parse supported_fields from metadata
	var meta map[string]interface{}
	if err := json.Unmarshal(t.MetadataJSON, &meta); err == nil {
		if fields, ok := meta["supported_fields"].([]interface{}); ok {
			for _, f := range fields {
				if str, ok := f.(string); ok {
					t.SupportedFields = append(t.SupportedFields, str)
				}
			}
		}
	}
	return t, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanTemplate(sc scanner) (*models.Template, error) {
	var t models.Template
	var metaRaw string
	var availableInt int
	var created, updated string
	err := sc.Scan(&t.ID, &t.Slug, &t.Name, &t.PackageSlug, &t.SourcePath, &t.ThumbnailPath, &t.DemoPath, &metaRaw, &availableInt, &created, &updated)
	if err != nil {
		return nil, err
	}
	t.Available = availableInt == 1
	t.MetadataJSON = json.RawMessage(metaRaw)
	t.CreatedAt = parseTime(created)
	t.UpdatedAt = parseTime(updated)
	return &t, nil
}

// ---------- Orders ----------

func generateReference() (string, string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	ref := "SKM-" + strings.ToUpper(hex.EncodeToString(b)[:12])
	hash := sha256.Sum256([]byte(ref))
	return ref, hex.EncodeToString(hash[:]), nil
}

func (s *Service) CreateOrder(req models.CreateOrderRequest) (*models.CreateOrderResponse, error) {
	// Basic validation
	if req.PackageSlug == "" || req.TemplateSlug == "" {
		return nil, fmt.Errorf("%w: package_slug and template_slug required", ErrValidation)
	}
	if req.Customer.Name == "" || req.Customer.Whatsapp == "" {
		return nil, fmt.Errorf("%w: customer name and whatsapp required", ErrValidation)
	}
	if req.IdempotencyKey == "" {
		return nil, fmt.Errorf("%w: idempotency_key required", ErrValidation)
	}

	// Idempotency check
	var existingRef string
	err := s.db.QueryRow(`SELECT reference FROM orders WHERE idempotency_key = ?`, req.IdempotencyKey).Scan(&existingRef)
	if err == nil {
		// Return existing order summary
		return s.buildCreateResponse(existingRef)
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Validate package
	var pkgID, pkgPrice int64
	var pkgActive int
	var pkgCurrency string
	err = s.db.QueryRow(`SELECT id, price, currency, active FROM packages WHERE slug = ?`, req.PackageSlug).
		Scan(&pkgID, &pkgPrice, &pkgCurrency, &pkgActive)
	if err == sql.ErrNoRows || pkgActive != 1 {
		return nil, ErrPackageUnavailable
	}
	if err != nil {
		return nil, err
	}

	// Validate template
	var tplID int64
	var tplPkg string
	var tplAvailable int
	err = s.db.QueryRow(`SELECT id, package_slug, available FROM templates WHERE slug = ?`, req.TemplateSlug).
		Scan(&tplID, &tplPkg, &tplAvailable)
	if err == sql.ErrNoRows || tplAvailable != 1 {
		return nil, ErrTemplateUnavailable
	}
	if err != nil {
		return nil, err
	}
	if tplPkg != req.PackageSlug {
		return nil, fmt.Errorf("%w: template does not belong to package", ErrValidation)
	}

	invJSON, err := json.Marshal(req.Invitation)
	if err != nil {
		return nil, err
	}

	ref, refHash, err := generateReference()
	if err != nil {
		return nil, err
	}

	ts := nowUTC()
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		INSERT INTO orders (
			reference, reference_hash, idempotency_key, package_id, template_id,
			customer_name, customer_whatsapp, invitation_data_json,
			total_amount, currency, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ref, refHash, req.IdempotencyKey, pkgID, tplID,
		req.Customer.Name, req.Customer.Whatsapp, string(invJSON),
		pkgPrice, pkgCurrency, models.OrderStatusPendingPayment, ts, ts)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			// Race on idempotency
			return s.buildCreateResponseByKey(req.IdempotencyKey)
		}
		return nil, err
	}

	orderID, _ := res.LastInsertId()

	// Create initial payment record
	_, err = tx.Exec(`
		INSERT INTO payments (order_id, method, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		orderID, s.cfg.PaymentMethod, models.PaymentStatusUnpaid, ts, ts)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.CreateOrderResponse{
		OrderReference: ref,
		Status:         models.OrderStatusPendingPayment,
		TotalAmount:    pkgPrice,
		Currency:       pkgCurrency,
		PaymentInstructions: models.PaymentInstructions{
			Method:      s.cfg.PaymentMethod,
			Description: s.cfg.PaymentDesc,
		},
	}, nil
}

func (s *Service) buildCreateResponse(ref string) (*models.CreateOrderResponse, error) {
	var status, currency string
	var amount int64
	err := s.db.QueryRow(`SELECT status, total_amount, currency FROM orders WHERE reference = ?`, ref).
		Scan(&status, &amount, &currency)
	if err != nil {
		return nil, err
	}
	return &models.CreateOrderResponse{
		OrderReference: ref,
		Status:         status,
		TotalAmount:    amount,
		Currency:       currency,
		PaymentInstructions: models.PaymentInstructions{
			Method:      s.cfg.PaymentMethod,
			Description: s.cfg.PaymentDesc,
		},
	}, nil
}

func (s *Service) buildCreateResponseByKey(key string) (*models.CreateOrderResponse, error) {
	var ref string
	err := s.db.QueryRow(`SELECT reference FROM orders WHERE idempotency_key = ?`, key).Scan(&ref)
	if err != nil {
		return nil, err
	}
	return s.buildCreateResponse(ref)
}

func (s *Service) GetOrderByReference(ref string) (*models.OrderStatusResponse, error) {
	var o models.Order
	var invID sql.NullInt64
	var invSlug sql.NullString
	var invPublished int
	var payStatus string

	err := s.db.QueryRow(`
		SELECT o.reference, o.status, o.total_amount, o.currency,
		       COALESCE(p.status, 'unpaid'),
		       i.id, i.slug, COALESCE(i.published, 0)
		FROM orders o
		LEFT JOIN payments p ON p.order_id = o.id
		LEFT JOIN invitations i ON i.order_id = o.id
		WHERE o.reference = ?
		ORDER BY p.id DESC LIMIT 1`, ref).
		Scan(&o.Reference, &o.Status, &o.TotalAmount, &o.Currency, &payStatus, &invID, &invSlug, &invPublished)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	resp := &models.OrderStatusResponse{
		OrderReference: o.Reference,
		Status:         o.Status,
		PaymentStatus:  payStatus,
		Published:      invPublished == 1,
	}
	if invPublished == 1 && invSlug.Valid {
		url := s.cfg.BaseURL + "/i/" + invSlug.String
		resp.InvitationURL = &url
	}
	return resp, nil
}

func (s *Service) SubmitPaymentProof(ref, method string, fileReader interface{ Read([]byte) (int, error) }, contentType string) error {
	var orderID int64
	var status string
	err := s.db.QueryRow(`SELECT id, status FROM orders WHERE reference = ?`, ref).Scan(&orderID, &status)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	if status != models.OrderStatusPendingPayment && status != models.OrderStatusPaymentSubmitted {
		return fmt.Errorf("%w: cannot submit proof in status %s", ErrInvalidTransition, status)
	}

	// Save file
	key, err := s.proofs.Save(fileReader, contentType)
	if err != nil {
		if strings.Contains(err.Error(), "too large") {
			return fmt.Errorf("%w: %s", ErrValidation, err.Error())
		}
		if strings.Contains(err.Error(), "unsupported") {
			return fmt.Errorf("%w: %s", ErrValidation, err.Error())
		}
		return err
	}

	ts := nowUTC()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update latest payment or insert
	var payID int64
	err = tx.QueryRow(`SELECT id FROM payments WHERE order_id = ? ORDER BY id DESC LIMIT 1`, orderID).Scan(&payID)
	if err == sql.ErrNoRows {
		_, err = tx.Exec(`
			INSERT INTO payments (order_id, method, proof_storage_key, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			orderID, method, key, models.PaymentStatusSubmitted, ts, ts)
	} else if err != nil {
		return err
	} else {
		_, err = tx.Exec(`
			UPDATE payments SET method = ?, proof_storage_key = ?, status = ?, updated_at = ?
			WHERE id = ?`, method, key, models.PaymentStatusSubmitted, ts, payID)
	}
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE orders SET status = ?, updated_at = ? WHERE id = ?`,
		models.OrderStatusPaymentSubmitted, ts, orderID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) UploadGalleryPhotos(ref string, uploads []GalleryUpload) ([]string, error) {
	if len(uploads) == 0 || len(uploads) > 8 {
		return nil, fmt.Errorf("%w: pilih 1 sampai 8 foto", ErrValidation)
	}
	var orderID int64
	var status string
	err := s.db.QueryRow(`SELECT id, status FROM orders WHERE reference = ?`, ref).Scan(&orderID, &status)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if status == models.OrderStatusPublished || status == models.OrderStatusRejected || status == models.OrderStatusCancelled {
		return nil, fmt.Errorf("%w: foto tidak dapat diubah pada status %s", ErrInvalidTransition, status)
	}

	keys := make([]string, 0, len(uploads))
	cleanupNew := func() {
		for _, key := range keys {
			if err := s.gallery.Remove(key); err != nil {
				fmt.Printf("remove uncommitted gallery photo %q: %v\n", key, err)
			}
		}
	}
	for _, upload := range uploads {
		key, err := s.gallery.Save(upload.Reader)
		if err != nil {
			cleanupNew()
			if strings.Contains(err.Error(), "too large") || strings.Contains(err.Error(), "empty") ||
				strings.Contains(err.Error(), "unsupported") || strings.Contains(err.Error(), "invalid") {
				return nil, fmt.Errorf("%w: %s", ErrValidation, err.Error())
			}
			return nil, err
		}
		keys = append(keys, key)
	}

	var previousKeys []string
	rows, err := s.db.Query(`SELECT storage_key FROM gallery_photos WHERE order_id = ?`, orderID)
	if err != nil {
		cleanupNew()
		return nil, err
	}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			rows.Close()
			cleanupNew()
			return nil, err
		}
		previousKeys = append(previousKeys, key)
	}
	if err := rows.Close(); err != nil {
		cleanupNew()
		return nil, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		cleanupNew()
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM gallery_photos WHERE order_id = ?`, orderID)
	if err != nil {
		cleanupNew()
		return nil, err
	}
	now := nowUTC()
	urls := make([]string, 0, len(keys))
	for _, key := range keys {
		_, err = tx.Exec(`INSERT INTO gallery_photos (storage_key, order_id, created_at) VALUES (?, ?, ?)`,
			key, orderID, now)
		if err != nil {
			cleanupNew()
			return nil, err
		}
		urls = append(urls, "/media/"+key)
	}

	var invitation map[string]interface{}
	var invitationJSON string
	if err := tx.QueryRow(`SELECT invitation_data_json FROM orders WHERE id = ?`, orderID).Scan(&invitationJSON); err != nil {
		cleanupNew()
		return nil, err
	}
	if err := json.Unmarshal([]byte(invitationJSON), &invitation); err != nil {
		cleanupNew()
		return nil, fmt.Errorf("decode invitation data: %w", err)
	}
	invitation["gallery"] = urls
	updatedJSON, err := json.Marshal(invitation)
	if err != nil {
		cleanupNew()
		return nil, fmt.Errorf("encode invitation data: %w", err)
	}
	if _, err := tx.Exec(`UPDATE orders SET invitation_data_json = ?, updated_at = ? WHERE id = ?`,
		string(updatedJSON), now, orderID); err != nil {
		cleanupNew()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		cleanupNew()
		return nil, err
	}
	for _, key := range previousKeys {
		if err := s.gallery.Remove(key); err != nil {
			fmt.Printf("remove replaced gallery photo %q: %v\n", key, err)
		}
	}
	return urls, nil
}

func (s *Service) PublishedPhotoKey(key string) (string, error) {
	var storageKey, publishedAt string
	err := s.db.QueryRow(`
		SELECT gp.storage_key, i.published_at
		FROM gallery_photos gp
		JOIN invitations i ON i.order_id = gp.order_id
		WHERE gp.storage_key = ? AND i.published = 1`, key).Scan(&storageKey, &publishedAt)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if !parseTime(publishedAt).AddDate(0, 0, s.cfg.InvitationRetentionDays).After(time.Now().UTC()) {
		return "", ErrNotFound
	}
	return storageKey, nil
}

func (s *Service) CleanupExpiredPhotos() error {
	cutoff := time.Now().UTC().AddDate(0, 0, -s.cfg.InvitationRetentionDays).Format(time.RFC3339)
	now := nowUTC()
	rows, err := s.db.Query(`
		SELECT storage_key FROM gallery_photos
		WHERE (expires_at IS NOT NULL AND expires_at <= ?)
		   OR (expires_at IS NULL AND created_at <= ?)`, now, cutoff)
	if err != nil {
		return err
	}
	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			rows.Close()
			return err
		}
		keys = append(keys, key)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, key := range keys {
		if _, err := tx.Exec(`DELETE FROM gallery_photos WHERE storage_key = ?`, key); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	for _, key := range keys {
		if err := s.gallery.Remove(key); err != nil {
			return fmt.Errorf("remove expired gallery photo %q: %w", key, err)
		}
	}
	return nil
}

// ---------- Admin auth ----------

func (s *Service) AuthenticateAdmin(username, password string) (*models.Admin, error) {
	var a models.Admin
	var hash string
	var activeInt int
	var created, updated string
	err := s.db.QueryRow(`
		SELECT id, username, password_hash, active, created_at, updated_at
		FROM admins WHERE username = ?`, username).
		Scan(&a.ID, &a.Username, &hash, &activeInt, &created, &updated)
	if err == sql.ErrNoRows {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}
	if activeInt != 1 {
		return nil, ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, ErrUnauthorized
	}
	a.Active = true
	a.CreatedAt = parseTime(created)
	a.UpdatedAt = parseTime(updated)
	return &a, nil
}

func (s *Service) CreateSession(adminID int64) (string, time.Time, error) {
	id := uuid.NewString()
	expires := time.Now().UTC().Add(s.cfg.SessionDuration)
	_, err := s.db.Exec(`
		INSERT INTO sessions (id, admin_id, expires_at, created_at)
		VALUES (?, ?, ?, ?)`,
		id, adminID, expires.Format(time.RFC3339), nowUTC())
	if err != nil {
		return "", time.Time{}, err
	}
	return id, expires, nil
}

func (s *Service) GetSession(sessionID string) (*models.Session, *models.Admin, error) {
	var sess models.Session
	var expiresStr, createdStr string
	err := s.db.QueryRow(`
		SELECT id, admin_id, expires_at, created_at FROM sessions WHERE id = ?`, sessionID).
		Scan(&sess.ID, &sess.AdminID, &expiresStr, &createdStr)
	if err == sql.ErrNoRows {
		return nil, nil, ErrUnauthorized
	}
	if err != nil {
		return nil, nil, err
	}
	sess.ExpiresAt = parseTime(expiresStr)
	sess.CreatedAt = parseTime(createdStr)
	if time.Now().UTC().After(sess.ExpiresAt) {
		_, _ = s.db.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
		return nil, nil, ErrUnauthorized
	}

	var a models.Admin
	var activeInt int
	var created, updated string
	err = s.db.QueryRow(`
		SELECT id, username, active, created_at, updated_at FROM admins WHERE id = ?`, sess.AdminID).
		Scan(&a.ID, &a.Username, &activeInt, &created, &updated)
	if err != nil || activeInt != 1 {
		return nil, nil, ErrUnauthorized
	}
	a.Active = true
	a.CreatedAt = parseTime(created)
	a.UpdatedAt = parseTime(updated)
	return &sess, &a, nil
}

func (s *Service) DeleteSession(sessionID string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
	return err
}

// ---------- Admin orders ----------

func (s *Service) ListOrders(status string, page, limit int) ([]models.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	where := "1=1"
	args := []interface{}{}
	if status != "" {
		where += " AND o.status = ?"
		args = append(args, status)
	}

	var total int
	countQ := `SELECT COUNT(1) FROM orders o WHERE ` + where
	if err := s.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT o.id, o.reference, o.customer_name, o.customer_whatsapp,
		       o.total_amount, o.currency, o.status, o.created_at, o.updated_at,
		       p.slug, t.slug
		FROM orders o
		JOIN packages p ON p.id = o.package_id
		JOIN templates t ON t.id = o.template_id
		WHERE ` + where + `
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []models.Order
	for rows.Next() {
		var o models.Order
		var created, updated string
		if err := rows.Scan(&o.ID, &o.Reference, &o.CustomerName, &o.CustomerWhatsapp,
			&o.TotalAmount, &o.Currency, &o.Status, &created, &updated,
			&o.PackageSlug, &o.TemplateSlug); err != nil {
			return nil, 0, err
		}
		o.CreatedAt = parseTime(created)
		o.UpdatedAt = parseTime(updated)
		list = append(list, o)
	}
	return list, total, rows.Err()
}

func (s *Service) GetOrderByID(id int64) (*models.Order, *models.Payment, error) {
	var o models.Order
	var invJSON string
	var created, updated string
	err := s.db.QueryRow(`
		SELECT o.id, o.reference, o.package_id, o.template_id,
		       o.customer_name, o.customer_whatsapp, o.invitation_data_json,
		       o.total_amount, o.currency, o.status, o.created_at, o.updated_at,
		       p.slug, t.slug
		FROM orders o
		JOIN packages p ON p.id = o.package_id
		JOIN templates t ON t.id = o.template_id
		WHERE o.id = ?`, id).
		Scan(&o.ID, &o.Reference, &o.PackageID, &o.TemplateID,
			&o.CustomerName, &o.CustomerWhatsapp, &invJSON,
			&o.TotalAmount, &o.Currency, &o.Status, &created, &updated,
			&o.PackageSlug, &o.TemplateSlug)
	if err == sql.ErrNoRows {
		return nil, nil, ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	o.InvitationDataJSON = json.RawMessage(invJSON)
	o.CreatedAt = parseTime(created)
	o.UpdatedAt = parseTime(updated)

	var pay models.Payment
	var proofKey sql.NullString
	var verifiedBy sql.NullInt64
	var verifiedAt, payCreated, payUpdated sql.NullString
	var rejReason sql.NullString
	err = s.db.QueryRow(`
		SELECT id, order_id, method, proof_storage_key, status, verified_by, verified_at, rejection_reason, created_at, updated_at
		FROM payments WHERE order_id = ? ORDER BY id DESC LIMIT 1`, id).
		Scan(&pay.ID, &pay.OrderID, &pay.Method, &proofKey, &pay.Status, &verifiedBy, &verifiedAt, &rejReason, &payCreated, &payUpdated)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}
	if err == nil {
		if proofKey.Valid {
			pay.ProofStorageKey = proofKey.String
		}
		if verifiedBy.Valid {
			pay.VerifiedBy = &verifiedBy.Int64
		}
		if verifiedAt.Valid {
			t := parseTime(verifiedAt.String)
			pay.VerifiedAt = &t
		}
		if rejReason.Valid {
			pay.RejectionReason = &rejReason.String
		}
		pay.CreatedAt = parseTime(payCreated.String)
		pay.UpdatedAt = parseTime(payUpdated.String)
		return &o, &pay, nil
	}
	return &o, nil, nil
}

func (s *Service) VerifyPayment(orderID int64, adminID int64, decision, note string) error {
	if decision != "approve" && decision != "reject" {
		return fmt.Errorf("%w: decision must be approve or reject", ErrValidation)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	err = tx.QueryRow(`SELECT status FROM orders WHERE id = ?`, orderID).Scan(&status)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != models.OrderStatusPaymentSubmitted {
		return fmt.Errorf("%w: expected payment_submitted, got %s", ErrInvalidTransition, status)
	}

	var payID int64
	var payStatus string
	err = tx.QueryRow(`SELECT id, status FROM payments WHERE order_id = ? ORDER BY id DESC LIMIT 1`, orderID).
		Scan(&payID, &payStatus)
	if err != nil {
		return err
	}
	if payStatus == models.PaymentStatusVerified {
		return ErrAlreadyVerified
	}

	ts := nowUTC()
	if decision == "approve" {
		_, err = tx.Exec(`
			UPDATE payments SET status = ?, verified_by = ?, verified_at = ?, updated_at = ?
			WHERE id = ?`, models.PaymentStatusVerified, adminID, ts, ts, payID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`UPDATE orders SET status = ?, updated_at = ? WHERE id = ?`,
			models.OrderStatusPaymentVerified, ts, orderID)
		if err != nil {
			return err
		}
	} else {
		_, err = tx.Exec(`
			UPDATE payments SET status = ?, verified_by = ?, verified_at = ?, rejection_reason = ?, updated_at = ?
			WHERE id = ?`, models.PaymentStatusRejected, adminID, ts, note, ts, payID)
		if err != nil {
			return err
		}
		// Keep order in payment_submitted or move to a rejected payment state; contract separates payment reject vs order reject
		_, err = tx.Exec(`UPDATE orders SET status = ?, updated_at = ? WHERE id = ?`,
			models.OrderStatusPendingPayment, ts, orderID)
		if err != nil {
			return err
		}
	}

	details, _ := json.Marshal(map[string]interface{}{
		"decision": decision,
		"note":     note,
	})
	_, err = tx.Exec(`
		INSERT INTO audit_logs (admin_id, action, entity_type, entity_id, details_json, created_at)
		VALUES (?, ?, 'order', ?, ?, ?)`,
		adminID, "verify_payment_"+decision, orderID, string(details), ts)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) PublishOrder(orderID int64, adminID int64, slug string) error {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug == "" || len(slug) < 3 {
		return fmt.Errorf("%w: slug must be at least 3 characters", ErrValidation)
	}
	// Basic slug validation
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return fmt.Errorf("%w: slug contains invalid characters", ErrValidation)
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	var templateID int64
	err = tx.QueryRow(`SELECT status, template_id FROM orders WHERE id = ?`, orderID).Scan(&status, &templateID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != models.OrderStatusPaymentVerified && status != models.OrderStatusInProgress {
		return fmt.Errorf("%w: order must be payment_verified or in_progress", ErrInvalidTransition)
	}

	// Check slug uniqueness
	var existing int
	err = tx.QueryRow(`SELECT COUNT(1) FROM invitations WHERE slug = ?`, slug).Scan(&existing)
	if err != nil {
		return err
	}
	if existing > 0 {
		return ErrSlugExists
	}

	ts := nowUTC()
	_, err = tx.Exec(`
		INSERT INTO invitations (order_id, slug, template_id, published, published_at, created_at, updated_at)
		VALUES (?, ?, ?, 1, ?, ?, ?)`,
		orderID, slug, templateID, ts, ts, ts)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return ErrSlugExists
		}
		return err
	}

	_, err = tx.Exec(`UPDATE orders SET status = ?, updated_at = ? WHERE id = ?`,
		models.OrderStatusPublished, ts, orderID)
	if err != nil {
		return err
	}

	details, _ := json.Marshal(map[string]string{"slug": slug})
	_, err = tx.Exec(`
		INSERT INTO audit_logs (admin_id, action, entity_type, entity_id, details_json, created_at)
		VALUES (?, 'publish', 'order', ?, ?, ?)`,
		adminID, orderID, string(details), ts)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) RejectOrder(orderID int64, adminID int64, note string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	err = tx.QueryRow(`SELECT status FROM orders WHERE id = ?`, orderID).Scan(&status)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status == models.OrderStatusPublished || status == models.OrderStatusCancelled {
		return fmt.Errorf("%w: cannot reject in status %s", ErrInvalidTransition, status)
	}

	ts := nowUTC()
	_, err = tx.Exec(`UPDATE orders SET status = ?, updated_at = ? WHERE id = ?`,
		models.OrderStatusRejected, ts, orderID)
	if err != nil {
		return err
	}

	details, _ := json.Marshal(map[string]string{"note": note})
	_, err = tx.Exec(`
		INSERT INTO audit_logs (admin_id, action, entity_type, entity_id, details_json, created_at)
		VALUES (?, 'reject_order', 'order', ?, ?, ?)`,
		adminID, orderID, string(details), ts)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ---------- Invitations (public) ----------

func (s *Service) GetPublishedInvitation(slug string) (*models.Invitation, error) {
	var inv models.Invitation
	var invJSON string
	var publishedInt int
	var publishedAt, created, updated sql.NullString
	var tplSlug string

	err := s.db.QueryRow(`
		SELECT i.id, i.order_id, i.slug, i.template_id, i.published, i.published_at,
		       i.created_at, i.updated_at, o.invitation_data_json, t.slug
		FROM invitations i
		JOIN orders o ON o.id = i.order_id
		JOIN templates t ON t.id = i.template_id
		WHERE i.slug = ? AND i.published = 1`, slug).
		Scan(&inv.ID, &inv.OrderID, &inv.Slug, &inv.TemplateID, &publishedInt, &publishedAt,
			&created, &updated, &invJSON, &tplSlug)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	inv.Published = true
	inv.TemplateSlug = tplSlug
	inv.Data = json.RawMessage(invJSON)
	if publishedAt.Valid {
		t := parseTime(publishedAt.String)
		inv.PublishedAt = &t
	}
	inv.CreatedAt = parseTime(created.String)
	inv.UpdatedAt = parseTime(updated.String)
	return &inv, nil
}

// ---------- Admin templates ----------

func (s *Service) ListAllTemplates() ([]models.Template, error) {
	rows, err := s.db.Query(`
		SELECT id, slug, name, package_slug, source_path, thumbnail_path, demo_path, metadata_json, available, created_at, updated_at
		FROM templates ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Template
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *t)
	}
	return list, rows.Err()
}

func (s *Service) UpdateTemplateAvailability(id int64, available bool) error {
	val := 0
	if available {
		val = 1
	}
	res, err := s.db.Exec(`UPDATE templates SET available = ?, updated_at = ? WHERE id = ?`,
		val, nowUTC(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
