<template>
  <div class="mobile-mj">
    <van-form @submit="generate">
     


      <div class="text-line section-title">画质选择</div>
      <div class="text-line">
        <van-row :gutter="10">
          <van-col :span="12" v-for="item in models" :key="item.value">
            <div :class="item.value === params.model ? 'model active' : 'model'" @click="changeModel(item)">
              <div class="icon">
                <van-image :src="item.img" fit="cover"></van-image>
              </div>
              <div class="text">
                <van-text-ellipsis :content="item.text" />
              </div>
            </div>
          </van-col>
        </van-row>
      </div>

      <div class="text-line section-title">图片比例</div>
      <div class="text-line">
        <van-row :gutter="8" justify="center">
          <van-col :span="6" v-for="item in rates" :key="item.value">
            <div :class="item.value === params.rate ? 'rate active' : 'rate'" @click="changeRate(item)">
              <div class="icon">
                <van-image :src="item.img" fit="cover"></van-image>
              </div>
              <div class="text">{{ item.text }}</div>
            </div>
          </van-col>
        </van-row>
      </div>

      <div class="text-line">
        <div class="text-line section-title">
          <span>选择素材</span>
          <van-button 
            type="primary" 
            size="small" 
            icon="search" 
            @click="goToMaterialList"
            class="material-search-btn"
          >
            选择素材
          </van-button>
        </div>

        <!-- 选择匿名图片 -->
        <div class="text-line">
          <div class="section-title">
            <span>匿名图片</span>
            <van-button
              type="primary"
              size="small"
              icon="photo-o"
              @click="openUploadImagePicker"
              class="prompt-search-btn"
            >
              选择匿名图片
            </van-button>
          </div>
          <div class="prompt-placeholder" v-if="!isLogin">
            <span class="placeholder-text">登录后可查看匿名上传图片</span>
          </div>
          <div class="prompt-placeholder" v-else-if="uploadCodeImages.length === 0">
            <span class="placeholder-text">暂无匿名上传图片（可在个人中心分享上传码）</span>
          </div>
        </div>

        <van-popup v-model:show="showUploadImagePicker" position="bottom" round class="upload-image-picker">
          <div class="picker-header">
            <div class="picker-title">选择匿名图片</div>
            <div class="picker-actions">
              <van-button size="small" type="default" @click="closeUploadImagePicker">取消</van-button>
              <van-button size="small" type="primary" @click="confirmUploadImagePicker">确定</van-button>
            </div>
          </div>

          <div class="picker-body">
            <van-empty
              v-if="uploadImagesLoading && uploadCodeImages.length === 0"
              description="加载中..."
            />
            <van-empty
              v-else-if="uploadCodeImages.length === 0"
              image="https://fastly.jsdelivr.net/npm/@vant/assets/custom-empty-image.png"
              image-size="80"
              description="暂无匿名上传图片"
            />
            <div v-else class="picker-grid">
              <div
                v-for="img in uploadCodeImages"
                :key="img.id"
                class="picker-item"
                :class="{ selected: selectedUploadImageUrlsSet.has(img.url) }"
                @click="toggleSelectUploadImage(img.url)"
              >
                <van-image :src="img.url" fit="cover" width="100%" height="110" />
                <div class="selected-mark" v-if="selectedUploadImageUrlsSet.has(img.url)">
                  <van-icon name="success" size="20" color="#fff" />
                </div>
              </div>
            </div>
          </div>
        </van-popup>
        
        <!-- 已选择的素材信息 -->
        <div class="text-line" v-if="selectedMaterial">
          <div class="selected-material-card">
            <div class="material-preview">
              <van-image 
                :src="selectedMaterial.image || selectedMaterial.preview || '/images/img-placeholder.jpg'" 
                fit="cover"
                width="80"
                height="80"
                radius="8"
              >
                <template v-slot:error>
                  <div class="material-placeholder-small">
                    <i class="iconfont icon-image"></i>
                  </div>
                </template>
              </van-image>
            </div>
            <div class="material-info">
              <div class="material-name">{{ selectedMaterial.title || selectedMaterial.name || '未命名素材' }}</div>
              <van-button 
                type="default" 
                size="mini" 
                @click="clearMaterial"
                class="clear-material-btn"
              >
                清除
              </van-button>
            </div>
          </div>
        </div>

        <!-- 提示词输入 -->
        <div class="text-line">
          <div class="section-title">
            <span>提示词</span>
            <van-button 
              type="primary" 
              size="small" 
              icon="search" 
              @click="goToPromptEditor"
              class="prompt-search-btn"
            >
              {{ params.prompt ? '编辑提示词' : '编辑提示词' }}
            </van-button>
          </div>
          <!-- 显示提示词预览 -->
          <div class="prompt-preview" v-if="params.prompt">
            <div class="prompt-text">{{ params.prompt }}</div>
            <van-button 
              type="default" 
              size="mini" 
              @click="clearPrompt"
              class="clear-prompt-btn"
            >
              清除
            </van-button>
          </div>
          <div class="prompt-placeholder" v-else>
            <span class="placeholder-text">点击上方按钮编辑提示词</span>
          </div>
        </div>

        <div class="text-line">
          <van-uploader v-model="imgList" :after-read="uploadImg" />
        </div>
      </div>

      <div class="text-line pt-6">
        <div class="power-info">
          <div class="power-item">
            <span class="power-label">nb模型</span>
            <span class="power-value">{{ mjPower }}算力</span>
          </div>
          <div class="power-divider">|</div>
          <div class="power-item">
            <span class="power-label">nb2模型</span>
            <span class="power-value">{{ mj2Power }}算力</span>
          </div>
          <div class="power-divider">|</div>
          <div class="power-item current">
            <span class="power-label">当前算力</span>
            <span class="power-value highlight">{{ power }}</span>
          </div>
        </div>
      </div>

      <div class="text-line">
        <van-button round block type="primary" native-type="submit" class="generate-btn">
          <span class="btn-text">立即生成</span>
          <span class="btn-icon">→</span>
        </van-button>
      </div>

      <!-- B 供应商测试入口 -->
      <div class="text-line" style="margin-top: 16px; padding-top: 16px; border-top: 1px dashed #e8e9eb;">
        <div class="section-title" style="margin-bottom: 12px;">
          <span style="font-size: 14px; color: #969799;">应急通道生成（应急用）</span>
        </div>


        <!-- <div class="text-line" style="margin-bottom: 12px;">
          <van-field
            v-model="geminiApiKey"
            placeholder="请输入 Gemini API Key（应急通道用）"
            type="password"
            clearable
            @blur="saveGeminiApiKey"
          >
            <template #label>
              <span style="font-size: 13px;">API Key:</span>
            </template>
          </van-field>
        </div> -->
        <van-button 
          round 
          block 
          type="warning" 
          @click="generateWithGemini" 
          :loading="geminiLoading"
          :disabled="!geminiApiKey || geminiLoading"
          class="test-btn"
        >
          <span class="btn-text">{{ geminiLoading ? '生成中...' : '应急通道生成' }}</span>
        </van-button>
        <div v-if="geminiTestResult" class="gemini-test-result">
          <div class="result-title">应急通道结果：</div>
          <div v-if="geminiTestResult.images && geminiTestResult.images.length > 0" class="result-images">
            <van-image
              v-for="(img, index) in geminiTestResult.images"
              :key="index"
              :src="img"
              fit="cover"
              width="100%"
              style="margin-bottom: 8px; border-radius: 8px;"
              @click="previewGeminiImage(img)"
            />
          </div>
          <div v-if="geminiTestResult.text" class="result-text">{{ geminiTestResult.text }}</div>
          <div v-if="geminiTestResult.error" class="result-error">{{ geminiTestResult.error }}</div>
        </div>
      </div>
    </van-form>

    <h3>任务列表</h3>
    <div class="running-job-list pt-3 pb-3">
      <van-empty
        v-if="runningJobs.length === 0"
        image="https://fastly.jsdelivr.net/npm/@vant/assets/custom-empty-image.png"
        image-size="80"
        description="暂无记录"
      />
      <van-grid :gutter="10" :column-num="3" v-else>
        <van-grid-item v-for="item in runningJobs" :key="item.id">
          <div v-if="item.progress > 0">
            <van-image src="/images/img-holder.png"></van-image>
            <div class="progress">
              <van-circle v-model:current-rate="item.progress" :rate="item.progress" :speed="100" :text="item.progress + '%'" :stroke-width="60" size="90px" />
            </div>
          </div>

          <div v-else class="task-in-queue">
            <div class="loading-wrapper">
              <div class="icon-container">
                <i class="iconfont icon-quick-start loading-icon"></i>
                <div class="pulse-ring"></div>
              </div>
              <div class="text-container">
                <span class="text">{{ loadingText }}</span>
                <span class="dots">
                  <span v-for="i in 3" :key="i" class="dot" :class="{ active: loadingDots >= i }"></span>
                </span>
              </div>
            </div>
          </div>

        </van-grid-item>
      </van-grid>
    </div>

    <h3>创作记录</h3>
    <div class="finish-job-list">
      <van-empty
        v-if="finishedJobs.length === 0"
        image="https://fastly.jsdelivr.net/npm/@vant/assets/custom-empty-image.png"
        image-size="80"
        description="暂无记录"
      />

      <van-list
        v-else
        v-model:error="error"
        v-model:loading="loading"
        :finished="finished"
        error-text="请求失败，点击重新加载"
        finished-text="没有更多了"
        @load="onLoad"
      >
        <van-grid :gutter="10" :column-num="2">
          <van-grid-item v-for="item in finishedJobs" :key="item.id">
            <div class="failed" v-if="item.progress === 101">
              <div class="title">任务失败</div>
              <div class="opt">
                <van-button size="small" @click="showErrMsg(item)">详情</van-button>
                <van-button type="primary" size="small" @click="retryTask(item)">重做</van-button>
                <van-button type="danger" @click="removeImage(item)" size="small">删除</van-button>
              </div>
            </div>
            <div class="job-item" v-else>
              <van-image :src="item['thumb_url']" :class="item['can_opt'] ? '' : 'upscale'" lazy-load @click="imageView(item)" fit="cover">
                <template v-slot:loading>
                  <van-loading type="spinner" size="20" />
                </template>
                <template v-slot:error>
                  <span style="margin-bottom: 20px">正在下载图片</span>
                  <van-loading type="circular" color="#1989fa" size="40" />
                </template>
              </van-image>

              <div class="remove">
                <el-button type="danger" :icon="Delete" @click="removeImage(item)" circle />
                <el-button type="warning" v-if="item.publish" @click="publishImage(item, false)" circle>
                  <i class="iconfont icon-cancel-share"></i>
                </el-button>
                <el-button type="success" @click="useAsInput(item)" circle title="以图生图">
                  <i class="iconfont icon-image"></i>
                </el-button>
                <el-button type="primary" @click="showPrompt(item)" circle>
                  <i class="iconfont icon-prompt"></i>
                </el-button>
                <!-- 下载图片按钮 -->
                <el-button type="primary" @click="downloadImage(item)" circle>
                  <i class="iconfont icon-download"></i>
                </el-button>
              </div>
            </div>
          </van-grid-item>
        </van-grid>
      </van-list>
    </div>
    <button style="display: none" class="copy-prompt" :data-clipboard-text="prompt" id="copy-btn">复制</button>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import { showConfirmDialog, showFailToast, showImagePreview, showNotify, showSuccessToast, showToast, showDialog, showLoadingToast, closeToast } from "vant";
