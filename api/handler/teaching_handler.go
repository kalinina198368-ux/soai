package handler

import (
	"encoding/json"
	"fmt"
	"geekai/core"
	"geekai/core/types"
	"geekai/service"
	"geekai/service/oss"
	"geekai/service/teaching"
	"geekai/store/model"
	"geekai/utils"
	"geekai/utils/resp"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TeachingHandler struct {
	BaseHandler
	uploader    *oss.UploaderManager
	userService *service.UserService
}

func NewTeachingHandler(app *core.AppServer, db *gorm.DB, manager *oss.UploaderManager, userService *service.UserService) *TeachingHandler {
	return &TeachingHandler{
		BaseHandler: BaseHandler{App: app, DB: db},
		uploader:    manager,
		userService: userService,
	}
}

func stripMarkdownJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		if idx := strings.Index(s, "\n"); idx >= 0 {
			s = s[idx+1:]
		}
		s = strings.TrimSpace(s)
		if strings.HasSuffix(s, "```") {
			s = strings.TrimSuffix(s, "```")
		}
	}
	return strings.TrimSpace(s)
}

// teachingOutlineWithContent 只保留有口播正文的章节（用户可能删掉中间/末尾段或留空占位，避免模型输出截断、非法 JSON）
func teachingOutlineWithContent(outline []map[string]interface{}) []map[string]interface{} {
	if len(outline) == 0 {
		return nil
	}
	out := make([]map[string]interface{}, 0, len(outline))
	for _, seg := range outline {
		if seg == nil {
			continue
		}
		text := ""
		if v, ok := seg["text"]; ok && v != nil {
			text = strings.TrimSpace(fmt.Sprint(v))
		}
		if text == "" {
			continue
		}
		cp := make(map[string]interface{}, len(seg)+1)
		for k, v := range seg {
			cp[k] = v
		}
		out = append(out, cp)
	}
	return out
}

// extractJSONArray 从模型输出中截取最外层 JSON 数组（忽略数组前后的说明文字），用于轻微脏数据容错
func extractJSONArray(s string) ([]byte, bool) {
	s = strings.TrimSpace(s)
	i := strings.Index(s, "[")
	if i < 0 {
		return nil, false
	}
	depth := 0
	inStr := false
	esc := false
	for j := i; j < len(s); j++ {
		c := s[j]
		if inStr {
			if esc {
				esc = false
				continue
			}
			if c == '\\' {
				esc = true
				continue
			}
			if c == '"' {
				inStr = false
			}
			continue
		}
		if c == '"' {
			inStr = true
			continue
		}
		if c == '[' {
			depth++
		} else if c == ']' {
			depth--
			if depth == 0 {
				return []byte(s[i : j+1]), true
			}
		}
	}
	return nil, false
}

const teachingImagePromptsLlmBatchSize = 8

func parseTeachingImagePromptsResponse(raw string) ([]map[string]interface{}, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("模型返回内容为空")
	}
	raw = stripMarkdownJSON(raw)
	var arr []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		if sub, ok := extractJSONArray(raw); ok {
			if err2 := json.Unmarshal(sub, &arr); err2 != nil {
				return nil, err2
			}
		} else {
			return nil, err
		}
	}
	return arr, nil
}

