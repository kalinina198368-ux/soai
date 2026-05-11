<template>
  <div class="page-luma">
    <div class="prompt-box">
      <!-- 模式切换和模型选择 -->
      <div class="header-controls">
        <div class="mode-tabs">
          <div 
            class="tab" 
            :class="{ active: !isImageToVideo }"
            @click="toggleVideoMode"
          >
            文生视频
          </div>
          <div 
            class="tab" 
            :class="{ active: isImageToVideo }"
            @click="toggleVideoMode"
          >
            图生视频
          </div>
        </div>
        <div class="model-selector">
          <el-select v-model="formData.model" size="small" style="width: 150px">
            <el-option label="sora-2" value="sora-2" />
          </el-select>
        </div>
      </div>

      <!-- 宽高比和时长选择 -->
      <div class="params-row">
        <div class="param-group">
          <span class="param-label">宽高比：</span>
          <div class="aspect-ratios">
            <div 
              v-for="ratio in aspectRatios" 
              :key="ratio.value"
              class="aspect-ratio-item"
              :class="{ active: formData.aspectRatio === ratio.value }"
              @click="selectAspectRatio(ratio.value)"
            >
              {{ ratio.label }}
            </div>
          </div>
        </div>
        <div class="param-group">
          <span class="param-label">时长：</span>
          <div class="duration-options">
            <div 
              class="duration-item"
              :class="{ active: formData.duration === '10s' }"
              @click="selectDuration('10s')"
            >10s</div>
            <div 
              class="duration-item"
              :class="{ active: formData.duration === '15s' }"
              @click="selectDuration('15s')"
            >15s</div>
          </div>
        </div>
      </div>

      <!-- 图生视频模式：多图片上传 -->
      <div v-if="isImageToVideo" class="image-upload-section">
        <div class="upload-header">
          <span class="upload-title">上传图片 ({{ uploadedImages.length }}/10)</span>
          <el-button 
            v-if="uploadedImages.length > 0" 
            size="small" 
            type="warning" 
            @click="clearAllImages"
          >
            清空
          </el-button>
        </div>
        <div class="image-grid">
          <div 
            v-for="image in uploadedImages" 
            :key="image.id" 
            class="image-item"
          >
            <el-image :src="image.url" fit="cover" />
            <div class="image-overlay">
              <el-icon @click="removeImage(image.id)"><CircleCloseFilled /></el-icon>
            </div>
            <div class="image-name">{{ image.name }}</div>
          </div>
          <div 
            v-if="uploadedImages.length < 10" 
            class="upload-item" 
            @click="uploadImages"
            :class="{ uploading: isUploading }"
          >
            <el-icon v-if="!isUploading"><Plus /></el-icon>
            <el-icon v-else class="is-loading"><Loading /></el-icon>
            <span>{{ isUploading ? '上传中...' : '添加图片' }}</span>
          </div>
        </div>
      </div>

      <!-- 文生视频模式：参考图片 -->
      <div v-else class="reference-image-section">
        <div v-if="referenceImagePreview" class="reference-image-preview">
          <div class="preview-item">
            <el-image :src="referenceImagePreview" fit="cover" />
            <div class="preview-overlay">
              <el-icon @click="removeReferenceImage"><CircleCloseFilled /></el-icon>
            </div>
          </div>
        </div>
        <div 
          class="upload-item" 
          @click="uploadReferenceImage"
          :class="{ 'has-image': formData.referenceImage }"
        >
          <el-icon v-if="!formData.referenceImage"><Plus /></el-icon>
          <el-icon v-else><Check /></el-icon>
          <span>{{ formData.referenceImage ? '更换参考图片' : '参考图片' }}</span>
        </div>
      </div>

      <div class="prompt-container">
        <div class="input-container">
          <textarea class="prompt-input" :rows="row" v-model="formData.prompt" placeholder="请输入视频创作描述" autofocus> </textarea>
          <div class="send-icon" @click="generateVideo">
            <i class="iconfont icon-send"></i>
          </div>
        </div>

        <div class="params">
          <div class="item-group">
            <el-button class="generate-btn" size="small" @click="generatePrompt" color="#5865f2">
              <i class="iconfont icon-chuangzuo" style="margin-right: 5px"></i>
              <span>生成AI视频提示词</span>
            </el-button>
          </div>
          <div class="item-group">
            <span class="label">循环参考图</span>
            <el-switch v-model="formData.loop" size="small" />
          </div>
          <div class="item-group">
            <span class="label">提示词优化</span>
            <el-switch v-model="formData.enhancePrompt" size="small" />
          </div>
        </div>
      </div>
    </div>

    <el-container class="video-container" v-loading="loading" element-loading-background="rgba(100,100,100,0.3)">
      <h2 class="h-title text-2xl mb-5 mt-2">你的作品</h2>

      <div class="list-box" v-if="!noData">
        <div v-for="item in list" :key="item.id">
          <div class="item">
            <div class="left">
              <div class="container">
                <div v-if="item.status === 'completed' && item.video_url">
                  <video class="video" :src="replaceImg(item.video_url)" preload="auto" loop="loop">您的浏览器不支持视频播放</video>
                  <button class="play flex justify-center items-center" @click="play(item)">
                    <img src="/images/play.svg" alt="" />
                  </button>
                </div>
                <div v-else-if="item.status === 'failed'" class="failed-container">
                  <el-icon><WarningFilled /></el-icon>
                  <span>生成失败</span>
                </div>
                <div v-else-if="item.status === 'processing'" class="processing-container">
                  <generating message="正在生成视频" />
                  <div class="progress-info" v-if="parseProgress(item) > 0">
                    <el-progress :percentage="parseProgress(item)" :show-text="true" />
                    <span class="progress-text">{{ parseProgress(item) }}%</span>
                  </div>
                </div>
                <generating message="等待中..." v-else />
              </div>
            </div>
            <div class="center">
              <div class="failed" v-if="item.status === 'failed'">
                任务执行失败：{{ item.err_msg || '未知错误' }}，任务提示词：{{ item.prompt }}
              </div>
              <div class="prompt" v-else>{{ item.prompt }}</div>
              <div class="video-meta" v-if="item.status === 'completed'">
                <el-tag size="small" type="primary">{{ formatDurationDisplay(getItemDuration(item)) }}</el-tag>
                <el-tag size="small" type="success">{{ getAspectRatioLabel(getItemAspectRatio(item)) }}</el-tag>
                <el-tag v-if="item.is_favorite" size="small" type="danger">收藏</el-tag>
              </div>
            </div>
            <div class="right" v-if="item.status === 'completed'">
              <div class="tools">
                <el-tooltip content="下载视频" placement="top">
                  <button class="btn btn-icon" @click="download(item)" :disabled="item.downloading">
                    <i class="iconfont icon-download" v-if="!item.downloading"></i>
                    <el-image src="/images/loading.gif" class="downloading" fit="cover" v-else />
                  </button>
                </el-tooltip>
                <el-tooltip content="收藏" placement="top">
                  <button class="btn btn-icon" @click="toggleFavorite(item)">
                    <i class="iconfont" :class="item.is_favorite ? 'icon-star-fill' : 'icon-star'"></i>
                  </button>
                </el-tooltip>
                <el-tooltip content="删除" placement="top">
                  <button class="btn btn-icon" @click="removeJob(item)">
                    <i class="iconfont icon-remove"></i>
                  </button>
                </el-tooltip>
              </div>
            </div>
            <div class="right-error" v-else>
              <el-button type="danger" @click="removeJob(item)" circle>
                <i class="iconfont icon-remove"></i>
              </el-button>
            </div>
          </div>
        </div>
      </div>
      <el-empty :image-size="100" :image="nodata" description="没有任何作品，赶紧去创作吧！" v-else />

      <div class="pagination">
        <el-pagination
          v-if="total > pageSize"
          background
          style="--el-pagination-button-bg-color: rgba(86, 86, 95, 0.2)"
          layout="total,prev, pager, next"
          :hide-on-single-page="true"
          v-model:current-page="page"
          v-model:page-size="pageSize"
          @current-change="fetchData(page)"
          :total="total"
        />
      </div>
    </el-container>
    <black-dialog v-model:show="showDialog" title="预览视频" hide-footer @cancal="showDialog = false" width="auto">
      <video style="width: 100%; max-height: 90vh" :src="currentVideoUrl" preload="auto" :autoplay="true" loop="loop" v-show="showDialog" controls>
        您的浏览器不支持视频播放
      </video>
    </black-dialog>
  </div>
