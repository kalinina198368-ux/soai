package mj

// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// * Copyright 2023 The Geek-AI Authors. All rights reserved.
// * Use of this source code is governed by a Apache-2.0 license
// * that can be found in the LICENSE file.
// * @Author yangjian102621@163.com
// * +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"geekai/core/types"
	logger2 "geekai/logger"
	"geekai/service"
	"geekai/store/model"
	"geekai/utils"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"gorm.io/gorm"
)

// Client MidJourney client
type Client struct {
	client         *req.Client
	licenseService *service.LicenseService
	db             *gorm.DB
}

type ImageRes struct {
	URLs          []string
	RevisedPrompt string
}

type nanoParams struct {
	AspectRatio string `json:"rate,omitempty"`
	Mode        string `json:"mode,omitempty"`
	Model       string `json:"model,omitempty"`
}

type bananaResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		RevisedPrompt string `json:"revised_prompt"`
		URL           string `json:"url"`
	} `json:"data"`
	Model string `json:"model"`
}

var logger = logger2.GetLogger()

func NewClient(licenseService *service.LicenseService, db *gorm.DB) *Client {
	return &Client{
		// 图片编辑操作可能需要较长时间，将超时时间设置为 3 分钟
		client:         req.C().SetTimeout(time.Minute * 5).SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"),
		licenseService: licenseService,
		db:             db,
	}
}

func parseNanoParams(raw string) nanoParams {
	if raw == "" {
		return nanoParams{}
	}
	var params nanoParams
	_ = json.Unmarshal([]byte(raw), &params)
	return params
}

func (c *Client) Imagine(task types.MjTask) (ImageRes, error) {
	params := parseNanoParams(task.Params)
	if params.Model == "" {
		params.Model = "nano-banana"
	}

	body := map[string]interface{}{
		"prompt":          task.Prompt,
		"model":           params.Model,
		"response_format": "url",
	}
	if params.AspectRatio != "" {
		body["aspect_ratio"] = params.AspectRatio
	}
	if task.NegPrompt != "" {
		body["negative_prompt"] = task.NegPrompt
	}

	resBody := bananaResponse{}
	// 图生图模式
	if len(task.ImgArr) > 0 {
		// 下载所有图片
		imageDataList := make([][]byte, 0, len(task.ImgArr))
		imageNames := make([]string, 0, len(task.ImgArr))
		for i, imgURL := range task.ImgArr {
			imageData, err := utils.DownloadImage(imgURL, "")
			if err != nil {
				return ImageRes{}, fmt.Errorf("下载参考图 %d 失败: %v", i+1, err)
			}
			imageDataList = append(imageDataList, imageData)
			// 从URL中提取文件名，如果没有则使用默认名称
			fileName := filepath.Base(imgURL)
			if fileName == "." || fileName == "/" {
				fileName = fmt.Sprintf("image_%d.png", i+1)
			}
			imageNames = append(imageNames, fileName)
		}

		// 构建 multipart form data
		var bodyBuffer bytes.Buffer
		writer := multipart.NewWriter(&bodyBuffer)

		// 添加表单字段
		_ = writer.WriteField("model", params.Model)
		_ = writer.WriteField("prompt", task.Prompt)
		_ = writer.WriteField("response_format", "url")
		if params.AspectRatio != "" {
			_ = writer.WriteField("aspect_ratio", params.AspectRatio)
		}
		if task.NegPrompt != "" {
			_ = writer.WriteField("negative_prompt", task.NegPrompt)
		}

		// 添加所有图片文件（使用相同的字段名 "image"）
		for i, imageData := range imageDataList {
			part, err := writer.CreateFormFile("image", imageNames[i])
			if err != nil {
				writer.Close()
				return ImageRes{}, fmt.Errorf("创建文件字段失败: %v", err)
			}
			_, err = io.Copy(part, bytes.NewReader(imageData))
			if err != nil {
				writer.Close()
				return ImageRes{}, fmt.Errorf("写入图片数据失败: %v", err)
			}
		}

		// 获取 content type（在关闭 writer 之前）
		contentType := writer.FormDataContentType()

		// 关闭 writer 以完成 multipart 数据
		err := writer.Close()
		if err != nil {
			return ImageRes{}, fmt.Errorf("关闭 multipart writer 失败: %v", err)
		}

		// 发送请求
		err = c.callAPI("v1/images/edits", task.ChannelId, func(r *req.Request) *req.Request {
			return r.SetHeader("Content-Type", contentType).
				SetBody(bodyBuffer.Bytes())
		}, &resBody)
		if err != nil {
			return ImageRes{}, err
		}
		return convertBananaResponse(resBody), nil
	}

	// 文生图模式
	err := c.callAPI("v1/images/generations", task.ChannelId, func(r *req.Request) *req.Request {
		return r.SetBody(body)
	}, &resBody)
	if err != nil {
		return ImageRes{}, err
	}
	return convertBananaResponse(resBody), nil
}

func (c *Client) Blend(task types.MjTask) (ImageRes, error) {
	return ImageRes{}, errors.New("当前模型暂不支持融图功能")
}

func (c *Client) SwapFace(task types.MjTask) (ImageRes, error) {
	return ImageRes{}, errors.New("当前模型暂不支持换脸功能")
}

func (c *Client) Upscale(task types.MjTask) (ImageRes, error) {
	return ImageRes{}, errors.New("当前模型暂不支持放大功能")
}

func (c *Client) Variation(task types.MjTask) (ImageRes, error) {
	return ImageRes{}, errors.New("当前模型暂不支持变换功能")
}

func (c *Client) callAPI(apiPath string, channel string, build func(r *req.Request) *req.Request, result interface{}) error {
	apiKey, err := c.getApiKey(channel)
	if err != nil {
		return err
	}

	if err = c.licenseService.IsValidApiURL(apiKey.ApiURL); err != nil {
		return err
	}

	base := strings.TrimSuffix(apiKey.ApiURL, "/")
	path := strings.TrimPrefix(apiPath, "/")
	apiURL := fmt.Sprintf("%s/%s", base, path)

	reqClient := c.client.R().
		SetHeader("Authorization", "Bearer "+apiKey.Value).
		SetSuccessResult(result)
	if build != nil {
		reqClient = build(reqClient)
	}

	resp, err := reqClient.Post(apiURL)
	if err != nil {
		return fmt.Errorf("请求 API 出错：%v", err)
	}

	if resp.IsErrorState() {
		errMsg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API 返回错误：%s", string(errMsg))
	}

	if err = c.db.Model(&apiKey).Update("last_used_at", time.Now().Unix()).Error; err != nil {
		logger.Error("update api key last used time error: ", err)
	}

	return nil
}

func (c *Client) getApiKey(channel string) (model.ApiKey, error) {
	session := c.db.Session(&gorm.Session{}).Where("type", "mj").Where("enabled", true)
	if channel != "" {
		session = session.Where("api_url", channel)
	}
	var apiKey model.ApiKey
	err := session.Order("last_used_at ASC").First(&apiKey).Error
	if err != nil {
		return model.ApiKey{}, fmt.Errorf("no available MidJourney api key: %v", err)
	}
	return apiKey, nil
}

func convertBananaResponse(res bananaResponse) ImageRes {
	imageRes := ImageRes{
		URLs: make([]string, 0),
	}
	for _, item := range res.Data {
		if item.URL != "" {
			imageRes.URLs = append(imageRes.URLs, item.URL)
		}
		if item.RevisedPrompt != "" && imageRes.RevisedPrompt == "" {
			imageRes.RevisedPrompt = item.RevisedPrompt
		}
	}
	return imageRes
}
