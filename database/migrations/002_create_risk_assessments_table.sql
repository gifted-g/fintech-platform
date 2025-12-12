-- Migration: Create risk_assessments table
-- Version: 002
-- Description: Table for storing risk assessment reports

CREATE TABLE IF NOT EXISTS risk_assessments (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    credit_score_id VARCHAR(255) REFERENCES credit_scores(id),
    risk_level VARCHAR(50) NOT NULL CHECK (risk_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    risk_score DECIMAL(5,2) NOT NULL CHECK (risk_score >= 0 AND risk_score <= 100),
    fraud_probability DECIMAL(5,2) NOT NULL CHECK (fraud_probability >= 0 AND fraud_probability <= 100),
    default_probability DECIMAL(5,2) NOT NULL CHECK (default_probability >= 0 AND default_probability <= 100),
    recommended_action VARCHAR(100) NOT NULL,
    risk_factors JSONB NOT NULL DEFAULT '[]',
    assessed_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_risk_assessments_user_id ON risk_assessments(user_id);
CREATE INDEX idx_risk_assessments_risk_level ON risk_assessments(risk_level);
CREATE INDEX idx_risk_assessments_assessed_at ON risk_assessments(assessed_at DESC);
CREATE INDEX idx_risk_assessments_user_assessed ON risk_assessments(user_id, assessed_at DESC);

-- Trigger
CREATE TRIGGER update_risk_assessments_updated_at
    BEFORE UPDATE ON risk_assessments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE risk_assessments IS 'Stores risk assessment reports for users';
COMMENT ON COLUMN risk_assessments.risk_level IS 'Overall risk level category';
COMMENT ON COLUMN risk_assessments.risk_score IS 'Numeric risk score (0-100)';
COMMENT ON COLUMN risk_assessments.fraud_probability IS 'Probability of fraud (0-100)';
COMMENT ON COLUMN risk_assessments.default_probability IS 'Probability of default (0-100)';
