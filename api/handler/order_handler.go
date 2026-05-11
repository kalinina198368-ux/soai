package handler

// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// * Copyright 2023 The Geek-AI Authors. All rights reserved.
// * Use of this source code is governed by a Apache-2.0 license
// * that can be found in the LICENSE file.
// * @Author yangjian102621@163.com
// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"geekai/core"
	"geekai/core/types"
	"geekai/store/model"
	"geekai/store/vo"
	"geekai/utils"
	"geekai/utils/resp"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	BaseHandler
}

const defaultDiscountPercent = 100.0

func NewOrderHandler(app *core.AppServer, db *gorm.DB) *OrderHandler {
	return &OrderHandler{BaseHandler: BaseHandler{App: app, DB: db}}
}

// List 订单列表
func (h *OrderHandler) List(c *gin.Context) {
	page := h.GetInt(c, "page", 1)
	pageSize := h.GetInt(c, "page_size", 20)
	userId := h.GetLoginUserId(c)
	//session := h.DB.Session(&gorm.Session{}).Where("user_id = ? AND status = ?", userId, types.OrderPaidSuccess)
	session := h.DB.Session(&gorm.Session{}).Where("user_id = ? ", userId)
	var total int64
	session.Model(&model.Order{}).Count(&total)
	var items []model.Order
	var list = make([]vo.Order, 0)
	offset := (page - 1) * pageSize
	res := session.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items)
	if res.Error == nil {
		for _, item := range items {
			var order vo.Order
			err := utils.CopyObject(item, &order)
			if err == nil {
				order.Id = item.Id
				order.CreatedAt = item.CreatedAt.Unix()
				order.UpdatedAt = item.UpdatedAt.Unix()
				payMethod, ok := types.PayMethods[item.PayWay]
				if !ok {
					payMethod = item.PayWay
				}
				payName, ok := types.PayNames[item.PayType]
				if !ok {
					payName = item.PayWay
				}
				order.PayMethod = payMethod
				order.PayName = payName
				list = append(list, order)
			} else {
				logger.Error(err)
			}
		}
	}
	resp.SUCCESS(c, vo.NewPage(total, page, pageSize, list))
}

// Query 查询订单状态
func (h *OrderHandler) Query(c *gin.Context) {
	orderNo := h.GetTrim(c, "order_no")
	var order model.Order
	res := h.DB.Where("order_no = ?", orderNo).First(&order)
	if res.Error != nil {
		resp.ERROR(c, "Order not found")
		return
	}

	if order.Status == types.OrderPaidSuccess {
		resp.SUCCESS(c, gin.H{"status": order.Status})
		return
	}

	counter := 0
	for {
		time.Sleep(time.Second)
		var item model.Order
		h.DB.Where("order_no = ?", orderNo).First(&item)
		if counter >= 15 || item.Status == types.OrderPaidSuccess || item.Status != order.Status {
			order.Status = item.Status
			break
		}
		counter++
	}

	resp.SUCCESS(c, gin.H{"status": order.Status})
}

