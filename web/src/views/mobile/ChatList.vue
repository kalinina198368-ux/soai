<template>
  <div class="app-background">
    <div class="container mobile-chat-list">
      <van-nav-bar
          :title="title"
          custom-class="navbar"
      >
        <template #left>
          <van-icon name="wap-nav" @click="toggleHistory"/>
        </template>
        <template #right>
          <van-icon name="delete-o" @click="clearAllChatHistory"/>
        </template>
      </van-nav-bar>

      <div class="new-chat">
        <div class="hero">
          <van-image
              class="hero-icon"
              src="https://cdn.jsdelivr.net/gh/DeepSeek-AI/art/logo.png"
              fit="contain"
          />
          <p class="hero-title">嗨！我是 DeepSeek</p>
          <p class="hero-desc">我可以帮你搜索、答疑、写作，请把你的任务交给我吧~</p>
        </div>

        <div class="model-section">
          <div class="section-title">选择模型</div>
          <div class="model-chips">
            <span
                v-for="model in models"
                :key="model.value"
                :class="['chip', {active: selectedModel && selectedModel.value === model.value}]"
                @click="selectModel(model)"
            >
              {{ model.text }}
            </span>
          </div>
        </div>

        <van-field
            v-model="roleDisplayName"
            readonly
            label="角色"
            placeholder="请选择聊天角色"
            class="field role-field"
            @click="openRolePicker"
        />

        <van-field
            v-model="chatTitle"
            label="标题"
            placeholder="给这次对话取个名字吧"
            class="field"
        />

        <van-button
            type="primary"
            block
            class="start-btn"
            @click="startChat"
        >
          开启新对话
        </van-button>
      </div>
    </div>

    <van-popup v-model:show="showHistory" position="left" class="history-popup">
      <div class="history-header">
        <div class="history-title">历史会话</div>
        <van-icon name="cross" @click="showHistory = false"/>
      </div>
      <van-search
          v-model="historyKeyword"
          input-align="center"
          placeholder="搜索会话标题"
          custom-class="van-search"
          @input="search"
      />
      <div class="history-body">
        <van-empty v-if="!chats.length && !loading" description="暂无会话"/>
        <div v-else class="history-list">
          <van-swipe-cell v-for="item in chats" :key="item.id">
            <van-cell @click="changeChat(item)">
              <div class="chat-list-item">
                <van-image
                    :src="item.icon"
                    round
                />
                <div class="van-ellipsis">{{ item.title }}</div>
              </div>
            </van-cell>
            <template #right>
              <van-button square text="修改" type="primary" @click="editChat(item)"/>
              <van-button square text="删除" type="danger" @click="removeChat(item)"/>
            </template>
          </van-swipe-cell>
        </div>
      </div>
    </van-popup>

    <van-popup v-model:show="showRolePicker" position="bottom" class="popup">
      <van-picker
          :columns="roles"
          title="选择聊天角色"
          @cancel="showRolePicker = false"
          @confirm="confirmRole"
      >
        <template #option="item">
          <div class="picker-option">
            <van-image
                v-if="item.icon"
                :src="item.icon"
                fit="cover"
                round
            />
            <span>{{ item.text }}</span>
          </div>
        </template>
      </van-picker>
    </van-popup>

    <van-dialog v-model:show="showEditChat" title="修改对话标题" show-cancel-button class="dialog" @confirm="saveTitle">
      <van-field v-model="tmpChatTitle" label="" placeholder="请输入对话标题" class="field"/>
    </van-dialog>

  </div>
</template>

<script setup>
import {ref, onMounted, watch} from "vue";
import {httpGet, httpPost} from "@/utils/http";
import {showConfirmDialog, showFailToast, showSuccessToast} from "vant";
import {checkSession} from "@/store/cache";
import {router} from "@/router";
import {removeArrayItem, showLoginDialog} from "@/utils/libs";

const title = ref("新对话")
const chats = ref([])
const allChats = ref([])
const loading = ref(false)
const finished = ref(false)
const error = ref(false)
const loginUser = ref(null)
const isLogin = ref(false)
const roles = ref([])
const models = ref([])
const selectedRole = ref(null)
const selectedModel = ref(null)
const showRolePicker = ref(false)
const chatTitle = ref("")
const showHistory = ref(false)
const historyKeyword = ref("")
const roleDisplayName = ref("")
const showEditChat = ref(false)
const item = ref({})
const tmpChatTitle = ref("")
const autoStarted = ref(false)