func (h *TeachingHandler) imagePromptsLLMOneBatch(topic, system string, modelID int, batch []map[string]interface{}, batchIdx, batchTotal, expect int) ([]map[string]interface{}, error) {
	brief, _ := json.Marshal(batch)
	user := fmt.Sprintf(
		"主题：%s\n本次为第 %d/%d 批分镜（本批共 %d 条）。你输出的 JSON 数组必须恰好包含 %d 个对象，顺序与下列输入一致，不要合并或省略。\n口播分镜 JSON：%s",
		strings.TrimSpace(topic), batchIdx, batchTotal, expect, expect, string(brief),
	)
	msgs := []interface{}{
		types.Message{Role: "system", Content: system},
		types.Message{Role: "user", Content: user},
	}
	tryParse := func(content string) ([]map[string]interface{}, error) {
		return parseTeachingImagePromptsResponse(content)
	}
	raw, err := utils.SendTeachingChatCompletionLongJSON(h.DB, modelID, msgs, 6144)
	if err != nil {
		return nil, err
	}
	arr, err := tryParse(raw)
	if err == nil && len(arr) == expect {
		return arr, nil
	}
	retryUser := user + fmt.Sprintf("\n\n【纠正】上一条输出无效或条数不等于 %d。请只输出合法 JSON 数组，恰好 %d 个元素，不要任何其它文字。", expect, expect)
	msgs2 := []interface{}{
		types.Message{Role: "system", Content: system},
		types.Message{Role: "user", Content: retryUser},
	}
	raw2, err2 := utils.SendTeachingChatCompletionLongJSON(h.DB, modelID, msgs2, 6144)
	if err2 != nil {
		if err != nil {
			return nil, fmt.Errorf("第%d/%d 批：%v", batchIdx, batchTotal, err)
		}
		return nil, fmt.Errorf("第%d/%d 批：条数期望%d实际%d", batchIdx, batchTotal, expect, len(arr))
	}
	arr2, err3 := tryParse(raw2)
	if err3 != nil {
		return nil, fmt.Errorf("第%d/%d 批解析失败：%v", batchIdx, batchTotal, err3)
	}
	if len(arr2) != expect {
		return nil, fmt.Errorf("第%d/%d 批：期望输出 %d 条配图项，实际 %d 条（可能因模型输出被截断，可尝试调大模型 MaxTokens 或减小分镜密度）", batchIdx, batchTotal, expect, len(arr2))
	}
	return arr2, nil
}

func (h *TeachingHandler) saveMP3Local(mp3 []byte) (string, error) {
	basePath := h.App.Config.OSS.Local.BasePath
	baseURL := h.App.Config.OSS.Local.BaseURL
	if basePath == "" || baseURL == "" {
		return "", fmt.Errorf("未配置本地 OSS 路径，无法保存教学语音文件")
	}
	path, err := utils.GenUploadPath(basePath, "teaching.mp3", false)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, mp3, 0644); err != nil {
		return "", err
	}
	return utils.GenUploadUrl(basePath, baseURL, path), nil
}

// Script 根据主题生成口播脚本草稿（JSON）
func (h *TeachingHandler) Script(c *gin.Context) {
	var data struct {
		Topic string `json:"topic"`
	}
	if err := c.ShouldBindJSON(&data); err != nil || strings.TrimSpace(data.Topic) == "" {
		resp.ERROR(c, types.InvalidArgs)
		return
	}
	modelID := h.App.SysConfig.TeachingScriptModelId
	system := `你是教学口播撰稿助手。只输出一个合法 JSON 对象，不要 Markdown、不要代码围栏、不要其它说明文字。
JSON 结构严格为：
{"title":"string","outline":[{"id":"seg-1","title":"小节标题","text":"该小节完整口播正文"}]}
要求：outline 至少 3 条，id 唯一，text 为口语化中文，适合短视频口播。title 包含用户给的主题。`
	user := fmt.Sprintf("教学主题：%s", strings.TrimSpace(data.Topic))
	msgs := []interface{}{
		types.Message{Role: "system", Content: system},
		types.Message{Role: "user", Content: user},
	}
	raw, err := utils.SendTeachingChatCompletion(h.DB, modelID, msgs)
	if err != nil {
		resp.ERROR(c, err.Error())
		return
	}
	raw = stripMarkdownJSON(raw)
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		resp.ERROR(c, "模型返回非合法 JSON："+err.Error())
		return
	}
	var cm model.ChatModel
	_ = h.DB.Where("id", modelID).First(&cm).Error
	if cm.Power > 0 {
		uid := h.GetLoginUserId(c)
		_ = h.userService.DecreasePower(int(uid), cm.Power, model.PowerLog{
			Type:   types.PowerConsume,
			Model:  cm.Value,
			Remark: "教学：生成口播脚本",
		})
	}
	resp.SUCCESS(c, out)
}

