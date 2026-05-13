package dto

// MilestoneMapCreateReq is the request DTO for creating a milestone map.
type MilestoneMapCreateReq struct {
	MapName string `json:"mapName" binding:"required,max=100"`
	MapDesc string `json:"mapDesc"`
}

// MilestoneMapUpdateReq is the request DTO for updating a milestone map.
// Only non-nil fields will be updated.
type MilestoneMapUpdateReq struct {
	MapName *string `json:"mapName"`
	MapDesc *string `json:"mapDesc"`
}

// MilestoneMapFilter holds filter parameters for listing MilestoneMaps.
type MilestoneMapFilter struct {
	Status *string `form:"status"`
}

// MilestoneCreateReq is the request DTO for creating a milestone.
type MilestoneCreateReq struct {
	MilestoneName   string  `json:"milestoneName" binding:"required,max=100"`
	ExpectedEndDate *string `json:"expectedEndDate" binding:"required"`
}

// MilestoneUpdateReq is the request DTO for updating a milestone.
// Only non-nil fields will be updated.
type MilestoneUpdateReq struct {
	MilestoneName   *string `json:"milestoneName"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	DbUpdateTime    string  `json:"dbUpdateTime" binding:"required"`
}
