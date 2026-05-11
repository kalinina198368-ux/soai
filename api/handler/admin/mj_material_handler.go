package admin

// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// * Copyright 2023 The Geek-AI Authors. All rights reserved.
// * Use of this source code is governed by a Apache-2.0 license
// * that can be found in the LICENSE file.
// * @Author yangjian102621@163.com
// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

import (
	"geekai/core"
	"geekai/core/types"
	"geekai/handler"
	"geekai/store/model"
	"geekai/store/vo"
	"geekai/utils"
	"geekai/utils/resp"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MjMaterialHandler struct {
	handler.BaseHandler
}

func NewMjMaterialHandler(app *core.AppServer, db *gorm.DB) *MjMaterialHandler {
	return &MjMaterialHandler{BaseHandler: handler.BaseHandler{App: app, DB: db}}
}

// CategoryList 获取分类列表
func (h *MjMaterialHandler) CategoryList(c *gin.Context) {
	var items []model.MjMaterialCategory
	err := h.DB.Order("sort_order DESC, id DESC").Find(&items).Error
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	list := make([]vo.MjMaterialCategory, 0, len(items))
	for _, v := range items {
		var item vo.MjMaterialCategory
		if err = utils.CopyObject(v, &item); err != nil {
			continue
		}
		item.Id = v.Id
		list = append(list, item)
	}

	resp.SUCCESS(c, list)
}

// CategorySave 创建或更新分类
func (h *MjMaterialHandler) CategorySave(c *gin.Context) {
	var data struct {
		Id        uint   `json:"id"`
		Name      string `json:"name"`
		Title     string `json:"title"`
		IsActive  bool   `json:"is_active"`
		SortOrder int    `json:"sort_order"`
		CreatedAt int64  `json:"created_at"` // 秒级时间戳
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	if data.Name == "" || data.Title == "" {
		resp.ERROR(c, "name 和 title 不能为空")
		return
	}

	item := model.MjMaterialCategory{
		Name:      data.Name,
		Title:     data.Title,
		IsActive:  data.IsActive,
		SortOrder: data.SortOrder,
	}
	item.Id = data.Id

	// 处理创建时间：前端传入秒级时间戳，否则回落到已有记录或当前时间
	var reqCreatedAt time.Time
	if data.CreatedAt > 0 {
		reqCreatedAt = time.Unix(data.CreatedAt, 0)
	}

	if item.Id > 0 {
		// 更新时保留原始创建时间，避免被清零
		if !reqCreatedAt.IsZero() {
			item.CreatedAt = reqCreatedAt
		} else {
			var old model.MjMaterialCategory
			if err := h.DB.First(&old, item.Id).Error; err == nil {
				item.CreatedAt = old.CreatedAt
			}
		}
		item.UpdatedAt = time.Now()
	} else {
		if reqCreatedAt.IsZero() {
			reqCreatedAt = time.Now()
		}
		item.CreatedAt = reqCreatedAt
		item.UpdatedAt = reqCreatedAt
	}

	if err := h.DB.Save(&item).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	var voItem vo.MjMaterialCategory
	_ = utils.CopyObject(item, &voItem)
	voItem.Id = item.Id
	resp.SUCCESS(c, voItem)
}

// CategoryEnable 启用/停用分类
func (h *MjMaterialHandler) CategoryEnable(c *gin.Context) {
	var data struct {
		Id       uint `json:"id"`
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	if err := h.DB.Model(&model.MjMaterialCategory{}).Where("id = ?", data.Id).Update("is_active", data.IsActive).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}
	resp.SUCCESS(c)
}

// CategorySort 批量调整分类排序
func (h *MjMaterialHandler) CategorySort(c *gin.Context) {
	var data struct {
		Ids   []uint `json:"ids"`
		Sorts []int  `json:"sorts"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}
	if len(data.Ids) != len(data.Sorts) {
		resp.ERROR(c, "参数长度不一致")
		return
	}

	for i, id := range data.Ids {
		if err := h.DB.Model(&model.MjMaterialCategory{}).Where("id = ?", id).Update("sort_order", data.Sorts[i]).Error; err != nil {
			resp.ERROR(c, err.Error())
			return
		}
	}
	resp.SUCCESS(c)
}

// CategoryRemove 删除分类
func (h *MjMaterialHandler) CategoryRemove(c *gin.Context) {
	id := h.GetInt(c, "id", 0)
	if id <= 0 {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	// 可选：同时删除该分类下素材
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("category_id = ?", id).Delete(&model.MjMaterial{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.MjMaterialCategory{}).Error
	})
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}
	resp.SUCCESS(c)
}

// MaterialList 素材列表
func (h *MjMaterialHandler) MaterialList(c *gin.Context) {
	var data struct {
		CategoryId int    `json:"category_id"`
		Title      string `json:"title"`
		Type       string `json:"type"`
		IsActive   *bool  `json:"is_active"`
		Page       int    `json:"page"`
		PageSize   int    `json:"page_size"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}
	if data.Page <= 0 {
		data.Page = 1
	}
	if data.PageSize <= 0 || data.PageSize > 200 {
		data.PageSize = 20
	}

	session := h.DB.Model(&model.MjMaterial{})
	if data.CategoryId > 0 {
		session = session.Where("category_id = ?", data.CategoryId)
	}
	if data.Title != "" {
		session = session.Where("title LIKE ?", "%"+data.Title+"%")
	}
	if data.Type != "" {
		session = session.Where("type = ?", data.Type)
	}
	if data.IsActive != nil {
		session = session.Where("is_active = ?", *data.IsActive)
	}

	var total int64
	if err := session.Count(&total).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	offset := (data.Page - 1) * data.PageSize
	var items []model.MjMaterial
	if err := session.Order("sort_order DESC, id DESC").Offset(offset).Limit(data.PageSize).Find(&items).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	list := make([]vo.MjMaterial, 0, len(items))
	for _, v := range items {
		var item vo.MjMaterial
		if err := utils.CopyObject(v, &item); err != nil {
			continue
		}
		item.Id = v.Id
		item.CategoryId = v.CategoryId
		item.IsActive = v.IsActive
		list = append(list, item)
	}

	resp.SUCCESS(c, vo.NewPage(total, data.Page, data.PageSize, list))
}

// MaterialSave 创建或更新素材
func (h *MjMaterialHandler) MaterialSave(c *gin.Context) {
	var data struct {
		Id         uint   `json:"id"`
		CategoryId uint   `json:"category_id"`
		Title      string `json:"title"`
		Name       string `json:"name"`
		Image      string `json:"image"`
		//Preview    string `json:"preview"`
		Prompt    string `json:"prompt"`
		Type      string `json:"type"`
		IsActive  bool   `json:"is_active"`
		SortOrder int    `json:"sort_order"`
		CreatedAt int64  `json:"created_at"` // 前端传秒级时间戳
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}
	if data.CategoryId == 0 || data.Title == "" || data.Name == "" {
		resp.ERROR(c, "category_id、title、name 不能为空")
		return
	}

	item := model.MjMaterial{
		CategoryId: data.CategoryId,
		Title:      data.Title,
		Name:       data.Name,
		Image:      data.Image,
		//Preview:    data.Preview,
		Prompt:    data.Prompt,
		Type:      data.Type,
		IsActive:  data.IsActive,
		SortOrder: data.SortOrder,
	}
	item.Id = data.Id
	// 处理创建时间：前端传入秒级时间戳，否则回落到已有记录或当前时间
	var reqCreatedAt time.Time
	if data.CreatedAt > 0 {
		reqCreatedAt = time.Unix(data.CreatedAt, 0)
	}

	if item.Id > 0 {
		// 更新时保留原始创建时间
		if !reqCreatedAt.IsZero() {
			item.CreatedAt = reqCreatedAt
		} else {
			var old model.MjMaterial
			if err := h.DB.First(&old, item.Id).Error; err == nil {
				item.CreatedAt = old.CreatedAt
			}
		}
		item.UpdatedAt = time.Now()
	} else {
		if reqCreatedAt.IsZero() {
			reqCreatedAt = time.Now()
		}
		item.CreatedAt = reqCreatedAt
		item.UpdatedAt = reqCreatedAt
	}

	if err := h.DB.Save(&item).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}

	var voItem vo.MjMaterial
	_ = utils.CopyObject(item, &voItem)
	voItem.Id = item.Id
	voItem.CategoryId = item.CategoryId
	//voItem.CreatedAt = item.CreatedAt
	resp.SUCCESS(c, voItem)
}

// MaterialEnable 启用/停用素材
func (h *MjMaterialHandler) MaterialEnable(c *gin.Context) {
	var data struct {
		Id       uint `json:"id"`
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	if err := h.DB.Model(&model.MjMaterial{}).Where("id = ?", data.Id).Update("is_active", data.IsActive).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}
	resp.SUCCESS(c)
}

// MaterialSort 批量调整素材排序
func (h *MjMaterialHandler) MaterialSort(c *gin.Context) {
	var data struct {
		Ids   []uint `json:"ids"`
		Sorts []int  `json:"sorts"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}
	if len(data.Ids) != len(data.Sorts) {
		resp.ERROR(c, "参数长度不一致")
		return
	}

	for i, id := range data.Ids {
		if err := h.DB.Model(&model.MjMaterial{}).Where("id = ?", id).Update("sort_order", data.Sorts[i]).Error; err != nil {
			resp.ERROR(c, err.Error())
			return
		}
	}
	resp.SUCCESS(c)
}

// MaterialRemove 删除素材
func (h *MjMaterialHandler) MaterialRemove(c *gin.Context) {
	id := h.GetInt(c, "id", 0)
	if id <= 0 {
		resp.ERROR(c, types.InvalidArgs)
		return
	}

	if err := h.DB.Where("id = ?", id).Delete(&model.MjMaterial{}).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}
	resp.SUCCESS(c)
}
