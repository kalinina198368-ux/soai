<template>
  <div class="app-background">
    <div class="mobile-chat" v-loading="loading" element-loading-text="正在连接会话...">
      <van-nav-bar>
        <template #title>
          <div class="chat-title-wrapper">
            <span class="chat-title">{{ title || "智能助手" }}</span>
            <span class="chat-subtitle">
              <van-icon :name="loading ? 'clock-o' : 'success'" />
              {{ navSubtitle }}
            </span>
          </div>
        </template>
        <template #left>
          <span class="setting">
            <van-icon name="bars" @click="openHistory" />
          </span>
        </template>
        <template #right>
          <div class="nav-quick-actions">
            <div
              v-for="action in navQuickActions"
              :key="action.route"
              class="nav-action-chip"
              :style="{ background: action.background }"
              @click="handleNavAction(action)"
            >
              <van-icon :name="action.icon" />
              <span>{{ action.label }}</span>
            </div>
          </div>
        </template>
      </van-nav-bar>

      <!-- <van-share-sheet v-model:show="showShare" title="立即分享给好友" :options="shareOptions" @select="shareChat" /> -->

      <div class="chat-list-wrapper">
        <div id="message-list-box" class="message-list-box" :style="{ paddingBottom: bottomPadding + 'px' }">
          <van-list v-model:error="error" :finished="finished" error-text="请求失败，点击重新加载" @load="onLoad">
            <van-cell v-for="item in chatData" :key="item" :border="false" class="message-line">
              <chat-prompt v-if="item.type === 'prompt'" :content="item.content" :icon="item.icon" />
              <chat-reply v-else-if="item.type === 'reply'" :content="item.content" :icon="item.icon" :org-content="item.orgContent" />
            </van-cell>
          </van-list>
        </div>
      </div>

      <div class="chat-box-wrapper" ref="bottomBoxRef">
        <van-cell-group inset style="--van-cell-background: var(--van-cell-background-light)">
          <!-- 模型选择小浮块 -->
          <div v-if="models.length" class="model-chip-wrapper">
            <div
              v-for="model in models"
              :key="model.id"
              class="model-chip"
              :class="{ active: modelId === model.id }"
              @click="selectModel(model.id)"
            >
              {{ model.text }}
            </div>
          </div>

          <!-- 图片预览区域 -->
          <div v-if="uploadedImages.length > 0" class="image-preview-container">
            <div v-for="(image, index) in uploadedImages" :key="index" class="image-preview-item">
              <van-image
                :src="image.url"
                fit="cover"
                width="60"
                height="60"
                radius="8"
                @click="previewImage(image.url)"
              />
              <van-icon 
                name="cross" 
                class="remove-image-icon" 
                @click="removeImage(index)"
              />
            </div>
          </div>

          <!-- 视频预览区域 -->
          <div v-if="uploadedVideos.length > 0" class="video-preview-container">
            <div v-for="(video, index) in uploadedVideos" :key="index" class="video-preview-item">
              <video
                :src="video.url"
                width="60"
                height="60"
                style="object-fit: cover; border-radius: 8px;"
                @click="previewVideo(video.url)"
              />
              <div class="video-play-icon">
                <van-icon name="play-circle-o" size="20" />
              </div>
              <van-icon 
                name="cross" 
                class="remove-image-icon" 
                @click="removeVideo(index)"
              />
            </div>
          </div>
          
          <van-field 
            v-model="prompt" 
            center 
            clearable 
            placeholder="输入你的问题"
            type="textarea"
            :rows="1"
            :autosize="{ maxRows: 4 }"
            @keyup.enter="handleEnterKey"
          >
            <template #left-icon>
              <div style="display: flex; gap: 12px; align-items: center;">
                <van-icon  v-if="isGemini3ProPreview"  name="photo-o" @click="uploadImages" class="upload-icon" />
                <van-icon v-if="isGemini3ProPreview" name="video-o" @click="uploadVideos" class="upload-icon" />
              </div>
            </template>


            <template #button>
              <van-button size="small" type="primary" @click="sendMessage" :disabled="!canSend">发送</van-button>
            </template>
            <template #extra>
              <div class="icon-box">
                <van-icon v-if="showStopGenerate" name="stop-circle-o" @click="stopGenerate" />
                <van-icon v-if="showReGenerate" name="play-circle-o" @click="reGenerate" />
              </div>
            </template>
          </van-field>
        </van-cell-group>
      </div>
    </div>

    <button id="copy-link-btn" style="display: none" :data-clipboard-text="url">复制链接地址</button>

    <!--    <van-overlay :show="showMic" z-index="100">-->
    <!--      <div class="mic-wrapper">-->
    <!--        <div class="image">-->
    <!--          <van-image-->
    <!--              width="100"-->
    <!--              height="100"-->
    <!--              src="/images/mic.gif"-->
    <!--          />-->
    <!--        </div>-->
    <!--        <van-button type="success" @click="stopVoice">说完了</van-button>-->
    <!--      </div>-->
    <!--    </van-overlay>-->
  </div>

  <!-- 历史会话列表，仿 DeepSeek 左侧浮出效果 -->
  <van-popup
    v-model:show="showHistory"
    position="left"
    class="history-popup"
    :style="{ width: '80%', height: '100%' }"
  >
    <div class="history-container">
      <van-nav-bar
        title="会话列表"
        left-arrow
        @click-left="showHistory = false"
      >
        <template #right>
          <van-button
            size="small"
            type="primary"
            class="history-new-chat-btn"
            @click="startNewChat"
          >
            <span class="history-new-chat-text">新建会话</span>
          </van-button>
        </template>
      </van-nav-bar>

      <div class="history-content">
        <van-list
          v-model:error="historyError"
          v-model:loading="historyLoading"
          :finished="historyFinished"
          error-text="请求失败，点击重新加载"
          finished-text="没有更多了"
          @load="loadHistoryList"
        >
          <van-swipe-cell v-for="chat in historyChats" :key="chat.chat_id">
            <van-cell is-link @click="openChat(chat)">
              <div class="history-item">
                <van-image
                  v-if="chat.icon"
                  :src="chat.icon"
                  round
                  width="32"
                  height="32"
                />
                <div class="history-title van-ellipsis">
                  {{ chat.title || '未命名会话' }}
                </div>
              </div>
            </van-cell>
            <template #right>
              <van-button
                square
                text="修改"
                type="primary"
                @click.stop="editChat(chat)"
              />
              <van-button
                square
                text="删除"
                type="danger"
                @click.stop="removeChat(chat)"
              />
            </template>
          </van-swipe-cell>
        </van-list>
      </div>
    </div>
  </van-popup>
  <van-dialog
    v-model:show="showEditChat"
    title="修改对话标题"
    show-cancel-button
    class="dialog"
    @confirm="saveTitle"
  >
    <van-field
      v-model="tmpChatTitle"
      label=""
      placeholder="请输入对话标题"
      class="field"
    />
  </van-dialog>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import { showImagePreview, showNotify, showToast, showSuccessToast, showFailToast, showConfirmDialog } from "vant";
