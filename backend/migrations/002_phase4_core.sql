-- ============================================
-- Phase 4: Core Engines — Scoring, Draws, Charities, Winners, Activity Logs
-- ============================================

-- ============================================
-- Charities table
-- ============================================
CREATE TABLE IF NOT EXISTS charities (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    website     VARCHAR(500) NOT NULL DEFAULT '',
    logo_url    VARCHAR(500) NOT NULL DEFAULT '',
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_charities_active ON charities(active);

CREATE TRIGGER trigger_charities_updated_at
    BEFORE UPDATE ON charities
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Scores table (Stableford: 1–45)
-- ============================================
CREATE TABLE IF NOT EXISTS scores (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score       INTEGER NOT NULL CHECK (score >= 1 AND score <= 45),
    round_date  DATE NOT NULL DEFAULT CURRENT_DATE,
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scores_user_id ON scores(user_id);
CREATE INDEX IF NOT EXISTS idx_scores_round_date ON scores(round_date);
CREATE INDEX IF NOT EXISTS idx_scores_user_score ON scores(user_id, score DESC);

-- ============================================
-- Draws table
-- ============================================
CREATE TABLE IF NOT EXISTS draws (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    draw_date       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    month           INTEGER NOT NULL CHECK (month >= 1 AND month <= 12),
    year            INTEGER NOT NULL CHECK (year >= 2020),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'running', 'completed', 'cancelled')),
    total_pool      DECIMAL(12,2) NOT NULL DEFAULT 0,
    winner_prize    DECIMAL(12,2) NOT NULL DEFAULT 0,
    charity_amount  DECIMAL(12,2) NOT NULL DEFAULT 0,
    platform_fee    DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_entries   INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(month, year)
);

CREATE INDEX IF NOT EXISTS idx_draws_status ON draws(status);
CREATE INDEX IF NOT EXISTS idx_draws_month_year ON draws(month, year);

-- ============================================
-- Draw entries table
-- ============================================
CREATE TABLE IF NOT EXISTS draw_entries (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    draw_id     UUID NOT NULL REFERENCES draws(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score_id    UUID REFERENCES scores(id) ON DELETE SET NULL,
    entry_score INTEGER NOT NULL CHECK (entry_score >= 1 AND entry_score <= 45),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(draw_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_draw_entries_draw_id ON draw_entries(draw_id);
CREATE INDEX IF NOT EXISTS idx_draw_entries_user_id ON draw_entries(user_id);

-- ============================================
-- Winners table
-- ============================================
CREATE TABLE IF NOT EXISTS winners (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    draw_id             UUID NOT NULL REFERENCES draws(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    prize_amount        DECIMAL(12,2) NOT NULL DEFAULT 0,
    proof_url           VARCHAR(500) NOT NULL DEFAULT '',
    proof_notes         TEXT NOT NULL DEFAULT '',
    verification_status VARCHAR(20) NOT NULL DEFAULT 'pending'
                        CHECK (verification_status IN ('pending', 'approved', 'rejected')),
    rejection_reason    TEXT NOT NULL DEFAULT '',
    verified_by         UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_winners_draw_id ON winners(draw_id);
CREATE INDEX IF NOT EXISTS idx_winners_user_id ON winners(user_id);
CREATE INDEX IF NOT EXISTS idx_winners_verification_status ON winners(verification_status);

-- ============================================
-- User charity selections (one active per user)
-- ============================================
CREATE TABLE IF NOT EXISTS user_charity_selections (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    charity_id              UUID NOT NULL REFERENCES charities(id) ON DELETE CASCADE,
    contribution_percentage INTEGER NOT NULL DEFAULT 10 CHECK (contribution_percentage >= 10 AND contribution_percentage <= 100),
    selected_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_charity_user_id ON user_charity_selections(user_id);
CREATE INDEX IF NOT EXISTS idx_user_charity_charity_id ON user_charity_selections(charity_id);

-- ============================================
-- Activity logs
-- ============================================
CREATE TABLE IF NOT EXISTS activity_logs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    action      VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL DEFAULT '',
    entity_id   VARCHAR(100) NOT NULL DEFAULT '',
    metadata    JSONB NOT NULL DEFAULT '{}',
    ip_address  VARCHAR(45) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_action ON activity_logs(action);
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_logs_entity ON activity_logs(entity_type, entity_id);
