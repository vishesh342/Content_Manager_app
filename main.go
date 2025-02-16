package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vishesh342/content-manager/api"
)

const(
	dbSource = "postgres://root:secret@localhost:5432/contentdb?sslmode=disable"
	address = "0.0.0.0:8080"
)
func main(){
	config, err := pgxpool.ParseConfig(dbSource)
    if err != nil {
        log.Fatalf("Unable to parse connection string: %v\n", err)
    }
    pool, err := pgxpool.NewWithConfig(context.Background(), config)
    if err != nil {
        log.Fatalf("Unable to create connection pool: %v\n", err)
    }

	defer pool.Close()

    // Initialize the DBConnector with the connection pool

	server,err := api.NewServer(pool)
	if err != nil {
		log.Fatalf("Unable to create server: %v\n", err)
	}
	
    err = server.Run(address)

	if err != nil{
		log.Fatal("Cannot Start Server...",err)
	}
}