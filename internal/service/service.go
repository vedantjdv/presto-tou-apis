package service

import (
	"context"
	"fmt"
	"time"

	"github.com/presto-tou-apis/internal/models"
	"github.com/presto-tou-apis/internal/repository"
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetPrice(ctx context.Context, chargerID int, utcTimestamp time.Time) (*models.PriceResponse, error) {
	charger, err := s.repo.GetCharger(ctx, chargerID)
	if err != nil {
		return nil, fmt.Errorf("error fetching charger: %v", err)
	}

	location, err := time.LoadLocation(charger.Timezone)
	if err != nil {
		// Fallback to UTC if timezone is invalid
		location = time.UTC
	}

	localTime := utcTimestamp.In(location)
	price, err := s.repo.GetPriceAtTime(ctx, chargerID, localTime)
	if err != nil {
		return nil, err
	}

	return &models.PriceResponse{
		ChargerID:   chargerID,
		Timestamp:   utcTimestamp,
		PricePerKWh: price,
		Currency:    "USD", // Default
	}, nil
}

func (s *Service) CreateCharger(ctx context.Context, charger *models.Charger) error {
	// Validate timezone
	if _, err := time.LoadLocation(charger.Timezone); err != nil {
		return fmt.Errorf("invalid timezone: %v", err)
	}
	return s.repo.CreateCharger(ctx, charger)
}

func (s *Service) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	return s.repo.CreateSchedule(ctx, schedule)
}

func (s *Service) ListSchedules(ctx context.Context) ([]models.Schedule, error) {
	return s.repo.ListSchedules(ctx)
}

func (s *Service) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	return s.repo.UpdateSchedule(ctx, schedule)
}

func (s *Service) AssignSchedule(ctx context.Context, chargerIDs []int, scheduleID int) error {
	return s.repo.AssignScheduleToChargers(ctx, chargerIDs, scheduleID)
}
