package handler

import (
	"github.com/gin-gonic/gin"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/service"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authSvc service.AuthService
}

// NewAuthHandler creates a new AuthHandler with the given AuthService.
func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.RespondError(c, apperrors.ErrValidation)
		return
	}

	token, user, err := h.authSvc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		apperrors.RespondError(c, err)
		return
	}

	apperrors.RespondOK(c, dto.LoginResp{
		Token: token,
		User:  userToDTO(user),
	})
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	apperrors.RespondOK(c, nil)
}

// userToDTO converts a model.User to dto.UserDTO.
func userToDTO(u *model.User) dto.UserDTO {
	return dto.UserDTO{
		ID:            u.ID,
		Username:      u.Username,
		DisplayName:   u.DisplayName,
		IsSuperAdmin:  u.IsSuperAdmin,
		CanCreateTeam: u.CanCreateTeam,
	}
}
