package models

import (
	"time"

	"gorm.io/gorm"
)

// SkillRating tracks a user's rating and notes for an AI skill
type SkillRating struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	SkillPath   string         `json:"skill_path" gorm:"not null;type:text;uniqueIndex"`
	SkillName   string         `json:"skill_name" gorm:"not null;type:text"`
	Category    string         `json:"category" gorm:"type:text;index"`
	Score       int            `json:"score" gorm:"not null;default:0"`  // 1-5 stars
	Tags        string         `json:"tags" gorm:"type:text"`            // comma-separated
	Notes       string         `json:"notes" gorm:"type:text"`
	UsageCount  int            `json:"usage_count" gorm:"not null;default:0"`
	Favorited   bool           `json:"favorited" gorm:"not null;default:false;index"`
	CreatedTime time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (SkillRating) TableName() string {
	return "skill_ratings"
}