</template>

<script setup>
import nodata from "@/assets/img/no-data.png";

import { onMounted, onUnmounted, reactive, ref } from "vue";
import { CircleCloseFilled, Plus, Loading, Check, WarningFilled } from "@element-plus/icons-vue";
import { httpDownload, httpPost, httpGet } from "@/utils/http";
import { checkSession, getClientId } from "@/store/cache";
import { closeLoading, showLoading, showMessageError, showMessageOK } from "@/utils/dialog";
import { replaceImg } from "@/utils/libs";
import { ElMessage, ElMessageBox } from "element-plus";
import Generating from "@/components/ui/Generating.vue";
import BlackDialog from "@/components/ui/BlackDialog.vue";
import { useSharedStore } from "@/store/sharedata";

const showDialog = ref(false);
const currentVideoUrl = ref("");
const row = ref(1);

// 图生视频相关状态
const isImageToVideo = ref(true);
const uploadedImages = ref([]);
const isUploading = ref(false);

// 文生视频参考图片状态
const referenceImagePreview = ref(null);

// 表单数据
const formData = reactive({
  prompt: "",
  model: "sora-2",
  aspectRatio: "16:9",
  duration: "10s",
  enhancePrompt: true,
  loop: false,
  referenceImage: null,
  endFrameImage: null,
});

