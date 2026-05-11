<template>
  <div class="mobile-team">
    <van-nav-bar title="" left-arrow @click-left="goBack" />
    
    <div class="content">
      <!-- 背景装饰 -->
      <div class="bg-decoration"></div>
      
      <!-- 团队统计卡片 -->
      <div class="stats-card">
        <div class="stats-header">
          <h3 class="stats-title">
            <i class="iconfont icon-chart"></i>
            团队统计
          </h3>
        </div>
        <div class="stats-grid">
          <div class="stat-item highlight">
            <div class="stat-icon"><i class="iconfont icon-user-add"></i></div>
            <div class="stat-value accent">{{ teamStats.directMembers }}</div>
            <div class="stat-label">直推</div>
          </div>
          <div class="stat-item">
            <div class="stat-icon"><i class="iconfont icon-users"></i></div>
            <div class="stat-value">{{ teamStats.indirectMembers }}</div>
            <div class="stat-label">间推</div>
          </div>

          <!-- <div class="stat-item">
            <div class="stat-icon"><i class="iconfont icon-lightning"></i></div>
            <div class="stat-value">{{ teamStats.activatedMembers }}</div>
            <div class="stat-label">激活</div>
          </div> -->

          <div class="stat-item">
            <div class="stat-icon"><i class="iconfont icon-team"></i></div>
            <div class="stat-value">{{ teamStats.totalMembers }}</div>
            <div class="stat-label">团队</div>
          </div>
        </div>

        
      </div>

      <!-- 团队层级 -->
      <div class="team-section">
        <div class="section-header">
          <h3 class="section-title">
            <i class="iconfont icon-users"></i>
            团队层级
          </h3>
          <div class="section-subtitle">查看您的团队结构</div>
        </div>
        
        <div class="level-tabs">
          <van-tabs v-model:active="activeTab" @change="onTabChange">
            <van-tab title="直推团队" name="direct">
              <div class="team-list">
                <van-empty
                  v-if="directTeam.length === 0 && !loading"
                  description="暂无直推成员"
                  :image-size="100"
                />
                <van-cell-group inset v-else>
                  <van-cell
                    v-for="member in directTeam"
                    :key="member.id"
                    class="team-member"
                  >
                    <template #icon>
                      <div class="member-avatar">
                        <van-image
                          :src="member.avatar || '/images/user-info.png'"
                          width="50"
                          height="50"
                          fit="cover"
                          round
                        />
                      </div>
                    </template>
                    <template #title>
                      <div class="member-info">
                        <div class="member-name">{{ member.nickname || member.username }}</div>
                        <!-- <div class="member-id">ID: {{ member.username }}</div> -->
                      </div>
                    </template>
                    <template #value>
                      <div class="member-stats">
                        <div class="stat-badge">
                          <span class="stat-label">算力</span>
                          <span class="stat-value">{{ member.power || 0 }}</span>
                        </div>
                      </div>
                    </template>
                    <template #label>
                      <div class="member-meta">
                        <span class="join-time">加入时间: {{ dateFormat(member.created_at) }}</span>
                      </div>
                    </template>
                  </van-cell>
                </van-cell-group>
              </div>
            </van-tab>
            <van-tab title="间推团队" name="indirect">
              <div class="team-list">
                <van-empty
                  v-if="indirectTeam.length === 0 && !loading"
                  description="暂无间推成员"
                  :image-size="100"
                />
                <van-cell-group inset v-else>
                  <van-cell
                    v-for="member in indirectTeam"
                    :key="member.id"
                    class="team-member"
                  >
                    <template #icon>
                      <div class="member-avatar">
                        <van-image
                          :src="member.avatar || '/images/user-info.png'"
                          width="50"
                          height="50"
                          fit="cover"
                          round
                        />
                      </div>
                    </template>
                    <template #title>
                      <div class="member-info">
                        <div class="member-name">{{ member.nickname || member.username }}</div>
                        <!-- <div class="member-id">ID: {{ member.username }}</div> -->
                      </div>
                    </template>
                    <template #value>
                      <div class="member-stats">
                        <div class="stat-badge">
                          <span class="stat-label">算力</span>
                          <span class="stat-value">{{ member.power || 0 }}</span>
                        </div>
                      </div>
                    </template>
                    <template #label>
                      <div class="member-meta">
                        <span class="join-time">加入时间: {{ dateFormat(member.created_at) }}</span>
                      </div>
                    </template>
                  </van-cell>
                </van-cell-group>
              </div>
            </van-tab>
          </van-tabs>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { httpGet } from "@/utils/http";
import { dateFormat } from "@/utils/libs";
import { showFailToast, showLoadingToast } from "vant";
import { useRouter } from "vue-router";
import { checkSession } from "@/store/cache";

const router = useRouter();

const activeTab = ref("direct");
const loading = ref(false);
const teamStats = ref({
  totalMembers: 0,
  directMembers: 0,
  indirectMembers: 0,
});
const directTeam = ref([]);
const indirectTeam = ref([]);

const goBack = () => {
  router.back();
};

const onTabChange = (name) => {
  if (name === "direct" && directTeam.value.length === 0) {
    fetchDirectTeam();
  } else if (name === "indirect" && indirectTeam.value.length === 0) {
    fetchIndirectTeam();
  }
};