import { useRouter } from "vue-router";
import { processContent, randString, renderInputText, UUID } from "@/utils/libs";
import { httpGet, httpPost } from "@/utils/http";
import hl from "highlight.js";
import "highlight.js/styles/a11y-dark.css";
import ChatPrompt from "@/components/mobile/ChatPrompt.vue";
import ChatReply from "@/components/mobile/ChatReply.vue";
import { checkSession, getClientId } from "@/store/cache";
import Clipboard from "clipboard";
import { showMessageError } from "@/utils/dialog";
import { useSharedStore } from "@/store/sharedata";
import emoji from "markdown-it-emoji";
import mathjaxPlugin from "markdown-it-mathjax3";
import MarkdownIt from "markdown-it";
import FileList from "@/components/FileList.vue";
const router = useRouter();

const navQuickActions = [
  {
    icon: "photo-o",
    label: "",
    route: "/mobile/image",
    background: "linear-gradient(120deg, #fdfbfb 0%, #ebedee 100%)",
  },
  {
    icon: "video-o",
    label: "",
    route: "/mobile/index",
    background: "linear-gradient(120deg, #fff8e1 0%, #ffe0b2 100%)",
  },
];

const handleNavAction = (action) => {
  if (!action?.route) {
    return;
  }
  router.push(action.route);
};

const roles = ref([]);
const roleId = ref(parseInt(router.currentRoute.value.query["role_id"]));
const role = ref({});
const models = ref([]);
// 当前选中的模型 ID（后端内部 ID）
const modelId = ref(parseInt(router.currentRoute.value.query["model_id"]));
// 当前选中的模型显示名称
const modelValue = ref("");
// 会话详情里返回的实际模型标识（例如 "gpt-4o"）
const detailModelValue = ref("");
const title = ref(router.currentRoute.value.query["title"]);
const chatId = ref(router.currentRoute.value.query["chat_id"]);
const loginUser = ref(null);
// const showMic = ref(false)

// 历史会话相关
const showHistory = ref(false);
const historyChats = ref([]);
const historyLoading = ref(false);
const historyFinished = ref(false);
const historyError = ref(false);

// 历史会话编辑相关
const showEditChat = ref(false);
const currentChatItem = ref(null);
const tmpChatTitle = ref("");

const editChat = (chat) => {
  showEditChat.value = true;
  currentChatItem.value = chat;
  tmpChatTitle.value = chat.title;
};