import { httpGet, httpPost } from "@/utils/http";
import Compressor from "compressorjs";
import { getSessionId } from "@/store/session";
import { checkSession, getClientId, getSystemInfo } from "@/store/cache";
import { useRouter, useRoute } from "vue-router";
import { Delete } from "@element-plus/icons-vue";
import { showLoginDialog } from "@/utils/libs";
import Clipboard from "clipboard";
import { useSharedStore } from "@/store/sharedata";
import { generateContent } from "@/services/geminiService";

const rates = [
  { css: "square", value: "1:1", text: "1:1", img: "/images/mj/rate_1_1.png" },
  { css: "size16-9", value: "16:9", text: "16:9", img: "/images/mj/rate_16_9.png" },
  { css: "size9-16", value: "9:16", text: "9:16", img: "/images/mj/rate_9_16.png" },
];

const models = [{ text: "高清画质", value: "nano-banana-2-4k", img: "/images/mj/mj-v5.2.png" },{ text: "标准画质", value: "nano-banana", img: "/images/mj/mj-v6.png" }];
const imgList = ref([]);
const params = ref({
  client_id: getClientId(),
  task_type: "image",
  nano_mode: "txt2img",
  rate: rates[0].value,
  model: models[0].value,
  chaos: 0,
  stylize: 0,
  seed: 0,
  img_arr: [],
  raw: false,
  iw: 0,
  prompt: "",
  neg_prompt: "",
  tile: false,
  quality: 0,
  cref: "",
  sref: "",
  cw: 0,
});
const userId = ref(0);
const router = useRouter();
const route = useRoute();
const runningJobs = ref([]);
const finishedJobs = ref([]);
const power = ref(0);
const isLogin = ref(false);
const prompt = ref("");
const store = useSharedStore();
const clipboard = ref(null);
const loadingDots = ref(0);
let loadingTimer = null;

