package handler

// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// * Copyright 2023 The Geek-AI Authors. All rights reserved.
// * Use of this source code is governed by a Apache-2.0 license
// * that can be found in the LICENSE file.
// * @Author yangjian102621@163.com
// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

import (
	"encoding/json"
	"fmt"
	"geekai/core"
	"geekai/core/types"
	"geekai/service"
	"geekai/service/mj"
	"geekai/service/oss"
	"geekai/store/model"
	"geekai/store/vo"
	"geekai/utils"
	"geekai/utils/resp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MidJourneyHandler struct {
	BaseHandler
	mjService   *mj.Service
	snowflake   *service.Snowflake
	uploader    *oss.UploaderManager
	userService *service.UserService
}

func NewMidJourneyHandler(app *core.AppServer, db *gorm.DB, snowflake *service.Snowflake, service *mj.Service, manager *oss.UploaderManager, userService *service.UserService) *MidJourneyHandler {
	return &MidJourneyHandler{
		snowflake:   snowflake,
		mjService:   service,
		uploader:    manager,
		userService: userService,
		BaseHandler: BaseHandler{
			App: app,
			DB:  db,
		},
	}
}

func (h *MidJourneyHandler) preCheck(c *gin.Context, requiredPower int) bool {
	user, err := h.GetLoginUser(c)
	if err != nil {
		resp.NotAuth(c)
		return false
	}

	if user.Power < requiredPower {
		resp.ERROR(c, "当前用户剩余算力不足以完成本次绘画！")
		return false
	}

	return true

}

// Image 创建一个绘画任务
func (h *MidJourneyHandler) Image(c *gin.Context) {
	var data struct {
		TaskType  string   `json:"task_type"`
		NanoMode  string   `json:"nano_mode"`
		ClientId  string   `json:"client_id"`
		Prompt    string   `json:"prompt"`
		NegPrompt string   `json:"neg_prompt"`
		Rate      string   `json:"rate"`
		Model     string   `json:"model"`   // 模型
		Chaos     int      `json:"chaos"`   // 创意度取值范围: 0-100
		Raw       bool     `json:"raw"`     // 是否开启原始模型
		Seed      int64    `json:"seed"`    // 随机数
		Stylize   int      `json:"stylize"` // 风格化
		ImgArr    []string `json:"img_arr"`
		Tile      bool     `json:"tile"`    // 重复平铺
		Quality   float32  `json:"quality"` // 画质
		Iw        float32  `json:"iw"`
		CRef      string   `json:"cref"` //生成角色一致的图像
		SRef      string   `json:"sref"` //生成风格一致的图像
		Cw        int      `json:"cw"`   // 参考程度
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	if data.TaskType != types.TaskImage.String() {
		resp.ERROR(c, "当前模型仅支持文生图和图生图")
		return
	}

	if data.NanoMode == "" {
		data.NanoMode = "txt2img"
	}
	//默认选择nano-banano模型
	if data.Model == "" {
		data.Model = "nano-banana"
	}

	// 根据模型选择不同的算力值进行预检查
	requiredPower := func() int {
		if data.Model == "nano-banana-2" {
			if h.App.SysConfig.MjPower2 > 0 {
				return h.App.SysConfig.MjPower2
			}
			return h.App.SysConfig.MjPower
		}
		return h.App.SysConfig.MjPower
	}()
	if !h.preCheck(c, requiredPower) {
		return
	}

	type nanoParams struct {
		AspectRatio string `json:"aspect_ratio,omitempty"`
		Mode        string `json:"mode,omitempty"`
		Model       string `json:"model,omitempty"`
	}
	opts := nanoParams{
		AspectRatio: data.Rate,
		Mode:        data.NanoMode,
		Model:       data.Model,
	}
	paramsBytes, err := json.Marshal(opts)
	if err != nil {
		resp.ERROR(c, "参数序列化失败："+err.Error())
		return
	}

	// 如果本地图片上传的是相对地址，处理成绝对地址
	for k, v := range data.ImgArr {
		if !strings.HasPrefix(v, "http") {
			data.ImgArr[k] = fmt.Sprintf("http://localhost:5678/%s", strings.TrimLeft(v, "/"))
		}
	}

	idValue, _ := c.Get(types.LoginUserID)
	userId := utils.IntValue(utils.InterfaceToString(idValue), 0)
	// generate task id
	taskId, err := h.snowflake.Next(true)
	if err != nil {
		resp.ERROR(c, "error with generate task id: "+err.Error())
		return
	}
	task := types.MjTask{
		ClientId:         data.ClientId,
		TaskId:           taskId,
		Type:             types.TaskType(data.TaskType),
		Prompt:           data.Prompt,
		NegPrompt:        data.NegPrompt,
		Params:           string(paramsBytes),
		UserId:           userId,
		ImgArr:           data.ImgArr,
		Mode:             h.App.SysConfig.MjMode,
		TranslateModelId: h.App.SysConfig.TranslateModelId,
	}
	job := model.MidJourneyJob{
		Type:     data.TaskType,
		UserId:   userId,
		TaskId:   taskId,
		TaskInfo: utils.JsonEncode(task),
		Progress: 0,
		Prompt:   data.Prompt,
		Power: func() int {
			if data.Model == "nano-banana-2" {
				if h.App.SysConfig.MjPower2 > 0 {
					return h.App.SysConfig.MjPower2
				}
				return h.App.SysConfig.MjPower
			}
			return h.App.SysConfig.MjPower
		}(),
		CreatedAt: time.Now(),
	}
	opt := "绘图"

	if res := h.DB.Create(&job); res.Error != nil || res.RowsAffected == 0 {
		resp.ERROR(c, "添加任务失败："+res.Error.Error())
		return
	}

	task.Id = job.Id
	h.mjService.PushTask(task)

	// update user's power
	err = h.userService.DecreasePower(job.UserId, job.Power, model.PowerLog{
		Type:   types.PowerConsume,
		Model:  data.Model,
		Remark: fmt.Sprintf("%s操作，任务ID：%s", opt, job.TaskId),
	})
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c)
}

