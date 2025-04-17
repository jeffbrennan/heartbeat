package main

import (
	"log"

	"heartbeat/api"
)

func main() {
	feed, err := api.GetFeedMessage(api.SubwayEndpointMap[api.BLUE])
	if err != nil {
		log.Fatalf("Error fetching subway data: %v", err)
	}
	log.Printf("Fetched subway data: %v", feed)

	// http.HandleFunc("/api/subway", api.GetSubwayData)
	// log.Println("Server starting at :8080...")
	// log.Fatal(http.ListenAndServe(":8080", nil))
}
