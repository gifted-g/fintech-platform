package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"

	"credit-scoring/internal/dto"
	"credit-scoring/internal/model"
	"credit-scoring/internal/repository"
	"credit-scoring/pkg/kafka"
	"credit-scoring/pkg/redis"
)

type CreditScoringService struct {
	repo     *repository.CreditRepository
	cache    *redis.RedisClient
	producer *kafka.Producer
	logger   *zap.Logger
}

func NewCreditScoringService(
	repo *repository.CreditRepository,
	cache *redis.RedisClient,
	producer *kafka.Producer,
	logger *zap.Logger,
) *CreditScoringService {
	return &CreditScoringService{
		repo:     repo,
		cache:    cache,
		producer: producer,
		logger:   logger,
	}
}

// CalculateScore calculates credit score using proprietary algorithm
func (s *CreditScoringService) CalculateScore(ctx context.Context, req *dto.CalculateScoreRequest) (*dto.CreditScore, error) {
	s.logger.Info("Calculating credit score", zap.String("userId", req.UserID))

	// Calculate base score components
	incomeScore := s.calculateIncomeScore(req.IncomeAmount)
	employmentScore := s.calculateEmploymentScore(req.EmploymentStatus)
	accountAgeScore := s.calculateAccountAgeScore(req.AccountAge)
	loanHistoryScore := s.calculateLoanHistoryScore(req.LoanHistory)

	// Weighted average
	totalScore := int(
		incomeScore*0.30 +
			employmentScore*0.25 +
			accountAgeScore*0.20 +
			loanHistoryScore*0.25,
	)

	// Ensure score is in valid range (300-850)
	if totalScore < 300 {
		totalScore = 300
	}
	if totalScore > 850 {
		totalScore = 850
	}

	// Determine grade
	grade := s.getGrade(totalScore)
	
	// Generate factors and recommendation
	factors := s.generateFactors(req, totalScore)
	recommendation := s.generateRecommendation(totalScore)

	creditScore := &dto.CreditScore{
		ID:             generateID(),
		UserID:         req.UserID,
		Score:          totalScore,
		Grade:          grade,
		Factors:        factors,
		Recommendation: recommendation,
		CalculatedAt:   time.Now(),
		ExpiresAt:      time.Now().Add(30 * 24 * time.Hour),
	}

	// Save to database
	dbModel := &model.CreditScore{
		ID:             creditScore.ID,
		UserID:         creditScore.UserID,
		Score:          creditScore.Score,
		Grade:          creditScore.Grade,
		Factors:        creditScore.Factors,
		Recommendation: creditScore.Recommendation,
		CalculatedAt:   creditScore.CalculatedAt,
		ExpiresAt:      creditScore.ExpiresAt,
	}

	if err := s.repo.Create(ctx, dbModel); err != nil {
		s.logger.Error("Failed to save credit score", zap.Error(err))
		return nil, err
	}

	// Cache the result
	cacheKey := fmt.Sprintf("credit_score:%s", req.UserID)
	if err := s.cache.Set(ctx, cacheKey, creditScore, 15*time.Minute); err != nil {
		s.logger.Warn("Failed to cache credit score", zap.Error(err))
	}

	// Publish event to Kafka
	event := map[string]interface{}{
		"eventType":  "credit_score_calculated",
		"userId":     req.UserID,
		"score":      totalScore,
		"grade":      grade,
		"timestamp":  time.Now(),
	}
	eventJSON, _ := json.Marshal(event)
	if err := s.producer.Publish(ctx, "credit-scoring-events", eventJSON); err != nil {
		s.logger.Warn("Failed to publish event", zap.Error(err))
	}

	return creditScore, nil
}

// GetScore retrieves the current credit score from cache or database
func (s *CreditScoringService) GetScore(ctx context.Context, userID string) (*dto.CreditScore, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("credit_score:%s", userID)
	var cachedScore dto.CreditScore
	if err := s.cache.Get(ctx, cacheKey, &cachedScore); err == nil {
		return &cachedScore, nil
	}

	// Fetch from database
	dbScore, err := s.repo.GetLatestByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	score := &dto.CreditScore{
		ID:             dbScore.ID,
		UserID:         dbScore.UserID,
		Score:          dbScore.Score,
		Grade:          dbScore.Grade,
		Factors:        dbScore.Factors,
		Recommendation: dbScore.Recommendation,
		CalculatedAt:   dbScore.CalculatedAt,
		ExpiresAt:      dbScore.ExpiresAt,
	}

	// Cache for next time
	s.cache.Set(ctx, cacheKey, score, 15*time.Minute)

	return score, nil
}