const saveTitle = () => {
  if (!currentChatItem.value) {
    return;
  }
  httpPost("/api/chat/update", {
    chat_id: currentChatItem.value.chat_id,
    title: tmpChatTitle.value,
  })
    .then(() => {
      showSuccessToast("操作成功！");
      currentChatItem.value.title = tmpChatTitle.value;
      // 如果当前正在查看的会话，就是被修改的这个，也同步更新顶部标题
      if (chatId.value === currentChatItem.value.chat_id) {
        title.value = tmpChatTitle.value;
      }
    })
    .catch((e) => {
      showFailToast("操作失败：" + e.message);
    });
};

const removeChat = (chat) => {
  showConfirmDialog({
    title: "操作提示",
    message: "确定要删除该会话吗？",
  })
    .then(() => {
      httpGet("/api/chat/remove?chat_id=" + chat.chat_id)
        .then(() => {
          showSuccessToast("会话已删除");
          historyChats.value = historyChats.value.filter(
            (item) => item.chat_id !== chat.chat_id
          );
          // 如果删除的是当前会话，则跳转到一个新会话
          if (chatId.value === chat.chat_id) {
            router.push({ path: "/mobile/chat/session" });
          }
        })
        .catch((e) => {
          showFailToast("操作失败：" + e.message);
        });
    })
    .catch(() => {
      // 用户取消
    });
};

// 图片上传相关
const uploadedImages = ref([]);
// 视频上传相关
const uploadedVideos = ref([]);
const isUploading = ref(false);

// 底部固定区域高度，用于给消息列表留出可见空间，避免被输入框和模型浮框遮住
const bottomBoxRef = ref(null);
const bottomPadding = ref(120);

const updateBottomPadding = () => {
  nextTick(() => {
    if (bottomBoxRef.value && bottomBoxRef.value.offsetHeight) {
      // 在实际高度基础上再多预留一些空间，保证最后一条消息完全露出
      bottomPadding.value = bottomBoxRef.value.offsetHeight + 24;
    }
  });
};

// 模型列表变化时（例如接口返回后出现模型小浮块），重新计算底部留白
watch(
  () => models.value.length,
  () => updateBottomPadding()
);

checkSession()
  .then((user) => {
    loginUser.value = user;
  })
  .catch(() => {

   // return showLoginDialog(router);

    router.push("/mobile/login");
  });

const loadModels = () => {
  // 加载模型
  httpGet("/api/model/list")
    .then((res) => {
      models.value = res.data;
      for (let i = 0; i < models.value.length; i++) {
        models.value[i].text = models.value[i].name;
        // 保留后端返回的实际模型标识（例如 "gpt-4o"）
        models.value[i].mValue = models.value[i].value;
        models.value[i].value = models.value[i].id;
      }
      // 如果会话详情接口返回了具体模型（如 "gpt-4o"），优先根据它匹配模型
      if (detailModelValue.value) {
        const matched = models.value.find(
          (m) =>
            m.mValue === detailModelValue.value ||
            m.name === detailModelValue.value
        );
        if (matched) {
          modelId.value = matched.id;
        }
      }
      // 如果依然没有选中模型，则回退到第一个模型
      if (!modelId.value && models.value.length) {
        modelId.value = models.value[0].id;
      }
      modelValue.value = getModelName(modelId.value);
      // 加载角色列表
      httpGet(`/api/app/list/user`, { id: roleId.value })
        .then((res) => {
          roles.value = res.data;
          if (!roleId.value) {
            roleId.value = roles.value[0]["id"];
          }
          // build data for role picker
          for (let i = 0; i < roles.value.length; i++) {
            roles.value[i].text = roles.value[i].name;
            roles.value[i].value = roles.value[i].id;
            roles.value[i].helloMsg = roles.value[i].hello_msg;
          }
          role.value = getRoleById(roleId.value);
          loadChatHistory();
        })
        .catch((e) => {
          showNotify({ type: "danger", message: "获取聊天角色失败: " + e.messages });
        });
    })
    .catch((e) => {
      showNotify({ type: "danger", message: "加载模型失败: " + e.message });
    });
};

const initChat = () => {
  if (chatId.value) {
    httpGet(`/api/chat/detail?chat_id=${chatId.value}`)
      .then((res) => {
        title.value = res.data.title;
        // 记录会话对应的模型 ID 和模型标识（例如 "gpt-4o"）
        modelId.value = res.data.model_id;
        detailModelValue.value = res.data.model || "";
        roleId.value = res.data.role_id;
        loadModels();
      })
      .catch(() => {
        loadModels();
      });
  } else {
    title.value = "新建对话";
    chatId.value = UUID();
    loadModels();
  }
};

initChat();

watch(
  () => router.currentRoute.value.query.chat_id,
  (newChatId, oldChatId) => {
    if (newChatId === oldChatId) {
      return;
    }
    chatId.value = newChatId || null;
    initChat();
  }
);

const chatData = ref([]);
const loading = ref(false);
const finished = ref(false);
const error = ref(false);
const navSubtitle = computed(() => {
  if (loading.value) {
    return "正在连接会话...";
  }
  return modelValue.value ? `模型 · ${modelValue.value}` : "准备就绪";
});

