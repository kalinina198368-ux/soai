package video_square

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
	logger2 "geekai/logger"
	"geekai/service/oss"
	"geekai/store/model"
	_ "geekai/utils"
	"time"

	"gorm.io/gorm"
)

var logger = logger2.GetLogger()

// Service 视频广场服务
type Service struct {
	db              *gorm.DB
	uploaderManager *oss.UploaderManager
	config          *types.AppConfig
}

// Sora2FeedResponse sora2官网feed响应结构
type Sora2FeedResponse struct {
	Items  []Sora2FeedItem `json:"items"`
	Cursor string          `json:"cursor"`
}

// Sora2FeedItem feed项
type Sora2FeedItem struct {
	Post    Sora2Post    `json:"post"`
	Profile Sora2Profile `json:"profile"`
}

// Sora2Post post数据
type Sora2Post struct {
	Id          string            `json:"id"`
	Text        string            `json:"text"`
	ViewCount   int               `json:"view_count"`
	LikeCount   int               `json:"like_count"`
	Attachments []Sora2Attachment `json:"attachments"`
}

// Sora2Attachment 附件数据
type Sora2Attachment struct {
	Id              string         `json:"id"`
	Kind            string         `json:"kind"`
	Url             string         `json:"url"`
	DownloadableUrl string         `json:"downloadable_url"`
	Width           int            `json:"width"`
	Height          int            `json:"height"`
	NFrames         int            `json:"n_frames"`
	Prompt          string         `json:"prompt"`
	Encodings       Sora2Encodings `json:"encodings"`
}

// Sora2Encodings 编码信息
type Sora2Encodings struct {
	Source    Sora2EncodingPath `json:"source"`
	Thumbnail Sora2EncodingPath `json:"thumbnail"`
}

// Sora2EncodingPath 编码路径
type Sora2EncodingPath struct {
	Path string `json:"path"`
}

// Sora2Profile 用户信息
type Sora2Profile struct {
	UserId            string `json:"user_id"`
	Username          string `json:"username"`
	ProfilePictureUrl string `json:"profile_picture_url"`
}

func NewService(db *gorm.DB, uploaderManager *oss.UploaderManager, config *types.AppConfig) *Service {
	return &Service{
		db:              db,
		uploaderManager: uploaderManager,
		config:          config,
	}
}

// FetchAndSaveVideos 从JSON字符串解析视频数据并保存
// jsonData: 从Postman获取的JSON字符串数据
func (s *Service) FetchAndSaveVideos(jsonData string) error {
	logger.Info("开始处理视频数据...")

	// 解析JSON
	var feedResp Sora2FeedResponse
	if err := json.Unmarshal([]byte(jsonData), &feedResp); err != nil {
		return fmt.Errorf("解析JSON失败: %v", err)
	}

	logger.Infof("获取到 %d 条视频数据", len(feedResp.Items))

	// 处理每条视频
	successCount := 0
	for _, item := range feedResp.Items {
		if err := s.processVideoItem(item); err != nil {
			logger.Errorf("处理视频失败 (post_id: %s): %v", item.Post.Id, err)
			continue
		}
		successCount++
	}

	logger.Infof("成功处理 %d 条视频", successCount)
	return nil
}

// processVideoItem 处理单个视频项
func (s *Service) processVideoItem(item Sora2FeedItem) error {
	// 检查是否已存在
	var existing model.VideoSquare
	result := s.db.Where("post_id = ?", item.Post.Id).First(&existing)
	if result.Error == nil {
		// 已存在，更新观看次数和点赞次数
		s.db.Model(&existing).Updates(map[string]interface{}{
			"view_count": item.Post.ViewCount,
			"like_count": item.Post.LikeCount,
			"updated_at": time.Now(),
		})
		return nil
	}

	// 查找视频附件
	var videoAttachment *Sora2Attachment
	for _, att := range item.Post.Attachments {
		if att.Kind == "sora" && att.DownloadableUrl != "" {
			videoAttachment = &att
			break
		}
	}

	if videoAttachment == nil {
		return fmt.Errorf("未找到视频附件")
	}

	// 获取提示词（优先使用attachment中的prompt，否则使用post的text）
	prompt := videoAttachment.Prompt
	if prompt == "" {
		prompt = item.Post.Text
	}

	// 下载视频到OSS
	videoURL, err := s.downloadVideoToOSS(videoAttachment.DownloadableUrl)
	if err != nil {
		return fmt.Errorf("下载视频失败: %v", err)
	}

	// 下载封面图到OSS
	posterURL := ""
	if videoAttachment.Encodings.Thumbnail.Path != "" {
		posterURL, err = s.uploaderManager.GetUploadHandler().PutUrlFile(videoAttachment.Encodings.Thumbnail.Path, false)
		if err != nil {
			logger.Warnf("下载封面图失败: %v", err)
		}
	}

	// 下载头像到OSS
	avatarURL := ""
	if item.Profile.ProfilePictureUrl != "" {
		avatarURL, err = s.uploaderManager.GetUploadHandler().PutUrlFile(item.Profile.ProfilePictureUrl, false)
		if err != nil {
			logger.Warnf("下载头像失败: %v", err)
		}
	}

	// 保存原始数据
	rawData, _ := json.Marshal(item)

	// 创建数据库记录
	videoSquare := model.VideoSquare{
		PostId:       item.Post.Id,
		OriginalURL:  videoAttachment.DownloadableUrl,
		OssURL:       videoURL,
		PosterURL:    posterURL,
		ThumbnailURL: videoAttachment.Encodings.Thumbnail.Path,
		Prompt:       prompt,
		Author:       item.Profile.Username,
		AvatarURL:    avatarURL,
		ViewCount:    item.Post.ViewCount,
		LikeCount:    item.Post.LikeCount,
		Width:        videoAttachment.Width,
		Height:       videoAttachment.Height,
		NFrames:      videoAttachment.NFrames,
		RawData:      string(rawData),
	}

	if err := s.db.Create(&videoSquare).Error; err != nil {
		return fmt.Errorf("保存数据库失败: %v", err)
	}

	logger.Infof("成功保存视频: post_id=%s, author=%s", item.Post.Id, item.Profile.Username)
	return nil
}

// downloadVideoToOSS 下载视频到OSS
func (s *Service) downloadVideoToOSS(videoURL string) (string, error) {
	// 直接使用PutUrlFile方法，它会自动下载并上传
	uploader := s.uploaderManager.GetUploadHandler()
	ossURL, err := uploader.PutUrlFile(videoURL, s.config.ProxyURL != "")
	if err != nil {
		return "", fmt.Errorf("上传视频到OSS失败: %v", err)
	}
	return ossURL, nil
}

// Run 已废弃，不再需要定时任务
// 现在需要手动调用 FetchAndSaveVideos(jsonData) 方法
func (s *Service) Run() {
	logger.Info("视频广场服务已启动，请手动调用 FetchAndSaveVideos(jsonData) 方法处理数据")
}
