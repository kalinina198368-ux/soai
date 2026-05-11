package handler

// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// * Copyright 2023 The Geek-AI Authors. All rights reserved.
// * Use of this source code is governed by a Apache-2.0 license
// * that can be found in the LICENSE file.
// * @Author yangjian102621@163.com
// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

import (
	"bytes"
	"encoding/json"
	"fmt"
	"geekai/core"
	"geekai/core/types"
	"geekai/service"
	"geekai/service/oss"
	"geekai/service/sora2"
	"geekai/store/model"
	"geekai/store/vo"
	"geekai/utils"
	"geekai/utils/resp"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Sora2Handler struct {
	BaseHandler
	sora2Service *sora2.Service
	uploader     *oss.UploaderManager
	userService  *service.UserService
}

func NewSora2Handler(app *core.AppServer, db *gorm.DB, service *sora2.Service, uploader *oss.UploaderManager, userService *service.UserService) *Sora2Handler {
	return &Sora2Handler{
		BaseHandler: BaseHandler{
			App: app,
			DB:  db,
		},
		sora2Service: service,
		uploader:     uploader,
		userService:  userService,
	}
}

// Generate 生成Sora2视频
func (h *Sora2Handler) Generate(c *gin.Context) {
	var data struct {
		Prompt         string   `json:"prompt" binding:"required"`
		Images         []string `json:"images"`       // 图生视频的图片数组
		Duration       string   `json:"duration"`     // 5, 10, 15, 30
		AspectRatio    string   `json:"aspect_ratio"` // 16:9, 9:16, 1:1
		Quality        string   `json:"quality"`      // standard, hd, uhd
		Style          string   `json:"style"`        // realistic, animated, artistic, cinematic
		NegativePrompt string   `json:"negative_prompt"`
		Seed           int64    `json:"seed"`
		Steps          int      `json:"steps"`
		Model          string   `json:"model"` // sora-2, sora-2-pro
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	user, err := h.GetLoginUser(c)
	if err != nil {
		resp.NotAuth(c)
		return
	}

	// 验证参数
	if data.Prompt == "" {
		resp.ERROR(c, "提示词不能为空")
		return
	}

	// 设置默认值
	if data.Duration == "" {
		data.Duration = "10"
	}
	if data.AspectRatio == "" {
		data.AspectRatio = "16:9"
	}
	if data.Quality == "" {
		data.Quality = "hd"
	}
	if data.Style == "" {
		data.Style = "realistic"
	}
	if data.Model == "" {
		data.Model = "sora-2"
	}
	if data.Steps == 0 {
		data.Steps = 30
	}

	// 计算消耗的算力
	power := h.calculatePower(data.Duration, data.Quality, data.Model)

	if user.Power < power {
		resp.ERROR(c, "您的算力不足，请充值后再试！")
		return
	}

	userId := int(h.GetLoginUserId(c))

	// 创建任务参数
	params := types.Sora2Params{
		Duration:       data.Duration,
		AspectRatio:    data.AspectRatio,
		Quality:        data.Quality,
		Style:          data.Style,
		NegativePrompt: data.NegativePrompt,
		Seed:           data.Seed,
		Steps:          data.Steps,
		Model:          data.Model,
	}

	task := types.Sora2Task{
		ClientId:         utils.RandString(10),
		UserId:           userId,
		Prompt:           data.Prompt,
		Images:           data.Images, // 添加图片参数
		Params:           params,
		TranslateModelId: h.App.SysConfig.TranslateModelId,
	}

	// 插入数据库
	job := model.Sora2Job{
		UserId:   userId,
		Prompt:   data.Prompt,
		Images:   utils.JsonEncode(data.Images), // 保存图片信息
		Power:    power,
		UpStatus: 0,
		TaskInfo: utils.JsonEncode(task),
		Status:   "pending",
	}

	tx := h.DB.Create(&job)
	if tx.Error != nil {
		resp.ERROR(c, tx.Error.Error())
		return
	}

	// 创建任务
	task.Id = job.Id
	h.sora2Service.PushTask(task)

	// 扣除用户算力
	err = h.userService.DecreasePower(job.UserId, job.Power, model.PowerLog{
		Type:   types.PowerConsume,
		Model:  "sora2",
		Remark: fmt.Sprintf("Sora2 视频生成，任务ID：%d", job.Id),
	})
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c, gin.H{
		"task_id": job.Id,
		"message": "任务已提交，正在生成中...",
	})
}

// 通过已有的视频来生成对应的视频角色
func (h *Sora2Handler) Characters(c *gin.Context) {
	var data struct {
		TaskId     string `json:"task_id" binding:"required"` // 前端传入的视频任务 ID（chatgpt_sora2_jobs.id）
		Timestamps string `json:"timestamps" binding:"required"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	user, err := h.GetLoginUser(c)
	if err != nil {
		resp.NotAuth(c)
		return
	}

	// 验证参数
	if data.TaskId == "" || data.Timestamps == "" {
		resp.ERROR(c, "参数异常")
		return
	}

	// 计算消耗的算力
	power := 10

	if user.Power < power {
		resp.ERROR(c, "您的算力不足，请充值后再试！")
		return
	}

	userId := int(h.GetLoginUserId(c))

	// 校验并获取对应的视频任务
	var job model.Sora2Job
	if err := h.DB.Where("task_id = ? AND user_id = ?", data.TaskId, userId).First(&job).Error; err != nil {
		resp.ERROR(c, "视频任务不存在或无权限访问")
		return
	}

	if job.TaskId == "" {
		resp.ERROR(c, "视频任务尚未提交到上游，无法提取角色")
		return
	}

	// 先在本地创建角色记录（pending 状态）
	role := model.Sora2Role{
		UserId: userId,
		// 记录本地 job 的 ID，方便后续关联
		TaskId: job.TaskId,
		Power:  power,
		Status: "pending",
	}
	if err := h.DB.Create(&role).Error; err != nil {
		resp.ERROR(c, "创建角色任务失败："+err.Error())
		return
	}

	// 调用上游角色提取接口
	reqBody := map[string]string{
		//"from_task":  fmt.Sprintf("video_%s", job.TaskId),
		"from_task":  data.TaskId,
		"timestamps": data.Timestamps,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	apiURL := "https://api.geekai.pro/sora/v1/characters"
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		h.DB.Model(&role).Updates(map[string]interface{}{
			"status":  "failed",
			"err_msg": err.Error(),
		})
		resp.ERROR(c, "创建上游请求失败："+err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-B2lsQ8mvegFgzyFJ932f66BcFe9542A28fA8437e1bA3C439")

	respApi, err := http.DefaultClient.Do(req)
	if err != nil {
		h.DB.Model(&role).Updates(map[string]interface{}{
			"status":  "failed",
			"err_msg": err.Error(),
		})
		resp.ERROR(c, "调用上游接口失败："+err.Error())
		return
	}
	defer respApi.Body.Close()

	respBody, _ := io.ReadAll(respApi.Body)

	if respApi.StatusCode != http.StatusOK && respApi.StatusCode != http.StatusCreated {
		errMsg := fmt.Sprintf("处理角色返回错误：%d, %s", respApi.StatusCode, string(respBody))
		h.DB.Model(&role).Updates(map[string]interface{}{
			"status":   "failed",
			"err_msg":  errMsg,
			"raw_data": string(respBody),
		})
		resp.ERROR(c, errMsg)
		return
	}

	// 解析上游返回数据
	var apiResp struct {
		Id                string `json:"id"`
		Username          string `json:"username"`
		DisplayName       string `json:"display_name"`
		Permalink         string `json:"permalink"`
		ProfilePictureURL string `json:"profile_picture_url"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		h.DB.Model(&role).Updates(map[string]interface{}{
			"status":   "failed",
			"err_msg":  "处理角色返回失败: " + err.Error(),
			"raw_data": string(respBody),
		})
		resp.ERROR(c, "处理角色返回失败："+err.Error())
		return
	}

	// 下载头像到 OSS（如果存在）
	var sysPictureUrl string
	if apiResp.ProfilePictureURL != "" {
		logger.Infof("try download profile picture: role_id=%d, profile_picture_url=%s", role.Id, apiResp.ProfilePictureURL)
		ossUrl, err := h.uploader.GetUploadHandler().PutUrlFile(apiResp.ProfilePictureURL, true)
		if err != nil {
			logger.Errorf("download profile picture with error: role_id=%d, error=%v", role.Id, err)
			// 下载失败不影响主流程，但记录错误
		} else {
			logger.Infof("download profile picture success: role_id=%d, sys_picture_url=%s", role.Id, ossUrl)
			sysPictureUrl = ossUrl
		}
	}

	// 更新角色记录为完成状态
	updateData := map[string]interface{}{
		"status":              "completed",
		"raw_data":            string(respBody),
		"username":            apiResp.Username,
		"display_name":        apiResp.DisplayName,
		"permalink":           apiResp.Permalink,
		"profile_picture_url": apiResp.ProfilePictureURL,
	}
	if sysPictureUrl != "" {
		updateData["sys_picture_url"] = sysPictureUrl
	}
	h.DB.Model(&role).Updates(updateData)

	// 扣除用户算力
	err = h.userService.DecreasePower(userId, power, model.PowerLog{
		Type:   types.PowerConsume,
		Model:  "sora2",
		Remark: fmt.Sprintf("Sora2 角色提取，视频任务ID：%d，角色任务ID：%d", job.Id, role.Id),
	})
	if err != nil {
		// 算力扣减失败不影响主流程，但记录错误
		h.DB.Model(&role).UpdateColumn("err_msg", fmt.Sprintf("算力扣减失败：%s", err.Error()))
	}

	resp.SUCCESS(c, gin.H{
		"id":                  apiResp.Id,
		"username":            apiResp.Username,
		"display_name":        apiResp.DisplayName,
		"permalink":           apiResp.Permalink,
		"profile_picture_url": apiResp.ProfilePictureURL,
		"message":             "角色提取成功",
	})
}

// Roles 返回当前用户的角色列表，可选按任务过滤
// GET /api/sora2/roles?task_id=xxx
func (h *Sora2Handler) Roles(c *gin.Context) {
	userId := h.GetLoginUserId(c)
	if userId == 0 {
		resp.NotAuth(c)
		return
	}

	taskId := c.Query("task_id")

	session := h.DB.
		Model(&model.Sora2Role{}).
		Where("user_id = ?", userId).
		Where("status = ?", "completed")

	if taskId != "" {
		session = session.Where("task_id = ?", taskId)
	}

	var list []model.Sora2Role
	if err := session.Order("id desc").Find(&list).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c, gin.H{
		"data": list,
	})
}

