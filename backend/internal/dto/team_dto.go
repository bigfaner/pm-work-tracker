package dto

// TeamMemberDTO is the response DTO for team membership with joined user info.
type TeamMemberDTO struct {
	ID          uint   `json:"id"`
	TeamID      uint   `json:"teamId"`
	UserID      uint   `json:"userId"`
	Role        string `json:"role"`
	JoinedAt    string `json:"joinedAt"`
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
}

// CreateTeamReq is the request DTO for creating a team.
type CreateTeamReq struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
	Code        string `json:"code" binding:"required,min=2,max=6,alpha"`
}

// UpdateTeamReq is the request DTO for updating a team.
type UpdateTeamReq struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// InviteMemberReq is the request DTO for inviting a member to a team.
type InviteMemberReq struct {
	Username string `json:"username" binding:"required"`
	RoleID   uint   `json:"roleId" binding:"required"`
}

// TransferPMReq is the request DTO for transferring PM role.
type TransferPMReq struct {
	NewPmUserID uint `json:"newPmUserId" binding:"required"`
}

// DisbandTeamReq is the request DTO for disbanding a team.
type DisbandTeamReq struct {
	ConfirmName string `json:"confirmName" binding:"required"`
}

// UserSearchDTO is a lightweight user DTO for the invite member search.
type UserSearchDTO struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

// TeamDetailResp is the response DTO for team detail.
type TeamDetailResp struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	PmID          uint   `json:"pmId"`
	PmDisplayName string `json:"pmDisplayName"`
	MemberCount   int    `json:"memberCount"`
	MainItemCount int    `json:"mainItemCount"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// TeamListResp is the response DTO for team listing with PM display name.
type TeamListResp struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	PmID          uint   `json:"pmId"`
	PmDisplayName string `json:"pmDisplayName"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// AdminTeamDTO is the response DTO for admin team listing with aggregated counts.
type AdminTeamDTO struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	PMDisplayName string `json:"pmDisplayName"`
	MemberCount   int    `json:"memberCount"`
	MainItemCount int    `json:"mainItemCount"`
	CreatedAt     string `json:"createdAt"`
}