const fetchTeamStats = () => {
  httpGet("/api/team/stats")
    .then((res) => {
      if (res.data) {
        teamStats.value = {
          totalMembers: res.data.total || 0,
          directMembers: res.data.direct || 0,
          indirectMembers: res.data.indirect || 0,
          activatedMembers: res.data.jhNum || 0,//激活
        };
      }
    })
    .catch((e) => {
      console.error("获取团队统计失败:", e);
      // 使用默认值
    });
};

const fetchDirectTeam = () => {
  loading.value = true;
  httpGet("/api/team/direct")
    .then((res) => {
      directTeam.value = res.data || [];
      loading.value = false;
    })
    .catch((e) => {
      loading.value = false;
      showFailToast("获取直推团队失败：" + e.message);
    });
};

const fetchIndirectTeam = () => {
  loading.value = true;
  httpGet("/api/team/indirect")
    .then((res) => {
      indirectTeam.value = res.data || [];
      loading.value = false;
    })
    .catch((e) => {
      loading.value = false;
      showFailToast("获取间推团队失败：" + e.message);
    });
};

onMounted(() => {
  checkSession()
    .then(() => {
      fetchTeamStats();
      fetchDirectTeam();
    })
    .catch(() => {
      router.push("/mobile/login");
    });
});
</script>

<style lang="stylus" scoped>
.mobile-team {
  min-height 100vh
  background #f5f5f5
  position relative
  overflow hidden

  .bg-decoration {
    position absolute
    top 0
    left 0
    right 0
    height 200px
    background rgba(168, 230, 207, 0.08)
    border-radius 0 0 60px 60px
    z-index 0
  }

  .content {
    position relative
    z-index 1
    padding 24px 20px 100px
    max-width 480px
    margin 0 auto

    .stats-card {
      background rgba(255, 255, 255, 0.95)
      border-radius 28px
      padding 32px
      margin-bottom 28px
      box-shadow 0 6px 24px rgba(168, 230, 207, 0.12)
      backdrop-filter blur(12px)
      border 1px solid rgba(168, 230, 207, 0.25)
      animation slideInUp 0.6s ease-out

      .stats-header {
        text-align center
        margin-bottom 24px

        .stats-title {
          font-size 26px
          font-weight 600
          color #2d5a4a
          margin 0
          display flex
          align-items center
          justify-content center
          gap 12px

          .iconfont {
            font-size 28px
            color #7fcdcd
          }
        }
      }

      .stats-grid {
        display grid
        grid-template-columns repeat(3, 3fr)
        gap 16px

        .stat-item {
          text-align center
          padding 20px
          background #f0fdf4
          border-radius 16px
          border 1px solid rgba(168, 230, 207, 0.25)

          .stat-value {
            font-size 28px
            font-weight 600
            color #2d5a4a
            margin-bottom 8px
          }

          .stat-label {
            font-size 13px
            color #7a8b8b
            font-weight 500
          }
        }
      }
    }

    .team-section {
      animation slideInUp 0.6s ease-out 0.2s both

      .section-header {
        text-align center
        margin-bottom 24px

        .section-title {
          font-size 26px
          font-weight 600
          color #2d5a4a
          margin 0 0 10px 0
          display flex
          align-items center
          justify-content center
          gap 12px

          .iconfont {
            font-size 28px
            color #7fcdcd
          }
        }

        .section-subtitle {
          font-size 16px
          color #7a8b8b
        }
      }

      .level-tabs {
        background rgba(255, 255, 255, 0.95)
        border-radius 24px
        padding 20px
        box-shadow 0 6px 24px rgba(168, 230, 207, 0.12)
        backdrop-filter blur(12px)
        border 1px solid rgba(168, 230, 207, 0.25)

        :deep(.van-tabs__nav) {
          background transparent
        }

        :deep(.van-tab) {
          color #7a8b8b
          font-weight 500
        }

        :deep(.van-tab--active) {
          color #2d5a4a
          font-weight 600
        }

        :deep(.van-tabs__line) {
          background #7fcdcd
        }

        .team-list {
          margin-top 20px

          .team-member {
            margin-bottom 12px
            border-radius 16px
            overflow hidden
            background #f0fdf4
            border 1px solid rgba(168, 230, 207, 0.25)

            .member-avatar {
              margin-right 12px
            }

            .member-info {
              .member-name {
                font-size 16px
                font-weight 600
                color #2d5a4a
                margin-bottom 4px
              }

              .member-id {
                font-size 12px
                color #7a8b8b
              }
            }

            .member-stats {
              .stat-badge {
                display flex
                flex-direction column
                align-items flex-end
                gap 4px

                .stat-label {
                  font-size 11px
                  color #7a8b8b
                }

                .stat-value {
                  font-size 18px
                  font-weight 600
                  color #2d5a4a
                }
              }
            }

            .member-meta {
              margin-top 8px
              padding-top 8px
              border-top 1px solid rgba(168, 230, 207, 0.2)

              .join-time {
                font-size 12px
                color #7a8b8b
              }
            }
          }
        }
      }
    }
  }
}

@keyframes slideInUp {
  from {
    opacity 0
    transform translateY(30px)
  }
  to {
    opacity 1
    transform translateY(0)
  }
}
</style>

