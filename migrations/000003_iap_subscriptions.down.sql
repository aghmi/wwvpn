DROP TABLE IF EXISTS device_subscriptions;

DROP INDEX IF EXISTS idx_device_subscriptions_device_id;
DROP INDEX IF EXISTS idx_device_subscriptions_subscription_id;
DROP INDEX IF EXISTS idx_subscriptions_store_tx;

ALTER TABLE subscriptions DROP COLUMN IF EXISTS raw_receipt;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS environment;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS product_id;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS platform;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS store_transaction_id;

ALTER TABLE subscriptions
  ADD CONSTRAINT subscriptions_device_id_key UNIQUE (device_id);

