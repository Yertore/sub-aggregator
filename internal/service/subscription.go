package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yertore/sub-aggregator/internal/model"
)

const (
	dateLayout   = "01-2006"
	defaultLimit = 20
	maxLimit     = 100
)

type Repository interface {
	Create(ctx context.Context, sub *model.Subscription) (*model.Subscription, error)
	GetByID(ctx context.Context, id string) (*model.Subscription, error)
	List(ctx context.Context, filter model.ListFilter) ([]*model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) (*model.Subscription, error)
	Delete(ctx context.Context, id string) error
	TotalCost(ctx context.Context, userID, serviceName, from, to string) (int, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
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
		if !endDate.After(startDate) {
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

func (s *Service) List(ctx context.Context, userID, serviceName string, limit, offset int) ([]*model.Subscription, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}

	filter := model.ListFilter{
		UserID:      userID,
		ServiceName: serviceName,
		Limit:       limit,
		Offset:      offset,
	}
	return s.repo.List(ctx, filter)
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

	newStart := existing.StartDate
	newEnd := existing.EndDate

	if req.StartDate != "" {
		parsed, err := parseMonthYear(req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date: %w", err)
		}
		newStart = parsed
	}
	if req.EndDate != "" {
		parsed, err := parseMonthYear(req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}
		newEnd = &parsed
	}

	if newEnd != nil && !newEnd.After(newStart) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	existing.StartDate = newStart
	existing.EndDate = newEnd

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
		toDate = t.Format("2006-01-02")
	}

	if fromDate != "" && toDate != "" {
		fromT, _ := time.Parse("2006-01-02", fromDate)
		toT, _ := time.Parse("2006-01-02", toDate)
		if toT.Before(fromT) {
			return 0, fmt.Errorf("to must be after from")
		}
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