// 宽高比选项
const aspectRatios = ref([
  { value: "16:9", label: "横屏" },
  { value: "9:16", label: "竖屏" },
]);

const store = useSharedStore();

// 设置WebSocket消息处理函数
const setupWebSocketHandler = () => {
  const checkAndSetup = () => {
    if (store.socket.conn && store.socket.conn.readyState === WebSocket.OPEN) {
      const handleWebSocketMessage = (event) => {
        try {
          let data;
          if (event.data instanceof Blob) {
            const reader = new FileReader();
            reader.readAsText(event.data, "UTF-8");
            reader.onload = () => {
              data = JSON.parse(String(reader.result));
              processSora2Message(data);
            };
          } else {
            data = JSON.parse(event.data);
            processSora2Message(data);
          }
        } catch (e) {
          console.warn("WebSocket消息解析失败:", e);
        }
      };

      const processSora2Message = (data) => {
        // 丢弃无关消息
        if (data.channel !== "sora2") {
          return;
        }

        console.log("🔔 收到Sora2 WebSocket消息:", data);

        // 处理完成/失败消息
        if (data.body === "FINISH" || data.body === "FAIL") {
          console.log("🔄 收到完成/失败消息，刷新列表");
          fetchData(1);
          return;
        }

        // 处理进度更新消息
        if (data.body && typeof data.body === "object") {
          const body = data.body;
          console.log("📊 处理进度更新消息:", body);

          // 更新列表中对应视频项的进度
          if (body.job_id) {
            const jobId = parseInt(body.job_id);
            const videoIndex = list.value.findIndex(item => item.id === jobId);

            if (videoIndex !== -1) {
              const currentItem = list.value[videoIndex];

              // 更新raw_data
              currentItem.raw_data = JSON.stringify(body);

              // 更新状态
              if (body.status === "IN_PROGRESS") {
                currentItem.status = "processing";
              } else if (body.status === "SUCCESS") {
                currentItem.status = "completed";
                // 如果状态是完成且有视频URL，更新video_url
                if (body.data && body.data.output) {
                  currentItem.video_url = body.data.output;
                }
              } else if (body.status === "FAILURE") {
                currentItem.status = "failed";
                currentItem.err_msg = body.err_msg || body.fail_reason || "生成失败";
              }
            }
          }
        }
      };

      // 添加事件监听器
      store.socket.conn.addEventListener("message", handleWebSocketMessage);

      // 保存处理器引用，用于清理
      store.sora2MessageHandler = handleWebSocketMessage;

      console.log("✅ WebSocket消息处理器已设置");
    } else {
      setTimeout(checkAndSetup, 1000);
    }
  };

  checkAndSetup();
};

