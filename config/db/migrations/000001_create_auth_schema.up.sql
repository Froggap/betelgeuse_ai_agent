CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.roles (
    id        TEXT PRIMARY KEY,
    name_role TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS auth.sessions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    refresh_token TEXT NOT NULL DEFAULT '',
    refresh_code  TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS auth.users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_name     TEXT NOT NULL UNIQUE,
    password      TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by    UUID,
    account_valid TEXT NOT NULL DEFAULT 'VALID',
    session_id    UUID NOT NULL REFERENCES auth.sessions (id),
    role_id       TEXT NOT NULL REFERENCES auth.roles (id)
);

INSERT INTO auth.roles (id, name_role)
VALUES ('ADM', 'Administrator')
ON CONFLICT (id) DO NOTHING;