// 判断当前模型是否为 gemini-3-pro-preview（支持视频描述）
const isGemini3ProPreview = computed(() => {
  if (!modelId.value || !models.value.length) {
    return false;
  }
  const currentModel = models.value.find(m => m.id === modelId.value);
  return currentModel && currentModel.mValue === 'gemini-3-flash-preview';
});
const store = useSharedStore();
const url = ref(location.protocol + "//" + location.host + "/mobile/chat/export?chat_id=" + chatId.value);
const md = new MarkdownIt({
  breaks: true,
  html: true,
  linkify: true,
  typographer: true,
  highlight: function (str, lang) {
    const codeIndex = parseInt(Date.now()) + Math.floor(Math.random() * 10000000);
    // 显示复制代码按钮
    const copyBtn = `<span class="copy-code-mobile" data-clipboard-action="copy" data-clipboard-target="#copy-target-${codeIndex}">复制</span>
<textarea style="position: absolute;top: -9999px;left: -9999px;z-index: -9999;" id="copy-target-${codeIndex}">${str.replace(
      /<\/textarea>/g,
      "&lt;/textarea>"
    )}</textarea>`;
    if (lang && hl.getLanguage(lang)) {
      const langHtml = `<span class="lang-name">${lang}</span>`;
      // 处理代码高亮
      const preCode = hl.highlight(lang, str, true).value;
      // 将代码包裹在 pre 中
      return `<pre class="code-container"><code class="language-${lang} hljs">${preCode}</code>${copyBtn} ${langHtml}</pre>`;
    }

    // 处理代码高亮
    const preCode = md.utils.escapeHtml(str);
    // 将代码包裹在 pre 中
    return `<pre class="code-container"><code class="language-${lang} hljs">${preCode}</code>${copyBtn}</pre>`;
  },
});
md.use(mathjaxPlugin);
md.use(emoji);
onMounted(() => {
  // 不再需要动态计算高度，因为使用固定定位
  // winHeight.value = window.innerHeight - navBarRef.value.$el.offsetHeight - bottomBarRef.value.$el.offsetHeight - 70;
  updateBottomPadding();
  window.addEventListener('resize', updateBottomPadding);

  const clipboard = new Clipboard(".content-mobile,.copy-code-mobile,#copy-link-btn");
  clipboard.on("success", (e) => {
    e.clearSelection();
    showNotify({ type: "success", message: "复制成功", duration: 1000 });
  });
  clipboard.on("error", () => {
    showNotify({ type: "danger", message: "复制失败", duration: 2000 });
  });

  store.addMessageHandler("chat", (data) => {
    if (data.channel !== "chat" || data.clientId !== getClientId()) {
      return;
    }

    if (data.type === "error") {
      showMessageError(data.body);
      return;
    }

    if (isNewMsg.value) {
      chatData.value.push({
        type: "reply",
        id: randString(32),
        icon: role.value.icon,
        content: md.render(processContent(data.body)),
      });
      if (!title.value) {
        title.value = previousText.value;
      }
      lineBuffer.value = data.body;
      isNewMsg.value = false;
    } else if (data.type === "end") {
      // 消息接收完毕
      enableInput();
      lineBuffer.value = ""; // 清空缓冲
      isNewMsg.value = true;
    } else {
      lineBuffer.value += data.body;
      const reply = chatData.value[chatData.value.length - 1];
      reply["orgContent"] = lineBuffer.value;
      reply["content"] = md.render(processContent(lineBuffer.value));

      nextTick(() => {
        hl.configure({ ignoreUnescapedHTML: true });
        const lines = document.querySelectorAll(".message-line");
        const blocks = lines[lines.length - 1].querySelectorAll("pre code");
        blocks.forEach((block) => {
          hl.highlightElement(block);
        });
        updateBottomPadding();
        scrollListBox();

        const items = document.querySelectorAll(".message-line");
        const imgs = items[items.length - 1].querySelectorAll("img");
        for (let i = 0; i < imgs.length; i++) {
          if (!imgs[i].src) {
            continue;
          }
          imgs[i].addEventListener("click", (e) => {
            e.stopPropagation();
            showImagePreview([imgs[i].src]);
          });
        }
      });
    }
  });
});

onUnmounted(() => {
  store.removeMessageHandler("chat");
  window.removeEventListener('resize', updateBottomPadding);
});

const newChat = (item) => {
  showPicker.value = false;
  const options = item.selectedOptions;
  roleId.value = options[0].value;
  modelId.value = options[1].value;
  modelValue.value = getModelName(modelId.value);
  chatId.value = UUID();
  chatData.value = [];
  role.value = getRoleById(roleId.value);
  title.value = "新建对话";
  loadChatHistory();
};

const onLoad = () => {
  // checkSession().then(() => {
  //   connect()
  // }).catch(() => {
  // })
};

