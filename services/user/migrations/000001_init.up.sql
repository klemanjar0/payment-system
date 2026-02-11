-- users
CREATE TYPE user_status AS ENUM ('pending', 'active', 'blocked', 'deleted');
CREATE TYPE kyc_status AS ENUM ('none', 'pending', 'verified', 'rejected');
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    status user_status NOT NULL DEFAULT 'pending',
    kyc_status kyc_status NOT NULL DEFAULT 'none',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_status ON users(status);
-- refresh tokens
create table refresh_tokens (
    id uuid primary key default gen_random_uuid(),
    user_id uuid references users(id) on delete cascade,
    device_info text,
    rotated_from uuid references refresh_tokens(id),
    revoked bool default false,
    created_at timestamptz default now(),
    expires_at timestamptz not null,
    last_used_at timestamptz
);
create index idx_refresh_tokens_user_id on refresh_tokens(user_id);
create index idx_refresh_tokens_expires on refresh_tokens(expires_at)
where revoked = false;
create index idx_refresh_tokens_rotated on refresh_tokens(rotated_from)
where rotated_from is not null;