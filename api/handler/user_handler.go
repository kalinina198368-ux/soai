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
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"gorm.io/gorm"
)

type UserHandler struct {
	BaseHandler
	searcher       *xdb.Searcher
	redis          *redis.Client
	licenseService *service.LicenseService
	captcha        *service.CaptchaService
	userService    *service.UserService
}

func NewUserHandler(
	app *core.AppServer,
	db *gorm.DB,
	searcher *xdb.Searcher,
	client *redis.Client,
	captcha *service.CaptchaService,
	userService *service.UserService,
	licenseService *service.LicenseService) *UserHandler {
	return &UserHandler{
		BaseHandler:    BaseHandler{DB: db, App: app},
		searcher:       searcher,
		redis:          client,
		captcha:        captcha,
		licenseService: licenseService,
		userService:    userService,
	}
}

// Register user register
func (h *UserHandler) Register(c *gin.Context) {
	// parameters process
	var data struct {
		RegWay     string `json:"reg_way"`
		Username   string `json:"username"`
		Mobile     string `json:"mobile"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		Code       string `json:"code"`
		InviteCode string `json:"invite_code"`
		Key        string `json:"key,omitempty"`
		Dots       string `json:"dots,omitempty"`
		X          int    `json:"x,omitempty"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	if h.App.SysConfig.EnabledVerify && data.RegWay == "username" {
		var check bool
		if data.X != 0 {
			check = h.captcha.SlideCheck(data)
		} else {
			check = h.captcha.Check(data)
		}
		if !check {
			resp.ERROR(c, "请先完人机验证")
			return
		}
	}

	data.Password = strings.TrimSpace(data.Password)
	if len(data.Password) < 8 {
		resp.ERROR(c, "密码长度不能少于8个字符")
		return
	}

	// 检测最大注册人数
	var totalUser int64
	h.DB.Model(&model.User{}).Count(&totalUser)
	//logger.Info("可以注册的人数为", h.licenseService.GetLicense().Configs.UserNum)
	//if h.licenseService.GetLicense().Configs.UserNum > 0 && int(totalUser) >= h.licenseService.GetLicense().Configs.UserNum {
	//	resp.ERROR(c, "当前注册用户数已达上限，请请升级 License")
	//	return
	//}

	// 检查验证码
	var key string
	if data.RegWay == "email" {
		key = CodeStorePrefix + data.Email
		code, err := h.redis.Get(c, key).Result()
		if err != nil || code != data.Code {
			resp.ERROR(c, "验证码错误")
			return
		}
	} else if data.RegWay == "mobile" {
		key = CodeStorePrefix + data.Mobile
		code, err := h.redis.Get(c, key).Result()
		if err != nil || code != data.Code {
			resp.ERROR(c, "验证码错误")
			return
		}
	}

	// 验证邀请码
	inviteCode := model.InviteCode{}
	if data.InviteCode != "" {
		res := h.DB.Where("code = ?", data.InviteCode).First(&inviteCode)
		if res.Error != nil {
			resp.ERROR(c, "无效的邀请码")
			return
		}
	}

	salt := utils.RandString(8)
	user := model.User{
		Username: data.Username,
		Password: utils.GenPassword(data.Password, salt),

		Avatar:     "/images/avatar/user.png",
		Salt:       salt,
		Status:     true,
		ChatModels: utils.JsonEncode([]int{46, 36, 15, 42, 52, 55}),
		ChatRoles:  utils.JsonEncode([]string{"gpt"}), // 默认只订阅通用助手角色
		Power:      h.App.SysConfig.InitPower,
	}

	// check if the username is existing
	var item model.User
	session := h.DB.Session(&gorm.Session{})
	if data.Mobile != "" {
		session = session.Where("mobile = ?", data.Mobile)
		user.Username = data.Mobile
		user.Mobile = data.Mobile
	} else if data.Email != "" {
		session = session.Where("email = ?", data.Email)
		user.Username = data.Email
		user.Email = data.Email
	} else if data.Username != "" {
		session = session.Where("username = ?", data.Username)
	}
	session.First(&item)
	if item.Id > 0 {
		resp.ERROR(c, "该用户名已经被注册")
		return
	}

	//获取zjfid inviteCode.UserId
	zjFid := inviteCode.UserId
	jjFid := uint(0)
	var zjUser model.User

	//获取jjfid
	res := h.DB.Model(&model.User{}).Where("id = ?", zjFid).Find(&zjUser)
	if res.Error == nil {
		jjFid = zjUser.ZjFid

	}

	user.ZjFid = zjFid
	user.JjFid = jjFid

	// 被邀请人也获得赠送算力
	//if data.InviteCode != "" {
	//	user.Power += h.App.SysConfig.InvitePower
	//}
	//if h.licenseService.GetLicense().Configs.DeCopy {
	//	user.Nickname = fmt.Sprintf("soai@%d", utils.RandomNumber(6))
	//} else {
	//	user.Nickname = fmt.Sprintf("soai@%d", utils.RandomNumber(6))
	//}

	tx := h.DB.Begin()
	if err := tx.Create(&user).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	// 记录邀请关系
	if data.InviteCode != "" {
		// 增加邀请数量
		h.DB.Model(&model.InviteCode{}).Where("code = ?", data.InviteCode).UpdateColumn("reg_num", gorm.Expr("reg_num + ?", 1))
		if h.App.SysConfig.InvitePower > 0 {

			//err := h.userService.IncreasePower(int(inviteCode.UserId), h.App.SysConfig.InvitePower, model.PowerLog{
			//	Type:   types.PowerInvite,
			//	Model:  "Invite",
			//	Remark: fmt.Sprintf("邀请用户注册奖励，金额：%d，邀请码：%s，新用户：%s", h.App.SysConfig.InvitePower, inviteCode.Code, user.Username),
			//})
			//if err != nil {
			//	tx.Rollback()
			//	resp.ERROR(c, err.Error())
			//	return
			//}
			//

		}

		// 添加邀请记录
		err := tx.Create(&model.InviteLog{
			InviterId:  inviteCode.UserId,
			UserId:     user.Id,
			Username:   user.Username,
			InviteCode: inviteCode.Code,
			Remark:     fmt.Sprintf("奖励 %d 算力", h.App.SysConfig.InvitePower),
		}).Error
		if err != nil {
			tx.Rollback()
			resp.ERROR(c, err.Error())
			return
		}
	}
	tx.Commit()

	_ = h.redis.Del(c, key) // 注册成功，删除短信验证码
	// 自动登录创建 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.Id,
		"expired": time.Now().Add(time.Second * time.Duration(h.App.Config.Session.MaxAge)).Unix(),
	})
	tokenString, err := token.SignedString([]byte(h.App.Config.Session.SecretKey))
	if err != nil {
		resp.ERROR(c, "Failed to generate token, "+err.Error())
		return
	}
	// 保存到 redis
	key = fmt.Sprintf("users/%d", user.Id)
	if _, err := h.redis.Set(c, key, tokenString, 0).Result(); err != nil {
		resp.ERROR(c, "error with save token: "+err.Error())
		return
	}
	resp.SUCCESS(c, gin.H{"token": tokenString, "user_id": user.Id, "username": user.Username})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Key      string `json:"key,omitempty"`
		Dots     string `json:"dots,omitempty"`
		X        int    `json:"x,omitempty"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}
	verifyKey := fmt.Sprintf("users/verify/%s", data.Username)
	needVerify, err := h.redis.Get(c, verifyKey).Bool()

	if h.App.SysConfig.EnabledVerify && needVerify {
		var check bool
		if data.X != 0 {
			check = h.captcha.SlideCheck(data)
		} else {
			check = h.captcha.Check(data)
		}
		if !check {
			resp.ERROR(c, "请先完人机验证")
			return
		}
	}

	var user model.User
	res := h.DB.Where("username = ?", data.Username).First(&user)
	if res.Error != nil {
		h.redis.Set(c, verifyKey, true, 0)
		resp.ERROR(c, "用户名不存在")
		return
	}

	password := utils.GenPassword(data.Password, user.Salt)
	if password != user.Password {
		h.redis.Set(c, verifyKey, true, 0)
		resp.ERROR(c, "用户名或密码错误")
		return
	}

	if user.Status == false {
		resp.ERROR(c, "该用户已被禁止登录，请联系管理员")
		return
	}

	// 更新最后登录时间和IP
	user.LastLoginIp = c.ClientIP()
	user.LastLoginAt = time.Now().Unix()
	h.DB.Model(&user).Updates(user)

	h.DB.Create(&model.UserLoginLog{
		UserId:       user.Id,
		Username:     user.Username,
		LoginIp:      c.ClientIP(),
		LoginAddress: utils.Ip2Region(h.searcher, c.ClientIP()),
	})

	// 创建 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.Id,
		"expired": time.Now().Add(time.Second * time.Duration(h.App.Config.Session.MaxAge)).Unix(),
	})
	tokenString, err := token.SignedString([]byte(h.App.Config.Session.SecretKey))
	if err != nil {
		resp.ERROR(c, "Failed to generate token, "+err.Error())
		return
	}
	// 保存到 redis
	sessionKey := fmt.Sprintf("users/%d", user.Id)
	if _, err = h.redis.Set(c, sessionKey, tokenString, 0).Result(); err != nil {
		resp.ERROR(c, "error with save token: "+err.Error())
		return
	}
	// 移除登录行为验证码
	h.redis.Del(c, verifyKey)
	resp.SUCCESS(c, gin.H{"token": tokenString, "user_id": user.Id, "username": user.Username})
}

// Logout 注 销
func (h *UserHandler) Logout(c *gin.Context) {
	key := h.GetUserKey(c)
	if _, err := h.redis.Del(c, key).Result(); err != nil {
		logger.Error("error with delete session: ", err)
	}
	resp.SUCCESS(c)
}

// CLogin 第三方登录请求二维码
func (h *UserHandler) CLogin(c *gin.Context) {
	returnURL := h.GetTrim(c, "return_url")
	var res types.BizVo
	apiURL := fmt.Sprintf("%s/api/clogin/request", h.App.Config.ApiConfig.ApiURL)
	r, err := req.C().R().SetBody(gin.H{"login_type": "wx", "return_url": returnURL}).
		SetHeader("AppId", h.App.Config.ApiConfig.AppId).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", h.App.Config.ApiConfig.Token)).
		SetSuccessResult(&res).
		Post(apiURL)
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}
	if r.IsErrorState() {
		resp.ERROR(c, "error with login http status: "+r.Status)
		return
	}

	if res.Code != types.Success {
		resp.ERROR(c, "error with http response: "+res.Message)
		return
	}

	resp.SUCCESS(c, res.Data)
}

// CLoginCallback 第三方登录回调
func (h *UserHandler) CLoginCallback(c *gin.Context) {
	loginType := c.Query("login_type")

	code := c.Query("code")
	userId := h.GetInt(c, "user_id", 0)
	action := c.Query("action")

	var res types.BizVo
	apiURL := fmt.Sprintf("%s/api/clogin/info", h.App.Config.ApiConfig.ApiURL)
	r, err := req.C().R().SetBody(gin.H{"login_type": loginType, "code": code}).
		SetHeader("AppId", h.App.Config.ApiConfig.AppId).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", h.App.Config.ApiConfig.Token)).
		SetSuccessResult(&res).
		Post(apiURL)
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}
	if r.IsErrorState() {
		resp.ERROR(c, "error with login http status: "+r.Status)
		return
	}

	if res.Code != types.Success {
		resp.ERROR(c, "error with http response: "+res.Message)
		return
	}

	// login successfully
	data := res.Data.(map[string]interface{})
	var user model.User
	if action == "bind" && userId > 0 {
		err = h.DB.Where("openid", data["openid"]).First(&user).Error
		if err == nil {
			resp.ERROR(c, "该微信已经绑定其他账号，请先解绑")
			return
		}

		err = h.DB.Where("id", userId).First(&user).Error
		if err != nil {
			resp.ERROR(c, "绑定用户不存在")
			return
		}

		err = h.DB.Model(&user).UpdateColumn("openid", data["openid"]).Error
		if err != nil {
			resp.ERROR(c, "更新用户信息失败，"+err.Error())
			return
		}

		resp.SUCCESS(c, gin.H{"token": ""})
		return
	}

	session := gin.H{}
	tx := h.DB.Where("openid", data["openid"]).First(&user)
	if tx.Error != nil {
		// create new user
		var totalUser int64
		h.DB.Model(&model.User{}).Count(&totalUser)
		if h.licenseService.GetLicense().Configs.UserNum > 0 && int(totalUser) >= h.licenseService.GetLicense().Configs.UserNum {
			resp.ERROR(c, "当前注册用户数已达上限，请请升级 License")
			return
		}

		salt := utils.RandString(8)
		password := fmt.Sprintf("%d", utils.RandomNumber(8))
		user = model.User{
			Username:  fmt.Sprintf("%s@%d", loginType, utils.RandomNumber(10)),
			Password:  utils.GenPassword(password, salt),
			Avatar:    fmt.Sprintf("%s", data["avatar"]),
			Salt:      salt,
			Status:    true,
			ChatRoles: utils.JsonEncode([]string{"gpt"}), // 默认只订阅通用助手角色
			Power:     h.App.SysConfig.InitPower,
			OpenId:    fmt.Sprintf("%s", data["openid"]),
			Nickname:  fmt.Sprintf("%s", data["nickname"]),
		}

		tx = h.DB.Create(&user)
		if tx.Error != nil {
			resp.ERROR(c, "保存数据失败")
			logger.Error(tx.Error)
			return
		}
		session["username"] = user.Username
		session["password"] = password
	} else { // login directly
		// 更新最后登录时间和IP
		user.LastLoginIp = c.ClientIP()
		user.LastLoginAt = time.Now().Unix()
		h.DB.Model(&user).Updates(user)

		h.DB.Create(&model.UserLoginLog{
			UserId:       user.Id,
			Username:     user.Username,
			LoginIp:      c.ClientIP(),
			LoginAddress: utils.Ip2Region(h.searcher, c.ClientIP()),
		})
	}

	// 创建 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.Id,
		"expired": time.Now().Add(time.Second * time.Duration(h.App.Config.Session.MaxAge)).Unix(),
	})
	tokenString, err := token.SignedString([]byte(h.App.Config.Session.SecretKey))
	if err != nil {
		resp.ERROR(c, "Failed to generate token, "+err.Error())
		return
	}
	// 保存到 redis
	key := fmt.Sprintf("users/%d", user.Id)
	if _, err := h.redis.Set(c, key, tokenString, 0).Result(); err != nil {
		resp.ERROR(c, "error with save token: "+err.Error())
		return
	}
	session["token"] = tokenString
	resp.SUCCESS(c, session)
}

// Session 获取/验证会话
func (h *UserHandler) Session(c *gin.Context) {
	user, err := h.GetLoginUser(c)
	if err != nil {
		resp.NotAuth(c, err.Error())
		return
	}

	var userVo vo.User
	err = utils.CopyObject(user, &userVo)
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}
	// 用户 VIP 到期
	if user.ExpiredTime > 0 && user.ExpiredTime < time.Now().Unix() {
		h.DB.Model(&user).UpdateColumn("vip", false)
	}
	userVo.Id = user.Id
	resp.SUCCESS(c, userVo)

}

type userProfile struct {
	Id          uint   `json:"id"`
	Nickname    string `json:"nickname"`
	Username    string `json:"username"`
	Avatar      string `json:"avatar"`
	Power       int    `json:"power"`
	ExpiredTime int64  `json:"expired_time"`
	Vip         bool   `json:"vip"`
	Lev         int    `json:"lev"`
	Points      int    `json:"points"`
}

func (h *UserHandler) Profile(c *gin.Context) {
	user, err := h.GetLoginUser(c)
	if err != nil {
		resp.NotAuth(c)
		return
	}

	h.DB.First(&user, user.Id)
	var profile userProfile
	err = utils.CopyObject(user, &profile)
	if err != nil {
		logger.Error("对象拷贝失败：", err.Error())
		resp.ERROR(c, "获取用户信息失败")
		return
	}

	profile.Id = user.Id
	resp.SUCCESS(c, profile)
}

func (h *UserHandler) ProfileUpdate(c *gin.Context) {
	var data userProfile
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	user, err := h.GetLoginUser(c)
	if err != nil {
		resp.NotAuth(c)
		return
	}
	h.DB.First(&user, user.Id)
	user.Avatar = data.Avatar
	user.Nickname = data.Nickname
	res := h.DB.Updates(&user)
	if res.Error != nil {
		resp.ERROR(c, "更新用户信息失败")
		return
	}

	resp.SUCCESS(c)
}

// UpdatePass 更新密码
func (h *UserHandler) UpdatePass(c *gin.Context) {
	var data struct {
		OldPass  string `json:"old_pass"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	if len(data.Password) < 8 {
		resp.ERROR(c, "密码长度不能少于8个字符")
		return
	}

	user, err := h.GetLoginUser(c)
	if err != nil {
		resp.NotAuth(c)
		return
	}

	password := utils.GenPassword(data.OldPass, user.Salt)
	logger.Debugf(user.Salt, ",", user.Password, ",", password, ",", data.OldPass)
	if password != user.Password {
		resp.ERROR(c, "原密码错误")
		return
	}

	newPass := utils.GenPassword(data.Password, user.Salt)
	err = h.DB.Model(&user).UpdateColumn("password", newPass).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c)
}