const loadChatHistory = () => {
  // 切换会话或重新加载时，先清空当前聊天数据与状态
  chatData.value = [];
  finished.value = false;
  error.value = false;

  httpGet("/api/chat/history?chat_id=" + chatId.value)
    .then((res) => {
      const role = getRoleById(roleId.value);
      // 加载状态结束
      finished.value = true;
      const data = res.data;
      if (data.length === 0) {
        chatData.value.push({
          type: "reply",
          id: randString(32),
          icon: role.icon,
          content: role.hello_msg,
          orgContent: role.hello_msg,
        });
        return;
      }

      for (let i = 0; i < data.length; i++) {
        if (data[i].type === "prompt") {
          // 处理用户消息中的图片URL
          data[i].content = md.render(processContent(data[i].content));
          chatData.value.push(data[i]);
          continue;
        }

        data[i].orgContent = data[i].content;
        data[i].content = md.render(processContent(data[i].content));
        chatData.value.push(data[i]);
      }

      nextTick(() => {
        hl.configure({ ignoreUnescapedHTML: true });
        const blocks = document.querySelector("#message-list-box").querySelectorAll("pre code");
        blocks.forEach((block) => {
          hl.highlightElement(block);
        });

        scrollListBox();
      });
    })
    .catch(() => {
      error.value = true;
    });
};

// 创建 socket 连接
const prompt = ref("");
const showStopGenerate = ref(false); // 停止生成
const showReGenerate = ref(false); // 重新生成
const previousText = ref(""); // 上一次提问
const lineBuffer = ref(""); // 输出缓冲行
const canSend = ref(true);
const isNewMsg = ref(true);
const stream = ref(store.chatStream);
watch(
  () => store.chatStream,
  (newValue) => {
    stream.value = newValue;
  }
);

// 监听图片列表变化，动态更新底部留白
watch(
  () => uploadedImages.value.length,
  () => updateBottomPadding()
);

// 监听视频列表变化，动态更新底部留白
watch(
  () => uploadedVideos.value.length,
  () => updateBottomPadding()
);
// const connect = function () {
//   // 初始化 WebSocket 对象
//   const _sessionId = getSessionId();
//   let host = process.env.VUE_APP_WS_HOST
//   if (host === '') {
//     if (location.protocol === 'https:') {
//       host = 'wss://' + location.host;
//     } else {
//       host = 'ws://' + location.host;
//     }
//   }
//   const _socket = new WebSocket(host + `/api/chat/new?session_id=${_sessionId}&role_id=${roleId.value}&chat_id=${chatId.value}&model_id=${modelId.value}&token=${getUserToken()}`);
//   _socket.addEventListener('open', () => {
//     loading.value = false
//     previousText.value = '';
//     canSend.value = true;
//
//     if (loadHistory.value) { // 加载历史消息
//      loadChatHistory()
//     }
//   });
//
//   _socket.addEventListener('message', event => {
//     if (event.data instanceof Blob) {
//       const reader = new FileReader();
//       reader.readAsText(event.data, "UTF-8");
//       reader.onload = () => {
//         const data = JSON.parse(String(reader.result));
//         if (data.type === 'error') {
//           showMessageError(data.message)
//           return
//         }
//
//         if (isNewMsg.value && data.type !== 'end') {
//           chatData.value.push({
//             type: "reply",
//             id: randString(32),
//             icon: role.value.icon,
//             content: data.content
//           });
//           if (!title.value) {
//             title.value = previousText.value
//           }
//           lineBuffer.value = data.content;
//           isNewMsg.value = false
//         } else if (data.type === 'end') { // 消息接收完毕
//           enableInput()
//           lineBuffer.value = ''; // 清空缓冲
//           isNewMsg.value = true
//         } else {
//           lineBuffer.value += data.content;
//           const reply = chatData.value[chatData.value.length - 1]
//           reply['orgContent'] = lineBuffer.value;
//           reply['content'] = md.render(processContent(lineBuffer.value));
//
//           nextTick(() => {
//             hl.configure({ignoreUnescapedHTML: true})
//             const lines = document.querySelectorAll('.message-line');
//             const blocks = lines[lines.length - 1].querySelectorAll('pre code');
//             blocks.forEach((block) => {
//               hl.highlightElement(block)
//             })
//             scrollListBox()
//
//             const items = document.querySelectorAll('.message-line')
//             const imgs = items[items.length - 1].querySelectorAll('img')
//             for (let i = 0; i < imgs.length; i++) {
//               if (!imgs[i].src) {
//                 continue
//               }
//               imgs[i].addEventListener('click', (e) => {
//                 e.stopPropagation()
//                 showImagePreview([imgs[i].src]);
//               })
//             }
//           })
//         }
//
//       };
//     }
//
//   });
//
//   _socket.addEventListener('close', () => {
//     // 停止发送消息
//     canSend.value = true
//     loadHistory.value = false
//     // 重连
//     connect()
//   });
//
//   socket.value = _socket;
// }

