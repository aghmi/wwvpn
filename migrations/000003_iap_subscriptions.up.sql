ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_device_id_key;

ALTER TABLE subscriptions
  ADD COLUMN IF NOT EXISTS store_transaction_id TEXT;

ALTER TABLE subscriptions
  ADD COLUMN IF NOT EXISTS platform TEXT,
  ADD COLUMN IF NOT EXISTS product_id TEXT,
  ADD COLUMN IF NOT EXISTS environment TEXT,
  ADD COLUMN IF NOT EXISTS raw_receipt TEXT;

UPDATE subscriptions
SET store_transaction_id = 'legacy:' || id::text,
    platform = COALESCE(platform, 'legacy'),
    product_id = COALESCE(product_id, plan),
    environment = environment,
    raw_receipt = raw_receipt
WHERE store_transaction_id IS NULL;

ALTER TABLE subscriptions
  ALTER COLUMN store_transaction_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_store_tx
  ON subscriptions (store_transaction_id);

CREATE TABLE IF NOT EXISTS device_subscriptions (
  device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
  subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
  linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (device_id, subscription_id)
);

CREATE INDEX IF NOT EXISTS idx_device_subscriptions_device_id
  ON device_subscriptions (device_id);

CREATE INDEX IF NOT EXISTS idx_device_subscriptions_subscription_id
  ON device_subscriptions (subscription_id);

INSERT INTO device_subscriptions (device_id, subscription_id)
SELECT s.device_id, s.id
FROM subscriptions s
ON CONFLICT (device_id, subscription_id) DO NOTHING;

