ALTER TABLE shop_orders
    ADD COLUMN IF NOT EXISTS fulfillment_note TEXT NOT NULL DEFAULT '';
