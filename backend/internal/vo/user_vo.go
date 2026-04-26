package vo

import (
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg"
)

// UserVO is the frontend-facing view object for a user.
type UserVO struct {
	BizKey       string `json:"bizKey"`
	Username     string `json:"username"`
	DisplayName  string `json:"displayName"`
	Email        string `json:"email"`
	Status       string `json:"status"`
	IsSuperAdmin bool   `json:"isSuperAdmin"`
}

// NewUserVO converts a model.User to a UserVO.
func NewUserVO(u *model.User) UserVO {
	return UserVO{
		BizKey:       pkg.FormatID(u.BizKey),
		Username:     u.Username,
		DisplayName:  u.DisplayName,
		Email:        u.Email,
		Status:       u.UserStatus,
		IsSuperAdmin: u.IsSuperAdmin,
	}
}
