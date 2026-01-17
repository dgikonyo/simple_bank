package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testQueries *Queries
var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	// Fallback to localhost if the variable isn't set (for local dev)
	dbHost := os.Getenv("DATABASE_HOST")
	if dbHost == "" {
		dbHost = "127.0.0.1" 
	}

	dbPort := os.Getenv("DATABASE_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	connStr := fmt.Sprintf("postgresql://root_user:root_secret@%s:%s/bank_db?sslmode=disable", dbHost, dbPort)

	// Add this log so you can see what's happening in the console
	fmt.Printf("Testing connection to: %s\n", connStr)

	var err error
	testDB, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)
	os.Exit(m.Run())
}