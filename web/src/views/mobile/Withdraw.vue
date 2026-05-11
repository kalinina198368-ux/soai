<template>
  <div class="mobile-withdraw">
    <van-nav-bar title="提现" left-arrow @click-left="goBack" />
    
    <div class="content">
      <!-- 背景装饰 -->
      <div class="bg-decoration"></div>
      
      <!-- 可提现积分卡片 -->
      <div class="balance-card">
        <div class="balance-header">
          <h3 class="balance-title">可提现积分</h3>
          <div class="balance-amount">
            <span class="amount">{{ availablePower }}</span>
            <span class="unit">积分</span>
          </div>
          <div class="balance-tip">* 最低提现积分：{{ minWithdrawPower }} 积分</div>
        </div>
      </div>

      <!-- 提现表单 -->
      <div class="withdraw-form-card">
        <div class="section-header">
          <h3 class="section-title">
            <i class="iconfont icon-gift"></i>
            提现信息
          </h3>
        </div>
        
        <van-form @submit="handleWithdraw">
          <van-cell-group inset>
            <van-field
              v-model="withdrawForm.power"
              type="number"
              label="提现积分"
              placeholder="请输入提现积分"
              :rules="powerRules"
              required
            >
              <template #button>
                <van-button
                  size="small"
                  type="primary"
                  plain
                  @click="withdrawForm.power = availablePower"
                >
                  全部提现
                </van-button>
              </template>
            </van-field>

            <van-field
              :model-value="calculatedAmount"
              label="折合人民币"
              readonly
              disabled
            >
              <template #input>
                <span class="amount-display">¥{{ calculatedAmount }}</span>
              </template>
            </van-field>

            <van-field
              v-model="withdrawForm.accountName"
              label="收款人姓名"
              placeholder="请输入收款人姓名"
              :rules="nameRules"
              required
            />

            <van-field
              label="收款码"
              required
            >
              <template #input>
                <van-uploader
                  v-model="qrcodeFileList"
                  :after-read="handleQrcodeUpload"
                  @delete="handleQrcodeDelete"
                  :max-count="1"
                  :deletable="true"
                  accept="image/*"
                >
                  <template #upload>
                    <div class="upload-slot">
                      <van-icon name="plus" size="24" />
                      <span>上传收款码</span>
                    </div>
                  </template>
                </van-uploader>
              </template>
            </van-field>

            <van-field
              v-model="withdrawForm.remark"
              type="textarea"
              label="备注"
              placeholder="选填，备注信息"
              rows="3"
              autosize
            />
          </van-cell-group>

          <div class="form-actions">
            <van-button
              type="primary"
              block
              round
              native-type="submit"
              :loading="submitting"
              :disabled="!canWithdraw"
            >
              提交提现申请
            </van-button>
          </div>
        </van-form>
      </div>

      <!-- 提现记录 -->
      <div class="withdraw-history-card">
        <div class="section-header">
          <h3 class="section-title">
            <i class="iconfont icon-list"></i>
            提现记录
          </h3>
        </div>

        <div class="history-list">
          <van-empty
            v-if="withdrawHistory.length === 0 && !loading"
            description="暂无提现记录"
            :image-size="100"
          />
          <van-cell-group inset v-else>
            <van-cell
              v-for="record in withdrawHistory"
              :key="record.id"
              class="history-item"
            >
              <template #title>
                <div class="record-info">
                  <div class="record-amount">{{ record.power || record.amount }} 积分</div>
                  <div class="record-status" :class="getStatusClass(record.status)">
                    {{ getStatusText(record.status) }}
                  </div>
                </div>
              </template>
              <template #value>
                <div class="record-time">{{ dateFormat(record.created_at) }}</div>
              </template>
              <template #label>
                <div class="record-detail">
                  <div class="detail-item" v-if="record.account_name">
                    <span class="label">收款人：</span>
                    <span class="value">{{ record.account_name }}</span>
                  </div>
                  <div class="detail-item" v-if="record.qrcode_url">
                    <span class="label">收款码：</span>
                    <van-image
                      :src="record.qrcode_url"
                      width="60"
                      height="60"
                      fit="cover"
                      round
                      @click="previewQrcode(record.qrcode_url)"
                      style="vertical-align: middle; margin-left: 8px; cursor: pointer;"
                    />
                  </div>
                  <div class="detail-item" v-if="record.remark">
                    <span class="label">备注：</span>
                    <span class="value">{{ record.remark }}</span>
                  </div>
                </div>
              </template>
            </van-cell>
          </van-cell-group>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from "vue";
