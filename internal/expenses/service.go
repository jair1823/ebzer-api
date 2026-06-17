package expenses

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
)

type Service interface {
	Create(ctx context.Context, dto CreateExpenseDTO) (int, error)
	GetByID(ctx context.Context, id int) (*Expense, error)
	GetAll(ctx context.Context, from, to, comercioID *string) ([]Expense, error)
	Update(ctx context.Context, id int, dto UpdateExpenseDTO) error
	Delete(ctx context.Context, id int, actorID *int) error
	CreateComercio(ctx context.Context, dto CreateComercioDTO) (int, error)
	GetComercios(ctx context.Context) ([]Comercio, error)
	UpdateComercio(ctx context.Context, id int, dto UpdateComercioDTO) error
	DeleteComercio(ctx context.Context, id int) error
	CreateProduct(ctx context.Context, dto CreateProductDTO) (int, error)
	GetProducts(ctx context.Context, comercioID *string) ([]Product, error)
	UpdateProduct(ctx context.Context, id int, dto UpdateProductDTO) error
	DeleteProduct(ctx context.Context, id int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, dto CreateExpenseDTO) (int, error) {
	if err := s.normalizeExpense(ctx, &dto.ComercioID, &dto.Date, &dto.Description, dto.Items); err != nil {
		return 0, err
	}

	id, err := s.repo.Create(ctx, dto)
	if isUniqueConstraintError(err) {
		return 0, errors.New("product name already exists for comercio")
	}
	return id, err
}

func (s *service) GetByID(ctx context.Context, id int) (*Expense, error) {
	if id <= 0 {
		return nil, errors.New("expense ID is required and must be > 0")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetAll(ctx context.Context, fromStr, toStr, comercioIDStr *string) ([]Expense, error) {
	from, err := normalizeFilterDate(fromStr, "from")
	if err != nil {
		return nil, err
	}
	to, err := normalizeFilterDate(toStr, "to")
	if err != nil {
		return nil, err
	}
	comercioID, err := parsePositiveID(comercioIDStr, "comercio_id")
	if err != nil {
		return nil, err
	}

	return s.repo.GetAll(ctx, from, to, comercioID)
}

func (s *service) Update(ctx context.Context, id int, dto UpdateExpenseDTO) error {
	if id <= 0 {
		return errors.New("expense ID is required and must be > 0")
	}
	if err := s.normalizeExpense(ctx, &dto.ComercioID, &dto.Date, &dto.Description, dto.Items); err != nil {
		return err
	}

	err := s.repo.Update(ctx, id, dto)
	if isUniqueConstraintError(err) {
		return errors.New("product name already exists for comercio")
	}
	return err
}

func (s *service) Delete(ctx context.Context, id int, actorID *int) error {
	if id <= 0 {
		return errors.New("expense ID is required and must be > 0")
	}
	return s.repo.Delete(ctx, id, actorID)
}

func (s *service) CreateComercio(ctx context.Context, dto CreateComercioDTO) (int, error) {
	name := strings.TrimSpace(dto.Name)
	if name == "" {
		return 0, errors.New("comercio name is required")
	}
	dto.Name = name
	dto.Description = normalizeOptionalText(dto.Description)

	id, err := s.repo.CreateComercio(ctx, dto)
	if isUniqueConstraintError(err) {
		return 0, errors.New("comercio name already exists")
	}
	return id, err
}

func (s *service) GetComercios(ctx context.Context) ([]Comercio, error) {
	return s.repo.GetComercios(ctx)
}

func (s *service) UpdateComercio(ctx context.Context, id int, dto UpdateComercioDTO) error {
	if id <= 0 {
		return errors.New("comercio ID is required and must be > 0")
	}
	if dto.Name != nil {
		name := strings.TrimSpace(*dto.Name)
		if name == "" {
			return errors.New("comercio name is required")
		}
		dto.Name = &name
	}
	dto.Description = normalizeOptionalText(dto.Description)

	err := s.repo.UpdateComercio(ctx, id, dto)
	if isUniqueConstraintError(err) {
		return errors.New("comercio name already exists")
	}
	return err
}

func (s *service) DeleteComercio(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("comercio ID is required and must be > 0")
	}
	return s.repo.DeleteComercio(ctx, id)
}

func (s *service) CreateProduct(ctx context.Context, dto CreateProductDTO) (int, error) {
	if err := s.normalizeProduct(ctx, &dto.ComercioID, &dto.Name, &dto.DefaultPrice); err != nil {
		return 0, err
	}
	id, err := s.repo.CreateProduct(ctx, dto)
	if isUniqueConstraintError(err) {
		return 0, errors.New("product name already exists for comercio")
	}
	return id, err
}

func (s *service) GetProducts(ctx context.Context, comercioIDStr *string) ([]Product, error) {
	comercioID, err := parsePositiveID(comercioIDStr, "comercio_id")
	if err != nil {
		return nil, err
	}
	return s.repo.GetProducts(ctx, comercioID)
}

func (s *service) UpdateProduct(ctx context.Context, id int, dto UpdateProductDTO) error {
	if id <= 0 {
		return errors.New("product ID is required and must be > 0")
	}
	if dto.ComercioID != nil {
		if *dto.ComercioID <= 0 {
			return errors.New("comercio ID must be > 0")
		}
		exists, err := s.repo.ComercioExists(ctx, *dto.ComercioID)
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("comercio not found")
		}
	}
	if dto.Name != nil {
		name := strings.TrimSpace(*dto.Name)
		if name == "" {
			return errors.New("product name is required")
		}
		dto.Name = &name
	}
	if dto.DefaultPrice != nil && *dto.DefaultPrice <= 0 {
		return errors.New("default price must be > 0")
	}

	err := s.repo.UpdateProduct(ctx, id, dto)
	if isUniqueConstraintError(err) {
		return errors.New("product name already exists for comercio")
	}
	return err
}

func (s *service) DeleteProduct(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("product ID is required and must be > 0")
	}
	return s.repo.DeleteProduct(ctx, id)
}

