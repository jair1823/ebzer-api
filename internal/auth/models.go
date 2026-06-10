package auth

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleGuest    Role = "guest"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleOperator, RoleGuest:
		return true
	default:
		return false
	}
}

type User struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         Role   `json:"role"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type AuthUser struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     Role   `json:"role"`
	IsActive bool   `json:"is_active"`
}

func NewAuthUser(user *User) AuthUser {
	return AuthUser{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: user.IsActive,
	}
}
