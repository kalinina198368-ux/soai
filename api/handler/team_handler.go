package handler

// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// * Copyright 2023 The Geek-AI Authors. All rights reserved.
// * Use of this source code is governed by a Apache-2.0 license
// * that can be found in the LICENSE file.
// * @Author yangjian102621@163.com
// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

import (
	"geekai/core"
	"geekai/store/model"
	"geekai/store/vo"
	"geekai/utils"
	"geekai/utils/resp"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InviteHandler 团队相关
type TeamHandler struct {
	BaseHandler
}

func NewTeamHandler(app *core.AppServer, db *gorm.DB) *TeamHandler {
	return &TeamHandler{BaseHandler: BaseHandler{App: app, DB: db}}
}

// Code 获取当前用户邀请码
func (h *TeamHandler) Code(c *gin.Context) {
	userId := h.GetLoginUserId(c)
	var inviteCode model.InviteCode
	res := h.DB.Where("user_id = ?", userId).First(&inviteCode)
	// 如果邀请码不存在，则创建一个
	if res.Error != nil {
		code := strings.ToUpper(utils.RandString(8))
		for {
			res = h.DB.Where("code = ?", code).First(&inviteCode)
			if res.Error != nil { // 不存在相同的邀请码则退出
				break
			}
		}
		inviteCode.UserId = userId
		inviteCode.Code = code
		h.DB.Create(&inviteCode)
	}
	//获取激活,充值 ,团队
	var total int64

	query := `
    SELECT COUNT(*) 
    FROM chatgpt_users 
    WHERE zj_fid = ? 
    OR jj_fid = ? 
`
	// 执行查询并获取计数
	h.DB.Raw(query, userId, userId).Scan(&total)
	//logger.Info("团队数量为", total)

	//获取激活总数
	var jhTotal int64
	result := h.DB.Raw("SELECT COUNT(*) FROM chatgpt_users WHERE zj_fid = ? AND is_jh = 1", userId).Scan(&jhTotal)
	if result.Error != nil {
		resp.ERROR(c, "获取邀请数据失败")
		return

	}

	var codeVo vo.InviteCode
	err := utils.CopyObject(inviteCode, &codeVo)
	if err != nil {
		resp.ERROR(c, "拷贝对象失败")
		return
	}

	codeVo.Hits = int(total) //团队数量
	codeVo.JhNum = jhTotal   //激活数量

	resp.SUCCESS(c, codeVo)
}

// List Log 用户邀请记录
func (h *TeamHandler) List(c *gin.Context) {
	page := h.GetInt(c, "page", 1)
	pageSize := h.GetInt(c, "page_size", 20)
	userId := h.GetLoginUserId(c)
	session := h.DB.Session(&gorm.Session{}).Where("inviter_id = ?", userId)
	var total int64
	session.Model(&model.InviteLog{}).Count(&total)
	var items []model.InviteLog
	var list = make([]vo.InviteLog, 0)
	offset := (page - 1) * pageSize
	res := session.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items)
	if res.Error == nil {
		for _, item := range items {
			var v vo.InviteLog
			err := utils.CopyObject(item, &v)
			if err == nil {
				v.Id = item.Id
				v.CreatedAt = item.CreatedAt.Unix()
				list = append(list, v)
			} else {
				logger.Error(err)
			}
		}
	}
	resp.SUCCESS(c, vo.NewPage(total, page, pageSize, list))
}

// Hits 访问邀请码
func (h *TeamHandler) Hits(c *gin.Context) {
	code := c.Query("code")
	h.DB.Model(&model.InviteCode{}).Where("code = ?", code).UpdateColumn("hits", gorm.Expr("hits + ?", 1))
	resp.SUCCESS(c)
}

// Stats 获取团队统计
func (h *TeamHandler) Stats(c *gin.Context) {
	userId := h.GetLoginUserId(c)

	// 总团队人数（直推 + 间推）
	var total int64
	query := `
		SELECT COUNT(DISTINCT id) 
		FROM chatgpt_users 
		WHERE zj_fid = ? OR jj_fid = ?
	`
	h.DB.Raw(query, userId, userId).Scan(&total)

	// 直推人数
	var direct int64
	h.DB.Table("chatgpt_users").Where("zj_fid = ?", userId).Count(&direct)

	//激活量
	var jhNum int64
	h.DB.Table("chatgpt_users").Where("is_jh=1 and  zj_fid = ?", userId).Count(&jhNum)

	// 间推人数
	var indirect int64
	h.DB.Table("chatgpt_users").Where("jj_fid = ?", userId).Count(&indirect)

	result := map[string]interface{}{
		"total":    total,
		"direct":   direct,
		"jhNum":    jhNum,
		"indirect": indirect,
	}

	resp.SUCCESS(c, result)
}

// Direct 获取直推团队列表
func (h *TeamHandler) Direct(c *gin.Context) {
	userId := h.GetLoginUserId(c)

	var users []model.User
	res := h.DB.Where("zj_fid = ?", userId).
		Select("id, username, nickname, avatar, power, created_at").
		Order("id DESC").
		Find(&users)

	if res.Error != nil {
		resp.ERROR(c, "获取直推团队失败")
		return
	}

	var list = make([]map[string]interface{}, 0)
	for _, user := range users {
		item := map[string]interface{}{
			"id":         user.Id,
			"username":   user.Username,
			"nickname":   user.Nickname,
			"avatar":     user.Avatar,
			"power":      user.Power,
			"created_at": user.CreatedAt.Unix(),
		}
		list = append(list, item)
	}

	resp.SUCCESS(c, list)
}

// Indirect 获取间推团队列表
func (h *TeamHandler) Indirect(c *gin.Context) {
	userId := h.GetLoginUserId(c)

	var users []model.User
	res := h.DB.Where("jj_fid = ?", userId).
		Select("id, username, nickname, avatar, power, created_at").
		Order("id DESC").
		Find(&users)

	if res.Error != nil {
		resp.ERROR(c, "获取间推团队失败")
		return
	}

	var list = make([]map[string]interface{}, 0)
	for _, user := range users {
		item := map[string]interface{}{
			"id":         user.Id,
			"username":   user.Username,
			"nickname":   user.Nickname,
			"avatar":     user.Avatar,
			"power":      user.Power,
			"created_at": user.CreatedAt.Unix(),
		}
		list = append(list, item)
	}

	resp.SUCCESS(c, list)
}
