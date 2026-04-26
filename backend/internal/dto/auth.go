package dto

import "pm-work-tracker/backend/internal/vo"

// LoginReq is the request DTO for login.
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResp is the response DTO for a successful login.
type LoginResp struct {
	Token string     `json:"token"`
	User  vo.UserVO  `json:"user"`
}

// TeamSummary is a lightweight team representation for user detail views.
type TeamSummary struct {
	BizKey string `json:"bizKey"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	TeamID uint   `json:"-"` // internal ID, not exposed to frontend
}

// CreateUserReq is the request DTO for creating a user.
type CreateUserReq struct {
	Username    string `json:"username" binding:"required,min=3,max=64"`
	DisplayName string `json:"displayName" binding:"required,min=1,max=64"`
	Email       string `json:"email" binding:"omitempty,max=100"`
	TeamKey     *string `json:"teamKey"`
}

// UpdateUserReq is the request DTO for updating a user.
type UpdateUserReq struct {
	DisplayName *string `json:"displayName" binding:"omitempty,min=1,max=64"`
	Email       *string `json:"email" binding:"omitempty,max=100"`
	TeamKey     *string `json:"teamKey"`
}

// UpdateUserStatusReq is the request DTO for toggling user status.
type UpdateUserStatusReq struct {
	Status string `json:"status" binding:"required,oneof=enabled disabled"`
}

// AdminUserDTO is the response DTO for admin user endpoints.
type AdminUserDTO struct {
	BizKey          string       `json:"bizKey"`
	Username        string       `json:"username"`
	DisplayName     string       `json:"displayName"`
	Email           string       `json:"email"`
	Status          string       `json:"userStatus"`
	IsSuperAdmin    bool         `json:"isSuperAdmin"`
	Teams           []TeamSummary `json:"teams"`
	InitialPassword string       `json:"initialPassword,omitempty"`
}
