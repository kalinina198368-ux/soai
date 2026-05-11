package model

import "time"

// UploadCode 用户上传码（用于匿名扫码上传）
type UploadCode struct {
	Id        uint `gorm:"primarykey;column:id"`
	UserId    uint
	Code      string
	CreatedAt time.Time
}