// 模拟订单提交到后台里面去
//
//func (h *OrderHandler) MockPaid(c *gin.Context) {
//
//	// 获取当前登录用户所有的聊天会话
//	user, err := h.GetLoginUser(c)
//	if err != nil {
//		resp.NotAuth(c)
//		return
//	}
//
//	//前端提交过来的参数
//	type MockPaidRequest struct {
//		OrderNo   string  `json:"order_no"`
//		Amount    float64 `json:"amount"`
//		Subject   string  `json:"subject"`
//		UserID    uint    `json:"user_id"`
//		ProductID uint    `json:"product_id"`
//		PayTime   int64   `json:"pay_time"`
//		PayMethod string  `json:"pay_method"`
//		Status    string  `json:"status"`
//	}
//
//	var req MockPaidRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		c.JSON(400, gin.H{
//			"code":    1,
//			"message": "invalid request: " + err.Error(),
//		})
//		return
//	}
//
//	//判断当前用户是否一致
//	if user.Id != req.UserID {
//		resp.ERROR(c, "当前用户鉴权失败！")
//		return
//	}
//
//	// 基础校验
//	if req.OrderNo == "" || req.UserID == 0 || req.ProductID == 0 || req.Amount < 0 {
//		c.JSON(400, gin.H{
//			"code":    1,
//			"message": "missing or invalid required fields",
//		})
//		return
//	}
//
//	// 插入数据库
//	orderModel := model.Order{
//		UserId:    req.UserID,
//		ProductId: req.ProductID,
//		Username:  user.Username,
//		OrderNo:   req.OrderNo,
//		Subject:   req.Subject,
//		Amount:    req.Amount,
//		Status:    types.OrderPaidSuccess,
//		Remark:    "",
//		PayWay:    "wx",
//		PayType:   "wx_h5",
//		PayTime:   req.PayTime,
//	}
//
//	tx := h.DB.Create(&orderModel)
//	logger.Info(tx)
//
//	//获取对应产品的算力值
//	var product model.Product
//	err1 := h.DB.Where("id", req.ProductID).First(&product).Error
//	if err1 != nil {
//		resp.ERROR(c, "产品不存在")
//		return
//	}
//
//	//更新用户的字段为已经激活
//	h.DB.Model(&model.User{}).Where("id", user.Id).UpdateColumn("is_jh", 1)
//
//	//获取当前用户的直推上级和间推上级
//	zjfid := user.ZjFid
//	jjfid := user.JjFid
//
//	// 用户自身折扣
//	userDiscount := h.getDiscountByLevel(user.Lev)
//
//	//添加用户的算力和日志（购买用户自己，按折扣计算算力）
//	h.AddPowerAndLog(req.UserID, product.Power, 0, req.UserID, userDiscount)
//
//	//添加直推的积分和日志（直推上级，佣金）
//	if zjfid > 0 {
//		h.AddPowerAndLog(zjfid, product.Power, 1, req.UserID, defaultDiscountPercent)
//	}
//
//	//添加间推的积分和日志（间推上级，佣金）
//	if jjfid > 0 {
//		h.AddPowerAndLog(jjfid, product.Power, 2, req.UserID, defaultDiscountPercent)
//	}
//
//	if err := h.getTeamReward(user.Id, req.Amount); err != nil {
//		logger.Error(fmt.Sprintf("用户 %d 团队奖发放失败: %v", user.Id, err))
//	}
//
//	if err := h.upgread(user.Id); err != nil {
//		logger.Error(fmt.Sprintf("用户 %d 等级升级失败: %v", user.Id, err))
//	}
//
//	resp.SUCCESS(c)
//
//}

