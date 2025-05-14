package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/ravindu/wallet-app-service/internal/config"
	"github.com/ravindu/wallet-app-service/internal/handler"
	"github.com/ravindu/wallet-app-service/internal/middleware"
	"github.com/ravindu/wallet-app-service/internal/repository"
	"github.com/ravindu/wallet-app-service/internal/usecase"
	"github.com/ravindu/wallet-app-service/pkg/database"
	"github.com/ravindu/wallet-app-service/pkg/logging"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to PostgreSQL
	db, err := database.NewPostgresDB(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL")
	
	// Connect to Redis
	var redisClient *redis.Client
	redisClient, err = database.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis, continuing without caching: %v", err)
		// Continue without Redis
	} else {
		defer redisClient.Close()
		log.Println("Connected to Redis")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Initialize use cases
	walletUsecase := usecase.NewWalletUsecase(userRepo, walletRepo, transactionRepo, redisClient)

	// Initialize handlers
	walletHandler := handler.NewWalletHandler(walletUsecase)

	// Set up router with middleware
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID) // Chi's built-in RequestID middleware
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))
	
	// Our custom RequestID middleware that checks for the Request-Id header
	r.Use(middleware.RequestID)
	
	// TODO: Uncomment to enable authentication
	// r.Use(middleware.AuthMiddleware)
	
	logger := logging.NewLogger()
	logger.Info(context.Background(), "Starting wallet application service")

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// TODO: Implement public routes (no auth required)
		// - Health check
		// - API documentation
		// - Authentication endpoints

		// TODO: Protected routes - require authentication
		// Wallet routes
		// Once auth is implemented, replace with:
		// r.Group(func(r chi.Router) {
		//     r.Use(middleware.AuthMiddleware)
		//     r.Post("/deposit", walletHandler.DepositHandler)
		//     r.Post("/withdraw", walletHandler.WithdrawHandler)
		//     r.Post("/transfer", walletHandler.TransferHandler)
		//     r.Get("/balance/{userID}", walletHandler.GetBalanceHandler)
		//     r.Get("/transactions/{userID}", walletHandler.GetTransactionHistoryHandler)
		// })

		// For now, routes are open without authentication
		r.Post("/deposit", walletHandler.DepositHandler)
		r.Post("/withdraw", walletHandler.WithdrawHandler)
		r.Post("/transfer", walletHandler.TransferHandler)
		r.Get("/balance/{userID}", walletHandler.GetBalanceHandler)
		r.Get("/transactions/{userID}", walletHandler.GetTransactionHistoryHandler)
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so it doesn't block
	go func() {
		log.Printf("Server listening on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}