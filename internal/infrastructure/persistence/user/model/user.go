package model

import gormlib "gorm.io/gorm"

type User struct {
	gormlib.Model
	UID      string `gorm:"column:uid;type:varchar(36);uniqueIndex;not null"`
	Name     string `gorm:"column:name;type:varchar(64);not null"`
	Gender   uint8  `gorm:"column:gender;type:tinyint unsigned;not null;default:0"`
	Email    string `gorm:"column:email;type:varchar(255);uniqueIndex;not null"`
	Password string `gorm:"column:password;type:varchar(255);not null"`
}

func (User) TableName() string {
	return "user"
}
