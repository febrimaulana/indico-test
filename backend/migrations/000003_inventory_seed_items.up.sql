-- 000003_inventory_seed_items: item demo untuk reviewer (idempotent).
INSERT INTO items (item_id, name, total_stock, reserved_stock)
VALUES ('item_4021', 'INDICO Flash Sale Item', 100, 0)
ON CONFLICT (item_id) DO NOTHING;

INSERT INTO items (item_id, name, total_stock, reserved_stock)
VALUES ('item_4022', 'INDICO Limited Edition Item', 50, 0)
ON CONFLICT (item_id) DO NOTHING;