// 素材库相关
const selectedMaterial = ref(null);

// 匿名上传图片（上传码）相关
const showUploadImagePicker = ref(false);
const uploadCodeImages = ref([]);
const uploadImagesLoading = ref(false);
const selectedUploadImageUrls = ref([]);
const selectedUploadImageUrlsSet = computed(() => new Set(selectedUploadImageUrls.value));

// B 供应商（Gemini API）测试相关
const geminiApiKey = ref("");
const geminiLoading = ref(false);
const geminiTestResult = ref(null);

// 加载保存的 Gemini API Key
const loadGeminiApiKey = () => {
  try {
    const saved = localStorage.getItem("gemini_api_key");
    if (saved) {
      geminiApiKey.value = saved;
    } else {
      // 如果没有保存的值，使用硬编码的固定值
      geminiApiKey.value = "sk-l2szN1KNOpvnXyJcBj1O7TJSlNyGTrsQ7vvWIdBGdzKU6Cjr";
      // 同时保存到 localStorage
      localStorage.setItem("gemini_api_key", geminiApiKey.value);
    }
  } catch (e) {
    console.error("加载 API Key 失败:", e);
    // 即使出错也设置默认值，确保按钮可用
    geminiApiKey.value = "sk-l2szN1KNOpvnXyJcBj1O7TJSlNyGTrsQ7vvWIdBGdzKU6Cjr";
  }
};

// 保存 Gemini API Key
const saveGeminiApiKey = () => {

  //给定一个固定值，防止用户输入错误
  geminiApiKey.value = "sk-l2szN1KNOpvnXyJcBj1O7TJSlNyGTrsQ7vvWIdBGdzKU6Cjr";
  try {
    if (geminiApiKey.value) {
      localStorage.setItem("gemini_api_key", geminiApiKey.value);
    } else {
      localStorage.removeItem("gemini_api_key");
    }
  } catch (e) {
    console.error("保存 API Key 失败:", e);
  }
};

// 预览 Gemini 生成的图片
const previewGeminiImage = (imageUrl) => {
  showImagePreview([imageUrl]);
};

