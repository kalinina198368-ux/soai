package sora2

// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// * Copyright 2023 The Geek-AI Authors. All rights reserved.
// * Use of this source code is governed by a Apache-2.0 license
// * that can be found in the LICENSE file.
// * @Author yangjian102621@163.com
// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

import (
	"encoding/json"
	"errors"
	"fmt"
	"geekai/core/types"
	logger2 "geekai/logger"
	"geekai/service"
	"geekai/service/oss"
	"geekai/store"
	"geekai/store/model"
	"geekai/utils"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"github.com/imroc/req/v3"
	"gorm.io/gorm"
)

var logger = logger2.GetLogger()

type Service struct {
	httpClient    *req.Client
	db            *gorm.DB
	uploadManager *oss.UploaderManager
	taskQueue     *store.RedisQueue
	notifyQueue   *store.RedisQueue
	wsService     *service.WebsocketService
	clientIds     map[uint]string
	userService   *service.UserService
}

func NewService(db *gorm.DB, manager *oss.UploaderManager, redisCli *redis.Client, wsService *service.WebsocketService, userService *service.UserService) *Service {
	return &Service{
		httpClient:    req.C().SetTimeout(time.Minute).SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"),
		db:            db,
		taskQueue:     store.NewRedisQueue("Sora2_Task_Queue", redisCli),
		notifyQueue:   store.NewRedisQueue("Sora2_Notify_Queue", redisCli),
		wsService:     wsService,
		uploadManager: manager,
		clientIds:     map[uint]string{},
		userService:   userService,
	}
}

func (s *Service) PushTask(task types.Sora2Task) {
	logger.Infof("add a new Sora2 task to the task list: %+v", task)
	s.taskQueue.RPush(task)
}

func (s *Service) Run() {

	logger.Info("进入到sora2调试页面")
	// 将数据库中未提交的任务加载到队列
	var jobs []model.Sora2Job
	s.db.Where("status", "pending").Find(&jobs)
	logger.Info(jobs)
	for _, v := range jobs {
		var task types.Sora2Task
		err := utils.JsonDecode(v.TaskInfo, &task)
		if err != nil {
			logger.Errorf("decode task info with error: %v", err)
			continue
		}
		task.Id = v.Id
		s.PushTask(task)
		s.clientIds[v.Id] = task.ClientId
	}
	logger.Info("Starting Sora2 job consumer...")
	go func() {
		for {
			var task types.Sora2Task
			err := s.taskQueue.LPop(&task)
			if err != nil {
				logger.Errorf("taking task with error: %v", err)
				continue
			}

			// Sora2 原生支持中文，无需翻译
			// Sora2 can understand Chinese prompts natively, no translation needed
			// This avoids translation loss and improves user experience
			//if utils.HasChinese(task.Prompt) {
			//	content, err := utils.OpenAIRequest(s.db, fmt.Sprintf(service.TranslatePromptTemplate, task.Prompt), task.TranslateModelId)
			//	if err == nil {
			//		task.Prompt = content
			//	} else {
			//		logger.Warnf("error with translate prompt: %v", err)
			//	}
			//}

			if task.ClientId != "" {
				s.clientIds[task.Id] = task.ClientId
			}

			var r Sora2RespVo
			r, err = s.Sora2Create(task)
			if err != nil {
				logger.Errorf("create Sora2 task with error: %v", err)
				err = s.db.Model(&model.Sora2Job{Id: task.Id}).UpdateColumns(map[string]interface{}{
					"err_msg":   err.Error(),
					"status":    "failed",
					"cover_url": "/images/failed.jpg",
				}).Error
				if err != nil {
					logger.Errorf("update task with error: %v", err)
				}
				s.notifyQueue.RPush(service.NotifyMessage{ClientId: task.ClientId, UserId: task.UserId, JobId: int(task.Id), Message: service.TaskStatusFailed})
				continue
			}

			// 更新任务信息
			err = s.db.Model(&model.Sora2Job{Id: task.Id}).UpdateColumns(map[string]interface{}{
				"task_id":    r.TaskId,
				"channel":    r.Channel,
				"prompt_ext": task.Prompt,
				"status":     "processing", // 任务提交成功后，状态设为 processing
			}).Error
			if err != nil {
				logger.Errorf("update Sora2 task with error: %v", err)
				s.PushTask(task)
			} else {
				logger.Infof("Sora2 task submitted successfully: task_id=%s, job_id=%d", r.TaskId, task.Id)
			}
		}
	}()
}

