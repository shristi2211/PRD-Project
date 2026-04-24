-- ============================================
-- Phase 6: Pre-Registration IP Subscription Lock
-- ============================================

CREATE TABLE IF NOT EXISTS ip_subscriptions (
    ip_address VARCHAR(45) PRIMARY KEY,
    plan_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
