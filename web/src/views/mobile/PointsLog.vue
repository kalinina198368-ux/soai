<template>
  <div class="mobile-power-log">
    <van-nav-bar title="" left-arrow @click-left="goBack" />
    
    <div class="content">
      <!-- 搜索筛选区域 -->
      <div class="filter-section">
        <van-field
          v-model="query.model"
          placeholder=""
          clearable
          class="filter-input"
        />
        
        <van-field
          v-model="query.typeText"
          readonly
          label="类型"
          placeholder="全部"
          is-link
          @click="showTypePicker = true"
          class="filter-input"
        />
        
        <van-popup v-model:show="showTypePicker" position="bottom">
          <van-picker
            :columns="typeColumns"
            @confirm="onTypeConfirm"
            @cancel="showTypePicker = false"
          />
        </van-popup>
        
        <van-field
          v-model="dateRangeText"
          readonly
          label="日期"
          placeholder="选择日期范围"
          is-link
          @click="showDatePicker = true"
          class="filter-input"
        />
        
        <van-calendar
          v-model:show="showDatePicker"
          type="range"
          @confirm="onDateConfirm"
        />
        
        <van-button type="primary" block @click="search" class="search-btn">
          搜索
        </van-button>
        
        <div class="total-power" v-if="totalPower > 0">
          <van-cell title="算力总额" :value="totalPower" />
        </div>
      </div>

      <!-- 数据列表 -->
      <div class="list-section" v-if="items.length > 0">
        <van-cell-group inset>
          <van-cell
            v-for="item in items"
            :key="item.id"
            class="power-item"
          >
            <template #title>
              <div class="item-title">
                <van-tag
                  :type="tagTypeMap[item.type] || 'default'"
                  size="small"
                  class="type-tag"
                >
                  {{ item.type_str }}
                </van-tag>
                <span class="model-text">{{ item.model || '未知模型' }}</span>
              </div>
            </template>
            <template #value>
              <div class="amount-cell">
                <span
                  class="amount"
                  :class="{ positive: item.mark === 1, negative: item.mark === 0 }"
                >
                  {{ item.mark === 1 ? '+' : '-' }}{{ item.amount }}
                </span>
                <div class="balance">余额: {{ item.balance }}</div>
              </div>
            </template>
            <template #label>
              <div class="item-label">
                <span class="time-text">{{ dateFormat(item.created_at) }}</span>
                <span class="remark-text" v-if="item.remark">{{ item.remark }}</span>
              </div>
            </template>
          </van-cell>
        </van-cell-group>
        
        <!-- 加载更多提示 -->
        <div v-if="isLoadingMore" class="load-more-tip">
          <van-loading type="spinner" size="16px" />
          <span>加载中...</span>
        </div>
        <div v-if="finished && items.length > 0" class="load-more-tip finished">
          <span>没有更多了</span>
        </div>
      </div>

      <!-- 空状态 -->
      <van-empty
        v-else-if="!loading"
        description="暂无数据"
        :image-size="100"
      />
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref, computed } from "vue";
import { httpPost } from "@/utils/http";
import { dateFormat } from "@/utils/libs";
import { showFailToast, showLoadingToast, showSuccessToast } from "vant";
import { useRouter } from "vue-router";
import { checkSession } from "@/store/cache";

const router = useRouter();

const items = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const loading = ref(false);
const finished = ref(false);
const isLoadingMore = ref(false);
const totalPower = ref(0);

// 滚动监听相关
let scrollHandler = null;
let loadMoreTimer = null; // 加载更多延迟定时器

const query = ref({
  model: "",
  date: [],
  type: 0,
  typeText: "全部",
});

const showTypePicker = ref(false);
const showDatePicker = ref(false);
const dateRangeText = ref("");

const typeColumns = [
  { text: "全部", value: 0 },
  { text: "充值", value: 1 },
  { text: "消费", value: 2 },
  { text: "退款", value: 3 },
  { text: "奖励", value: 4 },
  { text: "众筹", value: 5 },
  { text: "转账", value: 8 },
  { text: "收款", value: 9 },
  { text: "质押", value: 10 },
  { text: "释放", value: 11 },
];

const tagTypeMap = {
  1: "success",
  2: "primary",
  3: "danger",
  4: "warning",
  5: "default",
  8: "primary",
  9: "success",
  10: "warning",
  11: "info",
};

const goBack = () => {
  router.back();
};

const onTypeConfirm = ({ selectedOptions }) => {
  const option = selectedOptions[0];
  query.value.type = option.value;
  query.value.typeText = option.text;
  showTypePicker.value = false;
};

const onDateConfirm = (values) => {
  if (values && values.length === 2) {
    const formatDate = (date) => {
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      return `${year}-${month}-${day}`;
    };
    
    query.value.date = [
      formatDate(values[0]),
      formatDate(values[1]),
    ];
    dateRangeText.value = `${query.value.date[0]} 至 ${query.value.date[1]}`;
  }
  showDatePicker.value = false;
};

