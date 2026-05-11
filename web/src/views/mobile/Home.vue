<template>
  <van-config-provider :theme="theme">
    <div class="mobile-home">
      <router-view/>

      <van-tabbar route v-model="active">
        <van-tabbar-item to="/mobile/chat" name="chat" icon="chat-o">对话</van-tabbar-item>
        <van-tabbar-item to="/mobile/image" name="image" icon="photo-o">图片</van-tabbar-item>
        <van-tabbar-item to="/mobile/index" name="home" icon="video-o">视频</van-tabbar-item>
        <van-tabbar-item to="/mobile/profile" name="profile" icon="user-o">我的
        </van-tabbar-item>
      </van-tabbar>

    </div>
  </van-config-provider>

</template>

<script setup>
import {ref, watch, onMounted} from "vue";
import {useSharedStore} from "@/store/sharedata";
import {useRoute} from "vue-router";
import {httpGet} from "@/utils/http";

const active = ref('home')
const store = useSharedStore()
const theme = ref(store.theme)
const route = useRoute()

watch(() => store.theme, (val) => {
  theme.value = val
})

// 设置微信分享的 meta 标签
const setWeChatShareMeta = (title, description, image, url) => {
  // 移除旧的 meta 标签
  const oldMetaTags = document.querySelectorAll('meta[property^="og:"], meta[name^="twitter:"]');
  oldMetaTags.forEach(tag => tag.remove());

  // 创建新的 meta 标签
  const metaTags = [
    { property: 'og:type', content: 'website' },
    { property: 'og:title', content: title },
    { property: 'og:description', content: description },
    { property: 'og:image', content: image },
    { property: 'og:url', content: url },
    { name: 'twitter:card', content: 'summary_large_image' },
    { name: 'twitter:title', content: title },
    { name: 'twitter:description', content: description },
    { name: 'twitter:image', content: image },
  ];

  metaTags.forEach(meta => {
    const tag = document.createElement('meta');
    if (meta.property) {
      tag.setAttribute('property', meta.property);
    } else {
      tag.setAttribute('name', meta.name);
    }
    tag.setAttribute('content', meta.content);
    document.head.appendChild(tag);
  });

  // 更新页面标题
  document.title = title;
};

// 检测referral参数，记录分享点击
onMounted(() => {
  const referral = route.query.referral;
  
  // 如果有 referral 参数，设置分享 meta 标签
  if (referral) {
    const shareUrl = `${location.protocol}//${location.host}${location.pathname}${location.search}`;
    const shareTitle = '邀请好友体验新功能 - So-AI';
    const shareDesc = '每成功邀请一位好友访问，您将获得 +50 积分奖励！';
    // 使用绝对路径，确保微信能正确抓取图片（建议图片尺寸：1200x630px）
    const shareImage = `${location.protocol}//${location.host}/images/logo.png`;
    
    // 设置微信分享 meta 标签
    setWeChatShareMeta(shareTitle, shareDesc, shareImage, shareUrl);
    
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
})

</script>

<style lang="stylus">
@import '@/assets/iconfont/iconfont.css';
.mobile-home {
  .container {
    .van-nav-bar {
      position fixed
      width 100%
    }

    padding 0 10px
  }

}

// 黑色主题
.van-theme-dark body {
  background #1c1c1e
}

.van-nav-bar {
  position fixed
  width 100%
}
</style>