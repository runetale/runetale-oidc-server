CREATE TABLE IF NOT EXISTS tenants (
  id BIGSERIAL PRIMARY KEY NOT NULL UNIQUE,
  tenant_id VARCHAR(255) NOT NULL,
  domain VARCHAR(255) NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tenants_tenant_id ON tenants(tenant_id);