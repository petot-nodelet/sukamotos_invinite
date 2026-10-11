CREATE TABLE IF NOT EXISTS gallery_photos (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    storage_key TEXT    NOT NULL UNIQUE,
    order_id    INTEGER NOT NULL,
    expires_at  TEXT,
    created_at  TEXT    NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_gallery_photos_order ON gallery_photos(order_id);
CREATE INDEX IF NOT EXISTS idx_gallery_photos_expiry ON gallery_photos(expires_at);
