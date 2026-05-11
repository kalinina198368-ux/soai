package vo

// SoraMaterialCategory Sora 视频素材分类 VO
type SoraMaterialCategory struct {
	Id       uint   `json:"id"`
	Name     string `json:"name"`      // 分类内部标识
	Title    string `json:"title"`     // 分类展示名称
	SortNum  int    `json:"sort"`      // 排序值
	IsActive bool   `json:"is_active"` // 是否启用
}

// SoraMaterial Sora 视频素材 VO
type SoraMaterial struct {
	Id         uint   `json:"id"`
	CategoryId uint   `json:"category_id"`
	Title      string `json:"title"`
	Name       string `json:"name"`
	Video      string `json:"video"`      // 素材视频地址
	Image      string `json:"image"`      // 素材封面图地址
	Prompt     string `json:"prompt"`     // 提示词
	IsActive   bool   `json:"is_active"`  // 是否激活
}