// GetHistory retrieves credit score history
func (s *CreditScoringService) GetHistory(ctx context.Context, userID string) (*dto.CreditScoreHistory, error) {
	dbScores, err := s.repo.GetHistoryByUserID(ctx, userID, 12) // Last 12 scores
	if err != nil {
		return nil, err
	}

	history := make([]dto.CreditScore, len(dbScores))
	for i, dbScore := range dbScores {
		history[i] = dto.CreditScore{
			ID:             dbScore.ID,
			UserID:         dbScore.UserID,
			Score:          dbScore.Score,
			Grade:          dbScore.Grade,
			Factors:        dbScore.Factors,
			Recommendation: dbScore.Recommendation,
			CalculatedAt:   dbScore.CalculatedAt,
			ExpiresAt:      dbScore.ExpiresAt,
		}
	}

	return &dto.CreditScoreHistory{
		UserID:  userID,
		History: history,
	}, nil
}

// RefreshScore forces recalculation of credit score
func (s *CreditScoringService) RefreshScore(ctx context.Context, userID string) (*dto.CreditScore, error) {
	// Invalidate cache
	cacheKey := fmt.Sprintf("credit_score:%s", userID)
	s.cache.Delete(ctx, cacheKey)

	// In production, you'd fetch fresh data and recalculate
	// For now, return the latest score
	return s.GetScore(ctx, userID)
}

// Helper functions for score calculation
func (s *CreditScoringService) calculateIncomeScore(income float64) float64 {
	if income < 50000 {
		return 300
	} else if income < 100000 {
		return 450
	} else if income < 200000 {
		return 600
	} else if income < 500000 {
		return 750
	}
	return 850
}

func (s *CreditScoringService) calculateEmploymentScore(status string) float64 {
	scores := map[string]float64{
		"employed":      750,
		"self-employed": 650,
		"unemployed":    350,
		"retired":       550,
	}
	return scores[status]
}

func (s *CreditScoringService) calculateAccountAgeScore(ageInMonths int) float64 {
	return math.Min(300+float64(ageInMonths)*10, 850)
}

func (s *CreditScoringService) calculateLoanHistoryScore(history []dto.LoanHistoryItem) float64 {
	if len(history) == 0 {
		return 500 // Neutral score
	}

	paidOnTime := 0
	for _, loan := range history {
		if loan.Status == "paid" {
			paidOnTime++
		}
	}

	ratio := float64(paidOnTime) / float64(len(history))
	return 300 + (ratio * 550)
}

func (s *CreditScoringService) getGrade(score int) string {
	if score >= 800 {
		return "Excellent"
	} else if score >= 740 {
		return "Very Good"
	} else if score >= 670 {
		return "Good"
	} else if score >= 580 {
		return "Fair"
	}
	return "Poor"
}

func (s *CreditScoringService) generateFactors(req *dto.CalculateScoreRequest, score int) []string {
	factors := []string{}
	
	if req.IncomeAmount < 100000 {
		factors = append(factors, "Low income level")
	}
	if req.AccountAge < 12 {
		factors = append(factors, "Short account history")
	}
	if req.EmploymentStatus == "unemployed" {
		factors = append(factors, "Current unemployment")
	}
	if score >= 700 {
		factors = append(factors, "Strong payment history")
		factors = append(factors, "Good financial stability")
	}
	
	return factors
}

func (s *CreditScoringService) generateRecommendation(score int) string {
	if score >= 740 {
		return "Excellent credit profile. Eligible for best rates and terms."
	} else if score >= 670 {
		return "Good credit profile. Eligible for competitive rates."
	} else if score >= 580 {
		return "Fair credit profile. May need additional documentation."
	}
	return "Credit profile needs improvement. Consider secured products."
}

func generateID() string {
	return fmt.Sprintf("cs_%d", time.Now().UnixNano())
}
