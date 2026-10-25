CREATE TABLE IF NOT EXISTS privacy_policies (
    id VARCHAR(64) PRIMARY KEY,
    version VARCHAR(50) NOT NULL UNIQUE,
    content TEXT NOT NULL,
    effective_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_privacy_policy_acknowledgments (
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    acknowledged_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, version)
);

-- Default Privacy Policy Seed
INSERT INTO privacy_policies (id, version, content, effective_date) VALUES
    ('privacy_policy_v1_0', '1.0', 'Fundly Privacy Policy version 1.0. We collect only the data necessary to provide budgeting and transaction features, store it securely, never sell it to third parties, and retain it only as long as your account is active or as required by law.', '2026-01-01 00:00:00+00')
ON CONFLICT (id) DO NOTHING;
