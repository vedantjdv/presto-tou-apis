package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/presto-tou-apis/internal/db"
	"github.com/presto-tou-apis/internal/models"
	"github.com/presto-tou-apis/internal/repository"
	"github.com/presto-tou-apis/internal/service"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) (*repository.Repository, *service.Service) {
	// Set the test DB URL if not set. Use 5439 as per docker-compose mapping
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://postgres:root@localhost:5439/tou_db?sslmode=disable")
	}

	pool, err := db.Connect()
	if err != nil {
		t.Fatalf("Could not connect to test database: %v", err)
	}

	// Clean up tables for a fresh test
	ctx := context.Background()
	_, _ = pool.Exec(ctx, "TRUNCATE TABLE charger_schedules, intervals, chargers, schedules RESTART IDENTITY CASCADE")

	repo := repository.NewRepository(pool)
	svc := service.NewService(repo)
	return repo, svc
}

func TestGetPrice_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo, svc := setupTestDB(t)
	h := NewHandler(svc)
	r := gin.Default()
	r.GET("/v1/chargers/:id/price", h.GetPrice)

	ctx := context.Background()

	// 1. Create a charger in NY
	charger := &models.Charger{Name: "NY-Charger", Timezone: "America/New_York"}
	err := repo.CreateCharger(ctx, charger)
	assert.NoError(t, err)

	// 2. Create a schedule with a specific price
	schedule := &models.Schedule{
		Name: "Peak Plan",
		Intervals: []models.Interval{
			{
				StartTime:   "00:00:00",
				EndTime:     "23:59:59",
				PricePerKWh: 0.99,
				DaysOfWeek:  127, // All days
			},
		},
	}
	err = repo.CreateSchedule(ctx, schedule)
	assert.NoError(t, err)

	// 3. Assign schedule to charger
	err = repo.AssignScheduleToCharger(ctx, charger.ID, schedule.ID)
	assert.NoError(t, err)

	t.Run("Get price successfully", func(t *testing.T) {
		// Midday in NY is roughly evening in UTC
		req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/chargers/%d/price?timestamp=2024-05-10T15:00:00Z", charger.ID), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.PriceResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, 0.99, resp.PricePerKWh)
	})

	t.Run("Invalid timestamp returns 400", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/chargers/%d/price?timestamp=not-a-date", charger.ID), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid timestamp format")
	})
}

func TestCreateCharger_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, svc := setupTestDB(t)
	h := NewHandler(svc)
	r := gin.Default()
	r.POST("/v1/chargers", h.CreateCharger)

	t.Run("Create valid charger", func(t *testing.T) {
		charger := models.Charger{Name: "LA-Charger", Timezone: "America/Los_Angeles"}
		body, _ := json.Marshal(charger)
		req, _ := http.NewRequest("POST", "/v1/chargers", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Reject invalid timezone", func(t *testing.T) {
		charger := models.Charger{Name: "Ghost-Charger", Timezone: "Invalid/Zone"}
		body, _ := json.Marshal(charger)
		req, _ := http.NewRequest("POST", "/v1/chargers", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid timezone")
	})
}
