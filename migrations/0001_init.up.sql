-- +migrate Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table (main entry point)
-- Users table using Auth0 sub (e.g. "auth0|abc123") as primary key
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,             -- Auth0's sub claim (e.g. "auth0|abc123")
    email TEXT UNIQUE NOT NULL,      -- Unique user email
    name TEXT NOT NULL,              -- Full name from JWT or signup
    business_name TEXT NOT NULL,     -- From signup form
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Optional performance index on email for fast lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Links table
CREATE TABLE IF NOT EXISTS links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slug TEXT UNIQUE NOT NULL,
    destination TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Clicks table
CREATE TABLE IF NOT EXISTS clicks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    ip TEXT,
    referrer TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);
