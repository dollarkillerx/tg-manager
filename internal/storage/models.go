package storage

import "time"

type ForwardRule struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	SourceChannelID int64     `gorm:"index;not null" json:"source_channel_id"`
	SourceName      string    `json:"source_name"`
	SourceHash      int64     `json:"source_hash"`
	TargetChannelID int64     `gorm:"not null" json:"target_channel_id"`
	TargetName      string    `json:"target_name"`
	TargetHash      int64     `json:"target_hash"`
	MatchPattern    string    `gorm:"not null" json:"match_pattern"`
	Enabled         bool      `gorm:"default:true" json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ForwardLog struct {
	ID              uint      `gorm:"primaryKey"`
	RuleID          uint      `gorm:"uniqueIndex:idx_rule_msg;not null"`
	MessageID       int       `gorm:"uniqueIndex:idx_rule_msg;not null"`
	SourceChannelID int64     `gorm:"not null"`
	TargetChannelID int64     `gorm:"not null"`
	CreatedAt       time.Time
}