// 使用 Gemini API 生成图片（B 供应商测试）
const generateWithGemini = async () => {
  if (!geminiApiKey.value) {
    return showFailToast("请先输入 Gemini API Key");
  }

  // 验证提示词
  if (!params.value.prompt || params.value.prompt.trim() === "") {
    return showFailToast("提示词不能为空！");
  }

  geminiLoading.value = true;
  geminiTestResult.value = null;

  try {
    // 准备图片 URL 列表
    const imageUrls = imgList.value.map((img) => img.url);

    // 映射比例设置
    const aspectRatioMap = {
      "1:1": "1:1",
      "16:9": "16:9",
      "9:16": "9:16",
    };

    // 准备设置
    const settings = {
      modelName: "gemini-3-pro-image-preview",
      resolution: "2K", // 可以根据需要调整
      aspectRatio: aspectRatioMap[params.value.rate] || "Auto",
      useGrounding: false,
      enableThinking: false,
      customEndpoint: "https://api.kuai.host", // 默认端点，可以根据需要修改
    };

    // 调用 Gemini API
    const result = await generateContent(
      geminiApiKey.value,
      params.value.prompt,
      imageUrls,
      settings
    );

    geminiTestResult.value = {
      images: result.images || [],
      text: result.text || "",
      error: null,
    };

    if (result.images && result.images.length > 0) {
      showSuccessToast(`测试成功！生成了 ${result.images.length} 张图片`);
    } else {
      showFailToast("生成成功，但未返回图片");
    }
  } catch (error) {
    console.error("Gemini API 测试失败:", error);
    geminiTestResult.value = {
      images: [],
      text: "",
      error: error.message || "生成失败，请检查 API Key 和网络连接",
    };
    showFailToast(error.message || "生成失败");
  } finally {
    geminiLoading.value = false;
  }
};

// 生成动态的"生成中"文本
const loadingText = computed(() => {
  return `生成中`;
});

// 跳转到素材列表页面
const goToMaterialList = () => {
  // 如果已选择素材，传递当前选中的素材ID
  const query = selectedMaterial.value ? { materialId: selectedMaterial.value.id } : {};
  router.push({
    path: "/mobile/material-list",
    query
  });
};

// 跳转到提示词编辑页面
const goToPromptEditor = () => {
  // 保存当前的图片列表到 sessionStorage
  if (imgList.value && imgList.value.length > 0) {
    const imgListData = imgList.value.map(img => ({
      url: img.url,
      status: img.status || 'done'
    }));
    sessionStorage.setItem('mj_imgList_backup', JSON.stringify(imgListData));
  }
  
  router.push({
    path: "/mobile/prompt-editor",
    query: {
      prompt: params.value.prompt || ""
    }
  });
};

// 清除选中的素材
const clearMaterial = () => {
  selectedMaterial.value = null;
  params.value.prompt = "";
};

// 清除提示词
const clearPrompt = () => {
  params.value.prompt = "";
};

onMounted(() => {
  clipboard.value = new Clipboard(".copy-prompt");
  clipboard.value.on("success", () => {
    showNotify({ type: "success", message: "复制成功", duration: 1000 });
  });
  clipboard.value.on("error", () => {
    showNotify({ type: "danger", message: "复制失败", duration: 2000 });
  });

  checkSession()
    .then((user) => {
      power.value = user["power"];
      userId.value = user.id;
      isLogin.value = true;
      fetchRunningJobs();
      fetchFinishJobs(1);
    })
    .catch(() => {
      // router.push('/login')
    });

  // 启动加载动画定时器
  loadingTimer = setInterval(() => {
    loadingDots.value = (loadingDots.value + 1) % 4; // 0, 1, 2, 3 循环
  }, 500); // 每500ms切换一次

  store.addMessageHandler("mj", (data) => {
    if (data.channel !== "mj" || data.clientId !== getClientId()) {
      return;
    }
    if (data.body === "FINISH" || data.body === "FAIL") {
      page.value = 1;
      fetchFinishJobs(1);
    }
    fetchRunningJobs();
  });

  // 检查路由参数，如果有选中的素材，则加载
  checkRouteMaterial();
  // 检查路由参数，如果有提示词，则加载
  checkRoutePrompt();
  
  // 加载保存的 Gemini API Key
  loadGeminiApiKey();
});

// 检查路由参数中的素材信息
const checkRouteMaterial = () => {
  if (route.query.materialData) {
    try {
      const material = JSON.parse(route.query.materialData);
      selectMaterial(material);
      // 清除路由参数，避免刷新时重复加载
      router.replace({ path: route.path, query: {} });
    } catch (e) {
      console.error("解析素材数据失败:", e);
    }
  }
};

// 检查路由参数中的提示词信息
const checkRoutePrompt = () => {
  if (route.query.prompt !== undefined) {
    params.value.prompt = route.query.prompt || "";
    
    // 恢复之前保存的图片列表（从提示词编辑页面返回时）
    try {
      const imgListBackup = sessionStorage.getItem('mj_imgList_backup');
      if (imgListBackup) {
        const restoredImgList = JSON.parse(imgListBackup);
        // 恢复图片列表，确保每个图片都有正确的状态
        imgList.value = restoredImgList.map(img => ({
          url: img.url,
          status: img.status || 'done'
        }));
        sessionStorage.removeItem('mj_imgList_backup'); // 恢复后清除备份
      }
    } catch (e) {
      console.error("恢复图片列表失败:", e);
      // 如果解析失败，清除可能损坏的备份数据
      sessionStorage.removeItem('mj_imgList_backup');
    }
    
    // 清除路由参数，避免刷新时重复加载
    const newQuery = { ...route.query };
    delete newQuery.prompt;
    router.replace({ path: route.path, query: newQuery });
  }
};

