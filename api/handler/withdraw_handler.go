package handler

// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// * Copyright 2023 The Geek-AI Authors. All rights reserved.
// * Use of this source code is governed by a Apache-2.0 license
// * that can be found in the LICENSE file.
// * @Author yangjian102621@163.com
// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

import (
	"fmt"
	"geekai/core"
	"geekai/core/types"
	"geekai/store/model"
	"geekai/utils/resp"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 提现
type WithdrawHandler struct {
	BaseHandler
}

func NewWithdrawHandler(app *core.AppServer, db *gorm.DB) *WithdrawHandler {
	return &WithdrawHandler{BaseHandler: BaseHandler{App: app, DB: db}}
}

// Apply 提交提现申请
func (h *WithdrawHandler) Apply(c *gin.Context) {
	userId := h.GetLoginUserId(c)

	var data struct {
		Power       int    `json:"power"`
		Amount      int    `json:"amount"`
		AccountName string `json:"account_name"`
		QrcodeUrl   string `json:"qrcode_url"`
		Remark      string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	// 验证必填字段
	if data.Power <= 0 {
		resp.ERROR(c, "提现算力必须大于0")
		return
	}
	if data.AccountName == "" {
		resp.ERROR(c, "请输入收款人姓名")
		return
	}
	if data.QrcodeUrl == "" {
		resp.ERROR(c, "请上传收款码")
		return
	}

	// 获取用户信息，验证算力
	var user model.User
	res := h.DB.First(&user, userId)
	if res.Error != nil {
		resp.ERROR(c, "用户不存在")
		return
	}

	// 检查算力是否足够
	if user.Power < data.Power {
		resp.ERROR(c, "算力不足")
		return
	}

	// 检查最低提现算力（可以从系统配置获取，这里暂时硬编码）
	minWithdrawPower := 100
	if data.Power < minWithdrawPower {
		resp.ERROR(c, "最低提现算力为100")
		return
	}

	// 获取用户名
	username := user.Username

	// 创建提现记录
	now := time.Now().Unix()
	withdrawData := map[string]interface{}{
		"user_id":      userId,
		"username":     username,
		"power":        data.Power,
		"amount":       data.Amount, //折合成人民币的金额
		"account_name": data.AccountName,
		"qrcode_url":   data.QrcodeUrl,
		"status":       0, // 0: 待审核
		"remark":       data.Remark,
		"created_at":   now,
		"updated_at":   now,
	}

	err := h.DB.Table("chatgpt_withdraws").Create(withdrawData).Error
	if err != nil {
		logger.Error("创建提现记录失败: ", err)
		resp.ERROR(c, "提交提现申请失败")
		return
	}

	// 扣除用户算力（提现申请时先冻结算力）
	// 注意：这里可以选择立即扣除，或者等审核通过后再扣除
	// 如果立即扣除，需要更新用户算力
	err = h.DB.Model(&user).Update("points", gorm.Expr("points - ?", data.Power)).Error
	if err != nil {
		logger.Error("扣除算力失败: ", err)
		// 如果扣除失败，可以选择回滚提现记录，或者只记录错误
		// 这里暂时只记录错误，不阻止提现申请
	}

	// 重新查询用户信息，获取扣除后的实时余额
	var updatedUser model.User
	if err := h.DB.First(&updatedUser, userId).Error; err != nil {
		logger.Error("获取用户实时余额失败: ", err)
		// 如果查询失败，使用计算值作为备用
		updatedUser.Points = user.Points - data.Power
	}

	//提现日志
	h.DB.Create(&model.PointsLog{
		UserId:    user.Id,
		Username:  user.Username,
		Type:      types.PowerWithdraw, //提现
		Amount:    data.Power,
		Mark:      types.PowerSub,
		Balance:   updatedUser.Points, // 使用扣除后的实时余额
		Model:     fmt.Sprintf("提现扣款"),
		Remark:    fmt.Sprintf("用户提现扣款:(%s) ,代理商等级:(%d)", user.Username, user.Lev), //session.Model.Value,
		CreatedAt: time.Now(),
	})

	resp.SUCCESS(c)
}

// History 获取提现记录列表
func (h *WithdrawHandler) History(c *gin.Context) {
	userId := h.GetLoginUserId(c)

	// 定义提现记录结构
	type WithdrawRecord struct {
		Id          uint   `gorm:"column:id"`
		UserId      uint   `gorm:"column:user_id"`
		Username    string `gorm:"column:username"`
		Power       int    `gorm:"column:power"`
		AccountName string `gorm:"column:account_name"`
		QrcodeUrl   string `gorm:"column:qrcode_url"`
		Status      int    `gorm:"column:status"`
		Remark      string `gorm:"column:remark"`
		CreatedAt   int64  `gorm:"column:created_at"`
		UpdatedAt   int64  `gorm:"column:updated_at"`
	}

	var items []WithdrawRecord
	res := h.DB.Table("chatgpt_withdraws").
		Where("user_id = ?", userId).
		Select("id, user_id, username, power, account_name, qrcode_url, status, remark, created_at, updated_at").
		Order("id DESC").
		Find(&items)

	if res.Error != nil {
		resp.ERROR(c, "获取提现记录失败")
		return
	}

	var list = make([]map[string]interface{}, 0)
	for _, item := range items {
		record := map[string]interface{}{
			"id":           item.Id,
			"user_id":      item.UserId,
			"username":     item.Username,
			"power":        item.Power,
			"account_name": item.AccountName,
			"qrcode_url":   item.QrcodeUrl,
			"status":       item.Status,
			"remark":       item.Remark,
			"created_at":   item.CreatedAt,
			"updated_at":   item.UpdatedAt,
		}
		list = append(list, record)
	}

	resp.SUCCESS(c, list)
}
