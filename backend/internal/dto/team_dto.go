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
