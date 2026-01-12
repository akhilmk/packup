-- Database schema for Itinera
-- Run this manually when needed during development

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    google_id TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    name TEXT,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS todos (
    id TEXT PRIMARY KEY,
    text TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created TIMESTAMPTZ NOT NULL DEFAULT now(),
    position DOUBLE PRECISION NOT NULL DEFAULT 0,
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE
);
