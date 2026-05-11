package vo

// MjMaterialCategory MidJourney 素材分类 VO
type MjMaterialCategory struct {
	Id       uint   `json:"id"`
	Name     string `json:"name"`      // 分类内部标识
	Title    string `json:"title"`     // 分类展示名称
	SortNum  int    `json:"sort"`      // 排序值
	IsActive bool   `json:"is_active"` // 是否启用
}

// MjMaterial MidJourney 素材 VO
type MjMaterial struct {
	Id         uint   `json:"id"`
	CategoryId uint   `json:"category_id"`
	Title      string `json:"title"`
	Name       string `json:"name"`
	Image      string `json:"image"`     // 素材图片地址
	Preview    string `json:"preview"`   // 预览图地址
	Prompt     string `json:"prompt"`    // 提示词
	Type       string `json:"type"`      // 类型：txt2img、img2img、all
	IsActive   bool   `json:"is_active"` // 是否激活

}
