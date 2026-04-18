package dto

// MainItemFilter holds filter parameters for listing MainItems.
type MainItemFilter struct {
	Status    string `form:"status"`
	Priority  string `form:"priority"`
	AssigneeID *uint `form:"assignee_id"`
	IsKeyItem *bool  `form:"is_key_item"`
	Archived  bool   `form:"archived"`
}

// SubItemFilter holds filter parameters for listing SubItems.
type SubItemFilter struct {
	Status    string `form:"status"`
	Priority  string `form:"priority"`
	AssigneeID *uint `form:"assignee_id"`
	IsKeyItem *bool  `form:"is_key_item"`
}

// SubItemCreateReq is the request DTO for creating a sub item.
type SubItemCreateReq struct {
	MainItemID      uint   `json:"main_item_id" binding:"required"`
	Title           string `json:"title" binding:"required,max=100"`
	Description     string `json:"description"`
	Priority        string `json:"priority" binding:"required,oneof=P1 P2 P3"`
	AssigneeID      *uint  `json:"assignee_id"`
	StartDate       *string `json:"start_date"`
	ExpectedEndDate *string `json:"expected_end_date"`
}

// SubItemUpdateReq is the request DTO for updating a sub item.
// Only non-nil fields will be updated.
type SubItemUpdateReq struct {
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	Priority        *string `json:"priority"`
	AssigneeID      *uint   `json:"assignee_id"`
	StartDate       *string `json:"start_date"`
	ExpectedEndDate *string `json:"expected_end_date"`
}

// ItemPoolFilter holds filter parameters for listing ItemPool entries.
type ItemPoolFilter struct {
	Status string `form:"status"`
}

// SubmitItemPoolReq is the request DTO for submitting an item to the pool.
type SubmitItemPoolReq struct {
	Title          string `json:"title" binding:"required,max=100"`
	Background     string `json:"background"`
	ExpectedOutput string `json:"expected_output"`
}

// AssignItemPoolReq is the request DTO for assigning a pool item to a main item.
type AssignItemPoolReq struct {
	MainItemID uint `json:"mainItemId" binding:"required"`
	AssigneeID uint `json:"assigneeId" binding:"required"`
}

// RejectItemPoolReq is the request DTO for rejecting a pool item.
type RejectItemPoolReq struct {
	Reason string `json:"reason" binding:"required,max=200"`
}

// MainItemCreateReq is the request DTO for creating a main item.
type MainItemCreateReq struct {
	Title           string  `json:"title" binding:"required,max=100"`
	Priority        string  `json:"priority" binding:"required,oneof=P0 P1 P2 P3"`
	AssigneeID      *uint   `json:"assignee_id"`
	StartDate       *string `json:"start_date"`
	ExpectedEndDate *string `json:"expected_end_date"`
	IsKeyItem       bool    `json:"is_key_item"`
}

// MainItemUpdateReq is the request DTO for updating a main item.
// Only non-nil fields will be updated.
type MainItemUpdateReq struct {
	Title           *string `json:"title"`
	Priority        *string `json:"priority"`
	AssigneeID      *uint   `json:"assignee_id"`
	StartDate       *string `json:"start_date"`
	ExpectedEndDate *string `json:"expected_end_date"`
	ActualEndDate   *string `json:"actual_end_date"`
	Status          *string `json:"status"`
	IsKeyItem       *bool   `json:"is_key_item"`
}

// Pagination holds page parameters.
type Pagination struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
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
	ID              uint    `json:"id"`
	Title           string  `json:"title"`
	Priority        string  `json:"priority"`
	StartDate       string  `json:"startDate"`
	ExpectedEndDate string  `json:"expectedEndDate"`
	Completion      float64 `json:"completion"`
	SubItemCount    int     `json:"subItemCount"`
}

// SubItemSnapshot represents a sub-item snapshot in the weekly comparison view.
type SubItemSnapshot struct {
	ID                  uint    `json:"id"`
	Title               string  `json:"title"`
	Priority            string  `json:"priority"`
	Status              string  `json:"status"`
	AssigneeName        string  `json:"assigneeName"`
	ExpectedEndDate     string  `json:"expectedEndDate"`
	Completion          float64 `json:"completion"`
	ProgressDescription string  `json:"progressDescription"`
	Delta               float64 `json:"delta"`
	IsNew               bool    `json:"isNew"`
	JustCompleted       bool    `json:"justCompleted"`
}

// WeeklyViewResult is the response DTO for the weekly view.
// Deprecated: Use WeeklyViewResponse for the new comparison format.
type WeeklyViewResult struct {
	WeekStart string           `json:"weekStart"`
	WeekEnd   string           `json:"weekEnd"`
	Groups    []WeeklyGroupDTO `json:"groups"`
}

// WeeklyGroupDTO groups sub-items under a main item for the weekly view.
type WeeklyGroupDTO struct {
	MainItem             MainItemSummaryDTO   `json:"mainItem"`
	NewlyCompleted       []SubItemWeekDTO     `json:"newlyCompleted"`
	HasProgress          []SubItemWeekDTO     `json:"hasProgress"`
	NoChangeFromLastWeek []SubItemSummaryDTO  `json:"noChangeFromLastWeek"`
}

// MainItemSummaryDTO is a lightweight summary of a main item.
type MainItemSummaryDTO struct {
	ID         uint    `json:"id"`
	Title      string  `json:"title"`
	Completion float64 `json:"completion"`
	IsKeyItem  bool    `json:"isKeyItem,omitempty"`
}

// SubItemSummaryDTO is a lightweight summary of a sub-item.
type SubItemSummaryDTO struct {
	ID         uint    `json:"id"`
	Title      string  `json:"title"`
	Status     string  `json:"status"`
	Completion float64 `json:"completion"`
}

// SubItemWeekDTO is a sub-item with its progress records for the week.
type SubItemWeekDTO struct {
	ID               uint                  `json:"id"`
	Title            string                `json:"title"`
	Status           string                `json:"status"`
	Completion       float64               `json:"completion"`
	MainItemID       uint                  `json:"mainItemId"`
	ProgressThisWeek []ProgressRecordDTO   `json:"progressThisWeek"`
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
	AssigneeID *uint    `form:"assigneeId" json:"assigneeId"`
	SortBy     string   `form:"sortBy" json:"sortBy"`                 // field name
	SortOrder  string   `form:"sortOrder" json:"sortOrder"`           // "asc"|"desc"
}

// TableRow represents a single row in the table view (main item or sub-item).
type TableRow struct {
	ID              uint    `json:"id"`
	Type            string  `json:"type"`        // "main"|"sub"
	Code            string  `json:"code"`
	Title           string  `json:"title"`
	Priority        string  `json:"priority"`
	AssigneeID      *uint   `json:"assigneeId"`
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
	ID             uint              `json:"id"`
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
	ID              uint    `json:"id"`
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
	ID           uint     `json:"id"`
	Title        string   `json:"title"`
	Completion   float64  `json:"completion"`
	Achievements []string `json:"achievements"`
	Blockers     []string `json:"blockers"`
}