// 真实支付
func (h *OrderHandler) PayPaid(c *gin.Context) {

	// 获取当前登录用户所有的聊天会话
	user, err := h.GetLoginUser(c)
	if err != nil {
		resp.NotAuth(c)
		return
	}

	//前端提交过来的参数
	type MockPaidRequest struct {
		OrderNo   string  `json:"order_no"`
		Amount    float64 `json:"amount"`
		Subject   string  `json:"subject"`
		UserID    uint    `json:"user_id"`
		ProductID uint    `json:"product_id"`
		PayTime   int64   `json:"pay_time"`
		PayMethod string  `json:"pay_method"`
		Status    string  `json:"status"`
	}

	var req MockPaidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    1,
			"message": "invalid request: " + err.Error(),
		})
		return
	}

	//判断当前用户是否一致
	if user.Id != req.UserID {
		resp.ERROR(c, "当前用户鉴权失败！")
		return
	}

	// 基础校验
	if req.OrderNo == "" || req.UserID == 0 || req.ProductID == 0 || req.Amount < 0 {
		c.JSON(400, gin.H{
			"code":    1,
			"message": "missing or invalid required fields",
		})
		return
	}

	// 插入数据库（当前仅记录订单，后续通过支付回调再发放算力和奖励）
	orderModel := model.Order{
		UserId:    req.UserID,
		ProductId: req.ProductID,
		Username:  user.Username,
		OrderNo:   req.OrderNo,
		Subject:   req.Subject,
		Amount:    req.Amount,
		Status:    types.OrderNotPaid,
		Remark:    "",
		PayWay:    "wx",
		PayType:   "wx_h5",
		PayTime:   req.PayTime,
	}

	tx := h.DB.Create(&orderModel)
	logger.Info(tx)

	// 以下逻辑改为由支付回调统一处理，避免未付款先发放算力和奖励
	// //获取对应产品的算力值
	// var product model.Product
	// err1 := h.DB.Where("id", req.ProductID).First(&product).Error
	// if err1 != nil {
	// 	resp.ERROR(c, "产品不存在")
	// 	return
	// }
	//
	// //更新用户的字段为已经激活
	// h.DB.Model(&model.User{}).Where("id", user.Id).UpdateColumn("is_jh", 1)
	//
	// //获取当前用户的直推上级和间推上级
	// zjfid := user.ZjFid
	// jjfid := user.JjFid
	//
	// // 用户自身折扣
	// userDiscount := h.getDiscountByLevel(user.Lev)
	//
	// //添加用户的算力和日志（购买用户自己，按折扣计算算力）
	// h.AddPowerAndLog(req.UserID, product.Power, 0, req.UserID, userDiscount)
	//
	// //添加直推的积分和日志（直推上级，佣金）
	// if zjfid > 0 {
	// 	h.AddPowerAndLog(zjfid, product.Power, 1, req.UserID, defaultDiscountPercent)
	// }
	//
	// //添加间推的积分和日志（间推上级，佣金）
	// if jjfid > 0 {
	// 	h.AddPowerAndLog(jjfid, product.Power, 2, req.UserID, defaultDiscountPercent)
	// }
	//
	// if err := h.getTeamReward(user.Id, req.Amount); err != nil {
	// 	logger.Error(fmt.Sprintf("用户 %d 团队奖发放失败: %v", user.Id, err))
	// }
	//
	// if err := h.upgread(user.Id); err != nil {
	// 	logger.Error(fmt.Sprintf("用户 %d 等级升级失败: %v", user.Id, err))
	// }

	resp.SUCCESS(c)

}

