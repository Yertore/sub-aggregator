package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yertore/sub-aggregator/internal/apperror"
	"github.com/Yertore/sub-aggregator/internal/model"
)

type mockRepo struct {
	createFn    func(ctx context.Context, sub *model.Subscription) (*model.Subscription, error)
	getByIDFn   func(ctx context.Context, id string) (*model.Subscription, error)
	listFn      func(ctx context.Context, filter model.ListFilter) ([]*model.Subscription, error)
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
func (m *mockRepo) List(ctx context.Context, filter model.ListFilter) ([]*model.Subscription, error) {
	return m.listFn(ctx, filter)
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

// Create
func TestCreate_Success(t *testing.T) {
	svc := New(&mockRepo{
		createFn: func(_ context.Context, sub *model.Subscription) (*model.Subscription, error) {
			sub.ID = "test-id"
			return sub, nil
		},
	})

	result, err := svc.Create(context.Background(), &model.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "07-2025",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "test-id" {
		t.Errorf("expected id %q, got %q", "test-id", result.ID)
	}
}

func TestCreate_InvalidPrice(t *testing.T) {
	svc := New(&mockRepo{})

	_, err := svc.Create(context.Background(), &model.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       -1,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "07-2025",
	})
	if err == nil {
		t.Fatal("expected error for negative price")
	}
}

func TestCreate_InvalidStartDate(t *testing.T) {
	svc := New(&mockRepo{})

	_, err := svc.Create(context.Background(), &model.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "2025-07",
	})
	if err == nil {
		t.Fatal("expected error for invalid date format")
	}
}

func TestCreate_EndDateNotAfterStartDate(t *testing.T) {
	svc := New(&mockRepo{})

	_, err := svc.Create(context.Background(), &model.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "07-2025",
		EndDate:     "06-2025",
	})
	if err == nil {
		t.Fatal("expected error when end_date <= start_date")
	}
}

// GetByID
func TestGetByID_NotFound(t *testing.T) {
	svc := New(&mockRepo{
		getByIDFn: func(_ context.Context, _ string) (*model.Subscription, error) {
			return nil, apperror.ErrNotFound
		},
	})

	_, err := svc.GetByID(context.Background(), "non-existing-id")
	if err != apperror.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// List
func TestList_DefaultLimitApplied(t *testing.T) {
	svc := New(&mockRepo{
		listFn: func(_ context.Context, filter model.ListFilter) ([]*model.Subscription, error) {
			if filter.Limit != defaultLimit {
				t.Errorf("expected default limit %d, got %d", defaultLimit, filter.Limit)
			}
			return nil, nil
		},
	})

	// limit=0 should be replaced with defaultLimit
	_, _ = svc.List(context.Background(), "", "", 0, 0)
}

func TestList_MaxLimitEnforced(t *testing.T) {
	svc := New(&mockRepo{
		listFn: func(_ context.Context, filter model.ListFilter) ([]*model.Subscription, error) {
			if filter.Limit > maxLimit {
				t.Errorf("limit %d exceeds maxLimit %d", filter.Limit, maxLimit)
			}
			return nil, nil
		},
	})

	// limit=500 should be capped at maxLimit
	_, _ = svc.List(context.Background(), "", "", 500, 0)
}

// Delete
func TestDelete_NotFound(t *testing.T) {
	svc := New(&mockRepo{
		deleteFn: func(_ context.Context, _ string) error {
			return apperror.ErrNotFound
		},
	})

	err := svc.Delete(context.Background(), "non-existing-id")
	if err != apperror.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// Update
func TestUpdate_NewStartDateAfterExistingEndDate(t *testing.T) {
	endDate := mustParseDate("09-2025")
	svc := New(&mockRepo{
		getByIDFn: func(_ context.Context, _ string) (*model.Subscription, error) {
			return &model.Subscription{
				ID:          "test-id",
				ServiceName: "Yandex Plus",
				Price:       400,
				StartDate:   mustParseDate("07-2025"),
				EndDate:     &endDate,
			}, nil
		},
	})

	_, err := svc.Update(context.Background(), "test-id", &model.UpdateSubscriptionRequest{
		StartDate: "10-2025",
	})
	if err == nil {
		t.Fatal("expected error when new start_date is after existing end_date")
	}
}

func TestUpdate_BothDatesUpdatedCorrectly(t *testing.T) {
	svc := New(&mockRepo{
		getByIDFn: func(_ context.Context, _ string) (*model.Subscription, error) {
			return &model.Subscription{
				ID:          "test-id",
				ServiceName: "Yandex Plus",
				Price:       400,
				StartDate:   mustParseDate("07-2025"),
			}, nil
		},
		updateFn: func(_ context.Context, sub *model.Subscription) (*model.Subscription, error) {
			return sub, nil
		},
	})

	result, err := svc.Update(context.Background(), "test-id", &model.UpdateSubscriptionRequest{
		StartDate: "08-2025",
		EndDate:   "12-2025",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EndDate == nil {
		t.Fatal("expected end_date to be set")
	}
}

// TotalCost
func TestTotalCost_InvalidDateFormat(t *testing.T) {
	svc := New(&mockRepo{})

	_, err := svc.TotalCost(context.Background(), "", "", "2025-01", "")
	if err == nil {
		t.Fatal("expected error for invalid date format")
	}
}

func TestTotalCost_ToBeforeFrom(t *testing.T) {
	svc := New(&mockRepo{})

	_, err := svc.TotalCost(context.Background(), "", "", "06-2025", "01-2025")
	if err == nil {
		t.Fatal("expected error when to is before from")
	}
}

func TestTotalCost_CallsRepoWithParsedDates(t *testing.T) {
	svc := New(&mockRepo{
		totalCostFn: func(_ context.Context, _, _, from, to string) (int, error) {
			if from != "2025-01-01" {
				t.Errorf("expected from=2025-01-01, got %s", from)
			}
			if to != "2025-12-01" {
				t.Errorf("expected to=2025-12-01, got %s", to)
			}
			return 2400, nil
		},
	})

	total, err := svc.TotalCost(context.Background(), "", "", "01-2025", "12-2025")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2400 {
		t.Errorf("expected total=2400, got %d", total)
	}
}


func TestParseMonthYear_Valid(t *testing.T) {
	t1, err := parseMonthYear("07-2025")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if t1.Month() != time.July || t1.Year() != 2025 {
		t.Errorf("unexpected result: %v", t1)
	}
}

func TestParseMonthYear_Invalid(t *testing.T) {
	_, err := parseMonthYear("invalid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func mustParseDate(s string) time.Time {
	t, err := parseMonthYear(s)
	if err != nil {
		panic(err)
	}
	return t
}