type ttsSeg struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// TTS 按分段合成语音（字节 OpenSpeech），音频保存到本地 OSS 目录并返回可访问 URL。
func (h *TeachingHandler) TTS(c *gin.Context) {
	var data struct {
		Segments []ttsSeg `json:"segments"`
	}
	if err := c.ShouldBindJSON(&data); err != nil || len(data.Segments) == 0 {
		resp.ERROR(c, types.InvalidArgs)
		return
	}
	cfg := h.App.SysConfig
	hasKeyAuth := strings.TrimSpace(cfg.TtsApiKey) != "" && strings.TrimSpace(cfg.TtsResourceId) != ""
	hasLegacyAuth := strings.TrimSpace(cfg.TtsAppId) != "" && strings.TrimSpace(cfg.TtsAccessKey) != "" && strings.TrimSpace(cfg.TtsResourceId) != ""
	if !hasKeyAuth && !hasLegacyAuth {
		resp.ERROR(c, "请配置 TTS：填写 tts_api_key + tts_resource_id（X-Api-Key 方式），或填写 tts_app_id + tts_access_key + tts_resource_id（旧方式）")
		return
	}
	if strings.TrimSpace(cfg.TtsSpeaker) == "" {
		resp.ERROR(c, "请在系统配置中填写 TTS 发音人 tts_speaker")
		return
	}
	client := &http.Client{Timeout: 120 * time.Second}
	type item struct {
		SegmentId   string `json:"segmentId"`
		FileName    string `json:"fileName"`
		URL         string `json:"url"`
		DurationSec int    `json:"durationSec"`
	}
	var items []item
	uid := fmt.Sprintf("u%d", h.GetLoginUserId(c))
	for _, seg := range data.Segments {
		if strings.TrimSpace(seg.Text) == "" {
			continue
		}
		mp3, err := teaching.SynthesizeMP3(client, cfg.TtsApiURL, cfg.TtsAppId, cfg.TtsAccessKey, cfg.TtsResourceId,
			seg.Text, cfg.TtsSpeaker, cfg.TtsAudioFormat, cfg.TtsSampleRate, uid, cfg.TtsApiKey)
		if err != nil {
			resp.ERROR(c, err.Error())
			return
		}
		url, err := h.saveMP3Local(mp3)
		if err != nil {
			resp.ERROR(c, err.Error())
			return
		}
		id := seg.ID
		if id == "" {
			id = fmt.Sprintf("seg-%d", len(items)+1)
		}
		items = append(items, item{
			SegmentId: id,
			FileName:  fmt.Sprintf("%s.mp3", id),
			URL:       url,
		})
	}
	if len(items) == 0 {
		resp.ERROR(c, "没有可合成的段落：请先填写至少一节口播正文")
		return
	}
	resp.SUCCESS(c, gin.H{"items": items})
}