// List 获取视频列表
func (h *Sora2Handler) List(c *gin.Context) {
	userId := h.GetLoginUserId(c)
	page := h.GetInt(c, "page", 1)
	pageSize := h.GetInt(c, "page_size", 20)
	status := c.Query("status")

	session := h.DB.Session(&gorm.Session{}).Where("user_id", userId)
	if status != "" {
		session = session.Where("status", status)
	}

	// 统计总数
	var total int64
	session.Model(&model.Sora2Job{}).Count(&total)

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		session = session.Offset(offset).Limit(pageSize)
	}

	var list []model.Sora2Job
	err := session.Order("id desc").Find(&list).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c, gin.H{
		"total": total,
		"page":  page,
		"size":  pageSize,
		"data":  list,
	})
}

// History 获取历史记录
func (h *Sora2Handler) History(c *gin.Context) {
	userId := h.GetLoginUserId(c)
	page := h.GetInt(c, "page", 1)
	pageSize := h.GetInt(c, "page_size", 10)

	session := h.DB.Session(&gorm.Session{}).
		Where("user_id", userId).
		Where("status", "completed")

	// 统计总数
	var total int64
	session.Model(&model.Sora2Job{}).Count(&total)

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		session = session.Offset(offset).Limit(pageSize)
	}

	var list []model.Sora2Job
	err := session.Order("id desc").Find(&list).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c, gin.H{
		"total": total,
		"page":  page,
		"size":  pageSize,
		"data":  list,
	})
}

