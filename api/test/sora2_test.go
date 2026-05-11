package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Sora2GenerateRequest Sora2视频生成请求
type Sora2GenerateRequest struct {
	Prompt         string `json:"prompt"`
	Duration       string `json:"duration"`
	AspectRatio    string `json:"aspect_ratio"`
	Quality        string `json:"quality"`
	Style          string `json:"style"`
	NegativePrompt string `json:"negative_prompt"`
	Seed           int64  `json:"seed"`
	Steps          int    `json:"steps"`
	Model          string `json:"model"`
}

func main() {
	// 测试Sora2视频生成API
	url := "http://localhost:8080/api/sora2/generate"

	requestData := Sora2GenerateRequest{
		Prompt:         "一只可爱的小猫在花园里玩耍，阳光明媚，画面温馨，慢镜头拍摄",
		Duration:       "10",
		AspectRatio:    "16:9",
		Quality:        "hd",
		Style:          "realistic",
		NegativePrompt: "模糊，低质量",
		Seed:           12345,
		Steps:          30,
		Model:          "sora-2",
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		fmt.Printf("JSON编码错误: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("创建请求错误: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer ") //

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求错误: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应错误: %v\n", err)
		return
	}

	fmt.Printf("响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))
}
