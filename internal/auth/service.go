package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrForbidden          = errors.New("forbidden")
)

type Service interface {
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)
	Refresh(ctx context.Context, refreshToken string) (RefreshResponse, error)
	CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, id int) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	ListUsers(ctx context.Context) ([]User, error)
	UpdateUser(ctx context.Context, id int, req UpdateUserRequest) (*User, error)
	DeactivateUser(ctx context.Context, id int) error
	BootstrapInitialAdmin(ctx context.Context) error
}

type service struct {
	repo       Repository
	jwt        *JWTService
	accessTTL  time.Duration
	refreshTTL time.Duration
	isProd     bool
}

func NewService(repo Repository, jwt *JWTService, cfg Config) Service {
	return &service{
		repo:       repo,
		jwt:        jwt,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		isProd:     cfg.IsProd,
	}
}

func (s *service) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	username := normalizeUsername(req.Username)
	if username == "" {
		return LoginResponse{}, ErrInvalidCredentials
	}
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return LoginResponse{}, err
	}
	if user == nil || !user.IsActive {
		return LoginResponse{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.issueTokenPair(user)
	if err != nil {
		return LoginResponse{}, err
	}
	return LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTTL.Seconds()),
		User:         NewAuthUser(user),
	}, nil
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (RefreshResponse, error) {
	claims, err := s.jwt.Validate(refreshToken, TokenTypeRefresh)
	if err != nil {
		return RefreshResponse{}, ErrInvalidCredentials
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return RefreshResponse{}, ErrInvalidCredentials
	}
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return RefreshResponse{}, err
	}
	if user == nil || !user.IsActive || user.Username != claims.Username || user.Role != claims.Role {
		return RefreshResponse{}, ErrInvalidCredentials
	}

	accessToken, err := s.jwt.Generate(user, TokenTypeAccess, s.accessTTL)
	if err != nil {
		return RefreshResponse{}, err
	}
	return RefreshResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *service) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Username = normalizeUsername(req.Username)
	email, err := normalizeEmail(req.Email)
	if err != nil {
		return nil, err
	}
	req.Email = email
	if req.Role == "" {
		req.Role = RoleOperator
	}
	if err := validateUserInput(req.Name, req.Username, req.Password, req.Role); err != nil {
		return nil, err
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	id, err := s.repo.Create(ctx, req, hash)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetUser(ctx context.Context, id int) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *service) ListUsers(ctx context.Context) ([]User, error) {
	return s.repo.List(ctx)
}

func (s *service) UpdateUser(ctx context.Context, id int, req UpdateUserRequest) (*User, error) {
	if req.Name != nil {
		v := strings.TrimSpace(*req.Name)
		if v == "" {
			return nil, errors.New("name is required")
		}
		req.Name = &v
	}
	if req.Username != nil {
		v := normalizeUsername(*req.Username)
		if v == "" {
			return nil, errors.New("username is required")
		}
		req.Username = &v
	}
	if req.Email != nil {
		v, err := normalizeEmail(*req.Email)
		if err != nil {
			return nil, err
		}
		req.Email = &v
	}
	if req.Role != nil && !req.Role.IsValid() {
		return nil, errors.New("invalid role")
	}

	var passwordHash *string
	if req.Password != nil {
		if len(*req.Password) < 8 {
			return nil, errors.New("password must be at least 8 characters")
		}
		hash, err := hashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		passwordHash = &hash
	}

	if err := s.repo.Update(ctx, id, req, passwordHash); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *service) DeactivateUser(ctx context.Context, id int) error {
	return s.repo.Deactivate(ctx, id)
}

func (s *service) BootstrapInitialAdmin(ctx context.Context) error {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	name := strings.TrimSpace(os.Getenv("INITIAL_ADMIN_NAME"))
	username := normalizeUsername(os.Getenv("INITIAL_ADMIN_USERNAME"))
	email := strings.TrimSpace(os.Getenv("INITIAL_ADMIN_EMAIL"))
	password := os.Getenv("INITIAL_ADMIN_PASSWORD")
	if name == "" && email == "" && password == "" && !s.isProd {
		log.Println("No users exist and INITIAL_ADMIN_* is not set; skipping local admin bootstrap")
		return nil
	}
	if name == "" || email == "" || password == "" {
		return errors.New("INITIAL_ADMIN_NAME, INITIAL_ADMIN_EMAIL, and INITIAL_ADMIN_PASSWORD are required when no users exist")
	}
	if username == "" {
		username = usernameFromEmail(email)
	}

	user, err := s.CreateUser(ctx, CreateUserRequest{
		Name:     name,
		Username: username,
		Email:    email,
		Password: password,
		Role:     RoleAdmin,
	})
	if err != nil {
		return fmt.Errorf("create initial admin: %w", err)
	}
	log.Printf("Created initial admin user: %s", user.Username)
	return nil
}

func (s *service) issueTokenPair(user *User) (string, string, error) {
	accessToken, err := s.jwt.Generate(user, TokenTypeAccess, s.accessTTL)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := s.jwt.Generate(user, TokenTypeRefresh, s.refreshTTL)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func validateUserInput(name, username, password string, role Role) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(username) == "" {
		return errors.New("username is required")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if !role.IsValid() {
		return errors.New("invalid role")
	}
	return nil
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func usernameFromEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return ""
	}
	return normalizeUsername(email[:at])
}

func normalizeEmail(email string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(email))
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value {
		return "", errors.New("valid email is required")
	}
	return value, nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