func (s *service) normalizeExpense(ctx context.Context, comercioID *int, date **string, description **string, items []CreateExpenseItemDTO) error {
	if comercioID == nil || *comercioID <= 0 {
		return errors.New("comercio ID is required and must be > 0")
	}
	exists, err := s.repo.ComercioExists(ctx, *comercioID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("comercio not found")
	}
	normalizedDate, err := normalizeExpenseDate(*date)
	if err != nil {
		return err
	}
	*date = normalizedDate
	*description = normalizeOptionalText(*description)
	if len(items) == 0 {
		return errors.New("at least one expense item is required")
	}
	for index := range items {
		productName := strings.TrimSpace(items[index].ProductName)
		if productName == "" {
			return errors.New("product name is required")
		}
		if items[index].ProductID != nil && *items[index].ProductID <= 0 {
			return errors.New("product ID must be > 0")
		}
		if items[index].Quantity <= 0 {
			return errors.New("quantity must be > 0")
		}
		if items[index].UnitPrice <= 0 {
			return errors.New("unit price must be > 0")
		}
		items[index].ProductName = productName
	}
	return nil
}

func (s *service) normalizeProduct(ctx context.Context, comercioID *int, name *string, defaultPrice *CustomFloat64) error {
	if comercioID == nil || *comercioID <= 0 {
		return errors.New("comercio ID is required and must be > 0")
	}
	exists, err := s.repo.ComercioExists(ctx, *comercioID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("comercio not found")
	}
	trimmed := strings.TrimSpace(*name)
	if trimmed == "" {
		return errors.New("product name is required")
	}
	if defaultPrice == nil || *defaultPrice <= 0 {
		return errors.New("default price must be > 0")
	}
	*name = trimmed
	return nil
}

func normalizeExpenseDate(date *string) (*string, error) {
	if date == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*date)
	if trimmed == "" {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return nil, errors.New("invalid date (expected format: YYYY-MM-DD)")
	}

	return &trimmed, nil
}

func normalizeFilterDate(date *string, name string) (*string, error) {
	if date == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*date)
	if trimmed == "" {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return nil, errors.New("invalid " + name + " date (expected format: YYYY-MM-DD)")
	}
	return &trimmed, nil
}

func parsePositiveID(value *string, name string) (*int, error) {
	if value == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}

	id, err := strconv.Atoi(trimmed)
	if err != nil || id <= 0 {
		return nil, errors.New("invalid " + name)
	}

	return &id, nil
}

func normalizeOptionalText(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
