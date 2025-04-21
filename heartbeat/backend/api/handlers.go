package api

import (
	"context"
	"fmt"
	transit_realtime "heartbeat/internal/proto/transit_realtime"
	"io"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
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

	sqlStr := generateSQLInsert(
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
func GetCleanTrips(feed *transit_realtime.FeedMessage) []Trip {
	trips := []Trip{}
	for _, entity := range feed.Entity {
		if entity == nil {
			continue
		}

		if entity.TripUpdate == nil {
			continue
		}

		var start_time string
		if entity.TripUpdate.Trip.StartTime == nil {
			start_time = "UNKNOWN"
		} else {
			start_time = *entity.TripUpdate.Trip.StartTime
		}

		trip_clean := Trip{
			trip_id:    *entity.TripUpdate.Trip.TripId,
			route_id:   *entity.TripUpdate.Trip.RouteId,
			start_date: *entity.TripUpdate.Trip.StartDate,
			start_time: start_time,
		}
		trips = append(trips, trip_clean)

	}
	return trips
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

	sqlStr := generateSQLInsert(
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
func GetSubwayDataByEndpoint(
	ctx context.Context,
	conn *pgx.Conn,
	endpoint SubwayEndpoint,
) error {
	feed, err := GetFeedMessage(SubwayEndpointMap[endpoint])
	if err != nil {
		return err
	}
	log.Printf(
		"Obtained %d entities from %s",
		len(feed.Entity),
		SubwayEndpointMap[endpoint],
	)

	var allRows int64 = 0
	rows, err := LoadTrips(conn, ctx, feed)
	if err != nil {
		return err
	}
	log.Printf("Loaded %d trips\n", rows)
	allRows += rows

	rows, err = LoadVehicles(conn, ctx, feed)
	if err != nil {
		return err
	}
	log.Printf("Loaded %d vehicles\n", rows)
	allRows += rows

	log.Printf(
		"Data loaded successfully from %s: %d rows",
		SubwayEndpointMap[endpoint],
		allRows,
	)
	return nil
}

func GetSubwayData(ctx context.Context, pool *pgxpool.Pool) error {
	eg, ctx := errgroup.WithContext(ctx)
	for endpoint, url := range SubwayEndpointMap {
		ep := endpoint
		url := url
		eg.Go(func() error {
			conn, err := pool.Acquire(ctx)
			if err != nil {
				return err
			}
			defer conn.Release()

			log.Printf("Loading data from %s", url)
			return GetSubwayDataByEndpoint(ctx, conn.Conn(), ep)
		})
	}
	return eg.Wait()
}

func buildArgs[T any](records []T, mapper func(T) []any) []any {
	args := []any{}
	for _, rec := range records {
		args = append(args, mapper(rec)...)
	}
	return args
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

func generateSQLInsert(
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
