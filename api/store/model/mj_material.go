package model

import "time"

// MjMaterialCategory MidJourney 素材分类
type MjMaterialCategory struct {
	Id        uint      `gorm:"primarykey;column:id"`
	Name      string    // 分类内部标识（英文名）
	Title     string    // 分类展示名称
	IsActive  bool      // 是否激活
	SortOrder int       // 排序值，越大越靠前
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

func (MjMaterialCategory) TableName() string {
	return "chatgpt_material_categorys"
}

// MjMaterial MidJourney 素材
type MjMaterial struct {
	Id         uint   `gorm:"primarykey;column:id"`
	CategoryId uint   // 分类 ID
	Title      string // 素材标题
	Name       string // 素材内部标识
	Image      string // 素材图片地址
	//Preview    string    // 预览图地址
	Prompt    string    // 提示词
	Type      string    // 类型：txt2img、img2img、all
	IsActive  bool      // 是否激活
	SortOrder int       // 排序值
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

func (MjMaterial) TableName() string {
	return "chatgpt_materials"
}