// ResetPass 找回密码
func (h *UserHandler) ResetPass(c *gin.Context) {
	var data struct {
		Type     string `json:"type"`     // 验证类别：mobile, email
		Mobile   string `json:"mobile"`   // 手机号
		Email    string `json:"email"`    // 邮箱地址
		Code     string `json:"code"`     // 验证码
		Password string `json:"password"` // 新密码
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	session := h.DB.Session(&gorm.Session{})
	var key string
	if data.Type == "email" {
		session = session.Where("email", data.Email)
		key = CodeStorePrefix + data.Email
	} else if data.Type == "mobile" {
		session = session.Where("mobile", data.Mobile)
		key = CodeStorePrefix + data.Mobile
	} else {
		resp.ERROR(c, "验证类别错误")
		return
	}
	var user model.User
	err := session.First(&user).Error
	if err != nil {
		resp.ERROR(c, "用户不存在！")
		return
	}

	// 检查验证码
	code, err := h.redis.Get(c, key).Result()
	if err != nil || code != data.Code {
		resp.ERROR(c, "验证码错误")
		return
	}

	password := utils.GenPassword(data.Password, user.Salt)
	err = h.DB.Model(&user).UpdateColumn("password", password).Error
	if err != nil {
		resp.ERROR(c, err.Error())
	} else {
		h.redis.Del(c, key)
		resp.SUCCESS(c)
	}
}

// BindMobile 绑定手机号
func (h *UserHandler) BindMobile(c *gin.Context) {
	var data struct {
		Mobile string `json:"mobile"`
		Code   string `json:"code"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	// 检查验证码
	key := CodeStorePrefix + data.Mobile
	code, err := h.redis.Get(c, key).Result()
	if err != nil || code != data.Code {
		resp.ERROR(c, "验证码错误")
		return
	}

	// 检查手机号是否被其他账号绑定
	var item model.User
	res := h.DB.Where("mobile", data.Mobile).First(&item)
	if res.Error == nil {
		resp.ERROR(c, "该手机号已经绑定了其他账号，请更换手机号")
		return
	}

	userId := h.GetLoginUserId(c)

	err = h.DB.Model(&item).Where("id", userId).UpdateColumn("mobile", data.Mobile).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	_ = h.redis.Del(c, key) // 删除短信验证码
	resp.SUCCESS(c)
}

// BindEmail 绑定邮箱
func (h *UserHandler) BindEmail(c *gin.Context) {
	var data struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	// 检查验证码
	key := CodeStorePrefix + data.Email
	code, err := h.redis.Get(c, key).Result()
	if err != nil || code != data.Code {
		resp.ERROR(c, "验证码错误")
		return
	}

	// 检查手机号是否被其他账号绑定
	var item model.User
	res := h.DB.Where("email", data.Email).First(&item)
	if res.Error == nil {
		resp.ERROR(c, "该邮箱地址已经绑定了其他账号，请更邮箱地址")
		return
	}

	userId := h.GetLoginUserId(c)

	err = h.DB.Model(&item).Where("id", userId).UpdateColumn("email", data.Email).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	_ = h.redis.Del(c, key) // 删除短信验证码
	resp.SUCCESS(c)
}

//转账操作

// 转账逻辑处理 更新密码
func (h *UserHandler) Transfer(c *gin.Context) {
	//uid指的是用户的推广id
	var data struct {
		Uid      string `json:"uid"`
		Power    string `json:"power"`
		Password string `json:"password"`
	}

	logger.Info(data)
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	user, err := h.GetLoginUser(c)
	if err != nil {
		resp.NotAuth(c)
		return
	}
	//如果是普通用户,则直接弹出,普通用户不支持转账

	//校验密码
	password := utils.GenPassword(data.Password, user.Salt)
	if password != user.Password {
		resp.ERROR(c, "密码输入错误")
		return
	}

	//判断是否存在这样的用户

	inviteCode := model.InviteCode{}
	res0 := h.DB.Where("code = ?", data.Uid).First(&inviteCode)
	logger.Info("res0------> ", res0)

	//userid
	UserId := inviteCode.UserId

	if res0.Error != nil {
		resp.ERROR(c, "用户id不存在，请核对后再操作")
		return
	}

	//判断是否是因为跨部门原因导致的，跨部门不能转账
	//
	//// 限制转账关系：仅允许上下级（同链条）之间转账，禁止跨部门/兄弟横向转账
	//// 发送方（当前用户）的上下级集合（20代内）
	//senderDown := getTenLevelUserIds2(h.DB, h.GetLoginUserId(c))
	//// 发送方的上级链，从其直接上级开始向上（20代内）
	//senderAncestors := getUpIds(h.DB, uint(user.ZjFid))
	//
	//// 被转账用户ID（接收方）
	//receiverId := uint(UserId)
	//
	//containsUint := func(list []uint, target uint) bool {
	//	for _, id := range list {
	//		if id == target {
	//			return true
	//		}
	//	}
	//	return false
	//}
	//
	//allowed := containsUint(senderDown, receiverId) || containsUint(senderAncestors, receiverId)
	//if !allowed {
	//	resp.ERROR(c, "禁止跨部门转账")
	//	return
	//}

	currentUserId := h.GetLoginUserId(c)

	logger.Info("当前用户id为", currentUserId)

	//判断只有代理商用户才能进行转账操作
	//var items []model.Order
	//h.DB.Where("user_id = ? ", currentUserId).Find(&items)
	//logger.Info("当前用户已经投入的产品为", items)
	//if len(items) == 0 {
	//	resp.ERROR(c, "请先激活产品才能转账")
	//	return
	//}

	// 先将字符串类型转换为浮点数类型
	dataPower, err := strconv.ParseFloat(data.Power, 64)

	if err != nil {
		// 处理转换失败的情况
		resp.ERROR(c, "数据转换失败")
		return
	}

	// 然后进行比较
	if int(dataPower) > user.Power {
		resp.ERROR(c, "用户余额不足")
		return
	}

	//获取当前系统最小转账额度
	//最低金额
	var miniPower = h.App.SysConfig.VipMonthPower

	logger.Info(miniPower)
	logger.Info(dataPower)
	//logger.Info( int(dataPower) < miniPower)

	//最高转账和最低转账限制
	//
	//if dataPower < miniPower {
	//	resp.ERROR(c, fmt.Sprintf("最低转账额度为 %d 算力", int(miniPower)))
	//	return
	//}

	//更新转账人余额
	res := h.DB.Model(&model.User{}).Where("id = ?", user.Id).UpdateColumn("power", gorm.Expr("power - ?", data.Power))

	//更新收款人余额
	h.DB.Model(&model.User{}).Where("id = ?", UserId).UpdateColumn("power", gorm.Expr("power + ?", data.Power))

	//添加logs
	//str := data.Power
	//num, err := strconv.Atoi(str)
	//if err != nil {
	//	fmt.Printf("转换出错: %v\n", err)
	//	return
	//}
	//fmt.Printf("转换后的整数为: %d\n", num)

	//res := h.DB.Model(&model.User{}).Where("id = ?", user.Id).UpdateColumn("power", gorm.Expr("power - ?", power))
	if res.Error == nil {
		// 记录算力消费日志
		var u model.User
		h.DB.Where("id", user.Id).First(&u)

		var u2 model.User
		h.DB.Where("id", UserId).First(&u2)

		h.DB.Create(&model.PowerLog{
			UserId:    user.Id,
			Username:  user.Username,
			Type:      types.PowerSend, //转账
			Amount:    int(dataPower),
			Mark:      types.PowerSub,
			Balance:   u.Power,
			Model:     fmt.Sprintf("转账:收款人(%s) ", u2.Username), //session.Model.Value,
			Remark:    fmt.Sprintf("用户转出算力：%s ", data.Power),
			CreatedAt: time.Now(),
		})

		h.DB.Create(&model.PowerLog{
			UserId:    u2.Id,
			Username:  u2.Username,
			Type:      types.PowerRece, //收款
			Amount:    int(dataPower),
			Mark:      types.PowerAdd,
			Balance:   u2.Power,
			Model:     fmt.Sprintf("收款:转账人(%s)", user.Username), //session.Model.Value,
			Remark:    fmt.Sprintf("用户收到算力转账：%s ", data.Power),
			CreatedAt: time.Now(),
		})

	} else {
		logger.Error("更新数据库失败: ", res.Error)
		resp.ERROR(c, "更新数据库失败")
		return

	}

	resp.SUCCESS(c)
}
