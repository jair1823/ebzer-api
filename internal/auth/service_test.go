package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type fakeRepository struct {
	users map[int]*User
	next  int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		users: map[int]*User{},
		next:  1,
	}
}

func (f *fakeRepository) Count(ctx context.Context) (int, error) {
	return len(f.users), nil
}

func (f *fakeRepository) Create(ctx context.Context, req CreateUserRequest, passwordHash string) (int, error) {
	id := f.next
	f.next++
	f.users[id] = &User{
		ID:           id,
		Name:         req.Name,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         req.Role,
		IsActive:     true,
	}
	return id, nil
}

func (f *fakeRepository) GetByID(ctx context.Context, id int) (*User, error) {
	user := f.users[id]
	if user == nil {
		return nil, nil
	}
	copy := *user
	return &copy, nil
}

func (f *fakeRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	for _, user := range f.users {
		if user.Email == email {
			copy := *user
			return &copy, nil
		}
	}
	return nil, nil
}

func (f *fakeRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	for _, user := range f.users {
		if user.Username == username {
			copy := *user
			return &copy, nil
		}
	}
	return nil, nil
}

func (f *fakeRepository) List(ctx context.Context) ([]User, error) {
	users := []User{}
	for _, user := range f.users {
		users = append(users, *user)
	}
	return users, nil
}

func (f *fakeRepository) Update(ctx context.Context, id int, req UpdateUserRequest, passwordHash *string) error {
	user := f.users[id]
	if user == nil {
		return errors.New("user not found")
	}
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if passwordHash != nil {
		user.PasswordHash = *passwordHash
	}
	return nil
}

func (f *fakeRepository) Deactivate(ctx context.Context, id int) error {
	user := f.users[id]
	if user == nil {
		return errors.New("user not found")
	}
	user.IsActive = false
	return nil
}

func newTestService(repo Repository) Service {
	cfg := Config{
		JWTSecret:  "test-secret",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	}
	return NewService(repo, NewJWTService(cfg.JWTSecret), cfg)
}

