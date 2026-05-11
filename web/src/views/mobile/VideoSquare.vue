<template>
  <div class="mobile-video-square">
    <van-nav-bar
      title="视频广场"
      left-text="返回"
      left-arrow
      @click-left="goBack"
      fixed
      safe-area-inset-top
      class="nav"
    />

    <div class="feed" ref="feedContainer">
      <section
        v-for="(video, index) in featuredVideos"
        :key="video.id"
        class="feed-item"
        :data-index="index"
        :ref="(el) => setItemRef(el, index)"
      >
        <video
          v-if="video.videoUrl"
          :src="video.videoUrl"
          :poster="video.poster"
          preload="metadata"
          playsinline
          webkit-playsinline
          muted
          loop
          class="feed-video"
          :ref="(el) => setVideoRef(el, index)"
          @click="togglePlay(index)"
          @play="handleVideoPlay(index)"
          @pause="handleVideoPause(index)"
          @loadedmetadata="() => incrementView(video.id)"
        >
          您的浏览器不支持视频播放
        </video>

        <button
          class="play-overlay"
          type="button"
          v-show="showPlayButton(index)"
          @click.stop="togglePlay(index)"
        >
          <van-icon name="play-circle-o" size="68" />
        </button>

        <button class="top-indicator" type="button" @click="toggleMute">
          <span class="channel">声音状态</span>
          <div class="volume">
            <van-icon :name="volumeIcon" size="20" />
            <span class="volume-label">
              {{ isMuted ? "静音" : "有声" }}
            </span>
          </div>
        </button>

        <div class="bottom-info">
          <div class="creator">
            <img class="avatar" :src="video.avatar" alt="" />
            <div>
              <p class="creator-name">@{{ video.author }}</p>
              <p class="prompt-label">Prompt</p>
            </div>
          </div>
          <p class="prompt-text">{{ video.prompt }}</p>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { httpGet, httpPost } from "@/utils/http";

const router = useRouter();

const featuredVideos = ref([]);
const loading = ref(false);
const page = ref(1);
const pageSize = ref(50);
const finished = ref(false);

const videoRefs = ref([]);
const itemRefs = ref([]);
const feedContainer = ref(null);
const playingStates = ref([]);
const activeIndex = ref(0);
const isMuted = ref(true);
let intersectionObserver;

const setVideoRef = (el, index) => {
  videoRefs.value[index] = el;
  if (el) {
    el.muted = isMuted.value;
  }
};

const setItemRef = (el, index) => {
  itemRefs.value[index] = el;
};

const updatePlayingState = (index, isPlaying) => {
  playingStates.value[index] = isPlaying;
  playingStates.value = [...playingStates.value];
};

const handleVideoPlay = (index) => {
  updatePlayingState(index, true);
  activeIndex.value = index;
};

const handleVideoPause = (index) => {
  updatePlayingState(index, false);
};

const showPlayButton = (index) => !playingStates.value[index];

const applyMuteState = (video) => {
  if (!video) return;
  video.muted = isMuted.value;
};

const playVideo = (index) => {
  const video = videoRefs.value[index];
  if (!video) return;
  applyMuteState(video);
  const playPromise = video.play();
  if (playPromise && typeof playPromise.then === "function") {
    playPromise.catch(() => {
      video.muted = true;
      video.play().catch(() => {});
    });
  }
};

const pauseVideo = (index) => {
  const video = videoRefs.value[index];
  if (video) {
    video.pause();
  }
};

const pauseOthers = (currentIndex) => {
  videoRefs.value.forEach((video, idx) => {
    if (idx !== currentIndex && video && !video.paused) {
      video.pause();
    }
  });
};

const setActiveVideo = (index) => {
  pauseOthers(index);
  playVideo(index);
  activeIndex.value = index;
};

const togglePlay = (index) => {
  const video = videoRefs.value[index];
  if (!video) return;
  if (video.paused) {
    setActiveVideo(index);
  } else {
    pauseVideo(index);
  }
};

const toggleMute = () => {
  isMuted.value = !isMuted.value;
  videoRefs.value.forEach((video) => {
    if (!video) return;
    applyMuteState(video);
    if (!isMuted.value && video.paused) {
      video.play().catch(() => {});
    }
  });
};

