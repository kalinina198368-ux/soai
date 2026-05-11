package main

// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// * Copyright 2023 The Geek-AI Authors. All rights reserved.
// * Use of this source code is governed by a Apache-2.0 license
// * that can be found in the LICENSE file.
// * @Author yangjian102621@163.com
// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

import (
	"context"
	"embed"
	"geekai/core"
	"geekai/core/types"
	"geekai/handler"
	"geekai/handler/admin"
	logger2 "geekai/logger"
	"geekai/service"
	"geekai/service/dalle"
	"geekai/service/mj"
	"geekai/service/oss"
	"geekai/service/payment"
	"geekai/service/sd"
	"geekai/service/sms"
	"geekai/service/sora2"
	"geekai/service/suno"
	"geekai/service/video"
	"geekai/service/video_square"
	"geekai/store"
	"geekai/store/model"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

var logger = logger2.GetLogger()

//go:embed res
var xdbFS embed.FS

// AppLifecycle 应用程序生命周期
type AppLifecycle struct {
}

// OnStart 应用程序启动时执行
func (l *AppLifecycle) OnStart(context.Context) error {
	logger.Info("AppLifecycle OnStart")
	return nil
}

// OnStop 应用程序停止时执行
func (l *AppLifecycle) OnStop(context.Context) error {
	logger.Info("AppLifecycle OnStop")
	return nil
}

func NewAppLifeCycle() *AppLifecycle {
	return &AppLifecycle{}
}

func main() {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config.toml"
	}
	debug, _ := strconv.ParseBool(os.Getenv("APP_DEBUG"))
	logger.Info("Loading config file: ", configFile)
	if !debug {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic Error:", err)
			}
		}()
	}

	app := fx.New(
		// 初始化配置应用配置
		fx.Provide(func() *types.AppConfig {
			config, err := core.LoadConfig(configFile)
			if err != nil {
				log.Fatal(err)
			}
			config.Path = configFile
			if debug {
				_ = core.SaveConfig(config)
			}
			return config
		}),
		// 创建应用服务
		fx.Provide(core.NewServer),
		// 初始化
		fx.Invoke(func(s *core.AppServer, client *redis.Client) {
			s.Init(debug, client)
		}),

		// 初始化数据库
		fx.Provide(store.NewGormConfig),
		fx.Provide(store.NewMysql),
		fx.Provide(store.NewRedisClient),
		fx.Provide(store.NewLevelDB),

		// 自动创建新增表（上传码）
		fx.Invoke(func(db *gorm.DB) {
			_ = db.AutoMigrate(&model.UploadCode{}, &model.UploadCodeImage{})
		}),

		fx.Provide(func() embed.FS {
			return xdbFS
		}),

		// 创建 Ip2Region 查询对象
		fx.Provide(func() (*xdb.Searcher, error) {
			file, err := xdbFS.Open("res/ip2region.xdb")
			if err != nil {
				return nil, err
			}
			cBuff, err := io.ReadAll(file)
			if err != nil {
				return nil, err
			}

			return xdb.NewWithBuffer(cBuff)
		}),

		// 创建控制器
		fx.Provide(handler.NewChatRoleHandler),
		fx.Provide(handler.NewUserHandler),
		fx.Provide(handler.NewChatHandler),
		fx.Provide(handler.NewNetHandler),
		fx.Provide(handler.NewUploadCodeHandler),
		fx.Provide(handler.NewSmsHandler),
		fx.Provide(handler.NewRedeemHandler),
		fx.Provide(handler.NewCaptchaHandler),
		fx.Provide(handler.NewMidJourneyHandler),
		fx.Provide(handler.NewChatModelHandler),
		fx.Provide(handler.NewSdJobHandler),
		fx.Provide(handler.NewPaymentHandler),
		fx.Provide(handler.NewOrderHandler),
		fx.Provide(handler.NewProductHandler),
		fx.Provide(handler.NewConfigHandler),
		fx.Provide(handler.NewPowerLogHandler),
		fx.Provide(handler.NewPointsLogHandler),

		fx.Provide(admin.NewConfigHandler),
		fx.Provide(admin.NewAdminHandler),
		fx.Provide(admin.NewApiKeyHandler),
		fx.Provide(admin.NewUserHandler),
		fx.Provide(admin.NewChatAppHandler),
		fx.Provide(admin.NewRedeemHandler),
		fx.Provide(admin.NewDashboardHandler),
		fx.Provide(admin.NewChatModelHandler),
		fx.Provide(admin.NewProductHandler),
		fx.Provide(admin.NewOrderHandler),
		fx.Provide(admin.NewWithdrawHandler), //提现
		fx.Provide(admin.NewChatHandler),
		fx.Provide(admin.NewPowerLogHandler),
		fx.Provide(admin.NewPointsLogHandler),

		// 创建服务
		fx.Provide(sms.NewSendServiceManager),
		fx.Provide(func(config *types.AppConfig) *service.CaptchaService {
			return service.NewCaptchaService(config.ApiConfig)
		}),
		fx.Provide(oss.NewUploaderManager),
		fx.Provide(dalle.NewService),
		fx.Invoke(func(s *dalle.Service) {
			s.Run()
			s.CheckTaskNotify()
			s.DownloadImages()
			s.CheckTaskStatus()
		}),

		// 邮件服务
		fx.Provide(service.NewSmtpService),
		// License 服务
		fx.Provide(service.NewLicenseService),
		fx.Invoke(func(licenseService *service.LicenseService) {
			// licenseService.SyncLicense()
		}),

		// MidJourney service pool
		fx.Provide(mj.NewService),
		fx.Provide(mj.NewClient),
		fx.Invoke(func(s *mj.Service) {
			s.Run()
			s.SyncTaskProgress()
			s.CheckTaskNotify()
			s.DownloadImages()
		}),

		// Stable Diffusion 机器人
		fx.Provide(sd.NewService),
		fx.Invoke(func(s *sd.Service, config *types.AppConfig) {
			s.Run()
			s.CheckTaskStatus()
			s.CheckTaskNotify()
		}),

		fx.Provide(suno.NewService),
		fx.Invoke(func(s *suno.Service) {
			s.Run()
			s.SyncTaskProgress()
			s.CheckTaskNotify()
			s.DownloadFiles()
		}),
		fx.Provide(video.NewService),
		fx.Invoke(func(s *video.Service) {
			s.Run()
			s.SyncTaskProgress()
			s.CheckTaskNotify()
			s.DownloadFiles()
		}),
		fx.Provide(sora2.NewService),
		fx.Invoke(func(s *sora2.Service) {
			s.Run()
			s.SyncTaskProgress()
			s.CheckTaskNotify()
			s.DownloadFiles()
		}),
		fx.Provide(video_square.NewService),
		fx.Invoke(func(s *video_square.Service) {
			s.Run()
		}),
		fx.Provide(service.NewUserService),
		fx.Provide(payment.NewAlipayService),
		fx.Provide(payment.NewHuPiPay),
		fx.Provide(payment.NewJPayService),
		fx.Provide(payment.NewWechatService),
		fx.Provide(service.NewSnowflake),
		fx.Provide(service.NewXXLJobExecutor),
		fx.Invoke(func(exec *service.XXLJobExecutor, config *types.AppConfig) {
			if config.XXLConfig.Enabled {
				go func() {
					log.Fatal(exec.Run())
				}()
			}
		}),

		// 注册路由
		fx.Invoke(func(s *core.AppServer, h *handler.ChatRoleHandler) {
			group := s.Engine.Group("/api/app/")
			group.GET("list", h.List)
			group.GET("list/user", h.ListByUser)
			group.POST("update", h.UpdateRole)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.UserHandler) {
			group := s.Engine.Group("/api/user/")
			group.POST("register", h.Register)
			group.POST("login", h.Login)
			group.GET("logout", h.Logout)
			group.GET("session", h.Session)
			group.GET("profile", h.Profile)
			group.POST("profile/update", h.ProfileUpdate)
			group.POST("password", h.UpdatePass)
			group.POST("bind/mobile", h.BindMobile)
			group.POST("bind/email", h.BindEmail)
			//转账
			group.POST("transfer", h.Transfer)
			group.POST("resetPass", h.ResetPass)
			group.GET("clogin", h.CLogin)
			group.GET("clogin/callback", h.CLoginCallback)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.ChatHandler) {
			group := s.Engine.Group("/api/chat/")
			group.GET("list", h.List)
			group.GET("detail", h.Detail)
			group.POST("update", h.Update)
			group.GET("remove", h.Remove)
			group.GET("history", h.History)
			group.GET("clear", h.Clear)
			group.POST("tokens", h.Tokens)
			group.GET("stop", h.StopGenerate)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.NetHandler) {
			s.Engine.POST("/api/upload", h.Upload)
			s.Engine.POST("/api/upload/list", h.List)
			s.Engine.GET("/api/upload/remove", h.Remove)
			s.Engine.GET("/api/download", h.Download)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.UploadCodeHandler) {
			group := s.Engine.Group("/api/upload_code/")
			group.GET("code", h.Code)
			group.GET("images", h.Images)
			group.GET("remove", h.Remove)
			group.POST("upload", h.Upload) // 匿名扫码上传（已在 needLogin 放行）
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.SmsHandler) {
			group := s.Engine.Group("/api/sms/")
			group.POST("code", h.SendCode)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.CaptchaHandler) {
			group := s.Engine.Group("/api/captcha/")
			group.GET("get", h.Get)
			group.POST("check", h.Check)
			group.GET("slide/get", h.SlideGet)
			group.POST("slide/check", h.SlideCheck)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.RedeemHandler) {
			group := s.Engine.Group("/api/redeem/")
			group.POST("verify", h.Verify)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.MidJourneyHandler) {
			group := s.Engine.Group("/api/mj/")
			group.POST("image", h.Image)
			group.POST("upscale", h.Upscale)
			group.POST("variation", h.Variation)
			group.POST("retry", h.Retry)
			group.GET("jobs", h.JobList)
			group.GET("imgWall", h.ImgWall)
			group.GET("remove", h.Remove)
			group.GET("publish", h.Publish)
			//后端素材
			group.GET("materials/categories", h.GetMaterialsCategories)
			//后端素材列表
			group.GET("materials/list", h.GetMaterialsList)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.SdJobHandler) {
			group := s.Engine.Group("/api/sd")
			group.POST("image", h.Image)
			group.GET("jobs", h.JobList)
			group.GET("imgWall", h.ImgWall)
			group.GET("remove", h.Remove)
			group.GET("publish", h.Publish)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.ConfigHandler) {
			group := s.Engine.Group("/api/config/")
			group.GET("get", h.Get)
			group.GET("license", h.License)
		}),

		// 管理后台控制器
		fx.Invoke(func(s *core.AppServer, h *admin.ConfigHandler) {
			group := s.Engine.Group("/api/admin/config")
			group.POST("update", h.Update)
			group.GET("get", h.Get)
			group.POST("active", h.Active)
			group.GET("fixData", h.FixData)
			group.GET("license", h.GetLicense)
		}),
		fx.Invoke(func(s *core.AppServer, h *admin.ManagerHandler) {
			group := s.Engine.Group("/api/admin/")
			group.POST("login", h.Login)
			group.GET("logout", h.Logout)
			group.GET("session", h.Session)
			group.GET("list", h.List)
			group.POST("save", h.Save)
			group.POST("enable", h.Enable)
			group.GET("remove", h.Remove)
			group.POST("resetPass", h.ResetPass)
		}),
		fx.Invoke(func(s *core.AppServer, h *admin.ApiKeyHandler) {
			group := s.Engine.Group("/api/admin/apikey/")
			group.POST("save", h.Save)
			group.GET("list", h.List)
			group.POST("set", h.Set)
			group.GET("remove", h.Remove)
		}),
		fx.Invoke(func(s *core.AppServer, h *admin.UserHandler) {
			group := s.Engine.Group("/api/admin/user/")
			group.GET("list", h.List)
			group.POST("save", h.Save)
			group.GET("remove", h.Remove)
			group.GET("loginLog", h.LoginLog)
			group.POST("resetPass", h.ResetPass)
		}),
		fx.Invoke(func(s *core.AppServer, h *admin.ChatAppHandler) {
			group := s.Engine.Group("/api/admin/role/")
			group.GET("list", h.List)
			group.POST("save", h.Save)
			group.POST("sort", h.Sort)
			group.POST("set", h.Set)
			group.GET("remove", h.Remove)
		}),
		fx.Invoke(func(s *core.AppServer, h *admin.RedeemHandler) {
			group := s.Engine.Group("/api/admin/redeem/")
			group.GET("list", h.List)
			group.POST("create", h.Create)
			group.POST("set", h.Set)
			group.GET("remove", h.Remove)
			group.POST("export", h.Export)
		}),
		fx.Invoke(func(s *core.AppServer, h *admin.DashboardHandler) {
			group := s.Engine.Group("/api/admin/dashboard/")
			group.GET("stats", h.Stats)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.ChatModelHandler) {
			group := s.Engine.Group("/api/model/")
			group.GET("list", h.List)
		}),
		fx.Invoke(func(s *core.AppServer, h *admin.ChatModelHandler) {
			group := s.Engine.Group("/api/admin/model/")
			group.POST("save", h.Save)
			group.GET("list", h.List)
			group.POST("set", h.Set)
			group.POST("sort", h.Sort)
			group.GET("remove", h.Remove)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.PaymentHandler) {
			group := s.Engine.Group("/api/payment/")
			group.POST("doPay", h.Pay)
			group.GET("payWays", h.GetPayWays)
			group.POST("notify/alipay", h.AlipayNotify)
			group.GET("notify/geek", h.GeekPayNotify)
			group.POST("notify/wechat", h.WechatPayNotify)
			group.POST("notify/hupi", h.HuPiPayNotify)
		}),
		fx.Invoke(func(s *core.AppServer, h *admin.ProductHandler) {
			group := s.Engine.Group("/api/admin/product/")
			group.POST("save", h.Save)
			group.GET("list", h.List)
			group.POST("enable", h.Enable)
			group.POST("sort", h.Sort)
			group.GET("remove", h.Remove)
		}),
		fx.Invoke(func(s *core.AppServer, h *admin.OrderHandler) {
			group := s.Engine.Group("/api/admin/order/")
			group.POST("list", h.List)
			group.GET("remove", h.Remove)
			group.GET("clear", h.Clear)
		}),
		//提现
		fx.Invoke(func(s *core.AppServer, h *admin.WithdrawHandler) {
			group := s.Engine.Group("/api/admin/withdraw/")
			group.POST("list", h.List)
			group.GET("remove", h.Remove)
			group.POST("approve", h.Approve) //审核通过
			group.POST("reject", h.Reject)
			//group.GET("clear", h.Clear)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.OrderHandler) {
			group := s.Engine.Group("/api/order/")
			group.GET("list", h.List)
			//group.POST("mockPaid", h.MockPaid) //真实联诚收银台支付
			group.POST("mockPaid", h.PayPaid) //真实联诚收银台支付
			group.GET("query", h.Query)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.ProductHandler) {
			group := s.Engine.Group("/api/product/")
			group.GET("list", h.List)
		}),

		fx.Provide(handler.NewInviteHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.InviteHandler) {
			group := s.Engine.Group("/api/invite/")
			group.GET("code", h.Code)
			group.GET("list", h.List)
			group.GET("hits", h.Hits)
			group.GET("shareClick", h.ShareClick)
			group.GET("shareClickLogList", h.ShareClickLogList)
		}),

		//团队
		fx.Provide(handler.NewTeamHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.TeamHandler) {
			group := s.Engine.Group("/api/team/")
			//group.GET("code", h.Code)
			group.GET("stats", h.Stats)
			group.GET("direct", h.Direct)
			group.GET("indirect", h.Indirect)
			//group.GET("list", h.List)
			//group.GET("hits", h.Hits)
		}),

		//提现
		fx.Provide(handler.NewWithdrawHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.WithdrawHandler) {
			group := s.Engine.Group("/api/withdraw/")
			group.GET("history", h.History)
			group.POST("apply", h.Apply)
			//group.GET("indirect", h.Indirect)
			//	http://localhost:5678/api/withdraw/history

		}),

		fx.Provide(admin.NewFunctionHandler),
		fx.Invoke(func(s *core.AppServer, h *admin.FunctionHandler) {
			group := s.Engine.Group("/api/admin/function/")
			group.POST("save", h.Save)
			group.POST("set", h.Set)
			group.GET("list", h.List)
			group.GET("remove", h.Remove)
			group.GET("token", h.GenToken)
		}),

		fx.Provide(admin.NewUploadHandler),
		fx.Invoke(func(s *core.AppServer, h *admin.UploadHandler) {
			s.Engine.POST("/api/admin/upload", h.Upload)
		}),

		fx.Provide(handler.NewFunctionHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.FunctionHandler) {
			group := s.Engine.Group("/api/function/")
			group.POST("weibo", h.WeiBo)
			group.POST("zaobao", h.ZaoBao)
			group.POST("dalle3", h.Dall3)
			group.GET("list", h.List)
		}),
		fx.Invoke(func(s *core.AppServer, h *admin.ChatHandler) {
			group := s.Engine.Group("/api/admin/chat/")
			group.POST("list", h.List)
			group.POST("message", h.Messages)
			group.GET("history", h.History)
			group.GET("remove", h.RemoveChat)
			group.GET("message/remove", h.RemoveMessage)
		}),
		fx.Invoke(func(s *core.AppServer, h *handler.PowerLogHandler) {
			group := s.Engine.Group("/api/powerLog/")
			group.POST("list", h.List)
		}),
		fx.Invoke(func(s *core.AppServer, h *admin.PowerLogHandler) {
			group := s.Engine.Group("/api/admin/powerLog/")
			group.POST("list", h.List)
		}),

		//积分日志
		fx.Invoke(func(s *core.AppServer, h *handler.PointsLogHandler) {
			group := s.Engine.Group("/api/pointsLog/")
			group.POST("list", h.List)
		}),
		fx.Invoke(func(s *core.AppServer, h *admin.PointsLogHandler) {
			group := s.Engine.Group("/api/admin/pointsLog/")
			group.POST("list", h.List)
		}),

		fx.Provide(admin.NewMenuHandler),
		fx.Invoke(func(s *core.AppServer, h *admin.MenuHandler) {
			group := s.Engine.Group("/api/admin/menu/")
			group.POST("save", h.Save)
			group.GET("list", h.List)
			group.POST("enable", h.Enable)
			group.POST("sort", h.Sort)
			group.GET("remove", h.Remove)
		}),
		fx.Provide(handler.NewMenuHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.MenuHandler) {
			group := s.Engine.Group("/api/menu/")
			group.GET("list", h.List)
		}),
		fx.Provide(handler.NewMarkMapHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.MarkMapHandler) {
			s.Engine.POST("/api/markMap/gen", h.Generate)
		}),
		fx.Provide(handler.NewDallJobHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.DallJobHandler) {
			group := s.Engine.Group("/api/dall")
			group.POST("image", h.Image)
			group.GET("jobs", h.JobList)
			group.GET("imgWall", h.ImgWall)
			group.GET("remove", h.Remove)
			group.GET("publish", h.Publish)
			group.GET("models", h.GetModels)
		}),
		fx.Provide(handler.NewSunoHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.SunoHandler) {
			group := s.Engine.Group("/api/suno")
			group.POST("create", h.Create)
			group.GET("list", h.List)
			group.GET("remove", h.Remove)
			group.GET("publish", h.Publish)
			group.POST("update", h.Update)
			group.GET("detail", h.Detail)
			group.GET("play", h.Play)
		}),
		fx.Provide(handler.NewVideoSquareHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.VideoSquareHandler) {
			// 用户端路由
			group := s.Engine.Group("/api/video-square/")
			group.GET("list", h.List)
			group.POST("view", h.IncrementView)
			// 管理员端路由
			adminGroup := s.Engine.Group("/api/admin/video-square/")
			adminGroup.POST("import", h.ImportData) // 手动导入JSON数据
		}),
		fx.Provide(handler.NewSora2Handler),
		fx.Invoke(func(s *core.AppServer, h *handler.Sora2Handler) {
			//http://localhost:5678/v2/videos/generate
			group := s.Engine.Group("/api/sora2")
			group.POST("generate", h.Generate)     //生成视频
			group.POST("characters", h.Characters) //通过taskid来生成视频角色
			group.GET("roles", h.Roles)            //获取当前用户的角色列表（可按task过滤）

			group.GET("list", h.List)
			group.GET("history", h.History)
			group.GET("favorites", h.Favorites)
			group.POST("favorite", h.ToggleFavorite)
			group.GET("download", h.Download)
			group.GET("remove", h.Remove)
			group.POST("publish", h.Publish)
			group.GET("progress/:task_id", h.GetProgress)

			// 素材相关接口
			group.GET("materials/categories", h.GetMaterialsCategories) //获取素材分类列表
			group.GET("materials/list", h.GetMaterialsList)             //获取素材列表
		}),
		fx.Provide(admin.NewChatAppTypeHandler),
		fx.Invoke(func(s *core.AppServer, h *admin.ChatAppTypeHandler) {
			group := s.Engine.Group("/api/admin/app/type")
			group.POST("save", h.Save)
			group.GET("list", h.List)
			group.GET("remove", h.Remove)
			group.POST("enable", h.Enable)
			group.POST("sort", h.Sort)
		}),
		fx.Provide(handler.NewChatAppTypeHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.ChatAppTypeHandler) {
			group := s.Engine.Group("/api/app/type")
			group.GET("list", h.List)
		}),
		fx.Provide(handler.NewTestHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.TestHandler) {
			group := s.Engine.Group("/api/test")
			group.Any("sse", h.PostTest, h.SseTest)
		}),
		fx.Provide(service.NewWebsocketService),
		fx.Provide(handler.NewWebsocketHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.WebsocketHandler) {
			s.Engine.Any("/api/ws", h.Client)
		}),
		fx.Provide(handler.NewPromptHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.PromptHandler) {
			group := s.Engine.Group("/api/prompt")
			group.POST("/lyric", h.Lyric)
			group.POST("/image", h.Image)
			group.POST("/video", h.Video)
			group.POST("/meta", h.MetaPrompt)
		}),
		fx.Invoke(func(s *core.AppServer, db *gorm.DB) {
			go func() {
				err := s.Run(db)
				if err != nil {
					logger.Error(err)
					os.Exit(0)
				}
			}()
		}),
		fx.Provide(NewAppLifeCycle),
		// 注册生命周期回调函数
		fx.Invoke(func(lifecycle fx.Lifecycle, lc *AppLifecycle) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return lc.OnStart(ctx)
				},
				OnStop: func(ctx context.Context) error {
					return lc.OnStop(ctx)
				},
			})
		}),
		fx.Provide(admin.NewImageHandler),
		fx.Invoke(func(s *core.AppServer, h *admin.ImageHandler) {
			group := s.Engine.Group("/api/admin/image")
			group.POST("/list/mj", h.MjList)
			group.POST("/list/sd", h.SdList)
			group.POST("/list/dall", h.DallList)
			group.GET("/remove", h.Remove)
		}),
		fx.Provide(admin.NewMjMaterialHandler),
		fx.Invoke(func(s *core.AppServer, h *admin.MjMaterialHandler) {
			group := s.Engine.Group("/api/admin/mj/material/")
			group.POST("category/list", h.CategoryList)
			group.POST("category/save", h.CategorySave)
			group.POST("category/enable", h.CategoryEnable)
			group.POST("category/sort", h.CategorySort)
			group.GET("category/remove", h.CategoryRemove)

			group.POST("list", h.MaterialList)
			group.POST("save", h.MaterialSave)
			group.POST("enable", h.MaterialEnable)
			group.POST("sort", h.MaterialSort)
			group.GET("remove", h.MaterialRemove)
		}),
		fx.Provide(admin.NewSoraMaterialHandler),
		fx.Invoke(func(s *core.AppServer, h *admin.SoraMaterialHandler) {
			group := s.Engine.Group("/api/admin/sora/material/")
			group.POST("category/list", h.CategoryList)
			group.POST("category/save", h.CategorySave)
			group.POST("category/enable", h.CategoryEnable)
			group.POST("category/sort", h.CategorySort)
			group.GET("category/remove", h.CategoryRemove)

			group.POST("list", h.MaterialList)
			group.POST("save", h.MaterialSave)
			group.POST("enable", h.MaterialEnable)
			group.POST("sort", h.MaterialSort)
			group.GET("remove", h.MaterialRemove)
		}),
		fx.Provide(admin.NewMediaHandler),
		fx.Invoke(func(s *core.AppServer, h *admin.MediaHandler) {
			group := s.Engine.Group("/api/admin/media")
			group.POST("/list/suno", h.SunoList)
			group.POST("/list/luma", h.SoraList)
			//	group.POST("/list/sora", h.SoraList)
			group.GET("/remove", h.Remove)
		}),
		fx.Provide(handler.NewRealtimeHandler),
		fx.Invoke(func(s *core.AppServer, h *handler.RealtimeHandler) {
			s.Engine.Any("/api/realtime", h.Connection)
			s.Engine.POST("/api/realtime/voice", h.VoiceChat)
		}),
	)
	// 启动应用程序
	go func() {
		if err := app.Start(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	// 监听退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 关闭应用程序
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Stop(ctx); err != nil {
		log.Fatal(err)
	}

}