const disableInput = (force) => {
  canSend.value = false;
  showReGenerate.value = false;
  showStopGenerate.value = !force;
};

const enableInput = () => {
  canSend.value = true;
  showReGenerate.value = previousText.value !== "";
  showStopGenerate.value = false;
};

// 将聊天框的滚动条滑动到最底部
const scrollListBox = () => {
  document.getElementById("message-list-box").scrollTo(0, document.getElementById("message-list-box").scrollHeight + 46);
};

const sendMessage = () => {
  if (canSend.value === false) {
    showToast("AI 正在作答中，请稍后...");
    return;
  }

  if (store.socket.conn.readyState !== WebSocket.OPEN) {
    showToast("连接断开，正在重连...");
    return;
  }

  if (prompt.value.trim().length === 0 && uploadedImages.value.length === 0 && uploadedVideos.value.length === 0) {
    showToast("请输入需要 AI 回答的问题或上传图片/视频");
    return false;
  }

  // 构建消息内容 - 将图片URL直接放在文本中，让后端ExtractImgURLs函数提取
  let messageContent = prompt.value;
  let messageDisplay = prompt.value;
  
  // 如果有图片，将图片URL添加到消息中（后端会提取这些URL并转换为image_url格式）
  if (uploadedImages.value.length > 0) {
    const imageUrls = uploadedImages.value.map(img => img.url);
    messageContent += '\n' + imageUrls.join('\n');
    messageDisplay += '\n' + imageUrls.join('\n');
  }
  
  // 如果有视频，将视频URL添加到消息中
  if (uploadedVideos.value.length > 0) {
    const videoUrls = uploadedVideos.value.map(video => video.url);
    messageContent += '\n' + videoUrls.join('\n');
    messageDisplay += '\n' + videoUrls.join('\n');
  }

  // 追加消息到聊天记录
  const processedText = processContent(messageDisplay);
  const renderedContent = md.render(processedText);
  
  chatData.value.push({
    type: "prompt",
    id: randString(32),
    icon: loginUser.value.avatar,
    content: renderedContent,
    created_at: new Date().getTime(),
  });

  nextTick(() => {
    scrollListBox();
  });

  disableInput(false);
  store.socket.conn.send(
    JSON.stringify({
      channel: "chat",
      type: "text",
      body: {
        role_id: roleId.value,
        model_id: modelId.value,
        chat_id: chatId.value,
        content: messageContent,
        stream: stream.value,
      },
    })
  );
  previousText.value = messageContent;
  prompt.value = "";
  uploadedImages.value = []; // 清空已上传的图片
  uploadedVideos.value = []; // 清空已上传的视频
  return true;
};

const stopGenerate = () => {
  showStopGenerate.value = false;
  httpGet("/api/chat/stop?session_id=" + getClientId()).then(() => {
    enableInput();
  });
};

const reGenerate = () => {
  disableInput(false);
  const text = "重新生成上述问题的答案：" + previousText.value;
  // 追加消息
  chatData.value.push({
    type: "prompt",
    id: randString(32),
    icon: loginUser.value.avatar,
    content: md.render(processContent(text)),
  });
  store.socket.conn.send(
    JSON.stringify({
      channel: "chat",
      type: "text",
      body: {
        role_id: roleId.value,
        model_id: modelId.value,
        chat_id: chatId.value,
        content: previousText.value,
        stream: stream.value,
      },
    })
  );
};

const showShare = ref(false);
const shareOptions = [
  { name: "微信", icon: "wechat" },
  { name: "复制链接", icon: "link" },
];
const shareChat = (option) => {
  showShare.value = false;
  if (option.icon === "wechat") {
    showToast({ message: "当前会话已经导出，请通过浏览器或者微信的自带分享功能分享给好友", duration: 5000 });
    router.push({
      path: "/mobile/chat/export",
      query: { title: title.value, chat_id: chatId.value, role: role.value.name, model: modelValue.value },
    });
  } else if (option.icon === "link") {
    document.getElementById("copy-link-btn").click();
  }
};

// const goHome = () => {
//   router.push({ path: "/mobile/chat" });
// };

const goHome = () => {
  router.push({ path: "/mobile/profile" });
};

// 选择模型（模型小浮块点击）
const selectModel = (id) => {
  modelId.value = id;
  modelValue.value = getModelName(id);
};

// 打开历史会话抽屉
const openHistory = () => {
  showHistory.value = true;
  // 每次打开时都刷新一次列表，确保能看到最新会话
  loadHistoryList();
};

// 加载历史会话列表
const loadHistoryList = () => {
  if (!loginUser.value || historyLoading.value) return;
  historyLoading.value = true;
  httpGet("/api/chat/list?user_id=" + loginUser.value.id)
    .then((res) => {
      historyChats.value = res.data || [];
      historyFinished.value = true;
      historyLoading.value = false;
    })
    .catch(() => {
      historyError.value = true;
      historyLoading.value = false;
      showNotify({ type: "danger", message: "加载会话列表失败" });
    });
};

