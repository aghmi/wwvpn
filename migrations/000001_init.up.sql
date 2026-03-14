CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    revenuecat_id VARCHAR(255),
    plan VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(device_id)
);

CREATE INDEX idx_subscriptions_device_id ON subscriptions(device_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);

CREATE TABLE servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    country VARCHAR(2) NOT NULL,
    city VARCHAR(100) NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    grpc_port INT NOT NULL DEFAULT 50051,
    awg_port INT NOT NULL DEFAULT 51820,
    vless_port INT NOT NULL DEFAULT 443,
    awg_public_key VARCHAR(255),
    vless_public_key VARCHAR(255),
    vless_short_id VARCHAR(16),
    camouflage_domain VARCHAR(255) DEFAULT 'www.microsoft.com',
    capacity INT NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_premium BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    server_id UUID NOT NULL REFERENCES servers(id),
    protocol VARCHAR(50) NOT NULL,
    connected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    disconnected_at TIMESTAMPTZ
);

CREATE INDEX idx_connections_device_id ON connections(device_id);
CREATE INDEX idx_connections_active ON connections(device_id) WHERE disconnected_at IS NULL;
