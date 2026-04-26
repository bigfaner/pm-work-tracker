package dto

// MainItemFilter holds filter parameters for listing MainItems.
type MainItemFilter struct {
	Status    string `form:"status"`
	Priority  string `form:"priority"`
	AssigneeKey *string `form:"assigneeKey" json:"assigneeKey"`
	IsKeyItem   *bool   `form:"isKeyItem"`
	Archived    bool    `form:"archived"`
}

// SubItemFilter holds filter parameters for listing SubItems.
type SubItemFilter struct {
	Status    string `form:"status"`
	Priority  string `form:"priority"`
	AssigneeKey *string `form:"assigneeKey" json:"assigneeKey"`
	IsKeyItem   *bool   `form:"isKeyItem"`
}

// SubItemCreateReq is the request DTO for creating a sub item.
type SubItemCreateReq struct {
	MainItemKey     string  `json:"mainItemKey" binding:"required"`
	Title           string  `json:"title" binding:"required,max=100"`
	Description     string  `json:"description"`
	Priority        string  `json:"priority" binding:"required,oneof=P1 P2 P3"`
	AssigneeKey     string  `json:"assigneeKey" binding:"required"`
	StartDate       *string `json:"startDate" binding:"required"`
	ExpectedEndDate *string `json:"expectedEndDate" binding:"required"`
}

// SubItemUpdateReq is the request DTO for updating a sub item.
// Only non-nil fields will be updated.
type SubItemUpdateReq struct {
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	Priority        *string `json:"priority"`
	AssigneeKey     *string `json:"assigneeKey"`
	StartDate       *string `json:"startDate"`
	ExpectedEndDate *string `json:"expectedEndDate"`
}

// ItemPoolFilter holds filter parameters for listing ItemPool entries.
type ItemPoolFilter struct {
	Status string `form:"status"`
}

// SubmitItemPoolReq is the request DTO for submitting an item to the pool.
type SubmitItemPoolReq struct {
	Title          string `json:"title" binding:"required,max=100"`
	Background     string `json:"background"`
	ExpectedOutput string `json:"expectedOutput"`
}

// AssignItemPoolReq is the request DTO for assigning a pool item to an existing main item as sub-item.
type AssignItemPoolReq struct {
	MainItemKey     string  `json:"mainItemKey" binding:"required"`
	AssigneeKey     *string `json:"assigneeKey" binding:"required"`
	Priority        string  `json:"priority"`
	StartDate       *string `json:"startDate" binding:"required"`
	ExpectedEndDate *string `json:"expectedEndDate" binding:"required"`
}

// ConvertToMainItemReq is the request DTO for converting a pool item to a new main item.
type ConvertToMainItemReq struct {
	Priority        string  `json:"priority" binding:"required,oneof=P0 P1 P2 P3"`
	AssigneeKey     *string `json:"assigneeKey" binding:"required"`
	StartDate       *string `json:"startDate" binding:"required"`
	ExpectedEndDate *string `json:"expectedEndDate" binding:"required"`
}

// RejectItemPoolReq is the request DTO for rejecting a pool item.
type RejectItemPoolReq struct {
	Reason string `json:"reason" binding:"required,max=200"`
}

// MainItemCreateReq is the request DTO for creating a main item.
type MainItemCreateReq struct {
	Title           string  `json:"title" binding:"required,max=100"`
	Description     string  `json:"description"`
	Priority        string  `json:"priority" binding:"required,oneof=P0 P1 P2 P3"`
	AssigneeKey     string  `json:"assigneeKey" binding:"required"`
	StartDate       *string `json:"startDate" binding:"required"`
	ExpectedEndDate *string `json:"expectedEndDate" binding:"required"`
	IsKeyItem       bool    `json:"isKeyItem"`
}

