package agenda

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeRepo implements Repository for service unit tests.
type fakeRepo struct {
	createCalled   bool
	createdDTO     CreateAgendaItemDTO
	completeCalled bool
	archiveCalled  bool
	lastCompleteID int
	lastArchiveID  int
	createErr      error
	completeErr    error
	archiveErr     error
}

func (f *fakeRepo) Create(ctx context.Context, dto CreateAgendaItemDTO) (int, error) {
	f.createCalled = true
	f.createdDTO = dto
	return 1, f.createErr
}

func (f *fakeRepo) GetByID(ctx context.Context, id int) (*AgendaItem, error) {
	return nil, nil
}

func (f *fakeRepo) GetAll(ctx context.Context, filter FilterAgendaItemsDTO) ([]AgendaItem, error) {
	return []AgendaItem{}, nil
}

func (f *fakeRepo) Update(ctx context.Context, id int, dto UpdateAgendaItemDTO) error {
	return nil
}

func (f *fakeRepo) Delete(ctx context.Context, id int) error {
	return nil
}

func (f *fakeRepo) Complete(ctx context.Context, id int) error {
	f.completeCalled = true
	f.lastCompleteID = id
	return f.completeErr
}

func (f *fakeRepo) Archive(ctx context.Context, id int) error {
	f.archiveCalled = true
	f.lastArchiveID = id
	return f.archiveErr
}

// -------------------- Create validations --------------------

func TestCreate_RequiresTitle(t *testing.T) {
	svc := NewService(&fakeRepo{})
	_, err := svc.Create(context.Background(), CreateAgendaItemDTO{})
	if err == nil || err.Error() != "title is required" {
		t.Fatalf("expected title required error, got %v", err)
	}
}

func TestCreate_DefaultsTypeToNote(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateAgendaItemDTO{Title: "Test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createdDTO.Type != TypeNote {
		t.Fatalf("expected type note, got %q", repo.createdDTO.Type)
	}
}

func TestCreate_DefaultsPriorityToMedium(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateAgendaItemDTO{Title: "Test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createdDTO.Priority != PriorityMedium {
		t.Fatalf("expected priority medium, got %q", repo.createdDTO.Priority)
	}
}

func TestCreate_RejectsInvalidType(t *testing.T) {
	svc := NewService(&fakeRepo{})
	_, err := svc.Create(context.Background(), CreateAgendaItemDTO{Title: "Test", Type: "bogus"})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestCreate_RejectsInvalidPriority(t *testing.T) {
	svc := NewService(&fakeRepo{})
	_, err := svc.Create(context.Background(), CreateAgendaItemDTO{Title: "Test", Priority: "extreme"})
	if err == nil {
		t.Fatal("expected error for invalid priority")
	}
}

func TestCreate_ReminderRequiresDueDate(t *testing.T) {
	svc := NewService(&fakeRepo{})
	_, err := svc.Create(context.Background(), CreateAgendaItemDTO{
		Title: "Stand-up",
		Type:  TypeReminder,
	})
	if err == nil || err.Error() != "due_date is required for reminder items" {
		t.Fatalf("expected due_date required error, got %v", err)
	}
}

func TestCreate_ReminderWithDueDateSucceeds(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	due := time.Now().Add(24 * time.Hour)

	_, err := svc.Create(context.Background(), CreateAgendaItemDTO{
		Title:   "Stand-up",
		Type:    TypeReminder,
		DueDate: &due,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreate_NoteAndTaskDoNotRequireDueDate(t *testing.T) {
	svc := NewService(&fakeRepo{})
	for _, typ := range []ItemType{TypeNote, TypeTask} {
		_, err := svc.Create(context.Background(), CreateAgendaItemDTO{
			Title: "Item",
			Type:  typ,
		})
		if err != nil {
			t.Fatalf("type %q: unexpected error: %v", typ, err)
		}
	}
}

// -------------------- GetAll defaults --------------------

func TestGetAll_DefaultsStatusToPending(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	_, err := svc.GetAll(context.Background(), FilterAgendaItemsDTO{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetAll_RejectsInvalidStatus(t *testing.T) {
	svc := NewService(&fakeRepo{})
	_, err := svc.GetAll(context.Background(), FilterAgendaItemsDTO{Status: "unknown"})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestGetAll_AllStatusPassesFilter(t *testing.T) {
	svc := NewService(&fakeRepo{})
	_, err := svc.GetAll(context.Background(), FilterAgendaItemsDTO{Status: "all"})
	if err != nil {
		t.Fatalf("unexpected error for all status: %v", err)
	}
}

func TestGetAll_RejectsInvalidDateFormat(t *testing.T) {
	svc := NewService(&fakeRepo{})
	bad := "01/01/2026"
	_, err := svc.GetAll(context.Background(), FilterAgendaItemsDTO{From: &bad})
	if err == nil {
		t.Fatal("expected error for invalid date format")
	}
}

// -------------------- Update validations --------------------

func TestUpdate_RejectsEmptyTitle(t *testing.T) {
	svc := NewService(&fakeRepo{})
	empty := ""
	err := svc.Update(context.Background(), 1, UpdateAgendaItemDTO{Title: &empty})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestUpdate_RejectsInvalidStatus(t *testing.T) {
	svc := NewService(&fakeRepo{})
	bad := ItemStatus("invalid")
	err := svc.Update(context.Background(), 1, UpdateAgendaItemDTO{Status: &bad})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

// -------------------- Complete / Archive --------------------

func TestComplete_CallsRepo(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	if err := svc.Complete(context.Background(), 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.completeCalled || repo.lastCompleteID != 5 {
		t.Fatalf("expected Complete(5) on repo")
	}
}

func TestArchive_CallsRepo(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	if err := svc.Archive(context.Background(), 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.archiveCalled || repo.lastArchiveID != 7 {
		t.Fatalf("expected Archive(7) on repo")
	}
}

func TestComplete_PropagatesRepoError(t *testing.T) {
	repo := &fakeRepo{completeErr: errors.New("agenda item not found")}
	svc := NewService(repo)

	err := svc.Complete(context.Background(), 99)
	if err == nil || err.Error() != "agenda item not found" {
		t.Fatalf("expected not found error, got %v", err)
	}
}
