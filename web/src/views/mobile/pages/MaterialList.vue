<template>
  <div class="material-list-page">
    <van-nav-bar
      title="选择素材"
      left-arrow
      @click-left="goBack"
    />
    
    <div class="material-content">
      <!-- 素材分类选择 -->
      <div class="category-filter" v-if="categories.length > 0">
        <van-dropdown-menu>
          <van-dropdown-item v-model="selectedCategory" :options="categoryOptions" @change="onCategoryChange" />
        </van-dropdown-menu>
      </div>
      
      <!-- 素材列表 -->
      <div class="material-grid" v-if="materials.length > 0">
        <van-grid :gutter="10" :column-num="2">
          <van-grid-item v-for="item in materials" :key="item.id">
            <div 
              :class="selectedMaterialId === item.id ? 'material-item active' : 'material-item'" 
              @click="selectMaterial(item)"
            >
              <div class="material-image-wrapper">
                <van-image 
                  :src="item.image || item.preview || '/images/img-placeholder.jpg'" 
                  fit="cover"
                  :lazy-load="true"
                >
                  <template v-slot:loading>
                    <van-loading type="spinner" size="20" />
                  </template>
                  <template v-slot:error>
                    <div class="material-placeholder">
                      <i class="iconfont icon-image"></i>
                    </div>
                  </template>
                </van-image>
                <!-- 选中标记 -->
                <div class="selected-badge" v-if="selectedMaterialId === item.id">
                  <van-icon name="success" color="#1989fa" size="24" />
                </div>
              </div>
              <div class="material-title">
                <van-text-ellipsis :content="item.title || item.name || '未命名素材'" />
              </div>
            </div>
          </van-grid-item>
        </van-grid>
        
        <!-- 分页加载 -->
        <van-list
          v-model:loading="materialLoading"
          v-model:error="materialError"
          :finished="materialFinished"
          error-text="请求失败，点击重新加载"
          finished-text="没有更多了"
          @load="loadMaterials"
        />
      </div>
      
      <!-- 空状态 -->
      <van-empty
        v-else-if="!materialLoading"
        image="https://fastly.jsdelivr.net/npm/@vant/assets/custom-empty-image.png"
        image-size="80"
        description="暂无素材"
      />
    </div>
    
    <!-- 底部确认按钮 -->
    <div class="material-footer" v-if="selectedMaterialId">
      <van-button 
        type="primary" 
        block 
        round
        @click="confirmSelect"
      >
        确认选择
      </van-button>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import { httpGet } from "@/utils/http";
import { showFailToast } from "vant";

const router = useRouter();
const route = useRoute();

// 素材库相关
const materials = ref([]);
const categories = ref([]);
const selectedCategory = ref("");
const selectedMaterialId = ref(null);
const selectedMaterial = ref(null);
const materialLoading = ref(false);
const materialError = ref(false);
const materialFinished = ref(false);
const materialPage = ref(1);
const materialPageSize = ref(20);

// 分类选项
const categoryOptions = computed(() => {
  const options = [{ text: "全部分类", value: "" }];
  categories.value.forEach(cat => {
    options.push({
      text: cat.name || cat.title || "未命名分类",
      value: cat.id || cat.value || ""
    });
  });
  return options;
});

// 返回上一页
const goBack = () => {
  router.back();
};

// 加载素材分类
const loadCategories = () => {
  httpGet("/api/mj/materials/categories")
    .then((res) => {
      categories.value = res.data || [];
    })
    .catch((e) => {
      console.warn("获取素材分类失败，使用默认分类:", e.message);
      categories.value = [];
    });
};

// 加载素材列表
const loadMaterials = () => {
  if (materialLoading.value || materialFinished.value) {
    return;
  }
  
  materialLoading.value = true;
  const categoryParam = selectedCategory.value ? `&category_id=${selectedCategory.value}` : "";
  
  httpGet(`/api/mj/materials/list?page=${materialPage.value}&page_size=${materialPageSize.value}${categoryParam}`)
    .then((res) => {
      const items = res.data.items || res.data || [];
      
      if (items.length < materialPageSize.value) {
        materialFinished.value = true;
      }
      
      if (materialPage.value === 1) {
        materials.value = items;
      } else {
        materials.value = materials.value.concat(items);
      }
      
      materialPage.value += 1;
      materialLoading.value = false;
    })
    .catch((e) => {
      materialLoading.value = false;
      materialError.value = true;
      console.error("获取素材列表失败:", e);
      if (e.message.includes("404") || e.message.includes("Not Found")) {
        materials.value = [];
        materialFinished.value = true;
      } else {
        showFailToast("获取素材列表失败：" + e.message);
      }
    });
};

// 分类改变
const onCategoryChange = () => {
  materialPage.value = 1;
  materialFinished.value = false;
  materials.value = [];
  loadMaterials();
};

// 选择素材
const selectMaterial = (material) => {
  selectedMaterialId.value = material.id;
  selectedMaterial.value = material;
};

// 确认选择并返回
const confirmSelect = () => {
  if (!selectedMaterial.value) {
    showFailToast("请先选择素材");
    return;
  }
  
  // 通过路由参数传递选中的素材信息
  router.push({
    path: "/mobile/image",
    query: {
      materialId: selectedMaterial.value.id,
      materialData: JSON.stringify(selectedMaterial.value)
    }
  });
};

onMounted(() => {
  // 加载素材分类和素材列表
  loadCategories();
  loadMaterials();
  
  // 如果从路由参数中获取已选择的素材ID，则选中它
  if (route.query.materialId) {
    selectedMaterialId.value = parseInt(route.query.materialId);
  }
});
</script>

<style lang="stylus">
.material-list-page
  min-height: 100vh
  background: #f7f8fa
  padding-bottom: 120px  // 为底部按钮和导航栏留出足够空间
  
  .material-content
    padding: 10px
    
    .category-filter
      margin-bottom: 10px
      background: #fff
      border-radius: 8px
      overflow: hidden
    
    .material-grid
      .material-item
        border: 2px solid #ebedf0
        border-radius: 8px
        overflow: hidden
        cursor: pointer
        transition: all 0.3s
        background: #fff
        position: relative
        
        &:active
          transform: scale(0.98)
        
        &.active
          border-color: #1989fa
          box-shadow: 0 0 8px rgba(25, 137, 250, 0.3)
        
        .material-image-wrapper
          position: relative
          width: 100%
          height: 120px
          
          .van-image
            width: 100%
            height: 100%
          
          .selected-badge
            position: absolute
            top: 5px
            right: 5px
            background: rgba(255, 255, 255, 0.9)
            border-radius: 50%
            width: 32px
            height: 32px
            display: flex
            align-items: center
            justify-content: center
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1)
        
        .material-title
          padding: 8px
          font-size: 12px
          color: #323233
          text-align: center
          background: #f7f8fa
          
    .material-placeholder
      width: 100%
      height: 120px
      display: flex
      align-items: center
      justify-content: center
      background: #f7f8fa
      color: #969799
      
      .iconfont
        font-size: 32px

  .material-footer
    position: fixed
    bottom: 50px  // 为底部导航栏留出空间（Vant tabbar 高度）
    left: 0
    right: 0
    padding: 10px
    padding-bottom: calc(10px + env(safe-area-inset-bottom))  // 适配安全区域
    background: #fff
    border-top: 1px solid #ebedf0
    box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.05)
    z-index: 100  // 确保在内容之上
</style>

