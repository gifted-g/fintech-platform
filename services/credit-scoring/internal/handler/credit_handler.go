package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"credit-scoring/internal/dto"
	"credit-scoring/internal/service"
	"credit-scoring/pkg/errors"
)

type CreditHandler struct {
	service *service.CreditScoringService
	logger  *zap.Logger
}

func NewCreditHandler(service *service.CreditScoringService, logger *zap.Logger) *CreditHandler {
	return &CreditHandler{
		service: service,
		logger:  logger,
	}
}

// CalculateScore calculates credit score for a user
func (h *CreditHandler) CalculateScore(c *gin.Context) {
	var req dto.CalculateScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, errors.NewAPIError("INVALID_REQUEST", err.Error()))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewAPIError("VALIDATION_ERROR", err.Error()))
		return
	}

	score, err := h.service.CalculateScore(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to calculate score", zap.Error(err), zap.String("userId", req.UserID))
		c.JSON(http.StatusInternalServerError, errors.NewAPIError("CALCULATION_ERROR", "Failed to calculate credit score"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Data:    score,
		Message: "Credit score calculated successfully",
	})
}

// GetScore retrieves the current credit score for a user
func (h *CreditHandler) GetScore(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, errors.NewAPIError("INVALID_REQUEST", "User ID is required"))
		return
	}

	score, err := h.service.GetScore(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get score", zap.Error(err), zap.String("userId", userID))
		c.JSON(http.StatusNotFound, errors.NewAPIError("NOT_FOUND", "Credit score not found"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Data:    score,
	})
}

// GetHistory retrieves credit score history for a user
func (h *CreditHandler) GetHistory(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, errors.NewAPIError("INVALID_REQUEST", "User ID is required"))
		return
	}

	history, err := h.service.GetHistory(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get history", zap.Error(err), zap.String("userId", userID))
		c.JSON(http.StatusInternalServerError, errors.NewAPIError("INTERNAL_ERROR", "Failed to retrieve history"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Data:    history,
	})
}

// RefreshScore forces a refresh of the credit score
func (h *CreditHandler) RefreshScore(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, errors.NewAPIError("INVALID_REQUEST", "User ID is required"))
		return
	}

	score, err := h.service.RefreshScore(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to refresh score", zap.Error(err), zap.String("userId", userID))
		c.JSON(http.StatusInternalServerError, errors.NewAPIError("REFRESH_ERROR", "Failed to refresh credit score"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Data:    score,
		Message: "Credit score refreshed successfully",
	})
}