import { httpGet, httpPost } from "@/utils/http";
import { dateFormat } from "@/utils/libs";
import { showFailToast, showLoadingToast, showSuccessToast, showImagePreview } from "vant";
import { useRouter } from "vue-router";
import { checkSession } from "@/store/cache";
import Compressor from "compressorjs";

const router = useRouter();

const loading = ref(false);
const submitting = ref(false);
const availablePower = ref(0);
const minWithdrawPower = ref(100);
const qrcodeFileList = ref([]);
const withdrawForm = ref({
  power: "",
  accountName: "",
  qrcodeUrl: "",
  remark: "",
});
const withdrawHistory = ref([]);

// 计算折合人民币金额（10:1，10个积分=1元）
const calculatedAmount = computed(() => {
  const power = Number(withdrawForm.value.power);
  if (isNaN(power) || power <= 0) {
    return "0.00";
  }
  // 10个积分 = 1元人民币
  const amount = power / 10;
  return amount.toFixed(2);
});

const powerRules = [
  { required: true, message: "请输入提现积分" },
  {
    validator: (value) => {
      const power = Number(value);
      if (isNaN(value) || power <= 0) {
        return "提现积分必须大于0";
      }
      if (power < minWithdrawPower.value) {
        return `最低提现积分为${minWithdrawPower.value}`;
      }
      if (power > availablePower.value) {
        return "提现积分不能超过可提现余额";
      }
      return true;
    },
  },
];

const nameRules = [
  { required: true, message: "请输入收款人姓名" },
  { pattern: /^[\u4e00-\u9fa5]{2,10}$/, message: "请输入正确的姓名格式" },
];

const canWithdraw = computed(() => {
  return (
    withdrawForm.value.power &&
    withdrawForm.value.accountName &&
    withdrawForm.value.qrcodeUrl &&
    Number(withdrawForm.value.power) >= minWithdrawPower.value &&
    Number(withdrawForm.value.power) <= availablePower.value
  );
});

const goBack = () => {
  router.back();
};

const getStatusText = (status) => {
  const statusMap = {
    0: "待审核",
    1: "审核通过",
    2: "已打款",
    3: "已拒绝",
  };
  return statusMap[status] || "未知";
};

const getStatusClass = (status) => {
  const classMap = {
    0: "status-pending",
    1: "status-approved",
    2: "status-paid",
    3: "status-rejected",
  };
  return classMap[status] || "";
};

const previewQrcode = (url) => {
  if (url) {
    showImagePreview([url]);
  }
};

const fetchPower = () => {
  // 从用户信息中获取积分
  httpGet("/api/user/profile")
    .then((res) => {
      if (res.data) {
        availablePower.value = res.data.points || 0;
        minWithdrawPower.value = 100; // 可以从系统配置获取
      }
    })
    .catch((e) => {
      console.error("获取积分失败:", e);
      // 使用默认值
    });
};

const fetchHistory = () => {
  loading.value = true;
  httpGet("/api/withdraw/history")
    .then((res) => {
      withdrawHistory.value = res.data || [];
      loading.value = false;
    })
    .catch((e) => {
      loading.value = false;
      showFailToast("获取提现记录失败：" + e.message);
    });
};

const handleQrcodeUpload = (file) => {
  file.status = "uploading";
  
  // 压缩图片并上传
  new Compressor(file.file, {
    quality: 0.6,
    maxWidth: 800,
    maxHeight: 800,
    success(result) {
      const formData = new FormData();
      formData.append("file", result, result.name);
      
      // 执行上传操作
      httpPost("/api/upload", formData)
        .then((res) => {
          file.url = res.data.url;
          withdrawForm.value.qrcodeUrl = res.data.url;
          file.status = "done";
          showSuccessToast("收款码上传成功");
        })
        .catch((e) => {
          file.status = "failed";
          file.message = "上传失败";
          showFailToast("收款码上传失败：" + e.message);
        });
    },
    error(err) {
      file.status = "failed";
      file.message = "图片处理失败";
      showFailToast("图片处理失败：" + err.message);
    },
  });
};

const handleQrcodeDelete = () => {
  withdrawForm.value.qrcodeUrl = "";
};