// ImagePrompts 根据口播大纲生成各分镜配图提示词（JSON 数组），语言与口播正文一致；默认按约 5 秒口播密度拆条并统一画风锚点。
// 颗粒度优先读系统配置（管理后台「教学」页）；请求体中带同名字段时可单次覆盖。
func (h *TeachingHandler) ImagePrompts(c *gin.Context) {
	var data struct {
		Topic                    string                   `json:"topic"`
		Outline                  []map[string]interface{} `json:"outline"`
		SparseVisualShots        *bool                    `json:"sparse_visual_shots,omitempty"`
		VisualShotMaxRunes       *int                     `json:"visual_shot_max_runes,omitempty"`
		VisualMaxTotalShots      *int                     `json:"visual_max_total_shots,omitempty"`
		ImagePromptsLlmBatchSize *int                     `json:"image_prompts_llm_batch_size,omitempty"`
	}
	if err := c.ShouldBindJSON(&data); err != nil || len(data.Outline) == 0 {
		resp.ERROR(c, types.InvalidArgs)
		return
	}
	outline := teachingOutlineWithContent(data.Outline)
	if len(outline) == 0 {
		resp.ERROR(c, "请至少保留一节有口播正文的段落后再生成配图提示词")
		return
	}
	cfg := h.App.SysConfig
	maxRunes := cfg.TeachingVisualShotMaxRunes
	if maxRunes <= 0 {
		maxRunes = 44
	}
	if data.VisualShotMaxRunes != nil && *data.VisualShotMaxRunes > 0 {
		maxRunes = *data.VisualShotMaxRunes
	}

	sparse := cfg.TeachingVisualSparse
	if data.SparseVisualShots != nil {
		sparse = *data.SparseVisualShots
	}

	maxTotalShots := cfg.TeachingVisualMaxTotalShots
	if maxTotalShots <= 0 {
		maxTotalShots = 40
	}
	if data.VisualMaxTotalShots != nil && *data.VisualMaxTotalShots > 0 {
		maxTotalShots = *data.VisualMaxTotalShots
	}

	workOutline := outline
	if !sparse {
		workOutline = expandOutlineVisualShots(outline, maxRunes, maxTotalShots)
	}
	if len(workOutline) == 0 {
		resp.ERROR(c, "请至少保留一节有口播正文的段落后再生成配图提示词")
		return
	}
	batchSize := cfg.TeachingImagePromptsLlmBatch
	if batchSize <= 0 {
		batchSize = teachingImagePromptsLlmBatchSize
	}
	if data.ImagePromptsLlmBatchSize != nil && *data.ImagePromptsLlmBatchSize > 0 {
		batchSize = *data.ImagePromptsLlmBatchSize
	}
	if batchSize < 4 {
		batchSize = 4
	}
	if batchSize > 12 {
		batchSize = 12
	}
	modelID := h.App.SysConfig.TeachingScriptModelId
	system := `你是视频分镜配图提示词助手。只输出合法 JSON 数组，不要 Markdown、不要代码围栏、不要其它说明文字。
数组每项结构：{"segmentId":"与输入该分镜 id 完全一致（含 seg-1__2 这类子分镜）","title":"画面标题","prompt":"生图用的画面描述正文（不要重复写统一画风句，只写本分镜独有画面）","aspectRatio":"16:9"}
要求：
1) 输出条数必须与用户说明的本批分镜条数完全一致，顺序一致；不要合并或省略。
2) segmentId 必须与输入 JSON 里该条的 id 字段一致。
3) 同一视频内人物身份、服饰主色、时代场景基调在各分镜之间必须一致；prompt 只描述本分镜的构图与动作变化，不要每帧换主角性别或换画风。
4) title 与 prompt 的语言必须与对应分镜口播正文一致：中文口播用简练简体中文；不要无故整段英文。
5) prompt 写清主体、环境、光线与构图；不出现可识别真人脸与真实名人姓名。
6) 各分镜 prompt 在画面内容上要有明显推进（时间/动作/景别变化），避免多张图看起来是同一张。`

	n := len(workOutline)
	totalBatches := (n + batchSize - 1) / batchSize
	var arr []map[string]interface{}
	for b := 0; b < totalBatches; b++ {
		start := b * batchSize
		end := start + batchSize
		if end > n {
			end = n
		}
		batch := workOutline[start:end]
		expect := len(batch)
		part, err := h.imagePromptsLLMOneBatch(strings.TrimSpace(data.Topic), system, modelID, batch, b+1, totalBatches, expect)
		if err != nil {
			resp.ERROR(c, err.Error())
			return
		}
		arr = append(arr, part...)
	}

	anchor := teachingVisualAnchor(data.Topic)
	for i := range arr {
		p, _ := arr[i]["prompt"].(string)
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "【统一画风") {
			arr[i]["prompt"] = anchor + " " + p
		}
	}
	var cm model.ChatModel
	_ = h.DB.Where("id", modelID).First(&cm).Error
	if cm.Power > 0 {
		uid := h.GetLoginUserId(c)
		_ = h.userService.DecreasePower(int(uid), cm.Power, model.PowerLog{
			Type:   types.PowerConsume,
			Model:  cm.Value,
			Remark: "教学：生图提示词",
		})
	}
	resp.SUCCESS(c, gin.H{"prompts": arr})
}

type imgPromptRow struct {
	SegmentId string `json:"segmentId"`
	Prompt    string `json:"prompt"`
}