// 在历史列表中打开某个会话
const openChat = (chat) => {
  showHistory.value = false;
  router.push({
    path: "/mobile/chat/session",
    query: { chat_id: chat.chat_id },
  });
};

// 从历史弹窗中新建会话
const startNewChat = () => {
  showHistory.value = false;
  router.push({ path: "/mobile/chat/session" });
};

const getRoleById = function (rid) {
  for (let i = 0; i < roles.value.length; i++) {
    if (roles.value[i]["id"] === rid) {
      return roles.value[i];
    }
  }
  return null;
};

const getModelName = (model_id) => {
  for (let i = 0; i < models.value.length; i++) {
    if (models.value[i].id === model_id) {
      return models.value[i].text;
    }
  }
  return "";
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

// 视频上传相关函数
const uploadVideos = () => {
  const input = document.createElement('input');
  input.type = 'file';
  input.accept = 'video/*';
  input.multiple = true;
  input.onchange = handleVideoUpload;
  input.click();
};

const handleImageUpload = async (event) => {
  const files = Array.from(event.target.files);
  if (files.length === 0) return;

  isUploading.value = true;
  
  try {
    for (const file of files) {
      const formData = new FormData();
      formData.append("file", file);
      
      const response = await httpPost("/api/upload", formData);
      
      // 将相对路径转换为绝对路径
      let imageUrl = response.data.url;
      if (!imageUrl.startsWith('http')) {
        // 如果是相对路径，转换为绝对路径
        imageUrl = location.protocol + "//" + location.host + imageUrl;
      }
      
      const imageData = {
        id: response.data.id,
        name: file.name,
        size: file.size,
        url: imageUrl,
        objKey: response.data.objKey
      };
      uploadedImages.value.push(imageData);
    }
    
    showToast({
      type: 'success',
      message: `成功上传 ${files.length} 张图片`
    });
  } catch (error) {
    showToast({
      type: 'fail',
      message: "图片上传失败：" + error.message
    });
  } finally {
    isUploading.value = false;
  }
};

const removeImage = (index) => {
  uploadedImages.value.splice(index, 1);
  showToast({
    type: 'success',
    message: "图片已删除"
  });
};

const previewImage = (url) => {
  showImagePreview([url]);
};

// 视频上传处理
const handleVideoUpload = async (event) => {
  const files = Array.from(event.target.files);
  if (files.length === 0) return;

  isUploading.value = true;
  
  try {
    for (const file of files) {
      const formData = new FormData();
      formData.append("file", file);
      
      const response = await httpPost("/api/upload", formData);
      
      // 将相对路径转换为绝对路径
      let videoUrl = response.data.url;
      if (!videoUrl.startsWith('http')) {
        // 如果是相对路径，转换为绝对路径
        videoUrl = location.protocol + "//" + location.host + videoUrl;
      }
      
      const videoData = {
        id: response.data.id,
        name: file.name,
        size: file.size,
        url: videoUrl,
        objKey: response.data.objKey
      };
      uploadedVideos.value.push(videoData);
    }
    
    showToast({
      type: 'success',
      message: `成功上传 ${files.length} 个视频`
    });
  } catch (error) {
    showToast({
      type: 'fail',
      message: "视频上传失败：" + error.message
    });
  } finally {
    isUploading.value = false;
  }
};

const removeVideo = (index) => {
  uploadedVideos.value.splice(index, 1);
  showToast({
    type: 'success',
    message: "视频已删除"
  });
};

const previewVideo = (url) => {
  // 使用 van-image-preview 的类似方式预览视频
  // 或者可以创建一个视频预览弹窗
  window.open(url, '_blank');
};

// 处理回车键事件
const handleEnterKey = (event) => {
  if (event.shiftKey) {
    // Shift + Enter 换行
    return;
  } else {
    // Enter 发送消息
    event.preventDefault();
    sendMessage();
  }
};

// // eslint-disable-next-line no-undef
// const recognition = new webkitSpeechRecognition() || SpeechRecognition();
// //recognition.lang = 'zh-CN' // 设置语音识别语言
// recognition.onresult = function (event) {
//   prompt.value = event.results[0][0].transcript
// };
//
// recognition.onerror = function (event) {
//   showMic.value = false
//   recognition.stop()
//   showNotify({type: 'danger', message: '语音识别错误:' + event.error})
// };
//
// recognition.onend = function () {
//   console.log('语音识别结束');
// };
// const inputVoice = () => {
//   showMic.value = true
//   recognition.start();
// }
//
// const stopVoice = () => {
//   showMic.value = false
//   recognition.stop()
// }
</script>

<style lang="stylus">
@import "@/assets/css/mobile/chat-session.styl"

.van-nav-bar {
  padding: 0 8px;

  .chat-title-wrapper {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    line-height: 1.1;
  }

  .chat-title {
    font-size: 16px;
    font-weight: 600;
    color: #1f2d3d;
  }

  .chat-subtitle {
    margin-top: 3px;
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    border-radius: 999px;
    background: rgba(31, 45, 61, 0.08);
    font-size: 11px;
    font-weight: 500;
    color: #4f5b70;

    .van-icon {
      font-size: 12px;
    }
  }

  .nav-quick-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .nav-action-chip {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 10px;
    border-radius: 999px;
    font-size: 12px;
    color: #1f2d3d;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
    cursor: pointer;
    transition: transform 0.2s ease, box-shadow 0.2s ease;

    .van-icon {
      font-size: 16px;
    }

    &:active {
      transform: scale(0.95);
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12);
    }
  }
}

