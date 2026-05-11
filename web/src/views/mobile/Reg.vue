<template>
  <div class="login flex w-full flex-col place-content-center h-lvh">
    <div class="title text-center text-3xl font-bold mt-8">{{ title }}</div>
    <div class="w-full p-8">
      <login-dialog @success="loginSuccess" />
    </div>
  </div>
</template>

<script setup>
import LoginDialog from "@/components/RegDialog.vue";
import { getSystemInfo } from "@/store/cache";
import { useRouter, useRoute } from "vue-router";
import { ref, onMounted } from "vue";
import { httpGet } from "@/utils/http";

const router = useRouter();
const route = useRoute();
const title = ref("注册");

const loginSuccess = () => {
  router.push("/mobile/chat");
};

onMounted(() => {
  getSystemInfo().then((res) => {
    title.value = res.data.title;
  });
  
  // 检测referral参数，记录分享点击
  const referral = route.query.referral;
  if (referral) {
    // 调用API记录分享点击
    httpGet(`/api/invite/shareClick?code=${referral}`)
      .then((res) => {
        if (res.data && res.data.rewarded) {
          console.log("分享点击已记录，分享者获得1个算力奖励");
        } else {
          console.log("分享点击已记录（24小时内已奖励过）");
        }
      })
      .catch((e) => {
        console.error("记录分享点击失败:", e);
      });
  }
});
</script>

<style scoped lang="stylus">
.login {
  background: var(--theme-bg);
  transition: all 0.3s ease;

  .logo {
    background: #ffffff;
    border-radius: 50%;
  }

  .title {
    color: var(--text-theme-color);
  }
}
</style>