// RenderImages 调用配置的 OpenAI 兼容 images/generations，并把结果转存到当前 OSS。
func (h *TeachingHandler) RenderImages(c *gin.Context) {
	var data struct {
		Prompts []imgPromptRow `json:"prompts"`
	}
	if err := c.ShouldBindJSON(&data); err != nil || len(data.Prompts) == 0 {
		resp.ERROR(c, types.InvalidArgs)
		return
	}
	cfg := h.App.SysConfig
	if strings.TrimSpace(cfg.TeachingImageApiURL) == "" || strings.TrimSpace(cfg.TeachingImageApiKey) == "" {
		resp.ERROR(c, "请在系统配置中填写教学配图 API 地址与 Key")
		return
	}
	if len(data.Prompts) > 48 {
		resp.ERROR(c, "单次最多 48 张")
		return
	}
	client := &http.Client{Timeout: 300 * time.Second}
	type asset struct {
		SegmentId string `json:"segmentId"`
		ImageURL  string `json:"imageUrl"`
		RemoteURL string `json:"remoteUrl,omitempty"`
		Revised   string `json:"revisedPrompt,omitempty"`
	}
	var assets []asset
	for _, row := range data.Prompts {
		if strings.TrimSpace(row.Prompt) == "" {
			continue
		}
		remote, b64, revised, err := teaching.GenerateImage(client, cfg.TeachingImageApiURL, cfg.TeachingImageApiKey,
			cfg.TeachingImageModel, row.Prompt, cfg.TeachingImageSize)
		if err != nil {
			resp.ERROR(c, err.Error())
			return
		}
		var final string
		if remote != "" {
			final, err = h.uploader.GetUploadHandler().PutUrlFile(remote, false)
			if err != nil {
				resp.ERROR(c, "转存配图失败："+err.Error())
				return
			}
		} else if b64 != "" {
			final, err = h.uploader.GetUploadHandler().PutBase64(b64)
			if err != nil {
				resp.ERROR(c, "上传配图失败："+err.Error())
				return
			}
		} else {
			resp.ERROR(c, "配图接口未返回 url 或 b64_json")
			return
		}
		assets = append(assets, asset{
			SegmentId: row.SegmentId,
			ImageURL:  final,
			RemoteURL: remote,
			Revised:   revised,
		})
	}
	if h.App.SysConfig.PromptPower > 0 {
		uid := h.GetLoginUserId(c)
		cost := h.App.SysConfig.PromptPower * len(assets)
		_ = h.userService.DecreasePower(int(uid), cost, model.PowerLog{
			Type:   types.PowerConsume,
			Model:  cfg.TeachingImageModel,
			Remark: fmt.Sprintf("教学：配图生成 x%d", len(assets)),
		})
	}
	resp.SUCCESS(c, gin.H{"assets": assets})
}

// RecordList 当前用户教学记录列表
func (h *TeachingHandler) RecordList(c *gin.Context) {
	uid := h.GetLoginUserId(c)
	if uid == 0 {
		resp.NotAuth(c)
		return
	}
	var items []model.TeachingRecord
	h.DB.Where("user_id", uid).Order("updated_at DESC").Limit(50).Find(&items)
	list := make([]gin.H, 0, len(items))
	for _, it := range items {
		list = append(list, gin.H{
			"id":            it.Id,
			"topic":         it.Topic,
			"title_summary": it.TitleSummary,
			"updated_at":    it.UpdatedAt.Unix(),
			"max_step":      it.MaxStep,
		})
	}
	resp.SUCCESS(c, gin.H{"list": list})
}

// decodeJSONColumn 将库中 JSON 列还原为对象/数组；若曾被存成 JSON 字符串（二次编码），会逐层解开。
func decodeJSONColumn(s string) interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	raw := []byte(s)
	for i := 0; i < 6; i++ {
		var v interface{}
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil
		}
		str, ok := v.(string)
		if !ok {
			return v
		}
		t := strings.TrimSpace(str)
		if t == "" {
			return nil
		}
		if t[0] != '[' && t[0] != '{' {
			return str
		}
		raw = []byte(t)
	}
	var last interface{}
	_ = json.Unmarshal(raw, &last)
	return last
}

// normalizeJSONStorage 把请求里的 RawMessage 规范成写入库的 JSON 文本（避免把数组/对象再包一层 JSON 字符串）。
func normalizeJSONStorage(raw json.RawMessage, empty string) string {
	if len(raw) == 0 {
		return empty
	}
	s := strings.TrimSpace(string(raw))
	if s == "" {
		return empty
	}
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}
	if str, ok := v.(string); ok {
		t := strings.TrimSpace(str)
		if t == "" {
			return empty
		}
		if err := json.Unmarshal([]byte(t), &v); err != nil {
			return t
		}
	}
	out, err := json.Marshal(v)
	if err != nil {
		return empty
	}
	return string(out)
}

