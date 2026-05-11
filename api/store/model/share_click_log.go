package model

import "time"

type ShareClickLog struct {
	Id         uint      `gorm:"primarykey;column:id"`
	InviteCode string    `gorm:"column:invite_code;index"`
	UserId     uint      `gorm:"column:user_id;index"`
	Ip         string    `gorm:"column:ip"`
	UserAgent  string    `gorm:"column:user_agent"`
	CreatedAt  time.Time `gorm:"column:created_at;index"`
}

// func (ShareClickLog) TableName() string {
// 	return "share_click_logs"
// }

