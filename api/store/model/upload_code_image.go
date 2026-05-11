package model

import "time"

// UploadCodeImage 上传码图片记录（匿名上传的图片归档到拥有者）
type UploadCodeImage struct {
	Id         uint `gorm:"primarykey;column:id"`
	UserId     uint // 拥有者用户ID
	Code       string
	URL        string
	ObjKey     string
	Name       string
	Ext        string
	Size       int64
	UploaderIP string
	CreatedAt  time.Time
}
