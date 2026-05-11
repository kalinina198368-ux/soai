package model

import "time"

// VideoSquare 视频广场数据模型
type VideoSquare struct {
	Id          uint      `gorm:"primarykey;column:id" json:"id"`
	PostId      string    `gorm:"column:post_id;uniqueIndex;size:100" json:"post_id"`      // sora2官网的post ID
	OriginalURL string    `gorm:"column:original_url;type:text" json:"original_url"`      // 原始视频URL
	OssURL      string    `gorm:"column:oss_url;type:text" json:"oss_url"`               // OSS上的视频URL
	PosterURL   string    `gorm:"column:poster_url;type:text" json:"poster_url"`          // 封面图URL（OSS）
	ThumbnailURL string   `gorm:"column:thumbnail_url;type:text" json:"thumbnail_url"`     // 缩略图URL（OSS）
	Prompt      string    `gorm:"column:prompt;type:text" json:"prompt"`                   // 提示词
	Author      string    `gorm:"column:author;size:100" json:"author"`                    // 作者用户名
	AvatarURL   string    `gorm:"column:avatar_url;type:text" json:"avatar_url"`          // 作者头像URL（OSS）
	ViewCount   int       `gorm:"column:view_count;default:0" json:"view_count"`           // 观看次数
	LikeCount   int       `gorm:"column:like_count;default:0" json:"like_count"`           // 点赞次数（从sora2同步）
	Width       int       `gorm:"column:width" json:"width"`                                // 视频宽度
	Height      int       `gorm:"column:height" json:"height"`                             // 视频高度
	NFrames     int       `gorm:"column:n_frames" json:"n_frames"`                         // 帧数
	RawData     string    `gorm:"column:raw_data;type:longtext" json:"raw_data"`           // 原始JSON数据
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (VideoSquare) TableName() string {
	return "chatgpt_video_square"
}

