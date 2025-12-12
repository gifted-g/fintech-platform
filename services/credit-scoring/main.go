package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"credit-scoring/internal/config"
	"credit-scoring/internal/handler"
	"credit-scoring/internal/middleware"
	"credit-scoring/internal/repository"
	"credit-scoring/internal/service"
	"credit-scoring/pkg/database"
	"credit-scoring/pkg/kafka"
	"credit-scoring/pkg/logger"
	"credit-scoring/pkg/redis"
	"credit-scoring/pkg/tracing"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using system environment variables")
	}

	// Initialize logger
	log, err := logger.NewLogger("credit-scoring")
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize tracing
	shutdown, err := tracing.InitTracer("credit-scoring", cfg.JaegerEndpoint)
	if err != nil {
		log.Error("Failed to initialize tracer", zap.Error(err))
	}
	defer shutdown(context.Background())

	// Initialize database
	db, err := database.NewPostgresDB(cfg.DatabaseURL, cfg.DatabaseMaxConns)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := redis.NewRedisClient(cfg.RedisURL, cfg.RedisPassword)
	if err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatal("Failed to initialize Kafka producer", zap.Error(err))
	}
	defer kafkaProducer.Close()

	// Initialize repositories
	creditRepo := repository.NewCreditRepository(db)
	
	// Initialize services
	creditService := service.NewCreditScoringService(
		creditRepo,
		redisClient,
		kafkaProducer,
		log,
	)

	// Initialize handlers
	creditHandler := handler.NewCreditHandler(creditService, log)

	// Setup router
	router := setupRouter(creditHandler, log, cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Starting Credit Scoring Service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited")
}

func setupRouter(creditHandler *handler.CreditHandler, log *zap.Logger, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Global middleware
	router.Use(middleware.Logger(log))
	router.Use(middleware.Recovery(log))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.RateLimiter(100, 100)) // 100 requests per second

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "credit-scoring",
			"version": "1.0.0",
		})
	})

	// Metrics endpoint
	router.GET("/metrics", middleware.PrometheusHandler())

	// API routes
	v1 := router.Group("/api/v1")
	v1.Use(middleware.Auth(cfg.JWTSecret))
	{
		credit := v1.Group("/credit")
		{
			credit.POST("/score", creditHandler.CalculateScore)
			credit.GET("/score/:userId", creditHandler.GetScore)
			credit.GET("/history/:userId", creditHandler.GetHistory)
			credit.POST("/refresh/:userId", creditHandler.RefreshScore)
		}
	}

	return router
}
