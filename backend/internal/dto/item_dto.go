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

// ItemPoolFilter holds filter parameters for listing ItemPool entries.
type ItemPoolFilter struct {
	Status string `form:"status"`
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