// Favorites 获取收藏列表
func (h *Sora2Handler) Favorites(c *gin.Context) {
	userId := h.GetLoginUserId(c)
	page := h.GetInt(c, "page", 1)
	pageSize := h.GetInt(c, "page_size", 10)

	session := h.DB.Session(&gorm.Session{}).
		Where("user_id", userId).
		Where("is_favorite", true).
		Where("status", "completed")

	// 统计总数
	var total int64
	session.Model(&model.Sora2Job{}).Count(&total)

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		session = session.Offset(offset).Limit(pageSize)
	}

	var list []model.Sora2Job
	err := session.Order("id desc").Find(&list).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c, gin.H{
		"total": total,
		"page":  page,
		"size":  pageSize,
		"data":  list,
	})
}

// ToggleFavorite 切换收藏状态
func (h *Sora2Handler) ToggleFavorite(c *gin.Context) {
	id := h.GetInt(c, "id", 0)
	userId := h.GetLoginUserId(c)

	var job model.Sora2Job
	err := h.DB.Where("id = ? AND user_id = ?", id, userId).First(&job).Error
	if err != nil {
		resp.ERROR(c, "视频不存在")
		return
	}

	// 切换收藏状态
	job.IsFavorite = !job.IsFavorite
	err = h.DB.Model(&job).UpdateColumn("is_favorite", job.IsFavorite).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	message := "已取消收藏"
	if job.IsFavorite {
		message = "已添加到收藏"
	}

	resp.SUCCESS(c, gin.H{
		"is_favorite": job.IsFavorite,
		"message":     message,
	})
}

