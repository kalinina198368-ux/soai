package mj

// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// * Copyright 2023 The Geek-AI Authors. All rights reserved.
// * Use of this source code is governed by a Apache-2.0 license
// * that can be found in the LICENSE file.
// * @Author yangjian102621@163.com
// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

import (
	"encoding/json"
	"fmt"
	"geekai/core/types"
	"geekai/service"
	"geekai/service/oss"
	"geekai/store"
	"geekai/store/model"
	"geekai/utils"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"gorm.io/gorm"
)

// Service MJ 绘画服务
type Service struct {
	client          *Client // MJ Client
	taskQueue       *store.RedisQueue
	notifyQueue     *store.RedisQueue
	db              *gorm.DB
	wsService       *service.WebsocketService
	uploaderManager *oss.UploaderManager
	userService     *service.UserService
	clientIds       map[uint]string
}

func NewService(redisCli *redis.Client, db *gorm.DB, client *Client, manager *oss.UploaderManager, wsService *service.WebsocketService, userService *service.UserService) *Service {
	return &Service{
		db:              db,
		taskQueue:       store.NewRedisQueue("MidJourney_Task_Queue", redisCli),
		notifyQueue:     store.NewRedisQueue("MidJourney_Notify_Queue", redisCli),
		client:          client,
		wsService:       wsService,
		uploaderManager: manager,
		clientIds:       map[uint]string{},
		userService:     userService,
	}
}

func (s *Service) Run() {
	// 将数据库中未提交的人物加载到队列
	var jobs []model.MidJourneyJob
	s.db.Where("task_id", "").Where("progress", 0).Find(&jobs)
	for _, v := range jobs {
		var task types.MjTask
		err := utils.JsonDecode(v.TaskInfo, &task)
		if err != nil {
			logger.Errorf("decode task info with error: %v", err)
			continue
		}
		task.Id = v.Id
		s.clientIds[task.Id] = task.ClientId
		s.PushTask(task)
	}

	logger.Info("Starting MidJourney job consumer for service")
	go func() {
		for {
			var task types.MjTask
			err := s.taskQueue.LPop(&task)
			if err != nil {
				logger.Errorf("taking task with error: %v", err)
				continue
			}

			//不需要翻译
			// translate prompt
			//if utils.HasChinese(task.Prompt) {
			//	content, err := utils.OpenAIRequest(s.db, fmt.Sprintf(service.TranslatePromptTemplate, task.Prompt), task.TranslateModelId)
			//	if err == nil {
			//		task.Prompt = content
			//	} else {
			//		logger.Warnf("error with translate prompt: %v", err)
			//	}
			//}
			//// translate negative prompt
			//if task.NegPrompt != "" && utils.HasChinese(task.NegPrompt) {
			//	content, err := utils.OpenAIRequest(s.db, fmt.Sprintf(service.TranslatePromptTemplate, task.NegPrompt), task.TranslateModelId)
			//	if err == nil {
			//		task.NegPrompt = content
			//	} else {
			//		logger.Warnf("error with translate prompt: %v", err)
			//	}
			//}

			// use fast mode as default
			if task.Mode == "" {
				task.Mode = "fast"
			}
			s.clientIds[task.Id] = task.ClientId

			var job model.MidJourneyJob
			tx := s.db.Where("id = ?", task.Id).First(&job)
			if tx.Error != nil {
				logger.Error("任务不存在，任务ID：", task.TaskId)
				continue
			}

			logger.Infof("handle a new Nano Banana task: %+v", task)
			// 通知前端任务开始执行
			s.notifyQueue.RPush(service.NotifyMessage{
				ClientId: task.ClientId,
				UserId:   job.UserId,
				JobId:    int(job.Id),
				Message:  service.TaskStatusRunning,
			})
			var res ImageRes
			switch task.Type {
			case types.TaskImage:
				res, err = s.client.Imagine(task)
				break
			case types.TaskUpscale:
				res, err = s.client.Upscale(task)
				break
			case types.TaskVariation:
				res, err = s.client.Variation(task)
				break
			case types.TaskBlend:
				res, err = s.client.Blend(task)
				break
			case types.TaskSwapFace:
				res, err = s.client.SwapFace(task)
				break
			}

			if err != nil {
				errMsg := err.Error()
				logger.Error("绘画任务执行失败：", errMsg)
				job.Progress = service.FailTaskProgress
				job.ErrMsg = errMsg
				// update the task progress
				s.db.Updates(&job)
				// 任务失败，通知前端
				s.notifyQueue.RPush(service.NotifyMessage{ClientId: task.ClientId, UserId: task.UserId, JobId: int(job.Id), Message: service.TaskStatusFailed})
				continue
			}
			if len(res.URLs) == 0 {
				errMsg := "nano-banana 未返回图片链接"
				logger.Error(errMsg)
				job.Progress = service.FailTaskProgress
				job.ErrMsg = errMsg
				s.db.Updates(&job)
				s.notifyQueue.RPush(service.NotifyMessage{ClientId: task.ClientId, UserId: task.UserId, JobId: int(job.Id), Message: service.TaskStatusFailed})
				continue
			}

			logger.Infof("nano-banana 任务完成：%+v", res)
			job.Progress = 100
			job.OrgURL = res.URLs[0]
			job.ChannelId = ""
			job.MessageId = ""
			job.ErrMsg = ""
			s.db.Updates(&job)
		}
	}()
}

