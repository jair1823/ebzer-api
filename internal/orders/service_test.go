package orders

import (
	"context"
	"testing"
	"time"

	"creaciones-api/internal/incomes"
)

// fakeRepo implements Repository for tests
type fakeRepo struct {
	createCalled bool
	createdDTO   CreateOrderDTO
	finishCalled bool
}

func (f *fakeRepo) Create(ctx context.Context, dto CreateOrderDTO) (int, error) {
	f.createCalled = true
	f.createdDTO = dto
	return 42, nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id int) (*Order, error) {
	return nil, nil
}

func (f *fakeRepo) GetAll(ctx context.Context, statusID *int, from *time.Time, to *time.Time) ([]Order, error) {
	return nil, nil
}

func (f *fakeRepo) Update(ctx context.Context, id int, dto UpdateOrderDTO) error {
	return nil
}

func (f *fakeRepo) FinishOrder(ctx context.Context, id int) error {
	f.finishCalled = true
	return nil
}

func (f *fakeRepo) Delete(ctx context.Context, id int) error {
	return nil
}

// fakeIncomeRepo implements incomes.Repository for tests (minimal stubs)
type fakeIncomeRepo struct{}

func (f *fakeIncomeRepo) Create(ctx context.Context, dto incomes.CreateIncomeDTO) (int, error) {
	return 0, nil
}
func (f *fakeIncomeRepo) GetByID(ctx context.Context, id int) (*incomes.Income, error) { return nil, nil }
func (f *fakeIncomeRepo) GetAll(ctx context.Context, from *time.Time, to *time.Time) ([]incomes.Income, error) {
	return nil, nil
}
func (f *fakeIncomeRepo) GetByOrderID(ctx context.Context, orderID int) ([]incomes.Income, error) { return nil, nil }
func (f *fakeIncomeRepo) Update(ctx context.Context, id int, dto incomes.UpdateIncomeDTO) error { return nil }
func (f *fakeIncomeRepo) Delete(ctx context.Context, id int) error { return nil }

func TestCreate_DefaultsToNew(t *testing.T) {
	repo := &fakeRepo{}
	var incRepo fakeIncomeRepo
	svc := NewService(repo, &incRepo)

	dto := CreateOrderDTO{
		Description:   "Test order",
		AmountCharged: CustomFloat64(100),
		StatusID:      0, // not provided
	}

	id, err := svc.Create(context.Background(), dto)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if id != 42 {
		t.Fatalf("expected id 42, got %d", id)
	}
	if !repo.createCalled {
		t.Fatalf("expected Create to be called on repo")
	}
	if repo.createdDTO.StatusID != 1 {
		t.Fatalf("expected StatusID defaulted to 1 (new), got %d", repo.createdDTO.StatusID)
	}
}

func TestFinishOrder_CallsRepo(t *testing.T) {
	repo := &fakeRepo{}
	var incRepo fakeIncomeRepo
	svc := NewService(repo, &incRepo)

	err := svc.FinishOrder(context.Background(), 7)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !repo.finishCalled {
		t.Fatalf("expected FinishOrder to call repo.FinishOrder")
	}
}
