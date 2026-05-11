package vo

type ShareClickLog struct {
	Id         uint   `json:"id"`
	InviteCode string `json:"invite_code"`
	Ip         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
	CreatedAt  int64  `json:"created_at"`
	Rewarded   bool   `json:"rewarded"` // 是否已奖励（24小时内）
}

