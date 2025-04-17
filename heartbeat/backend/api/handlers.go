package api

import (
	"encoding/json"
	transit_realtime "heartbeat/internal/proto/transit_realtime"
	"io"
	"net/http"

	"google.golang.org/protobuf/proto"
)

type SubwayEndpoint int

const (
	BLUE SubwayEndpoint = iota
	ORANGE
	GREEN
	BROWN
	YELLOW
	GRAY
	NUMBERED
)

var SubwayEndpointMap = map[SubwayEndpoint]string{
	BLUE:     "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-ace",
	ORANGE:   "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-bdfm",
	GREEN:    "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-g",
	BROWN:    "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-jz",
	YELLOW:   "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-nqrw",
	GRAY:     "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-l",
	NUMBERED: "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs",
}

// GetFeedMessage queries the given URL, reads the GTFS-realtime binary data,
// and unmarshals it into a FeedMessage defined in the generated proto file.
func GetFeedMessage(url string) (*transit_realtime.FeedMessage, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	feed := &transit_realtime.FeedMessage{}
	if err := proto.Unmarshal(data, feed); err != nil {
		return nil, err
	}
	return feed, nil
}

// GetSubwayData is an HTTP handler that fetches one of the subway endpoints,
// parses the GTFS-realtime feed into a protobuf struct, and returns data as JSON.
func GetSubwayData(w http.ResponseWriter, r *http.Request) {
	// For example, using the BLUE endpoint.
	url := SubwayEndpointMap[BLUE]

	feed, err := GetFeedMessage(url)
	if err != nil {
		http.Error(
			w,
			"Error fetching feed: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// For demonstration, returns the entire feed object as JSON.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feed)
}
