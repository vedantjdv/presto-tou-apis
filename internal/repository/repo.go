package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/presto-tou-apis/internal/models"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Chargers
func (r *Repository) CreateCharger(ctx context.Context, charger *models.Charger) error {
	query := `INSERT INTO chargers (name, timezone) VALUES ($1, $2) RETURNING id, created_date`
	return r.pool.QueryRow(ctx, query, charger.Name, charger.Timezone).Scan(&charger.ID, &charger.CreatedDate)
}

func (r *Repository) GetCharger(ctx context.Context, id int) (*models.Charger, error) {
	var charger models.Charger
	query := `SELECT id, name, timezone, created_date FROM chargers WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&charger.ID, &charger.Name, &charger.Timezone, &charger.CreatedDate)
	if err != nil {
		return nil, err
	}
	return &charger, nil
}

// Schedules
func (r *Repository) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO schedules (name, description) VALUES ($1, $2) RETURNING id, created_date`
	err = tx.QueryRow(ctx, query, schedule.Name, schedule.Description).Scan(&schedule.ID, &schedule.CreatedDate)
	if err != nil {
		return err
	}

	for i := range schedule.Intervals {
		interval := &schedule.Intervals[i]
		interval.ScheduleID = schedule.ID
		q := `INSERT INTO intervals (schedule_id, start_time, end_time, price_per_kwh, days_of_week) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id, created_date`
		err = tx.QueryRow(ctx, q, interval.ScheduleID, interval.StartTime, interval.EndTime, interval.PricePerKWh, interval.DaysOfWeek).Scan(&interval.ID, &interval.CreatedDate)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) ListSchedules(ctx context.Context) ([]models.Schedule, error) {
	query := `
		SELECT 
			s.id, s.name, s.description, s.created_date,
			i.id, i.start_time::TEXT, i.end_time::TEXT, i.price_per_kwh, i.days_of_week
		FROM schedules s
		LEFT JOIN intervals i ON s.id = i.schedule_id
		ORDER BY s.id, i.start_time
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.Schedule
	scheduleMap := make(map[int]*models.Schedule)

	for rows.Next() {
		var sID int
		var sName, sDesc string
		var sCreated time.Time
		var iID *int
		var iStart, iEnd *string
		var iPrice *float64
		var iDays *int

		err := rows.Scan(
			&sID, &sName, &sDesc, &sCreated,
			&iID, &iStart, &iEnd, &iPrice, &iDays,
		)
		if err != nil {
			return nil, err
		}

		s, ok := scheduleMap[sID]
		if !ok {
			newS := models.Schedule{
				ID:          sID,
				Name:        sName,
				Description: sDesc,
				CreatedDate: sCreated,
				Intervals:   []models.Interval{},
			}
			schedules = append(schedules, newS)
			s = &schedules[len(schedules)-1]
			scheduleMap[sID] = s
		}

		if iID != nil {
			s.Intervals = append(s.Intervals, models.Interval{
				ID:          *iID,
				ScheduleID:  sID,
				StartTime:   *iStart,
				EndTime:     *iEnd,
				PricePerKWh: *iPrice,
				DaysOfWeek:  *iDays,
			})
		}
	}

	return schedules, nil
}

func (r *Repository) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `UPDATE schedules SET name = $1, description = $2 WHERE id = $3`
	_, err = tx.Exec(ctx, query, schedule.Name, schedule.Description, schedule.ID)
	if err != nil {
		return err
	}

	// For simplicity, delete old intervals and insert new ones
	_, err = tx.Exec(ctx, `DELETE FROM intervals WHERE schedule_id = $1`, schedule.ID)
	if err != nil {
		return err
	}

	for i := range schedule.Intervals {
		interval := &schedule.Intervals[i]
		interval.ScheduleID = schedule.ID
		q := `INSERT INTO intervals (schedule_id, start_time, end_time, price_per_kwh, days_of_week) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id, created_date`
		err = tx.QueryRow(ctx, q, interval.ScheduleID, interval.StartTime, interval.EndTime, interval.PricePerKWh, interval.DaysOfWeek).Scan(&interval.ID, &interval.CreatedDate)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// Charger Schedules
func (r *Repository) AssignScheduleToCharger(ctx context.Context, chargerID, scheduleID int) error {
	query := `INSERT INTO charger_schedules (charger_id, schedule_id) VALUES ($1, $2)
              ON CONFLICT (charger_id, schedule_id) DO NOTHING`
	_, err := r.pool.Exec(ctx, query, chargerID, scheduleID)
	return err
}

func (r *Repository) BulkAssignScheduleToChargers(ctx context.Context, chargerIDs []int, scheduleID int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Optional: Clear existing schedules for these chargers if we want 1-to-1
	// For now, let's just insert/ignore to match the current pattern
	query := `INSERT INTO charger_schedules (charger_id, schedule_id) VALUES ($1, $2)
              ON CONFLICT (charger_id, schedule_id) DO NOTHING`

	for _, id := range chargerIDs {
		_, err := tx.Exec(ctx, query, id, scheduleID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetPriceAtTime(ctx context.Context, chargerID int, localTime time.Time) (float64, error) {
	// 1. Get the schedule(s) assigned to the charger
	// 2. Find the interval that matches the local time and day of week
	
	// Convert day of week to bitmask: 1=Mon, 2=Tue, 4=Wed, 8=Thu, 16=Fri, 32=Sat, 64=Sun
	day := int(localTime.Weekday())
	// Go's Weekday: Sunday=0, Monday=1, ..., Saturday=6
	// We need: Mon=1, Tue=2, Wed=4, Thu=8, Fri=16, Sat=32, Sun=64
	bitmask := 0
	if day == 0 {
		bitmask = 64 // Sun
	} else {
		bitmask = 1 << (day - 1)
	}

	timeStr := localTime.Format("15:04:05")

	query := `
		SELECT i.price_per_kwh
		FROM intervals i
		JOIN charger_schedules cs ON i.schedule_id = cs.schedule_id
		WHERE cs.charger_id = $1
		  AND i.days_of_week & $2 > 0
		  AND i.start_time <= $3::TIME
		  AND i.end_time > $3::TIME
		ORDER BY i.created_date DESC
		LIMIT 1
	`
	var price float64
	err := r.pool.QueryRow(ctx, query, chargerID, bitmask, timeStr).Scan(&price)
	if err == pgx.ErrNoRows {
		return 0, fmt.Errorf("no pricing interval found for charger %d at %s", chargerID, timeStr)
	}
	return price, err
}