// Download 下载视频
func (h *Sora2Handler) Download(c *gin.Context) {
	id := h.GetInt(c, "id", 0)
	userId := h.GetLoginUserId(c)

	var job model.Sora2Job
	err := h.DB.Where("id = ? AND user_id = ?", id, userId).First(&job).Error
	if err != nil {
		resp.ERROR(c, "视频不存在")
		return
	}

	if job.Status != "completed" {
		resp.ERROR(c, "视频尚未生成完成")
		return
	}

	if job.VideoURL == "" {
		resp.ERROR(c, "视频文件不存在")
		return
	}

	// 返回下载链接
	resp.SUCCESS(c, gin.H{
		"download_url": job.VideoURL,
		"filename":     fmt.Sprintf("sora2_video_%d.mp4", job.Id),
	})
}

// Remove 删除视频
func (h *Sora2Handler) Remove(c *gin.Context) {
	id := h.GetInt(c, "id", 0)
	userId := h.GetLoginUserId(c)

	var job model.Sora2Job
	err := h.DB.Where("id = ? AND user_id = ?", id, userId).First(&job).Error
	if err != nil {
		resp.ERROR(c, "视频不存在")
		return
	}

	// 只有失败或者超时的任务才能删除
	if job.Status == "processing" && time.Now().Before(job.CreatedAt.Add(time.Minute*30)) {
		resp.ERROR(c, "只有失败和超时(30分钟)的任务才能删除！")
		return
	}

	// 删除任务
	err = h.DB.Delete(&job).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	// 删除文件
	if job.VideoURL != "" {
		_ = h.uploader.GetUploadHandler().Delete(job.VideoURL)
	}
	if job.CoverURL != "" {
		_ = h.uploader.GetUploadHandler().Delete(job.CoverURL)
	}

	resp.SUCCESS(c)
}

// Publish 发布/取消发布视频
func (h *Sora2Handler) Publish(c *gin.Context) {
	id := h.GetInt(c, "id", 0)
	userId := h.GetLoginUserId(c)
	publish := h.GetBool(c, "publish")

	var job model.Sora2Job
	err := h.DB.Where("id = ? AND user_id = ?", id, userId).First(&job).Error
	if err != nil {
		resp.ERROR(c, "视频不存在")
		return
	}

	if job.Status != "completed" {
		resp.ERROR(c, "只有已完成的视频才能发布")
		return
	}

	err = h.DB.Model(&job).UpdateColumn("publish", publish).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	message := "已取消发布"
	if publish {
		message = "已发布到画廊"
	}

	resp.SUCCESS(c, gin.H{"message": message})
}

// calculatePower 计算消耗的算力
func (h *Sora2Handler) calculatePower(duration, quality, model string) int {

	basePower := h.App.SysConfig.LumaPower //使用luma配置好的算力模型
	//	basePower := 15 // 基础算力
	//basePower := 10 // 基础算力
	//
	//// 根据时长增加算力
	//durationInt, _ := strconv.Atoi(duration)
	//if durationInt > 10 {
	//	basePower += (durationInt - 10) * 2
	//}
	//
	//// 根据质量增加算力
	//switch quality {
	//case "hd":
	//	basePower += 8
	//case "uhd":
	//	basePower += 15
	//}
	//
	//// 根据模型增加算力
	//if model == "sora-2-pro" {
	//	basePower += 5
	//}

	return basePower
}

