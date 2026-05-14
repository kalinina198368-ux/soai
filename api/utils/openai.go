package utils

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
	"geekai/store/model"
	"io"
	"time"

	"github.com/imroc/req/v3"
	"github.com/pkoukk/tiktoken-go"
	"gorm.io/gorm"
)

func CalcTokens(text string, model string) (int, error) {
	encoding, ok := tiktoken.MODEL_TO_ENCODING[model]
	if !ok {
		encoding = "cl100k_base"
	}
	tke, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return 0, fmt.Errorf("getEncoding: %v", err)
	}

	token := tke.Encode(text, nil, nil)
	return len(token), nil
}

type OpenAIResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func OpenAIRequest(db *gorm.DB, prompt string, modelId int) (string, error) {
	messages := make([]interface{}, 1)
	messages[0] = types.Message{
		Role:    "user",
		Content: prompt,
	}
	return SendOpenAIMessage(db, messages, modelId)
}

func SendOpenAIMessage(db *gorm.DB, messages []interface{}, modelId int) (string, error) {
	var chatModel model.ChatModel
	db.Where("id", modelId).First(&chatModel)
	if chatModel.Value == "" {
		chatModel.Value = "gpt-4o" // 默认使用 gpt-4o
	}
	var apiKey model.ApiKey
	if chatModel.KeyId > 0 {
		_ = db.Where("id", chatModel.KeyId).First(&apiKey).Error
	}
	if apiKey.Id == 0 {
		err := db.Session(&gorm.Session{}).Where("type", "chat").Where("enabled", true).Order("last_used_at ASC").First(&apiKey).Error
		if err != nil {
			return "", fmt.Errorf("error with fetch OpenAI API KEY：%v", err)
		}
	}

	var response OpenAIResponse
	client := req.C()
	if len(apiKey.ProxyURL) > 5 {
		client.SetProxyURL(apiKey.ApiURL)
	}
	apiURL := fmt.Sprintf("%s/v1/chat/completions", apiKey.ApiURL)
	logger.Infof("Sending %s request, API KEY:%s, PROXY: %s, Model: %s", apiKey.ApiURL, apiURL, apiKey.ProxyURL, chatModel.Name)
	r, err := client.R().SetHeader("Body-Type", "application/json").
		SetHeader("Authorization", "Bearer "+apiKey.Value).
		SetBody(types.ApiRequest{
			Model:       chatModel.Value,
			Temperature: 0.9,
			MaxTokens:   1024,
			Stream:      false,
			Messages:    messages,
		}).Post(apiURL)
	if err != nil {
		return "", fmt.Errorf("请求 OpenAI API失败：%v", err)
	}

	if r.IsErrorState() {
		return "", fmt.Errorf("请求 OpenAI API失败：%v", r.Status)
	}

	body, _ := io.ReadAll(r.Body)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("解析API数据失败：%v, %s", err, string(body))
	}

	// 更新 API KEY 的最后使用时间
	db.Model(&apiKey).UpdateColumn("last_used_at", time.Now().Unix())

	return response.Choices[0].Message.Content, nil
}

// sendTeachingChatCompletionOnce 教学文案模型非流式对话；minResponseTokens>0 时 max_tokens 取 max(模型表, minResponseTokens)，且不超过 16384。
func sendTeachingChatCompletionOnce(db *gorm.DB, modelId int, messages []interface{}, minResponseTokens int) (string, error) {
	if modelId <= 0 {
		return "", fmt.Errorf("未配置教学文案模型，请在系统配置中设置 teaching_script_model_id")
	}
	var chatModel model.ChatModel
	if err := db.Where("id", modelId).First(&chatModel).Error; err != nil {
		return "", fmt.Errorf("教学模型 id=%d 不存在或未启用: %w", modelId, err)
	}
	if chatModel.Value == "" {
		chatModel.Value = "gpt-4o"
	}
	var apiKey model.ApiKey
	if chatModel.KeyId > 0 {
		_ = db.Where("id", chatModel.KeyId).First(&apiKey).Error
	}
	if apiKey.Id == 0 {
		err := db.Where("type", "chat").Where("enabled", true).Order("last_used_at ASC").First(&apiKey).Error
		if err != nil {
			return "", fmt.Errorf("获取 OpenAI API KEY 失败：%v", err)
		}
	}

	temp := float32(0.7)
	if chatModel.Temperature > 0 {
		temp = chatModel.Temperature
	}
	maxTok := chatModel.MaxTokens
	if maxTok <= 0 {
		maxTok = 4096
	}
	if minResponseTokens > 0 && maxTok < minResponseTokens {
		maxTok = minResponseTokens
	}
	if maxTok > 16384 {
		maxTok = 16384
	}

	var response OpenAIResponse
	client := req.C()
	if len(apiKey.ProxyURL) > 5 {
		client.SetProxyURL(apiKey.ProxyURL)
	}
	apiURL := fmt.Sprintf("%s/v1/chat/completions", apiKey.ApiURL)
	logger.Infof("Teaching chat model=%s api=%s max_tokens=%d", chatModel.Value, apiURL, maxTok)
	r, err := client.R().SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+apiKey.Value).
		SetBody(types.ApiRequest{
			Model:       chatModel.Value,
			Temperature: temp,
			MaxTokens:   maxTok,
			Stream:      false,
			Messages:    messages,
		}).Post(apiURL)
	if err != nil {
		return "", fmt.Errorf("请求 OpenAI API失败：%v", err)
	}
	if r.IsErrorState() {
		body, _ := io.ReadAll(r.Body)
		return "", fmt.Errorf("请求 OpenAI API失败：%v, %s", r.Status, string(body))
	}
	body, _ := io.ReadAll(r.Body)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("解析API数据失败：%v, %s", err, string(body))
	}
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("模型返回空 choices：%s", string(body))
	}
	db.Model(&apiKey).UpdateColumn("last_used_at", time.Now().Unix())
	return response.Choices[0].Message.Content, nil
}

// SendTeachingChatCompletion 使用后台「语言模型」配置发起非流式对话；温度与 max_tokens 取自模型表。
func SendTeachingChatCompletion(db *gorm.DB, modelId int, messages []interface{}) (string, error) {
	return sendTeachingChatCompletionOnce(db, modelId, messages, 0)
}

// SendTeachingChatCompletionLongJSON 用于教学生图提示词等长 JSON 输出，保证响应 max_tokens 至少为 minResponseTokens（避免截断导致非法 JSON）。
func SendTeachingChatCompletionLongJSON(db *gorm.DB, modelId int, messages []interface{}, minResponseTokens int) (string, error) {
	if minResponseTokens < 4096 {
		minResponseTokens = 4096
	}
	return sendTeachingChatCompletionOnce(db, modelId, messages, minResponseTokens)
}
