<template>
  <div class="sora-video-page">
    <!-- 顶部导航 -->
    <div class="top-nav">
      <van-tabs v-model:active="activeTab" class="nav-tabs" animated>
        <van-tab title="AI视频" name="sora"></van-tab>
      </van-tabs>

      <!-- 视频广场入口：位于 AI 视频 右侧，可点击跳转到下方视频列表区域 -->
      <!-- <div class="video-square" @click="goVideoSquare">
        <div class="video-square-item">
          <div class="video-square-icon">
            <span class="play-icon">▶</span>
          </div>
          <div class="video-square-text">
            <div class="title">视频广场</div>
            <div class="subtitle">刷最新视频</div>
          </div>
        </div>
      </div> -->
     
    </div>

    <!-- 模式切换：普通模式 / 模板模式 -->
    <div class="mode-section">
      <van-tabs v-model:active="activeMode" @change="onModeChange" animated>
        <van-tab title="专业模式" name="normal"></van-tab>
        <van-tab title="智能模式" name="template"></van-tab>
      </van-tabs>
    </div>

     <!-- <div class="model-selector">
        <van-field
          v-model="formData.model"
          is-link
          readonly
          placeholder="Model: sora-2"
          @click="showModelPicker = true"
        />
      </div> -->

    <!-- 宽高比选择 -->
    <div class="aspect-ratio-section">
      <div class="aspect-ratios">
        <div 
          v-for="ratio in aspectRatios" 
          :key="ratio.value"
          class="aspect-ratio-item"
          :class="{ active: formData.aspectRatio === ratio.value }"
          @click="selectAspectRatio(ratio.value)"
        >
          <div class="ratio-icon" :class="ratio.icon"></div>
          <span>{{ ratio.label }}</span>
        </div>
      </div>
    </div>
    
    <!-- 时长选择 -->
    <div class="duration-section">
      <div class="duration-label">时长</div>
      <div class="duration-options">
        <div 
          class="duration-chip"
          :class="{ active: formData.duration === '10s' }"
          @click="selectDuration('10s')"
        >10s</div>
        <div 
          class="duration-chip"
          :class="{ active: formData.duration === '15s' }"
          @click="selectDuration('15s')"
        >15s</div>
      </div>
    </div>

    <!-- 模板模式：素材选择 -->
    <div v-if="activeMode === 'template'" class="template-section">
      <div class="text-line section-title">选择素材</div>
      <div class="text-line">
        <!-- 素材分类选择 -->
        <van-dropdown-menu v-if="categories.length > 0">
          <van-dropdown-item v-model="selectedCategory" :options="categoryOptions" @change="onCategoryChange" />
        </van-dropdown-menu>
        
        <!-- 素材列表 -->
        <div class="material-list" v-if="materials.length > 0">
          <van-grid :gutter="10" :column-num="2">
            <van-grid-item v-for="item in materials" :key="item.id">
              <div 
                :class="selectedMaterial?.id === item.id ? 'material-item active' : 'material-item'" 
                @click="selectMaterial(item)"
              >
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
        
        <!-- 已选择的提示词预览 -->
        <div class="text-line" v-if="selectedMaterial">
          <div class="section-title">已选择素材：{{ selectedMaterial.title || selectedMaterial.name }}</div>
          <van-field
            v-model="formData.prompt"
            rows="3"
            autosize
            type="textarea"
            placeholder="提示词已自动填充"
            readonly
          />
        </div>
      </div>
    </div>

    <!-- 视频描述输入（普通模式） -->
    <div v-else class="prompt-section">
      <van-field
        v-model="formData.prompt"
        type="textarea"
        placeholder="视频创作描述"
        rows="6"
        class="prompt-input"
        @input="handlePromptInput"
      />
    </div>

    <!-- 角色浮动条（类似 Sora 官网的角色选择） -->
    <div v-if="isLogin && roleList.length" class="role-floating-bar">
      <div class="role-floating-inner">
        <div class="role-avatars-scroll">
          <div class="role-avatar-item create">
            <div class="avatar-circle">
              <span class="plus">+</span>
            </div>
            <div class="avatar-name">Create</div>
          </div>
          <div
            v-for="role in roleList"
            :key="role.id"
            :class="['role-avatar-item', { selected: isRoleSelected(role) }]"
            @click="toggleRoleSelection(role)"
          >
            <div class="avatar-circle">
              <img
                v-if="role.sys_picture_url || role.profile_picture_url"
                :src="role.sys_picture_url || role.profile_picture_url"
                :alt="role.display_name || role.username"
              />
              <span v-else class="avatar-text">
                {{ (role.display_name || role.username || '?').slice(0, 1) }}
              </span>
            </div>
            <div
              v-if="isRoleSelected(role)"
              class="avatar-remove"
              @click.stop="removeRoleSelection(role)"
            >
              ×
            </div>
            <div class="avatar-name">
              {{ role.display_name || role.username || '角色' }}
            </div>
          </div>
        </div>
        <div class="role-bar-tip">
          点击角色，会自动在提示词中添加“@角色名 ”来出演视频
        </div>
      </div>
    </div>

    <!-- 图片上传和时长选择 -->
    <div class="controls-section">
      <!-- 图生视频模式：图片上传区域 -->
      <div v-if="isImageToVideo" class="image-upload-section">
        <div class="upload-header">
          <div class="upload-header-left">
            <span class="upload-title">上传图片 ({{ uploadedImages.length }}/10)</span>
            <!-- 使用须知：放在上传图片标题右侧 -->
            <div class="notes-section" @click="showNoticePopup = true">
              <van-icon name="question-o" class="notes-icon" />
              <span class="notes-trigger-text">使用须知</span>
            </div>
          </div>
          <van-button 
            v-if="uploadedImages.length > 0" 
            size="mini" 
            type="warning" 
            @click="clearAllImages"
          >
            清空
          </van-button>
        </div>

        <!-- 图片预览网格 -->
        <div class="image-grid">
          <div 
            v-for="image in uploadedImages" 
            :key="image.id" 
            class="image-item"
          >
            <img :src="image.previewUrl || image.url" :alt="image.name" />
            <div class="image-overlay">
              <van-icon name="cross" @click="removeImage(image.id)" />
            </div>
            <div class="image-name">{{ image.name }}</div>
          </div>
          
          <!-- 上传按钮 -->
          <div 
            v-if="uploadedImages.length < 10" 
            class="upload-item" 
            @click="uploadImages"
            :class="{ uploading: isUploading }"
          >
            <van-loading v-if="isUploading" type="spinner" size="24px" />
            <van-icon v-else name="plus" />
            <span>{{ isUploading ? '上传中...' : '添加图片' }}</span>
          </div>
        </div>
        </div>

      <!-- 文生视频模式：参考图片上传 -->
      <div v-else class="image-uploads">
        <!-- 参考图片预览 -->
        <div v-if="referenceImagePreview" class="reference-image-preview">
          <div class="preview-item">
            <img :src="referenceImagePreview" alt="参考图片预览" />
            <div class="preview-overlay">
              <van-icon name="cross" @click="removeReferenceImage" />
            </div>
          </div>
        </div>
        
        <!-- 上传按钮 -->
        <div 
          class="upload-item" 
          @click="uploadReferenceImage"
          :class="{ 'has-image': formData.referenceImage }"
        >
          <van-icon v-if="!formData.referenceImage" name="plus" />
          <van-icon v-else name="success" />
          <span>{{ formData.referenceImage ? '更换参考图片' : '参考图片' }}</span>
        </div>
      </div>
      
      
    </div>

    <!-- 生成按钮 -->
    <div class="generate-section">
      <van-button 
        type="primary" 
        size="large" 
        block
        @click="generateVideo"
        :loading="isGenerating"
        :disabled="activeMode === 'template' ? !selectedMaterial : !formData.prompt.trim()"
        class="generate-btn"
      >
        <van-icon name="play" v-if="!isGenerating" />
        <span>{{ isGenerating ? '生成中...' : (isImageToVideo ? '立即生成' : '立即生成') }}</span>
      </van-button>
    </div>

    <!-- 生成进度 -->
    <div v-if="isGenerating" class="progress-section">
      <van-card>
        <template #header>
          <div class="card-header">
            <van-icon name="clock-o" />
            <span>生成进度</span>
          </div>
        </template>
        <div class="progress-content">
          <van-progress 
            :percentage="generationProgress" 
            :show-pivot="false"
            stroke-width="8"
            color="#1989fa"
          />
          <p class="progress-text">{{ generationStatus }}</p>
          <van-loading type="spinner" size="24px" />
        </div>
      </van-card>
    </div>

    <!-- 视频列表（视频广场内容区域） -->
    <div
      v-if="isLogin"
      ref="videoListSectionRef"
      class="video-list-section"
      v-loading="loading"
    >
        <div class="video-list">
        <div v-for="(item, index) in list" :key="item.id || index" class="video-item">
            <div class="video-container">
            <!-- 视频内容：列表中不展示原生控制条，只做预览入口 -->
            <div v-if="item.status === 'completed' && item.video_url">
              <video 
                :src="item.video_url"
                preload="metadata"
                class="generated-video"
                @click="previewVideo(item)"
              >
                您的浏览器不支持视频播放
              </video>
              <div class="play-overlay" @click="previewVideo(item)">
                <van-icon name="play" />
            </div>
            </div>
            <div v-else-if="item.status === 'failed'" class="failed-container">
              <van-icon name="warning-o" />
              <span>生成失败</span>
            </div>
            <div v-else-if="item.status === 'processing'" class="processing-container">
              <div class="progress-wrapper">
                <van-loading type="spinner" size="24px" />
                <div class="progress-text">正在生成中...</div>
                <!-- 调试信息 -->

                <!-- <div style="background: yellow; padding: 5px; margin: 5px; font-size: 12px;">
                  调试: 状态={{ item.status }}, 进度={{ parseProgress(item) }}%, raw_data长度={{ item.raw_data ? item.raw_data.length : 0 }}
                </div> -->
                
                <van-progress
                  :percentage="parseProgress(item)"
                  color="#1989fa"
                  track-color="#f2f3f5"
                  stroke-width="8"
                  class="video-progress"
                />
                <div class="progress-percentage">{{ parseProgress(item) }}%</div>
              </div>
            </div>
            <div v-else class="generating-container">
              <van-loading type="spinner" size="24px" />
              <span>等待中...</span>
              <!-- 调试信息 -->
              <!-- <div style="background: orange; padding: 5px; margin: 5px; font-size: 12px;">
                调试: 状态={{ item.status }}, 进度={{ parseProgress(item) }}%, raw_data长度={{ item.raw_data ? item.raw_data.length : 0 }}
              </div> -->
            </div>
          </div>
            <div class="video-info">
            <div class="prompt-wrapper">
              <p class="video-prompt">
                {{ truncatePrompt(item.prompt) }}
              </p>
            </div>
              <div class="video-meta">
              <van-tag size="small" type="primary">{{ formatDurationDisplay(getItemDuration(item)) }}</van-tag>
              <van-tag size="small" type="success">{{ getAspectRatioLabel(getItemAspectRatio(item)) }}</van-tag>
              <!-- 详情 -->
              <van-tag 
                v-if="item.status === 'completed'"
                size="small" 
                type="default" 
                style="cursor: pointer;"
                @click="showDetail(item)"
              >
                <van-icon name="info-o" style="margin-right: 2px;" />
                详情
              </van-tag>
              <!-- <van-tag size="small" type="warning">{{ item.quality || 'hd' }}</van-tag> -->
              <van-tag v-if="item.is_favorite" size="small" type="danger">收藏</van-tag>
              </div>
            <div class="video-actions" v-if="item.status === 'completed'">
                  <!-- 角色定制 按钮（区分 3 种状态，可视化更清晰） -->
                  <van-button 
                    size="small"
                    plain
                    hairline
                    type="primary"
                    class="role-btn"
                    :class="[
                      'role-btn',
                      `role-btn--${getVideoRoleStatus(item)}`
                    ]"
                    @click="roleCustomize(item)"
                  >
                    <van-icon :name="getVideoRoleIcon(item)" />
                    {{ getVideoRoleLabel(item) }}
                  </van-button>

                <van-button 
                  size="small" 
                  type="primary" 
                @click="downloadVideo(item)"
                >
                  <van-icon name="down" />
                  下载
                </van-button>
                <van-button
                  size="small"
                  type="warning"
                  @click="removeVideo(item)"
                >
                  <van-icon name="delete-o" />
                  删除
                </van-button>
              </div>
            <div class="video-actions" v-else-if="item.status === 'failed'">
              <van-button
                size="small"
                type="danger"
                @click="removeVideo(item)"
              >
                <van-icon name="delete-o" />
                删除
              </van-button>
            </div>
          </div>
        </div>
        <div v-if="list.length === 0 && !loading" class="empty-state">
          <van-empty description="还没有生成任何视频，快去创作吧！" />
        </div>
        <!-- 加载更多提示 -->
        <div v-if="isLoadingMore" class="load-more-tip">
          <van-loading type="spinner" size="16px" />
          <span>加载中...</span>
        </div>
        <div v-if="finished && list.length > 0" class="load-more-tip finished">
          <span>没有更多了</span>
        </div>
      </div>
    </div>

    <!-- 示例提示词 -->
    <div class="examples-section">
      <van-card>
        <template #header>
          <div class="card-header">
            <van-icon name="bulb-o" />
            <span>创意示例</span>
          </div>
        </template>
        <div class="examples-list">
          <van-cell 
            v-for="(example, index) in examplePrompts" 
            :key="index"
            :title="example.title"
            :label="example.prompt"
            is-link
            @click="useExample(example)"
            class="example-item"
          />
        </div>
      </van-card>
    </div>

    <!-- 模型选择器弹窗 -->
    <van-popup v-model:show="showModelPicker" position="bottom">
      <van-picker
        :columns="modelOptions"
        @confirm="onModelConfirm"
        @cancel="showModelPicker = false"
      />
    </van-popup>

    <!-- 视频预览对话框 -->
    <van-popup
      v-model:show="showDialog"
      position="center"
      :style="{ width: '90%', height: '80%' }"
    >
      <div class="video-preview">
        <div class="preview-header">
          <span>视频预览</span>
          <van-icon name="cross" @click="closePreview" />
        </div>
        <video
          v-if="currentVideoUrl"
          ref="previewVideoRef"
          :src="currentVideoUrl"
          controls
          preload="auto"
          autoplay
          loop
          class="preview-video"
        >
          您的浏览器不支持视频播放
        </video>
      </div>
    </van-popup>

    <!-- 详情弹窗 -->
    <van-popup
      v-model:show="showDetailDialog"
      position="center"
      :style="{ width: '90%', maxHeight: '80%' }"
      round
    >
      <div class="detail-dialog" v-if="currentDetailItem">
        <div class="detail-header">
          <span>视频详情</span>
          <van-icon name="cross" @click="closeDetail" />
        </div>
        <div class="detail-content">
          <div class="detail-section">
            <div class="detail-label">
              <span>提示词</span>
              <van-button 
                v-if="currentDetailItem.prompt"
                size="mini" 
                type="primary" 
                plain
                @click="copyPrompt(currentDetailItem.prompt)"
              >
                <van-icon name="copy" style="margin-right: 4px;" />
                复制
              </van-button>
            </div>
            <div class="detail-prompt">{{ currentDetailItem.prompt || '无描述' }}</div>
          </div>
          <div v-if="getItemImages(currentDetailItem).length > 0" class="detail-section">
            <div class="detail-label">参考图片</div>
            <div class="detail-images">
              <div 
                v-for="(img, idx) in getItemImages(currentDetailItem)" 
                :key="idx"
                class="detail-image-item"
              >
                <img :src="img" :alt="`图片${idx + 1}`" @click="previewImage(img)" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </van-popup>

    <!-- 角色片段选择弹窗（从顶部下拉出现） -->
    <van-popup
      v-model:show="showRoleDialog"
      position="bottom"
      round
      :style="{ height: '90%' }"
    >
      <div class="role-dialog" v-if="currentRoleItem">
        <div class="role-header">
          <span>选择要提取角色的片段</span>
          <van-icon name="cross" @click="closeRoleDialog" />
        </div>
        <div class="role-content">
          <div class="role-video-wrapper">
            <video
              :src="currentRoleItem.video_url"
              ref="roleVideoRef"
              controls
              preload="metadata"
              class="role-video"
              @loadedmetadata="onRoleVideoLoaded"
            >
              您的浏览器不支持视频播放
            </video>
          </div>
          <div class="role-form">
            <div class="role-tip">
              从视频中选择一个时间段，系统会从该片段中提取主要人物或物品，用于后续复用。
            </div>
            <div
              class="role-thumbnail-strip"
              v-if="roleThumbnails.length"
            >
              <div
                v-for="thumb in roleThumbnails"
                :key="thumb.time"
                class="role-thumb-item"
                :class="{
                  active:
                    thumb.time >= roleRange[0] &&
                    thumb.time <= roleRange[1],
                }"
              >
                <img :src="thumb.url" :alt="`第 ${thumb.time.toFixed(1)} 秒`" />
              </div>
            </div>
            <div class="role-slider-wrapper">
              <van-slider
                v-model="roleRange"
                range
                :min="0"
                :max="roleDuration"
                :step="1"
                bar-height="4px"
                @update:model-value="handleRoleSliderInput"
              />
              <div class="role-time-display">
                <span>开始：{{ roleRange[0] }}s</span>
                <span>结束：{{ roleRange[1] }}s</span>
              </div>
            </div>
            <div class="role-duration-tip">
              视频总时长约 {{ formatDurationDisplay(roleDuration) }}，拖动滑块选择需要提取角色的时间段。
            </div>
            <van-button
              type="primary"
              block
              class="role-submit-btn"
              :loading="isRoleSubmitting"
              @click="confirmRoleCustomize"
            >
              确认提取
            </van-button>
          </div>
        </div>
      </div>
    </van-popup>

    <!-- 已有角色信息浮层（点击某个已有角色视频时展示） -->
    <van-popup
      v-model:show="showRolePicker"
      position="bottom"
      round
      :style="{ height: '40%' }"
    >
      <div v-if="currentRoleInfo" class="role-info-dialog">
        <div class="role-info-header">
          <span>已提取的角色</span>
          <van-icon name="cross" @click="showRolePicker = false" />
        </div>
        <div class="role-info-content">
          <div class="role-info-main">
            <div class="role-info-avatar">
              <img
                v-if="currentRoleInfo.sys_picture_url || currentRoleInfo.profile_picture_url"
                :src="currentRoleInfo.sys_picture_url || currentRoleInfo.profile_picture_url"
                :alt="currentRoleInfo.display_name || currentRoleInfo.username"
              />
              <span v-else>
                {{
                  (currentRoleInfo.display_name ||
                    currentRoleInfo.username ||
                    "?"
                  ).slice(0, 1)
                }}
              </span>
            </div>
            <div class="role-info-text">
              <div class="role-info-name">
                {{ currentRoleInfo.display_name || currentRoleInfo.username }}
              </div>
              <div class="role-info-username">
                @{{ currentRoleInfo.username }}
              </div>
            </div>
          </div>


          <!-- <div
            v-if="currentRoleInfo.permalink"
            class="role-info-link"
            @click="window.open(currentRoleInfo.permalink, '_blank')"
          >
            查看角色主页
          </div> -->


          <div class="role-info-tip">
            你可以在提示词中说明“@{{ currentRoleInfo.username }}”，系统会尽量保持人物外观一致。
          </div>
        </div>
      </div>
    </van-popup>

    <!-- 注意事项弹窗内容 -->
    <van-popup
      v-model:show="showNoticePopup"
      position="center"
      :style="{ width: '88%', maxWidth: '420px', borderRadius: '12px' }"
      round
    >
      <div class="notice-dialog">
        <div class="notice-header">
          <div class="title">
            <van-icon name="warning-o" />
            <span>使用须知</span>
          </div>
          <van-icon
            name="cross"
            class="close"
            @click.stop="showNoticePopup = false"
          />
        </div>
        <div class="notice-content">
          <ul>
            <li>1、提交的图片中是否涉及真人（非常像真人的也不行）。</li>
            <li>2、提示词内容是否违规（暴力、色情、版权、活着的名人）。</li>
            <li>3、生成结果审查是否合格（这也是大家经常看到的生成了 90% 多后失败的原因）。</li>
            <li>4、图片上传多张，其实 Sora 模型 99% 情况下只识别第一张，望周知。</li>
          </ul>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref, reactive, computed } from "vue";
