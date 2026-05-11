package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/presto-tou-apis/internal/models"
	"github.com/presto-tou-apis/internal/service"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func validateTime(t string) error {
	_, err := time.Parse("15:04:05", t)
	return err
}

// GetPrice godoc
// @Summary Get price for a charger at a specific time
// @Description Returns the price per kWh for a given charger and timestamp. Defaults to now in UTC if no timestamp provided.
// @Tags chargers
// @Accept json
// @Produce json
// @Param id path int true "Charger ID"
// @Param timestamp query string false "Timestamp in RFC3339 format (e.g. 2024-05-10T15:00:00Z)"
// @Success 200 {object} models.PriceResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /v1/chargers/{id}/price [get]
func (h *Handler) GetPrice(c *gin.Context) {
	chargerIDStr := c.Param("id")
	chargerID, err := strconv.Atoi(chargerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid charger id"})
		return
	}

	timeParam := c.Query("timestamp")
	var timestamp time.Time
	if timeParam == "" {
		timestamp = time.Now().UTC()
	} else {
		timestamp, err = time.Parse(time.RFC3339, timeParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timestamp format, use RFC3339"})
			return
		}
	}

	resp, err := h.svc.GetPrice(c.Request.Context(), chargerID, timestamp)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateCharger godoc
// @Summary Create a new charger
// @Description Creates a new charger with the given name and timezone. See /v1/timezones for a list of valid timezones.
// @Tags chargers
// @Accept json
// @Produce json
// @Param charger body models.ChargerInput true "Charger object"
// @Success 201 {object} models.Charger
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/chargers [post]
func (h *Handler) CreateCharger(c *gin.Context) {
	var input models.ChargerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate timezone
	if _, err := time.LoadLocation(input.Timezone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timezone: " + input.Timezone + ". Use /v1/timezones to see samples."})
		return
	}

	charger := models.Charger{
		Name:     input.Name,
		Timezone: input.Timezone,
	}

	if err := h.svc.CreateCharger(c.Request.Context(), &charger); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, charger)
}

// ListTimezones godoc
// @Summary List common valid timezones
// @Description Returns a list of common valid timezones that can be used when creating a charger.
// @Tags chargers
// @Produce json
// @Success 200 {array} string
// @Router /v1/timezones [get]
func (h *Handler) ListTimezones(c *gin.Context) {
	commonTimezones := []string{
		"UTC",
		"America/New_York",
		"America/Chicago",
		"America/Denver",
		"America/Los_Angeles",
		"Europe/London",
		"Europe/Paris",
		"Asia/Tokyo",
		"Asia/Kolkata",
		"Australia/Sydney",
	}
	c.JSON(http.StatusOK, commonTimezones)
}

// CreateSchedule godoc
// @Summary Create a pricing schedule
// @Description Creates a new schedule with a list of intervals.
// @Tags schedules
// @Accept json
// @Produce json
// @Param schedule body models.ScheduleInput true "Schedule object"
// @Success 201 {object} models.Schedule
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/schedules [post]
func (h *Handler) CreateSchedule(c *gin.Context) {
	var input models.ScheduleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	intervals := make([]models.Interval, len(input.Intervals))
	for i, v := range input.Intervals {
		// Validation
		if err := validateTime(v.StartTime); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format (HH:MM:SS): " + v.StartTime})
			return
		}
		if err := validateTime(v.EndTime); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format (HH:MM:SS): " + v.EndTime})
			return
		}
		if v.PricePerKWh < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "price_per_kwh cannot be negative"})
			return
		}
		if v.DaysOfWeek < 0 || v.DaysOfWeek > 127 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "days_of_week must be between 0 and 127"})
			return
		}

		intervals[i] = models.Interval{
			StartTime:   v.StartTime,
			EndTime:     v.EndTime,
			PricePerKWh: v.PricePerKWh,
			DaysOfWeek:  v.DaysOfWeek,
		}
	}

	schedule := models.Schedule{
		Name:        input.Name,
		Description: input.Description,
		Intervals:   intervals,
	}

	if err := h.svc.CreateSchedule(c.Request.Context(), &schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

// ListSchedules godoc
// @Summary List all schedules
// @Description Returns a list of all available pricing schedules.
// @Tags schedules
// @Produce json
// @Success 200 {array} models.Schedule
// @Failure 500 {object} map[string]string
// @Router /v1/schedules [get]
func (h *Handler) ListSchedules(c *gin.Context) {
	schedules, err := h.svc.ListSchedules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, schedules)
}

// UpdateSchedule godoc
// @Summary Update a schedule
// @Description Updates the name, description, and intervals of an existing schedule.
// @Tags schedules
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Param schedule body models.ScheduleInput true "Schedule object"
// @Success 200 {object} models.Schedule
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/schedules/{id} [put]
func (h *Handler) UpdateSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}

	var input models.ScheduleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	intervals := make([]models.Interval, len(input.Intervals))
	for i, v := range input.Intervals {
		// Validation
		if err := validateTime(v.StartTime); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format (HH:MM:SS): " + v.StartTime})
			return
		}
		if err := validateTime(v.EndTime); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format (HH:MM:SS): " + v.EndTime})
			return
		}
		if v.PricePerKWh < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "price_per_kwh cannot be negative"})
			return
		}
		if v.DaysOfWeek < 0 || v.DaysOfWeek > 127 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "days_of_week must be between 0 and 127"})
			return
		}

		intervals[i] = models.Interval{
			StartTime:   v.StartTime,
			EndTime:     v.EndTime,
			PricePerKWh: v.PricePerKWh,
			DaysOfWeek:  v.DaysOfWeek,
		}
	}

	schedule := models.Schedule{
		ID:          id,
		Name:        input.Name,
		Description: input.Description,
		Intervals:   intervals,
	}

	if err := h.svc.UpdateSchedule(c.Request.Context(), &schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// AssignSchedule godoc
// @Summary Assign a schedule to a charger
// @Description Links a charger to a specific pricing schedule.
// @Tags chargers
// @Accept json
// @Produce json
// @Param id path int true "Charger ID"
// @Param body body map[string]int true "Schedule ID (e.g. {'schedule_id': 1})"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/chargers/{id}/schedule [post]
func (h *Handler) AssignSchedule(c *gin.Context) {
	chargerIDStr := c.Param("id")
	chargerID, err := strconv.Atoi(chargerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid charger id"})
		return
	}

	var req struct {
		ScheduleID int `json:"schedule_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.AssignSchedule(c.Request.Context(), chargerID, req.ScheduleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "schedule assigned successfully"})
}

// BulkAssignSchedule godoc
// @Summary Assign a schedule to multiple chargers
// @Description Links multiple chargers to a specific pricing schedule in one request.
// @Tags chargers
// @Accept json
// @Produce json
// @Param body body models.BulkScheduleInput true "Bulk assignment object"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/chargers/bulk-schedule [post]
func (h *Handler) BulkAssignSchedule(c *gin.Context) {
	var input models.BulkScheduleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.BulkAssignSchedule(c.Request.Context(), input.ChargerIDs, input.ScheduleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "schedules assigned successfully"})
}
