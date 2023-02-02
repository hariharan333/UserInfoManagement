package models

type User struct {
	Id        int64  `gorm:"primaryKey" json:"id"`
	FullName  string `gorm:"type:varchar(45)" json:"fullName"`
	Email     string `gorm:"type:varchar(45)" json:"email"`
	PhoneNo   string `gorm:"type:varchar(45)" json:"phoneNo"`
	Image     string `gorm:"type:blob" json:"image"`
	CreateaAt string `gorm:"type:varchar(45)" json:"createAt"`
	UpdatedAt string `gorm:"type:varchar(45)" json:"updatedAt"`
}
