package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yertore/sub-aggregator/internal/model"
	"github.com/Yertore/sub-aggregator/internal/repository"
)

const dateLayout = "01-2006"

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req *model.CreateSubscriptionRequest) (*model.Subscription, error) {
	if req.ServiceName == "" {
		return nil, fmt.Errorf("service_name is required")
	}
	if req.Price <= 0 {
		return nil, fmt.Errorf("price must be greater than 0")
	}
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}

	sub := &model.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
	}

	if req.EndDate != "" {
		endDate, err := parseMonthYear(req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}
		if endDate.Before(startDate) {
			return nil, fmt.Errorf("end_date must be after start_date")
		}
		sub.EndDate = &endDate
	}

	return s.repo.Create(ctx, sub)
}

func (s *Service) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, userID, serviceName string) ([]*model.Subscription, error) {
	return s.repo.List(ctx, userID, serviceName)
}

func (s *Service) Update(ctx context.Context, id string, req *model.UpdateSubscriptionRequest) (*model.Subscription, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ServiceName != "" {
		existing.ServiceName = req.ServiceName
	}
	if req.Price != 0 {
		if req.Price < 0 {
			return nil, fmt.Errorf("price must be greater than 0")
		}
		existing.Price = req.Price
	}
	if req.StartDate != "" {
		startDate, err := parseMonthYear(req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date: %w", err)
		}
		existing.StartDate = startDate
	}
	if req.EndDate != "" {
		endDate, err := parseMonthYear(req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}
		if endDate.Before(existing.StartDate) {
			return nil, fmt.Errorf("end_date must be after start_date")
		}
		existing.EndDate = &endDate
	}

	return s.repo.Update(ctx, existing)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) TotalCost(ctx context.Context, userID, serviceName, from, to string) (int, error) {
	var fromDate, toDate string

	if from != "" {
		t, err := parseMonthYear(from)
		if err != nil {
			return 0, fmt.Errorf("invalid from: %w", err)
		}
		fromDate = t.Format("2006-01-02")
	}

	if to != "" {
		t, err := parseMonthYear(to)
		if err != nil {
			return 0, fmt.Errorf("invalid to: %w", err)
		}
		toDate = t.AddDate(0, 1, -1).Format("2006-01-02")
	}

	return s.repo.TotalCost(ctx, userID, serviceName, fromDate, toDate)
}

func parseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("expected MM-YYYY format, got %q", s)
	}
	return t, nil
}
