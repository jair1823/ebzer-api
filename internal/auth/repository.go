package auth

import (
	"context"
	"database/sql"
	"errors"
)

type Repository interface {
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, req CreateUserRequest, passwordHash string) (int, error)
	GetByID(ctx context.Context, id int) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]User, error)
	Update(ctx context.Context, id int, req UpdateUserRequest, passwordHash *string) error
	Deactivate(ctx context.Context, id int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

const userSelectColumns = `id, name, email, password_hash, role, is_active, created_at, updated_at`

func scanUser(row interface {
	Scan(dest ...any) error
}) (*User, error) {
	var user User
	var isActive int
	if err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&isActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}
	user.IsActive = isActive != 0
	return &user, nil
}

func (r *repository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (r *repository) Create(ctx context.Context, req CreateUserRequest, passwordHash string) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, LOWER($2), $3, $4)
		RETURNING id;
	`, req.Name, req.Email, passwordHash, req.Role).Scan(&id)
	return id, err
}

func (r *repository) GetByID(ctx context.Context, id int) (*User, error) {
	user, err := scanUser(r.db.QueryRowContext(ctx,
		"SELECT "+userSelectColumns+" FROM users WHERE id = $1", id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return user, err
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	user, err := scanUser(r.db.QueryRowContext(ctx,
		"SELECT "+userSelectColumns+" FROM users WHERE LOWER(email) = LOWER($1)", email))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return user, err
}

func (r *repository) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT "+userSelectColumns+" FROM users ORDER BY name COLLATE NOCASE ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}
	return users, rows.Err()
}

func (r *repository) Update(ctx context.Context, id int, req UpdateUserRequest, passwordHash *string) error {
	var role *string
	if req.Role != nil {
		v := string(*req.Role)
		role = &v
	}

	var isActive *int
	if req.IsActive != nil {
		v := 0
		if *req.IsActive {
			v = 1
		}
		isActive = &v
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE users SET
			name = COALESCE($1, name),
			email = COALESCE(LOWER($2), email),
			password_hash = COALESCE($3, password_hash),
			role = COALESCE($4, role),
			is_active = COALESCE($5, is_active),
			updated_at = datetime('now')
		WHERE id = $6;
	`, req.Name, req.Email, passwordHash, role, isActive, id)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *repository) Deactivate(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE users SET is_active = 0, updated_at = datetime('now') WHERE id = $1", id)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("user not found")
	}
	return nil
}
