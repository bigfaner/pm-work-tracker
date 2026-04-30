package dto

// TeamMemberDTO is the response DTO for team membership with joined user info.
type TeamMemberDTO struct {
	BizKey      string `json:"bizKey"`
	TeamKey     string `json:"teamKey"`
	UserKey     string `json:"userKey"`
	Role        string `json:"role"`
	RoleKey     string `json:"roleKey"`
	RoleName    string `json:"roleName"`
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
	RoleKey  string `json:"roleKey" binding:"required"`
}

// TransferPMReq is the request DTO for transferring PM role.
type TransferPMReq struct {
	NewPmUserKey string `json:"newPmUserKey" binding:"required"`
}

// DisbandTeamReq is the request DTO for disbanding a team.
type DisbandTeamReq struct {
	ConfirmName string `json:"confirmName" binding:"required"`
}

// UpdateMemberRoleReq is the request DTO for updating a team member's role.
type UpdateMemberRoleReq struct {
	RoleKey string `json:"roleKey" binding:"required"`
}

// UserSearchDTO is a lightweight user DTO for the invite member search.
type UserSearchDTO struct {
	BizKey      string `json:"bizKey"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

// TeamDetailResp is the response DTO for team detail.
type TeamDetailResp struct {
	BizKey        string `json:"bizKey"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Code          string `json:"code"`
	PmKey         string `json:"pmKey"`
	PmDisplayName string `json:"pmDisplayName"`
	MemberCount   int    `json:"memberCount"`
	MainItemCount int    `json:"mainItemCount"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// TeamListResp is the response DTO for team listing with PM display name.
type TeamListResp struct {
	BizKey        string `json:"bizKey"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Code          string `json:"code"`
	PmKey         string `json:"pmKey"`
	PmDisplayName string `json:"pmDisplayName"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// AdminTeamDTO is the response DTO for admin team listing with aggregated counts.
type AdminTeamDTO struct {
	BizKey        string `json:"bizKey"`
	Name          string `json:"name"`
	PMDisplayName string `json:"pmDisplayName"`
	MemberCount   int    `json:"memberCount"`
	MainItemCount int    `json:"mainItemCount"`
	CreatedAt     string `json:"createdAt"`
}

// TeamListPage is a typed pagination wrapper for team listing.
type TeamListPage struct {
	Items    []*TeamListResp `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}
