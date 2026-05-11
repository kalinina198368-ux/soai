package handler

// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// * Copyright 2023 The Geek-AI Authors. All rights reserved.
// * Use of this source code is governed by a Apache-2.0 license
// * that can be found in the LICENSE file.
// * @Author yangjian102621@163.com
// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"geekai/core"
	"geekai/core/types"
	"geekai/service"
	"geekai/service/payment"
	"geekai/store/model"
	"geekai/utils"
	"geekai/utils/resp"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PayWay struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// PaymentHandler 支付服务回调 handler
type PaymentHandler struct {
	BaseHandler
	alipayService    *payment.AlipayService
	huPiPayService   *payment.HuPiPayService
	geekPayService   *payment.GeekPayService
	wechatPayService *payment.WechatPayService
	snowflake        *service.Snowflake
	userService      *service.UserService
	fs               embed.FS
	lock             sync.Mutex
	signKey          string // 用来签名的随机秘钥
	orderHandler     *OrderHandler
}

func NewPaymentHandler(
	server *core.AppServer,
	alipayService *payment.AlipayService,
	huPiPayService *payment.HuPiPayService,
	geekPayService *payment.GeekPayService,
	wechatPayService *payment.WechatPayService,
	db *gorm.DB,
	userService *service.UserService,
	snowflake *service.Snowflake,
	fs embed.FS) *PaymentHandler {
	return &PaymentHandler{
		alipayService:    alipayService,
		huPiPayService:   huPiPayService,
		geekPayService:   geekPayService,
		wechatPayService: wechatPayService,
		snowflake:        snowflake,
		userService:      userService,
		fs:               fs,
		lock:             sync.Mutex{},
		BaseHandler: BaseHandler{
			App: server,
			DB:  db,
		},
		signKey:      utils.RandString(32),
		orderHandler: NewOrderHandler(server, db),
	}
}