import { useRouter } from "vue-router";
import { checkSession, getSystemInfo } from "@/store/cache";
import { httpGet, httpPost } from "@/utils/http";
import { showNotify, showToast, showConfirmDialog, showImagePreview } from "vant";
import { useSharedStore } from "@/store/sharedata";
import { showLoginDialog } from "@/utils/libs";

const title = ref(process.env.VUE_APP_TITLE);
const router = useRouter();
const isLogin = ref(false);
const slogan = ref("输入您的创意描述，AI将为您生成精彩的视频内容");

// 注意事项弹窗
const showNoticePopup = ref(false);

// 视频广场区域 DOM 引用（用于顶部入口点击后滚动到此处）
const videoListSectionRef = ref(null);

// 表单数据
const formData = reactive({
  prompt: "",
  model: "sora-2",
  aspectRatio: "9:16",
  duration: "15s",
  enhancePrompt: true,
  loop: false,
  referenceImage: null,
  endFrameImage: null,
});

// 图生视频相关状态
const isImageToVideo = ref(true);
const uploadedImages = ref([]);
const isUploading = ref(false);

// 文生视频参考图片状态
const referenceImagePreview = ref(null);

// 滚动监听相关
let scrollHandler = null;

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
    if (distanceToBottom < 100 && !finished.value && !isLoadingMore.value && !loading.value && isLogin.value) {
      loadMore();
    }
  };
  
  // 添加滚动监听
  window.addEventListener('scroll', scrollHandler, { passive: true });
};

