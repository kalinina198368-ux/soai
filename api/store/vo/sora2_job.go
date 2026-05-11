package vo

import "time"

type Sora2Job struct {
	Id           uint      `gorm:"primarykey;column:id" json:"id"`
	UserId       int       `json:"user_id"`
	Channel      string    `json:"channel"`       // 频道
	TaskId       string    `json:"task_id"`       // 外部任务ID
	TaskInfo     string    `json:"task_info"`     // 原始任务信息
	Prompt       string    `json:"prompt"`        // 提示词
	PromptExt    string    `json:"prompt_ext"`    // 优化后提示词
	Images       string    `json:"images"`        // 图生视频的图片信息(JSON格式) - LONGTEXT
	CoverURL     string    `json:"cover_url"`     // 封面图 URL
	VideoURL     string    `json:"video_url"`     // 无水印视频 URL
	WaterURL     string    `json:"water_url"`     // 有水印视频 URL
	ThumbnailURL string    `json:"thumbnail_url"` // 缩略图 URL
	Status       string    `json:"status"`        // 任务状态: pending, processing, completed, failed
	Publish      bool      `json:"publish"`       // 是否发布
	IsFavorite   bool      `json:"is_favorite"`   // 是否收藏
	ErrMsg       string    `json:"err_msg"`       // 错误信息
	RawData      string    `json:"raw_data"`      // 原始数据 json
	Power        int       `json:"power"`         // 消耗算力
	Views        int       `json:"views"`         // 观看次数
	Likes        int       `json:"likes"`         // 点赞次数
	Shares       int       `json:"shares"`        // 分享次数
	CreatedAt    int64     `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