const handleWithdraw = () => {
  if (!canWithdraw.value) {
    return;
  }

  submitting.value = true;
  const loadingToast = showLoadingToast({
    message: "提交中...",
    forbidClick: true,
    duration: 0,
  });

  httpPost("/api/withdraw/apply", {
    power: Number(withdrawForm.value.power),
    amount: Number(calculatedAmount.value), // 折合成人民币的金额 10:1 10个积分等于1元人民币
    account_name: withdrawForm.value.accountName,
    qrcode_url: withdrawForm.value.qrcodeUrl,
    remark: withdrawForm.value.remark,
  })
    .then(() => {
      loadingToast.close();
      submitting.value = false;
      showSuccessToast("提现申请已提交，请等待审核");
      // 重置表单
      withdrawForm.value = {
        power: "",
        accountName: "",
        qrcodeUrl: "",
        remark: "",
      };
      qrcodeFileList.value = [];
      // 刷新积分和记录
      fetchPower();
      fetchHistory();
    })
    .catch((e) => {
      loadingToast.close();
      submitting.value = false;
      showFailToast("提交失败：" + e.message);
    });
};

onMounted(() => {
  checkSession()
    .then(() => {
      fetchPower();
      fetchHistory();
    })
    .catch(() => {
      router.push("/mobile/login");
    });
});
</script>

<style lang="stylus" scoped>
.mobile-withdraw {
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

    .balance-card {
      background rgba(255, 255, 255, 0.95)
      border-radius 28px
      padding 32px
      margin-bottom 28px
      box-shadow 0 6px 24px rgba(168, 230, 207, 0.12)
      backdrop-filter blur(12px)
      border 1px solid rgba(168, 230, 207, 0.25)
      animation slideInUp 0.6s ease-out
      text-align center

      .balance-header {
        .balance-title {
          font-size 18px
          font-weight 500
          color #7a8b8b
          margin 0 0 16px 0
        }

        .balance-amount {
          display flex
          align-items baseline
          justify-content center
          gap 8px
          margin-bottom 16px

          .amount {
            font-size 48px
            font-weight 600
            color #2d5a4a
          }

          .unit {
            font-size 20px
            font-weight 500
            color #7a8b8b
            margin-left 8px
          }
        }

        .balance-tip {
          font-size 12px
          color #7a8b8b
        }
      }
    }

    .upload-slot {
      display flex
      flex-direction column
      align-items center
      justify-content center
      width 80px
      height 80px
      background #f0fdf4
      border 1px dashed rgba(168, 230, 207, 0.5)
      border-radius 8px
      color #7a8b8b
      font-size 12px
      gap 4px

      &:active {
        background #e8f9f0
      }
    }

    .withdraw-form-card {
      background rgba(255, 255, 255, 0.95)
      border-radius 24px
      padding 28px
      margin-bottom 28px
      box-shadow 0 6px 24px rgba(168, 230, 207, 0.12)
      backdrop-filter blur(12px)
      border 1px solid rgba(168, 230, 207, 0.25)
      animation slideInUp 0.6s ease-out 0.2s both

      .section-header {
        text-align center
        margin-bottom 24px

        .section-title {
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

      .amount-display {
        color #f56c6c
        font-size 18px
        font-weight 600
      }

      .form-actions {
        margin-top 24px
        padding 0 4px

        .van-button {
          height 48px
          font-size 16px
          font-weight 500
          border-radius 14px
          box-shadow 0 3px 12px rgba(168, 230, 207, 0.2)
        }
      }
    }

    .withdraw-history-card {
      background rgba(255, 255, 255, 0.95)
      border-radius 24px
      padding 28px
      box-shadow 0 6px 24px rgba(168, 230, 207, 0.12)
      backdrop-filter blur(12px)
      border 1px solid rgba(168, 230, 207, 0.25)
      animation slideInUp 0.6s ease-out 0.35s both

      .section-header {
        text-align center
        margin-bottom 24px

        .section-title {
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

      .history-list {
        .history-item {
          margin-bottom 12px
          border-radius 16px
          overflow hidden
          background #f0fdf4
          border 1px solid rgba(168, 230, 207, 0.25)

          .record-info {
            .record-amount {
              font-size 20px
              font-weight 600
              color #2d5a4a
              margin-bottom 4px
            }

            .record-status {
              display inline-block
              padding 4px 12px
              border-radius 12px
              font-size 12px
              font-weight 500

              &.status-pending {
                background #fff3cd
                color #856404
              }

              &.status-approved {
                background #d1ecf1
                color #0c5460
              }

              &.status-paid {
                background #d4edda
                color #155724
              }

              &.status-rejected {
                background #f8d7da
                color #721c24
              }
            }
          }

          .record-time {
            font-size 12px
            color #7a8b8b
          }

          .record-detail {
            margin-top 8px
            padding-top 8px
            border-top 1px solid rgba(168, 230, 207, 0.2)

            .detail-item {
              font-size 12px
              color #7a8b8b
              margin-bottom 4px

              &:last-child {
                margin-bottom 0
              }

              .label {
                font-weight 500
              }

              .value {
                color #2d5a4a
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