onUnmounted(() => {
  stopProgressPolling();
  // 清理WebSocket事件监听器
  if (store.socket.conn && store.sora2MessageHandler) {
    store.socket.conn.removeEventListener("message", store.sora2MessageHandler);
    delete store.sora2MessageHandler;
  }
  // 清理滚动监听
  if (scrollHandler) {
    window.removeEventListener('scroll', scrollHandler);
    scrollHandler = null;
  }
  // 清理图片预览URL
  if (referenceImagePreview.value) {
    URL.revokeObjectURL(referenceImagePreview.value);
  }
  // 清理图生视频图片预览URL
  uploadedImages.value.forEach(img => {
    if (img.previewUrl) {
      URL.revokeObjectURL(img.previewUrl);
    }
  });
  // 清理视频预览
  if (previewVideoRef.value) {
    previewVideoRef.value.pause();
    previewVideoRef.value.currentTime = 0;
  }
});

// 生成状态
const isGenerating = ref(false);
const generationProgress = ref(0);
const generationStatus = ref("准备生成...");
const generatedVideos = ref([]);

// 选择器状态
const showModelPicker = ref(false);
const activeTab = ref("sora");
const activeMode = ref("normal"); // 模式切换：normal 普通模式, template 模板模式

const progressTimer = ref(null);
const currentTaskId = ref(null);
const store = useSharedStore();

// 视频列表相关
const loading = ref(false);
const list = ref([]);
const noData = ref(true);
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);
const finished = ref(false); // 是否已加载完所有数据
const isLoadingMore = ref(false); // 是否正在加载更多

// 视频预览相关
const showDialog = ref(false);
const currentVideoUrl = ref("");
const previewVideoRef = ref(null);

// 角色裁剪视频引用 & 缩略图
const roleVideoRef = ref(null);
const roleThumbnails = ref([]);
const isGeneratingRoleThumbs = ref(false);
const lastRoleRange = ref([0, 3]);
// 角色列表（用于提示词区域和快捷选择）
const roleList = ref([]);
const selectedRoleKeys = ref([]); // 已在提示词中使用的角色（username 列表）
const isLoadingRoles = ref(false);
const showRolePicker = ref(false);
const currentRoleInfo = ref(null);
// 已经检查过“是否有角色”的任务 ID 集合，避免重复请求
const checkedRoleTaskIds = new Set();

// 详情弹窗相关
const showDetailDialog = ref(false);
const currentDetailItem = ref(null);

// 角色定制弹窗相关
const showRoleDialog = ref(false);
const currentRoleItem = ref(null);
const roleDuration = ref(10);
const roleRange = ref([0, 3]);
const isRoleSubmitting = ref(false);

// 素材库相关
const materials = ref([]);
const categories = ref([]);
const selectedCategory = ref("");
const selectedMaterial = ref(null);
const materialLoading = ref(false);
const materialError = ref(false);
const materialFinished = ref(false);
const materialPage = ref(1);
const materialPageSize = ref(20);


// 宽高比选项
const aspectRatios = ref([
  { value: "16:9", label: "横屏", icon: "wide" },
  { value: "9:16", label: "竖屏", icon: "tall" },
]);

// 模型选项
const modelOptions = ref([
  { text: "sora-2", value: "sora-2" },
  // { text: "sora-2-pro", value: "sora-2-pro" },
]);

// 示例提示词
const examplePrompts = ref([
  {
    title: "自然风光",
    prompt: "一只可爱的小猫在花园里玩耍，阳光明媚，画面温馨，慢镜头拍摄",
  },
  {
    title: "城市夜景",
    prompt: "繁华都市的夜晚，霓虹灯闪烁，车流如织，航拍视角展现城市魅力",
  },
  {
    title: "科幻场景",
    prompt: "未来世界的机器人正在建造一座巨大的太空站，激光切割金属，火花四溅",
  },
  {
    title: "美食制作",
    prompt: "厨师在厨房里制作精美的蛋糕，奶油裱花，装饰水果，特写镜头",
  },
  {
    title: "运动场景",
    prompt: "运动员在跑道上冲刺，汗水飞溅，慢动作展现肌肉线条和力量感",
  },
  {
    title: "艺术创作",
    prompt: "画家在画布上挥洒颜料，色彩斑斓，创作过程充满艺术气息",
  },
]);

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

// 加载素材分类
const loadCategories = () => {
  httpGet("/api/sora2/materials/categories")
    .then((res) => {
      categories.value = res.data || [];
    })
    .catch((e) => {
      // 如果接口不存在，使用空数组
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
  
  httpGet(`/api/sora2/materials/list?page=${materialPage.value}&page_size=${materialPageSize.value}${categoryParam}`)
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
      // 如果接口不存在，使用示例数据
      if (e.message.includes("404") || e.message.includes("Not Found")) {
        materials.value = [];
        materialFinished.value = true;
      } else {
        showToast({
          type: 'fail',
          message: "获取素材列表失败：" + e.message
        });
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
  selectedMaterial.value = material;
  // 自动填充提示词
  formData.prompt = material.prompt || material.text || "";
  showToast({
    type: 'success',
    message: "已选择素材：" + (material.title || material.name || "未命名")
  });
};

// 模式切换
const onModeChange = (mode) => {
  if (mode === "template") {
    // 切换到模板模式时，加载素材分类和素材列表
    if (categories.value.length === 0) {
      loadCategories();
    }
    if (materials.value.length === 0) {
      loadMaterials();
    }
  } else {
    // 切换到普通模式时，清空选择的素材
    selectedMaterial.value = null;
  }
};

// 初始化
onMounted(() => {
  console.log("🎬 Sora2 Index.vue mounted - 调试模式");
  console.log("formData:", formData);
  console.log("Initial isLogin state:", isLogin.value);
  
  // 强制设置标题和标语
  title.value = "🎬 Sora2 视频生成";
  slogan.value = "输入您的创意描述，AI将为您生成精彩的视频内容";
  
  getSystemInfo()
    .then((res) => {
      title.value = "🎬 Sora2 视频生成";
      slogan.value = "输入您的创意描述，AI将为您生成精彩的视频内容";
      console.log("System info loaded, but using Sora2 title:", title.value, slogan.value);
    })
    .catch((e) => {
      console.log("获取系统配置失败，使用默认值");
      title.value = "🎬 Sora2 视频生成";
      slogan.value = "输入您的创意描述，AI将为您生成精彩的视频内容";
    });

  checkSession()
    .then((user) => {
      isLogin.value = true;
      console.log("User logged in:", user);
      fetchData(1);
      // 登录后预加载当前用户角色列表
      loadAllRoles();
    })
    .catch(() => {
      console.log("User not logged in");
      isLogin.value = false;
    });

  // 设置WebSocket消息处理 - 等待连接建立
  setupWebSocketHandler();
  
  // 添加滚动监听，实现下拉加载更多
  setupScrollListener();
});

// 设置WebSocket消息处理函数
const setupWebSocketHandler = () => {
  const checkAndSetup = () => {
    // console.log("🔌 检查WebSocket连接状态:", store.socket.conn);
    // console.log("🔌 WebSocket readyState:", store.socket.conn?.readyState);
    // console.log("🔌 WebSocket URL:", store.socket.conn?.url);
    
    if (store.socket.conn && store.socket.conn.readyState === WebSocket.OPEN) {
      console.log("✅ WebSocket已连接，设置消息处理");
      
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
        // 丢弃无关消息 - 注意：sora2不需要检查clientId，因为后端发送给所有客户端
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
          
          // 更新全局进度
          if (body.progress) {
            const progress = parseInt(body.progress, 10) || 0;
            generationProgress.value = Math.max(0, Math.min(100, progress));
            console.log("🌐 更新全局进度:", generationProgress.value);
          }

          // 更新列表中对应视频项的进度
          if (body.job_id) {
            const jobId = parseInt(body.job_id);
            const videoIndex = list.value.findIndex(item => item.id === jobId);
            
            console.log(`🔍 查找视频项: job_id=${jobId}, 找到索引: ${videoIndex}`);
            
            if (videoIndex !== -1) {
              console.log(`📝 更新前 - 状态: ${list.value[videoIndex].status}, 进度: ${parseProgress(list.value[videoIndex])}%`);
              
              // 使用Vue 3的响应式更新方式
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
                // 视频完成后，立即检查是否有角色可用（静默预加载）
                // currentItem 已经是 list.value 中的项，所以可以直接传入
                const taskId = currentItem.task_id || currentItem.taskId || currentItem.id;
                if (taskId && !checkedRoleTaskIds.has(taskId)) {
                  checkedRoleTaskIds.add(taskId);
                  loadRolesByTask(taskId, currentItem, { silent: true });
                }
              } else if (body.status === "FAILURE") {
                currentItem.status = "failed";
              }
              
              console.log(`📝 更新后 - 状态: ${currentItem.status}, 进度: ${parseProgress(currentItem)}%`);
            } else {
              console.log(`❌ 未找到对应的视频项，job_id: ${jobId}`);
            }
          }

          // 处理任务完成
          if (body.status === "SUCCESS" || body.status === "completed") {
            isGenerating.value = false;
            generationStatus.value = "生成完成";
            showToast({
              type: 'success',
              message: "视频生成完成！"
            });
            fetchData(1);
          } else if (body.status === "FAILURE" || body.status === "failed") {
            isGenerating.value = false;
            generationStatus.value = body.err_msg || body.fail_reason || "生成失败";
            showToast({
              type: 'fail',
              message: generationStatus.value
            });
            fetchData(1);
          }
        }
      };

      // 添加事件监听器
      store.socket.conn.addEventListener("message", handleWebSocketMessage);
      
      // 保存处理器引用，用于清理
      store.sora2MessageHandler = handleWebSocketMessage;
      
      console.log("✅ WebSocket消息处理器已设置");
    } else {
      console.log("⏳ WebSocket未连接，1秒后重试");
      setTimeout(checkAndSetup, 1000);
    }
  };
  
  checkAndSetup();
};