onMounted(() => {
  checkSession().then(() => {
    fetchData(1);
  });

  // 设置WebSocket消息处理
  setupWebSocketHandler();
});

onUnmounted(() => {
  // 清理WebSocket事件监听器
  if (store.socket.conn && store.sora2MessageHandler) {
    store.socket.conn.removeEventListener("message", store.sora2MessageHandler);
    delete store.sora2MessageHandler;
  }
  // 清理图片预览URL
  if (referenceImagePreview.value) {
    URL.revokeObjectURL(referenceImagePreview.value);
  }
  // 清理图生视频图片URL
  uploadedImages.value.forEach(img => {
    if (img.url) {
      URL.revokeObjectURL(img.url);
    }
  });
});

// 选择宽高比
const selectAspectRatio = (ratio) => {
  formData.aspectRatio = ratio;
};

// 选择时长
const selectDuration = (duration) => {
  formData.duration = duration;
};

// 切换文生视频/图生视频模式
const toggleVideoMode = () => {
  isImageToVideo.value = !isImageToVideo.value;
  if (!isImageToVideo.value) {
    // 切换到文生视频时清空图生视频的图片
    clearAllImages();
  } else {
    // 切换到图生视频时清空参考图片
    if (referenceImagePreview.value) {
      URL.revokeObjectURL(referenceImagePreview.value);
      referenceImagePreview.value = null;
      formData.referenceImage = null;
    }
  }
};

// 上传参考图片（文生视频模式）
const uploadReferenceImage = () => {
  const input = document.createElement('input');
  input.type = 'file';
  input.accept = 'image/*';
  input.multiple = false;
  input.onchange = handleReferenceImageUpload;
  input.click();
};

const handleReferenceImageUpload = async (event) => {
  const file = event.target.files[0];
  if (!file) return;

  try {
    const base64 = await fileToBase64(file);
    formData.referenceImage = base64;

    // 创建预览URL
    referenceImagePreview.value = URL.createObjectURL(file);

    ElMessage.success("参考图片上传成功");
  } catch (error) {
    ElMessage.error("参考图片上传失败：" + error.message);
  }
};

// 删除参考图片
const removeReferenceImage = () => {
  if (referenceImagePreview.value) {
    URL.revokeObjectURL(referenceImagePreview.value);
  }
  referenceImagePreview.value = null;
  formData.referenceImage = null;

  ElMessage.success("参考图片已删除");
};

// 图片上传相关函数
const uploadImages = () => {
  const input = document.createElement('input');
  input.type = 'file';
  input.accept = 'image/*';
  input.multiple = true;
  input.onchange = handleImageUpload;
  input.click();
};

const handleImageUpload = async (event) => {
  const files = Array.from(event.target.files);
  if (files.length === 0) return;

  isUploading.value = true;

  try {
    for (const file of files) {
      if (uploadedImages.value.length >= 10) {
        ElMessage.warning("最多只能上传10张图片");
        break;
      }
      const base64 = await fileToBase64(file);
      const imageData = {
        id: Date.now() + Math.random(),
        name: file.name,
        size: file.size,
        base64: base64,
        url: URL.createObjectURL(file)
      };
      uploadedImages.value.push(imageData);
    }

    ElMessage.success(`成功上传 ${Math.min(files.length, 10 - uploadedImages.value.length + files.length)} 张图片`);
  } catch (error) {
    ElMessage.error("图片上传失败：" + error.message);
  } finally {
    isUploading.value = false;
  }
};

const fileToBase64 = (file) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => resolve(reader.result);
    reader.onerror = error => reject(error);
  });
};

const removeImage = (imageId) => {
  const index = uploadedImages.value.findIndex(img => img.id === imageId);
  if (index !== -1) {
    // 释放URL对象
    if (uploadedImages.value[index].url) {
      URL.revokeObjectURL(uploadedImages.value[index].url);
    }
    uploadedImages.value.splice(index, 1);
    ElMessage.success("图片已删除");
  }
};