// 添加算力和日志等情况
// userId: 要添加算力的用户ID
// powerAmount: 基础算力值（产品算力）
// relationType: 关系类型 0=购买用户自己, 1=直推上级, 2=间推上级
// buyerUserId: 购买用户ID（用于记录日志）
// discount: 各等级折扣（百分比），默认100
func (h *OrderHandler) AddPowerAndLog(userId uint, powerAmount int, relationType int, buyerUserId uint, discount float64) {
	// 检查用户ID是否有效
	if userId == 0 {
		return
	}

	if discount <= 0 {
		discount = defaultDiscountPercent
	}

	// 获取用户信息
	var targetUser model.User
	if err := h.DB.Where("id = ?", userId).First(&targetUser).Error; err != nil {
		logger.Error(fmt.Sprintf("获取用户信息失败: %v", err))
		return
	}

	var actualPower int
	var logType types.PowerType
	var remark string
	var modelName string

	// 根据关系类型计算实际算力
	if relationType == 0 {
		// 购买用户自己，根据折扣发放算力
		actualPower = h.applyDiscount(powerAmount, discount)
		logType = types.PowerRecharge // 充值类型
		if discount != defaultDiscountPercent {
			remark = fmt.Sprintf("购买产品获得算力， %d折", int(discount))
		} else {
			remark = "购买产品获得算力"
		}
		modelName = "订单充值"
	} else {
		// 上级用户，需要根据等级计算佣金
		// 获取系统配置
		var config model.Config
		if err := h.DB.Where("marker = ?", "system").First(&config).Error; err != nil {
			logger.Error(fmt.Sprintf("获取系统配置失败: %v", err))
			return
		}

		// 解析系统配置
		var systemConfig map[string]interface{}
		if err := json.Unmarshal([]byte(config.Config), &systemConfig); err != nil {
			logger.Error(fmt.Sprintf("解析系统配置失败: %v", err))
			return
		}

		starLevel := 0

		starLevel = int(targetUser.Lev)

		// 根据等级和关系类型获取佣金比例
		var commissionRate float64
		configKey := ""
		if relationType == 1 {
			// 直推
			switch starLevel {
			case 9: // 代理商
				configKey = "agent_gold_direct_commission"
			case 8: // 代理商
				configKey = "agent_silver_direct_commission"
			case 7: // 代理商
				configKey = "agent_bronze_direct_commission"
			case 3: // 三星
				configKey = "star3_direct_commission"
			case 2: // 二星
				configKey = "star2_direct_commission"
			case 1: // 一星
				configKey = "star1_direct_commission"
			default: // 普通会员
				configKey = "normal_direct_commission"
			}
		} else if relationType == 2 {
			// 间推
			switch starLevel {
			case 9: // 代理商
				configKey = "agent_gold_indirect_commission"
			case 8: // 代理商
				configKey = "agent_silver_indirect_commission"
			case 7: // 代理商
				configKey = "agent_bronze_indirect_commission"
			case 3: // 三星
				configKey = "star3_indirect_commission"
			case 2: // 二星
				configKey = "star2_indirect_commission"
			case 1: // 一星
				configKey = "star1_indirect_commission"
			default: // 普通会员
				configKey = "normal_indirect_commission"
			}
		}

		// 获取佣金比例（百分比）
		if configKey != "" {
			if rate, ok := systemConfig[configKey]; ok {
				if rateVal, ok := rate.(float64); ok {
					commissionRate = rateVal
				}
			}
		}

		// 计算实际佣金算力（按百分比计算）
		actualPower = int(float64(powerAmount) * commissionRate / 100.0)
		logType = types.PowerSale // 佣金
		relationName := "直推"
		if relationType == 2 {
			relationName = "间推"
		}
		remark = fmt.Sprintf("%s佣金，购买用户ID: %d，产品算力: %d，佣金比例: %.2f%%", relationName, buyerUserId, powerAmount, commissionRate)
		modelName = "分销佣金"
	}

	// 如果算力为0，不处理
	if actualPower == 0 {
		return
	}

	// 记录算力日志
	//powerLog := model.PowerLog{
	//	UserId:    targetUser.Id,
	//	Username:  targetUser.Username,
	//	Type:      logType,
	//	Amount:    actualPower,
	//	Balance:   newPower,
	//	Mark:      types.PowerAdd, // 1表示增加
	//	Model:     modelName,
	//	Remark:    remark,
	//	CreatedAt: time.Now(),
	//}

	//如果是自己则直接记录算力,否则记录积分
	if relationType == 0 {

		// 更新用户算力
		oldPower := targetUser.Power
		newPower := oldPower + actualPower
		if err := h.DB.Model(&targetUser).Update("power", newPower).Error; err != nil {
			logger.Error(fmt.Sprintf("更新用户算力失败: %v", err))
			return
		}

		// 记录算力日志
		powerLog := model.PowerLog{
			UserId:    targetUser.Id,
			Username:  targetUser.Username,
			Type:      logType,
			Amount:    actualPower,
			Balance:   newPower,
			Mark:      types.PowerAdd, // 1表示增加
			Model:     modelName,
			Remark:    remark,
			CreatedAt: time.Now(),
		}

		if err := h.DB.Create(&powerLog).Error; err != nil {
			logger.Error(fmt.Sprintf("创建算力日志失败: %v", err))
			// 如果日志创建失败，回滚算力
			h.DB.Model(&targetUser).Update("power", oldPower)
			return
		}
		logger.Info(fmt.Sprintf("用户 %d (%s) 获得算力 %d，余额: %d -> %d", targetUser.Id, targetUser.Username, actualPower, oldPower, newPower))

	} else {
		// 更新用户积分
		oldPower := targetUser.Points
		newPower := oldPower + actualPower
		if err := h.DB.Model(&targetUser).Update("points", newPower).Error; err != nil {
			logger.Error(fmt.Sprintf("更新用户积分失败: %v", err))
			return
		}

		// 记录积分日志
		pointsLog := model.PointsLog{
			UserId:    targetUser.Id,
			Username:  targetUser.Username,
			Type:      logType,
			Amount:    actualPower,
			Balance:   newPower,
			Mark:      types.PowerAdd, // 1表示增加
			Model:     modelName,
			Remark:    remark,
			CreatedAt: time.Now(),
		}

		if err := h.DB.Create(&pointsLog).Error; err != nil {
			logger.Error(fmt.Sprintf("创建积分日志失败: %v", err))
			// 如果日志创建失败，回滚算力
			h.DB.Model(&targetUser).Update("points", oldPower)
			return
		}

		logger.Info(fmt.Sprintf("用户 %d (%s) 获得积分 %d，余额: %d -> %d", targetUser.Id, targetUser.Username, actualPower, oldPower, newPower))

	}

}