func (h *PaymentHandler) Pay(c *gin.Context) {
	var data struct {
		PayWay    string `json:"pay_way"`
		PayType   string `json:"pay_type"`
		ProductId int    `json:"product_id"`
		UserId    int    `json:"user_id"`
		Device    string `json:"device"`
		Host      string `json:"host"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	var product model.Product
	err := h.DB.Where("id", data.ProductId).First(&product).Error
	if err != nil {
		resp.ERROR(c, "Product not found")
		return
	}

	orderNo, err := h.snowflake.Next(false)
	if err != nil {
		resp.ERROR(c, "error with generate trade no: "+err.Error())
		return
	}
	var user model.User
	err = h.DB.Where("id", data.UserId).First(&user).Error
	if err != nil {
		resp.NotAuth(c)
		return
	}

	amount := product.Discount
	var payURL, returnURL, notifyURL string
	switch data.PayWay {
	case "alipay":
		if h.App.Config.AlipayConfig.NotifyURL != "" { // 用于本地调试支付
			notifyURL = h.App.Config.AlipayConfig.NotifyURL
		} else {
			notifyURL = fmt.Sprintf("%s/api/payment/notify/alipay", data.Host)
		}
		if h.App.Config.AlipayConfig.ReturnURL != "" { // 用于本地调试支付
			returnURL = h.App.Config.AlipayConfig.ReturnURL
		} else {
			returnURL = fmt.Sprintf("%s/payReturn", data.Host)
		}
		money := fmt.Sprintf("%.2f", amount)
		if data.Device == "wechat" {
			payURL, err = h.alipayService.PayMobile(payment.AlipayParams{
				OutTradeNo: orderNo,
				Subject:    product.Name,
				TotalFee:   money,
				ReturnURL:  returnURL,
				NotifyURL:  notifyURL,
			})
		} else {
			payURL, err = h.alipayService.PayPC(payment.AlipayParams{
				OutTradeNo: orderNo,
				Subject:    product.Name,
				TotalFee:   money,
				ReturnURL:  returnURL,
				NotifyURL:  notifyURL,
			})
		}

		if err != nil {
			resp.ERROR(c, "error with generate pay url: "+err.Error())
			return
		}
		break
	case "wechat":
		if h.App.Config.WechatPayConfig.NotifyURL != "" {
			notifyURL = h.App.Config.WechatPayConfig.NotifyURL
		} else {
			notifyURL = fmt.Sprintf("%s/api/payment/notify/wechat", data.Host)
		}
		if data.Device == "wechat" {
			payURL, err = h.wechatPayService.PayUrlH5(payment.WechatPayParams{
				OutTradeNo: orderNo,
				TotalFee:   int(amount * 100),
				Subject:    product.Name,
				NotifyURL:  notifyURL,
				ClientIP:   c.ClientIP(),
			})
		} else {
			payURL, err = h.wechatPayService.PayUrlNative(payment.WechatPayParams{
				OutTradeNo: orderNo,
				TotalFee:   int(amount * 100),
				Subject:    product.Name,
				NotifyURL:  notifyURL,
			})
		}
		if err != nil {
			resp.ERROR(c, err.Error())
			return
		}
		break
	case "hupi":
		if h.App.Config.HuPiPayConfig.NotifyURL != "" {
			notifyURL = h.App.Config.HuPiPayConfig.NotifyURL
		} else {
			notifyURL = fmt.Sprintf("%s/api/payment/notify/hupi", data.Host)
		}
		if h.App.Config.HuPiPayConfig.ReturnURL != "" {
			returnURL = h.App.Config.HuPiPayConfig.ReturnURL
		} else {
			returnURL = fmt.Sprintf("%s/payReturn", data.Host)
		}
		r, err := h.huPiPayService.Pay(payment.HuPiPayParams{
			Version:      "1.1",
			TradeOrderId: orderNo,
			TotalFee:     fmt.Sprintf("%f", amount),
			Title:        product.Name,
			NotifyURL:    notifyURL,
			ReturnURL:    returnURL,
			WapName:      "GeekAI助手",
		})
		if err != nil {
			resp.ERROR(c, err.Error())
			return
		}
		payURL = r.URL
		break
	case "geek":
		if h.App.Config.GeekPayConfig.NotifyURL != "" {
			notifyURL = h.App.Config.GeekPayConfig.NotifyURL
		} else {
			notifyURL = fmt.Sprintf("%s/api/payment/notify/geek", data.Host)
		}
		if h.App.Config.GeekPayConfig.ReturnURL != "" {
			data.Host = utils.GetBaseURL(h.App.Config.GeekPayConfig.ReturnURL)
		}
		if data.Device == "wechat" { // 微信客户端打开，调回手机端用户中心页面
			returnURL = fmt.Sprintf("%s/mobile/profile", data.Host)
		} else {
			returnURL = fmt.Sprintf("%s/payReturn", data.Host)
		}
		params := payment.GeekPayParams{
			OutTradeNo: orderNo,
			Method:     "web",
			Name:       product.Name,
			Money:      fmt.Sprintf("%f", amount),
			ClientIP:   c.ClientIP(),
			Device:     data.Device,
			Type:       data.PayType,
			ReturnURL:  returnURL,
			NotifyURL:  notifyURL,
		}

		res, err := h.geekPayService.Pay(params)
		if err != nil {
			resp.ERROR(c, err.Error())
			return
		}
		payURL = res.PayURL
	default:
		resp.ERROR(c, "不支持的支付渠道")
		return
	}

	// 创建订单
	remark := types.OrderRemark{
		Days:     product.Days,
		Power:    product.Power,
		Name:     product.Name,
		Price:    product.Price,
		Discount: product.Discount,
	}
	order := model.Order{
		UserId:    user.Id,
		Username:  user.Username,
		ProductId: product.Id,
		OrderNo:   orderNo,
		Subject:   product.Name,
		Amount:    amount,
		Status:    types.OrderNotPaid,
		PayWay:    data.PayWay,
		PayType:   data.PayType,
		Remark:    utils.JsonEncode(remark),
	}
	err = h.DB.Create(&order).Error
	if err != nil {
		resp.ERROR(c, "error with create order: "+err.Error())
		return
	}
	resp.SUCCESS(c, payURL)
}

// 异步通知回调公共逻辑
func (h *PaymentHandler) notify(orderNo string, tradeNo string) error {
	var order model.Order
	err := h.DB.Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return fmt.Errorf("error with fetch order: %v", err)
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	// 已支付订单，直接返回
	if order.Status == types.OrderPaidSuccess {
		return nil
	}

	var user model.User
	err = h.DB.First(&user, order.UserId).Error
	if err != nil {
		return fmt.Errorf("error with fetch user info: %v", err)
	}

	// 根据订单中的产品ID获取产品信息（从产品表获取算力等信息，而不是依赖订单的 remark 字段）
	var product model.Product
	err = h.DB.First(&product, order.ProductId).Error
	if err != nil {
		return fmt.Errorf("error with fetch product info: %v", err)
	}

	// 增加用户算力（使用产品自身的算力值）
	err = h.userService.IncreasePower(int(order.UserId), product.Power, model.PowerLog{
		Type:   types.PowerRecharge,
		Model:  order.PayWay,
		Remark: fmt.Sprintf("充值算力，金额：%f，订单号：%s", order.Amount, order.OrderNo),
	})
	if err != nil {
		return err
	}

	// 更新订单状态
	order.PayTime = time.Now().Unix()
	order.Status = types.OrderPaidSuccess
	order.TradeNo = tradeNo
	err = h.DB.Updates(&order).Error
	if err != nil {
		return fmt.Errorf("error with update order info: %v", err)
	}

	// 更新产品销量
	err = h.DB.Model(&model.Product{}).Where("id = ?", order.ProductId).
		UpdateColumn("sales", gorm.Expr("sales + ?", 1)).Error
	if err != nil {
		return fmt.Errorf("error with update product sales: %v", err)
	}

	if err := h.orderHandler.upgread(order.UserId); err != nil {
		logger.Error(fmt.Sprintf("用户 %d 等级升级失败: %v", order.UserId, err))
	}

	return nil
}

// GetPayWays 获取支付方式
func (h *PaymentHandler) GetPayWays(c *gin.Context) {
	payWays := make([]gin.H, 0)
	if h.App.Config.AlipayConfig.Enabled {
		payWays = append(payWays, gin.H{"pay_way": "alipay", "pay_type": "alipay"})
	}
	if h.App.Config.HuPiPayConfig.Enabled {
		payWays = append(payWays, gin.H{"pay_way": "hupi", "pay_type": "wxpay"})
	}
	if h.App.Config.GeekPayConfig.Enabled {
		for _, v := range h.App.Config.GeekPayConfig.Methods {
			payWays = append(payWays, gin.H{"pay_way": "geek", "pay_type": v})
		}
	}
	if h.App.Config.WechatPayConfig.Enabled {
		payWays = append(payWays, gin.H{"pay_way": "wechat", "pay_type": "wxpay"})
	}
	resp.SUCCESS(c, payWays)
}

// HuPiPayNotify 虎皮椒支付异步回调
func (h *PaymentHandler) HuPiPayNotify(c *gin.Context) {
	err := c.Request.ParseForm()
	if err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	orderNo := c.Request.Form.Get("trade_order_id")
	tradeNo := c.Request.Form.Get("open_order_id")
	logger.Infof("收到虎皮椒订单支付回调，%+v", c.Request.Form)

	if err = h.huPiPayService.Check(orderNo); err != nil {
		logger.Error("订单校验失败：", err)
		c.String(http.StatusOK, "fail")
		return
	}

	err = h.notify(orderNo, tradeNo)
	if err != nil {
		logger.Error(err)
		c.String(http.StatusOK, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}

// AlipayNotify 支付宝支付回调
func (h *PaymentHandler) AlipayNotify(c *gin.Context) {
	err := c.Request.ParseForm()
	if err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	result := h.alipayService.TradeVerify(c.Request)
	logger.Infof("收到支付宝商号订单支付回调：%+v", result)
	if !result.Success() {
		logger.Error("订单校验失败：", result.Message)
		c.String(http.StatusOK, "fail")
		return
	}

	tradeNo := c.Request.Form.Get("trade_no")
	err = h.notify(result.OutTradeNo, tradeNo)
	if err != nil {
		logger.Error(err)
		c.String(http.StatusOK, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}

// GeekPayNotify 支付异步回调
func (h *PaymentHandler) GeekPayNotify(c *gin.Context) {
	var params = make(map[string]string)
	for k := range c.Request.URL.Query() {
		params[k] = c.Query(k)
	}

	logger.Infof("收到GeekPay订单支付回调：%+v", params)
	// 检查支付状态
	if params["trade_status"] != "TRADE_SUCCESS" {
		c.String(http.StatusOK, "success")
		return
	}

	sign := h.geekPayService.Sign(params)
	if sign != c.Query("sign") {
		logger.Errorf("签名验证失败, %s, %s", sign, c.Query("sign"))
		c.String(http.StatusOK, "fail")
		return
	}

	err := h.notify(params["out_trade_no"], params["trade_no"])
	if err != nil {
		logger.Error(err)
		c.String(http.StatusOK, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}

//
//// WechatPayNotify 微信商户支付异步回调
//func (h *PaymentHandler) WechatPayNotify(c *gin.Context) {
//	// 打印完整回调请求参数（便于本地调试）
//	bodyBytes, _ := io.ReadAll(c.Request.Body)
//	logger.Infof("WechatPayNotify RawBody: %s", string(bodyBytes))
//	// 由于读取了 Body，这里需要重新写回给后续逻辑使用
//	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
//
//	err := c.Request.ParseForm()
//	if err != nil {
//		c.String(http.StatusOK, "fail")
//		return
//	}
//
//	logger.Infof("WechatPayNotify Method: %s, URL: %s", c.Request.Method, c.Request.URL.String())
//	logger.Infof("WechatPayNotify Header: %+v", c.Request.Header)
//	logger.Infof("WechatPayNotify Query: %+v", c.Request.URL.Query())
//	logger.Infof("WechatPayNotify Form: %+v", c.Request.Form)
//
//	result := h.wechatPayService.TradeVerify(c.Request)
//	logger.Infof("收到微信商号订单支付回调：%+v", result)
//	if !result.Success() {
//		logger.Error("订单校验失败：", err)
//		c.JSON(http.StatusBadRequest, gin.H{
//			"code":    "FAIL",
//			"message": err.Error(),
//		})
//		return
//	}
//
//	err = h.notify(result.OutTradeNo, result.TradeId)
//	if err != nil {
//		logger.Error(err)
//		c.String(http.StatusOK, "fail")
//		return
//	}
//
//	c.String(http.StatusOK, "success")
//}

// WechatPayNotify 微信商户支付异步回调
func (h *PaymentHandler) WechatPayNotify(c *gin.Context) {
	// 打印完整回调请求参数（便于本地调试）
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	logger.Infof("WechatPayNotify RawBody: %s", string(bodyBytes))

	// 给我打印完整的排序后的字符串（key_sign 不参与）
	var notifyParams map[string]string
	if err := json.Unmarshal(bodyBytes, &notifyParams); err != nil {
		// 如果不是 JSON 格式，记录错误方便排查
		logger.Errorf("WechatPayNotify 解析回调 JSON 失败: %v", err)
	} else {
		// 1. 按字段名字典序排序（排除 key_sign）
		keys := make([]string, 0, len(notifyParams))
		for k := range notifyParams {
			if k == "key_sign" {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// 2. 拼接成 a=value1&attach=&b=value2... 的形式（空字符串也参与）
		var sb strings.Builder
		for i, k := range keys {
			if i > 0 {
				sb.WriteByte('&')
			}
			sb.WriteString(k)
			sb.WriteByte('=')
			sb.WriteString(notifyParams[k])
		}
		string1 := sb.String()
		logger.Infof("WechatPayNotify sign string1（已排序，未包含 key_sign）: %s", string1)

		//不做md5签名的校验

		// 3. 将对应的字符串转成 md5，并对比 md5 后的值和 key_sign 是否匹配
		//    注意：这里的 accessToken 需要你自行填写成文档中的 access_token 密钥
		//const accessToken = "48bfefa440cf4faa9a95012b2cc583a9" // TODO: 将这里替换为真实的 access_token
		//if accessToken != "" {
		//	string2 := string1 + "&access_token=" + accessToken
		//	sum := md5.Sum([]byte(string2))
		//	localSign := strings.ToLower(hex.EncodeToString(sum[:]))
		//	remoteSign := strings.ToLower(notifyParams["key_sign"])
		//
		//	logger.Infof("WechatPayNotify localSign: %s, remote key_sign: %s, match: %v",
		//		localSign, remoteSign, localSign == remoteSign)
		//}
	}

	// 由于读取了 Body，这里需要重新写回给后续逻辑使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	err := c.Request.ParseForm()
	if err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	logger.Infof("WechatPayNotify Method: %s, URL: %s", c.Request.Method, c.Request.URL.String())
	logger.Infof("WechatPayNotify Header: %+v", c.Request.Header)
	logger.Infof("WechatPayNotify Query: %+v", c.Request.URL.Query())
	logger.Infof("WechatPayNotify Form: %+v", c.Request.Form)

	// 先不校验签名，如果 return_msg == 支付成功 且 terminal_trace 在系统里面存在的话，
	// 根据订单号来将对应的订单更改为充值到账
	if notifyParams == nil {
		logger.Error("WechatPayNotify notifyParams 为空")
		c.JSON(http.StatusOK, gin.H{
			"return_code": "02",
			"return_msg":  "参数错误",
		})
		return
	}

	// 判断回调是否支付成功（按文档约定的字段和值来判断）
	if notifyParams["return_msg"] != "支付成功" {
		logger.Errorf("WechatPayNotify 支付未成功，return_msg=%s", notifyParams["return_msg"])
		c.JSON(http.StatusOK, gin.H{
			"return_code": "02",
			"return_msg":  "支付未成功",
		})
		return
	}

	orderNo := notifyParams["terminal_trace"] // 文档中说明为你方订单号
	if orderNo == "" {
		logger.Error("WechatPayNotify 缺少 terminal_trace（订单号）")
		c.JSON(http.StatusOK, gin.H{
			"return_code": "02",
			"return_msg":  "缺少订单号",
		})
		return
	}

	// tradeNo 可以根据文档选择合适的字段；如果没有就先用 terminal_trace 占位
	tradeNo := notifyParams["out_trade_no"]
	if tradeNo == "" {
		tradeNo = orderNo
	}

	if err := h.notify(orderNo, tradeNo); err != nil {
		logger.Errorf("WechatPayNotify 订单处理失败, orderNo=%s, tradeNo=%s, err=%v", orderNo, tradeNo, err)
		c.JSON(http.StatusOK, gin.H{
			"return_code": "02",
			"return_msg":  "订单处理失败",
		})
		return
	}

	

	//需要返回
	//{"return_code":"01","return_msg":"success"}
	c.JSON(http.StatusOK, gin.H{
		"return_code": "01",
		"return_msg":  "success",
	})
}
