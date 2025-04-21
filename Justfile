@proto:
    protoc \
    --proto_path=heartbeat/backend/internal/proto \
    --go_out=paths=source_relative:heartbeat/backend/internal/proto/transit_realtime \
    --go-grpc_out=paths=source_relative:heartbeat/backend/internal/proto \
    heartbeat/backend/internal/proto/gtfs-realtime.proto \
    heartbeat/backend/internal/proto/nyct-ext.proto

@dbinit:
    psql -U postgres -d heartbeat -a -f "heartbeat/database/init.sql"

@up:
    docker compose up