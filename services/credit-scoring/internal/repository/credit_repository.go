package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"

	"credit-scoring/internal/model"
)

type CreditRepository struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) *CreditRepository {
	return &CreditRepository{db: db}
}

func (r *CreditRepository) Create(ctx context.Context, score *model.CreditScore) error {
	query := `
		INSERT INTO credit_scores (id, user_id, score, grade, factors, recommendation, calculated_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		score.ID,
		score.UserID,
		score.Score,
		score.Grade,
		pq.Array(score.Factors),
		score.Recommendation,
		score.CalculatedAt,
		score.ExpiresAt,
	)
	return err
}

func (r *CreditRepository) GetLatestByUserID(ctx context.Context, userID string) (*model.CreditScore, error) {
	query := `
		SELECT id, user_id, score, grade, factors, recommendation, calculated_at, expires_at, created_at, updated_at
		FROM credit_scores
		WHERE user_id = $1
		ORDER BY calculated_at DESC
		LIMIT 1
	`
	
	score := &model.CreditScore{}
	var factors pq.StringArray
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&score.ID,
		&score.UserID,
		&score.Score,
		&score.Grade,
		&factors,
		&score.Recommendation,
		&score.CalculatedAt,
		&score.ExpiresAt,
		&score.CreatedAt,
		&score.UpdatedAt,
	)
	
	score.Factors = []string(factors)
	
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	
	return score, err
}

func (r *CreditRepository) GetHistoryByUserID(ctx context.Context, userID string, limit int) ([]*model.CreditScore, error) {
	query := `
		SELECT id, user_id, score, grade, factors, recommendation, calculated_at, expires_at, created_at, updated_at
		FROM credit_scores
		WHERE user_id = $1
		ORDER BY calculated_at DESC
		LIMIT $2
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []*model.CreditScore
	for rows.Next() {
		score := &model.CreditScore{}
		var factors pq.StringArray
		
		if err := rows.Scan(
			&score.ID,
			&score.UserID,
			&score.Score,
			&score.Grade,
			&factors,
			&score.Recommendation,
			&score.CalculatedAt,
			&score.ExpiresAt,
			&score.CreatedAt,
			&score.UpdatedAt,
		); err != nil {
			return nil, err
		}
		
		score.Factors = []string(factors)
		scores = append(scores, score)
	}

	return scores, rows.Err()
}

var ErrNotFound = sql.ErrNoRows
