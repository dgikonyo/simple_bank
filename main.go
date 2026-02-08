package main

import (
	"context"
	"fmt"
	"log"
	"simple_bank/api"
	"simple_bank/internal/db"
	"simple_bank/util"

	"github.com/jackc/pgx/v5/pgxpool"
)


func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Cannot load config:", err)
	}

	// Add this log so you can see what's happening in the console
	fmt.Printf("Testing connection to: %s\n", config.DBSource)

	conn, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.StartServer(config.ServerAddress)
	if err != nil {
		log.Fatal("Cannot start server:", err)
	}

}
