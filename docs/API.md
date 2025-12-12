# API Documentation

## Authentication

All API requests require authentication using JWT tokens.

### Get Access Token

\`\`\`http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password"
}
\`\`\`

Response:
\`\`\`json
{
  "success": true,
  "data": {
    "accessToken": "eyJhbGc...",
    "refreshToken": "eyJhbGc...",
    "expiresIn": 900
  }
}
\`\`\`

### Use Token

\`\`\`http
GET /api/v1/credit/score/user123
Authorization: Bearer eyJhbGc...
\`\`\`

## Credit Scoring API

### Calculate Credit Score

\`\`\`http
POST /api/v1/credit/score
Authorization: Bearer {token}
Content-Type: application/json

{
  "userId": "user123",
  "incomeAmount": 150000,
  "employmentStatus": "employed",
  "accountAge": 24,
  "transactionData": {},
  "loanHistory": []
}
\`\`\`

Response:
\`\`\`json
{
  "success": true,
  "data": {
    "id": "cs_1234567890",
    "userId": "user123",
    "score": 720,
    "grade": "Good",
    "factors": [
      "Strong payment history",
      "Good financial stability"
    ],
    "recommendation": "Good credit profile. Eligible for competitive rates.",
    "calculatedAt": "2025-01-15T10:30:00Z",
    "expiresAt": "2025-02-15T10:30:00Z"
  }
}
\`\`\`

### Get Credit Score

\`\`\`http
GET /api/v1/credit/score/:userId
Authorization: Bearer {token}
\`\`\`

### Get Credit History

\`\`\`http
GET /api/v1/credit/history/:userId
Authorization: Bearer {token}
\`\`\`

## Risk Assessment API

### Assess Risk

\`\`\`http
POST /api/v1/risk/assess
Authorization: Bearer {token}
Content-Type: application/json

{
  "userId": "user123",
  "loanAmount": 50000,
  "loanPurpose": "business",
  "creditScoreId": "cs_1234567890"
}
\`\`\`

Response:
\`\`\`json
{
  "success": true,
  "data": {
    "id": "ra_9876543210",
    "userId": "user123",
    "riskLevel": "MEDIUM",
    "riskScore": 65.5,
    "fraudProbability": 12.3,
    "defaultProbability": 8.7,
    "recommendedAction": "APPROVE_WITH_CONDITIONS"
  }
}
\`\`\`

## Error Responses

\`\`\`json
{
  "code": "VALIDATION_ERROR",
  "message": "Income amount is required"
}
\`\`\`

Common error codes:
- `UNAUTHORIZED` - Missing or invalid token
- `FORBIDDEN` - Insufficient permissions
- `VALIDATION_ERROR` - Invalid request data
- `NOT_FOUND` - Resource not found
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `INTERNAL_ERROR` - Server error
