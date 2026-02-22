package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"simple_bank/util"
	"testing"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testQueries *Queries
var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../../")
	if err != nil {
		log.Fatal("Cannot load config:", err)
	}
	// Add this log so you can see what's happening in the console
	fmt.Printf("Testing connection to: %s\n", config.DBSource)

	testDB, err = pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)
	os.Exit(m.Run())
}