func (h *OrderHandler) applyDiscount(powerAmount int, discount float64) int {
	if powerAmount <= 0 {
		return 0
	}
	actual := int(math.Round(float64(powerAmount) / (discount / 100.0)))
	if actual < 0 {
		return 0
	}
	return actual
}

func (h *OrderHandler) getDiscountByLevel(level int) float64 {
	cfg := h.App.SysConfig
	switch level {
	case 9:
		return h.normalizeDiscount(float64(cfg.AgentGoldDiscount))
	case 8:
		return h.normalizeDiscount(float64(cfg.AgentSilverDiscount))
	case 7:
		return h.normalizeDiscount(float64(cfg.AgentBronzeDiscount))
	case 3:
		return h.normalizeDiscount(float64(cfg.Star3Discount))
	case 2:
		return h.normalizeDiscount(float64(cfg.Star2Discount))
	case 1:
		return h.normalizeDiscount(float64(cfg.Star1Discount))
	default:
		return defaultDiscountPercent
	}
}

func (h *OrderHandler) getDiscountByUserID(userId uint) float64 {
	if userId == 0 {
		return defaultDiscountPercent
	}
	var u model.User
	if err := h.DB.Select("id, lev").Where("id = ?", userId).First(&u).Error; err != nil {
		logger.Warnf("获取用户 %d 折扣失败: %v", userId, err)
		return defaultDiscountPercent
	}
	return h.getDiscountByLevel(u.Lev)
}

func (h *OrderHandler) normalizeDiscount(value float64) float64 {
	if value <= 0 {
		return defaultDiscountPercent
	}
	return value
}

