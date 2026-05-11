<template>
  <div class="mobile-help-center">
    <van-nav-bar title="帮助中心" left-arrow @click-left="goBack" />
    
    <div class="content">
      <!-- 搜索框 -->
      <!-- <div class="search-section">
        <van-search
          v-model="searchKeyword"
          placeholder="搜索问题"
          @search="onSearch"
          @clear="onClear"
        />
      </div> -->

      <!-- 常见问题分类 -->
      <div class="category-section">
        <div class="section-header">
          <h3 class="section-title">
            <i class="iconfont icon-question"></i>
            常见问题
          </h3>
        </div>
        
        <van-cell-group inset>
          <van-cell
            v-for="category in categories"
            :key="category.id"
            :title="category.name"
            is-link
            @click="showCategoryQuestions(category)"
            class="category-item"
          >
            <template #icon>
              <van-icon :name="category.icon" size="20" :color="category.color" />
            </template>
          </van-cell>
        </van-cell-group>
      </div>

      <!-- 热门问题 -->
      <div class="hot-section" v-if="hotQuestions.length > 0">
        <div class="section-header">
          <h3 class="section-title">
            <i class="iconfont icon-fire"></i>
            热门问题
          </h3>
        </div>
        
        <van-collapse v-model="activeNames" class="question-collapse">
          <van-collapse-item
            v-for="question in hotQuestions"
            :key="question.id"
            :name="question.id"
            :title="question.title"
            class="question-item"
          >
            <div class="question-answer" v-html="question.answer"></div>
          </van-collapse-item>
        </van-collapse>
      </div>

      <!-- 搜索结果 -->
      <div class="search-results" v-if="searchResults.length > 0">
        <div class="section-header">
          <h3 class="section-title">搜索结果</h3>
        </div>
        
        <van-collapse v-model="activeNames" class="question-collapse">
          <van-collapse-item
            v-for="result in searchResults"
            :key="result.id"
            :name="result.id"
            :title="result.title"
            class="question-item"
          >
            <div class="question-answer" v-html="result.answer"></div>
          </van-collapse-item>
        </van-collapse>
      </div>

      <!-- 空状态 -->
      <van-empty
        v-if="searchKeyword && searchResults.length === 0 && !loading"
        description="未找到相关问题"
        :image-size="100"
      />

      <!-- 联系客服 -->
      <div class="contact-section">
        <van-button
          type="primary"
          block
          @click="contactService"
          class="contact-btn"
        >
          <van-icon name="chat-o" />
          联系客服
        </van-button>
      </div>
    </div>

    <!-- 分类问题弹窗 -->
    <van-popup
      v-model:show="showCategoryDialog"
      position="bottom"
      :style="{ height: '80%' }"
      round
    >
      <div class="category-dialog">
        <div class="dialog-header">
          <h3>{{ currentCategory?.name }}</h3>
          <van-icon name="cross" @click="showCategoryDialog = false" />
        </div>
        <div class="dialog-content">
          <van-collapse v-model="categoryActiveNames" class="question-collapse">
            <van-collapse-item
              v-for="question in categoryQuestions"
              :key="question.id"
              :name="question.id"
              :title="question.title"
              class="question-item"
            >
              <div class="question-answer" v-html="question.answer"></div>
            </van-collapse-item>
          </van-collapse>
          <van-empty
            v-if="categoryQuestions.length === 0"
            description="暂无问题"
            :image-size="80"
          />
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { showToast } from "vant";

const router = useRouter();

const searchKeyword = ref("");
const loading = ref(false);
const activeNames = ref([]);
const categoryActiveNames = ref([]);
const showCategoryDialog = ref(false);
const currentCategory = ref(null);
const categoryQuestions = ref([]);
const searchResults = ref([]);

// 问题分类
const categories = ref([
  {
    id: 1,
    name: "账号相关",
    icon: "user-o",
    color: "#409EFF",
    count: 5,
  },
  {
    id: 2,
    name: "充值问题",
    icon: "gold-coin-o",
    color: "#FF9800",
    count: 8,
  },
  {
    id: 3,
    name: "使用指南",
    icon: "guide-o",
    color: "#67C23A",
    count: 12,
  },

  // {
  //   id: 4,
  //   name: "功能说明",
  //   icon: "setting-o",
  //   color: "#909399",
  //   count: 10,
  // },
  // {
  //   id: 5,
  //   name: "其他问题",
  //   icon: "question-o",
  //   color: "#F56C6C",
  //   count: 6,
  // },

]);

