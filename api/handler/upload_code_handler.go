package handler

// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// * Copyright 2023 The Geek-AI Authors. All rights reserved.
// * Use of this source code is governed by a Apache-2.0 license
// * that can be found in the LICENSE file.
// * @Author yangjian102621@163.com
// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

import (
	"geekai/core"
	"geekai/core/types"
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

// UploadCodeHandler 上传码（匿名扫码上传图片）
type UploadCodeHandler struct {
	BaseHandler
	uploaderManager *oss.UploaderManager
}

func NewUploadCodeHandler(app *core.AppServer, db *gorm.DB, manager *oss.UploaderManager) *UploadCodeHandler {
	return &UploadCodeHandler{
		BaseHandler:     BaseHandler{App: app, DB: db},
		uploaderManager: manager,
	}
}

// Code 获取当前登录用户的上传码（不存在则创建）
func (h *UploadCodeHandler) Code(c *gin.Context) {
	userId := h.GetLoginUserId(c)
	if userId == 0 {
		resp.NotAuth(c, "请先登录")
		return
	}

	var uploadCode model.UploadCode
	res := h.DB.Where("user_id = ?", userId).First(&uploadCode)
	if res.Error != nil {
		code := strings.ToUpper(utils.RandString(10))
		for {
			var existed model.UploadCode
			r2 := h.DB.Where("code = ?", code).First(&existed)
			if r2.Error != nil {
				break
			}
			code = strings.ToUpper(utils.RandString(10))
		}
		uploadCode.UserId = uint(userId)
		uploadCode.Code = code
		uploadCode.CreatedAt = time.Now()
		_ = h.DB.Create(&uploadCode).Error
	}

	resp.SUCCESS(c, vo.UploadCode{
		Id:        uploadCode.Id,
		UserId:    uploadCode.UserId,
		Code:      uploadCode.Code,
		CreatedAt: uploadCode.CreatedAt.Unix(),
	})
}

// Upload 匿名扫码上传图片（不要求登录）
// multipart/form-data: file=<file>
// query/form: code=<upload_code>
func (h *UploadCodeHandler) Upload(c *gin.Context) {
	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		code = strings.TrimSpace(c.PostForm("code"))
	}
	if code == "" {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	// 找到上传码归属用户
	var uploadCode model.UploadCode
	res := h.DB.Where("code = ?", code).First(&uploadCode)
	if res.Error != nil || uploadCode.Id == 0 || uploadCode.UserId == 0 {
		resp.ERROR(c, "无效的上传码")
		return
	}

	file, err := h.uploaderManager.GetUploadHandler().PutFile(c, "file")
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	// 文件名过长截断（与 NetHandler 保持一致）
	if len(file.Name) > 100 {
		file.Name = file.Name[:90] + file.Ext
	}

	item := model.UploadCodeImage{
		UserId:     uploadCode.UserId,
		Code:       uploadCode.Code,
		URL:        file.URL,
		ObjKey:     file.ObjKey,
		Name:       file.Name,
		Ext:        file.Ext,
		Size:       file.Size,
		UploaderIP: c.ClientIP(),
		CreatedAt:  time.Now(),
	}
	if err := h.DB.Create(&item).Error; err != nil {
		resp.ERROR(c, "保存上传记录失败："+err.Error())
		return
	}

	// 返回与 /api/upload 一致的 file 结构，方便前端直接使用 url
	resp.SUCCESS(c, file)
}

// Images 获取当前用户上传码下的图片列表（需要登录）
func (h *UploadCodeHandler) Images(c *gin.Context) {
	userId := h.GetLoginUserId(c)
	if userId == 0 {
		resp.NotAuth(c, "请先登录")
		return
	}

	page := h.GetInt(c, "page", 1)
	pageSize := h.GetInt(c, "page_size", 30)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 30
	}

	session := h.DB.Session(&gorm.Session{}).Where("user_id = ?", userId)

	// 统计总数
	var total int64
	session.Model(&model.UploadCodeImage{}).Count(&total)

	var items []model.UploadCodeImage
	offset := (page - 1) * pageSize
	err := session.Order("id desc").Offset(offset).Limit(pageSize).Find(&items).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	list := make([]vo.UploadCodeImage, 0, len(items))
	for _, v := range items {
		list = append(list, vo.UploadCodeImage{
			Id:         v.Id,
			UserId:     v.UserId,
			Code:       v.Code,
			URL:        v.URL,
			ObjKey:     v.ObjKey,
			Name:       v.Name,
			Ext:        v.Ext,
			Size:       v.Size,
			UploaderIP: v.UploaderIP,
			CreatedAt:  v.CreatedAt.Unix(),
		})
	}

	resp.SUCCESS(c, vo.NewPage(total, page, pageSize, list))
}

// Remove 删除一条上传码图片记录（需要登录）
func (h *UploadCodeHandler) Remove(c *gin.Context) {
	userId := h.GetLoginUserId(c)
	if userId == 0 {
		resp.NotAuth(c, "请先登录")
		return
	}
	id := h.GetInt(c, "id", 0)
	if id <= 0 {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	var item model.UploadCodeImage
	tx := h.DB.Where("id = ? AND user_id = ?", id, userId).First(&item)
	if tx.Error != nil || item.Id == 0 {
		resp.ERROR(c, "记录不存在")
		return
	}

	// 删除记录（不强制删除对象存储文件，避免误删）
	if err := h.DB.Delete(&model.UploadCodeImage{}, "id = ? AND user_id = ?", id, userId).Error; err != nil {
		resp.ERROR(c, "删除失败："+err.Error())
		return
	}

	resp.SUCCESS(c)
}
