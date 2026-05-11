<template>
    <div class="mobile-share">
      <van-nav-bar
        :title="isHelpMode ? '为我助力' : '分享邀请'"
        left-arrow
        @click-left="router.back()"
      />
  
      <div class="share-container">
        <!-- 助力模式：其他用户通过分享链接访问 -->
        <div v-if="isHelpMode" class="share-content">
          <div class="share-icon-wrapper">
            <van-icon name="like-o" size="80" color="#07c160" />
          </div>
          <h2 class="share-title">为我助力</h2>
          
          <!-- 已助力提示 -->
          <div v-if="alreadyHelped && helpedUsername" class="helped-notice">
            <van-icon name="success" size="24" color="#07c160" style="margin-right: 8px;" />
            <span>今天已为 <span class="username-highlight">{{ helpedUsername }}</span> 用户助力了</span>
          </div>
          
          <p v-else class="share-desc">您的好友邀请您体验新功能<br>点击助力按钮，为好友助力！</p>
          
          <van-button
            v-if="!alreadyHelped"
            type="primary"
            size="large"
            round
            class="share-button"
            @click="handleHelpClick"
          >
            <van-icon name="like-o" size="20" style="margin-right: 8px;" />
            去助力
          </van-button>

          <div class="share-tips">
            <p v-if="alreadyHelped">• 您今天已经为好友助力过了</p>
            <p v-else>• 点击助力按钮为好友助力</p>
            <p>• 助力成功后，好友将获得算力奖励</p>
            <p>• 感谢您的支持！</p>
          </div>
        </div>

        <!-- 分享模式：用户自己访问分享页面 -->
        <div v-else class="share-content">
          <div class="share-icon-wrapper">
            <van-icon name="share-o" size="80" color="#409EFF" />
          </div>
          <h2 class="share-title">邀请好友体验新功能</h2>
          <p class="share-desc">每成功邀请一位好友访问，您将获得 +2算力奖励！</p>
          
          <van-button
            type="primary"
            size="large"
            round
            class="share-button"
            @click="handleShareClick"
          >
            <van-icon name="share-o" size="20" style="margin-right: 8px;" />
            点击分享
          </van-button>
  
          <div class="share-tips">
            <p>• 点击右上角 <span class="highlight">...</span> 进行分享</p>
            <p>• 选择发送给朋友或分享到朋友圈</p>
            <p>• 好友通过您的链接访问后，算力将自动到账</p>
          </div>

          <!-- 助力日志列表 -->
          <div class="help-log-section">
            <h3 class="help-log-title">
              <van-icon name="records" size="20" style="margin-right: 8px;" />
              助力记录
              <span class="help-log-count" v-if="helpLogList.length > 0">({{ helpLogList.length }})</span>
            </h3>
            
            <div v-if="loadingHelpLog" class="help-log-loading">
              <van-loading type="spinner" size="20px">加载中...</van-loading>
            </div>
            
            <div v-else-if="helpLogList.length === 0" class="help-log-empty">
              <van-icon name="info-o" size="40" color="#ccc" />
              <p>暂无助力记录</p>
              <p class="help-log-empty-tip">分享给好友，邀请他们为您助力吧！</p>
            </div>
            
            <div v-else class="help-log-list">
              <div 
                v-for="log in helpLogList" 
                :key="log.id" 
                class="help-log-item"
              >
                <div class="help-log-item-header">
                  <div class="help-log-item-info">
                    <van-icon name="location-o" size="16" color="#999" />
                    <span class="help-log-ip">{{ log.ip }}</span>
                    <van-tag 
                      v-if="log.rewarded" 
                      type="success" 
                      size="mini"
                      style="margin-left: 8px;"
                    >
                      已奖励
                    </van-tag>
                  </div>
                  <span class="help-log-time">{{ formatTime(log.created_at) }}</span>
                </div>
                <div v-if="log.user_agent" class="help-log-user-agent">
                  {{ formatUserAgent(log.user_agent) }}
                </div>
              </div>
            </div>
            
            <van-button
              v-if="hasMoreHelpLog"
              plain
              type="primary"
              size="small"
              class="load-more-btn"
              @click="loadMoreHelpLog"
              :loading="loadingMoreHelpLog"
            >
              加载更多
            </van-button>
          </div>
        </div>
      </div>
  
      <!-- 分享引导遮罩层 -->
      <div 
        v-if="showShareOverlay" 
        class="share-overlay" 
        @click="hideShareGuide"
      >
        <div class="share-arrow"></div>
        <div class="share-guide-content">
          <h3 class="share-guide-title">点击右上角 <span class="share-dots">...</span></h3>
          <p class="share-guide-text">
            选择 <span class="share-highlight">[发送给朋友]</span> <br> 
            或 <span class="share-highlight">[分享到朋友圈]</span>
          </p>
          <div class="share-tip-box">
            好友通过您的链接访问后<br>算力将自动到账
          </div>
        </div>
      </div>
    </div>
  </template>
  
  <script setup>
  import { ref, onMounted, computed } from "vue";
  import { useRoute, useRouter } from "vue-router";
  import { httpGet } from "@/utils/http";
  import { showLoginDialog } from "@/utils/libs";
  import { checkSession } from "@/store/cache";
  
  const route = useRoute();
  const router = useRouter();
  const showShareOverlay = ref(false);
  
  // 检查是否登录
  const isLogin = ref(false);
  
  // 助力状态
  const alreadyHelped = ref(false);
  const helpedUsername = ref("");
  
  // 助力日志
  const helpLogList = ref([]);
  const loadingHelpLog = ref(false);
  const loadingMoreHelpLog = ref(false);
  const helpLogPage = ref(1);
  const helpLogPageSize = ref(20);
  const hasMoreHelpLog = ref(false);
  
  // 判断是否是助力模式（其他用户通过分享链接访问）
  const isHelpMode = computed(() => {
    const referral = route.query.referral;
    if (!referral) {
      return false; // 没有 referral 参数，是用户自己访问分享页面
    }
    // 检查是否是用户自己点击分享跳转过来的
    const shouldShowGuide = sessionStorage.getItem('showShareGuide') === 'true';
    // 如果有 referral 且不是用户自己，则是助力模式
    return !shouldShowGuide;
  });
  
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
  
  // 处理分享按钮点击
  const handleShareClick = () => {
    if (!isLogin.value) {
      return showLoginDialog(router);
    }
    
    // 显示分享引导遮罩层
    showShareOverlay.value = true;
  };
  
  // 处理助力按钮点击
  const handleHelpClick = () => {
    if (!isLogin.value) {
      return showLoginDialog(router);
    }
    
    // 如果已登录，重新调用API记录助力（如果之前没有记录成功）
    const referral = route.query.referral;
    if (referral && !alreadyHelped.value) {
      httpGet(`/api/invite/shareClick?code=${referral}`)
        .then((res) => {
          if (res.data) {
            if (res.data.rewarded) {
              // 助力成功
              alreadyHelped.value = true;
              if (res.data.username) {
                helpedUsername.value = res.data.username;
              }
              console.log("助力成功！");
            } else if (res.data.alreadyHelped) {
              // 已经助力过
              alreadyHelped.value = true;
              if (res.data.username) {
                helpedUsername.value = res.data.username;
              }
            }
          }
        })
        .catch((e) => {
          console.error("助力失败:", e);
        });
    }
  };
  
  // 隐藏分享引导遮罩层
  const hideShareGuide = () => {
    showShareOverlay.value = false;
  };
  
  // 格式化时间
  const formatTime = (timestamp) => {
    const date = new Date(timestamp * 1000);
    const now = new Date();
    const diff = now - date;
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);
    
    if (minutes < 1) {
      return "刚刚";
    } else if (minutes < 60) {
      return `${minutes}分钟前`;
    } else if (hours < 24) {
      return `${hours}小时前`;
    } else if (days < 7) {
      return `${days}天前`;
    } else {
      const month = date.getMonth() + 1;
      const day = date.getDate();
      const hour = date.getHours();
      const minute = date.getMinutes();
      return `${month}-${day} ${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')}`;
    }
  };
  
  // 格式化UserAgent
  const formatUserAgent = (userAgent) => {
    if (!userAgent) return "";
    // 简化显示，只显示主要信息
    if (userAgent.includes("Mobile")) {
      return "移动设备";
    } else if (userAgent.includes("Windows")) {
      return "Windows";
    } else if (userAgent.includes("Mac")) {
      return "Mac";
    } else if (userAgent.includes("Linux")) {
      return "Linux";
    }
    return "其他设备";
  };
  
  // 加载助力日志
  const loadHelpLog = (page = 1, append = false) => {
    if (!isLogin.value) return;
    
    if (append) {
      loadingMoreHelpLog.value = true;
    } else {
      loadingHelpLog.value = true;
    }
    
    httpGet(`/api/invite/shareClickLogList?page=${page}&page_size=${helpLogPageSize.value}`)
      .then((res) => {
        if (res.data && res.data.items) {
          if (append) {
            helpLogList.value = [...helpLogList.value, ...res.data.items];
          } else {
            helpLogList.value = res.data.items;
          }
          // 判断是否还有更多
          const total = res.data.total || 0;
          const currentTotal = helpLogList.value.length;
          hasMoreHelpLog.value = currentTotal < total;
          helpLogPage.value = page;
        }
      })
      .catch((e) => {
        console.error("加载助力日志失败:", e);
      })
      .finally(() => {
        loadingHelpLog.value = false;
        loadingMoreHelpLog.value = false;
      });
  };
  
  // 加载更多助力日志
  const loadMoreHelpLog = () => {
    if (!loadingMoreHelpLog.value && hasMoreHelpLog.value) {
      loadHelpLog(helpLogPage.value + 1, true);
    }
  };
  
  // 初始化
  onMounted(() => {
    // 检查登录状态
    checkSession()
      .then(() => {
        isLogin.value = true;
        // 如果不是助力模式，加载助力日志
        if (!isHelpMode.value) {
          loadHelpLog();
        }
      })
      .catch(() => {
        isLogin.value = false;
      });
    
    const referral = route.query.referral;
    
    // 如果有 referral 参数，设置分享 meta 标签
    if (referral) {
      // 使用当前页面的完整 URL（包含 referral 参数）
      const shareUrl = `${location.protocol}//${location.host}${location.pathname}${location.search}`;
      const shareTitle = '邀请好友体验新功能 - So-AI';
      const shareDesc = '每成功邀请一位好友访问，您将获得 +2算力奖励！';
      // 使用绝对路径，确保微信能正确抓取图片（建议图片尺寸：1200x630px）
      const shareImage = `${location.protocol}//${location.host}/images/logo.png`;
      
      // 设置微信分享 meta 标签
      setWeChatShareMeta(shareTitle, shareDesc, shareImage, shareUrl);
      
      // 检查是否是用户自己点击分享跳转过来的（通过 sessionStorage 标记）
      const shouldShowGuide = sessionStorage.getItem('showShareGuide') === 'true';
      if (shouldShowGuide) {
        // 清除标记，下次访问时不再显示
        sessionStorage.removeItem('showShareGuide');
        // 如果是用户自己跳转的，不记录分享点击，只显示分享引导
        // 这里不自动显示遮罩层，等用户点击分享按钮时再显示
      } else {
        // 防止重复调用：使用 sessionStorage 标记已调用过的邀请码
        const shareClickKey = `shareClick_${referral}`;
        const hasClicked = sessionStorage.getItem(shareClickKey);
        
        if (!hasClicked) {
          // 标记已调用，防止重复请求
          sessionStorage.setItem(shareClickKey, 'true');
          
          // 如果是他人通过分享链接访问的，调用API记录分享点击
          // 无论是否登录都调用，因为助力不需要登录也能看到
          httpGet(`/api/invite/shareClick?code=${referral}`)
            .then((res) => {
              if (res.data) {
                // 检查是否已经助力过
                if (res.data.alreadyHelped) {
                  alreadyHelped.value = true;
                  if (res.data.username) {
                    helpedUsername.value = res.data.username;
                  }
                }
                
                if (res.data.rewarded) {
                  console.log("分享点击已记录，分享者获得1算力奖励");
                  // 如果本次助力成功，也标记为已助力
                  alreadyHelped.value = true;
                  if (res.data.username) {
                    helpedUsername.value = res.data.username;
                  }
                } else {
                  console.log("分享点击已记录（24小时内已奖励过）");
                }
              }
            })
            .catch((e) => {
              // 如果请求失败，清除标记，允许重试
              sessionStorage.removeItem(shareClickKey);
              console.error("记录分享点击失败:", e);
            });
        } else {
          console.log("分享点击已记录（本次会话已处理）");
        }
      }
    } else {
      // 如果没有 referral 参数，设置默认的分享 meta 标签
      const shareUrl = `${location.protocol}//${location.host}${location.pathname}`;
      const shareTitle = '邀请好友体验新功能 - So-AI';
      const shareDesc = '每成功邀请一位好友访问，您将获得+2算力奖励！';
      const shareImage = `${location.protocol}//${location.host}/images/logo.png`;
      setWeChatShareMeta(shareTitle, shareDesc, shareImage, shareUrl);
    }
  });
  </script>
  
  <style lang="stylus">
  .mobile-share {
    min-height: 100vh;
    background: #f5f5f5;
  
    .share-container {
      padding: 40px 20px;
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: calc(100vh - 46px);
  
      .share-content {
        text-align: center;
        max-width: 400px;
        width: 100%;
  
        .share-icon-wrapper {
          margin-bottom: 32px;
          display: flex;
          justify-content: center;
        }
  
        .share-title {
          font-size: 24px;
          font-weight: 600;
          color: #333;
          margin: 0 0 16px 0;
        }
  
        .share-desc {
          font-size: 16px;
          color: #666;
          line-height: 1.6;
          margin: 0 0 40px 0;
        }
  
        .share-button {
          width: 200px;
          height: 50px;
          font-size: 18px;
          margin-bottom: 40px;
        }
  
        .share-tips {
          text-align: left;
          background: #fff;
          border-radius: 12px;
          padding: 20px;
          font-size: 14px;
          color: #666;
          line-height: 2;

          p {
            margin: 8px 0;
          }

          .highlight {
            color: #409EFF;
            font-weight: 600;
            font-size: 18px;
          }
        }

        .helped-notice {
          display: flex;
          align-items: center;
          justify-content: center;
          background: #f0f9ff;
          border: 1px solid #07c160;
          border-radius: 12px;
          padding: 16px 20px;
          margin: 0 0 40px 0;
          font-size: 16px;
          color: #333;
          line-height: 1.6;

          .username-highlight {
            color: #07c160;
            font-weight: 600;
          }
        }

        .help-log-section {
          margin-top: 40px;
          background: #fff;
          border-radius: 12px;
          padding: 20px;
          text-align: left;

          .help-log-title {
            display: flex;
            align-items: center;
            font-size: 18px;
            font-weight: 600;
            color: #333;
            margin: 0 0 20px 0;
            padding-bottom: 12px;
            border-bottom: 1px solid #eee;

            .help-log-count {
              color: #409EFF;
              font-weight: normal;
              margin-left: 4px;
            }
          }

          .help-log-loading {
            text-align: center;
            padding: 40px 0;
            color: #999;
          }

          .help-log-empty {
            text-align: center;
            padding: 40px 20px;
            color: #999;

            p {
              margin: 12px 0;
              font-size: 14px;
            }

            .help-log-empty-tip {
              font-size: 12px;
              color: #ccc;
            }
          }

          .help-log-list {
            .help-log-item {
              padding: 16px 0;
              border-bottom: 1px solid #f5f5f5;

              &:last-child {
                border-bottom: none;
              }

              .help-log-item-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 8px;

                .help-log-item-info {
                  display: flex;
                  align-items: center;
                  font-size: 14px;
                  color: #666;

                  .help-log-ip {
                    margin-left: 6px;
                    font-family: monospace;
                  }
                }

                .help-log-time {
                  font-size: 12px;
                  color: #999;
                }
              }

              .help-log-user-agent {
                font-size: 12px;
                color: #999;
                margin-top: 4px;
              }
            }
          }

          .load-more-btn {
            width: 100%;
            margin-top: 16px;
          }
        }
      }
    }
  }
  
  .share-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.85);
    z-index: 9999;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding-top: 96px;
    animation: fadeIn 0.3s ease-out;
  
    .share-arrow {
      position: absolute;
      top: 10px;
      right: 20px;
      width: 60px;
      height: 60px;
      background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="white"><path d="M15 5l-1.41 1.41L18.17 11H2v2h16.17l-4.59 4.59L15 19l7-7-7-7z" transform="rotate(-45 12 12)"/></svg>');
      background-repeat: no-repeat;
      background-size: contain;
      animation: floatArrow 1.5s infinite ease-in-out;
    }
  
    .share-guide-content {
      text-align: center;
      color: white;
      padding: 0 32px;
      max-width: 320px;
  
      .share-guide-title {
        font-size: 20px;
        font-weight: 600;
        margin: 0 0 16px 0;
        color: white;
  
        .share-dots {
          font-size: 28px;
          display: inline-block;
        }
      }
  
      .share-guide-text {
        font-size: 18px;
        line-height: 1.6;
        margin: 0 0 32px 0;
        color: white;
  
        .share-highlight {
          color: #07c160;
          font-weight: 600;
        }
      }
  
      .share-tip-box {
        background: rgba(255, 255, 255, 0.1);
        border-radius: 12px;
        padding: 16px;
        font-size: 14px;
        color: rgba(255, 255, 255, 0.9);
        line-height: 1.6;
      }
    }
  }
  
  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  
  @keyframes floatArrow {
    0%, 100% {
      transform: translateY(0) rotate(10deg);
    }
    50% {
      transform: translateY(-10px) rotate(10deg);
    }
  }
  </style>
  