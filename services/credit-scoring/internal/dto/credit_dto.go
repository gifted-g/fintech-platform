package dto

import (
	"fmt"
	"time"
)

type CalculateScoreRequest struct {
	UserID           string                 `json:"userId" binding:"required"`
	IncomeAmount     float64                `json:"incomeAmount" binding:"required,min=0"`
	EmploymentStatus string                 `json:"employmentStatus" binding:"required"`
	AccountAge       int                    `json:"accountAge" binding:"required,min=0"`
	TransactionData  map[string]interface{} `json:"transactionData"`
	LoanHistory      []LoanHistoryItem      `json:"loanHistory"`
}

type LoanHistoryItem struct {
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	PaymentDate time.Time `json:"paymentDate"`
}

func (r *CalculateScoreRequest) Validate() error {
	validStatuses := map[string]bool{
		"employed":     true,
		"self-employed": true,
		"unemployed":   true,
		"retired":      true,
	}

	if !validStatuses[r.EmploymentStatus] {
		return fmt.Errorf("invalid employment status: %s", r.EmploymentStatus)
	}

	if r.IncomeAmount < 0 {
		return fmt.Errorf("income amount cannot be negative")
	}

	return nil
}

type CreditScore struct {
	ID            string    `json:"id"`
	UserID        string    `json:"userId"`
	Score         int       `json:"score"`
	Grade         string    `json:"grade"`
	Factors       []string  `json:"factors"`
	Recommendation string   `json:"recommendation"`
	CalculatedAt  time.Time `json:"calculatedAt"`
	ExpiresAt     time.Time `json:"expiresAt"`
}

type CreditScoreHistory struct {
	UserID  string        `json:"userId"`
	History []CreditScore `json:"history"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}
