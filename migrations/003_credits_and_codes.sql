-- User balance for rotation credits
ALTER TABLE users ADD COLUMN IF NOT EXISTS balance NUMERIC(10,2) NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_rotations INT NOT NULL DEFAULT 0;

-- Redemption codes
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

-- Redemption log
CREATE TABLE IF NOT EXISTS redemption_log (
    id          BIGSERIAL PRIMARY KEY,
    code_id     BIGINT REFERENCES redemption_codes(id),
    user_id     BIGINT REFERENCES users(id),
    rotations   INT NOT NULL,
    redeemed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- User fingerprints for reducing SC registration risk
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

CREATE INDEX IF NOT EXISTS idx_redemption_codes_code ON redemption_codes(code);
CREATE INDEX IF NOT EXISTS idx_user_fingerprints_user ON user_fingerprints(user_id);
