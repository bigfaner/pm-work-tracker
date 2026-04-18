package dto

// LoginReq is the request DTO for login.
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserDTO is the user object returned in login responses.
type UserDTO struct {
	ID            uint   `json:"id"`
	Username      string `json:"username"`
	DisplayName   string `json:"displayName"`
	Email         string `json:"email"`
	Status        string `json:"status"`
	IsSuperAdmin  bool   `json:"isSuperAdmin"`
	CanCreateTeam bool   `json:"canCreateTeam"`
}

// LoginResp is the response DTO for a successful login.
type LoginResp struct {
	Token string   `json:"token"`
	User  UserDTO  `json:"user"`
}
