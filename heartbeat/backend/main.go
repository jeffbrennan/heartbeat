package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"

	"heartbeat/api"
)

func main() {
	ctx := context.Background()
	url := "postgresql://localhost:5432/heartbeat"
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	feed, err := api.GetFeedMessage(api.SubwayEndpointMap[api.BLUE])
	if err != nil {
		panic(err)
	}

	var allRows int64 = 0
	rows, err := api.LoadTrips(conn, ctx, feed)
	if err != nil {
		log.Printf("Unable to load trips: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Loaded %d trips\n", rows)
	allRows += rows

	rows, err = api.LoadVehicles(conn, ctx, feed)
	if err != nil {
		log.Printf("Unable to load vehicles: %v\n", err)
		os.Exit(1)
	}
	log.Printf("Loaded %d vehicles\n", rows)
	allRows += rows

	log.Printf("Data loaded successfully: %d rows", allRows)
}
