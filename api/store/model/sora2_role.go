package model

import "time"

// Sora2Role 对应表 chatgpt_sora2_roles
type Sora2Role struct {
	Id                int       `gorm:"primarykey;column:id" json:"id"`
	UserId            int       `json:"user_id"`
	TaskId            string    `json:"task_id"` // 对应的视频 job 的 ID（chatgpt_sora2_jobs.id）
	ErrMsg            string    `json:"err_msg"`
	RawData           string    `json:"raw_data"`
	Power             int       `json:"power"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Status            string    `json:"status"`              // pending, processing, completed, failed
	ApiId             int       `json:"api_id"`              // 预留字段
	Username          string    `json:"username"`            // 对应的 username
	DisplayName       string    `json:"display_name"`        // 对应的 display_name
	Permalink         string    `json:"permalink"`           // 对应的 permalink
	ProfilePictureURL string    `json:"profile_picture_url"` // 头像地址
	SysPictureUrl     string    `json:"sys_picture_url"`     //系统头像地址 也就是将ProfilePictureURL下载到oss中，之后的url
}

func (Sora2Role) TableName() string {
	return "chatgpt_sora2_roles"
}