type Sora2RespVo struct {
	TaskId     string `json:"task_id"`
	Platform   string `json:"platform"`
	Action     string `json:"action"`
	Status     string `json:"status"`
	FailReason string `json:"fail_reason"`
	SubmitTime int64  `json:"submit_time"`
	StartTime  int64  `json:"start_time"`
	FinishTime int64  `json:"finish_time"`
	Progress   string `json:"progress"`
	Data       struct {
		Output string `json:"output"`
	} `json:"data"`
	SearchItem string `json:"search_item"`
	Channel    string `json:"channel,omitempty"`
}

func (s *Service) Sora2Create(task types.Sora2Task) (Sora2RespVo, error) {
	// 读取 API KEY
	var apiKey model.ApiKey
	session := s.db.Session(&gorm.Session{}).Where("type", "luma").Where("enabled", true)
	if task.Channel != "" {
		session = session.Where("api_url", task.Channel)
	}
	tx := session.Order("last_used_at DESC").First(&apiKey)
	if tx.Error != nil {
		return Sora2RespVo{}, errors.New("no available API KEY for Sora2")
	}

	// 构建请求参数
	reqBody := map[string]interface{}{
		"prompt":       task.Prompt,
		"model":        task.Params.Model,
		"aspect_ratio": task.Params.AspectRatio,
		"duration":     task.Params.Duration,
	}

	// 添加图片参数（图生视频模式）
	if len(task.Images) > 0 {
		reqBody["images"] = task.Images
		logger.Infof("图生视频模式，图片数量: %d", len(task.Images))
	}

	// 添加可选参数
	if task.Params.Quality == "hd" {
		reqBody["hd"] = true
	}
	if task.Params.NegativePrompt != "" {
		reqBody["negative_prompt"] = task.Params.NegativePrompt
	}
	if task.Params.Seed > 0 {
		reqBody["seed"] = task.Params.Seed
	}
	if task.Params.Steps > 0 {
		reqBody["steps"] = task.Params.Steps
	}

	var res Sora2RespVo
	apiURL := fmt.Sprintf("%s/v2/videos/generations", apiKey.ApiURL)
	logger.Infof("Sora2 API URL: %s, request body: %+v", apiURL, reqBody)

	startTime := time.Now()
	r, err := s.httpClient.R().
		SetHeader("Authorization", "Bearer "+apiKey.Value).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post(apiURL)

	elapsed := time.Since(startTime)
	logger.Infof("Sora2 API request completed in %v", elapsed)

	//{
	//    "task_id": "f0aa213c-c09e-4e19-a0e5-c698fe48acf1"
	//}

	if err != nil {
		return Sora2RespVo{}, fmt.Errorf("请求 API 出错：%v", err)
	}

	if r.StatusCode != 200 && r.StatusCode != 201 {
		return Sora2RespVo{}, fmt.Errorf("请求 API 出错：%d, %s", r.StatusCode, r.String())
	}

	body, _ := io.ReadAll(r.Body)
	err = json.Unmarshal(body, &res)
	if err != nil {
		return Sora2RespVo{}, fmt.Errorf("解析API数据失败：%v, %s", err, string(body))
	}

	// update the last_use_at for api key
	apiKey.LastUsedAt = time.Now().Unix()
	session.Updates(&apiKey)
	res.Channel = apiKey.ApiURL
	return res, nil
}

func (s *Service) CheckTaskNotify() {
	go func() {
		logger.Info("Running Sora2 task notify checking ...")
		for {
			var message service.NotifyMessage
			err := s.notifyQueue.LPop(&message)
			if err != nil {
				continue
			}
			logger.Debugf("Receive notify message: %+v", message)
			client := s.wsService.Clients.Get(message.ClientId)
			if client == nil {
				continue
			}
			utils.SendChannelMsg(client, types.ChSora2, message.Message)
		}
	}()
}

