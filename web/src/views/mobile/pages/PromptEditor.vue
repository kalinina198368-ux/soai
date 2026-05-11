<template>
  <div class="prompt-editor-page">
    <van-nav-bar
      title="编辑提示词"
      left-arrow
      @click-left="goBack"
    />

    <div class="editor-content">
      <div class="textarea-wrapper">
        <van-field
          v-model="promptText"
          type="textarea"
          placeholder="请输入或粘贴提示词"
          show-word-limit
          maxlength="5000"
        />
      </div>

      <div class="helper-text">
        尽量用简短准确的描述，包含主体、动作、风格等信息。
      </div>
    </div>

    <div class="editor-actions">
      <van-button type="primary" block round @click="savePrompt">
        保存并返回
      </van-button>
      <van-button class="mt-8" type="default" block round @click="clearPrompt">
        清空
      </van-button>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { showToast } from "vant";

const route = useRoute();
const router = useRouter();

const promptText = ref("");

onMounted(() => {
  promptText.value = route.query.prompt ? String(route.query.prompt) : "";
});

const goBack = () => {
  router.back();
};

const savePrompt = () => {
  const newQuery = { ...route.query, prompt: promptText.value.trim() };
  router.push({
    path: "/mobile/image",
    query: newQuery,
  });
  showToast({ type: "success", message: "已保存提示词" });
};

const clearPrompt = () => {
  promptText.value = "";
};
</script>

<style lang="stylus" scoped>
.prompt-editor-page
  display: flex
  flex-direction: column
  height: 100vh
  background: #f7f8fa

  // 覆盖全局 fixed 定位，使用 sticky 避免遮挡内容
  :deep(.van-nav-bar)
    position: sticky
    top: 0
    z-index: 100
    flex-shrink: 0

.editor-content
  flex: 1
  display: flex
  flex-direction: column
  padding: 16px
  overflow: hidden
  min-height: 0

.textarea-wrapper
  flex: 1
  display: flex
  flex-direction: column
  min-height: 0
  overflow: hidden

  :deep(.van-cell)
    flex: 1
    display: flex
    flex-direction: column
    padding: 12px 16px
    min-height: 0
    height: 100%

  :deep(.van-field__body)
    flex: 1
    display: flex
    flex-direction: column
    min-height: 0
    height: 100%

  :deep(textarea)
    flex: 1
    width: 100%
    height: 100%
    min-height: 0
    resize: none
    overflow-y: auto
    border: none
    outline: none

.helper-text
  margin-top: 8px
  font-size: 12px
  color: #888
  line-height: 1.5
  flex-shrink: 0

.editor-actions
  padding: 16px
  background: #fff
  flex-shrink: 0

.mt-8
  margin-top: 8px
</style>
