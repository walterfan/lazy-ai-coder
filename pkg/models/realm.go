package models

import (
	"time"

	"gorm.io/gorm"
)

// Realm represents the realm table in the database
type Realm struct {
	ID          string         `json:"id" gorm:"primaryKey;type:text"`
	Name        string         `json:"name" gorm:"unique;not null;type:text"`
	Description string         `json:"description" gorm:"type:text"`
	CreatedBy   string         `json:"created_by" gorm:"type:text"`
	CreatedTime time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:text"`
	UpdatedTime time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
