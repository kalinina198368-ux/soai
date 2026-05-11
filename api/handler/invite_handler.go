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
	"geekai/service"
	"geekai/store/model"
	"geekai/store/vo"
	"geekai/utils"
	"geekai/utils/resp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InviteHandler 用户邀请
type InviteHandler struct {
	BaseHandler
	userService *service.UserService
}

func NewInviteHandler(app *core.AppServer, db *gorm.DB, userService *service.UserService) *InviteHandler {
	return &InviteHandler{
		BaseHandler: BaseHandler{App: app, DB: db},
		userService: userService,
	}
}

// Code 获取当前用户邀请码
func (h *InviteHandler) Code(c *gin.Context) {
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
func (h *InviteHandler) List(c *gin.Context) {
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
func (h *InviteHandler) Hits(c *gin.Context) {
	code := c.Query("code")
	h.DB.Model(&model.InviteCode{}).Where("code = ?", code).UpdateColumn("hits", gorm.Expr("hits + ?", 1))
	resp.SUCCESS(c)
}

// ShareClick 分享链接点击，给分享者添加1个算力
func (h *InviteHandler) ShareClick(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		resp.ERROR(c, "邀请码不能为空")
		return
	}

	// 查找邀请码对应的用户
	var inviteCode model.InviteCode
	res := h.DB.Where("code = ?", code).First(&inviteCode)
	if res.Error != nil {
		resp.ERROR(c, "无效的邀请码")
		return
	}

	//自己不能给自己分享
	if inviteCode.UserId == h.GetLoginUserId(c) {
		resp.ERROR(c, "不能给自己分享")
		return
	}

	//后台配置是否ip和24小时锁定

	// 防止重复点击（24小时内同一IP和邀请码只奖励一次）
	clientIP := c.ClientIP()
	oneDayAgo := time.Now().Add(-24 * time.Hour)

	// 使用事务确保检查、创建日志和增加算力的原子性
	tx := h.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 在事务中检查最近24小时内是否有相同IP和邀请码的记录（双重检查锁定）
	var recentClick model.ShareClickLog
	checkRes := tx.Where("invite_code = ? AND ip = ? AND created_at > ?", code, clientIP, oneDayAgo).First(&recentClick)

	// 获取被助力用户信息
	var targetUser model.User
	var username string = ""
	if err := h.DB.Where("id = ?", inviteCode.UserId).First(&targetUser).Error; err == nil {
		if targetUser.Nickname != "" {
			username = targetUser.Nickname
		} else {
			username = targetUser.Username
		}
	}

	// 如果24小时内已经点击过（通过IP判断），则认为已经助力过
	alreadyHelped := checkRes.Error == nil

	if checkRes.Error == nil {
		// 24小时内已经点击过，不重复奖励
		tx.Rollback()
		resp.SUCCESS(c, gin.H{
			"message":       "已记录",
			"rewarded":      false,
			"alreadyHelped": alreadyHelped,
			"username":      username,
		})
		return
	}

	// 记录点击日志
	clickLog := model.ShareClickLog{
		InviteCode: code,
		UserId:     inviteCode.UserId,
		Ip:         clientIP,
		UserAgent:  c.GetHeader("User-Agent"),
		CreatedAt:  time.Now(),
	}
	if err := tx.Create(&clickLog).Error; err != nil {
		tx.Rollback()
		logger.Error("记录分享点击日志失败: ", err)
		// 如果是因为重复记录导致的错误，返回已记录
		resp.SUCCESS(c, gin.H{
			"message":       "已记录",
			"rewarded":      false,
			"alreadyHelped": alreadyHelped,
			"username":      username,
		})
		return
	}

	// 提交日志记录的事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败: ", err)
		resp.ERROR(c, "处理失败")
		return
	}

	// 给分享者添加1个算力（IncreasePower内部已有事务和锁保护）
	if h.userService != nil {
		err := h.userService.IncreasePower(int(inviteCode.UserId), 2, model.PowerLog{
			Type:   types.PowerInvite,
			Model:  "Share",
			Remark: fmt.Sprintf("分享链接被点击奖励，邀请码：%s，点击IP：%s", code, clientIP),
		})
		if err != nil {
			logger.Error("增加算力失败: ", err)
			resp.ERROR(c, "奖励失败")
			return
		}
	}

	// 获取被助力用户信息（如果之前没有获取）
	if username == "" {
		var targetUser model.User
		if err := h.DB.Where("id = ?", inviteCode.UserId).First(&targetUser).Error; err == nil {
			if targetUser.Nickname != "" {
				username = targetUser.Nickname
			} else {
				username = targetUser.Username
			}
		}
	}

	resp.SUCCESS(c, gin.H{
		"message":       "奖励成功",
		"rewarded":      true,
		"alreadyHelped": false,
		"username":      username,
	})
}

// ShareClickLogList 获取助力日志列表
func (h *InviteHandler) ShareClickLogList(c *gin.Context) {
	page := h.GetInt(c, "page", 1)
	pageSize := h.GetInt(c, "page_size", 20)
	userId := h.GetLoginUserId(c)

	// 获取当前用户的邀请码
	var inviteCode model.InviteCode
	res := h.DB.Where("user_id = ?", userId).First(&inviteCode)
	if res.Error != nil {
		resp.ERROR(c, "未找到邀请码")
		return
	}

	// 查询该邀请码的所有点击记录
	session := h.DB.Session(&gorm.Session{}).Where("invite_code = ?", inviteCode.Code)
	var total int64
	session.Model(&model.ShareClickLog{}).Count(&total)

	var items []model.ShareClickLog
	var list = make([]vo.ShareClickLog, 0)
	offset := (page - 1) * pageSize
	res = session.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items)

	if res.Error == nil {
		oneDayAgo := time.Now().Add(-24 * time.Hour)
		for _, item := range items {
			var v vo.ShareClickLog
			v.Id = item.Id
			v.InviteCode = item.InviteCode
			// 隐藏IP的部分信息，保护隐私（只显示前3段，最后一段用*代替）
			ipParts := strings.Split(item.Ip, ".")
			if len(ipParts) == 4 {
				v.Ip = fmt.Sprintf("%s.%s.%s.*", ipParts[0], ipParts[1], ipParts[2])
			} else {
				v.Ip = item.Ip // IPv6或其他格式，直接显示
			}
			v.UserAgent = item.UserAgent
			v.CreatedAt = item.CreatedAt.Unix()
			// 判断是否已奖励（24小时内）
			v.Rewarded = item.CreatedAt.After(oneDayAgo)
			list = append(list, v)
		}
	}

	resp.SUCCESS(c, vo.NewPage(total, page, pageSize, list))
}

//团队列表
