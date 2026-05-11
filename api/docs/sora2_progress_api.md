# Sora2 任务进度查询 API 文档

## 概述
本文档描述了 Sora2 视频生成任务的进度查询接口，支持实时获取任务状态和进度信息。

## 接口列表

### 1. 查询任务进度

**接口地址：** `GET /api/sora2/progress/:task_id`

**请求参数：**
- `task_id` (路径参数): 任务ID，必需

**请求头：**
- `Authorization`: Bearer token (可选，用于身份验证)

**响应格式：**

#### 成功响应 (200)
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "veo3:1756693796-YQVHH4A3Lg",
    "status": "SUCCESS",
    "progress": "100",
    "video_url": "https://filesystem.site/cdn/20250901/018eg2SgUpHMT6EEuQbfeRLWeUhE75.mp4",
    "cover_url": "",
    "water_url": "https://filesystem.site/cdn/20250901/018eg2SgUpHMT6EEuQbfeRLWeUhE75.mp4",
    "err_msg": "",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:05:00Z"
  }
}
```

#### 外部API响应格式
外部API `/v2/videos/generations/{task_id}` 返回的原始格式：
```json
{
  "task_id": "veo3:1756693796-YQVHH4A3Lg",
  "platform": "google",
  "action": "google-videos",
  "status": "SUCCESS",
  "fail_reason": "",
  "submit_time": 1756693797,
  "start_time": 1756693808,
  "finish_time": 1756693898,
  "progress": "100%",
  "data": {
    "output": "https://filesystem.site/cdn/20250901/018eg2SgUpHMT6EEuQbfeRLWeUhE75.mp4"
  },
  "search_item": ""
}
```

#### 错误响应
```json
{
  "code": 1,
  "message": "任务不存在",
  "data": null
}
```

**状态枚举：**
- `NOT_START`: 未开始
- `IN_PROGRESS`: 正在执行
- `SUCCESS`: 执行完成
- `FAILURE`: 失败

**进度值说明：**
- `0`: 未开始或失败
- `1-99`: 进行中（具体数值表示完成百分比）
- `100`: 完成

## 使用示例

### JavaScript 示例
```javascript
// 查询任务进度
async function getTaskProgress(taskId) {
  try {
    const response = await fetch(`/api/sora2/progress/${taskId}`, {
      method: 'GET',
      headers: {
        'Authorization': 'Bearer ' + token,
        'Content-Type': 'application/json'
      }
    });
    
    const result = await response.json();
    
    if (result.code === 0) {
      const data = result.data;
      console.log('任务状态:', data.status);
      console.log('进度:', data.progress + '%');
      
      if (data.status === 'SUCCESS') {
        console.log('视频URL:', data.video_url);
      } else if (data.status === 'FAILURE') {
        console.log('错误信息:', data.err_msg);
      }
    } else {
      console.error('查询失败:', result.message);
    }
  } catch (error) {
    console.error('请求失败:', error);
  }
}

// 轮询进度
function pollProgress(taskId, callback) {
  const interval = setInterval(async () => {
    const result = await getTaskProgress(taskId);
    
    if (result) {
      callback(result);
      
      // 如果任务完成或失败，停止轮询
      if (result.status === 'SUCCESS' || result.status === 'FAILURE') {
        clearInterval(interval);
      }
    }
  }, 2000); // 每2秒查询一次
  
  return interval;
}
```

### Go 示例
```go
package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
)

func getTaskProgress(taskId string) error {
    url := fmt.Sprintf("/api/sora2/progress/%s", taskId)
    method := "GET"

    client := &http.Client{}
    req, err := http.NewRequest(method, url, nil)
    if err != nil {
        return err
    }

    req.Header.Add("Authorization", "Bearer YOUR_TOKEN")
    req.Header.Add("Content-Type", "application/json")

    res, err := client.Do(req)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return err
    }

    fmt.Println(string(body))
    return nil
}
```

## WebSocket 实时推送

除了轮询查询，系统还支持通过 WebSocket 实时推送任务进度更新。

### WebSocket 消息格式
```json
{
  "type": "task_progress",
  "data": {
    "task_id": "task_123456",
    "status": "IN_PROGRESS",
    "progress": "75",
    "video_url": "",
    "cover_url": "",
    "water_url": "",
    "err_msg": "",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:05:00Z"
  }
}
```

### 前端 WebSocket 示例
```javascript
// 连接 WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  
  if (message.type === 'task_progress') {
    const data = message.data;
    console.log('收到进度更新:', data);
    
    // 更新UI
    updateProgressUI(data);
  }
};

function updateProgressUI(data) {
  // 更新进度条
  document.getElementById('progress').style.width = data.progress + '%';
  
  // 更新状态文本
  document.getElementById('status').textContent = getStatusText(data.status);
  
  // 如果完成，显示视频
  if (data.status === 'SUCCESS' && data.video_url) {
    document.getElementById('video').src = data.video_url;
  }
}
```

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | 任务不存在 |
| 2 | 无权限访问此任务 |
| 3 | 请先登录 |
| 4 | 任务ID不能为空 |

## 注意事项

1. **轮询频率**: 建议每2-5秒查询一次，避免过于频繁的请求
2. **超时处理**: 如果任务长时间无响应，建议设置超时机制
3. **错误处理**: 需要妥善处理网络错误和API错误
4. **权限验证**: 只能查询自己创建的任务进度
5. **资源清理**: 任务完成后及时清理轮询定时器

## 最佳实践

1. **结合使用**: 建议同时使用轮询和WebSocket，轮询作为备用方案
2. **状态缓存**: 前端可以缓存任务状态，减少不必要的请求
3. **用户体验**: 提供清晰的进度提示和错误信息
4. **性能优化**: 避免同时轮询过多任务
