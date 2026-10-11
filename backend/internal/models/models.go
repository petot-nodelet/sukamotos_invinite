package models

import (
	"encoding/json"
	"time"
)

// Order statuses
const (
	OrderStatusPendingPayment   = "pending_payment"
	OrderStatusPaymentSubmitted = "payment_submitted"
	OrderStatusPaymentVerified  = "payment_verified"
	OrderStatusInProgress       = "in_progress"
	OrderStatusPublished        = "published"
	OrderStatusRejected         = "rejected"
	OrderStatusCancelled        = "cancelled"
)

// Payment statuses
const (
	PaymentStatusUnpaid    = "unpaid"
	PaymentStatusSubmitted = "submitted"
	PaymentStatusVerified  = "verified"
	PaymentStatusRejected  = "rejected"
)

type Package struct {
	ID           int64           `json:"id"`
	Slug         string          `json:"slug"`
	Name         string          `json:"name"`
	Price        int64           `json:"price"`
	Currency     string          `json:"currency"`
	FeaturesJSON json.RawMessage `json:"-"`
	Features     []string        `json:"features"`
	Active       bool            `json:"active"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type Template struct {
	ID              int64           `json:"id"`
	Slug            string          `json:"slug"`
	Name            string          `json:"name"`
	PackageSlug     string          `json:"package_slug"`
	SourcePath      string          `json:"-"`
	ThumbnailPath   string          `json:"thumbnail_url"`
	DemoPath        string          `json:"demo_url"`
	MetadataJSON    json.RawMessage `json:"-"`
	SupportedFields []string        `json:"supported_fields,omitempty"`
	Available       bool            `json:"available"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type Admin struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Order struct {
	ID                 int64           `json:"id"`
	Reference          string          `json:"order_reference"`
	ReferenceHash      string          `json:"-"`
	IdempotencyKey     string          `json:"-"`
	PackageID          int64           `json:"package_id,omitempty"`
	TemplateID         int64           `json:"template_id,omitempty"`
	CustomerName       string          `json:"customer_name,omitempty"`
	CustomerWhatsapp   string          `json:"customer_whatsapp,omitempty"`
	InvitationDataJSON json.RawMessage `json:"invitation,omitempty"`
	TotalAmount        int64           `json:"total_amount"`
	Currency           string          `json:"currency"`
	Status             string          `json:"status"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`

	// Joined / computed fields for responses
	PackageSlug   string  `json:"package_slug,omitempty"`
	TemplateSlug  string  `json:"template_slug,omitempty"`
	PaymentStatus string  `json:"payment_status,omitempty"`
	Published     bool    `json:"published,omitempty"`
	InvitationURL *string `json:"invitation_url,omitempty"`
}

type Payment struct {
	ID              int64      `json:"id"`
	OrderID         int64      `json:"order_id"`
	Method          string     `json:"method"`
	ProofStorageKey string     `json:"-"`
	Status          string     `json:"status"`
	VerifiedBy      *int64     `json:"verified_by,omitempty"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Invitation struct {
	ID          int64      `json:"id"`
	OrderID     int64      `json:"order_id"`
	Slug        string     `json:"slug"`
	TemplateID  int64      `json:"template_id"`
	Published   bool       `json:"published"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// For public response
	TemplateSlug string          `json:"template_slug,omitempty"`
	Data         json.RawMessage `json:"invitation,omitempty"`
}

type AuditLog struct {
	ID          int64           `json:"id"`
	AdminID     *int64          `json:"admin_id,omitempty"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entity_type"`
	EntityID    *int64          `json:"entity_id,omitempty"`
	DetailsJSON json.RawMessage `json:"details"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Session struct {
	ID        string    `json:"id"`
	AdminID   int64     `json:"admin_id"`
	CSRFToken string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Request / response DTOs

type CreateOrderRequest struct {
	PackageSlug    string                 `json:"package_slug"`
	TemplateSlug   string                 `json:"template_slug"`
	Customer       CustomerInfo           `json:"customer"`
	Invitation     map[string]interface{} `json:"invitation"`
	IdempotencyKey string                 `json:"idempotency_key"`
}

type CustomerInfo struct {
	Name     string `json:"name"`
	Whatsapp string `json:"whatsapp"`
}

type CreateOrderResponse struct {
	OrderReference      string              `json:"order_reference"`
	Status              string              `json:"status"`
	TotalAmount         int64               `json:"total_amount"`
	Currency            string              `json:"currency"`
	PaymentInstructions PaymentInstructions `json:"payment_instructions"`
}

type PaymentInstructions struct {
	Method      string `json:"method"`
	Description string `json:"description"`
}

type OrderStatusResponse struct {
	OrderReference string  `json:"order_reference"`
	Status         string  `json:"status"`
	PaymentStatus  string  `json:"payment_status"`
	Published      bool    `json:"published"`
	InvitationURL  *string `json:"invitation_url"`
}

type VerifyPaymentRequest struct {
	Decision string `json:"decision"` // approve | reject
	Note     string `json:"note"`
}

type PublishRequest struct {
	Slug string `json:"slug"`
}

type RejectOrderRequest struct {
	Note string `json:"note"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
