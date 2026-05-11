package vo

type UploadCodeImage struct {
	Id         uint   `json:"id"`
	UserId     uint   `json:"user_id"`
	Code       string `json:"code"`
	URL        string `json:"url"`
	ObjKey     string `json:"obj_key"`
	Name       string `json:"name"`
	Ext        string `json:"ext"`
	Size       int64  `json:"size"`
	UploaderIP string `json:"uploader_ip"`
	CreatedAt  int64  `json:"created_at"`
}
