package model

// TeachingRecord 用户教学创作草稿/记录（脚本、TTS、提示词、配图素材 JSON）
type TeachingRecord struct {
	BaseModel
	UserId       uint   `gorm:"index;column:user_id"`
	Topic        string `gorm:"size:255;column:topic"`
	TitleSummary string `gorm:"size:255;column:title_summary"`
	ScriptJSON   string `gorm:"type:mediumtext;column:script_json"`
	TtsJSON      string `gorm:"type:mediumtext;column:tts_json"`
	PromptsJSON  string `gorm:"type:mediumtext;column:prompts_json"`
	AssetsJSON   string `gorm:"type:mediumtext;column:assets_json"`
	MaxStep      int    `gorm:"column:max_step"` // 当前已推进到的步骤 0-3
}

func (TeachingRecord) TableName() string {
	return "teaching_records"
}
