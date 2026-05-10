# EV Charger TOU Pricing API

A backend service for managing Time-of-Use (TOU) pricing for EV chargers. Built with Go, Gin, and PostgreSQL.

## Features
- **Normalized Schema**: Efficient storage of chargers, schedules, and pricing intervals.
- **Timezone Support**: Pricing lookups handle charger-specific timezones.
- **Flexible Intervals**: Support for recurring intervals based on days of the week.
- **RESTful API**: Easy integration with chargers and management tools.

## Prerequisites
- [Go 1.21+](https://golang.org/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)

## Getting Started

1. **Clone the repository**:
   ```bash
   git clone <repo-url>
   cd presto-tou-apis
   ```

2. **Start the database**:
   Ensure you have a PostgreSQL database running locally and update the `DATABASE_URL` in your `.env` file.

3. **Run the service**:
   ```bash
   go run cmd/server/main.go
   ```
   The server will start on `http://localhost:8071`.
   You can access the Swagger UI at `http://localhost:8071/swagger/index.html`.

## API Documentation

### 1. Create a Charger
`POST /v1/chargers`
```json
{
  "name": "Charger-A",
  "timezone": "America/Los_Angeles"
}
```

### 2. Create a Pricing Schedule
`POST /v1/schedules`
- `days_of_week` is a bitmask (1=Mon, 2=Tue, 4=Wed, 8=Thu, 16=Fri, 32=Sat, 64=Sun).
- Example: 127 = All days, 62 = Weekdays (Mon-Fri).
```json
{
  "name": "Peak Summer Schedule",
  "description": "High rates during summer afternoons",
  "intervals": [
    {
      "start_time": "00:00:00",
      "end_time": "12:00:00",
      "price_per_kwh": 0.15,
      "days_of_week": 127
    },
    {
      "start_time": "12:00:00",
      "end_time": "18:00:00",
      "price_per_kwh": 0.45,
      "days_of_week": 62
    },
    {
      "start_time": "18:00:00",
      "end_time": "23:59:59",
      "price_per_kwh": 0.20,
      "days_of_week": 127
    }
  ]
}
```

### 3. Assign Schedule to Charger
`POST /v1/chargers/:id/schedule`
```json
{
  "schedule_id": 1
}
```

### 4. Bulk Assign Schedule (New)
`POST /v1/chargers/bulk-schedule`
```json
{
  "charger_ids": [1, 2, 3],
  "schedule_id": 2
}
```

### 5. Get Pricing for a Charger
`GET /v1/chargers/:id/price?timestamp=2024-05-08T14:30:00Z`
- `timestamp` is optional (defaults to current time in UTC).

## Technical Implementation Details
- **Database**: PostgreSQL (relational schema).
- **Concurrency**: Safe for concurrent requests using Gin and PostgreSQL connection pooling.
- **Timezone Logic**: The service converts the incoming UTC timestamp to the charger's local time before matching intervals.
- **Bitmask Optimization**: Using bitmasks for `days_of_week` allows efficient filtering of recurring schedules in SQL.
- **Bulk Operations**: Optimized bulk assignment using database transactions to ensure consistency across multiple chargers.