type reqVo struct {
	Index       int    `json:"index"`
	ClientId    string `json:"client_id"`
	ChannelId   string `json:"channel_id"`
	MessageId   string `json:"message_id"`
	MessageHash string `json:"message_hash"`
}

// Upscale send upscale command to MidJourney Bot
func (h *MidJourneyHandler) Upscale(c *gin.Context) {
	var data reqVo
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	if !h.preCheck(c, h.App.SysConfig.MjActionPower) {
		return
	}

	idValue, _ := c.Get(types.LoginUserID)
	userId := utils.IntValue(utils.InterfaceToString(idValue), 0)
	taskId, _ := h.snowflake.Next(true)
	task := types.MjTask{
		ClientId:    data.ClientId,
		Type:        types.TaskUpscale,
		UserId:      userId,
		ChannelId:   data.ChannelId,
		Index:       data.Index,
		MessageId:   data.MessageId,
		MessageHash: data.MessageHash,
		Mode:        h.App.SysConfig.MjMode,
	}
	job := model.MidJourneyJob{
		Type:      types.TaskUpscale.String(),
		UserId:    userId,
		TaskId:    taskId,
		TaskInfo:  utils.JsonEncode(task),
		Progress:  0,
		Power:     h.App.SysConfig.MjActionPower,
		CreatedAt: time.Now(),
	}
	if res := h.DB.Create(&job); res.Error != nil || res.RowsAffected == 0 {
		resp.ERROR(c, "添加任务失败："+res.Error.Error())
		return
	}

	task.Id = job.Id
	h.mjService.PushTask(task)

	// update user's power
	err := h.userService.DecreasePower(job.UserId, job.Power, model.PowerLog{
		Type:   types.PowerConsume,
		Model:  "mid-journey",
		Remark: fmt.Sprintf("Upscale 操作，任务ID：%s", job.TaskId),
	})
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c)
}

// Variation send variation command to MidJourney Bot
func (h *MidJourneyHandler) Variation(c *gin.Context) {
	var data reqVo
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	if !h.preCheck(c, h.App.SysConfig.MjActionPower) {
		return
	}

	idValue, _ := c.Get(types.LoginUserID)
	userId := utils.IntValue(utils.InterfaceToString(idValue), 0)
	taskId, _ := h.snowflake.Next(true)
	task := types.MjTask{
		Type:        types.TaskVariation,
		ClientId:    data.ClientId,
		UserId:      userId,
		Index:       data.Index,
		ChannelId:   data.ChannelId,
		MessageId:   data.MessageId,
		MessageHash: data.MessageHash,
		Mode:        h.App.SysConfig.MjMode,
	}
	job := model.MidJourneyJob{
		Type:      types.TaskVariation.String(),
		ChannelId: data.ChannelId,
		UserId:    userId,
		TaskId:    taskId,
		TaskInfo:  utils.JsonEncode(task),
		Progress:  0,
		Power:     h.App.SysConfig.MjActionPower,
		CreatedAt: time.Now(),
	}
	if res := h.DB.Create(&job); res.Error != nil || res.RowsAffected == 0 {
		resp.ERROR(c, "添加任务失败："+res.Error.Error())
		return
	}

	task.Id = job.Id
	h.mjService.PushTask(task)

	err := h.userService.DecreasePower(job.UserId, job.Power, model.PowerLog{
		Type:   types.PowerConsume,
		Model:  "mid-journey",
		Remark: fmt.Sprintf("Variation 操作，任务ID：%s", job.TaskId),
	})
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c)
}

