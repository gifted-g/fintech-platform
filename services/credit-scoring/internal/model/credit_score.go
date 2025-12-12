package model

import "time"

type CreditScore struct {
	ID             string    `db:"id"`
	UserID         string    `db:"user_id"`
	Score          int       `db:"score"`
	Grade          string    `db:"grade"`
	Factors        []string  `db:"factors"`
	Recommendation string    `db:"recommendation"`
	CalculatedAt   time.Time `db:"calculated_at"`
	ExpiresAt      time.Time `db:"expires_at"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
