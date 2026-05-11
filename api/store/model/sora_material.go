package model

import "time"

// SoraMaterialCategory Sora 视频素材分类
type SoraMaterialCategory struct {
	Id        uint      `gorm:"primarykey;column:id"`
	Name      string    // 分类内部标识（英文名）
	Title     string    // 分类展示名称
	IsActive  bool      // 是否激活
	SortOrder int       // 排序值，越大越靠前
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

func (SoraMaterialCategory) TableName() string {
	return "chatgpt_sora_material_categorys"
}

// SoraMaterial Sora 视频素材
type SoraMaterial struct {
	Id         uint      `gorm:"primarykey;column:id"`
	CategoryId uint      // 分类 ID
	Title      string    // 素材标题
	Name       string    // 素材内部标识
	Video      string    // 素材视频地址
	Image      string    // 素材封面图地址
	Prompt     string    // 提示词
	IsActive   bool      // 是否激活
	SortOrder  int       // 排序值
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间
}

func (SoraMaterial) TableName() string {
	return "chatgpt_sora_materials"
}

