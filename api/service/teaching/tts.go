package teaching

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const DefaultTTSURL = "https://openspeech.bytedance.com/api/v3/tts/unidirectional"

type ttsErrEnvelope struct {
	Header struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Reqid   string `json:"reqid"`
	} `json:"header"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func formatTTSHTTPError(status int, body []byte) error {
	raw := strings.TrimSpace(string(body))
	var env ttsErrEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return fmt.Errorf("字节TTS HTTP %d：%s", status, raw)
	}
	if env.Header.Message != "" {
		hint := ttsGrantHint(env.Header.Code)
		return fmt.Errorf("字节TTS HTTP %d（业务码 %d）：%s。%s", status, env.Header.Code, env.Header.Message, hint)
	}
	if env.Message != "" && env.Code != 0 {
		return fmt.Errorf("字节TTS HTTP %d（业务码 %d）：%s", status, env.Code, env.Message)
	}
	return fmt.Errorf("字节TTS HTTP %d：%s", status, raw)
}

// ttsGrantHint 针对常见火山/字节语音 SaaS 错误补充排查说明（以官方控制台为准）。
func ttsGrantHint(code int) string {
	switch code {
	case 45000010:
		return "请核对 X-Api-Key / X-Api-Resource-Id（或旧版 AppId+AccessKey+ResourceId）与控制台授权一致。"
	default:
		return "请核对系统配置里 TTS 的 X-Api-Key、ResourceId、请求地址与火山/字节语音控制台是否一致。"
	}
}

// ttsStreamChunk 流式单行 JSON：成功帧常见为 {"code":0,"message":"","data":"<base64>"}；结束帧可能含 code=20000000。
type ttsStreamChunk struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
	Header  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"header"`
}

// SynthesizeMP3 字节 OpenSpeech TTS v3 HTTP 单向流式。
// 鉴权二选一：① singleApiKey 非空时仅发送 X-Api-Key + X-Api-Resource-Id（火山 seed-tts-2.0 等）；② 否则发送 X-Api-App-Id + X-Api-Access-Key + X-Api-Resource-Id。
func SynthesizeMP3(client *http.Client, apiURL, appID, accessKey, resourceID, text, speaker, format string, sampleRate int, uid, singleApiKey string) ([]byte, error) {
	if client == nil {
		client = http.DefaultClient
	}
	if apiURL == "" {
		apiURL = DefaultTTSURL
	}
	if format == "" {
		format = "mp3"
	}
	if sampleRate <= 0 {
		sampleRate = 24000
	}
	if uid == "" {
		uid = "soai-teaching"
	}
	if strings.TrimSpace(resourceID) == "" {
		return nil, fmt.Errorf("tts: X-Api-Resource-Id 不能为空")
	}
	additions, _ := json.Marshal(map[string]any{
		"explicit_language":       "zh",
		"disable_markdown_filter": true,
		"enable_timestamp":        true,
	})
	payload := map[string]any{
		"user": map[string]string{"uid": uid},
		"req_params": map[string]any{
			"text":    text,
			"speaker": speaker,
			"audio_params": map[string]any{
				"format":           format,
				"sample_rate":      sampleRate,
				"enable_timestamp": true,
			},
			"additions": string(additions),
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(singleApiKey) != "" {
		req.Header.Set("X-Api-Key", strings.TrimSpace(singleApiKey))
	} else {
		if strings.TrimSpace(appID) == "" || strings.TrimSpace(accessKey) == "" {
			return nil, fmt.Errorf("tts: 请配置 tts_api_key（X-Api-Key），或同时配置 tts_app_id 与 tts_access_key")
		}
		req.Header.Set("X-Api-App-Id", strings.TrimSpace(appID))
		req.Header.Set("X-Api-Access-Key", strings.TrimSpace(accessKey))
	}
	req.Header.Set("X-Api-Resource-Id", strings.TrimSpace(resourceID))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, formatTTSHTTPError(resp.StatusCode, b)
	}

	var audio bytes.Buffer
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var chunk ttsStreamChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			continue
		}
		if chunk.Header != nil && chunk.Header.Code != 0 {
			return nil, fmt.Errorf("tts stream header code=%d: %s", chunk.Header.Code, chunk.Header.Message)
		}
		if chunk.Code == 0 && chunk.Data != "" {
			raw, err := base64.StdEncoding.DecodeString(chunk.Data)
			if err != nil {
				continue
			}
			audio.Write(raw)
			continue
		}
		if chunk.Code == 20000000 {
			break
		}
		if chunk.Code > 0 {
			msg := chunk.Message
			if msg == "" {
				msg = strings.TrimSpace(string(line))
			}
			return nil, fmt.Errorf("tts stream code=%d: %s", chunk.Code, msg)
		}
	}
	if audio.Len() == 0 {
		return nil, fmt.Errorf("tts: empty audio stream")
	}
	return audio.Bytes(), nil
}