// ImgWall 照片墙
func (h *MidJourneyHandler) ImgWall(c *gin.Context) {
	page := h.GetInt(c, "page", 0)
	pageSize := h.GetInt(c, "page_size", 0)
	err, jobs := h.getData(true, 0, page, pageSize, true)
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c, jobs)
}

// JobList 获取 MJ 任务列表
func (h *MidJourneyHandler) JobList(c *gin.Context) {
	finish := h.GetBool(c, "finish")
	userId := h.GetLoginUserId(c)
	page := h.GetInt(c, "page", 0)
	pageSize := h.GetInt(c, "page_size", 0)
	publish := h.GetBool(c, "publish")

	err, jobs := h.getData(finish, userId, page, pageSize, publish)
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c, jobs)
}

// JobList 获取 MJ 任务列表
func (h *MidJourneyHandler) getData(finish bool, userId uint, page int, pageSize int, publish bool) (error, vo.Page) {
	session := h.DB.Session(&gorm.Session{})
	if finish {
		session = session.Where("progress >= ?", 100).Order("id DESC")
	} else {
		session = session.Where("progress < ?", 100).Order("id ASC")
	}
	if userId > 0 {
		session = session.Where("user_id = ? and publish = ?", userId, 0)
	}
	if publish {
		session = session.Where("publish = ?", publish)
	}
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		session = session.Offset(offset).Limit(pageSize)
	}

	// 统计总数
	var total int64
	session.Model(&model.MidJourneyJob{}).Count(&total)

	var items []model.MidJourneyJob
	res := session.Find(&items)
	if res.Error != nil {
		return res.Error, vo.Page{}
	}

	var jobs = make([]vo.MidJourneyJob, 0)
	for _, item := range items {
		var job vo.MidJourneyJob
		err := utils.CopyObject(item, &job)
		if err != nil {
			continue
		}
		jobs = append(jobs, job)
	}
	return nil, vo.NewPage(total, page, pageSize, jobs)
}

// Remove remove task image
func (h *MidJourneyHandler) Remove(c *gin.Context) {
	id := h.GetInt(c, "id", 0)
	userId := h.GetInt(c, "user_id", 0)
	var job model.MidJourneyJob
	if res := h.DB.Where("id = ? AND user_id = ?", id, userId).First(&job); res.Error != nil {
		resp.ERROR(c, "记录不存在")
		return
	}

	//.UpdateColumn("publish", action)
	//action :=true
	err := h.DB.Model(&model.MidJourneyJob{Id: uint(id), UserId: userId}).UpdateColumn("publish", 1).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}
	// remove job
	//err := h.DB.Delete(&job).Error
	//if err != nil {
	//	resp.ERROR(c, err.Error())
	//	return
	//}

	// remove image
	err = h.uploader.GetUploadHandler().Delete(job.ImgURL)
	if err != nil {
		logger.Error("remove image failed: ", err)
	}

	resp.SUCCESS(c)
}

// Publish 发布图片到画廊显示
func (h *MidJourneyHandler) Publish(c *gin.Context) {
	id := h.GetInt(c, "id", 0)
	userId := h.GetInt(c, "user_id", 0)
	action := h.GetBool(c, "action") // 发布动作，true => 发布，false => 取消分享
	err := h.DB.Model(&model.MidJourneyJob{Id: uint(id), UserId: userId}).UpdateColumn("publish", action).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c)
}

