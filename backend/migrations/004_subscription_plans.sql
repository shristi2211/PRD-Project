-- ============================================
-- Phase 6: Subscription Plans (Free / Monthly / Yearly)
-- ============================================

-- Add subscription_type column
ALTER TABLE users ADD COLUMN IF NOT EXISTS subscription_type VARCHAR(20) NOT NULL DEFAULT 'free';

-- Migrate existing data based on subscription_active flag
UPDATE users SET subscription_type = 'monthly' WHERE subscription_active = true AND subscription_type = 'free';

-- Index for fast plan-based filtering
CREATE INDEX IF NOT EXISTS idx_users_subscription_type ON users(subscription_type);
