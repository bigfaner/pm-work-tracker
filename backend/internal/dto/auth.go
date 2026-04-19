package dto

// LoginReq is the request DTO for login.
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserDTO is the user object returned in login responses.
type UserDTO struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	DisplayName  string `json:"displayName"`
	Email        string `json:"email"`
	Status       string `json:"status"`
	IsSuperAdmin bool   `json:"isSuperAdmin"`
}

// LoginResp is the response DTO for a successful login.
type LoginResp struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

// TeamSummary is a lightweight team representation for user detail views.
type TeamSummary struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// CreateUserReq is the request DTO for creating a user.
type CreateUserReq struct {
	Username    string `json:"username" binding:"required,min=3,max=64"`
	DisplayName string `json:"displayName" binding:"required,min=1,max=64"`
	Email       string `json:"email" binding:"omitempty,max=100"`
	TeamID      *uint  `json:"teamId"`
}

// UpdateUserReq is the request DTO for updating a user.
type UpdateUserReq struct {
	DisplayName *string `json:"displayName" binding:"omitempty,min=1,max=64"`
	Email       *string `json:"email" binding:"omitempty,max=100"`
	TeamID      *uint   `json:"teamId"`
}

// UpdateUserStatusReq is the request DTO for toggling user status.
type UpdateUserStatusReq struct {
	Status string `json:"status" binding:"required,oneof=enabled disabled"`
}

// AdminUserDTO is the response DTO for admin user endpoints.
type AdminUserDTO struct {
	ID              uint          `json:"id"`
	Username        string        `json:"username"`
	DisplayName     string        `json:"displayName"`
	Email           string        `json:"email"`
	Status          string        `json:"status"`
	IsSuperAdmin    bool          `json:"isSuperAdmin"`
	Teams           []TeamSummary `json:"teams"`
	InitialPassword string        `json:"initialPassword,omitempty"`
}
