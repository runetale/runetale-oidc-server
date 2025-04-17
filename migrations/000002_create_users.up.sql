CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY NOT NULL UNIQUE,
  username VARCHAR(255) NOT NULL,
  user_id VARCHAR(255) NOT NULL UNIQUE,
  tenant_id INT NOT NULL,
  provider_id VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  domain VARCHAR(255) NOT NULL,
  sub VARCHAR(255) NOT NULL,
  aud VARCHAR(255) NOT NULL,
  azp VARCHAR(255) NOT NULL,
  picture VARCHAR(255) NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,

  CONSTRAINT fk_tenant_id FOREIGN KEY(tenant_id)
    REFERENCES public.tenants(id)
);

CREATE INDEX IF NOT EXISTS idx_users_user_id ON users(user_id);