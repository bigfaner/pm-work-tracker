package model

type User struct {
	BaseModel
	Username     string `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	DisplayName  string `gorm:"type:varchar(64);not null" json:"displayName"`
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`
	Email        string `gorm:"type:varchar(100)" json:"email"`
	UserStatus   string `gorm:"type:varchar(10);not null;default:'enabled'" json:"userStatus"`
	IsSuperAdmin bool   `gorm:"not null;default:false" json:"isSuperAdmin"`
}

func (User) TableName() string {
	return "pmw_users"
}
