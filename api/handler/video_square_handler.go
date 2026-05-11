package handler

// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// * Copyright 2023 The Geek-AI Authors. All rights reserved.
// * Use of this source code is governed by a Apache-2.0 license
// * that can be found in the LICENSE file.
// * @Author yangjian102621@163.com
// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

import (
	"geekai/core"
	"geekai/service/video_square"
	"geekai/store/model"
	"geekai/utils/resp"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VideoSquareHandler struct {
	BaseHandler
	service *video_square.Service
}

func NewVideoSquareHandler(app *core.AppServer, db *gorm.DB, service *video_square.Service) *VideoSquareHandler {
	return &VideoSquareHandler{
		BaseHandler: BaseHandler{
			App: app,
			DB:  db,
		},
		service: service,
	}
}

// List 获取视频广场列表
func (h *VideoSquareHandler) List(c *gin.Context) {
	page := h.GetInt(c, "page", 1)
	pageSize := h.GetInt(c, "page_size", 20)
	
	session := h.DB.Model(&model.VideoSquare{}).Where("oss_url != ?", "")
	
	// 统计总数
	var total int64
	session.Count(&total)
	
	// 分页查询
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		session = session.Offset(offset).Limit(pageSize)
	}
	
	var list []model.VideoSquare
	err := session.Order("id desc").Find(&list).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}
	
	// 转换为前端需要的格式
	items := make([]map[string]interface{}, 0)
	for _, v := range list {
		item := map[string]interface{}{
			"id":        v.Id,
			"videoUrl":  v.OssURL,
			"poster":    v.PosterURL,
			"prompt":    v.Prompt,
			"author":    v.Author,
			"avatar":    v.AvatarURL,
			"viewCount": v.ViewCount,
			"likeCount": v.LikeCount,
		}
		items = append(items, item)
	}
	
	resp.SUCCESS(c, map[string]interface{}{
		"items": items,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// IncrementView 增加观看次数
func (h *VideoSquareHandler) IncrementView(c *gin.Context) {
	var data struct {
		Id int `json:"id"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, "id is required")
		return
	}
	
	if data.Id == 0 {
		resp.ERROR(c, "id is required")
		return
	}
	
	err := h.DB.Model(&model.VideoSquare{}).Where("id = ?", data.Id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}
	
	resp.SUCCESS(c)
}

// ImportData 手动导入JSON数据
func (h *VideoSquareHandler) ImportData(c *gin.Context) {
	var data struct {
		JsonData string `json:"json_data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, "json_data is required")
		return
	}

	if data.JsonData == "" {
		resp.ERROR(c, "json_data cannot be empty")
		return
	}

	// 调用service处理数据
	if err := h.service.FetchAndSaveVideos(data.JsonData); err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	resp.SUCCESS(c, "数据导入成功")
}

