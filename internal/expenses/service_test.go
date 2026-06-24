package expenses

import (
	"context"
	"encoding/json"
	"testing"
)

type fakeExpenseRepository struct {
	comercios map[int]bool
	createDTO CreateExpenseDTO
	updateDTO UpdateExpenseDTO
}

func (r *fakeExpenseRepository) Create(_ context.Context, dto CreateExpenseDTO) (int, error) {
	r.createDTO = dto
	return 10, nil
}

func (r *fakeExpenseRepository) GetByID(_ context.Context, _ int) (*Expense, error) {
	return nil, nil
}

func (r *fakeExpenseRepository) GetAll(_ context.Context, _ *string, _ *string, _ *int) ([]Expense, error) {
	return nil, nil
}

func (r *fakeExpenseRepository) Update(_ context.Context, _ int, dto UpdateExpenseDTO) error {
	r.updateDTO = dto
	return nil
}

func (r *fakeExpenseRepository) Delete(_ context.Context, _ int, _ *int) error {
	return nil
}

func (r *fakeExpenseRepository) ComercioExists(_ context.Context, id int) (bool, error) {
	return r.comercios[id], nil
}

func (r *fakeExpenseRepository) CreateComercio(_ context.Context, _ CreateComercioDTO) (int, error) {
	return 1, nil
}

func (r *fakeExpenseRepository) GetComercios(_ context.Context) ([]Comercio, error) {
	return nil, nil
}

func (r *fakeExpenseRepository) UpdateComercio(_ context.Context, _ int, _ UpdateComercioDTO) error {
	return nil
}

func (r *fakeExpenseRepository) DeleteComercio(_ context.Context, _ int) error {
	return nil
}

func (r *fakeExpenseRepository) CreateProduct(_ context.Context, _ CreateProductDTO) (int, error) {
	return 1, nil
}

func (r *fakeExpenseRepository) GetProducts(_ context.Context, _ *int) ([]Product, error) {
	return nil, nil
}

func (r *fakeExpenseRepository) UpdateProduct(_ context.Context, _ int, _ UpdateProductDTO) error {
	return nil
}

func (r *fakeExpenseRepository) DeleteProduct(_ context.Context, _ int) error {
	return nil
}