func TestServiceLoginAndRefresh(t *testing.T) {
	repo := newFakeRepository()
	svc := newTestService(repo)

	user, err := svc.CreateUser(context.Background(), CreateUserRequest{
		Name:     "Owner",
		Username: "Owner",
		Email:    "owner@example.com",
		Password: "password123",
		Role:     RoleAdmin,
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	login, err := svc.Login(context.Background(), LoginRequest{
		Username: user.Username,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if login.AccessToken == "" || login.RefreshToken == "" {
		t.Fatal("expected access and refresh tokens")
	}
	if login.User.Role != RoleAdmin {
		t.Fatalf("expected admin role, got %q", login.User.Role)
	}
	if login.User.Username != "owner" || login.User.Email != "owner@example.com" {
		t.Fatalf("unexpected login user: %#v", login.User)
	}

	refresh, err := svc.Refresh(context.Background(), login.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if refresh.AccessToken == "" {
		t.Fatal("expected refreshed access token")
	}
}

func TestServiceRejectsInvalidPasswordAndInactiveUser(t *testing.T) {
	repo := newFakeRepository()
	svc := newTestService(repo)

	user, err := svc.CreateUser(context.Background(), CreateUserRequest{
		Name:     "Operator",
		Username: "operator",
		Email:    "operator@example.com",
		Password: "password123",
		Role:     RoleOperator,
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	if _, err := svc.Login(context.Background(), LoginRequest{
		Username: user.Username,
		Password: "wrong-password",
	}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}

	if err := svc.DeactivateUser(context.Background(), user.ID); err != nil {
		t.Fatalf("DeactivateUser returned error: %v", err)
	}
	if _, err := svc.Login(context.Background(), LoginRequest{
		Username: user.Username,
		Password: "password123",
	}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected inactive user to be rejected, got %v", err)
	}
}

func TestJWTRejectsWrongTokenType(t *testing.T) {
	jwt := NewJWTService("test-secret")
	user := &User{ID: 7, Username: "guest", Email: "guest@example.com", Role: RoleGuest}

	refreshToken, err := jwt.Generate(user, TokenTypeRefresh, time.Hour)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if _, err := jwt.Validate(refreshToken, TokenTypeAccess); err == nil {
		t.Fatal("expected refresh token to be rejected as access token")
	}
}

func TestServiceRejectsEmailAsLoginIdentifier(t *testing.T) {
	repo := newFakeRepository()
	svc := newTestService(repo)

	user, err := svc.CreateUser(context.Background(), CreateUserRequest{
		Name:     "Owner",
		Username: "owner",
		Email:    "owner@example.com",
		Password: "password123",
		Role:     RoleAdmin,
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	if _, err := svc.Login(context.Background(), LoginRequest{
		Username: user.Email,
		Password: "password123",
	}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected email login identifier to be rejected, got %v", err)
	}
}

func TestRefreshRejectsChangedUsernameClaim(t *testing.T) {
	repo := newFakeRepository()
	svc := newTestService(repo)

	user, err := svc.CreateUser(context.Background(), CreateUserRequest{
		Name:     "Owner",
		Username: "owner",
		Email:    "owner@example.com",
		Password: "password123",
		Role:     RoleAdmin,
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	login, err := svc.Login(context.Background(), LoginRequest{
		Username: user.Username,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	changedUsername := "owner2"
	if _, err := svc.UpdateUser(context.Background(), user.ID, UpdateUserRequest{
		Username: &changedUsername,
	}); err != nil {
		t.Fatalf("UpdateUser returned error: %v", err)
	}

	if _, err := svc.Refresh(context.Background(), login.RefreshToken); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected changed username claim to be rejected, got %v", err)
	}
}

func TestServiceChangePassword(t *testing.T) {
	repo := newFakeRepository()
	svc := newTestService(repo)
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, CreateUserRequest{
		Name:     "Operator",
		Username: "operator",
		Email:    "operator@example.com",
		Password: "password123",
		Role:     RoleOperator,
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	err = svc.ChangePassword(ctx, user.ID, ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "new-password123",
	})
	if err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}

	if _, err := svc.Login(ctx, LoginRequest{
		Username: user.Username,
		Password: "password123",
	}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected old password to be rejected, got %v", err)
	}
	if _, err := svc.Login(ctx, LoginRequest{
		Username: user.Username,
		Password: "new-password123",
	}); err != nil {
		t.Fatalf("expected new password to work, got %v", err)
	}
}

func TestServiceChangePasswordRejectsInvalidInput(t *testing.T) {
	repo := newFakeRepository()
	svc := newTestService(repo)
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, CreateUserRequest{
		Name:     "Guest",
		Username: "guest",
		Email:    "guest@example.com",
		Password: "password123",
		Role:     RoleGuest,
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
	originalHash := repo.users[user.ID].PasswordHash

	err = svc.ChangePassword(ctx, user.ID, ChangePasswordRequest{
		CurrentPassword: "wrong-password",
		NewPassword:     "new-password123",
	})
	if !errors.Is(err, ErrInvalidCurrentPassword) {
		t.Fatalf("expected invalid current password, got %v", err)
	}

	err = svc.ChangePassword(ctx, user.ID, ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "short",
	})
	if err == nil || err.Error() != "password must be at least 8 characters" {
		t.Fatalf("expected short password validation error, got %v", err)
	}

	if repo.users[user.ID].PasswordHash != originalHash {
		t.Fatal("expected password hash to remain unchanged after rejected changes")
	}

	repo.users[user.ID].IsActive = false
	if err := svc.ChangePassword(ctx, user.ID, ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "new-password123",
	}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected inactive user to be rejected, got %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(repo.users[user.ID].PasswordHash), []byte("password123")); err != nil {
		t.Fatal("expected original password to remain valid")
	}
}
