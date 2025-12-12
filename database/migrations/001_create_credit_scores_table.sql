-- Migration: Create credit_scores table
-- Version: 001
-- Description: Initial table for storing credit score calculations

CREATE TABLE IF NOT EXISTS credit_scores (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    score INTEGER NOT NULL CHECK (score >= 300 AND score <= 850),
    grade VARCHAR(50) NOT NULL,
    factors TEXT[] NOT NULL DEFAULT '{}',
    recommendation TEXT NOT NULL,
    calculated_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient queries
CREATE INDEX idx_credit_scores_user_id ON credit_scores(user_id);
CREATE INDEX idx_credit_scores_calculated_at ON credit_scores(calculated_at DESC);
CREATE INDEX idx_credit_scores_user_calculated ON credit_scores(user_id, calculated_at DESC);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_credit_scores_updated_at
    BEFORE UPDATE ON credit_scores
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE credit_scores IS 'Stores calculated credit scores for users';
COMMENT ON COLUMN credit_scores.id IS 'Unique identifier for the credit score record';
COMMENT ON COLUMN credit_scores.user_id IS 'Reference to the user';
COMMENT ON COLUMN credit_scores.score IS 'Calculated credit score (300-850)';
COMMENT ON COLUMN credit_scores.grade IS 'Credit grade (Excellent, Very Good, Good, Fair, Poor)';
COMMENT ON COLUMN credit_scores.factors IS 'Array of factors that influenced the score';
COMMENT ON COLUMN credit_scores.recommendation IS 'Recommendation based on the score';
COMMENT ON COLUMN credit_scores.calculated_at IS 'When the score was calculated';
COMMENT ON COLUMN credit_scores.expires_at IS 'When the score expires and needs recalculation';