// 选择器确认事件
const onModelConfirm = ({ selectedValues }) => {
  formData.model = selectedValues[0];
  showModelPicker.value = false;
};

// 选择宽高比
const selectAspectRatio = (ratio) => {
  formData.aspectRatio = ratio;
};

// 选择时长
const selectDuration = (duration) => {
  formData.duration = duration;
};

const startProgressPolling = (taskId) => {
  stopProgressPolling();
};

const stopProgressPolling = () => {
  if (progressTimer.value) {
    clearInterval(progressTimer.value);
    progressTimer.value = null;
  }
};

// 上传参考图片（文生视频模式）
const uploadReferenceImage = () => {
  if (!isLogin.value) {
    return showLoginDialog(router);
  }
  
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
    const formData_upload = new FormData();
    formData_upload.append("file", file);
    
    const response = await httpPost("/api/upload", formData_upload);
    
    // 将相对路径转换为绝对路径
    let imageUrl = response.data.url;
    if (!imageUrl.startsWith('http')) {
      // 如果是相对路径，转换为绝对路径
      imageUrl = location.protocol + "//" + location.host + imageUrl;
    }
    
    formData.referenceImage = imageUrl; // 使用HTTP URL
    
    // 创建预览URL
    referenceImagePreview.value = URL.createObjectURL(file);
    
    showToast({
      type: 'success',
      message: "参考图片上传成功"
    });
  } catch (error) {
    showToast({
      type: 'fail',
      message: "参考图片上传失败：" + error.message
    });
  }
};

// 删除参考图片
const removeReferenceImage = () => {
  if (referenceImagePreview.value) {
    URL.revokeObjectURL(referenceImagePreview.value);
  }
  referenceImagePreview.value = null;
  formData.referenceImage = null;
  
  showToast({
    type: 'success',
    message: "参考图片已删除"
  });
};

// 图片上传相关函数
const uploadImages = () => {
  if (!isLogin.value) {
    return showLoginDialog(router);
  }
  
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
        id: Date.now() + Math.random(),
        name: file.name,
        size: file.size,
        url: imageUrl, // 使用服务器返回的URL
        previewUrl: URL.createObjectURL(file) // 用于本地预览
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
    // 释放预览URL对象
    if (uploadedImages.value[index].previewUrl) {
      URL.revokeObjectURL(uploadedImages.value[index].previewUrl);
    }
    uploadedImages.value.splice(index, 1);
    showToast({
      type: 'success',
      message: "图片已删除"
    });
  }
};

const clearAllImages = () => {
  uploadedImages.value.forEach(img => {
    // 释放预览URL对象
    if (img.previewUrl) {
      URL.revokeObjectURL(img.previewUrl);
    }
  });
  uploadedImages.value = [];
  showToast({
    type: 'success',
    message: "已清空所有图片"
  });
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

// 使用示例提示词
const useExample = (example) => {
  formData.prompt = example.prompt;
  syncSelectedRolesWithPrompt(formData.prompt);
  showToast({
    type: 'success',
    message: `已应用示例：${example.title}`
  });
};

// 解析raw_data中的进度
const parseProgress = (item) => {
  const rawDataStr = item.raw_data;
  console.log(`🔍 parseProgress: 解析项目 ${item.id}, raw_data:`, rawDataStr);
  
  if (!rawDataStr) {
    console.log(`❌ parseProgress: 没有raw_data，返回0`);
    return 0;
  }
  
  try {
    const rawData = JSON.parse(rawDataStr);
    const progress = parseInt(rawData.progress || "0", 10) || 0;
    console.log(`✅ parseProgress: 解析成功，进度: ${progress}%`);
    return progress;
  } catch (e) {
    console.error("❌ parseProgress: 解析raw_data失败:", e);
    return 0;
  }
};

// 获取视频列表
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
  
  httpGet("/api/sora2/list", {
    page: page.value,
    page_size: pageSize.value,
  })
    .then((res) => {
      total.value = res.data.total;
      const newData = (res.data.data || []).map((item) => {
        // 进入列表时，直接根据 prompt/prompt_ext 判断一次“是否角色视频”
        // 这样角色按钮的样式一开始就正确，而不是等用户点了才变
        try {
          if (isRoleBasedVideo(item)) {
            // 这里不需要额外字段，getVideoRoleStatus 会再次调用 isRoleBasedVideo
          }
        } catch (e) {
          console.warn("判断角色视频状态失败", e);
        }
        return item;
      });
      
      if (isLoadMore) {
        // 加载更多时，追加数据
        list.value = [...list.value, ...newData];
        isLoadingMore.value = false;
      } else {
        // 首次加载或刷新时，替换数据
        list.value = newData;
        loading.value = false;
      }
      
      noData.value = list.value.length === 0;
      
      // 判断是否已加载完所有数据
      const totalPages = Math.ceil(total.value / pageSize.value);
      finished.value = page.value >= totalPages || newData.length < pageSize.value;

      // 对当前页中已完成的视频，预先去后端查一次"是否已生成过角色"（静默检查）
      // 只更新 item.has_role，不弹出角色选择弹窗
      // 这样列表加载后就能知道每个视频是否有角色可用，无需点击按钮才知道
      // 注意：使用 list.value 中的项，确保响应式更新
      const itemsToCheck = isLoadMore ? newData : list.value;
      itemsToCheck.forEach((item) => {
        if (!item || item.status !== "completed") return;
        const taskId = item.task_id || item.taskId || item.id;
        if (!taskId) return;
        // 避免重复请求：如果已经检查过，跳过
        if (checkedRoleTaskIds.has(taskId)) return;
        checkedRoleTaskIds.add(taskId);
        // 静默预加载角色信息，只更新 item.has_role 字段
        // 传入 list.value 中的项，确保响应式更新
        const listItem = list.value.find(li => {
          const liTaskId = li.task_id || li.taskId || li.id;
          return liTaskId === taskId;
        });
        loadRolesByTask(taskId, listItem || item, { silent: true });
      });
    })
    .catch((error) => {
      console.error("获取视频列表失败:", error);
      if (isLoadMore) {
        isLoadingMore.value = false;
      } else {
        loading.value = false;
      }
      noData.value = true;
      finished.value = false;
    });
};

// 加载更多数据
const loadMore = () => {
  if (finished.value || isLoadingMore.value || loading.value) {
    return;
  }
  
  page.value += 1;
  fetchData(page.value, true);
};

// 顶部「视频广场」入口点击：跳转到专属视频广场页面
const goVideoSquare = () => {
  router.push({ name: "mobile-video-square" });
};


// 处理输入框输入事件
const handlePromptInput = (value) => {

  // console.log("handlePromptInput:", value);

  if (!isLogin.value) {
    return showLoginDialog(router);
  }

  // 同步选中状态：如果用户手动删除了某个 @角色 标签，则自动取消选中
  syncSelectedRolesWithPrompt(value);

  // 当输入达到4个字符且未登录时，提示登录
  // if (value && value.length >= 4 && !isLogin.value) {
  //   showLoginDialog(router);
  //   showToast({
  //     type: 'warning',
  //     message: "请先登录后再继续输入"
  //   });
  // }
};