func (s *Service) DownloadFiles() {
	go func() {
		logger.Info("Starting Sora2 download files service...")
		var items []model.Sora2Job
		for {
			// 查询已完成且video_url不为空的任务 且时间是在30分钟内 且up_status为0
			res := s.db.Where("status = ?", "completed").Where("video_url != '' AND video_url IS NOT NULL").Where("created_at > ?", time.Now().Add(-30*time.Minute)).Where("up_status = ?", 0).Find(&items)

			if res.Error != nil {
				logger.Errorf("query completed tasks with error: %v", res.Error)
				time.Sleep(time.Second * 10)
				continue
			}

			if len(items) > 0 {
				logger.Infof("Found %d completed tasks without video_url, starting download...", len(items))
			}

			for _, v := range items {
				if v.WaterURL == "" {
					logger.Warnf("Task %d has no water_url, skipping", v.Id)
					continue
				}

				logger.Infof("try download video: task_id=%d, water_url=%s", v.Id, v.WaterURL)
				videoURL, err := s.uploadManager.GetUploadHandler().PutUrlFile(v.WaterURL, true)
				if err != nil {
					logger.Errorf("download video with error: task_id=%d, error=%v", v.Id, err)
					continue
				}
				logger.Infof("download video success: task_id=%d, video_url=%s", v.Id, videoURL)
				v.VideoURL = videoURL

				if v.CoverURL == "" && v.ThumbnailURL != "" {
					logger.Infof("try download cover: task_id=%d, thumbnail_url=%s", v.Id, v.ThumbnailURL)
					coverURL, err := s.uploadManager.GetUploadHandler().PutUrlFile(v.ThumbnailURL, true)
					if err != nil {
						logger.Errorf("download cover with error: task_id=%d, error=%v", v.Id, err)
					} else {
						logger.Infof("download cover success: task_id=%d, cover_url=%s", v.Id, coverURL)
						v.CoverURL = coverURL
					}
				}

				err = s.db.Model(&model.Sora2Job{Id: v.Id}).UpdateColumns(map[string]interface{}{
					"video_url": v.VideoURL,
					"cover_url": v.CoverURL,
				}).Error
				if err != nil {
					logger.Errorf("update task video_url failed: task_id=%d, error=%v", v.Id, err)
					continue
				}
				logger.Infof("update task video_url success: task_id=%d", v.Id)
				//更新对应的数据库up_status为1
				err = s.db.Model(&model.Sora2Job{Id: v.Id}).UpdateColumns(map[string]interface{}{
					"up_status": 1,
				}).Error
				if err != nil {
					logger.Errorf("update task up_status failed: task_id=%d, error=%v", v.Id, err)
					continue
				}
				logger.Infof("update task up_status success: task_id=%d", v.Id)


				




				s.notifyQueue.RPush(service.NotifyMessage{ClientId: s.clientIds[v.Id], UserId: v.UserId, JobId: int(v.Id), Message: service.TaskStatusFinished})
			}

			time.Sleep(time.Second * 10)
		}
	}()
}

