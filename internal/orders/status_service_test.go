package orders

import (
	"context"
	"errors"
	"testing"
)

// fakeStatusRepo implements StatusRepository for tests
type fakeStatusRepo struct {
	statuses      map[int]*OrderStatus
	nextID        int
	reorderCalled bool
	deactivated   []int
}

func newFakeStatusRepo(initial ...OrderStatus) *fakeStatusRepo {
	r := &fakeStatusRepo{statuses: map[int]*OrderStatus{}, nextID: 1}
	for i := range initial {
		s := initial[i]
		r.statuses[s.ID] = &s
		if s.ID >= r.nextID {
			r.nextID = s.ID + 1
		}
	}
	return r
}

func (r *fakeStatusRepo) Create(_ context.Context, dto CreateOrderStatusDTO) (int, error) {
	id := r.nextID
	r.nextID++
	color := dto.Color
	if color == "" {
		color = "#6B7280"
	}
	r.statuses[id] = &OrderStatus{
		ID:            id,
		Name:          dto.Name,
		DisplayName:   dto.DisplayName,
		Color:         color,
		OrderPosition: dto.OrderPosition,
		IsFinalStatus: dto.IsFinalStatus,
		IsActive:      true,
	}
	return id, nil
}

func (r *fakeStatusRepo) GetByID(_ context.Context, id int) (*OrderStatus, error) {
	return r.statuses[id], nil
}

func (r *fakeStatusRepo) GetByName(_ context.Context, name string) (*OrderStatus, error) {
	for _, s := range r.statuses {
		if s.Name == name {
			return s, nil
		}
	}
	return nil, nil
}

func (r *fakeStatusRepo) GetAll(_ context.Context, activeOnly bool) ([]OrderStatus, error) {
	var list []OrderStatus
	for _, s := range r.statuses {
		if !activeOnly || s.IsActive {
			list = append(list, *s)
		}
	}
	return list, nil
}

func (r *fakeStatusRepo) Update(_ context.Context, id int, dto UpdateOrderStatusDTO) error {
	s, ok := r.statuses[id]
	if !ok {
		return errors.New("order status not found")
	}
	if dto.DisplayName != nil {
		s.DisplayName = *dto.DisplayName
	}
	if dto.Color != nil {
		s.Color = *dto.Color
	}
	if dto.OrderPosition != nil {
		s.OrderPosition = *dto.OrderPosition
	}
	if dto.IsFinalStatus != nil {
		s.IsFinalStatus = *dto.IsFinalStatus
	}
	if dto.IsActive != nil {
		s.IsActive = *dto.IsActive
	}
	return nil
}

func (r *fakeStatusRepo) Deactivate(_ context.Context, id int) error {
	s, ok := r.statuses[id]
	if !ok {
		return errors.New("order status not found")
	}
	s.IsActive = false
	r.deactivated = append(r.deactivated, id)
	return nil
}

func (r *fakeStatusRepo) Reorder(_ context.Context, items []StatusOrderItem) error {
	r.reorderCalled = true
	for _, item := range items {
		if s, ok := r.statuses[item.ID]; ok {
			s.OrderPosition = item.Position
		}
	}
	return nil
}

func (r *fakeStatusRepo) CountOrdersForStatus(_ context.Context, statusID int) (int, error) {
	return 0, nil
}

// ---- Tests ----

func TestStatusCreate_ValidSlug(t *testing.T) {
	svc := NewStatusService(newFakeStatusRepo())
	id, err := svc.Create(context.Background(), CreateOrderStatusDTO{
		Name:          "in_progress",
		DisplayName:   "In Progress",
		Color:         "#F59E0B",
		OrderPosition: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}
}

func TestStatusCreate_InvalidSlug(t *testing.T) {
	svc := NewStatusService(newFakeStatusRepo())
	_, err := svc.Create(context.Background(), CreateOrderStatusDTO{
		Name:          "In Progress", // spaces/uppercase not allowed
		DisplayName:   "In Progress",
		OrderPosition: 2,
	})
	if err == nil {
		t.Fatal("expected validation error for invalid slug")
	}
}

func TestStatusCreate_DuplicateName(t *testing.T) {
	repo := newFakeStatusRepo(OrderStatus{ID: 1, Name: "in_progress", IsActive: true})
	svc := NewStatusService(repo)
	_, err := svc.Create(context.Background(), CreateOrderStatusDTO{
		Name:          "in_progress",
		DisplayName:   "In Progress",
		OrderPosition: 2,
	})
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
}

func TestStatusCreate_InvalidColor(t *testing.T) {
	svc := NewStatusService(newFakeStatusRepo())
	_, err := svc.Create(context.Background(), CreateOrderStatusDTO{
		Name:          "pending",
		DisplayName:   "Pending",
		Color:         "notacolor",
		OrderPosition: 2,
	})
	if err == nil {
		t.Fatal("expected error for invalid color")
	}
}

func TestStatusDeactivate_SystemStatusBlocked(t *testing.T) {
	repo := newFakeStatusRepo(OrderStatus{ID: 1, Name: "new", IsSystemStatus: true, IsActive: true})
	svc := NewStatusService(repo)
	err := svc.Deactivate(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error: system status should not be deactivatable")
	}
}

func TestStatusDeactivate_CustomStatus(t *testing.T) {
	repo := newFakeStatusRepo(OrderStatus{ID: 5, Name: "in_progress", IsSystemStatus: false, IsActive: true})
	svc := NewStatusService(repo)
	err := svc.Deactivate(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.statuses[5].IsActive {
		t.Fatal("expected status to be deactivated")
	}
}

func TestStatusUpdate_CannotDeactivateSystem(t *testing.T) {
	repo := newFakeStatusRepo(OrderStatus{ID: 1, Name: "new", IsSystemStatus: true, IsActive: true})
	svc := NewStatusService(repo)
	isActive := false
	err := svc.Update(context.Background(), 1, UpdateOrderStatusDTO{IsActive: &isActive})
	if err == nil {
		t.Fatal("expected error: cannot deactivate system status via Update")
	}
}

func TestStatusReorder_DuplicatePosition(t *testing.T) {
	svc := NewStatusService(newFakeStatusRepo())
	err := svc.Reorder(context.Background(), ReorderStatusesDTO{
		StatusOrders: []StatusOrderItem{
			{ID: 1, Position: 2},
			{ID: 2, Position: 2}, // duplicate
		},
	})
	if err == nil {
		t.Fatal("expected error for duplicate positions")
	}
}

func TestStatusReorder_ValidOrder(t *testing.T) {
	repo := newFakeStatusRepo(
		OrderStatus{ID: 1, Name: "new", OrderPosition: 1},
		OrderStatus{ID: 2, Name: "in_progress", OrderPosition: 2},
	)
	svc := NewStatusService(repo)
	err := svc.Reorder(context.Background(), ReorderStatusesDTO{
		StatusOrders: []StatusOrderItem{
			{ID: 1, Position: 2},
			{ID: 2, Position: 1},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.reorderCalled {
		t.Fatal("expected Reorder to be called on repo")
	}
}
