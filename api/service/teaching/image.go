package teaching

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type imageGenReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Size   string `json:"size"`
}

type imageGenResp struct {
	Data []struct {
		URL           string `json:"url"`
		B64JSON       string `json:"b64_json"`
		RevisedPrompt string `json:"revised_prompt"`
	} `json:"data"`
	Model string `json:"model"`
}

// GenerateImage OpenAI 兼容 POST {base}/v1/images/generations
func GenerateImage(client *http.Client, baseURL, bearerToken, model, prompt, size string) (remoteURL string, b64 string, revised string, err error) {
	if client == nil {
		client = &http.Client{Timeout: 300 * time.Second}
	}
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return "", "", "", fmt.Errorf("empty image api base url")
	}
	endpoint := base + "/v1/images/generations"
	if model == "" {
		model = "gpt-image-2"
	}
	if size == "" {
		size = "1024x1024"
	}
	body, err := json.Marshal(imageGenReq{Model: model, Prompt: prompt, Size: size})
	if err != nil {
		return "", "", "", err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(bearerToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("image api %d: %s", resp.StatusCode, string(raw))
	}
	var out imageGenResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", "", "", fmt.Errorf("decode image response: %w, body=%s", err, string(raw))
	}
	if len(out.Data) == 0 {
		return "", "", "", fmt.Errorf("image api: empty data[]")
	}
	first := out.Data[0]
	return first.URL, first.B64JSON, first.RevisedPrompt, nil
}