// SyncTaskProgress 异步拉取任务
func (s *Service) SyncTaskProgress() {
	go func() {
		var jobs []model.Sora2Job
		for {
			res := s.db.Where("status", "processing").Where("task_id <> ?", "").Find(&jobs)
			if res.Error != nil {
				continue
			}

			for _, job := range jobs {
				task, err := s.QuerySora2Task(job.TaskId, job.Channel)
				if err != nil {
					logger.Errorf("query task with error: %v", err)
					// 更新任务信息
					s.db.Model(&model.Sora2Job{Id: job.Id}).UpdateColumns(map[string]interface{}{
						"status":  "failed",
						"err_msg": err.Error(),
					})
					continue
				}

				logger.Debugf("task: %+v", task)

				// 根据新的API响应格式处理任务状态
				dbStatus := s.mapApiStatusToDbStatus(task.Status)
				progress := s.extractProgressFromApi(task.Progress)

				if task.Status == "SUCCESS" { // 任务完成
					data := map[string]interface{}{
						"status":    dbStatus,
						"video_url": task.Data.Output,
						"water_url": task.Data.Output, // 使用同一个URL
						"raw_data":  utils.JsonEncode(task),
						"err_msg":   task.FailReason,
					}
					err = s.db.Model(&model.Sora2Job{Id: job.Id}).UpdateColumns(data).Error
					if err != nil {
						logger.Errorf("更新数据库失败：%v", err)
						continue
					}

					// 推送完成状态到WebSocket
					s.pushProgressUpdate(job.Id, "100", task.Data.Output, "", task.Data.Output, "")
				} else if task.Status == "FAILURE" { // 任务失败
					s.db.Model(&model.Sora2Job{Id: job.Id}).UpdateColumns(map[string]interface{}{
						"status":  dbStatus,
						"err_msg": task.FailReason,
					})

					// 推送失败状态到WebSocket
					s.pushProgressUpdate(job.Id, "failed", "", "", "", task.FailReason)
				} else if task.Status == "IN_PROGRESS" { // 任务进行中
					// 更新进度和raw_data
					data := map[string]interface{}{
						"status":   dbStatus,
						"raw_data": utils.JsonEncode(task), // 保存完整的API响应，包含进度信息
					}
					err = s.db.Model(&model.Sora2Job{Id: job.Id}).UpdateColumns(data).Error
					if err != nil {
						logger.Errorf("更新进度信息失败：%v", err)
						continue
					}

					logger.Infof("Sora2任务进度更新: task_id=%s, progress=%s, job_id=%d", job.TaskId, progress, job.Id)

					// 推送进度更新到WebSocket
					s.pushProgressUpdate(job.Id, progress, "", "", "", "")
				}
			}

			// 找出失败的任务，并恢复其扣减算力
			s.db.Where("status", "failed").Where("power > ?", 0).Find(&jobs)
			for _, job := range jobs {
				err := s.userService.IncreasePower(job.UserId, job.Power, model.PowerLog{
					Type:   types.PowerRefund,
					Model:  "sora2",
					Remark: fmt.Sprintf("Sora2 任务失败，退回算力。任务ID：%s，Err:%s", job.TaskId, job.ErrMsg),
				})
				if err != nil {
					continue
				}
				// 更新任务状态
				s.db.Model(&job).UpdateColumn("power", 0)
			}
			time.Sleep(time.Second * 10)
		}
	}()
}

type Sora2TaskVo struct {
	TaskId     string `json:"task_id"`
	Platform   string `json:"platform"`
	Action     string `json:"action"`
	Status     string `json:"status"`
	FailReason string `json:"fail_reason"`
	SubmitTime int64  `json:"submit_time"`
	StartTime  int64  `json:"start_time"`
	FinishTime int64  `json:"finish_time"`
	Progress   string `json:"progress"`
	Data       struct {
		Output string `json:"output"`
	} `json:"data"`
	SearchItem string `json:"search_item"`
}

func (s *Service) QuerySora2Task(taskId string, channel string) (Sora2TaskVo, error) {
	// 读取 API KEY
	var apiKey model.ApiKey
	err := s.db.Session(&gorm.Session{}).Where("type", "luma").
		Where("api_url", channel).
		Where("enabled", true).
		Order("last_used_at DESC").First(&apiKey).Error
	if err != nil {
		return Sora2TaskVo{}, errors.New("no available API KEY for Sora2")
	}

	apiURL := fmt.Sprintf("%s/v2/videos/generations/%s", apiKey.ApiURL, taskId)
	var res Sora2TaskVo
	r, err := s.httpClient.R().SetHeader("Authorization", "Bearer "+apiKey.Value).Get(apiURL)

	if err != nil {
		return Sora2TaskVo{}, fmt.Errorf("请求 API 失败：%v", err)
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return Sora2TaskVo{}, fmt.Errorf("API 返回失败：%v", r.String())
	}

	body, _ := io.ReadAll(r.Body)
	err = json.Unmarshal(body, &res)
	if err != nil {
		return Sora2TaskVo{}, fmt.Errorf("解析API数据失败：%v, %s", err, string(body))
	}

	return res, nil
}