// 生成视频
const generateVideo = async () => {
  // 模板模式下需要先选择素材
  if (activeMode.value === "template") {
    if (!selectedMaterial.value) {
      showToast({
        type: 'fail',
        message: "请先选择素材！"
      });
      return;
    }
  } else {
    if (!formData.prompt.trim()) {
      showToast({
        type: 'fail',
        message: "请输入视频描述"
      });
      return;
    }
  }
  
  // 图生视频模式下图片上传是可选的，不再强制要求
  
  if (!isLogin.value) {
    return showLoginDialog(router);
  }

  try {
    isGenerating.value = true;
    generationProgress.value = 0;
    generationStatus.value = "正在提交生成任务...";

    let createPayload;
    
    if (isImageToVideo.value) {
      // 图生视频模式
      createPayload = {
        prompt: formData.prompt,
        model: formData.model,
        images: uploadedImages.value.map(img => img.url), // 使用HTTP URL
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
    currentTaskId.value = taskId;
    generationStatus.value = "任务已提交，等待进度...";

    showToast({
      type: 'success',
      message: createRes?.data?.message || "任务已提交，正在生成中..."
    });
    
    // 模板模式下，重置选择的素材
    if (activeMode.value === "template") {
      selectedMaterial.value = null;
      formData.prompt = "";
    }
    
    // 立即刷新视频列表，显示新创建的 processing 任务
    fetchData(1);
  } catch (error) {
    isGenerating.value = false;
    generationProgress.value = 0;
    showToast({
      type: 'fail',
      message: error.message || "生成失败，请稍后重试"
    });
  }
};

// 下载视频
const downloadVideo = async (item) => {
  try {
    const videoUrl = item.video_url;
    if (!videoUrl) {
      showToast({
        type: 'fail',
        message: "视频链接不存在"
      });
      return;
    }

    const link = document.createElement("a");
    link.href = videoUrl;
    link.download = `sora2_video_${item.id}.mp4`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    
    showToast({
      type: 'success',
      message: "视频下载开始"
    });
  } catch (error) {
    showToast({
      type: 'fail',
      message: "下载失败：" + error.message
    });
  }
};

// 切换收藏状态
const toggleFavorite = (item) => {
  httpGet("/api/sora2/favorite", { id: item.id })
    .then((res) => {
      item.is_favorite = res.data.is_favorite;
      showToast({
        type: 'success',
        message: res.data.message
      });
    })
    .catch((e) => {
      showToast({
        type: 'fail',
        message: "操作失败：" + e.message
      });
    });
};

// 删除视频
const removeVideo = (item) => {
  showConfirmDialog({
    title: '确认删除',
    message: '确定要删除这个视频吗？删除后无法恢复。',
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    confirmButtonColor: '#ff4444'
  }).then(() => {
    // 用户确认删除
    httpGet("/api/sora2/remove", { id: item.id })
      .then(() => {
        showToast({
          type: 'success',
          message: "视频删除成功"
        });
        fetchData(1);
      })
      .catch((e) => {
        showToast({
          type: 'fail',
          message: "视频删除失败：" + e.message
        });
      });
  }).catch(() => {
    // 用户取消删除，不做任何操作
  });
};

// 预览视频
const previewVideo = (item) => {
  currentVideoUrl.value = item.video_url;
  showDialog.value = true;
};

// 显示详情
const showDetail = (item) => {
  currentDetailItem.value = item;
  showDetailDialog.value = true;
};

// 关闭详情
const closeDetail = () => {
  showDetailDialog.value = false;
  currentDetailItem.value = null;
};

// 复制提示词
const copyPrompt = async (text) => {
  if (!text) return;
  try {
    await navigator.clipboard.writeText(text);
    showToast({
      message: '提示词已复制',
      type: 'success',
      duration: 2000,
    });
  } catch (err) {
    // 降级方案：使用传统方法
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.opacity = '0';
    document.body.appendChild(textArea);
    textArea.select();
    try {
      document.execCommand('copy');
      showToast({
        message: '提示词已复制',
        type: 'success',
        duration: 2000,
      });
    } catch (e) {
      showToast({
        message: '复制失败，请手动复制',
        type: 'fail',
        duration: 2000,
      });
    }
    document.body.removeChild(textArea);
  }
};

// 预览图片
const previewImage = (url) => {
  showImagePreview({
    images: [url],
    startPosition: 0,
  });
};

// 关闭预览
const closePreview = () => {
  // 停止视频播放
  if (previewVideoRef.value) {
    previewVideoRef.value.pause();
    previewVideoRef.value.currentTime = 0;
  }
  showDialog.value = false;
  currentVideoUrl.value = "";
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

// 从后端返回的数据中解析时长，优先 raw_data，其次 task_info，最后回退默认
const getItemDuration = (item) => {
  // 1) 尝试从 raw_data 解析（进度/完成时服务器推送的数据体）
  try {
    if (item?.raw_data) {
      const data = JSON.parse(item.raw_data);
      const v = data?.params?.duration || data?.duration;
      if (v !== undefined && v !== null && String(v).trim() !== "") {
        return String(v);
      }
    }
  } catch (_) {}

  // 2) 尝试从 task_info 解析（创建任务时保存的入参）
  try {
    if (item?.task_info) {
      const info = JSON.parse(item.task_info);
      // 常见结构：{ params: { duration, aspect_ratio, ... }, ... }
      const v = info?.params?.duration || info?.duration || info?.task?.params?.duration;
      if (v !== undefined && v !== null && String(v).trim() !== "") {
        return String(v);
      }
    }
  } catch (_) {}

  // 3) 回退
  return String(item?.duration || "10");
};

// 从后端返回的数据中解析宽高比，优先 raw_data，其次 task_info，最后回退默认
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

// 打开角色定制弹窗
const roleCustomize = async (item) => {
  if (!isLogin.value) {
    return showLoginDialog(router);
  }

  // 如果该视频本身就是使用某个角色制作的（在 prompt_ext 中包含 @username）
  // 则不再允许从该视频再次提取角色，直接给出提示
  if (isRoleBasedVideo(item)) {
    showToast({
      type: "warning",
      message: "该视频已使用角色创作，无需再次从该视频提取角色",
    });
    return;
  }

  if (!item || !item.video_url) {
    showToast({
      type: "fail",
      message: "当前视频还未生成完成，无法提取角色",
    });
    return;
  }

  // 先检查该视频是否已经有提取好的角色
  const hasRole = await loadRolesByTask(item.task_id || item.taskId || item.id, item);
  
  // 如果已经有角色，只显示角色信息弹窗，不显示视频选择弹窗
  if (hasRole) {
    return;
  }

  // 如果视频还没有角色，显示视频选择弹窗用于提取角色
  currentRoleItem.value = item;

  // 默认选择前 3 秒
  let durationStr = getItemDuration(item);
  let durationNum = parseInt(String(durationStr), 10);
  if (isNaN(durationNum) || durationNum <= 0) {
    durationNum = 10;
  }
  roleDuration.value = durationNum;
  roleRange.value = [0, Math.min(3, durationNum)];

  showRoleDialog.value = true;
};

// 判断一个视频是否为“角色视频”
// 后端有时会把带 @角色名 的内容放在 prompt_ext，有时直接放在 prompt 里
// 这里做一个兼容：优先用 prompt_ext，没有则退回到 prompt
const isRoleBasedVideo = (item) => {
  if (!item) return false;
  const ext = item.prompt_ext;
  const base = item.prompt;
  if (!ext && !base) return false;
  const text = String(ext || base);

  // 如果已经有角色列表，则精确匹配 @username
  if (Array.isArray(roleList.value) && roleList.value.length > 0) {
    const hasMatch = roleList.value.some((role) => {
      const name = role?.username || role?.display_name;
      if (!name) return false;
      const tag = "@" + name;
      return text.includes(tag);
    });
    if (hasMatch) return true;
  }

  // 兜底：简单判断是否包含 “@xxx” 结构（防止后端还没返回角色列表时漏判）
  return /@\w+/.test(text);
};

// 获取当前视频的“角色状态”
// - can_generate: 可生成角色（普通视频，尚未从该视频提取过角色）
// - has_role: 已成功生成角色（从该视频提取过角色，但视频本身不是用角色生成的）
// - role_based: 当前视频本身就是使用角色生成的
const getVideoRoleStatus = (item) => {
  if (!item) return "can_generate";
  if (isRoleBasedVideo(item)) {
    return "role_based";
  }
  if (item.has_role) {
    return "has_role";
  }
  return "can_generate";
};

// 角色按钮展示文案
const getVideoRoleLabel = (item) => {
  const status = getVideoRoleStatus(item);
  if (status === "role_based") return "角色创作";
  if (status === "has_role") return "角色可用";
  return "提取角色";
};

// 角色按钮图标
const getVideoRoleIcon = (item) => {
  const status = getVideoRoleStatus(item);
  if (status === "role_based") return "friends-o";
  if (status === "has_role") return "contact";
  return "user-o";
};

// 加载当前用户的全部角色列表（用于提示词区域展示）
const loadAllRoles = async () => {
  if (!isLogin.value || isLoadingRoles.value) return;
  try {
    isLoadingRoles.value = true;
    const res = await httpGet("/api/sora2/roles", {});
    roleList.value = Array.isArray(res?.data?.data) ? res.data.data : [];
  } catch (e) {
    // 静默失败即可，不影响主流程
    console.warn("加载角色列表失败", e);
  } finally {
    isLoadingRoles.value = false;
  }
};

// 按任务加载角色（用于点击视频「角色」按钮时判断是否已有角色）
// 可选传入 videoItem，用于在列表中标记"该视频已成功生成角色"
// options:
//   - silent: true 时只更新 videoItem.has_role，不弹出角色选择浮层（用于列表预加载）
// 返回 hasRole 布尔值，表示是否有角色
const loadRolesByTask = async (taskId, videoItem = null, options = {}) => {
  if (!isLogin.value || !taskId) return false;
  const { silent = false } = options;
  try {
    const res = await httpGet("/api/sora2/roles", { task_id: taskId });
    const roleList = Array.isArray(res?.data?.data) ? res.data.data : [];
    const hasRole = roleList.length > 0;

    // 在对应视频项上记录一个标记字段，方便列表渲染状态
    // 优先更新 list.value 中对应的项，确保响应式更新
    if (videoItem) {
      // 先尝试在 list.value 中找到对应的项并更新（确保响应式）
      const taskIdToMatch = videoItem.task_id || videoItem.taskId || videoItem.id;
      const listItem = list.value.find(item => {
        const itemTaskId = item.task_id || item.taskId || item.id;
        return itemTaskId === taskIdToMatch || itemTaskId === taskId;
      });
      
      if (listItem) {
        // 直接更新 list.value 中的项，确保响应式更新
        listItem.has_role = hasRole;
      } else {
        // 如果找不到，直接更新传入的 videoItem（兼容处理）
        videoItem.has_role = hasRole;
      }
    }

    // 非 silent 模式下，点击按钮时才弹出角色选择浮层
    if (!silent) {
      if (hasRole) {
        currentRoleInfo.value = roleList[0];
        showRolePicker.value = true;
      } else {
        currentRoleInfo.value = null;
        showRolePicker.value = false;
      }
    }
    
    return hasRole;
  } catch (e) {
    console.warn("按任务加载角色失败", e);
    return false;
  }
};

// 关闭角色定制弹窗
const closeRoleDialog = () => {
  showRoleDialog.value = false;
  currentRoleItem.value = null;
};

// 视频元信息加载完成时，自动更新总时长
const onRoleVideoLoaded = (e) => {
  try {
    const d = e?.target?.duration;
    if (!isNaN(d) && d > 0) {
      const dur = Math.floor(d);
      roleDuration.value = dur || 1;
      // 确保当前选择区间不越界
      const [s, ed] = roleRange.value || [0, 0];
      roleRange.value = [
        Math.min(Math.max(0, s), dur),
        Math.min(Math.max(0, ed || 3), dur),
      ];

      // 元数据就绪后，异步生成缩略图条
      generateRoleThumbnails(e.target);
    }
  } catch (_) {}
};

// 生成角色视频缩略图（纯前端 Canvas 截帧）
const generateRoleThumbnails = async (videoEl) => {
  try {
    if (!videoEl || isGeneratingRoleThumbs.value) return;
    const duration = Number(videoEl.duration) || 0;
    if (!duration || duration <= 0) return;

    isGeneratingRoleThumbs.value = true;
    const maxThumbs = 20; // 最多 20 张缩略图
    const count = Math.min(Math.ceil(duration), maxThumbs);

    const canvas = document.createElement("canvas");
    const ctx = canvas.getContext("2d");
    if (!ctx) {
      isGeneratingRoleThumbs.value = false;
      return;
    }

    // 简单按 16:9 比例生成缩略图
    const width = 160;
    const height = 90;
    canvas.width = width;
    canvas.height = height;

    const frames = [];

    const seekAndCapture = (time) =>
      new Promise((resolve) => {
        const handleSeeked = () => {
          try {
            ctx.drawImage(videoEl, 0, 0, width, height);
            const url = canvas.toDataURL("image/jpeg");
            frames.push({ time, url });
          } catch (_) {
            // 跨域视频可能无法截帧，直接忽略错误
          }
          videoEl.removeEventListener("seeked", handleSeeked);
          resolve();
        };
        videoEl.addEventListener("seeked", handleSeeked);
        videoEl.currentTime = time;
      });

    // 记录当前播放状态，生成后恢复
    const prevPaused = videoEl.paused;
    const prevTime = videoEl.currentTime;
    videoEl.pause();

    for (let i = 0; i < count; i++) {
      const t = (duration * i) / Math.max(count - 1, 1);
      await seekAndCapture(t);
    }

    // 恢复播放状态
    videoEl.currentTime = prevTime;
    if (!prevPaused) {
      videoEl.play().catch(() => {});
    }

    roleThumbnails.value = frames;
  } catch (_) {
    // 忽略所有生成缩略图的异常，避免影响主流程
  } finally {
    isGeneratingRoleThumbs.value = false;
  }
};

// 拖动角色滑块时，实时更新视频当前帧，模拟官网裁剪体验
const handleRoleSliderInput = (val) => {
  // 这里 val 是 [start, end]
  const prev = lastRoleRange.value || [roleRange.value[0], roleRange.value[1]];
  roleRange.value = val;
  const [start, end] = val || [0, 0];

  const videoEl = roleVideoRef.value;
  if (!videoEl) return;

  // 根据哪一侧变化更大来判断是拖动开始还是结束手柄
  const diffStart = Math.abs(Number(start) - Number(prev[0] ?? start));
  const diffEnd = Math.abs(Number(end) - Number(prev[1] ?? end));

  let previewTime;
  if (diffStart > diffEnd) {
    // 用户主要在拖动「开始」手柄
    previewTime = Number(start);
  } else if (diffEnd > 0) {
    // 用户主要在拖动「结束」手柄
    previewTime = Number(end);
  } else {
    // 回退：都没明显变化时，使用开始时间
    previewTime = Number(start);
  }

  previewTime = Math.max(0, Math.min(roleDuration.value, previewTime));

  try {
    videoEl.currentTime = previewTime;
    videoEl.pause();
  } catch (_) {
    // 忽略偶发的设置 currentTime 异常
  }

  // 记录本次区间，下一次用来判断哪一侧在变化
  lastRoleRange.value = [start, end];
};

// 确认角色提取
const confirmRoleCustomize = async () => {
  if (!isLogin.value) {
    return showLoginDialog(router);
  }
  if (!currentRoleItem.value) {
    return;
  }

  const [startRaw, endRaw] = roleRange.value || [0, 0];
  const start = Math.floor(Number(startRaw));
  const end = Math.floor(Number(endRaw));

  if (isNaN(start) || isNaN(end)) {
    showToast({
      type: "fail",
      message: "请输入有效的开始和结束时间",
    });
    return;
  }

  if (start < 0 || end <= start) {
    showToast({
      type: "fail",
      message: "结束时间必须大于开始时间，且开始时间不能小于 0",
    });
    return;
  }

  const diff = end - start;
  if (diff < 1 || diff > 3) {
    showToast({
      type: "fail",
      message: "时间范围差值必须在 1～3 秒之间，视频片段不能超过3秒",
    });
    return;
  }

  let durationStr = getItemDuration(currentRoleItem.value);
  let durationNum = parseInt(String(durationStr), 10);
  if (!isNaN(durationNum) && end > durationNum) {
    showToast({
      type: "fail",
      message: "结束时间不能超过视频总时长",
    });
    return;
  }

  const timestamps = `${start},${end}`;

  try {
    isRoleSubmitting.value = true;
    const res = await httpPost("/api/sora2/characters", {
      task_id: currentRoleItem.value.task_id,
      timestamps,
    });

    showToast({
      type: "success",
      message: res?.data?.message || "角色提取任务已提交",
    });
    showRoleDialog.value = false;
    // 重新加载角色列表（全局 + 当前任务），并在当前视频项上标记 has_role
    loadAllRoles();
    loadRolesByTask(currentRoleItem.value.task_id, currentRoleItem.value);
  } catch (e) {
    showToast({
      type: "fail",
      message: e?.message || "角色提取失败，请稍后重试",
    });
  } finally {
    isRoleSubmitting.value = false;
  }
};

// 角色选择相关工具函数
const getRoleKey = (role) => {
  const username = role?.username;
  return username ? String(username) : null;
};

const getRoleTag = (roleOrKey) => {
  const username =
    typeof roleOrKey === "string" ? roleOrKey : roleOrKey?.username;
  // 注意：末尾一定要带一个空格，方便 Sora2 正确识别角色标签
  return username ? `@${username} ` : "";
};

const isRoleSelected = (role) => {
  const key = getRoleKey(role);
  if (!key) return false;
  return selectedRoleKeys.value.includes(key);
};

const toggleRoleSelection = (role) => {
  if (!role) return;
  if (isRoleSelected(role)) {
    removeRoleSelection(role);
  } else {
    addRoleSelection(role);
  }
};

const addRoleSelection = (role) => {
  const key = getRoleKey(role);
  if (!key || isRoleSelected(role)) return;
  selectedRoleKeys.value.push(key);
  addRoleTagToPrompt(role);
};

const removeRoleSelection = (role) => {
  const key = getRoleKey(role);
  if (!key) return;
  selectedRoleKeys.value = selectedRoleKeys.value.filter(
    (username) => username !== key
  );
  removeRoleTagFromPrompt(key);
};

const addRoleTagToPrompt = (role) => {
  const tag = getRoleTag(role);
  if (!tag) return;
  const text = formData.prompt || "";
  if (text.includes(tag)) return;
  formData.prompt = text ? `${text}\n${tag}` : tag;
};

const removeRoleTagFromPrompt = (roleOrKey) => {
  const tag = getRoleTag(roleOrKey);
  if (!tag) return;
  const text = formData.prompt || "";
  if (!text.includes(tag)) return;
  // 直接移除所有该标签（包含后面的空格），再整理多余空格和空行
  let next = text.replaceAll(tag, "");
  // 合并多余空行
  next = next.replace(/\n{3,}/g, "\n\n");
  // 去掉首尾空白
  formData.prompt = next.trim();
};

const syncSelectedRolesWithPrompt = (currentText) => {
  // 兼容 Vant 输入事件在某些情况下传入对象/非字符串的情况，统一转成字符串
  const raw = currentText ?? formData.prompt ?? "";
  const text = typeof raw === "string" ? raw : String(raw || "");

  selectedRoleKeys.value = selectedRoleKeys.value.filter((username) => {
    const tag = getRoleTag(username);
    return tag && text.includes(tag);
  });
};

// 截断prompt，最多显示50个字
const truncatePrompt = (prompt) => {
  if (!prompt) return "";
  const maxLength = 50;
  if (prompt.length <= maxLength) {
    return prompt;
  }
  return prompt.substring(0, maxLength) + "...";
};

// 从后端返回的数据中解析图片，优先 task_info，其次 raw_data
const getItemImages = (item) => {
  const images = [];
  
  // 1) 尝试从 task_info 解析（创建任务时保存的入参）
  try {
    if (item?.task_info) {
      const info = JSON.parse(item.task_info);
      const imgList = info?.images || info?.params?.images || [];
      if (Array.isArray(imgList) && imgList.length > 0) {
        images.push(...imgList);
      }
    }
  } catch (_) {}
  
  // 2) 尝试从 raw_data 解析（进度/完成时服务器推送的数据体）
  try {
    if (item?.raw_data) {
      const data = JSON.parse(item.raw_data);
      const imgList = data?.images || data?.params?.images || [];
      if (Array.isArray(imgList) && imgList.length > 0) {
        // 合并但不重复
        imgList.forEach(img => {
          if (!images.includes(img)) {
            images.push(img);
          }
        });
      }
    }
  } catch (_) {}
  
  return images;
};
</script>

<!--  -->

<style scoped>
.generated-video {
  width: 100%;
  max-width: 400px;
  height: auto;
  display: block;
}
.video-container {
  position: relative;
}
.play-overlay {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  cursor: pointer;
}
</style>




<style scoped lang="stylus">
// Sora2 视频生成页面样式 - 与系统主题保持一致
.sora-video-page
  min-height: 100vh
  background: var(--theme-bg)
  color: var(--text-color)
  padding: 20px

// 顶部导航
.top-nav
  display: flex
  justify-content: space-between
  align-items: center
  margin-bottom: 20px

// 模式切换
.mode-section
  margin-bottom: 20px
  
  :deep(.van-tabs__wrap)
    border-bottom: 1px solid var(--border-color)
    background: transparent
  
  :deep(.van-tabs__nav)
    background: transparent
  
  :deep(.van-tab)
    font-size: 14px
    color: var(--text-color)
    background: transparent
    
    &.van-tab--active
      color: #00d4aa
      font-weight: 600
  
  .nav-tabs
    flex: 1
    margin-right: 16px
    
    // 隐藏标签面板内容（因为只有一个标签，不需要显示面板）
    :deep(.van-tab__panel)
      display: none
    
    // 去除白色背景
    :deep(.van-tabs__wrap)
      border-bottom: none
      background: transparent
    
    :deep(.van-tabs__nav)
      background: transparent
    
    :deep(.van-tab)
      font-size: 14px
      color: var(--text-color)
      background: transparent
      
      &.van-tab--active
        color: var(--text-color)
        font-weight: 600
        background: transparent
  
  // 顶部右侧「视频广场」入口样式
  .video-square
    min-width: 110px
    cursor: pointer
    
    .video-square-item
      display: flex
      align-items: center
      gap: 8px
      padding: 8px 10px
      border-radius: 999px
      background: rgba(0, 212, 170, 0.12)
      border: 1px solid rgba(0, 212, 170, 0.5)
      transition: all 0.2s ease
      
      .video-square-icon
        width: 32px
        height: 32px
        border-radius: 50%
        background: #00d4aa
        display: flex
        align-items: center
        justify-content: center
        box-shadow: 0 6px 12px rgba(0, 212, 170, 0.3)
        
        .play-icon
          font-size: 14px
          color: #fff
          transform: translateX(1px)
      
      .video-square-text
        display: flex
        flex-direction: column
        line-height: 1.2
        
        .title
          font-size: 13px
          font-weight: 600
          color: var(--text-color)
        
        .subtitle
          font-size: 11px
          color: #00d4aa
    
    &:active
      transform: scale(0.97)
      opacity: 0.9

  .model-selector
    .van-field
      background: var(--card-bg)
      border: 1px solid var(--border-color)
      border-radius: 6px
      color: var(--text-color)
      
      .van-field__control
        color: var(--text-color)

// 宽高比选择
.aspect-ratio-section
  margin-bottom: 30px
  
  .aspect-ratios
    display: flex
    gap: 12px
    flex-wrap: wrap
    
    .aspect-ratio-item
      display: flex
      flex-direction: column
      align-items: center
      padding: 12px 16px
      background: var(--card-bg)
      border: 2px solid var(--border-color)
      border-radius: 8px
      cursor: pointer
      transition: all 0.3s ease
      min-width: 60px
      
      &.active
        border-color: #00d4aa
        background: #00d4aa20
      
      .ratio-icon
        width: 24px
        height: 16px
        margin-bottom: 8px
        background: #666
        border-radius: 2px
        
        &.square
          width: 16px
          height: 16px
        &.portrait
          width: 12px
          height: 16px
        &.landscape
          width: 16px
          height: 12px
        &.wide
          width: 20px
          height: 12px
        &.tall
          width: 12px
          height: 20px
        &.ultrawide
          width: 24px
          height: 10px
      
      span
        font-size: 12px
        color: var(--text-color)

// 时长选择
.duration-section
  display: flex
  align-items: center
  justify-content: space-between
  gap: 12px
  margin-bottom: 16px
  
  .duration-label
    font-size: 14px
    color: var(--text-color)
    font-weight: 600
  
  .duration-options
    display: flex
    gap: 10px
    
    .duration-chip
      min-width: 54px
      text-align: center
      padding: 8px 14px
      border-radius: 999px
      background: var(--card-bg)
      border: 2px solid var(--border-color)
      color: var(--text-color)
      cursor: pointer
      user-select: none
      transition: all 0.2s ease
      
      &.active
        background: #00d4aa
        border-color: #00d4aa
        color: #000
        font-weight: 700

// 视频描述输入
.prompt-section
  /* 为下面的角色浮动条预留一点空间，让两者形成一个整体卡片区域 */
  margin-bottom: 18px
  
  .prompt-input
    background: var(--card-bg)
    border: 1px solid var(--border-color)
    border-radius: 8px
    
    .van-field__control
      color: var(--text-color)
      font-size: 16px
      
    &::placeholder
      color: var(--text-color-secondary)

// 模板模式
.template-section
  margin-bottom: 18px
  
  .text-line
    margin-bottom: 16px
    padding: 0 4px
    
    &.section-title
      font-size: 14px
      font-weight: 600
      color: var(--text-color)
      margin-bottom: 12px
      padding: 0
  
  .material-list
    margin-top: 10px
    
  .material-item
    border: 2px solid #ebedf0
    border-radius: 8px
    overflow: hidden
    cursor: pointer
    transition: all 0.3s
    background: #fff
    
    &:active
      transform: scale(0.98)
    
    &.active
      border-color: #1989fa
      box-shadow: 0 0 8px rgba(25, 137, 250, 0.3)
    
    .van-image
      width: 100%
      height: 120px
      
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

// 控制区域
.controls-section
  display: flex
  justify-content: space-between
  align-items: flex-start
  margin-bottom: 30px
  gap: 20px
  
  .image-uploads
    display: flex
    gap: 12px
    flex-wrap: wrap
    
    .reference-image-preview
      .preview-item
        position: relative
        width: 80px
        height: 80px
        border-radius: 8px
        overflow: hidden
        background: var(--card-bg)
        border: 1px solid var(--border-color)
        
        img
          width: 100%
          height: 100%
          object-fit: cover
        
        .preview-overlay
          position: absolute
          top: 0
          right: 0
          background: rgba(0, 0, 0, 0.6)
          border-radius: 0 8px 0 8px
          padding: 4px
          cursor: pointer
          
          .van-icon
            color: white
            font-size: 16px
    
    .upload-item
      display: flex
      flex-direction: column
      align-items: center
      padding: 16px
      background: var(--card-bg)
      border: 2px dashed var(--border-color)
      border-radius: 8px
      cursor: pointer
      transition: all 0.3s ease
      min-width: 80px
      
      &:hover
        border-color: #00d4aa
      
      &.has-image
        border-color: #00d4aa
        background: rgba(0, 212, 170, 0.1)
        
        .van-icon
          color: #00d4aa
      
      .van-icon
        font-size: 24px
        color: var(--text-color-secondary)
        margin-bottom: 8px
      
      span
        font-size: 12px
        color: var(--text-color-secondary)

// 图片上传区域
.image-upload-section
  flex: 1
  
  .upload-header
    display: flex
    justify-content: space-between
    align-items: center
    margin-bottom: 16px
    
    .upload-header-left
      display: inline-flex
      align-items: center

    .upload-title
      font-size: 14px
      font-weight: 600
      color: var(--text-color)
  
  .image-grid
    display: grid
    grid-template-columns: repeat(auto-fill, minmax(80px, 1fr))
    gap: 12px
    
    .image-item
      position: relative
      aspect-ratio: 1
      border-radius: 8px
      overflow: hidden
      background: var(--card-bg)
      border: 1px solid var(--border-color)
      
      img
        width: 100%
        height: 100%
        object-fit: cover
      
      .image-overlay
        position: absolute
        top: 0
        right: 0
        background: rgba(0, 0, 0, 0.6)
        border-radius: 0 8px 0 8px
        padding: 4px
        cursor: pointer
        
        .van-icon
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
        padding: 2px 4px
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
      background: var(--card-bg)
      border: 2px dashed var(--border-color)
      border-radius: 8px
      cursor: pointer
      transition: all 0.3s ease
      
      &.uploading
        border-color: #00d4aa
        background: rgba(0, 212, 170, 0.1)
      
      &:hover
        border-color: #00d4aa
      
      .van-icon, .van-loading
        font-size: 20px
        color: var(--text-color-secondary)
        margin-bottom: 4px
      
      span
        font-size: 10px
        color: var(--text-color-secondary)
        text-align: center
  
  .duration-selector
    display: flex
    gap: 8px
    
    .duration-item
      padding: 8px 16px
      background: var(--card-bg)
      border: 2px solid var(--border-color)
      border-radius: 6px
      cursor: pointer
      transition: all 0.3s ease
      color: var(--text-color-secondary)
      
      &.active
        background: #00d4aa
        border-color: #00d4aa
        color: #000
        font-weight: 600

// 开关选项
.toggle-section
  display: flex
  justify-content: space-between
  margin-bottom: 30px
  
  .toggle-item
    display: flex
    align-items: center
    gap: 12px
    
    .toggle-label
      color: var(--text-color)
      font-size: 14px

// 生成按钮
.generate-section
  margin-bottom: 30px
  
  .generate-btn
    height: 48px
    font-size: 16px
    font-weight: 600
    border-radius: 8px
    background: #00d4aa
    border: none
    color: #000
    
    .van-icon
      margin-right: 8px
    
    &:hover
      filter: brightness(1.05)
  
// 注意事项触发入口
.notes-section
  display: inline-flex
  align-items: center
  gap: 6px
  padding: 4px 8px
  border-radius: 999px
  background: rgba(255, 87, 87, 0.06)
  color: #ff5757
  font-size: 12px
  margin-left: 8px
  cursor: pointer
  user-select: none

  .notes-icon
    color: #ff5757
    font-size: 14px

  .notes-trigger-text
    font-size: 12px
    color: #ff5757

// 进度区域
.progress-section
  margin: 20px 0
  
  .van-card
    background: var(--card-bg)
    border: 1px solid var(--border-color)
    border-radius: 8px
    box-shadow: 0 2px 10px rgba(0,0,0,0.08)

.progress-content
  padding: 20px
  text-align: center
  
  .progress-text
    margin: 16px 0
    font-size: 14px
    color: var(--text-color)
    font-weight: 500

// 视频列表区域
.video-list-section
  margin: 20px 0
  
  .van-card
    background: var(--card-bg)
    border: 1px solid var(--border-color)
    border-radius: 8px

// 视频列表
.video-list
  padding: 16px

// 加载更多提示
.load-more-tip
  display: flex
  align-items: center
  justify-content: center
  padding: 16px
  color: var(--text-color-secondary)
  font-size: 14px
  gap: 8px
  
  &.finished
    color: var(--text-color-secondary)
    opacity: 0.6

.video-item
  background: var(--card-bg)
  border: 1px solid var(--border-color)
  border-radius: 8px
  overflow: hidden
  margin-bottom: 16px
  box-shadow: 0 2px 10px rgba(0,0,0,0.06)
  
  &:last-child
    margin-bottom: 0

.video-container
  position: relative
  background: #000
  border-radius: 8px
  overflow: hidden
  
  .generated-video
    width: 100%
    height: 200px
    object-fit: cover
    cursor: pointer
  
  .play-overlay
    position: absolute
    top: 50%
    left: 50%
    transform: translate(-50%, -50%)
    background: rgba(0, 0, 0, 0.6)
    border-radius: 50%
    width: 60px
    height: 60px
    display: flex
    align-items: center
    justify-content: center
    cursor: pointer
    transition: all 0.3s ease
    
    .van-icon
      color: white
      font-size: 24px
      margin-left: 4px
    
    &:hover
      background: rgba(0, 0, 0, 0.8)
  
  .failed-container,
  .generating-container
    height: 200px
    display: flex
    flex-direction: column
    align-items: center
    justify-content: center
    color: var(--text-color-secondary)
    
    .van-icon
      font-size: 48px
      margin-bottom: 12px
      color: #ff6b6b
    
    span
      font-size: 14px

  .processing-container
    height: 200px
    display: flex
    flex-direction: column
    align-items: center
    justify-content: center
    color: var(--text-color-secondary)
    
    .progress-wrapper
      display: flex
      flex-direction: column
      align-items: center
      width: 100%
      padding: 20px
      
      .van-loading
        margin-bottom: 16px
      
      .progress-text
        font-size: 14px
        font-weight: 500
        margin-bottom: 16px
        color: var(--text-color)
      
      .video-progress
        width: 100%
        max-width: 300px
        margin-bottom: 8px
      
      .progress-percentage
        font-size: 12px
        font-weight: 600
        color: #1989fa

.video-info
  padding: 16px
  
  .prompt-wrapper
    margin-bottom: 12px
  
  .video-prompt
    font-size: 14px
    color: var(--text-color)
    margin: 0
    line-height: 1.4
    position: relative
    display: block
    width: 100%
    word-break: break-word
  
  .video-meta
    margin-bottom: 12px
    
    .van-tag
      margin-right: 8px
      margin-bottom: 4px
  
  .video-actions
    display: flex
    gap: 8px
    
    .van-button
      flex: 1
      background: var(--card-bg)
      border: 1px solid var(--border-color)
      color: var(--text-color)
      
      .van-icon
        margin-right: 4px

      &.role-btn
        position: relative
        border-radius: 999px
        padding: 0 10px
        font-size: 12px

        // 默认状态：可生成角色（普通视频，尚未从该视频提取过角色）
        &.role-btn--can_generate
          border-color: #00bcd4
          color: #00bcd4
          background: rgba(0, 188, 212, 0.06)

        // 状态 2：已成功生成角色（但不是角色视频）
        &.role-btn--has_role
          border-color: #e6a23c
          color: #e6a23c
          background: rgba(230, 162, 60, 0.06)

          &::after
            content: ""
            position: absolute
            top: 4px
            right: 6px
            width: 6px
            height: 6px
            border-radius: 50%
            background: #e6a23c

        // 状态 3：当前视频就是使用角色生成的
        &.role-btn--role_based
          border-color: #67c23a
          color: #67c23a
          background: rgba(103, 194, 58, 0.08)

          &::after
            content: ""
            position: absolute
            top: 4px
            right: 6px
            width: 8px
            height: 8px
            border-radius: 50%
            background: #67c23a

// 视频预览样式
.video-preview
  background: var(--card-bg)
  border-radius: 8px
  overflow: hidden
  
  .preview-header
    display: flex
    justify-content: space-between
    align-items: center
    padding: 16px
    background: var(--theme-bg)
    border-bottom: 1px solid var(--border-color)
    
    span
      font-size: 16px
      font-weight: 600
      color: var(--text-color)
    
    .van-icon
      font-size: 20px
      color: var(--text-color-secondary)
      cursor: pointer
      
      &:hover
        color: var(--text-color)
  
  .preview-video
    width: 100%
    height: calc(80vh - 60px)
    object-fit: contain
    background: #000

// 分页样式
.pagination-section
  margin: 20px 0
  display: flex
  justify-content: center
  
  .van-pagination
    background: var(--card-bg)
    border-radius: 8px
    padding: 8px

// 选择器样式
.van-picker
  background: var(--card-bg)
  
  .van-picker__toolbar
    background: var(--theme-bg)
    border-bottom: 1px solid var(--border-color)
  
  .van-picker__confirm
    color: #00d4aa

// 开关样式
.van-switch
  --van-switch-on-background: #00d4aa

// 响应式优化
@media (max-width: 768px)
  .sora-video-page
    padding: 16px
  
  .top-nav
    flex-direction: row
    gap: 12px
    align-items: center
    
    .nav-tabs
      width: 100%
      justify-content: center
    
    .video-square
      flex-shrink: 0
      
    .model-selector
      width: 100%
  
  .aspect-ratios
    justify-content: center
    
    .aspect-ratio-item
      min-width: 50px
      padding: 8px 12px
  
  .controls-section
    flex-direction: column
    gap: 20px
    align-items: stretch
    
    .image-uploads
      justify-content: center
      
    .image-upload-section
      .image-grid
        grid-template-columns: repeat(auto-fill, minmax(70px, 1fr))
        gap: 8px
      
    .duration-selector
      justify-content: center
  
  .toggle-section
    flex-direction: column
    gap: 16px

// 动画效果
.aspect-ratio-item,
.duration-item,
.upload-item
  transition: all 0.3s ease
  
  &:hover
    transform: translateY(-2px)

.generate-btn
  transition: all 0.3s ease
  
  &:active
    transform: scale(0.98)

// 详情弹窗样式
.detail-dialog
  background: var(--card-bg)
  border-radius: 8px
  overflow: hidden
  
  .detail-header
    display: flex
    justify-content: space-between
    align-items: center
    padding: 16px
    background: var(--theme-bg)
    border-bottom: 1px solid var(--border-color)
    
    span
      font-size: 16px
      font-weight: 600
      color: var(--text-color)
    
    .van-icon
      font-size: 20px
      color: var(--text-color-secondary)
      cursor: pointer
      
      &:hover
        color: var(--text-color)
  
  .detail-content
    padding: 16px
    max-height: 60vh
    overflow-y: auto
    
    .detail-section
      margin-bottom: 20px
      
      &:last-child
        margin-bottom: 0
      
      .detail-label
        font-size: 14px
        font-weight: 600
        color: var(--text-color)
        margin-bottom: 8px
        display: flex
        align-items: center
        justify-content: space-between
      
      .detail-prompt
        font-size: 14px
        color: var(--text-color)
        line-height: 1.6
        word-break: break-word
        white-space: pre-wrap
        padding: 12px
        background: var(--theme-bg)
        border-radius: 6px
        border: 1px solid var(--border-color)
      
      .detail-images
        display: flex
        flex-wrap: wrap
        gap: 12px
        
        .detail-image-item
          width: 100px
          height: 100px
          border-radius: 8px
          overflow: hidden
          background: var(--theme-bg)
          border: 1px solid var(--border-color)
          cursor: pointer
          transition: transform 0.2s ease
          
          &:hover
            transform: scale(1.05)
          
          img
            width: 100%
            height: 100%
            object-fit: cover
            display: block

// 注意事项弹窗样式
.notice-dialog
  background: var(--card-bg)
  border-radius: 12px
  overflow: hidden

  .notice-header
    display: flex
    align-items: center
    justify-content: space-between
    padding: 12px 14px
    border-bottom: 1px solid var(--border-color)
    background: var(--theme-bg)

    .title
      display: flex
      align-items: center
      gap: 6px

      .van-icon
        color: #ff5757
        font-size: 18px

      span
        font-size: 15px
        font-weight: 600
        color: var(--text-color)

    .close
      font-size: 18px
      color: var(--text-color-secondary)

  .notice-content
    padding: 12px 16px 16px

    ul
      list-style: none
      padding: 0
      margin: 0

      li
        position: relative
        padding-left: 0
        margin-bottom: 6px
        font-size: 13px
        line-height: 1.5
        color: var(--text-color-secondary)

        &:last-child
          margin-bottom: 0
</style>

<style scoped>
.role-dialog {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--card-bg);
}
.role-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
}
.role-header span {
  font-size: 16px;
  font-weight: 600;
}
.role-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 12px 16px 16px;
  overflow-y: auto;
}
.role-video-wrapper {
  margin-bottom: 12px;
  max-height: 220px; /* 控制视频区域不要太高，方便下面操作滑块 */
  display: flex;
  justify-content: center;
  align-items: center;
}
.role-video {
  width: 100%;
  height: 100%;
  max-height: 220px;
  border-radius: 8px;
  background: #000;
  object-fit: contain; /* 保持画面完整，不会被裁剪 */
}
.role-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.role-tip {
  font-size: 13px;
  color: var(--text-color-secondary);
}
.role-duration-tip {
  font-size: 12px;
  color: var(--text-color-secondary);
}
.role-submit-btn {
  margin-top: 8px;
}
.role-slider-wrapper {
  margin-top: 8px;
}
.role-time-display {
  display: flex;
  justify-content: space-between;
  margin-top: 6px;
  font-size: 12px;
  color: var(--text-color-secondary);
}

