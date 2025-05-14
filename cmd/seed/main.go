package main

import (
	"context"
	"log"
	"time"

	"github.com/ravindu/wallet-app-service/internal/config"
	"github.com/ravindu/wallet-app-service/internal/domain"
	"github.com/ravindu/wallet-app-service/pkg/database"
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

	ctx := context.Background()

	// Create test users
	users := []struct {
		username string
		email    string
	}{
		{"alice", "alice@example.com"},
		{"bob", "bob@example.com"},
		{"charlie", "charlie@example.com"},
	}

	for _, u := range users {
		now := time.Now()
		
		// Insert user
		var userID int64
		err := db.QueryRow(ctx, `
			INSERT INTO users (username, email, created_at, updated_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (username) DO UPDATE SET email = $2
			RETURNING id
		`, u.username, u.email, now, now).Scan(&userID)
		
		if err != nil {
			log.Printf("Error creating user %s: %v", u.username, err)
			continue
		}
		
		// Check if wallet exists
		var walletExists bool
		err = db.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM wallets WHERE user_id = $1)
		`, userID).Scan(&walletExists)
		
		if err != nil {
			log.Printf("Error checking wallet for user %s: %v", u.username, err)
			continue
		}
		
		// Create wallet if it doesn't exist
		if !walletExists {
			_, err = db.Exec(ctx, `
				INSERT INTO wallets (user_id, balance, currency, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5)
			`, userID, 1000.00, string(domain.USD), now, now)
			
			if err != nil {
				log.Printf("Error creating wallet for user %s: %v", u.username, err)
				continue
			}
		}
		
		log.Printf("Created/updated user %s with ID %d", u.username, userID)
	}

	log.Println("Database seeding completed")
}