func TestServiceCreateValidatesComercioAndItems(t *testing.T) {
	repo := &fakeExpenseRepository{comercios: map[int]bool{2: true}}
	service := NewService(repo)

	if _, err := service.Create(context.Background(), CreateExpenseDTO{
		Items: []CreateExpenseItemDTO{{ProductName: "Tela", Quantity: 1, UnitPrice: 1}},
	}); err == nil || err.Error() != "comercio ID is required and must be > 0" {
		t.Fatalf("expected comercio ID error, got %v", err)
	}

	if _, err := service.Create(context.Background(), CreateExpenseDTO{
		ComercioID: 99,
		Items:      []CreateExpenseItemDTO{{ProductName: "Tela", Quantity: 1, UnitPrice: 1}},
	}); err == nil || err.Error() != "comercio not found" {
		t.Fatalf("expected missing comercio error, got %v", err)
	}

	if _, err := service.Create(context.Background(), CreateExpenseDTO{
		ComercioID: 2,
	}); err == nil || err.Error() != "amount or expense items is required" {
		t.Fatalf("expected total source required error, got %v", err)
	}

	if _, err := service.Create(context.Background(), CreateExpenseDTO{
		ComercioID: 2,
		Items:      []CreateExpenseItemDTO{{ProductName: " ", Quantity: 1, UnitPrice: 1}},
	}); err == nil || err.Error() != "product name is required" {
		t.Fatalf("expected product name error, got %v", err)
	}

	expenseDate := " 2026-06-05 "
	description := " Compra "
	if _, err := service.Create(context.Background(), CreateExpenseDTO{
		ComercioID:  2,
		Date:        &expenseDate,
		Description: &description,
		Items: []CreateExpenseItemDTO{
			{ProductName: " Tela ", Quantity: 2, UnitPrice: 1200},
		},
	}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if *repo.createDTO.Date != "2026-06-05" {
		t.Fatalf("expected normalized date, got %q", *repo.createDTO.Date)
	}
	if *repo.createDTO.Description != "Compra" {
		t.Fatalf("expected trimmed description, got %q", *repo.createDTO.Description)
	}
	if repo.createDTO.Items[0].ProductName != "Tela" {
		t.Fatalf("expected trimmed product name, got %q", repo.createDTO.Items[0].ProductName)
	}
}

func TestCreateExpenseItemDTORejectsDecimalQuantity(t *testing.T) {
	var dto CreateExpenseItemDTO
	err := json.Unmarshal([]byte(`{"product_name":"Tela","quantity":1.9,"unit_price":1200}`), &dto)
	if err == nil || err.Error() != "value must be an integer" {
		t.Fatalf("expected integer quantity error, got %v", err)
	}
}

func TestServiceCreateAcceptsSimpleExpense(t *testing.T) {
	repo := &fakeExpenseRepository{comercios: map[int]bool{2: true}}
	service := NewService(repo)
	amount := CustomFloat64(1250)

	if _, err := service.Create(context.Background(), CreateExpenseDTO{
		ComercioID: 2,
		Amount:     &amount,
	}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if repo.createDTO.Amount == nil || *repo.createDTO.Amount != amount {
		t.Fatalf("expected amount to be preserved, got %#v", repo.createDTO.Amount)
	}
	if len(repo.createDTO.Items) != 0 {
		t.Fatalf("expected no items for simple expense, got %#v", repo.createDTO.Items)
	}
}

func TestServiceCreateRejectsInvalidSimpleExpenseAmount(t *testing.T) {
	repo := &fakeExpenseRepository{comercios: map[int]bool{2: true}}
	service := NewService(repo)
	zeroAmount := CustomFloat64(0)
	negativeAmount := CustomFloat64(-1)

	if _, err := service.Create(context.Background(), CreateExpenseDTO{
		ComercioID: 2,
		Amount:     &zeroAmount,
	}); err == nil || err.Error() != "amount must be > 0" {
		t.Fatalf("expected amount error, got %v", err)
	}

	if _, err := service.Create(context.Background(), CreateExpenseDTO{
		ComercioID: 2,
		Amount:     &negativeAmount,
	}); err == nil || err.Error() != "amount must be > 0" {
		t.Fatalf("expected amount error, got %v", err)
	}
}

func TestServiceCreateRejectsAmountAndItemsTogether(t *testing.T) {
	repo := &fakeExpenseRepository{comercios: map[int]bool{2: true}}
	service := NewService(repo)
	amount := CustomFloat64(1250)

	if _, err := service.Create(context.Background(), CreateExpenseDTO{
		ComercioID: 2,
		Amount:     &amount,
		Items:      []CreateExpenseItemDTO{{ProductName: "Tela", Quantity: 1, UnitPrice: 1}},
	}); err == nil || err.Error() != "use either amount or expense items, not both" {
		t.Fatalf("expected mutually exclusive total source error, got %v", err)
	}
}

func TestServiceUpdateValidatesExpenseFields(t *testing.T) {
	repo := &fakeExpenseRepository{comercios: map[int]bool{3: true}}
	service := NewService(repo)

	if err := service.Update(context.Background(), 0, UpdateExpenseDTO{}); err == nil || err.Error() != "expense ID is required and must be > 0" {
		t.Fatalf("expected expense ID error, got %v", err)
	}

	if err := service.Update(context.Background(), 1, UpdateExpenseDTO{
		ComercioID: 3,
		Items:      []CreateExpenseItemDTO{{ProductName: "Tela", Quantity: 0, UnitPrice: 1}},
	}); err == nil || err.Error() != "quantity must be > 0" {
		t.Fatalf("expected quantity error, got %v", err)
	}

	if err := service.Update(context.Background(), 1, UpdateExpenseDTO{
		ComercioID: 3,
		Items:      []CreateExpenseItemDTO{{ProductName: "Tela", Quantity: 1, UnitPrice: 0}},
	}); err == nil || err.Error() != "unit price must be > 0" {
		t.Fatalf("expected unit price error, got %v", err)
	}
}

func TestServiceFiltersValidateComercioID(t *testing.T) {
	service := NewService(&fakeExpenseRepository{})

	badComercio := "abc"
	if _, err := service.GetAll(context.Background(), nil, nil, &badComercio); err == nil || err.Error() != "invalid comercio_id" {
		t.Fatalf("expected invalid comercio error, got %v", err)
	}
}

var _ Repository = (*fakeExpenseRepository)(nil)
