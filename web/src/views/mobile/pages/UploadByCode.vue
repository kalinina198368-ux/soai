<template>
  <div class="upload-by-code-page">
    <van-nav-bar title="扫码上传图片" left-arrow @click-left="goBack" />

    <div class="content">
      <div class="tip-card">
        <div class="tip-title">匿名上传</div>
        <div class="tip-desc">
          您上传的图片会在24小时内自动销毁,请及时通知操作人员使用。
        </div>
        <div class="tip-code" v-if="code">上传码：{{ code }}</div>
      </div>

      <van-empty
        v-if="!code"
        image="https://fastly.jsdelivr.net/npm/@vant/assets/custom-empty-image.png"
        image-size="80"
        description="缺少上传码，请重新扫码打开"
      />

      <div v-else class="uploader-card">
        <van-uploader
          v-model="fileList"
          :after-read="uploadImg"
          :max-count="9"
          multiple
        />
        <div class="uploader-actions">
          <van-button type="primary" block round :loading="loading" :disabled="loading" @click="noop">
            {{ loading ? "上传中..." : "上传完成即可关闭页面" }}
          </van-button>
        </div>
      </div>

      <div v-if="uploadedUrls.length > 0" class="result-card">
        <div class="result-title">已上传（{{ uploadedUrls.length }}）</div>
        <van-grid :gutter="10" :column-num="3">
          <van-grid-item v-for="(u, idx) in uploadedUrls" :key="u + idx">
            <van-image :src="u" fit="cover" width="100%" height="100" @click="preview(u)" />
          </van-grid-item>
        </van-grid>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { showFailToast, showImagePreview, showSuccessToast } from "vant";
import Compressor from "compressorjs";
import { httpPost } from "@/utils/http";

const route = useRoute();
const router = useRouter();

const code = ref(route.query.code ? String(route.query.code).trim() : "");
const fileList = ref([]);
const uploadedUrls = ref([]);
const loading = ref(false);

const goBack = () => {
  router.back();
};

const preview = (url) => {
  showImagePreview([url]);
};

const noop = () => {};

// 图片上传（压缩后上传到上传码通道）
const uploadImg = (file) => {
  if (!code.value) {
    showFailToast("缺少上传码");
    return;
  }
  file.status = "uploading";
  loading.value = true;

  new Compressor(file.file, {
    quality: 0.7,
    success(result) {
      const formData = new FormData();
      formData.append("file", result, result.name);
      httpPost(`/api/upload_code/upload?code=${encodeURIComponent(code.value)}`, formData)
        .then((res) => {
          const url = res?.data?.url;
          if (url) {
            file.url = url;
            uploadedUrls.value.unshift(url);
            showSuccessToast("上传成功");
          } else {
            showFailToast("上传成功，但未返回图片地址");
          }
          file.status = "done";
        })
        .catch((e) => {
          file.status = "failed";
          file.message = "上传失败";
          showFailToast("上传失败：" + (e.message || "未知错误"));
        })
        .finally(() => {
          // 如果还有正在上传的文件，这里不强制置 false；简单处理为延迟关闭
          loading.value = false;
        });
    },
    error(err) {
      loading.value = false;
      showFailToast(err?.message || "图片处理失败");
    },
  });
};
</script>

<style lang="stylus" scoped>
.upload-by-code-page
  min-height: 100vh
  background: #f7f8fa

  :deep(.van-nav-bar)
    position: sticky
    top: 0
    z-index: 100

.content
  padding: 12px

.tip-card
  background: #fff
  border-radius: 12px
  padding: 14px
  box-shadow: 0 4px 12px rgba(0,0,0,.04)
  margin-bottom: 12px

  .tip-title
    font-size: 15px
    font-weight: 600
    color: #1a1a1a
    margin-bottom: 6px

  .tip-desc
    font-size: 13px
    color: #666
    line-height: 1.6

  .tip-code
    margin-top: 10px
    font-size: 12px
    color: #969799

.uploader-card
  background: #fff
  border-radius: 12px
  padding: 14px
  box-shadow: 0 4px 12px rgba(0,0,0,.04)

.uploader-actions
  margin-top: 12px

.result-card
  margin-top: 12px
  background: #fff
  border-radius: 12px
  padding: 14px
  box-shadow: 0 4px 12px rgba(0,0,0,.04)

  .result-title
    font-size: 14px
    font-weight: 600
    color: #1a1a1a
    margin-bottom: 10px

</style>