// MainItemUpdateReq is the request DTO for updating a main item.
// Only non-nil fields will be updated.
type MainItemUpdateReq struct {
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	Priority        *string `json:"priority"`
	AssigneeKey     *string `json:"assigneeKey"`
	StartDate       *string `json:"startDate"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	ActualEndDate   *string `json:"actualEndDate"`
	IsKeyItem       *bool   `json:"isKeyItem"`
}

// ChangeStatusReq is the request DTO for changing item status.
type ChangeStatusReq struct {
	Status string `json:"status" binding:"required"`
}

// AssignSubItemReq is the request DTO for assigning a sub-item to a user.
type AssignSubItemReq struct {
	AssigneeKey string `json:"assigneeKey" binding:"required"`
}

// CreateRoleReq is the request DTO for creating a role.
type CreateRoleReq struct {
	Name            string   `json:"name" binding:"required"`
	Description     string   `json:"description"`
	PermissionCodes []string `json:"permissionCodes" binding:"required,min=1"`
}

// UpdateRoleReq is the request DTO for updating a role.
type UpdateRoleReq struct {
	Name            *string  `json:"name,omitempty"`
	Description     *string  `json:"description,omitempty"`
	PermissionCodes []string `json:"permissionCodes,omitempty"`
}

// Pagination holds page parameters.
type Pagination struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"pageSize" json:"pageSize"`
}