// 监听路由变化，处理从素材列表页面返回的情况
watch(() => route.query.materialData, (materialData) => {
  if (materialData) {
    checkRouteMaterial();
  }
});

// 监听路由变化，处理从提示词编辑页面返回的情况
watch(() => route.query.prompt, (prompt) => {
  if (prompt !== undefined) {
    checkRoutePrompt();
  }
});

onUnmounted(() => {
  clipboard.value.destroy();
  store.removeMessageHandler("mj");
  // 清理加载动画定时器
  if (loadingTimer) {
    clearInterval(loadingTimer);
    loadingTimer = null;
  }
});

const mjPower = ref(1);
const mj2Power = ref(1);
const mjActionPower = ref(1);
getSystemInfo()
  .then((res) => {
    mjPower.value = res.data["mj_power"];
    mj2Power.value = res.data["mj_power2"];
    mjActionPower.value = res.data["mj_action_power"];
  })
  .catch((e) => {
    showNotify({ type: "danger", message: "获取系统配置失败：" + e.message });
  });

// 选择素材
const selectMaterial = (material) => {
  selectedMaterial.value = material;
  // 自动填充提示词（如果提示词为空，则填充素材的提示词）
  if (!params.value.prompt || params.value.prompt === "") {
    params.value.prompt = material.prompt || material.text || "";
  }
};

// 获取上传码图片列表（当前登录用户）
const fetchUploadCodeImages = async (pageNum = 1) => {
  if (!isLogin.value) {
    return;
  }
  uploadImagesLoading.value = true;
  try {
    const res = await httpGet("/api/upload_code/images", { page: pageNum, page_size: 60 });
    const pageData = res?.data;
    if (pageData && Array.isArray(pageData.items)) {
      uploadCodeImages.value = pageData.items;
    } else {
      uploadCodeImages.value = [];
    }
  } catch (e) {
    uploadCodeImages.value = [];
    showFailToast("获取匿名图片失败：" + (e.message || "未知错误"));
  } finally {
    uploadImagesLoading.value = false;
  }
};

const openUploadImagePicker = async () => {
  if (!isLogin.value) {
    return showLoginDialog(router);
  }
  showUploadImagePicker.value = true;
  selectedUploadImageUrls.value = [];
  await fetchUploadCodeImages(1);
};

const closeUploadImagePicker = () => {
  showUploadImagePicker.value = false;
};

const toggleSelectUploadImage = (url) => {
  const idx = selectedUploadImageUrls.value.indexOf(url);
  if (idx >= 0) {
    selectedUploadImageUrls.value.splice(idx, 1);
  } else {
    selectedUploadImageUrls.value.push(url);
  }
};

const confirmUploadImagePicker = () => {
  if (selectedUploadImageUrls.value.length === 0) {
    showFailToast("请先选择图片");
    return;
  }
  // 选择匿名图片后，直接覆盖输入图片列表
  imgList.value = selectedUploadImageUrls.value.map((u) => ({ url: u, status: "done" }));
  showUploadImagePicker.value = false;
  showSuccessToast(`已选择 ${selectedUploadImageUrls.value.length} 张图片`);
};

// 获取运行中的任务
const fetchRunningJobs = () => {
  if (!isLogin.value) {
    return;
  }
  httpGet(`/api/mj/jobs?finish=0&user_id=${userId.value}`)
    .then((res) => {
      const jobs = res.data.items;
      const _jobs = [];
      for (let i = 0; i < jobs.length; i++) {
        if (jobs[i].progress === -1) {
          showNotify({
            message: `任务执行失败：${jobs[i]["err_msg"]}`,
            type: "danger",
          });
          if (jobs[i].type === "image") {
            power.value += mjPower.value;
          } else {
            power.value += mjActionPower.value;
          }
          continue;
        }
        _jobs.push(jobs[i]);
      }
      runningJobs.value = _jobs;
    })
    .catch((e) => {
      showNotify({ type: "danger", message: "获取任务失败：" + e.message });
    });
};

const loading = ref(false);
const finished = ref(false);
const error = ref(false);
const page = ref(0);
const pageSize = ref(10);
const fetchFinishJobs = (pageNum) => {
  loading.value = true;
  // 获取已完成的任务
  httpGet(`/api/mj/jobs?finish=1&page=${pageNum}&page_size=${pageSize.value}`)
    .then((res) => {
      const jobs = res.data.items;
      for (let i = 0; i < jobs.length; i++) {
        if (jobs[i].type === "upscale" || jobs[i].type === "swapFace") {
          jobs[i]["thumb_url"] = jobs[i]["img_url"] + "?imageView2/1/w/480/h/600/q/75";
        } else {
          jobs[i]["thumb_url"] = jobs[i]["img_url"] + "?imageView2/1/w/480/h/480/q/75";
        }
      }
      if (jobs.length < pageSize.value) {
        finished.value = true;
      }
      if (pageNum === 1) {
        finishedJobs.value = jobs;
      } else {
        finishedJobs.value = finishedJobs.value.concat(jobs);
      }
      nextTick(() => (loading.value = false));
    })
    .catch((e) => {
      loading.value = false;
      error.value = true;
      showFailToast("获取任务失败：" + e.message);
    });
};

