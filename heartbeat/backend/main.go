package main

import (
	"context"
	"fmt"
	"heartbeat/api"
	"os"

	"github.com/jackc/pgx/v5"
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
	api.GetSubwayData(ctx, conn)

}
