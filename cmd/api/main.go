package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"simple_bank/internal/db"
)

func main() {
	// 1. Load Configuration from Environment
	dbHost := getEnv("DATABASE_HOST", "127.0.0.1")
	dbPort := getEnv("DATABASE_PORT", "5432")
	dbUser := getEnv("DATABASE_USER", "root_user")
	dbPass := getEnv("DATABASE_PASSWORD", "root_secret")
	dbName := getEnv("DATABASE_NAME", "bank_db")

	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	// 2. Connect to Database with a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	connPool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer connPool.Close()

	// 3. Initialize SQLC Store
	queries := db.New(connPool)

	// 4. Setup Router
	mux := http.NewServeMux()

	// Health check route
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello! The Simple Bank API is running. ;-)")
	})

	// List Accounts Route
	mux.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		// Use the request context r.Context() so the DB query 
		// stops if the user closes their browser
		accounts, err := queries.ListAccounts(r.Context())
		if err != nil {
			log.Printf("Error fetching accounts: %v", err)
			http.Error(w, "Failed to fetch accounts", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "Fetched %d accounts", len(accounts))
	})

	// 5. Start Server
	serverAddr := ":8080"
	log.Printf("Starting server on %s...", serverAddr)
	
	server := &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

// Helper function to handle default values for environment variables
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}