// pushProgressUpdate 推送进度更新到WebSocket
func (s *Service) pushProgressUpdate(taskId uint, progress string, videoURL, coverURL, waterURL, errMsg string) {
	// 查找任务信息
	var job model.Sora2Job
	err := s.db.Where("id = ?", taskId).First(&job).Error
	if err != nil {
		logger.Errorf("find task for websocket push error: %v", err)
		return
	}

	// 构建进度消息 - 直接发送API响应格式，而不是包装在raw_data中
	var bodyData gin.H

	// 如果有raw_data，解析它作为body
	if job.RawData != "" {
		var rawData map[string]interface{}
		if err := json.Unmarshal([]byte(job.RawData), &rawData); err == nil {
			bodyData = gin.H(rawData)
			// 确保包含job_id用于前端匹配
			bodyData["job_id"] = job.Id
		} else {
			// 如果解析失败，使用默认格式
			bodyData = gin.H{
				"task_id":   job.TaskId,
				"progress":  progress,
				"video_url": videoURL,
				"cover_url": coverURL,
				"water_url": waterURL,
				"err_msg":   errMsg,
				"status":    s.mapProgressToStatus(progress),
				"job_id":    job.Id,
			}
		}
	} else {
		// 如果没有raw_data，使用默认格式
		bodyData = gin.H{
			"task_id":   job.TaskId,
			"progress":  progress,
			"video_url": videoURL,
			"cover_url": coverURL,
			"water_url": waterURL,
			"err_msg":   errMsg,
			"status":    s.mapProgressToStatus(progress),
			"job_id":    job.Id,
		}
	}

	progressMsg := types.ReplyMessage{
		Channel:  types.ChSora2,
		ClientId: "",
		Type:     types.MsgTypeText,
		Body:     bodyData,
	}

	// 发送给特定用户的所有客户端
	if job.UserId > 0 {
		clients := s.wsService.Clients.ToList()
		logger.Infof("推送Sora2进度更新: task_id=%s, progress=%s, user_id=%d, clients_count=%d", job.TaskId, progress, job.UserId, len(clients))
		logger.Infof("WebSocket消息内容: %+v", progressMsg)
		for _, client := range clients {
			// 这里需要根据实际情况获取用户ID，暂时发送给所有客户端
			logger.Infof("发送WebSocket消息给客户端: %s", client.Id)
			client.SendJson(progressMsg)
		}
	} else {
		logger.Warnf("任务没有用户ID，无法发送WebSocket消息: job_id=%d", job.Id)
	}

	// 如果有ClientId，也发送给特定客户端
	if clientId, exists := s.clientIds[taskId]; exists {
		if client := s.wsService.Clients.Get(clientId); client != nil {
			client.SendJson(progressMsg)
		}
	}
}

// mapApiStatusToDbStatus 将API状态映射为数据库状态
func (s *Service) mapApiStatusToDbStatus(apiStatus string) string {
	switch apiStatus {
	case "SUCCESS":
		return "completed"
	case "FAILURE":
		return "failed"
	case "IN_PROGRESS":
		return "processing"
	default:
		return "pending"
	}
}

// extractProgressFromApi 从API进度字符串中提取数字
func (s *Service) extractProgressFromApi(progressStr string) string {
	// 处理 "100%" 格式的进度字符串
	if len(progressStr) > 0 && progressStr[len(progressStr)-1] == '%' {
		return progressStr[:len(progressStr)-1]
	}
	return progressStr
}

// mapProgressToStatus 将进度值映射为状态
func (s *Service) mapProgressToStatus(progress string) string {
	switch progress {
	case "0":
		return "NOT_START"
	case "100":
		return "SUCCESS"
	case "failed", "error":
		return "FAILURE"
	default:
		// 如果是数字，检查是否在1-99之间
		if progress >= "1" && progress <= "99" {
			return "IN_PROGRESS"
		}
		return "NOT_START"
	}
}
