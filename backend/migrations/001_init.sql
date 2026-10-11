-- SUKAMOTO Wedding Invitation - Initial Schema
-- SQLite

PRAGMA foreign_keys = ON;

-- Packages
CREATE TABLE IF NOT EXISTS packages (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    slug        TEXT    NOT NULL UNIQUE,
    name        TEXT    NOT NULL,
    price       INTEGER NOT NULL DEFAULT 0, -- in smallest currency unit (IDR)
    currency    TEXT    NOT NULL DEFAULT 'IDR',
    features_json TEXT  NOT NULL DEFAULT '[]',
    active      INTEGER NOT NULL DEFAULT 1,
    created_at  TEXT    NOT NULL,
    updated_at  TEXT    NOT NULL
);

-- Templates
CREATE TABLE IF NOT EXISTS templates (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    slug            TEXT    NOT NULL UNIQUE,
    name            TEXT    NOT NULL,
    package_slug    TEXT    NOT NULL,
    source_path     TEXT    NOT NULL DEFAULT '',
    thumbnail_path  TEXT    NOT NULL DEFAULT '',
    demo_path       TEXT    NOT NULL DEFAULT '',
    metadata_json   TEXT    NOT NULL DEFAULT '{}',
    available       INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT    NOT NULL,
    updated_at      TEXT    NOT NULL,
    FOREIGN KEY (package_slug) REFERENCES packages(slug)
);

CREATE INDEX IF NOT EXISTS idx_templates_package ON templates(package_slug);
CREATE INDEX IF NOT EXISTS idx_templates_available ON templates(available);

-- Admins
CREATE TABLE IF NOT EXISTS admins (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT    NOT NULL UNIQUE,
    password_hash TEXT    NOT NULL,
    active        INTEGER NOT NULL DEFAULT 1,
    created_at    TEXT    NOT NULL,
    updated_at    TEXT    NOT NULL
);

-- Orders
CREATE TABLE IF NOT EXISTS orders (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    reference         TEXT    NOT NULL UNIQUE, -- public high-entropy reference
    reference_hash    TEXT    NOT NULL UNIQUE, -- for internal lookups if needed
    idempotency_key   TEXT    NOT NULL UNIQUE,
    package_id        INTEGER NOT NULL,
    template_id       INTEGER NOT NULL,
    customer_name     TEXT    NOT NULL,
    customer_whatsapp TEXT    NOT NULL,
    invitation_data_json TEXT NOT NULL DEFAULT '{}',
    total_amount      INTEGER NOT NULL,
    currency          TEXT    NOT NULL DEFAULT 'IDR',
    status            TEXT    NOT NULL DEFAULT 'pending_payment',
    created_at        TEXT    NOT NULL,
    updated_at        TEXT    NOT NULL,
    FOREIGN KEY (package_id) REFERENCES packages(id),
    FOREIGN KEY (template_id) REFERENCES templates(id)
);

CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created ON orders(created_at);

-- Payments
CREATE TABLE IF NOT EXISTS payments (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id          INTEGER NOT NULL,
    method            TEXT    NOT NULL,
    proof_storage_key TEXT,
    status            TEXT    NOT NULL DEFAULT 'unpaid',
    verified_by       INTEGER,
    verified_at       TEXT,
    rejection_reason  TEXT,
    created_at        TEXT    NOT NULL,
    updated_at        TEXT    NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (verified_by) REFERENCES admins(id)
);

CREATE INDEX IF NOT EXISTS idx_payments_order ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);

-- Invitations
CREATE TABLE IF NOT EXISTS invitations (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id      INTEGER NOT NULL UNIQUE,
    slug          TEXT    NOT NULL UNIQUE,
    template_id   INTEGER NOT NULL,
    published     INTEGER NOT NULL DEFAULT 0,
    published_at  TEXT,
    created_at    TEXT    NOT NULL,
    updated_at    TEXT    NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (template_id) REFERENCES templates(id)
);

CREATE INDEX IF NOT EXISTS idx_invitations_slug ON invitations(slug);
CREATE INDEX IF NOT EXISTS idx_invitations_published ON invitations(published);

-- Audit logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_id    INTEGER,
    action      TEXT    NOT NULL,
    entity_type TEXT    NOT NULL,
    entity_id   INTEGER,
    details_json TEXT   NOT NULL DEFAULT '{}',
    created_at  TEXT    NOT NULL,
    FOREIGN KEY (admin_id) REFERENCES admins(id)
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_admin ON audit_logs(admin_id);

-- Sessions (simple cookie-based admin sessions)
CREATE TABLE IF NOT EXISTS sessions (
    id          TEXT    PRIMARY KEY,
    admin_id    INTEGER NOT NULL,
    expires_at  TEXT    NOT NULL,
    created_at  TEXT    NOT NULL,
    FOREIGN KEY (admin_id) REFERENCES admins(id)
);

CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
