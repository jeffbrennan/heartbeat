# heartbeat

dashboard monitoring [real-time NYC transit data](https://api.mta.info/#/subwayRealTimeFeeds)

## Tech Stack

- [Go](https://go.dev/): backend
- [Metabase](https://github.com/metabase/metabase): visualization
- [Postgres](https://www.postgresql.org/): database
- [Docker](https://github.com/moby/moby): containerization

## Project Structure

```
heartbeat/
├── backend/
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── api/
│   │   └── handlers.go
│   └── models/
│       └── hospital_data.go
├── database/
│   ├── init.sql
│   └── docker-compose.yml
├── metabase/
│   └── docker-compose.override.yml
└── README.md
```

## Setup Instructions

1. Clone the repository:
   ```
   git clone <repository-url>
   cd heartbeat
   ```

2. Set up the database:
   - Navigate to the `database` directory and run:
     ```
     docker-compose up -d
     ```

3. Run the backend:
   - Navigate to the `backend` directory and run:
     ```
     go run main.go
     ```

4. Access the Metabase dashboard:
   - Open your browser and go to `http://localhost:3000`.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.