const onLoad = () => {
  page.value += 1;
  fetchFinishJobs(page.value);
};

// 切换图片比例
const changeRate = (item) => {
  params.value.rate = item.value;
};
// 切换模型
const changeModel = (item) => {
  params.value.model = item.value;
};

const imgKey = ref("");
const beforeUpload = (key) => {
  imgKey.value = key;
};

// 图片上传
const uploadImg = (file) => {
  file.status = "uploading";
  // 压缩图片并上传
  new Compressor(file.file, {
    quality: 0.6,
    success(result) {
      const formData = new FormData();
      formData.append("file", result, result.name);
      // 执行上传操作
      httpPost("/api/upload", formData)
        .then((res) => {
          file.url = res.data.url;
          if (imgKey.value !== "") {
            // 单张图片上传
            params.value[imgKey.value] = res.data.url;
            imgKey.value = "";
          }
          file.status = "done";
        })
        .catch((e) => {
          file.status = "failed";
          file.message = "上传失败";
          showFailToast("图片上传失败：" + e.message);
        });
    },
    error(err) {
      console.log(err.message);
    },
  });
};

const send = (url, index, item) => {
  httpPost(url, {
    client_id: getClientId(),
    index: index,
    channel_id: item.channel_id,
    message_id: item.message_id,
    message_hash: item.hash,
    session_id: getSessionId(),
    prompt: item.prompt,
  })
    .then(() => {
      showSuccessToast("任务推送成功，请耐心等待任务执行...");
      power.value -= mjActionPower.value;
      fetchRunningJobs();
    })
    .catch((e) => {
      showFailToast("任务推送失败：" + e.message);
    });
};

// 图片放大任务
const upscale = (index, item) => {
  send("/api/mj/upscale", index, item);
};

// 图片变换任务
const variation = (index, item) => {
  send("/api/mj/variation", index, item);
};

const generate = () => {
  if (!isLogin.value) {
    return showLoginDialog(router);
  }

  // 如果没有选择素材（专业模式），提示词不能为空
  if (!selectedMaterial.value && params.value.prompt === "") {
    return showFailToast("提示词不能为空！");
  }

  // 如果选择了素材（模板模式），但提示词为空且是图片任务，则提示
  if (selectedMaterial.value && params.value.prompt === "" && params.value.task_type === "image") {
    return showFailToast("提示词不能为空！");
  }
  
  // 根据是否上传图片自动判断模式
  params.value.nano_mode = imgList.value.length > 0 ? "img2img" : "txt2img";
  params.value.session_id = getSessionId();
  params.value.img_arr = imgList.value.map((img) => img.url);
  httpPost("/api/mj/image", params.value)
    .then(() => {
      showToast("绘画任务推送成功，请耐心等待任务执行");
      power.value -= mjPower.value;
      fetchRunningJobs();
      // 重置选择的素材和表单
      selectedMaterial.value = null;
      params.value.prompt = "";
      imgList.value = [];
    })
    .catch((e) => {
      showFailToast("任务推送失败：" + e.message);
    });
};

const removeImage = (item) => {
  showConfirmDialog({
    title: "删除提示",
    message: "此操作将会删除任务和图片，继续操作码?",
  })
    .then(() => {
      httpGet("/api/mj/remove", { id: item.id, user_id: item.user_id })
        .then(() => {
          showSuccessToast("任务删除成功");
          fetchFinishJobs(1);
        })
        .catch((e) => {
          showFailToast("任务删除失败：" + e.message);
        });
    })
    .catch(() => {
      showToast("您取消了操作");
    });
};
// 发布图片到作品墙
const publishImage = (item, action) => {
  let text = "图片发布";
  if (action === false) {
    text = "取消发布";
  }
  httpGet("/api/mj/publish", { id: item.id, action: action, user_id: item.user_id })
    .then(() => {
      showSuccessToast(text + "成功");
      item.publish = action;
    })
    .catch((e) => {
      showFailToast(text + "失败：" + e.message);
    });
};

