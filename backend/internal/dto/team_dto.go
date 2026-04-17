package dto

// TeamMemberDTO is the response DTO for team membership with joined user info.
type TeamMemberDTO struct {
	ID          uint   `json:"id"`
	TeamID      uint   `json:"team_id"`
	UserID      uint   `json:"user_id"`
	Role        string `json:"role"`
	JoinedAt    string `json:"joined_at"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
}

// CreateTeamReq is the request DTO for creating a team.
type CreateTeamReq struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// InviteMemberReq is the request DTO for inviting a member to a team.
type InviteMemberReq struct {
	Username string `json:"username" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=member"`
}