const clearAllImages = () => {
  uploadedImages.value.forEach(img => {
    if (img.url) {
      URL.revokeObjectURL(img.url);
    }
  });
  uploadedImages.value = [];
  ElMessage.success("已清空所有图片");
};

const loading = ref(false);
const list = ref([]);
const noData = ref(true);
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);

const fetchData = (_page) => {
  if (_page) {
    page.value = _page;
  }

  loading.value = true;

  httpGet("/api/sora2/list", {
    page: page.value,
    page_size: pageSize.value,
  })
    .then((res) => {
      total.value = res.data.total;
      loading.value = false;
      list.value = res.data.data || [];
      noData.value = list.value.length === 0;
    })
    .catch((error) => {
      console.error("获取视频列表失败:", error);
      loading.value = false;
      noData.value = true;
    });
};

// 解析raw_data中的进度
const parseProgress = (item) => {
  const rawDataStr = item.raw_data;
  if (!rawDataStr) {
    return 0;
  }

  try {
    const rawData = JSON.parse(rawDataStr);
    const progress = parseInt(rawData.progress || "0", 10) || 0;
    return progress;
  } catch (e) {
    console.error("parseProgress: 解析raw_data失败:", e);
    return 0;
  }
};

// 生成视频
const generateVideo = async () => {
  if (!formData.prompt.trim()) {
    return showMessageError("请输入视频描述");
  }

  try {
    let createPayload;

    if (isImageToVideo.value) {
      // 图生视频模式
      createPayload = {
        prompt: formData.prompt,
        model: formData.model,
        images: uploadedImages.value.map(img => img.base64), // 使用base64格式
        aspect_ratio: formData.aspectRatio,
        duration: mapDurationToSeconds(formData.duration),
        enhance_prompt: formData.enhancePrompt,
        loop: formData.loop,
      };
    } else {
      // 文生视频模式
      createPayload = {
        prompt: formData.prompt,
        model: formData.model,
        aspect_ratio: formData.aspectRatio,
        duration: mapDurationToSeconds(formData.duration),
        enhance_prompt: formData.enhancePrompt,
        loop: formData.loop,
        reference_image: formData.referenceImage,
        end_frame_image: formData.endFrameImage,
      };
    }

    console.log("📤 提交生成任务:", createPayload);

    const createRes = await httpPost("/api/sora2/generate", createPayload);
    const taskId = createRes?.data?.task_id;
    if (!taskId) {
      throw new Error("生成任务提交失败，未返回 task_id");
    }

    showMessageOK(createRes?.data?.message || "任务已提交，正在生成中...");

    // 立即刷新视频列表，显示新创建的 processing 任务
    fetchData(1);
  } catch (error) {
    showMessageError(error.message || "生成失败，请稍后重试");
  }
};

// 格式化时长
const mapDurationToSeconds = (durationStr) => {
  if (!durationStr) return "10";
  const m = String(durationStr).match(/^(\d+)/);
  return m ? m[1] : "10";
};

// 获取宽高比标签
const getAspectRatioLabel = (aspectRatio) => {
  const ratioMap = {
    "16:9": "横屏",
    "9:16": "竖屏",
    "1:1": "方形",
    "4:3": "标准",
    "3:4": "竖屏标准"
  };
  return ratioMap[aspectRatio] || aspectRatio || "16:9";
};

// 从后端返回的数据中解析时长
const getItemDuration = (item) => {
  // 1) 尝试从 raw_data 解析
  try {
    if (item?.raw_data) {
      const data = JSON.parse(item.raw_data);
      const v = data?.params?.duration || data?.duration;
      if (v !== undefined && v !== null && String(v).trim() !== "") {
        return String(v);
      }
    }
  } catch (_) {}

  // 2) 尝试从 task_info 解析
  try {
    if (item?.task_info) {
      const info = JSON.parse(item.task_info);
      const v = info?.params?.duration || info?.duration || info?.task?.params?.duration;
      if (v !== undefined && v !== null && String(v).trim() !== "") {
        return String(v);
      }
    }
  } catch (_) {}

  // 3) 回退
  return String(item?.duration || "10");
};

