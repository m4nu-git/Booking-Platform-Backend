package dto

type LoginUserRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type CreateUserRequestDTO struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RoleDTO struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type ProfileResponseDTO struct {
	Id        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Roles     []RoleDTO `json:"roles"`
}