// 升级
func (h *OrderHandler) upgread(userId uint) error {
	if userId == 0 {
		return fmt.Errorf("invalid user id")
	}

	var user model.User
	if err := h.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		return fmt.Errorf("fetch user failed: %w", err)
	}

	var total sql.NullFloat64
	err := h.DB.Model(&model.Order{}).
		Select("SUM(amount)").
		Where("user_id = ? AND status = ?", userId, types.OrderPaidSuccess).
		Scan(&total).Error
	if err != nil {
		return fmt.Errorf("fetch total recharge failed: %w", err)
	}

	//最大的单笔充值
	var maxSingle sql.NullFloat64
	err = h.DB.Model(&model.Order{}).
		Select("MAX(amount)").
		Where("user_id = ? AND status = ?", userId, types.OrderPaidSuccess).
		Scan(&maxSingle).Error
	if err != nil {
		return fmt.Errorf("fetch max single recharge failed: %w", err)
	}

	var totalRecharge float64
	if total.Valid {
		totalRecharge = total.Float64
	}

	var singleRecharge float64
	if maxSingle.Valid {
		singleRecharge = maxSingle.Float64
	}

	cfg := h.App.SysConfig
	newLevel := user.Lev

	applyLevel := func(required int, level int, actual float64) {
		if required <= 0 || newLevel >= level {
			return
		}

		if actual >= float64(required) {
			newLevel = level
		}
	}

	// 先判断代理商等级（7-9）
	applyLevel(cfg.AgentGoldUpgradeRecharge, 9, singleRecharge)
	applyLevel(cfg.AgentSilverUpgradeRecharge, 8, singleRecharge)
	applyLevel(cfg.AgentBronzeUpgradeRecharge, 7, singleRecharge)

	// 普通会员等级（0-3），仅在尚未满足代理商等级的情况下判断
	if newLevel < 7 {
		applyLevel(cfg.Star3UpgradeRecharge, 3, totalRecharge)
		applyLevel(cfg.Star2UpgradeRecharge, 2, totalRecharge)
		applyLevel(cfg.Star1UpgradeRecharge, 1, totalRecharge)
	}

	if newLevel == user.Lev {
		return nil
	}

	if err := h.DB.Model(&model.User{}).Where("id = ?", user.Id).Update("lev", newLevel).Error; err != nil {
		return fmt.Errorf("update user level failed: %w", err)
	}

	logger.Infof("用户 %d 等级升级，累计充值 %.2f，单笔最大充值 %.2f，等级 %d -> %d", user.Id, totalRecharge, singleRecharge, user.Lev, newLevel)
	return nil
}

// 团队奖励 获取团队奖励
func (h *OrderHandler) getTeamReward(userId uint, rechargeAmount float64) error {
	if userId == 0 || rechargeAmount <= 0 {
		return nil
	}

	var user model.User
	if err := h.DB.Where("id = ?", userId).Select("id, zj_fid").First(&user).Error; err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	cfg := h.App.SysConfig
	levelRate := map[int]float64{
		7: float64(cfg.AgentBronzeTeamReward),
		8: float64(cfg.AgentSilverTeamReward),
		9: float64(cfg.AgentGoldTeamReward),
	}

	maxRate := 0.0
	for _, rate := range levelRate {
		if rate > maxRate {
			maxRate = rate
		}
	}

	parentId := user.ZjFid
	prevRate := 0.0
	visited := make(map[uint]struct{})

	for parentId > 0 {
		if _, ok := visited[parentId]; ok {
			break
		}
		visited[parentId] = struct{}{}

		var parent model.User
		if err := h.DB.Where("id = ?", parentId).First(&parent).Error; err != nil {
			return fmt.Errorf("获取上级用户 %d 失败: %w", parentId, err)
		}

		rate, isAgent := levelRate[parent.Lev]
		if isAgent && rate > prevRate {
			diff := rate - prevRate
			reward := int(math.Round(rechargeAmount * diff / 100.0))
			if reward > 0 {
				oldPoints := parent.Points
				newPoints := oldPoints + reward
				if err := h.DB.Model(&parent).Update("points", newPoints).Error; err != nil {
					return fmt.Errorf("更新用户积分失败: %w", err)
				}

				pointsLog := model.PointsLog{
					UserId:    parent.Id,
					Username:  parent.Username,
					Type:      types.PowerTeamReward,
					Amount:    reward,
					Balance:   newPoints,
					Mark:      types.PowerAdd,
					Model:     "团队奖",
					Remark:    fmt.Sprintf("下级用户 %d 充值 %.2f，级差 %.2f%% 团队奖励", userId, rechargeAmount, diff),
					CreatedAt: time.Now(),
				}

				if err := h.DB.Create(&pointsLog).Error; err != nil {
					h.DB.Model(&parent).Update("points", oldPoints)
					return fmt.Errorf("创建积分日志失败: %w", err)
				}

				logger.Infof("用户 %d 获得团队奖励 %d（级差 %.2f%%），来自下级 %d", parent.Id, reward, diff, userId)
			}
			prevRate = rate
		}

		if prevRate >= maxRate {
			break
		}

		parentId = parent.ZjFid
	}

	return nil
}
