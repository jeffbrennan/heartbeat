package main

import (
	"context"
	"fmt"
	"heartbeat/api"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	url := "postgresql://postgres:postgres@db:5432/heartbeat"
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse configuration: %v\n", err)
		os.Exit(1)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	var ticks int = 0
	ticker := time.NewTicker(30 * time.Second)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ticks++
			log.Printf("[tick %d] fetching updated subway data", ticks)
			err = api.GetSubwayData(ctx, pool)

			if err != nil {
				log.Fatal("Error fetching subway data: ", err)
			}

		case <-ctx.Done():
			log.Println("Exiting scheduler")
			return
		}
	}
}
