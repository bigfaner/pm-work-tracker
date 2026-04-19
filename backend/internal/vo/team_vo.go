package vo

import "pm-work-tracker/backend/internal/dto"

// TeamMemberVO is the frontend-facing view object for a team member.
type TeamMemberVO struct {
	ID          uint   `json:"id"`
	TeamID      uint   `json:"teamId"`
	UserID      uint   `json:"userId"`
	Role        string `json:"role"`
	JoinedAt    string `json:"joinedAt"`
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
}

// NewTeamMemberVO converts a dto.TeamMemberDTO to a TeamMemberVO.
func NewTeamMemberVO(d dto.TeamMemberDTO) TeamMemberVO {
	return TeamMemberVO{
		ID:          d.ID,
		TeamID:      d.TeamID,
		UserID:      d.UserID,
		Role:        d.Role,
		JoinedAt:    d.JoinedAt,
		DisplayName: d.DisplayName,
		Username:    d.Username,
	}
}

// NewTeamMemberVOs converts a slice of TeamMemberDTO to TeamMemberVO.
func NewTeamMemberVOs(dtos []dto.TeamMemberDTO) []TeamMemberVO {
	result := make([]TeamMemberVO, 0, len(dtos))
	for _, d := range dtos {
		result = append(result, NewTeamMemberVO(d))
	}
	return result
}