// PageResult is a generic paginated result.
type PageResult[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

// Weekly view DTOs.

// WeeklyViewResponse is the response DTO for the weekly comparison view.
type WeeklyViewResponse struct {
	WeekStart string                 `json:"weekStart"`
	WeekEnd   string                 `json:"weekEnd"`
	Stats     WeeklyStats            `json:"stats"`
	Groups    []WeeklyComparisonGroup `json:"groups"`
}

// WeeklyStats holds summary statistics for the weekly view.
type WeeklyStats struct {
	ActiveSubItems int `json:"activeSubItems"`
	NewlyCompleted int `json:"newlyCompleted"`
	InProgress     int `json:"inProgress"`
	Blocked        int `json:"blocked"`
	Pending        int `json:"pending"`
	Pausing        int `json:"pausing"`
	Overdue        int `json:"overdue"`
}

// WeeklyComparisonGroup groups sub-items under a main item with week comparison.
type WeeklyComparisonGroup struct {
	MainItem          WeeklyMainItemSummary `json:"mainItem"`
	LastWeek          []SubItemSnapshot     `json:"lastWeek"`
	ThisWeek          []SubItemSnapshot     `json:"thisWeek"`
	CompletedNoChange []SubItemSnapshot     `json:"completedNoChange"`
}

// WeeklyMainItemSummary is a summary of a main item for the weekly comparison view.
type WeeklyMainItemSummary struct {
	BizKey          string  `json:"bizKey"`
	Code            string  `json:"code"`
	Title           string  `json:"title"`
	Priority        string  `json:"priority"`
	Status          string  `json:"status"`
	StartDate       string  `json:"startDate"`
	ExpectedEndDate string  `json:"expectedEndDate"`
	ActualEndDate   *string `json:"actualEndDate"`
	Completion      float64 `json:"completion"`
	SubItemCount    int     `json:"subItemCount"`
}

// SubItemSnapshot represents a sub-item snapshot in the weekly comparison view.
type SubItemSnapshot struct {
	BizKey              string              `json:"bizKey"`
	Code                string              `json:"code"`
	Title               string              `json:"title"`
	Priority            string              `json:"priority"`
	Status              string              `json:"status"`
	AssigneeName        string              `json:"assigneeName"`
	StartDate           string              `json:"startDate"`
	ExpectedEndDate     string              `json:"expectedEndDate"`
	ActualEndDate       *string             `json:"actualEndDate"`
	Completion          float64             `json:"completion"`
	ProgressDescription string              `json:"progressDescription"`
	ProgressRecords     []ProgressRecordDTO `json:"progressRecords"`
	Delta               float64             `json:"delta"`
	IsNew               bool                `json:"isNew"`
	JustCompleted       bool                `json:"justCompleted"`
}

// MainItemSummaryDTO is a lightweight summary of a main item.
type MainItemSummaryDTO struct {
	BizKey     string  `json:"bizKey"`
	Title      string  `json:"title"`
	Completion float64 `json:"completion"`
	IsKeyItem  bool    `json:"isKeyItem,omitempty"`
}

// ProgressRecordDTO is a lightweight progress record for weekly view.
type ProgressRecordDTO struct {
	ID          uint    `json:"id"`
	Completion  float64 `json:"completion"`
	Achievement string  `json:"achievement"`
	Blocker     string  `json:"blocker"`
	CreatedAt   string  `json:"createdAt"`
}

// Table view DTOs.

// TableFilter holds filter and sort parameters for the table view.
type TableFilter struct {
	Type       string   `form:"type" json:"type"`                     // "main"|"sub"|"" (empty = both)
	Priority   []string `form:"priority" json:"priority"`             // ["P1","P2","P3"]
	Status     []string `form:"status" json:"status"`
	AssigneeKey *string `form:"assigneeKey" json:"assigneeKey"`
	SortBy     string   `form:"sortBy" json:"sortBy"`                 // field name
	SortOrder  string   `form:"sortOrder" json:"sortOrder"`           // "asc"|"desc"
}

// TableRow represents a single row in the table view (main item or sub-item).
type TableRow struct {
	BizKey         string  `json:"bizKey"`
	Type            string  `json:"type"`        // "main"|"sub"
	Code            string  `json:"code"`
	Title           string  `json:"title"`
	Priority        string  `json:"priority"`
	AssigneeID      *string `json:"assigneeKey"`
	AssigneeName    string  `json:"assigneeName"`
	Status          string  `json:"status"`
	Completion      float64 `json:"completion"`
	ExpectedEndDate *string `json:"expectedEndDate"`
	ActualEndDate   *string `json:"actualEndDate"`
}

// Gantt view DTOs.

// GanttFilter holds filter parameters for the gantt view.
type GanttFilter struct {
	Status string `form:"status"`
}

// GanttResult is the response DTO for the gantt view.
type GanttResult struct {
	Items []GanttMainItemDTO `json:"items"`
}

// GanttMainItemDTO represents a main item in the gantt view.
type GanttMainItemDTO struct {
	BizKey         string            `json:"bizKey"`
	Title          string            `json:"title"`
	Priority       string            `json:"priority"`
	StartDate      string            `json:"startDate"`
	ExpectedEndDate string           `json:"expectedEndDate"`
	Completion     float64           `json:"completion"`
	Status         string            `json:"status"`
	IsOverdue      bool              `json:"isOverdue"`
	SubItems       []GanttSubItemDTO `json:"subItems"`
}

// GanttSubItemDTO represents a sub-item summary in the gantt view.
type GanttSubItemDTO struct {
	BizKey          string  `json:"bizKey"`
	Title           string  `json:"title"`
	StartDate       string  `json:"startDate"`
	ExpectedEndDate string  `json:"expectedEndDate"`
	Completion      float64 `json:"completion"`
	Status          string  `json:"status"`
}

// Report view DTOs.

// ReportPreview is the response DTO for the weekly report preview.
type ReportPreview struct {
	WeekStart string             `json:"weekStart"`
	WeekEnd   string             `json:"weekEnd"`
	Sections  []ReportSectionDTO `json:"sections"`
}

// ReportSectionDTO groups sub-items under a main item for the report.
type ReportSectionDTO struct {
	MainItem MainItemSummaryDTO `json:"mainItem"`
	SubItems []ReportSubItemDTO `json:"subItems"`
}

// ReportSubItemDTO represents a sub-item in the weekly report.
type ReportSubItemDTO struct {
	BizKey       string   `json:"bizKey"`
	Title        string   `json:"title"`
	Completion   float64  `json:"completion"`
	AssigneeID   *string  `json:"assigneeKey"`
	AssigneeName string   `json:"assigneeName"`
	Achievements []string `json:"achievements"`
	Blockers     []string `json:"blockers"`
}