// GetProgress 查询任务进度
func (h *Sora2Handler) GetProgress(c *gin.Context) {
	taskId := c.Param("task_id")
	if taskId == "" {
		resp.ERROR(c, "任务ID不能为空")
		return
	}

	// 查询任务状态
	var job model.Sora2Job
	//err := h.DB.Where("task_id = ?", taskId).First(&job).Error
	err := h.DB.Where("id = ?", taskId).First(&job).Error
	if err != nil {
		resp.ERROR(c, "任务不存在")
		return
	}

	// 检查用户权限
	user, err := h.GetLoginUser(c)
	logger.Info(user)
	if err != nil {
		resp.ERROR(c, "请先登录")
		return
	}

	//if job.UserId != user.Id {
	//	resp.ERROR(c, "无权限访问此任务")
	//	return
	//}

	// 构建响应数据
	//progressData := gin.H{
	//	"task_id":    job.TaskId,
	//	"status":     h.mapStatusToEnum(job.Status),
	//	"progress":   h.mapStatusToProgress(job.Status),
	//	"video_url":  job.VideoURL,
	//	"cover_url":  job.CoverURL,
	//	"water_url":  job.WaterURL,
	//	"err_msg":    job.ErrMsg,
	//	"created_at": job.CreatedAt,
	//	"updated_at": job.UpdatedAt,
	//}

	progressData := gin.H{
		"task_id":    job.TaskId,
		"status":     h.mapStatusToEnum(job.Status),
		"progress":   h.mapStatusToProgress(job.Status),
		"video_url":  job.VideoURL,
		"cover_url":  job.CoverURL,
		"water_url":  job.WaterURL,
		"err_msg":    job.ErrMsg,
		"created_at": job.CreatedAt,
		"updated_at": job.UpdatedAt,
	}

	resp.SUCCESS(c, progressData)
}

// mapStatusToEnum 将数据库状态映射为API枚举状态
func (h *Sora2Handler) mapStatusToEnum(status string) string {
	switch status {
	case "pending":
		return "NOT_START"
	case "processing":
		return "IN_PROGRESS"
	case "completed":
		return "SUCCESS"
	case "failed":
		return "FAILURE"
	default:
		return "NOT_START"
	}
}

// mapStatusToProgress 将数据库状态映射为进度值
func (h *Sora2Handler) mapStatusToProgress(status string) string {
	switch status {
	case "pending":
		return "0"
	case "processing":
		return "50" // 处理中显示50%
	case "completed":
		return "100"
	case "failed":
		return "0"
	default:
		return "0"
	}
}

// GetMaterialsCategories 获取 Sora2 素材分类列表
func (h *Sora2Handler) GetMaterialsCategories(c *gin.Context) {
	var items []model.SoraMaterialCategory
	err := h.DB.Where("is_active = ?", true).Order("sort_order DESC, id DESC").Find(&items).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	var categories = make([]vo.SoraMaterialCategory, 0, len(items))
	for _, v := range items {
		var cat vo.SoraMaterialCategory
		if err = utils.CopyObject(v, &cat); err != nil {
			continue
		}
		cat.Id = v.Id
		cat.SortNum = v.SortOrder // 映射 SortOrder 到 SortNum
		categories = append(categories, cat)
	}

	resp.SUCCESS(c, categories)
}

// GetMaterialsList 获取 Sora2 素材列表
func (h *Sora2Handler) GetMaterialsList(c *gin.Context) {
	page := h.GetInt(c, "page", 1)
	pageSize := h.GetInt(c, "page_size", 20)
	categoryId := h.GetInt(c, "category_id", 0)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	session := h.DB.Model(&model.SoraMaterial{}).Where("is_active = ?", true)
	if categoryId > 0 {
		session = session.Where("category_id = ?", categoryId)
	}

	var total int64
	if err := session.Count(&total).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	offset := (page - 1) * pageSize
	var items []model.SoraMaterial
	if err := session.Order("sort_order DESC, id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	materials := make([]vo.SoraMaterial, 0, len(items))
	for _, v := range items {
		var m vo.SoraMaterial
		if err := utils.CopyObject(v, &m); err != nil {
			continue
		}
		m.Id = v.Id
		m.CategoryId = v.CategoryId
		materials = append(materials, m)
	}

	resp.SUCCESS(c, vo.NewPage(total, page, pageSize, materials))
}