// GetMaterialsCategories 获取 MidJourney 素材分类列表
func (h *MidJourneyHandler) GetMaterialsCategories(c *gin.Context) {
	var items []model.MjMaterialCategory
	err := h.DB.Where("is_active = ?", true).Order("sort_order DESC, id DESC").Find(&items).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	var categories = make([]vo.MjMaterialCategory, 0, len(items))
	for _, v := range items {
		var cat vo.MjMaterialCategory
		if err = utils.CopyObject(v, &cat); err != nil {
			continue
		}
		cat.Id = v.Id
		categories = append(categories, cat)
	}

	resp.SUCCESS(c, categories)
}

// GetMaterialsList 获取 MidJourney 素材列表
func (h *MidJourneyHandler) GetMaterialsList(c *gin.Context) {
	page := h.GetInt(c, "page", 1)
	pageSize := h.GetInt(c, "page_size", 20)
	categoryId := h.GetInt(c, "category_id", 0)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	session := h.DB.Model(&model.MjMaterial{}).Where("is_active = ?", true)
	if categoryId > 0 {
		session = session.Where("category_id = ?", categoryId)
	}

	var total int64
	if err := session.Count(&total).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	offset := (page - 1) * pageSize
	var items []model.MjMaterial
	if err := session.Order("sort_order DESC, id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	materials := make([]vo.MjMaterial, 0, len(items))
	for _, v := range items {
		var m vo.MjMaterial
		if err := utils.CopyObject(v, &m); err != nil {
			continue
		}
		m.Id = v.Id
		m.CategoryId = v.CategoryId
		materials = append(materials, m)
	}

	resp.SUCCESS(c, vo.NewPage(total, page, pageSize, materials))
}

// Retry 重试失败的任务
func (h *MidJourneyHandler) Retry(c *gin.Context) {
	var data struct {
		Id       int    `json:"id"`
		UserId   int    `json:"user_id"`
		ClientId string `json:"client_id"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	if data.Id == 0 {
		resp.ERROR(c, "任务ID不能为空")
		return
	}

	idValue, _ := c.Get(types.LoginUserID)
	userId := utils.IntValue(utils.InterfaceToString(idValue), 0)
	if userId == 0 {
		resp.NotAuth(c)
		return
	}

	var job model.MidJourneyJob
	if res := h.DB.Where("id = ? AND user_id = ?", data.Id, userId).First(&job); res.Error != nil {
		resp.ERROR(c, "任务不存在")
		return
	}

	// 检查任务是否失败
	if job.Progress != service.FailTaskProgress {
		resp.ERROR(c, "只能重试失败的任务")
		return
	}

	// 解析任务信息
	var task types.MjTask
	if err := utils.JsonDecode(job.TaskInfo, &task); err != nil {
		resp.ERROR(c, "解析任务信息失败："+err.Error())
		return
	}

	// 检查算力
	requiredPower := job.Power
	if !h.preCheck(c, requiredPower) {
		return
	}

	// 生成新的任务ID
	taskId, err := h.snowflake.Next(true)
	if err != nil {
		resp.ERROR(c, "生成任务ID失败："+err.Error())
		return
	}

	// 创建新任务
	newJob := model.MidJourneyJob{
		Type:      job.Type,
		UserId:    userId,
		TaskId:    taskId,
		TaskInfo:  job.TaskInfo, // 使用原始任务信息
		Progress:  0,
		Prompt:    job.Prompt,
		Power:     job.Power,
		CreatedAt: time.Now(),
	}

	if res := h.DB.Create(&newJob); res.Error != nil || res.RowsAffected == 0 {
		resp.ERROR(c, "创建重试任务失败："+res.Error.Error())
		return
	}

	// 更新任务信息中的ID和TaskId
	task.Id = newJob.Id
	task.TaskId = taskId
	task.ClientId = data.ClientId
	task.UserId = userId

	// 推送任务到队列
	h.mjService.PushTask(task)

	// 扣减算力
	err = h.userService.DecreasePower(newJob.UserId, newJob.Power, model.PowerLog{
		Type: types.PowerConsume,
		Model: func() string {
			// 从任务参数中解析模型名称
			if task.Params != "" {
				type nanoParams struct {
					Model string `json:"model,omitempty"`
				}
				var params nanoParams
				if err := json.Unmarshal([]byte(task.Params), &params); err == nil && params.Model != "" {
					return params.Model
				}
			}
			return "nano-banana"
		}(),
		Remark: fmt.Sprintf("重试任务，原任务ID：%s，新任务ID：%s", job.TaskId, newJob.TaskId),
	})
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c)
}