/* 角色浮动条（提示词下方，类似 Sora 官网） */
.role-floating-bar {
  position: relative;
  /* 轻微上移一点，形成“压在输入框下面”的悬浮感 */
  margin-top: -6px;
  margin-bottom: 22px;
  z-index: 1;
}
.role-floating-inner {
  border-radius: 18px;
  background: var(--card-bg);
  padding: 10px 18px;
  color: var(--text-color);
  border: 1px solid rgba(0, 0, 0, 0.05);
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.18);
}
.role-avatars-scroll {
  display: flex;
  align-items: center;
  gap: 12px;
  overflow-x: auto;
  padding-bottom: 4px;
}
.role-avatar-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 56px;
  cursor: pointer;
  position: relative;
}
.role-avatar-item .avatar-circle {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  background: var(--theme-bg);
  border: 2px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}
.role-avatar-item .avatar-circle img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.role-avatar-item .avatar-circle .avatar-text {
  font-size: 16px;
  color: var(--text-color);
}
.role-avatar-item .avatar-name {
  margin-top: 4px;
  font-size: 11px;
  color: var(--text-color-secondary);
  text-align: center;
  max-width: 60px;
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow: hidden;
}
.role-avatar-item.selected .avatar-circle {
  border-color: #00d4aa;
  background: rgba(0, 212, 170, 0.12);
  box-shadow: 0 0 0 4px rgba(0, 212, 170, 0.12);
}
.role-avatar-item .avatar-remove {
  position: absolute;
  top: 0;
  right: 0;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #ff6b6b;
  color: #fff;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
  cursor: pointer;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.15);
}
.role-avatar-item.create .avatar-circle {
  border-style: dashed;
}
.role-avatar-item.create .plus {
  font-size: 22px;
  line-height: 1;
}
.role-bar-tip {
  margin-top: 4px;
  font-size: 10px;
  color: var(--text-color-secondary);
}

/* 已有角色信息弹层 */
.role-info-dialog {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 12px 16px 16px;
}
.role-info-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}
.role-info-header span {
  font-size: 16px;
  font-weight: 600;
}
.role-info-content {
  flex: 1;
  padding-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.role-info-main {
  display: flex;
  align-items: center;
  gap: 12px;
}
.role-info-avatar {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  overflow: hidden;
  background: #333;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #fff;
}
.role-info-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.role-info-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.role-info-name {
  font-size: 16px;
  font-weight: 600;
}
.role-info-username {
  font-size: 13px;
  color: var(--text-color-secondary);
}
.role-info-link {
  font-size: 13px;
  color: #409eff;
  cursor: pointer;
}
.role-info-tip {
  font-size: 12px;
  color: var(--text-color-secondary);
}
</style>