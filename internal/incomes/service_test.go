package incomes

import (
	"context"
	"testing"
	"time"
)

type fakeIncomeRepository struct {
	createDTO CreateIncomeDTO
	updateDTO UpdateIncomeDTO
}

func (r *fakeIncomeRepository) Create(_ context.Context, dto CreateIncomeDTO) (int, error) {
	r.createDTO = dto
	return 1, nil
}

func (r *fakeIncomeRepository) GetByID(_ context.Context, _ int) (*Income, error) {
	return nil, nil
}

func (r *fakeIncomeRepository) GetAll(_ context.Context, _ *time.Time, _ *time.Time) ([]Income, error) {
	return nil, nil
}

func (r *fakeIncomeRepository) GetByOrderID(_ context.Context, _ int) ([]Income, error) {
	return nil, nil
}

func (r *fakeIncomeRepository) Update(_ context.Context, _ int, dto UpdateIncomeDTO) error {
	r.updateDTO = dto
	return nil
}

func (r *fakeIncomeRepository) Delete(_ context.Context, _ int, _ *int) error {
	return nil
}

func TestServiceCreateValidatesDateAndAmount(t *testing.T) {
	repo := &fakeIncomeRepository{}
	service := NewService(repo)

	invalidDate := "06/01/2026"
	if _, err := service.Create(context.Background(), CreateIncomeDTO{
		OrderID: 1,
		Amount:  CustomFloat64(10),
		Date:    &invalidDate,
	}); err == nil || err.Error() != "invalid date (expected format: YYYY-MM-DD)" {
		t.Fatalf("expected invalid date error, got %v", err)
	}

	if _, err := service.Create(context.Background(), CreateIncomeDTO{
		OrderID: 1,
		Amount:  CustomFloat64(0),
	}); err == nil || err.Error() != "amount must be > 0" {
		t.Fatalf("expected amount error, got %v", err)
	}
}

func TestServiceUpdateValidatesAndNormalizesDate(t *testing.T) {
	repo := &fakeIncomeRepository{}
	service := NewService(repo)

	if err := service.Update(context.Background(), 1, UpdateIncomeDTO{
		OrderID: intPtr(0),
	}); err == nil || err.Error() != "order ID must be > 0" {
		t.Fatalf("expected order ID error, got %v", err)
	}

	paymentDate := " 2026-06-02 "
	if err := service.Update(context.Background(), 1, UpdateIncomeDTO{
		Date: &paymentDate,
	}); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if repo.updateDTO.Date == nil {
		t.Fatal("expected normalized date")
	}
	if *repo.updateDTO.Date != "2026-06-02" {
		t.Fatalf("expected normalized date, got %q", *repo.updateDTO.Date)
	}
}

func intPtr(value int) *int {
	return &value
}

var _ Repository = (*fakeIncomeRepository)(nil)