// 从后端返回的数据中解析宽高比
const getItemAspectRatio = (item) => {
  // 1) raw_data
  try {
    if (item?.raw_data) {
      const data = JSON.parse(item.raw_data);
      const v = data?.params?.aspect_ratio || data?.aspect_ratio;
      if (v !== undefined && v !== null && String(v).trim() !== "") {
        return String(v);
      }
    }
  } catch (_) {}

  // 2) task_info
  try {
    if (item?.task_info) {
      const info = JSON.parse(item.task_info);
      const v = info?.params?.aspect_ratio || info?.aspect_ratio || info?.task?.params?.aspect_ratio;
      if (v !== undefined && v !== null && String(v).trim() !== "") {
        return String(v);
      }
    }
  } catch (_) {}

  // 3) 回退
  return String(item?.aspect_ratio || "16:9");
};

// 显示 10s / 15s
const formatDurationDisplay = (d) => {
  let v = String(d || "10").trim();
  if (/^\d+$/.test(v)) return v + "s";
  if (v.toLowerCase().endsWith("s")) return v;
  return v;
};

const download = (item) => {
  const url = replaceImg(item.video_url);
  const downloadURL = `${process.env.VUE_APP_API_HOST}/api/download?url=${url}`;
  // parse filename
  const urlObj = new URL(url);
  const fileName = urlObj.pathname.split("/").pop();
  item.downloading = true;
  httpDownload(downloadURL)
    .then((response) => {
      const blob = new Blob([response.data]);
      const link = document.createElement("a");
      link.href = URL.createObjectURL(blob);
      link.download = fileName;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(link.href);
      item.downloading = false;
    })
    .catch(() => {
      showMessageError("下载失败");
      item.downloading = false;
    });
};

const play = (item) => {
  currentVideoUrl.value = replaceImg(item.video_url);
  showDialog.value = true;
};

const removeJob = (item) => {
  ElMessageBox.confirm("此操作将会删除任务相关文件，继续操作吗?", "删除提示", {
    confirmButtonText: "确认",
    cancelButtonText: "取消",
    type: "warning",
  })
    .then(() => {
      httpGet("/api/sora2/remove", { id: item.id })
        .then(() => {
          ElMessage.success("任务删除成功");
          fetchData(1);
        })
        .catch((e) => {
          ElMessage.error("任务删除失败：" + e.message);
        });
    })
    .catch(() => {});
};

// 切换收藏状态
const toggleFavorite = (item) => {
  httpGet("/api/sora2/favorite", { id: item.id })
    .then((res) => {
      item.is_favorite = res.data.is_favorite;
      ElMessage.success(res.data.message);
    })
    .catch((e) => {
      ElMessage.error("操作失败：" + e.message);
    });
};

const generatePrompt = () => {
  if (formData.prompt === "") {
    return showMessageError("请输入原始提示词");
  }
  showLoading("正在生成视频脚本...");
  httpPost("/api/prompt/video", { prompt: formData.prompt })
    .then((res) => {
      formData.prompt = res.data;
      closeLoading();
    })
    .catch((e) => {
      showMessageError("生成提示词失败：" + e.message);
      closeLoading();
    });
};
</script>

<style lang="stylus" scoped>
@import "@/assets/css/luma.styl"

// 新增样式
.header-controls
  display: flex
  justify-content: space-between
  align-items: center
  margin-bottom: 20px

  .mode-tabs
    display: flex
    gap: 10px

    .tab
      padding: 8px 16px
      border-radius: 6px
      cursor: pointer
      transition: all 0.3s ease
      color: #666
      background: rgba(255, 255, 255, 0.1)

      &.active
        background: #5865f2
        color: #fff
        font-weight: 600

      &:hover
        background: rgba(88, 101, 242, 0.2)

  .model-selector
    display: flex
    align-items: center

.params-row
  display: flex
  gap: 30px
  margin-bottom: 20px
  flex-wrap: wrap

  .param-group
    display: flex
    align-items: center
    gap: 10px

    .param-label
      font-size: 14px
      color: #666
      white-space: nowrap

    .aspect-ratios
      display: flex
      gap: 8px

      .aspect-ratio-item
        padding: 6px 12px
        border-radius: 4px
        cursor: pointer
        transition: all 0.3s ease
        background: rgba(255, 255, 255, 0.1)
        color: #666

        &.active
          background: #5865f2
          color: #fff

        &:hover
          background: rgba(88, 101, 242, 0.2)

    .duration-options
      display: flex
      gap: 8px

      .duration-item
        padding: 6px 12px
        border-radius: 4px
        cursor: pointer
        transition: all 0.3s ease
        background: rgba(255, 255, 255, 0.1)
        color: #666

        &.active
          background: #5865f2
          color: #fff

        &:hover
          background: rgba(88, 101, 242, 0.2)

