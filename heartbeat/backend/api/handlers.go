package api

import (
	"context"
	"encoding/json"
	"fmt"
	transit_realtime "heartbeat/internal/proto/transit_realtime"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
)

type Vehicle struct {
	trip_id       string
	stop_id       string
	timestamp     uint64
	status        string
	stop_sequence int32
}

type Trip struct {
	trip_id    string
	route_id   string
	start_date string
	start_time string
}

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

func LoadVehicles(
	conn *pgx.Conn,
	ctx context.Context,
	feed *transit_realtime.FeedMessage,
) (int64, error) {
	vehicles := GetCleanVehicles(feed)
	if len(vehicles) == 0 {
		return 0, nil
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	sqlStr := GenerateSQLInsert(
		"fct_vehicles",
		[]string{
			"trip_id",
			"stop_id",
			"timestamp",
			"status",
			"stop_sequence",
		},
		[]string{"trip_id", "stop_id", "timestamp"},
		len(vehicles),
	)
	args := buildArgs(vehicles, func(v Vehicle) []any {
		return []any{
			v.trip_id,
			v.stop_id,
			v.timestamp,
			v.status,
			v.stop_sequence,
		}
	})

	commandTag, err := tx.Exec(ctx, sqlStr, args...)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return commandTag.RowsAffected(), nil
}

func GetCleanVehicles(feed *transit_realtime.FeedMessage) []Vehicle {
	vehicles := []Vehicle{}

	for _, entity := range feed.Entity {
		if entity.Vehicle == nil {
			continue
		}

		// potentially nil vars
		var status string
		var stop_sequence int32

		if entity.Vehicle.CurrentStatus == nil {
			status = "UNKNOWN"
		} else {
			status = entity.Vehicle.CurrentStatus.String()
		}

		if entity.Vehicle.CurrentStopSequence == nil {
			stop_sequence = -1
		} else {
			stop_sequence = int32(*entity.Vehicle.CurrentStopSequence)
		}

		vehicle_clean := Vehicle{
			trip_id:       *entity.Vehicle.Trip.TripId,
			stop_id:       *entity.Vehicle.StopId,
			timestamp:     *entity.Vehicle.Timestamp,
			status:        status,
			stop_sequence: stop_sequence,
		}
		vehicles = append(vehicles, vehicle_clean)

	}
	return vehicles
}

func GetCleanTrips(feed *transit_realtime.FeedMessage) []Trip {
	trips := []Trip{}

	for _, entity := range feed.Entity {
		if entity.TripUpdate == nil {
			continue
		}
		trip_clean := Trip{
			trip_id:    *entity.TripUpdate.Trip.TripId,
			route_id:   *entity.TripUpdate.Trip.RouteId,
			start_date: *entity.TripUpdate.Trip.StartDate,
			start_time: *entity.TripUpdate.Trip.StartTime,
		}
		trips = append(trips, trip_clean)

	}
	return trips
}

func generatePlaceholders(nRows int, nCols int) string {
	placeholders := ""
	for i := range nRows {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "("
		for j := range nCols {
			if j > 0 {
				placeholders += ", "
			}
			placeholders += fmt.Sprintf("$%d", i*nCols+j+1)
		}
		placeholders += ")"
	}
	return placeholders
}

func joinColumns(columns []string) string {
	str := ""
	for i, col := range columns {
		if i > 0 {
			str += ", "
		}
		str += col
	}
	return str
}

func buildArgs[T any](records []T, mapper func(T) []any) []any {
	args := []any{}
	for _, rec := range records {
		args = append(args, mapper(rec)...)
	}
	return args
}

func GenerateSQLInsert(
	tableName string,
	cols []string,
	pkCols []string,
	nRows int,
) string {
	sqlStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES ",
		tableName,
		joinColumns(cols),
	)
	sqlStr += generatePlaceholders((nRows), len(cols))
	sqlStr += fmt.Sprintf("ON CONFLICT (%s) DO NOTHING", joinColumns(pkCols))
	return sqlStr
}

func LoadTrips(
	conn *pgx.Conn,
	ctx context.Context,
	feed *transit_realtime.FeedMessage,
) (int64, error) {
	trips := GetCleanTrips(feed)
	if len(trips) == 0 {
		return 0, nil
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	sqlStr := GenerateSQLInsert(
		"fct_trips",
		[]string{"trip_id", "route_id", "start_date", "start_time"},
		[]string{"trip_id", "route_id", "start_date"},
		len(trips),
	)
	args := buildArgs(trips, func(t Trip) []any {
		return []any{t.trip_id, t.route_id, t.start_date, t.start_time}
	})

	commandTag, err := tx.Exec(ctx, sqlStr, args...)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return commandTag.RowsAffected(), nil
}

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

func GetSubwayData(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feed)
}