const volumeIcon = computed(() => (isMuted.value ? "volume-o" : "volume"));

const goBack = () => {
  router.back();
};

const handleIntersect = (entries) => {
  const visibleEntry = entries
    .filter((entry) => entry.isIntersecting)
    .sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];
  if (visibleEntry) {
    const index = Number(visibleEntry.target.dataset.index);
    setActiveVideo(index);
    // 记录观看
    if (featuredVideos.value[index]) {
      incrementView(featuredVideos.value[index].id);
    }
  }
};

const initObserver = () => {
  if (intersectionObserver) {
    intersectionObserver.disconnect();
  }
  intersectionObserver = new IntersectionObserver(handleIntersect, {
    threshold: 0.85,
  });
  itemRefs.value.forEach((item) => {
    if (item) {
      intersectionObserver.observe(item);
    }
  });
};

// 获取视频列表
const fetchVideos = async () => {
  if (loading.value || finished.value) return;
  
  loading.value = true;
  try {
    const res = await httpGet("/api/video-square/list", {
      page: page.value,
      page_size: pageSize.value,
    });
    
    if (res.code === 0) {
      const items = res.data.items || [];
      if (items.length === 0) {
        finished.value = true;
      } else {
        // 转换为前端需要的格式
        const videos = items.map((item) => ({
          id: item.id,
          videoUrl: item.videoUrl,
          poster: item.poster || "",
          prompt: item.prompt || "",
          author: item.author || "",
          avatar: item.avatar || "",
        }));
        
        if (page.value === 1) {
          featuredVideos.value = videos;
        } else {
          featuredVideos.value = [...featuredVideos.value, ...videos];
        }
        
        page.value++;
        
        // 如果还有更多数据，继续加载
        if (items.length < pageSize.value) {
          finished.value = true;
        }
      }
    }
  } catch (error) {
    console.error("获取视频列表失败:", error);
  } finally {
    loading.value = false;
  }
};

// 增加观看次数
const incrementView = (videoId) => {
  httpPost("/api/video-square/view", { id: videoId }).catch((err) => {
    console.error("更新观看次数失败:", err);
  });
};

onMounted(() => {
  fetchVideos().then(() => {
    nextTick(() => {
      initObserver();
      if (featuredVideos.value.length > 0) {
        setActiveVideo(0);
      }
    });
  });
});

onBeforeUnmount(() => {
  if (intersectionObserver) {
    intersectionObserver.disconnect();
  }
  videoRefs.value.forEach((video) => {
    if (video) {
      video.pause();
    }
  });
});
</script>

<style scoped>
.mobile-video-square {
  background: #05060a;
  min-height: 100vh;
  color: #fff;
  padding-top: 46px;
  overflow: hidden;
}

.nav {
  background: rgba(5, 6, 10, 0.4);
  backdrop-filter: blur(10px);
}

.feed {
  height: calc(100vh - 46px);
  overflow-y: auto;
  scroll-snap-type: y mandatory;
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.feed::-webkit-scrollbar {
  display: none;
}

.feed-item {
  position: relative;
  height: calc(100vh - 46px);
  scroll-snap-align: start;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #000;
}

.feed-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.play-overlay {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  border: none;
  background: rgba(0, 0, 0, 0.4);
  border-radius: 50%;
  width: 96px;
  height: 96px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.top-indicator {
  position: absolute;
  top: 60px;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  font-size: 18px;
  text-transform: capitalize;
  color: rgba(255, 255, 255, 0.9);
  background: none;
  border: none;
  cursor: pointer;
}

.channel {
  font-weight: 600;
}

.volume {
  display: flex;
  align-items: center;
  gap: 6px;
}

.volume-label {
  font-size: 14px;
  text-transform: none;
  color: rgba(255, 255, 255, 0.85);
}

.bottom-info {
  position: absolute;
  left: 16px;
  right: 16px;
  bottom: 40px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.creator {
  display: flex;
  align-items: center;
  gap: 12px;
}

.avatar {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.7);
}

.creator-name {
  margin: 0;
  font-weight: 600;
  font-size: 15px;
}

.prompt-label {
  margin: 0;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: rgba(255, 255, 255, 0.7);
}

.prompt-text {
  margin: 0;
  font-size: 15px;
  line-height: 1.6;
  color: #fff;
}
</style>

