-- Database schema for Itinera
-- Run this manually when needed during development

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    google_id TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    name TEXT,
    avatar_url TEXT,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
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
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    created_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    is_admin_todo BOOLEAN NOT NULL DEFAULT false
);

-- Junction table for per-user state of admin todos
CREATE TABLE IF NOT EXISTS user_todo_state (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    todo_id TEXT NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    position DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, todo_id)
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_todos_admin ON todos(is_admin_todo) WHERE is_admin_todo = true;
CREATE INDEX IF NOT EXISTS idx_user_todo_state_user ON user_todo_state(user_id);
CREATE INDEX IF NOT EXISTS idx_user_todo_state_todo ON user_todo_state(todo_id);