.image-upload-section
  margin-bottom: 20px

  .upload-header
    display: flex
    justify-content: space-between
    align-items: center
    margin-bottom: 12px

    .upload-title
      font-size: 14px
      font-weight: 600
      color: #666

  .image-grid
    display: grid
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr))
    gap: 12px

    .image-item
      position: relative
      aspect-ratio: 1
      border-radius: 8px
      overflow: hidden
      background: rgba(255, 255, 255, 0.1)
      border: 1px solid rgba(255, 255, 255, 0.2)

      .image-overlay
        position: absolute
        top: 4px
        right: 4px
        background: rgba(0, 0, 0, 0.6)
        border-radius: 50%
        padding: 4px
        cursor: pointer
        display: flex
        align-items: center
        justify-content: center

        .el-icon
          color: white
          font-size: 16px

      .image-name
        position: absolute
        bottom: 0
        left: 0
        right: 0
        background: rgba(0, 0, 0, 0.7)
        color: white
        font-size: 10px
        padding: 4px
        text-align: center
        white-space: nowrap
        overflow: hidden
        text-overflow: ellipsis

    .upload-item
      display: flex
      flex-direction: column
      align-items: center
      justify-content: center
      aspect-ratio: 1
      background: rgba(255, 255, 255, 0.1)
      border: 2px dashed rgba(255, 255, 255, 0.3)
      border-radius: 8px
      cursor: pointer
      transition: all 0.3s ease

      &.uploading
        border-color: #5865f2
        background: rgba(88, 101, 242, 0.1)

      &:hover
        border-color: #5865f2
        background: rgba(88, 101, 242, 0.1)

      .el-icon
        font-size: 24px
        color: #666
        margin-bottom: 8px

      span
        font-size: 12px
        color: #666

.reference-image-section
  display: flex
  gap: 12px
  margin-bottom: 20px
  align-items: flex-start

  .reference-image-preview
    .preview-item
      position: relative
      width: 100px
      height: 100px
      border-radius: 8px
      overflow: hidden
      background: rgba(255, 255, 255, 0.1)
      border: 1px solid rgba(255, 255, 255, 0.2)

      .preview-overlay
        position: absolute
        top: 4px
        right: 4px
        background: rgba(0, 0, 0, 0.6)
        border-radius: 50%
        padding: 4px
        cursor: pointer
        display: flex
        align-items: center
        justify-content: center

        .el-icon
          color: white
          font-size: 16px

  .upload-item
    display: flex
    flex-direction: column
    align-items: center
    justify-content: center
    padding: 20px
    background: rgba(255, 255, 255, 0.1)
    border: 2px dashed rgba(255, 255, 255, 0.3)
    border-radius: 8px
    cursor: pointer
    transition: all 0.3s ease
    min-width: 100px

    &.has-image
      border-color: #5865f2
      background: rgba(88, 101, 242, 0.1)

    &:hover
      border-color: #5865f2
      background: rgba(88, 101, 242, 0.1)

    .el-icon
      font-size: 24px
      color: #666
      margin-bottom: 8px

    span
      font-size: 12px
      color: #666

.video-meta
  display: flex
  gap: 8px
  margin-top: 8px
  flex-wrap: wrap

.processing-container
  display: flex
  flex-direction: column
  align-items: center
  justify-content: center
  padding: 20px

  .progress-info
    width: 100%
    margin-top: 16px

    .progress-text
      text-align: center
      margin-top: 8px
      font-size: 12px
      color: #666

.failed-container
  display: flex
  flex-direction: column
  align-items: center
  justify-content: center
  padding: 20px
  color: #f56c6c

  .el-icon
    font-size: 48px
    margin-bottom: 12px
</style>
