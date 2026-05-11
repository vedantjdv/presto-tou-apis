package models

import (
	"time"
)

type Charger struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Address     string    `json:"address" db:"address"`
	Timezone    string    `json:"timezone" db:"timezone"`
	CreatedDate time.Time `json:"created_date" db:"created_date"`
}

type ChargerInput struct {
	Name     string `json:"name" binding:"required"`
	Address  string `json:"address"`
	Timezone string `json:"timezone" binding:"required"`
}

type Schedule struct {
	ID          int        `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	Intervals   []Interval `json:"intervals,omitempty"`
	CreatedDate time.Time  `json:"created_date" db:"created_date"`
}

type ScheduleInput struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	Intervals   []IntervalInput `json:"intervals" binding:"required"`
}

type Interval struct {
	ID          int       `json:"id" db:"id"`
	ScheduleID  int       `json:"-" db:"schedule_id"`
	StartTime   string    `json:"start_time" db:"start_time"` // HH:MM:SS
	EndTime     string    `json:"end_time" db:"end_time"`     // HH:MM:SS
	PricePerKWh float64   `json:"price_per_kwh" db:"price_per_kwh"`
	DaysOfWeek  int       `json:"days_of_week" db:"days_of_week"`
	CreatedDate time.Time `json:"-" db:"created_date"`
}

type IntervalInput struct {
	StartTime   string  `json:"start_time" binding:"required"`
	EndTime     string  `json:"end_time" binding:"required"`
	PricePerKWh float64 `json:"price_per_kwh" binding:"required"`
	DaysOfWeek  int     `json:"days_of_week" binding:"required"`
}

type ChargerSchedule struct {
	ChargerID  int `json:"charger_id" db:"charger_id"`
	ScheduleID int `json:"schedule_id" db:"schedule_id"`
}

type PriceRequest struct {
	Timestamp time.Time `json:"timestamp" form:"timestamp"`
}

type PriceResponse struct {
	ChargerID   int       `json:"charger_id"`
	Timestamp   time.Time `json:"timestamp"`
	PricePerKWh float64   `json:"price_per_kwh"`
	Currency    string    `json:"currency"`
}

type BulkScheduleInput struct {
	ChargerIDs []int `json:"charger_ids" binding:"required"`
	ScheduleID int   `json:"schedule_id" binding:"required"`
}
