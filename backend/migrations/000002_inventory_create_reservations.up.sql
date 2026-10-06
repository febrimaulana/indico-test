-- 000002_inventory_create_reservations: reservasi ber-expiry 5 menit.
CREATE TABLE IF NOT EXISTS reservations (
    reservation_id TEXT PRIMARY KEY,
    user_id        TEXT        NOT NULL,
    item_id        TEXT        NOT NULL REFERENCES items (item_id),
    quantity       INTEGER     NOT NULL CHECK (quantity > 0),
    status         TEXT        NOT NULL CHECK (status IN ('active', 'confirmed', 'expired')),
    expires_at     TIMESTAMPTZ NOT NULL,
    confirmed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Sweeper hanya memindai reservasi aktif yang jatuh tempo.
CREATE INDEX IF NOT EXISTS idx_reservations_active_expiry
    ON reservations (expires_at) WHERE status = 'active';

-- Lookup per item (breakdown stok) dan per user.
CREATE INDEX IF NOT EXISTS idx_reservations_item_status ON reservations (item_id, status);
CREATE INDEX IF NOT EXISTS idx_reservations_user_id ON reservations (user_id);
