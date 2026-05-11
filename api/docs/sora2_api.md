# Sora2 视频生成功能

## 功能概述

Sora2 视频生成功能基于 OpenAI 的 Sora2 API，支持通过文本提示词生成高质量的视频内容。

## 主要特性

- 🎬 **文生视频**: 通过文本描述生成视频
- ⚙️ **参数控制**: 支持时长、比例、质量、风格等参数调节
- 🎨 **多种风格**: 支持写实、动画、艺术、电影等多种风格
- 📱 **移动端优化**: 专为移动端设计的用户界面
- 💾 **历史记录**: 自动保存生成历史
- ⭐ **收藏功能**: 收藏喜欢的视频
- 📊 **统计分析**: 观看、点赞、分享数据统计

## API 接口

### 1. 生成视频

**POST** `/api/sora2/generate`

**请求参数:**
```json
{
  "prompt": "一只可爱的小猫在花园里玩耍，阳光明媚，画面温馨，慢镜头拍摄",
  "duration": "10",
  "aspect_ratio": "16:9",
  "quality": "hd",
  "style": "realistic",
  "negative_prompt": "模糊，低质量",
  "seed": 12345,
  "steps": 30,
  "model": "sora-2"
}
```

**参数说明:**
- `prompt` (必需): 视频描述提示词
- `duration`: 视频时长 (5, 10, 15, 30秒)
- `aspect_ratio`: 视频比例 (16:9, 9:16, 1:1)
- `quality`: 视频质量 (standard, hd, uhd)
- `style`: 视频风格 (realistic, animated, artistic, cinematic)
- `negative_prompt`: 负面提示词
- `seed`: 随机种子
- `steps`: 生成步数 (20-50)
- `model`: 模型 (sora-2, sora-2-pro)

### 2. 获取视频列表

**GET** `/api/sora2/list`

**查询参数:**
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 20)
- `status`: 状态筛选 (pending, processing, completed, failed)

### 3. 获取历史记录

**GET** `/api/sora2/history`

**查询参数:**
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 10)

### 4. 获取收藏列表

**GET** `/api/sora2/favorites`

**查询参数:**
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 10)

### 5. 切换收藏状态

**POST** `/api/sora2/favorite`

**查询参数:**
- `id`: 视频ID

### 6. 下载视频

**GET** `/api/sora2/download`

**查询参数:**
- `id`: 视频ID

### 7. 删除视频

**GET** `/api/sora2/remove`

**查询参数:**
- `id`: 视频ID

### 8. 发布/取消发布

**POST** `/api/sora2/publish`

**查询参数:**
- `id`: 视频ID
- `publish`: 是否发布 (true/false)

## 数据库表结构

### chatgpt_sora2_jobs 表

```sql
CREATE TABLE `chatgpt_sora2_jobs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `channel` varchar(255) DEFAULT NULL COMMENT '频道',
  `task_id` varchar(255) DEFAULT NULL COMMENT '外部任务ID',
  `task_info` text COMMENT '原始任务信息',
  `prompt` text NOT NULL COMMENT '提示词',
  `prompt_ext` text COMMENT '优化后提示词',
  `cover_url` varchar(500) DEFAULT NULL COMMENT '封面图URL',
  `video_url` varchar(500) DEFAULT NULL COMMENT '无水印视频URL',
  `water_url` varchar(500) DEFAULT NULL COMMENT '有水印视频URL',
  `thumbnail_url` varchar(500) DEFAULT NULL COMMENT '缩略图URL',
  `status` varchar(50) DEFAULT 'pending' COMMENT '任务状态',
  `publish` tinyint(1) DEFAULT 0 COMMENT '是否发布',
  `is_favorite` tinyint(1) DEFAULT 0 COMMENT '是否收藏',
  `err_msg` text COMMENT '错误信息',
  `raw_data` text COMMENT '原始数据json',
  `power` int(11) DEFAULT 0 COMMENT '消耗算力',
  `views` int(11) DEFAULT 0 COMMENT '观看次数',
  `likes` int(11) DEFAULT 0 COMMENT '点赞次数',
  `shares` int(11) DEFAULT 0 COMMENT '分享次数',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Sora2视频生成任务表';
```

## 配置说明

### API Key 配置

需要在 `chatgpt_api_keys` 表中添加 Sora2 API Key：

```sql
INSERT INTO `chatgpt_api_keys` (`name`, `value`, `type`, `api_url`, `enabled`) 
VALUES ('Sora2 API Key', 'your_api_key_here', 'sora2', 'https://api.openai.com', 1);
```

### 算力消耗

不同参数组合的算力消耗：

- **基础算力**: 10
- **时长**: 每增加1秒 +2算力
- **质量**: HD +8算力, UHD +15算力
- **模型**: sora-2-pro +5算力

## 任务状态

- `pending`: 等待处理
- `processing`: 正在生成
- `completed`: 生成完成
- `failed`: 生成失败

## WebSocket 通知

支持通过 WebSocket 实时推送任务状态更新：

```javascript
// 连接 WebSocket
const ws = new WebSocket('ws://localhost:8080/api/ws');

// 监听 Sora2 频道消息
ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  if (message.channel === 'sora2') {
    console.log('Sora2 任务更新:', message.body);
  }
};
```

## 使用示例

### JavaScript 示例

```javascript
// 生成视频
async function generateVideo() {
  const response = await fetch('/api/sora2/generate', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    },
    body: JSON.stringify({
      prompt: '一只可爱的小猫在花园里玩耍',
      duration: '10',
      aspect_ratio: '16:9',
      quality: 'hd',
      style: 'realistic'
    })
  });
  
  const result = await response.json();
  console.log('生成结果:', result);
}

// 获取视频列表
async function getVideoList() {
  const response = await fetch('/api/sora2/list?page=1&page_size=20');
  const result = await response.json();
  console.log('视频列表:', result);
}
```

### Go 示例

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type Sora2Request struct {
    Prompt      string `json:"prompt"`
    Duration    string `json:"duration"`
    AspectRatio string `json:"aspect_ratio"`
    Quality     string `json:"quality"`
    Style       string `json:"style"`
}

func generateVideo() {
    url := "http://localhost:8080/api/sora2/generate"
    
    requestData := Sora2Request{
        Prompt:      "一只可爱的小猫在花园里玩耍",
        Duration:    "10",
        AspectRatio: "16:9",
        Quality:     "hd",
        Style:       "realistic",
    }
    
    jsonData, _ := json.Marshal(requestData)
    
    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer YOUR_TOKEN")
    
    client := &http.Client{}
    resp, _ := client.Do(req)
    defer resp.Body.Close()
}
```

## 注意事项

1. **API Key**: 确保配置了有效的 Sora2 API Key
2. **算力检查**: 生成前会检查用户算力是否足够
3. **任务队列**: 使用 Redis 队列管理任务，支持高并发
4. **文件下载**: 自动下载生成的视频文件到本地存储
5. **错误处理**: 完善的错误处理和算力退还机制
6. **审查机制**: 遵循 OpenAI 的内容审查政策

## 故障排除

### 常见问题

1. **任务失败**: 检查 API Key 是否有效，网络连接是否正常
2. **算力不足**: 用户需要充值算力才能生成视频
3. **文件下载失败**: 检查存储配置和网络连接
4. **任务超时**: 长时间未完成的任务会自动标记为失败

### 日志查看

```bash
# 查看 Sora2 服务日志
tail -f logs/app.log | grep sora2

# 查看任务队列状态
redis-cli llen Sora2_Task_Queue
```
