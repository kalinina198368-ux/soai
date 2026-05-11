<template>
  <div class="mobile-mj">
    <van-form @submit="generate">
      <div class="text-line section-title">图片比例</div>
      <div class="text-line">
        <van-row :gutter="10" justify="center">
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
      <div class="text-line section-title">模型选择</div>
      <div class="text-line">
        <van-row :gutter="12">
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
      <!-- <div class="text-line">
        <van-field label="创意度">
          <template #input>
            <van-slider v-model.number="params.chaos" :max="100" :step="1" @update:model-value="showToast('当前值：' + params.chaos)" />
          </template>
        </van-field>
      </div>

      <div class="text-line">
        <van-field label="风格化">
          <template #input>
            <van-slider v-model.number="params.stylize" :max="1000" :step="1" @update:model-value="showToast('当前值：' + params.stylize)" />
          </template>
        </van-field>
      </div> -->

      <!-- <div class="text-line">
        <van-field label="原始模式">
          <template #input>
            <van-switch v-model="params.raw" />
          </template>
        </van-field>
      </div> -->

      <div class="text-line">
        <van-field
          v-model="params.prompt"
          rows="3"
          autosize
          type="textarea"
          placeholder="图片创作描述（上传图片即为图生图，不上传即为文生图）"
        />
      </div>

      <div class="text-line">
        <van-uploader v-model="imgList" :after-read="uploadImg" />
      </div>
<!-- 
      <div class="text-line">
        <van-collapse v-model="activeColspan">
          <van-collapse-item title="反向提示词" name="neg_prompt">
            <van-field v-model="params.neg_prompt" rows="3" autosize type="textarea" placeholder="不想出现在图片上的元素(例如：树，建筑)" />
          </van-collapse-item>
        </van-collapse>
      </div> -->

      <div class="text-line pt-6">
        <el-tag>nb模型消耗{{ mjPower }}算力 |  nb2模型消耗{{ mj2Power }}算力, 当前算力：{{ power }}</el-tag>
      </div>

      <div class="text-line">
        <van-button round block type="primary" native-type="submit"> 立即生成 </van-button>
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
                <!-- <el-button type="success" v-else @click="publishImage(item, true)" circle>
                  <i class="iconfont icon-share-bold"></i>
                </el-button> -->
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
<!-- @cancel="showServiceDialog = false" -->
    <button style="display: none" class="copy-prompt" :data-clipboard-text="prompt" id="copy-btn">复制</button>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref } from "vue";
import { showConfirmDialog, showFailToast, showImagePreview, showNotify, showSuccessToast, showToast, showDialog, showLoadingToast, closeToast } from "vant";
import { httpGet, httpPost } from "@/utils/http";
import Compressor from "compressorjs";
import { getSessionId } from "@/store/session";
import { checkSession, getClientId, getSystemInfo } from "@/store/cache";
import { useRouter } from "vue-router";
import { Delete } from "@element-plus/icons-vue";
import { showLoginDialog } from "@/utils/libs";
import Clipboard from "clipboard";
import { useSharedStore } from "@/store/sharedata";

const activeColspan = ref([""]);

const rates = [
  { css: "square", value: "1:1", text: "方屏", img: "/images/mj/rate_1_1.png" },
  { css: "size16-9", value: "16:9", text: "横屏", img: "/images/mj/rate_16_9.png" },
  { css: "size9-16", value: "9:16", text: "竖屏", img: "/images/mj/rate_9_16.png" },
];


const models = [{ text: "nb2(高质)", value: "nano-banana-2-4k", img: "/images/mj/mj-v5.2.png" },{ text: "nb(标质)", value: "nano-banana", img: "/images/mj/mj-v6.png" }];
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
const runningJobs = ref([]);
const finishedJobs = ref([]);
const power = ref(0);
const isLogin = ref(false);
const prompt = ref("");
const store = useSharedStore();
const clipboard = ref(null);
const loadingDots = ref(0);
let loadingTimer = null;

// 生成动态的"生成中"文本
const loadingText = computed(() => {
  //const dots = "。".repeat(loadingDots.value);
  return `生成中`;
});

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

// 获取运行中的任务
const fetchRunningJobs = (userId) => {
  httpGet(`/api/mj/jobs?finish=0&user_id=${userId}`)
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
const fetchFinishJobs = (page) => {
  loading.value = true;
  // 获取已完成的任务
  httpGet(`/api/mj/jobs?finish=1&page=${page}&page_size=${pageSize.value}`)
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
      if (page === 1) {
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

  if (params.value.prompt === "" && params.value.task_type === "image") {
    return showFailToast("请输入绘画提示词！");
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
   // const downloadUrl = `http://localhost:5678/api/download?url=${encodeURIComponent(imageUrl)}`;
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
</style>
