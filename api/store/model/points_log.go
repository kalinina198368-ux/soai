package model

import (
	"geekai/core/types"
	"time"
)

// PointsLog 积分日志记录模型
type PointsLog struct {
	Id        uint `gorm:"primarykey;column:id"`
	UserId    uint
	Username  string
	Type      types.PowerType
	Amount    int
	Balance   int
	Model     string          // 模型
	Remark    string          // 备注
	Mark      types.PowerMark // 资金类型
	CreatedAt time.Time
}
