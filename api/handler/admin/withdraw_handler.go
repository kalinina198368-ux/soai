package admin

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
	"geekai/handler"
	"geekai/store/model"
	"geekai/store/vo"
	"geekai/utils"
	"geekai/utils/resp"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 提现管理端
type WithdrawHandler struct {
	handler.BaseHandler
}

func NewWithdrawHandler(app *core.AppServer, db *gorm.DB) *WithdrawHandler {
	return &WithdrawHandler{BaseHandler: handler.BaseHandler{App: app, DB: db}}
}

func (h *WithdrawHandler) List(c *gin.Context) {
	var data struct {
		Username  string   `json:"username"`
		Status    int      `json:"status"`
		StartDate string   `json:"start_date"`
		EndDate   string   `json:"end_date"`
		CreatedAt []string `json:"created_at"` // 兼容前端可能传递的 created_at 数组
		Page      int      `json:"page"`
		PageSize  int      `json:"page_size"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	session := h.DB.Session(&gorm.Session{})

	// 用户名筛选
	if data.Username != "" {
		session = session.Where("username LIKE ?", "%"+data.Username+"%")
	}

	// 日期筛选（支持两种格式）
	var startDate, endDate string
	if len(data.CreatedAt) == 2 {
		startDate = data.CreatedAt[0]
		endDate = data.CreatedAt[1]
	} else if data.StartDate != "" && data.EndDate != "" {
		startDate = data.StartDate
		endDate = data.EndDate
	}
	if startDate != "" && endDate != "" {
		start := utils.Str2stamp(startDate + " 00:00:00")
		end := utils.Str2stamp(endDate + " 23:59:59")
		session = session.Where("created_at >= ? AND created_at <= ?", start, end)
	}

	// 状态筛选
	if data.Status >= 0 {
		session = session.Where("status", data.Status)
	}

	var total int64
	session.Table("chatgpt_withdraws").Count(&total)

	// 定义提现记录结构
	type WithdrawRecord struct {
		Id          uint   `gorm:"column:id"`
		UserId      uint   `gorm:"column:user_id"`
		Username    string `gorm:"column:username"`
		Power       int    `gorm:"column:power"`
		Amount      int    `gorm:"column:amount"`
		AccountName string `gorm:"column:account_name"`
		QrcodeUrl   string `gorm:"column:qrcode_url"`
		Status      int    `gorm:"column:status"`
		Remark      string `gorm:"column:remark"`
		CreatedAt   int64  `gorm:"column:created_at"`
		UpdatedAt   int64  `gorm:"column:updated_at"`
	}

	var items []WithdrawRecord
	offset := (data.Page - 1) * data.PageSize
	res := session.Table("chatgpt_withdraws").
		Select("id, user_id, username, power,amount, account_name, qrcode_url, status, remark, created_at, updated_at").
		Order("id DESC").
		Offset(offset).
		Limit(data.PageSize).
		Find(&items)

	var list = make([]map[string]interface{}, 0)
	if res.Error == nil {
		for _, item := range items {
			record := map[string]interface{}{
				"id":           item.Id,
				"user_id":      item.UserId,
				"username":     item.Username,
				"power":        item.Power,
				"amount":       item.Amount,
				"account_name": item.AccountName,
				"qrcode_url":   item.QrcodeUrl,
				"status":       item.Status,
				"remark":       item.Remark,
				"created_at":   item.CreatedAt,
				"updated_at":   item.UpdatedAt,
			}
			list = append(list, record)
		}
	}
	resp.SUCCESS(c, vo.NewPage(total, data.Page, data.PageSize, list))
}

func (h *WithdrawHandler) Remove(c *gin.Context) {
	id := h.GetInt(c, "id", 0)

	if id <= 0 {
		resp.ERROR(c, "参数错误")
		return
	}

	// 获取提现记录信息
	type WithdrawRecord struct {
		UserId uint `gorm:"column:user_id"`
		Power  int  `gorm:"column:power"`
		Status int  `gorm:"column:status"`
	}

	var record WithdrawRecord
	res := h.DB.Table("chatgpt_withdraws").Where("id = ?", id).First(&record)
	if res.Error != nil {
		resp.ERROR(c, "记录不存在！")
		return
	}

	// 如果状态是待审核（0），需要返还算力
	if record.Status == 0 {
		// 返还算力给用户
		err := h.DB.Model(&model.User{}).Where("id = ?", record.UserId).
			Update("points", gorm.Expr("points + ?", record.Power)).Error
		if err != nil {
			logger.Error("返还算力失败: ", err)
			resp.ERROR(c, "返还算力失败")
			return
		}
	}

	// 删除提现记录
	err := h.DB.Table("chatgpt_withdraws").Where("id = ?", id).Delete(nil).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c)
}

// Approve 审核通过提现申请
func (h *WithdrawHandler) Approve(c *gin.Context) {
	var data struct {
		Id uint `json:"id"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	// 检查记录是否存在且状态为待审核
	type WithdrawRecord struct {
		Status int `gorm:"column:status"`
	}
	var record WithdrawRecord
	res := h.DB.Table("chatgpt_withdraws").Where("id = ?", data.Id).First(&record)
	if res.Error != nil {
		resp.ERROR(c, "记录不存在！")
		return
	}

	// 只能审核待审核状态的申请
	if record.Status != 0 {
		resp.ERROR(c, "只能审核待审核状态的申请")
		return
	}

	// 更新状态和更新时间
	now := time.Now().Unix()
	updates := map[string]interface{}{
		"status":     1, // 审核通过
		"updated_at": now,
	}

	err := h.DB.Table("chatgpt_withdraws").Where("id = ?", data.Id).Updates(updates).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c)
}

