-- Database schema for Packup
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
    is_default_task BOOLEAN NOT NULL DEFAULT false,
    shared_with_admin BOOLEAN NOT NULL DEFAULT false,
    hidden_from_user BOOLEAN NOT NULL DEFAULT false
);

-- Add column if it doesn't exist (for existing databases)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='todos' AND column_name='hidden_from_user') THEN
        ALTER TABLE todos ADD COLUMN hidden_from_user BOOLEAN NOT NULL DEFAULT false;
    END IF;
END $$;

-- Junction table for per-user state of default tasks
CREATE TABLE IF NOT EXISTS user_todo_state (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    todo_id TEXT NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    position DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, todo_id)
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_todos_default ON todos(is_default_task) WHERE is_default_task = true;
CREATE INDEX IF NOT EXISTS idx_user_todo_state_user ON user_todo_state(user_id);
CREATE INDEX IF NOT EXISTS idx_user_todo_state_todo ON user_todo_state(todo_id);
