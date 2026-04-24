-- ============================================
-- Migration: Multi-Charity Support
-- Allows users to allocate portions of their 30% pool to multiple charities
-- ============================================

-- Drop the old UNIQUE constraint on user_id so they can select multiple charities
ALTER TABLE user_charity_selections DROP CONSTRAINT IF EXISTS user_charity_selections_user_id_key;

-- Drop the 10-100% check constraint, because charities can now receive smaller fragments
ALTER TABLE user_charity_selections DROP CONSTRAINT IF EXISTS user_charity_selections_contribution_percentage_check;

-- Ensure a user can still only have one allocation per specific charity (composite key)
ALTER TABLE user_charity_selections ADD CONSTRAINT user_charity_user_charity_unique UNIQUE (user_id, charity_id);
