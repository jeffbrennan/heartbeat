-- 
-- dim_routes
-- 
drop table if exists dim_routes;
create table dim_routes (
    agency_id text,
    route_id text,
    route_short_name text,
    route_long_name text,
    route_type smallint,
    route_desc text,
    route_url text,
    route_color text,
    route_text_color text,
    primary key (agency_id, route_id)
);
\copy dim_routes from 'heartbeat/database/sources/mta/routes.txt' (FORMAT CSV, HEADER, DELIMITER(','));
-- 
-- dim_stops
-- 
drop table if exists dim_stops;
create table dim_stops (
    stop_id text,
    stop_name text,
    stop_lat real,
    stop_lon real,
    location_type smallint,
    parent_station text,
    primary key (stop_id)
);
\copy dim_stops from 'heartbeat/database/sources/mta/stops.txt' (FORMAT CSV, HEADER, DELIMITER(','));
-- 
--  dim_trips
-- 
drop table if exists dim_trips;
create table dim_trips (
    route_id text,
    trip_id text,
    service_id text,
    trip_headsign text,
    direction_id smallint,
    shape_id text,
    primary key (route_id, trip_id)
);
\copy dim_trips from 'heartbeat/database/sources/mta/trips.txt' (FORMAT CSV, HEADER, DELIMITER(','));
-- 
-- dim_shapes
-- 
drop table if exists dim_shapes;
create table dim_shapes (
    shape_id text,
    shape_pt_sequence smallint,
    shape_pt_lat real,
    shape_pt_lon real,
    primary key (shape_id, shape_pt_sequence)
);
\copy dim_shapes from 'heartbeat/database/sources/mta/shapes.txt' (FORMAT CSV, HEADER, DELIMITER(','));
-- 
-- dim_transfers
-- 
drop table if exists dim_transfers;
create table dim_transfers (
    from_stop_id text,
    to_stop_id text,
    transfer_type smallint,
    transfer_time smallint,
    primary key (from_stop_id, to_stop_id)
);
\copy dim_transfers from 'heartbeat/database/sources/mta/transfers.txt' (FORMAT CSV, HEADER, DELIMITER(','));