// 确保整个聊天容器占满屏幕高度
.mobile-chat {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

// 聊天内容区域自适应高度
.chat-list-wrapper {
  flex: 1;
  overflow: hidden;
  padding: 10px 0 0 0;
  
  .message-list-box {
    height: 100%;
    overflow-y: auto;
    padding-bottom: 10px;
  }
}

// 输入框区域固定在底部
.chat-box-wrapper {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  background: white;
  border-top: 1px solid #ebedf0;
  
  // 移除van-cell-group的默认边距
  .van-cell-group--inset {
    margin: 0 !important;
    border-radius: 0 !important;
  }
  
  // 移除van-cell的默认内边距
  .van-cell {
    padding-left: 0 !important;
    padding-right: 0 !important;
  }
}

.image-preview-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 8px 16px;
  background: white; /* 改为白色背景，与聊天背景一致 */
  border-bottom: 1px solid #ebedf0;
  min-height: 76px;
  width: 100%; /* 确保占满宽度 */
  box-sizing: border-box; /* 包含padding在内计算宽度 */
}

.image-preview-item {
  position: relative;
  display: inline-block;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.video-preview-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 8px 16px;
  background: white;
  border-bottom: 1px solid #ebedf0;
  min-height: 76px;
  width: 100%;
  box-sizing: border-box;
}

.video-preview-item {
  position: relative;
  display: inline-block;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  cursor: pointer;
}

.video-play-icon {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: white;
  background: rgba(0, 0, 0, 0.5);
  border-radius: 50%;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
  z-index: 1;
}

.remove-image-icon {
  position: absolute;
  top: 4px;
  right: 4px;
  background: rgba(255, 71, 87, 0.9);
  color: white;
  border-radius: 50%;
  width: 18px;
  height: 18px;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 2;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  transition: all 0.2s ease;
}

.remove-image-icon:hover {
  background: #ff4757;
  transform: scale(1.1);
}

.upload-icon {
  color: #1989fa;
  cursor: pointer;
  font-size: 18px;
}

.upload-icon:hover {
  color: #0570db;
}

.icon-box {
  display: flex;
  align-items: center;
  gap: 8px;
}

// 多行输入框样式调整
.van-field__control {
  resize: none;
  min-height: 20px;
  max-height: 80px;
}

// 发送按钮禁用状态
.van-button--disabled {
  opacity: 0.5;
}

// 移除van-sticky的默认样式，因为我们使用fixed定位
.chat-box-wrapper .van-sticky {
  position: static;
}

// 确保输入框区域也没有留白
.chat-box-wrapper .van-field {
  padding: 10px 16px; /* 给输入框添加适当的内边距 */
  background: white;
}

// 模型选择小浮块样式
.model-chip-wrapper {
  display: flex;
  flex-wrap: nowrap;
  overflow-x: auto;
  padding: 8px 12px 6px 12px;
  gap: 8px;
  background: #f7f8fa;
}

.model-chip {
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 14px;
  background: #ffffff;
  border: 1px solid #e5e5e5;
  white-space: nowrap;
  flex-shrink: 0;
}

.model-chip.active {
  background: #1989fa;
  color: #ffffff;
  border-color: #1989fa;
}

// 历史会话弹窗样式
.history-popup {
  background: #f7f8fa;
}

.history-popup .van-nav-bar {
  /* 覆盖 Home.vue 里对全局 .van-nav-bar 设置的 position: fixed，避免遮挡列表首行 */
  position: sticky;
  top: 0;
  width: 100%;
  z-index: 2;
}

.history-new-chat-btn {
  border: none;
  background: linear-gradient(120deg, #5b8cff 0%, #4866ff 100%);
  box-shadow: 0 6px 16px rgba(90, 133, 255, 0.35);
  padding: 0 16px;
  border-radius: 999px;

  .history-new-chat-text {
    font-weight: 600;
    letter-spacing: 0.08em;
  }
}

.history-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.history-content {
  flex: 1;
  overflow-y: auto;
}

.history-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.history-title {
  flex: 1;
}
</style>
