-- Migration: Create fraud_alerts table
-- Version: 004
-- Description: Table for storing fraud detection alerts

CREATE TABLE IF NOT EXISTS fraud_alerts (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    alert_type VARCHAR(100) NOT NULL,
    severity VARCHAR(50) NOT NULL CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    fraud_score DECIMAL(5,2) NOT NULL CHECK (fraud_score >= 0 AND fraud_score <= 100),
    description TEXT NOT NULL,
    indicators JSONB NOT NULL DEFAULT '[]',
    transaction_data JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'INVESTIGATING', 'CONFIRMED', 'FALSE_POSITIVE', 'RESOLVED')),
    resolved_at TIMESTAMP,
    resolved_by VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_fraud_alerts_user_id ON fraud_alerts(user_id);
CREATE INDEX idx_fraud_alerts_severity ON fraud_alerts(severity);
CREATE INDEX idx_fraud_alerts_status ON fraud_alerts(status);
CREATE INDEX idx_fraud_alerts_created_at ON fraud_alerts(created_at DESC);
CREATE INDEX idx_fraud_alerts_user_created ON fraud_alerts(user_id, created_at DESC);

-- Trigger
CREATE TRIGGER update_fraud_alerts_updated_at
    BEFORE UPDATE ON fraud_alerts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE fraud_alerts IS 'Stores fraud detection alerts and their investigation status';
COMMENT ON COLUMN fraud_alerts.fraud_score IS 'Calculated fraud probability score (0-100)';
COMMENT ON COLUMN fraud_alerts.indicators IS 'JSON array of fraud indicators detected';
