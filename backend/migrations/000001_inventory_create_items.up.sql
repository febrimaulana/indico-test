-- 000001_inventory_create_items: tabel stok utama.
CREATE TABLE IF NOT EXISTS items (
    item_id        TEXT PRIMARY KEY,
    name           TEXT        NOT NULL,
    total_stock    INTEGER     NOT NULL CHECK (total_stock >= 0),
    reserved_stock INTEGER     NOT NULL DEFAULT 0 CHECK (reserved_stock >= 0),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Jaring terakhir anti stok negatif / over-reserve.
    CONSTRAINT items_reserved_within_total CHECK (reserved_stock <= total_stock)
);
