-- Migration: Create user_verifications table
-- Version: 003
-- Description: Table for storing user verification records (BVN, NIN, KYC)

CREATE TABLE IF NOT EXISTS user_verifications (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    verification_type VARCHAR(50) NOT NULL CHECK (verification_type IN ('BVN', 'NIN', 'PASSPORT', 'DRIVERS_LICENSE', 'KYC')),
    verification_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('PENDING', 'VERIFIED', 'FAILED', 'EXPIRED')),
    verified_data JSONB,
    verification_method VARCHAR(100),
    verified_at TIMESTAMP,
    expires_at TIMESTAMP,
    failure_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, verification_type)
);

-- Indexes
CREATE INDEX idx_user_verifications_user_id ON user_verifications(user_id);
CREATE INDEX idx_user_verifications_status ON user_verifications(status);
CREATE INDEX idx_user_verifications_type ON user_verifications(verification_type);
CREATE INDEX idx_user_verifications_verified_at ON user_verifications(verified_at DESC);

-- Trigger
CREATE TRIGGER update_user_verifications_updated_at
    BEFORE UPDATE ON user_verifications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE user_verifications IS 'Stores user identity verification records';
COMMENT ON COLUMN user_verifications.verification_type IS 'Type of verification (BVN, NIN, etc.)';
COMMENT ON COLUMN user_verifications.verification_id IS 'The actual BVN/NIN/ID number (encrypted)';
COMMENT ON COLUMN user_verifications.verified_data IS 'JSON data returned from verification service';
