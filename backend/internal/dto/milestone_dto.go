package dto

// MilestoneMapCreateReq is the request DTO for creating a milestone map.
type MilestoneMapCreateReq struct {
	MapName         string  `json:"mapName" binding:"required,max=100"`
	MapDesc         string  `json:"mapDesc"`
	AssigneeBizKey  string  `json:"assigneeBizKey" binding:"required"`
	PlanStartDate   *string `json:"planStartDate"`
	ExpectedEndDate *string `json:"expectedEndDate"`
}

// MilestoneMapUpdateReq is the request DTO for updating a milestone map.
// Only non-nil fields will be updated.
type MilestoneMapUpdateReq struct {
	MapName         *string `json:"mapName"`
	MapDesc         *string `json:"mapDesc"`
	AssigneeBizKey  *string `json:"assigneeBizKey"`
	PlanStartDate   *string `json:"planStartDate"`
	ExpectedEndDate *string `json:"expectedEndDate"`
}

// MilestoneMapFilter holds filter parameters for listing milestone maps.
type MilestoneMapFilter struct {
	Name        *string `form:"name" json:"name"`
	AssigneeKey *string `form:"assigneeKey" json:"assigneeKey"`
	Status      *string `form:"status" json:"status"`
}

// MilestoneCreateReq is the request DTO for creating a milestone.
type MilestoneCreateReq struct {
	MilestoneName   string `json:"milestoneName" binding:"required,max=100"`
	MilestoneDesc   string `json:"milestoneDesc"`
	ExpectedEndDate string `json:"expectedEndDate" binding:"required"`
}

// MilestoneUpdateReq is the request DTO for updating a milestone.
// Only non-nil fields will be updated.
type MilestoneUpdateReq struct {
	MilestoneName   *string `json:"milestoneName"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	MilestoneDesc   *string `json:"milestoneDesc"`
}

// MilestoneTeamFilter holds filter parameters for listing milestones by team.
type MilestoneTeamFilter struct {
	Name             *string `form:"name"`
	Status           *string `form:"status"`
	ExcludeCancelled *bool   `form:"excludeCancelled"`
}
