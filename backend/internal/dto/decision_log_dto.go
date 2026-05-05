package dto

// DecisionLogCreateReq is the request DTO for creating a decision log.
type DecisionLogCreateReq struct {
	Category  string   `json:"category" binding:"required,oneof=technical resource requirement schedule risk other"`
	Tags      []string `json:"tags" binding:"dive,max=20"`
	Content   string   `json:"content" binding:"required,max=2000"`
	LogStatus string   `json:"logStatus" binding:"required,oneof=draft published"`
}

// DecisionLogUpdateReq is the request DTO for updating a draft decision log.
type DecisionLogUpdateReq struct {
	Category string   `json:"category" binding:"required,oneof=technical resource requirement schedule risk other"`
	Tags     []string `json:"tags" binding:"dive,max=20"`
	Content  string   `json:"content" binding:"required,max=2000"`
}