// RecordGet 单条记录（含各阶段 JSON）
func (h *TeachingHandler) RecordGet(c *gin.Context) {
	uid := h.GetLoginUserId(c)
	if uid == 0 {
		resp.NotAuth(c)
		return
	}
	id := uint(utils.IntValue(c.Param("id"), 0))
	if id == 0 {
		resp.ERROR(c, types.InvalidArgs)
		return
	}
	var rec model.TeachingRecord
	if err := h.DB.Where("id", id).Where("user_id", uid).First(&rec).Error; err != nil {
		resp.ERROR(c, "记录不存在")
		return
	}
	resp.SUCCESS(c, gin.H{
		"id":            rec.Id,
		"topic":         rec.Topic,
		"title_summary": rec.TitleSummary,
		"max_step":      rec.MaxStep,
		"updated_at":    rec.UpdatedAt.Unix(),
		"script":        decodeJSONColumn(rec.ScriptJSON),
		"tts_items":     decodeJSONColumn(rec.TtsJSON),
		"image_prompts": decodeJSONColumn(rec.PromptsJSON),
		"assets":        decodeJSONColumn(rec.AssetsJSON),
	})
}

type teachingRecordSaveBody struct {
	ID           uint            `json:"id"`
	Topic        string          `json:"topic"`
	Script       json.RawMessage `json:"script"`
	TtsItems     json.RawMessage `json:"tts_items"`
	ImagePrompts json.RawMessage `json:"image_prompts"`
	Assets       json.RawMessage `json:"assets"`
	MaxStep      int             `json:"max_step"`
}

// RecordSave 新建或更新教学记录
func (h *TeachingHandler) RecordSave(c *gin.Context) {
	uid := h.GetLoginUserId(c)
	if uid == 0 {
		resp.NotAuth(c)
		return
	}
	var body teachingRecordSaveBody
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.ERROR(c, types.InvalidArgs)
		return
	}
	scriptStr := normalizeJSONStorage(body.Script, "{}")
	ttsStr := normalizeJSONStorage(body.TtsItems, "[]")
	promptsStr := normalizeJSONStorage(body.ImagePrompts, "[]")
	assetsStr := normalizeJSONStorage(body.Assets, "[]")

	titleSum := strings.TrimSpace(body.Topic)
	if scriptStr != "" && scriptStr != "{}" {
		var m map[string]interface{}
		if json.Unmarshal([]byte(scriptStr), &m) == nil {
			if t, ok := m["title"].(string); ok {
				t = strings.TrimSpace(t)
				if t != "" {
					r := []rune(t)
					if len(r) > 120 {
						titleSum = string(r[:120]) + "..."
					} else {
						titleSum = t
					}
				}
			}
		}
	}
	if titleSum == "" {
		titleSum = strings.TrimSpace(body.Topic)
	}
	maxStep := body.MaxStep
	if maxStep < 0 {
		maxStep = 0
	}
	if maxStep > 3 {
		maxStep = 3
	}

	if body.ID > 0 {
		var old model.TeachingRecord
		if err := h.DB.Where("id", body.ID).Where("user_id", uid).First(&old).Error; err != nil {
			resp.ERROR(c, "记录不存在")
			return
		}
		updates := map[string]interface{}{
			"topic":         strings.TrimSpace(body.Topic),
			"title_summary": titleSum,
			"script_json":   scriptStr,
			"tts_json":      ttsStr,
			"prompts_json":  promptsStr,
			"assets_json":   assetsStr,
			"max_step":      maxStep,
			"updated_at":    time.Now(),
		}
		if err := h.DB.Model(&model.TeachingRecord{}).Where("id", body.ID).Where("user_id", uid).Updates(updates).Error; err != nil {
			resp.ERROR(c, err.Error())
			return
		}
		resp.SUCCESS(c, gin.H{"id": body.ID})
		return
	}

	if strings.TrimSpace(body.Topic) == "" {
		resp.ERROR(c, "新建记录需要填写主题")
		return
	}

	rec := model.TeachingRecord{
		UserId:       uid,
		Topic:        strings.TrimSpace(body.Topic),
		TitleSummary: titleSum,
		ScriptJSON:   scriptStr,
		TtsJSON:      ttsStr,
		PromptsJSON:  promptsStr,
		AssetsJSON:   assetsStr,
		MaxStep:      maxStep,
	}
	if err := h.DB.Create(&rec).Error; err != nil {
		resp.ERROR(c, err.Error())
		return
	}
	resp.SUCCESS(c, gin.H{"id": rec.Id})
}