// 设置滚动监听
const setupScrollListener = () => {
  scrollHandler = () => {
    // 获取滚动位置
    const scrollTop = window.pageYOffset || document.documentElement.scrollTop || document.body.scrollTop;
    const windowHeight = window.innerHeight || document.documentElement.clientHeight;
    const documentHeight = document.documentElement.scrollHeight || document.body.scrollHeight;
    
    // 计算距离底部的距离（提前100px触发加载）
    const distanceToBottom = documentHeight - (scrollTop + windowHeight);
    
    // 当距离底部小于100px且未加载完且不在加载中时，触发加载更多
    if (distanceToBottom < 100 && !finished.value && !isLoadingMore.value && !loading.value) {
      // 清除之前的定时器
      if (loadMoreTimer) {
        clearTimeout(loadMoreTimer);
      }
      
      // 延迟200毫秒后执行加载，避免过于频繁的请求
      loadMoreTimer = setTimeout(() => {
        loadMore();
        loadMoreTimer = null;
      }, 500);
    }
  };
  
  // 添加滚动监听
  window.addEventListener('scroll', scrollHandler, { passive: true });
};

const search = () => {
  page.value = 1;
  items.value = [];
  finished.value = false;
  fetchData(1);
};

// 加载更多数据
const loadMore = () => {
  // 再次检查状态，防止延迟期间状态发生变化
  if (finished.value || isLoadingMore.value || loading.value) {
    return;
  }
  
  page.value += 1;
  fetchData(page.value, true);
};

const fetchData = (_page, isLoadMore = false) => {
  if (_page) {
    page.value = _page;
  }
  
  // 如果是首次加载或刷新，重置finished状态
  if (!isLoadMore) {
    finished.value = false;
  }
  
  // 如果是加载更多，使用 isLoadingMore，否则使用 loading
  if (isLoadMore) {
    isLoadingMore.value = true;
  } else {
    loading.value = true;
  }
  
  // 构建请求参数
  const params = {
    model: query.value.model,
    date: query.value.date,
    page: page.value,
    page_size: pageSize.value,
  };
  
  // 如果类型不是"全部"，添加类型参数
  if (query.value.type > 0) {
    params.type = query.value.type;
  }
  
  httpPost("/api/pointsLog/list", params)
    .then((res) => {
      if (res.data) {
        // 后端返回的数据结构：vo.NewPage(total, page, pageSize, items)
        // 兼容不同的数据结构：res.data.data 或 res.data
        const data = res.data.data || res.data;
        const newItems = data.items || [];
        
        if (isLoadMore) {
          // 加载更多时，追加数据
          items.value = [...items.value, ...newItems];
          isLoadingMore.value = false;
        } else {
          // 首次加载或刷新时，替换数据
          items.value = newItems;
          loading.value = false;
        }
        
        total.value = data.total || 0;
        totalPower.value = res.data.stat || 0;
        
        // 判断是否已加载完所有数据
        const totalPages = Math.ceil(total.value / pageSize.value);
        finished.value = page.value >= totalPages || newItems.length < pageSize.value;
      } else {
        // 没有返回数据，认为加载完成
        finished.value = true;
        if (isLoadMore) {
          isLoadingMore.value = false;
        } else {
          loading.value = false;
        }
      }
    })
    .catch((e) => {
      if (isLoadMore) {
        isLoadingMore.value = false;
      } else {
        loading.value = false;
      }
      // 请求失败时，不设置 finished，允许用户重试
      // 但如果已经是第一页且没有数据，则设置为完成状态
      if (page.value === 1 && items.value.length === 0) {
        finished.value = true;
      }
      showFailToast("获取数据失败：" + e.message);
    });
};

onMounted(() => {
  checkSession()
    .then(() => {
      fetchData(1);
      // 添加滚动监听，实现下拉加载更多
      setupScrollListener();
    })
    .catch(() => {
      router.push("/mobile/login");
    });
});

onUnmounted(() => {
  // 清理滚动监听
  if (scrollHandler) {
    window.removeEventListener('scroll', scrollHandler);
    scrollHandler = null;
  }
  // 清理加载更多定时器
  if (loadMoreTimer) {
    clearTimeout(loadMoreTimer);
    loadMoreTimer = null;
  }
});
</script>

<style lang="stylus" scoped>
.mobile-power-log {
  min-height 100vh
  background #f5f5f5

  .content {
    padding 16px
    padding-bottom 80px

    .filter-section {
      background white
      border-radius 12px
      padding 16px
      margin-bottom 16px

      .filter-input {
        margin-bottom 12px
        padding 0

        &:last-of-type {
          margin-bottom 16px
        }
      }

      .search-btn {
        margin-top 8px
        border-radius 8px
      }

      .total-power {
        margin-top 16px
        padding-top 16px
        border-top 1px solid #ebedf0

        :deep(.van-cell__value) {
          color #1989fa
          font-weight 600
          font-size 18px
        }
      }
    }

    .list-section {
      .power-item {
        margin-bottom 12px
        border-radius 12px
        overflow hidden

        .item-title {
          display flex
          align-items center
          gap 8px

          .type-tag {
            flex-shrink 0
          }

          .model-text {
            font-size 14px
            color #323233
          }
        }

        .amount-cell {
          text-align right

          .amount {
            display block
            font-size 18px
            font-weight 600
            margin-bottom 4px

            &.positive {
              color #07c160
            }

            &.negative {
              color #ee0a24
            }
          }

          .balance {
            font-size 12px
            color #969799
          }
        }

        .item-label {
          display flex
          flex-direction column
          gap 4px
          margin-top 8px

          .time-text {
            font-size 12px
            color #969799
          }

          .remark-text {
            font-size 12px
            color #969799
            padding-top 4px
            border-top 1px solid #f0f0f0
          }
        }
      }
      
      // 加载更多提示
      .load-more-tip {
        display flex
        align-items center
        justify-content center
        padding 16px
        color #969799
        font-size 14px
        gap 8px
        
        &.finished {
          color #969799
          opacity 0.6
        }
      }
    }
  }
}
</style>