type CBReq struct {
	Id          string      `json:"id"`
	Action      string      `json:"action"`
	Status      string      `json:"status"`
	Prompt      string      `json:"prompt"`
	PromptEn    string      `json:"promptEn"`
	Description string      `json:"description"`
	SubmitTime  int64       `json:"submitTime"`
	StartTime   int64       `json:"startTime"`
	FinishTime  int64       `json:"finishTime"`
	Progress    string      `json:"progress"`
	ImageUrl    string      `json:"imageUrl"`
	FailReason  interface{} `json:"failReason"`
	Properties  struct {
		FinalPrompt string `json:"finalPrompt"`
	} `json:"properties"`
}

func GetImageHash(action string) string {
	split := strings.Split(action, "::")
	if len(split) > 5 {
		return split[4]
	}
	return split[len(split)-1]
}

func (s *Service) CheckTaskNotify() {
	go func() {
		for {
			var message service.NotifyMessage
			err := s.notifyQueue.LPop(&message)
			if err != nil {
				continue
			}
			logger.Debugf("receive a new mj notify message: %+v", message)
			client := s.wsService.Clients.Get(message.ClientId)
			if client == nil {
				continue
			}
			utils.SendChannelMsg(client, types.ChMj, message.Message)
		}
	}()
}

func (s *Service) DownloadImages() {
	go func() {
		var items []model.MidJourneyJob
		for {
			res := s.db.Where("img_url = ? AND progress = ?", "", 100).Find(&items)
			if res.Error != nil {
				continue
			}

			// download images
			for _, v := range items {
				if v.OrgURL == "" {
					continue
				}

				logger.Infof("try to download image: %s", v.OrgURL)
				// 如果是返回的是 discord 图片地址，则使用代理下载
				proxy := false
				if strings.HasPrefix(v.OrgURL, "https://cdn.discordapp.com") {
					proxy = true
				}
				imgURL, err := s.uploaderManager.GetUploadHandler().PutUrlFile(v.OrgURL, proxy)

				if err != nil {
					logger.Errorf("error with download image %s, %v", v.OrgURL, err)
					continue
				} else {
					logger.Infof("download image %s successfully.", v.OrgURL)
				}

				v.ImgURL = imgURL
				s.db.Updates(&v)

				s.notifyQueue.RPush(service.NotifyMessage{
					ClientId: s.clientIds[v.Id],
					UserId:   v.UserId,
					JobId:    int(v.Id),
					Message:  service.TaskStatusFinished})
			}

			time.Sleep(time.Second * 5)
		}
	}()
}

// PushTask push a new mj task in to task queue
func (s *Service) PushTask(task types.MjTask) {
	logger.Debugf("add a new MidJourney task to the task list: %+v", task)
	s.taskQueue.RPush(task)
}

// SyncTaskProgress 异步拉取任务
func (s *Service) SyncTaskProgress() {
	go func() {
		var jobs []model.MidJourneyJob
		for {
			err := s.db.Where("progress < ?", 100).Find(&jobs).Error
			if err != nil {
				continue
			}

			for _, job := range jobs {
				// 10 分钟还没完成的任务标记为失败
				if time.Now().Sub(job.CreatedAt) > time.Minute*10 {
					job.Progress = service.FailTaskProgress
					job.ErrMsg = "任务超时"
					s.db.Updates(&job)
					s.notifyQueue.RPush(service.NotifyMessage{
						ClientId: s.clientIds[job.Id],
						UserId:   job.UserId,
						JobId:    int(job.Id),
						Message:  service.TaskStatusFailed,
					})
				}
			}

			// 找出失败的任务，并恢复其扣减算力
			s.db.Where("progress", service.FailTaskProgress).Where("power > ?", 0).Find(&jobs)
			for _, job := range jobs {
				// 从任务参数中解析模型名称
				modelName := "nano-banana" // 默认模型
				var task types.MjTask
				if err := utils.JsonDecode(job.TaskInfo, &task); err == nil {
					// task.Params 对应 JSON 中的 full_prompt 字段，它是一个 JSON 字符串
					if task.Params != "" {
						type nanoParams struct {
							Model string `json:"model,omitempty"`
						}
						var params nanoParams
						if err := json.Unmarshal([]byte(task.Params), &params); err == nil && params.Model != "" {
							modelName = params.Model
						}
					}
				}

				err := s.userService.IncreasePower(job.UserId, job.Power, model.PowerLog{
					Type:   types.PowerRefund,
					Model:  modelName,
					Remark: fmt.Sprintf("任务失败，退回算力。任务ID：", job.Id),
				})
				if err != nil {
					continue
				}
				// 更新任务状态
				s.db.Model(&job).UpdateColumn("power", 0)
			}

			time.Sleep(time.Second * 5)
		}
	}()
}
