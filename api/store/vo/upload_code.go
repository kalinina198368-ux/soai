package vo

type UploadCode struct {
	Id        uint   `json:"id"`
	UserId    uint   `json:"user_id"`
	Code      string `json:"code"`
	CreatedAt int64  `json:"created_at"`
}
