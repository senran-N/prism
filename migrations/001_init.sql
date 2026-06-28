CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    github_id       BIGINT UNIQUE,
    github_login    TEXT NOT NULL DEFAULT '',
    avatar_url      TEXT NOT NULL DEFAULT '',
    github_token    TEXT NOT NULL DEFAULT '',
    selected_repo   TEXT NOT NULL DEFAULT '',
    linuxdo_id      BIGINT UNIQUE,
    linuxdo_username TEXT NOT NULL DEFAULT '',
    linuxdo_name    TEXT NOT NULL DEFAULT '',
    trust_level     INT NOT NULL DEFAULT 0,
    is_banned       BOOLEAN NOT NULL DEFAULT false,
    ban_reason      TEXT NOT NULL DEFAULT '',
    is_admin        BOOLEAN NOT NULL DEFAULT false,
    balance         NUMERIC(10,2) NOT NULL DEFAULT 0,
    total_rotations INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sc_accounts (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL,
    password      TEXT NOT NULL,
    workspace_id  TEXT NOT NULL DEFAULT '',
    user_id       TEXT NOT NULL DEFAULT '',
    project_id    TEXT NOT NULL DEFAULT '',
    repo_id       TEXT NOT NULL DEFAULT '',
    credits       NUMERIC(10,2) NOT NULL DEFAULT 20.00,
    status        TEXT NOT NULL DEFAULT 'registered',
    github_bound  BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS tasks (
    id            TEXT PRIMARY KEY,
    user_id       BIGINT REFERENCES users(id),
    account_id    TEXT,
    ticket_id     TEXT NOT NULL,
    project_id    TEXT NOT NULL DEFAULT '',
    description   TEXT NOT NULL,
    model         TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'created',
    cost          NUMERIC(10,2) NOT NULL DEFAULT 0,
    view_url      TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS redemption_codes (
    id          BIGSERIAL PRIMARY KEY,
    code        TEXT NOT NULL UNIQUE,
    rotations   INT NOT NULL DEFAULT 1,
    used_count  INT NOT NULL DEFAULT 0,
    max_uses    INT NOT NULL DEFAULT 1,
    created_by  BIGINT REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS redemption_log (
    id          BIGSERIAL PRIMARY KEY,
    code_id     BIGINT REFERENCES redemption_codes(id),
    user_id     BIGINT REFERENCES users(id),
    rotations   INT NOT NULL,
    redeemed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_fingerprints (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT REFERENCES users(id),
    ip          TEXT NOT NULL DEFAULT '',
    user_agent  TEXT NOT NULL DEFAULT '',
    language    TEXT NOT NULL DEFAULT '',
    timezone    TEXT NOT NULL DEFAULT '',
    screen      TEXT NOT NULL DEFAULT '',
    platform    TEXT NOT NULL DEFAULT '',
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_sc_accounts_status ON sc_accounts(status);
CREATE INDEX IF NOT EXISTS idx_redemption_codes_code ON redemption_codes(code);
CREATE INDEX IF NOT EXISTS idx_user_fingerprints_user ON user_fingerprints(user_id);