// Reject 拒绝提现申请
func (h *WithdrawHandler) Reject(c *gin.Context) {
	var data struct {
		Id     uint   `json:"id"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	// 获取提现记录信息
	type WithdrawRecord struct {
		UserId   uint   `gorm:"column:user_id"`
		Username string `gorm:"column:username"`
		Power    int    `gorm:"column:power"`
		Status   int    `gorm:"column:status"`
	}
	var record WithdrawRecord
	res := h.DB.Table("chatgpt_withdraws").Where("id = ?", data.Id).First(&record)
	if res.Error != nil {
		resp.ERROR(c, "记录不存在！")
		return
	}

	// 只能拒绝待审核状态的申请
	if record.Status != 0 {
		resp.ERROR(c, "只能拒绝待审核状态的申请")
		return
	}

	// 返还算力给用户（因为申请时已经扣除了算力）
	err := h.DB.Model(&model.User{}).Where("id = ?", record.UserId).
		Update("points", gorm.Expr("points + ?", record.Power)).Error
	if err != nil {
		logger.Error("返还算力失败: ", err)
		resp.ERROR(c, "返还算力失败")
		return
	}

	//var userModel model.User
	//
	////获取用户模型相关
	//res2 := h.DB.Model(&model.User{}).Where("id = ?", record.UserId).First(&userModel)
	//if res2 != nil {
	//	logger.Error("获取用户模型相关: ", err)
	//	resp.ERROR(c, "获取用户模型相关")
	//	return
	//}

	var userModel model.User
	//获取用户模型相关
	if err := h.DB.Model(&model.User{}).Where("id = ?", record.UserId).First(&userModel).Error; err != nil {
		logger.Error("获取用户模型相关: ", err)
		resp.ERROR(c, "获取用户模型相关")
		return
	}

	//提现退款日志
	h.DB.Create(&model.PointsLog{
		UserId:    record.UserId,
		Username:  record.Username,
		Type:      types.PowerWithdrawNG, //退款
		Amount:    record.Power,
		Mark:      types.PowerAdd,
		Balance:   userModel.Points,
		Model:     fmt.Sprintf("提现退款"),
		Remark:    fmt.Sprintf("用户提现退款:(%s) ,代理商等级:(%d)", record.Username, userModel.Lev), //session.Model.Value,
		CreatedAt: time.Now(),
	})

	// 更新状态、备注和更新时间
	now := time.Now().Unix()
	updates := map[string]interface{}{
		"status":     3, // 已拒绝
		"updated_at": now,
	}
	if data.Reason != "" {
		updates["remark"] = data.Reason
	}

	err = h.DB.Table("chatgpt_withdraws").Where("id = ?", data.Id).Updates(updates).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c)
}

// Paid 标记已打款
func (h *WithdrawHandler) Paid(c *gin.Context) {
	var data struct {
		Id uint `json:"id"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	// 检查记录是否存在且状态为审核通过
	type WithdrawRecord struct {
		Status int `gorm:"column:status"`
	}
	var record WithdrawRecord
	res := h.DB.Table("chatgpt_withdraws").Where("id = ?", data.Id).First(&record)
	if res.Error != nil {
		resp.ERROR(c, "记录不存在！")
		return
	}

	// 只能标记审核通过状态的申请为已打款
	if record.Status != 1 {
		resp.ERROR(c, "只能标记审核通过状态的申请为已打款")
		return
	}

	// 更新状态和更新时间
	now := time.Now().Unix()
	updates := map[string]interface{}{
		"status":     2, // 已打款
		"updated_at": now,
	}

	err := h.DB.Table("chatgpt_withdraws").Where("id = ?", data.Id).Updates(updates).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c)
}