const showPrompt = (item) => {
  prompt.value = item.prompt;
  
  // 从 task_info 中解析 full_prompt 获取 mode
  let mode = "txt2img";
  let imgArr = [];
  try {
    if (item.task_info) {
      // 解析 task_info 获取 task 对象
      const taskInfo = typeof item.task_info === 'string' 
        ? JSON.parse(item.task_info) 
        : item.task_info;
      
      // 从 task 对象中获取 full_prompt (即 Params 字段)
      if (taskInfo.full_prompt) {
        const fullPromptObj = typeof taskInfo.full_prompt === 'string' 
          ? JSON.parse(taskInfo.full_prompt) 
          : taskInfo.full_prompt;
        mode = fullPromptObj.mode || "txt2img";
      }
      
      // 从 task 对象中获取 img_arr
      if (taskInfo.img_arr && Array.isArray(taskInfo.img_arr)) {
        imgArr = taskInfo.img_arr;
      }
    }
  } catch (e) {
    console.error("解析 task_info 失败:", e);
  }
  
  // 构建消息内容
  let message = item.prompt;
  
  // 如果是图生图，添加原始图片
  if (mode === "img2img" && imgArr.length > 0) {
    const imagesHtml = imgArr.map((imgUrl, index) => {
      // 多图时使用网格布局，单图时居中显示
      const imageStyle = imgArr.length > 1 
        ? `max-width: 100%; border-radius: 8px; display: block;`
        : `max-width: 100%; border-radius: 8px; display: block; margin: 0 auto;`;
      
      const containerStyle = imgArr.length > 1
        ? `margin: 10px 5px; flex: 1; min-width: 0;`
        : `margin: 10px 0;`;
      
      const label = imgArr.length > 1 ? `<div style="margin-bottom: 5px; font-size: 12px; color: #666;">图片 ${index + 1}</div>` : '';
      
      return `<div style="${containerStyle}">${label}<img src="${imgUrl}" style="${imageStyle}" /></div>`;
    }).join('');
    
    const imagesContainerStyle = imgArr.length > 1
      ? `display: flex; flex-wrap: wrap; gap: 10px; margin-top: 10px;`
      : `margin-top: 10px;`;
    
    message = `<div><div style="margin-bottom: 10px;"><strong>提示词：</strong>${item.prompt}</div><div style="margin-top: 15px;"><strong>原始图片${imgArr.length > 1 ? ` (${imgArr.length}张)` : ''}：</strong><div style="${imagesContainerStyle}">${imagesHtml}</div></div></div>`;
  }
  
  showDialog({
    title: mode === "img2img" ? "图生图提示词" : "文生图提示词",
    message: message,
    confirmButtonText: "复制提示词",
    cancelButtonText: "关闭",
    allowHtml: true,
  })
    .then(() => {
      document.querySelector("#copy-btn").click();
    })
    .catch(() => {});
};

const showErrMsg = (item) => {
  showDialog({
    title: "错误详情",
    message: item["err_msg"],
  }).then(() => {
    // on close
  });
};

// 重做失败的任务
const retryTask = (item) => {
  if (!isLogin.value) {
    return showLoginDialog(router);
  }
  
  httpPost("/api/mj/retry", {
    id: item.id,
    user_id: item.user_id,
    client_id: getClientId(),
  })
    .then(() => {
      showSuccessToast("任务重试成功，请耐心等待任务执行");
      fetchRunningJobs();
    })
    .catch((e) => {
      showFailToast("任务重试失败：" + e.message);
    });
};

// 使用输出图片作为输入（以图生图）
const useAsInput = (item) => {
  if (!item.img_url) {
    showFailToast("图片地址不存在");
    return;
  }
  
  // 将当前图片添加到输入列表
  imgList.value = [{ url: item.img_url, status: 'done' }];
  
  // 恢复原始提示词（如果有）
  if (item.prompt) {
    params.value.prompt = item.prompt;
  }
  
  // 从 task_info 中恢复其他参数（如果有）
  try {
    if (item.task_info) {
      const taskInfo = typeof item.task_info === 'string' 
        ? JSON.parse(item.task_info) 
        : item.task_info;
      
      if (taskInfo.full_prompt) {
        const fullPromptObj = typeof taskInfo.full_prompt === 'string' 
          ? JSON.parse(taskInfo.full_prompt) 
          : taskInfo.full_prompt;
        
        if (fullPromptObj.rate) {
          params.value.rate = fullPromptObj.rate;
        }
        if (fullPromptObj.model) {
          params.value.model = fullPromptObj.model;
        }
      }
    }
  } catch (e) {
    console.error("解析任务参数失败:", e);
  }
  
  // 清除选中的素材
  selectedMaterial.value = null;
  
  // 滚动到顶部，让用户编辑提示词
  window.scrollTo({ top: 0, behavior: 'smooth' });
  
  showSuccessToast("已使用该图片作为输入，请编辑提示词后生成");
};

const imageView = (item) => {
  showImagePreview([item["img_url"]]);
};

// 下载图片
const downloadImage = async (item) => {
  try {
    const imageUrl = item["img_url"];
    if (!imageUrl) {
      showFailToast("图片地址不存在");
      return;
    }

    // 显示加载提示
    const loadingToast = showLoadingToast({
      message: "正在下载图片...",
      forbidClick: true,
      duration: 0,
    });

    // 通过后端代理下载图片，避免跨域问题
    const downloadUrl = `api/download?url=${encodeURIComponent(imageUrl)}`;
    
    // 使用 fetch 获取图片 blob
    const response = await fetch(downloadUrl);
    if (!response.ok) {
      throw new Error("下载失败");
    }
    
    const blob = await response.blob();
    
    // 创建下载链接
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    
    // 从 URL 中提取文件名，如果没有则使用默认名称
    const urlParts = imageUrl.split("/");
    let fileName = urlParts[urlParts.length - 1];
    // 移除查询参数
    fileName = fileName.split("?")[0];
    // 如果没有扩展名，添加 .png
    if (!fileName.includes(".")) {
      fileName = `image_${item.id || Date.now()}.png`;
    }
    link.download = fileName;
    
    // 触发下载
    document.body.appendChild(link);
    link.click();
    
    // 清理
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
    
    // 关闭加载提示
    closeToast();
    showSuccessToast("图片下载成功");
  } catch (error) {
    closeToast();
    console.error("下载图片失败:", error);
    showFailToast("图片下载失败：" + (error.message || "未知错误"));
  }
};

</script>

<style lang="stylus">
@import "@/assets/css/mobile/image-mj.styl"

