package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Yertore/sub-aggregator/internal/apperror"
	"github.com/Yertore/sub-aggregator/internal/model"
)

// mockRepo is a manual mock for the Repository interface.
type mockRepo struct {
	createFn    func(ctx context.Context, sub *model.Subscription) (*model.Subscription, error)
	getByIDFn   func(ctx context.Context, id string) (*model.Subscription, error)
	listFn      func(ctx context.Context, userID, serviceName string) ([]*model.Subscription, error)
	updateFn    func(ctx context.Context, sub *model.Subscription) (*model.Subscription, error)
	deleteFn    func(ctx context.Context, id string) error
	totalCostFn func(ctx context.Context, userID, serviceName, from, to string) (int, error)
}

func (m *mockRepo) Create(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	return m.createFn(ctx, sub)
}
func (m *mockRepo) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockRepo) List(ctx context.Context, userID, serviceName string) ([]*model.Subscription, error) {
	return m.listFn(ctx, userID, serviceName)
}
func (m *mockRepo) Update(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	return m.updateFn(ctx, sub)
}
func (m *mockRepo) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}
func (m *mockRepo) TotalCost(ctx context.Context, userID, serviceName, from, to string) (int, error) {
	return m.totalCostFn(ctx, userID, serviceName, from, to)
}

func TestCreate_Success(t *testing.T) {
	repo := &mockRepo{
		createFn: func(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
			sub.ID = "test-id"
			return sub, nil
		},
	}
	svc := New(repo)

	req := &model.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "07-2025",
	}

	result, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.ServiceName != "Yandex Plus" {
		t.Errorf("expected service_name 'Yandex Plus', got '%s'", result.ServiceName)
	}
	if result.Price != 400 {
		t.Errorf("expected price 400, got %d", result.Price)
	}
}

func TestCreate_InvalidPrice(t *testing.T) {
	svc := New(&mockRepo{})

	req := &model.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       -1,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "07-2025",
	}

	_, err := svc.Create(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for negative price, got nil")
	}
}

func TestCreate_InvalidStartDate(t *testing.T) {
	svc := New(&mockRepo{})

	req := &model.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "2025-07-01",
	}

	_, err := svc.Create(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for wrong date format, got nil")
	}
}

func TestCreate_EndDateBeforeStartDate(t *testing.T) {
	svc := New(&mockRepo{})

	req := &model.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "07-2025",
		EndDate:     "01-2025",
	}

	_, err := svc.Create(context.Background(), req)
	if err == nil {
		t.Fatal("expected error when end_date before start_date, got nil")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id string) (*model.Subscription, error) {
			return nil, apperror.ErrNotFound
		},
	}
	svc := New(repo)

	_, err := svc.GetByID(context.Background(), "non-existent-id")
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := &mockRepo{
		deleteFn: func(ctx context.Context, id string) error {
			return apperror.ErrNotFound
		},
	}
	svc := New(repo)

	err := svc.Delete(context.Background(), "non-existent-id")
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestTotalCost_InvalidDateFormat(t *testing.T) {
	svc := New(&mockRepo{})

	_, err := svc.TotalCost(context.Background(), "", "", "2025-01-01", "")
	if err == nil {
		t.Fatal("expected error for wrong date format, got nil")
	}
}

func TestTotalCost_Success(t *testing.T) {
	repo := &mockRepo{
		totalCostFn: func(ctx context.Context, userID, serviceName, from, to string) (int, error) {
			return 1200, nil
		},
	}
	svc := New(repo)

	total, err := svc.TotalCost(context.Background(), "", "", "01-2025", "12-2025")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if total != 1200 {
		t.Errorf("expected total 1200, got %d", total)
	}
}

func TestParseMonthYear(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		wantT   time.Time
	}{
		{"07-2025", false, time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)},
		{"01-2024", false, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"2025-07-01", true, time.Time{}},
		{"invalid", true, time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseMonthYear(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMonthYear(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && !got.Equal(tt.wantT) {
				t.Errorf("parseMonthYear(%q) = %v, want %v", tt.input, got, tt.wantT)
			}
		})
	}
}