checkSession().then((user) => {
  loginUser.value = user
  isLogin.value = true
  // 加载角色列表
  httpGet(`/api/app/list/user`).then((res) => {
    if (res.data) {
      setRoles(res.data)
    }
  }).catch(() => {
    showFailToast("加载聊天角色失败")
  })

  // 加载模型
  httpGet('/api/model/list?enable=1').then(res => {
    if (res.data) {
      setModels(res.data)
    }
  }).catch(e => {
    showFailToast("加载模型失败: " + e.message)
  })

}).catch(() => {
  loading.value = false
  finished.value = true

  // 加载角色列表
  httpGet('/api/app/list/user').then((res) => {
    if (res.data) {
      setRoles(res.data)
    }
  }).catch(() => {
    showFailToast("加载聊天角色失败")
  })

  // 加载模型
  httpGet('/api/model/list').then(res => {
    if (res.data) {
      setModels(res.data)
    }
  }).catch(e => {
    showFailToast("加载模型失败: " + e.message)
  })
})

const setRoles = (items) => {
  roles.value = items.map((role) => ({
    text: role.name,
    value: role.id,
    icon: role.icon,
    helloMsg: role.hello_msg
  }))
  if (!selectedRole.value && roles.value.length) {
    selectedRole.value = roles.value[0]
    roleDisplayName.value = roles.value[0].text
  }
}

const setModels = (items) => {
  models.value = items.map(model => ({text: model.name, value: model.id}))
  if (!selectedModel.value && models.value.length) {
    selectedModel.value = models.value[0]
  }
}

// 角色与模型加载完成后，自动进入会话（默认第一个角色 + 第一个模型）
watch(
  [selectedRole, selectedModel, isLogin],
  ([role, model, login]) => {
    if (!autoStarted.value && login && role && model) {
      autoStarted.value = true
      startChat()
    }
  },
  { immediate: false }
)

const onLoad = () => {
  checkSession().then(() => {
    httpGet("/api/chat/list?user_id=" + loginUser.value.id).then((res) => {
      if (res.data) {
        chats.value = res.data;
        allChats.value = res.data;
        finished.value = true
      }
      loading.value = false;
    }).catch(() => {
      error.value = true
      showFailToast("加载会话列表失败")
    })
  }).catch(() => {})
};

const search = () => {
  if (historyKeyword.value === '') {
    chats.value = allChats.value
    return
  }
  const items = [];
  for (let i = 0; i < allChats.value.length; i++) {
    if (allChats.value[i].title.toLowerCase().indexOf(historyKeyword.value.toLowerCase()) !== -1) {
      items.push(allChats.value[i]);
    }
  }
  chats.value = items;
}

const toggleHistory = () => {
  showHistory.value = !showHistory.value
  if (showHistory.value && !chats.value.length) {
    onLoad()
  }
}

const clearAllChatHistory = () => {
  if (!isLogin.value) {
    return showLoginDialog(router)
  }

  showConfirmDialog({
    title: '操作提示',
    message: '确定要删除所有的会话记录吗？'
  }).then(() => {
    httpGet("/api/chat/clear").then(() => {
      showSuccessToast('所有聊天记录已清空')
      chats.value = [];
    }).catch(e => {
      showFailToast("操作失败：" + e.message)
    })
  }).catch(() => {
    // on cancel
  })
}

const selectModel = (model) => {
  selectedModel.value = model
}

const openRolePicker = () => {
  if (!roles.value.length) {
    return showFailToast("暂无可选角色")
  }
  showRolePicker.value = true
}

const confirmRole = ({selectedOptions}) => {
  if (selectedOptions && selectedOptions.length) {
    selectedRole.value = selectedOptions[0]
    roleDisplayName.value = selectedOptions[0].text
  }
  showRolePicker.value = false
}

const startChat = () => {
  if (!isLogin.value) {
    return showLoginDialog(router)
  }
  if (!selectedRole.value || !selectedModel.value) {
    return showFailToast("请先选择角色和模型")
  }
  const titleValue = chatTitle.value ? encodeURIComponent(chatTitle.value) : encodeURIComponent("新对话")
  router.push(`/mobile/chat/session?title=${titleValue}&role_id=${selectedRole.value.value}&model_id=${selectedModel.value.value}`)
}

const changeChat = (chat) => {
  router.push(`/mobile/chat/session?chat_id=${chat.chat_id}`)
}

const editChat = (row) => {
  showEditChat.value = true
  item.value = row
  tmpChatTitle.value = row.title
}
const saveTitle = () => {
  httpPost('/api/chat/update', {chat_id: item.value.chat_id, title: tmpChatTitle.value}).then(() => {
    showSuccessToast("操作成功！");
    item.value.title = tmpChatTitle.value;
  }).catch(e => {
    showFailToast("操作失败：" + e.message);
  })
}

const removeChat = (item) => {
  httpGet('/api/chat/remove?chat_id=' + item.chat_id).then(() => {
    chats.value = removeArrayItem(chats.value, item, function (e1, e2) {
      return e1.id === e2.id
    })
    allChats.value = removeArrayItem(allChats.value, item, function (e1, e2) {
      return e1.id === e2.id
    })
  }).catch(e => {
    showFailToast('操作失败：' + e.message);
  })

}

onMounted(() => {
  onLoad()
})

</script>

<style lang="stylus" scoped>
@import "@/assets/css/mobile/chat-list.styl"
</style>