.section-title
  display: flex
  justify-content: space-between
  align-items: center
  width: 100%
  position: relative
  
  > span
    flex-shrink: 0
    margin-right: 12px
    z-index: 1
  
  .material-search-btn,
  .prompt-search-btn
    margin-left: auto
    flex-shrink: 0
    white-space: nowrap
    z-index: 2
    position: relative
    min-width: 90px  // 确保按钮有最小宽度

.selected-material-card
  display: flex
  align-items: center
  padding: 16px
  background: linear-gradient(135deg, #ffffff 0%, #f8f9ff 100%)
  border-radius: 16px
  border: 2px solid #e8e9eb
  box-shadow: 0 4px 16px rgba(124, 92, 255, 0.08)
  margin: 12px 0
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1)
  position: relative
  overflow: hidden

  &::before
    content: ""
    position: absolute
    top: 0
    left: 0
    right: 0
    height: 3px
    background: linear-gradient(90deg, #7c5cff 0%, #4a9eff 100%)
    opacity: 0.6
  
  .material-preview
    margin-right: 16px
    flex-shrink: 0
    position: relative
    
    &::after
      content: ""
      position: absolute
      inset: 0
      border-radius: 12px
      padding: 2px
      background: linear-gradient(135deg, #7c5cff, #4a9eff)
      -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0)
      -webkit-mask-composite: xor
      mask-composite: exclude
      pointer-events: none
    
  .material-info
    flex: 1
    display: flex
    justify-content: space-between
    align-items: center
    min-width: 0
    
    .material-name
      font-size: 15px
      color: #1a1a1a
      font-weight: 600
      flex: 1
      margin-right: 12px
      overflow: hidden
      text-overflow: ellipsis
      white-space: nowrap
      letter-spacing: 0.2px
      
    .clear-material-btn
      flex-shrink: 0
      height: 28px
      padding: 0 12px
      border-radius: 14px
      font-size: 12px
      font-weight: 500
      border-color: #e8e9eb
      transition: all 0.2s ease

      &:active
        transform: scale(0.95)
        background: #f7f8fa

.material-placeholder-small
  width: 80px
  height: 80px
  display: flex
  align-items: center
  justify-content: center
  background: linear-gradient(135deg, #f7f8fa 0%, #eef0f3 100%)
  color: #969799
  border-radius: 12px
  border: 2px dashed #e8e9eb
  
  .iconfont
    font-size: 28px
    opacity: 0.6

.prompt-preview
  background: linear-gradient(135deg, #ffffff 0%, #f8f9ff 100%)
  border-radius: 12px
  padding: 16px
  margin-top: 12px
  border: 2px solid #e8e9eb
  box-shadow: 0 4px 16px rgba(124, 92, 255, 0.08)
  position: relative
  
  .prompt-text
    font-size: 14px
    color: #323233
    line-height: 1.6
    max-height: 120px
    overflow: hidden
    text-overflow: ellipsis
    display: -webkit-box
    -webkit-line-clamp: 5
    -webkit-box-orient: vertical
    margin-bottom: 12px
    word-break: break-word
    
  .clear-prompt-btn
    height: 28px
    padding: 0 12px
    border-radius: 14px
    font-size: 12px
    font-weight: 500
    border-color: #e8e9eb
    transition: all 0.2s ease
    
    &:active
      transform: scale(0.95)
      background: #f7f8fa

.prompt-placeholder
  background: #f7f8fa
  border-radius: 12px
  padding: 16px
  margin-top: 12px
  text-align: center
  border: 2px dashed #e8e9eb
  
  .placeholder-text
    font-size: 14px
    color: #969799

.test-btn
  margin-top: 12px

.gemini-test-result
  margin-top: 16px
  padding: 16px
  background: linear-gradient(135deg, #fff8e1 0%, #fff3c4 100%)
  border-radius: 12px
  border: 2px solid #ffc107
  
  .result-title
    font-size: 14px
    font-weight: 600
    color: #f57c00
    margin-bottom: 12px
  
  .result-images
    margin-bottom: 12px
    
    :deep(.van-image)
      border-radius: 8px
      overflow: hidden
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1)
  
  .result-text
    font-size: 13px
    color: #666
    line-height: 1.6
    margin-bottom: 8px
    word-break: break-word
  
  .result-error
    font-size: 13px
    color: #f44336
    line-height: 1.6
    word-break: break-word

.upload-image-picker
  height: 70vh
  display: flex
  flex-direction: column

  .picker-header
    padding: 12px 14px
    display: flex
    align-items: center
    justify-content: space-between
    border-bottom: 1px solid #f2f3f5

    .picker-title
      font-size: 15px
      font-weight: 600
      color: #1a1a1a

    .picker-actions
      display: flex
      gap: 8px

  .picker-body
    padding: 12px
    overflow: auto
    flex: 1

  .picker-grid
    display: grid
    grid-template-columns: repeat(3, 1fr)
    gap: 10px

  .picker-item
    position: relative
    border-radius: 10px
    overflow: hidden
    border: 2px solid transparent

    &.selected
      border-color: #1989fa

    .selected-mark
      position: absolute
      right: 6px
      top: 6px
      width: 24px
      height: 24px
      border-radius: 12px
      background: rgba(25, 137, 250, .9)
      display: flex
      align-items: center
      justify-content: center
</style>