// 热门问题
const hotQuestionsData = [
  {
    id: "hot-1",
    title: "如何注册账号？",
    answer: "您可以通过手机号或邮箱进行注册。点击登录页面的「注册」按钮，填写相关信息即可完成注册。",
  },
  {
    id: "hot-2",
    title: "如何充值算力值？",
    answer: "进入「会员充值中心」，选择您需要的套餐，选择支付方式完成支付即可。支付成功后算力值会自动到账。",
  },
  {
    id: "hot-3",
    title: "算力值是什么？",
    answer: "算力值是平台内的虚拟货币，用于使用各种AI功能。不同的功能消耗的算力值不同，您可以在使用前查看消耗说明。",
  },
  {
    id: "hot-4",
    title: "如何查看消费记录？",
    answer: "在个人中心点击「明细」按钮，可以查看您的所有消费记录，包括充值记录和使用记录。",
  },
  {
    id: "hot-5",
    title: "如何联系客服？",
    answer: "您可以在个人中心点击「客服」按钮，查看客服联系方式，客服工作时间为工作日9:00-18:00。",
  },
];

const hotQuestions = ref(hotQuestionsData);

// 所有问题（用于搜索）
const allQuestionsData = [
  ...hotQuestionsData,
  {
    id: "q-1",
    title: "忘记密码怎么办？",
    answer: "您可以点击登录页面的「忘记密码」链接，通过手机号或邮箱找回密码。",
    category: 1,
  },
  {
    id: "q-2",
    title: "如何修改个人信息？",
    answer: "在个人中心点击头像或用户名，可以修改昵称、头像等个人信息。",
    category: 1,
  },
  {
    id: "q-3",
    title: "支付失败怎么办？",
    answer: "请检查您的支付方式是否正常，账户余额是否充足。如问题持续，请联系客服处理。",
    category: 2,
  },
  {
    id: "q-4",
    title: "如何使用AI对话功能？",
    answer: "进入「对话」，选择或创建对话，输入您的问题即可开始对话。您可以选择不同的AI模型来获得不同的回答风格。",
    category: 3,
  },
  {
    id: "q-5",
    title: "如何生成图片？",
    answer: "进入「绘画」，选择MJ绘图，输入图片描述，设置参数后即可生成图片。",
    category: 3,
  },
 {
    id: "q-6",
    title: "如何生成视频？",
    answer: "进入「sora」，输入文字或者添加对应的参考图，设置对应的时长和横竖屏设置，设置参数后即可生成视频。",
    category: 3,
  },
];

const allQuestions = ref(allQuestionsData);

const goBack = () => {
  router.back();
};

const onSearch = (value) => {
  if (!value.trim()) {
    showToast("请输入搜索关键词");
    return;
  }
  
  loading.value = true;
  // 模拟搜索
  setTimeout(() => {
    searchResults.value = allQuestions.value.filter((q) =>
      q.title.toLowerCase().includes(value.toLowerCase()) ||
      q.answer.toLowerCase().includes(value.toLowerCase())
    );
    loading.value = false;
    
    if (searchResults.value.length === 0) {
      showToast("未找到相关问题");
    }
  }, 300);
};

const onClear = () => {
  searchResults.value = [];
};

const showCategoryQuestions = (category) => {
  currentCategory.value = category;
  // 模拟获取分类问题
  categoryQuestions.value = allQuestions.value.filter(
    (q) => q.category === category.id
  );
  showCategoryDialog.value = true;
};

const contactService = () => {
  // 返回上一页并触发客服弹窗
  router.back();
  // 这里可以通过事件总线或其他方式通知 Profile 页面打开客服弹窗
  // 或者直接跳转到个人中心
  setTimeout(() => {
    router.push("/mobile/profile");
  }, 100);
};

onMounted(() => {
  // 可以在这里加载实际的帮助数据
});
</script>

<style scoped lang="scss">
.mobile-help-center {
  min-height: 100vh;
  background: #f5f5f5;

  .content {
    padding: 16px;
    padding-bottom: 80px;
  }

  .search-section {
    margin-bottom: 16px;
  }

  .section-header {
    margin: 24px 0 12px;
    
    .section-title {
      font-size: 18px;
      font-weight: 600;
      color: #323233;
      display: flex;
      align-items: center;
      gap: 8px;
      
      .iconfont {
        font-size: 20px;
        color: #409EFF;
      }
    }
  }

  .category-item {
    margin-bottom: 8px;
  }

  .question-collapse {
    margin-bottom: 16px;
    
    .question-item {
      margin-bottom: 8px;
    }
    
    .question-answer {
      padding: 12px 0;
      line-height: 1.6;
      color: #646566;
      font-size: 14px;
    }
  }

  .contact-section {
    margin-top: 32px;
    padding: 0 16px;
    
    .contact-btn {
      height: 44px;
      border-radius: 22px;
      font-size: 16px;
      
      .van-icon {
        margin-right: 8px;
      }
    }
  }

  .category-dialog {
    height: 100%;
    display: flex;
    flex-direction: column;
    
    .dialog-header {
      padding: 16px;
      border-bottom: 1px solid #ebedf0;
      display: flex;
      justify-content: space-between;
      align-items: center;
      
      h3 {
        font-size: 18px;
        font-weight: 600;
        margin: 0;
      }
      
      .van-icon {
        font-size: 20px;
        color: #969799;
      }
    }
    
    .dialog-content {
      flex: 1;
      overflow-y: auto;
      padding: 16px;
    }
  }
}
</style>
