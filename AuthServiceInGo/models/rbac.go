package models

type BaseModel struct {
	Id        int64  `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Role struct {
	BaseModel
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Permission struct {
	BaseModel
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

type RolePermission struct {
	BaseModel
	RoleId       int64 `json:"role_id"`
	PermissionId int64 `json:"permission_id"`
}

type UserRole struct {
	BaseModel
	UserId int64 `json:"user_id"`
	RoleId int64 `json:"role_id"`
}

// DTO for responses that need user + role info together
type UserRoleResponse struct {
	UserName    string `json:"username"`
	Email       string `json:"email"`
	RoleName    string `json:"role_name"`
	Description string `json:"description"`
}
