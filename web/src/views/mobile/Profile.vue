<template>
  <div class="mobile-user-profile">
    <!-- 背景装饰 -->
    <div class="bg-decoration"></div>
    
    <div class="content">
      <!-- 用户信息卡片 -->
      <div class="user-card" v-if="isLogin">
        <div class="avatar-section">
          <div class="avatar-container" @click="triggerAvatarUpload">
            <van-image :src="fileList[0].url" size="90" width="90" fit="cover" round />
            <div class="avatar-ring"></div>
            <div class="avatar-upload-overlay">
              <van-icon name="photograph" size="24" color="#fff" />
            </div>
            <input 
              ref="avatarInputRef" 
              type="file" 
              accept="image/*" 
              style="display: none;" 
              @change="handleAvatarUpload"
            />
          </div>
          <div class="user-info">
            <h2 class="username">{{  form.username }}</h2>
            <!-- <p class="user-id">ID: {{ form.username }}</p> -->
            <div class="user-level" :class="`level-${form.lev || 0}`">
              <span class="level-text">{{ getLevelText(form.lev || 0) }}</span>
            </div>
          </div>
        </div> 
        
        <div class="stats-section">
          <div class="stat-item">
            <div class="stat-value">{{ form.power }}</div>
            <div class="stat-label-wrapper">
              <div class="stat-label">算力</div>
              <van-icon 
                name="info-o" 
                size="18" 
                color="#409EFF" 
                class="stat-icon"
                @click.stop="handlePowerDetailClick"
              />
              <van-button 
                size="mini" 
                type="primary" 
                class="recharge-btn"
                @click.stop="showRechargeDialog = true"
              >
                充值
              </van-button>
            </div>
          </div>

          <div class="stat-divider"></div>
          <div class="stat-item" >
            <div class="stat-value">{{ form.points}}</div>
            <div class="stat-label-wrapper">
              <div class="stat-label">积分</div>
              <van-icon 
                name="info-o" 
                size="18" 
                color="#409EFF" 
                class="stat-icon"
                @click.stop="handlePointsDetailClick"
              />
              <van-button 
                size="mini" 
                type="primary" 
                class="withdraw-btn"
                @click.stop="handleWithdrawClick"
              >
                提现
              </van-button>
            </div>
          </div>
        </div>

        <!-- 我的订单列表（仅在存在订单时展示，避免空状态显得突兀） -->
        <!-- <div class="order-section" v-if="isLogin && orderList.length">
          <div class="section-header">
            <h3 class="section-title">
              <i class="iconfont icon-order"></i>
              我的订单
            </h3>
            <span class="section-subtitle">最近 3 笔充值记录</span>
          </div>
          <div class="order-list">
            <div v-for="item in orderList" :key="item.id" class="order-item">
              <div class="order-left">
                <div class="order-title">{{ item.subject || '充值套餐' }}</div>
                <div class="order-meta">
                  <span class="order-time">{{ formatOrderTime(item.created_at || item.pay_time) }}</span>
                  <span class="order-id">单号：{{ item.order_no || item.orderNo }}</span>
                </div>
              </div>
              <div class="order-right">
                <div class="order-amount">￥{{ Number(item.amount || item.total_amount).toFixed(2) }}</div>
                <div
                  class="order-status"
                  :class="getOrderStatusInfo(item.status).type"
                >
                  {{ getOrderStatusInfo(item.status).text }}
                </div>
              </div>
            </div>
          </div>
        </div> -->
      </div>

      <!-- 快速操作菜单 -->
      <div class="quick-actions" v-if="isLogin">
        <van-grid :column-num="4" :border="false" :gutter="16" :clickable="true">
          
    

          <!--明细报表 -->
          <!-- <van-grid-item @click="router.push('/mobile/powerlog')" class="grid-item">
            <template #icon>
              <van-icon name="balance-list-o" size="28" color="#409EFF" />
            </template>
            <template #text><span class="grid-text">明细</span></template>
          </van-grid-item> -->



           <van-grid-item @click="router.push('/mobile/team')" class="grid-item">
            <template #icon>
              <van-icon name="friends-o" size="28" color="#FF9800" />
            </template>
            <template #text><span class="grid-text">团队</span></template>
          </van-grid-item>

          <!-- 我的订单 -->
          <van-grid-item @click="handleOrderClick" class="grid-item">
            <template #icon>
              <van-icon name="orders-o" size="28" color="#409EFF" />
            </template>
            <template #text><span class="grid-text">订单</span></template>
          </van-grid-item>


<!-- 
          <van-grid-item @click="showSettings = true" class="grid-item">
            <template #icon>
              <van-icon name="setting" size="28" color="#909399" />
            </template>
            <template #text><span class="grid-text">设置</span></template>
          </van-grid-item> -->


          <van-grid-item @click="handleTransferClick" class="grid-item">
            <template #icon>
              <van-icon name="exchange" size="28" color="#67C23A" />
            </template>
            <template #text><span class="grid-text">转账</span></template>
          </van-grid-item>
<!-- 
          <van-grid-item @click="showServiceDialog = true" class="grid-item">
            <template #icon>
              <van-icon name="chat-o" size="28" color="#409EFF" />
            </template>
            <template #text><span class="grid-text">客服</span></template>
          </van-grid-item> -->

                
          <van-grid-item @click="showPasswordDialog = true" class="grid-item">
            <template #icon>
              <van-icon name="setting" size="28" color="#409EFF" />
            </template>
            <template #text><span class="grid-text">设置</span></template>
          </van-grid-item>

  

          <!-- 分享 -->
          <van-grid-item @click="share" class="grid-item">
            <template #icon>
              <van-icon name="share-o" size="28" color="#409EFF" />
            </template>
            <template #text><span class="grid-text">分享</span></template>
          </van-grid-item>

          <!-- 客服 -->
          <van-grid-item @click="showServiceDialog = true" class="grid-item">
            <template #icon>
              <van-icon name="chat-o" size="28" color="#67C23A" />
            </template>
            <template #text><span class="grid-text">客服</span></template>
          </van-grid-item>

                  <!-- 帮助 -->
         <van-grid-item @click="router.push('/mobile/help')" class="grid-item">
            <template #icon>
              <van-icon name="question-o" size="28" color="#409EFF" />
            </template>
            <template #text><span class="grid-text">帮助</span></template>
          </van-grid-item>



          <van-grid-item @click="logout" class="grid-item">
            <template #icon>
              <van-icon name="close" size="28" color="#606266" />
            </template>
            <template #text><span class="grid-text">退出</span></template>
          </van-grid-item>
        </van-grid>
      </div>

      <!-- 推荐码区域 -->
      <div class="invite-section" v-if="isLogin">
        <div class="section-header">
          <h3 class="section-title">
            <i class="iconfont icon-gift"></i>
            我的推荐码
          </h3>
          <div class="section-subtitle">邀请好友一起使用</div>
        </div>
        
        <div class="invite-card">
          <div class="invite-qrcode">
            <van-image 
              :src="qrImg" 
              width="200" 
              height="200" 
              fit="cover"
              class="qrcode-image"
            />
            <p class="qrcode-tip">扫码注册</p>
          </div>
          
          <div class="invite-url-box">
            <div class="invite-url">
              <span>{{ inviteURL }}</span>
              <i 
                class="iconfont icon-copy copy-icon" 
                @click="copyInviteURL"
                title="复制链接"
              ></i>
            </div>
          </div>
        </div>
      </div>

      <!-- 我的上传码 -->
      <div class="invite-section" v-if="isLogin">
        <div class="section-header">
          <h3 class="section-title">
            <i class="iconfont icon-image"></i>
            我的上传码
          </h3>
          <div class="section-subtitle">分享给朋友扫码即可匿名上传图片</div>
        </div>

        <div class="invite-card">
          <div class="invite-qrcode">
            <van-image
              :src="uploadQrImg"
              width="200"
              height="200"
              fit="cover"
              class="qrcode-image"
            />
            <p class="qrcode-tip">扫码上传</p>
          </div>

          <div class="invite-url-box">
            <div class="invite-url">
              <span>{{ uploadURL }}</span>
              <i
                class="iconfont icon-copy copy-icon"
                @click="copyUploadURL"
                title="复制链接"
              ></i>
            </div>
          </div>
        </div>
      </div>

      <!-- 邀请统计区域 -->
      <!-- <div class="invite-stats-section" v-if="isLogin">
        <div class="section-header">
          <h3 class="section-title">
            <i class="iconfont icon-chart"></i>
            邀请统计
          </h3>
        </div>
        
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-value">{{ inviteStats.regNum }}</div>
            <div class="stat-label">直推量</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ inviteStats.jhNum }}</div>
            <div class="stat-label">激活量</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ inviteStats.hits }}</div>
            <div class="stat-label">团队数量</div>
          </div>
        </div>
      </div> -->




    </div>

    <van-dialog
      v-model:show="showPasswordDialog"
      title="修改密码"
      show-cancel-button
      @confirm="updatePass"
      @cancel="showPasswordDialog = false"
      :before-close="beforeClose"
    >
      <van-form>
        <van-cell-group inset>
          <van-field v-model="pass.old" placeholder="旧密码" />
          <van-field v-model="pass.new" type="password" placeholder="新密码" />
          <van-field v-model="pass.renew" type="password" placeholder="确认密码" />
        </van-cell-group>
      </van-form>
    </van-dialog>

    <van-action-sheet v-model:show="showSettings" title="用户设置">
      <div class="setting-content">
        <van-form>
          <van-cell-group inset>
            <van-field name="switch" label="暗黑主题">
              <template #input>
                <van-switch v-model="dark" @change="(val) => store.setTheme(val ? 'dark' : 'light')" />
              </template>
            </van-field>

            <van-field name="switch" label="流式输出">
              <template #input>
                <van-switch v-model="stream" @change="(val) => store.setChatStream(val)" />
              </template>
            </van-field>
            <!--            <van-field-->
            <!--                v-model="password"-->
            <!--                type="password"-->
            <!--                name="密码"-->
            <!--                label="密码"-->
            <!--                placeholder="密码"-->
            <!--                :rules="[{ required: true, message: '请填写密码' }]"-->
            <!--            />-->
          </van-cell-group>
          <!--          <div style="margin: 16px;">-->
          <!--            <van-button round block type="primary" native-type="submit">-->
          <!--              提交-->
          <!--            </van-button>-->
          <!--          </div>-->
        </van-form>
      </div>
    </van-action-sheet>

    <!-- 联系客服弹窗 -->
    <van-dialog
      v-model:show="showServiceDialog"
      title="联系客服"
      show-cancel-button
      @cancel="showServiceDialog = false"
      :before-close="beforeClose"
    >
      <div class="service-content">
        <div class="service-section">
          <h4 class="section-title">
            <i class="iconfont icon-phone"></i>
            客服联系方式
          </h4>
          <div class="contact-list">
            <div class="contact-item">
              <div class="contact-label">客服微信</div>
              <div class="contact-value">
                <span>{{ serviceInfo.wechat }}</span>
                <van-button 
                  size="mini" 
                  type="primary" 
                  @click="copyToClipboard(serviceInfo.wechat)"
                  class="copy-btn"
                >
                  复制
                </van-button>
              </div>
            </div>
            <!-- <div class="contact-item">
              <div class="contact-label">客服QQ</div>
              <div class="contact-value">
                <span>{{ serviceInfo.qq }}</span>
                <van-button 
                  size="mini" 
                  type="primary" 
                  @click="copyToClipboard(serviceInfo.qq)"
                  class="copy-btn"
                >
                  复制
                </van-button>
              </div>
            </div> -->

            <!-- <div class="contact-item">
              <div class="contact-label">客服电话</div>
              <div class="contact-value">
                <span>{{ serviceInfo.phone }}</span>
                <van-button 
                  size="mini" 
                  type="primary" 
                  @click="callPhone(serviceInfo.phone)"
                  class="copy-btn"
                >
                  拨打
                </van-button>
              </div>
            </div> -->


          </div>
        </div>

        <div class="service-section">
          <h4 class="section-title">
            <i class="iconfont icon-qrcode"></i>
            充值未到账请联系客服
          </h4>
          <div class="qrcode-container">
            <van-image 
              :src="serviceInfo.qrcode" 
              width="200" 
              height="200" 
              fit="cover"
              class="qrcode-image"
            />
            <p class="qrcode-tip">扫描二维码添加客服微信</p>
          </div>
        </div>

        <!-- <div class="service-section">
          <h4 class="section-title">
            <i class="iconfont icon-clock"></i>
            服务时间
          </h4>
          <div class="service-time">
            <p>工作日：9:00 - 18:00</p>
            <p>周末：10:00 - 17:00</p>
            <p class="time-tip">* 非服务时间请留言，我们会尽快回复</p>
          </div>
        </div> -->
      </div>
    </van-dialog>

    <van-dialog
      v-model:show="cashierDialog"
      title="收银台"
      :show-confirm-button="false"
      :show-cancel-button="false"
      :close-on-click-overlay="false"
      class="cashier-dialog"
    >
      <div class="cashier-content">
        <van-loading
          v-if="cashierLoading"
          type="spinner"
          size="24px"
          text="正在生成订单..."
          vertical
        />
        <div v-else>
          <div class="cashier-summary">
            <div class="summary-row">
              <span class="label">套餐：</span>
              <span class="value">{{ cashierProductName }}</span>
            </div>
            <div class="summary-row">
              <span class="label">支付方式：</span>
              <span class="value">{{ cashierPayDisplayName }}</span>
            </div>
            <div class="summary-row">
              <span class="label">订单号：</span>
              <span
                class="value order-no"
                :class="{ disabled: !cashierOrder.orderNo }"
                @click="copyOrderNo"
              >
                {{ cashierOrder.orderNo || "生成中..." }}
              </span>
            </div>
            <div class="summary-row">
              <span class="label">应付金额：</span>
              <span class="value amount">￥{{ cashierAmountText }}</span>
            </div>
          </div>

          <div v-if="isWechatPay" class="cashier-pay">
            <div v-if="cashierQRCode" class="qrcode-wrapper">
              <van-image :src="cashierQRCode" width="220" height="220" fit="cover" />
              <p class="cashier-tip">请使用微信扫描二维码完成支付</p>
            </div>
            <van-loading v-else type="spinner" size="24px" text="二维码生成中..." vertical />
            <van-button block type="primary" plain class="mt-3" @click="refreshCashierOrder">
              二维码失效？点击刷新
            </van-button>
          </div>
          <div v-else class="cashier-pay">
            <p class="cashier-tip">点击下方按钮前往第三方收银台完成支付</p>
            <van-button block type="primary" :disabled="!cashierPayUrl" @click="handleOpenPayLink">
              立即前往支付
            </van-button>
          </div>

          <div class="cashier-actions">
            <van-button block type="success" class="mt-3" @click="handleCashierSuccess">我已完成支付</van-button>
            <van-button block type="default" plain class="mt-2" @click="handleCashierClose">稍后再付</van-button>
          </div>
        </div>
      </div>
    </van-dialog>


    <!-- 转账操作 -->
    
  <van-dialog v-model:show="transferDialog" title="转账" show-cancel-button
              @confirm="updateTransferBefore"
              @cancel="transferDialog = false "
              :before-close="beforeClose">
    <van-form>
      <van-cell-group inset>
        <van-field
            v-model="transfer.uid"
            placeholder="收款人推荐码"
        />
        <van-field
            v-model="transfer.power"
            placeholder="算力"
        />
     
      </van-cell-group>
    </van-form>
  </van-dialog>

  
  <!-- 密碼确认框 -->
  <van-dialog v-model:show="passwordDialog" title="输入密码" show-cancel-button
              @confirm="updateTransfer"
              @cancel="passwordDialog = false">
    <van-field
        v-model="transfer.password"
        type="password"
        placeholder="请输入密码"
    />
  </van-dialog>

  <van-dialog
      v-model:show="mockPayDialog"
      title="收银台（模拟）"
      :show-cancel-button="true"
      cancel-button-text="取消"
      confirm-button-text="我已完成支付"
      @confirm="confirmMockPaid"
    >
      <div class="cashier-content">
        <div class="cashier-summary">
          <div class="summary-row">
            <span class="label">套餐：</span>
            <span class="value">{{ mockOrder.subject }}</span>
          </div>
          <div class="summary-row">
            <span class="label">订单号：</span>
            <span class="value order-no">{{ mockOrder.orderNo }}</span>
          </div>
          <div class="summary-row">
            <span class="label">应付金额：</span>
            <span class="value amount">￥{{ Number(mockOrder.amount || 0).toFixed(2) }}</span>
          </div>
        </div>
        <p class="cashier-tip">这是模拟收银台，不会真实扣款。点击“我已完成支付”将视为支付成功。</p>
      </div>
    </van-dialog>

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

    <!-- 充值套餐弹窗 -->
    <van-popup
      v-model:show="showRechargeDialog"
      position="center"
      :style="{ width: '90%', maxWidth: '500px', borderRadius: '20px', overflow: 'hidden' }"
      closeable
      close-icon-position="top-right"
      :close-on-click-overlay="true"
      class="recharge-popup-wrapper"
    >
      <div class="recharge-popup">
        <div class="section-header">
          <h3 class="section-title">充值套餐</h3>
          <div class="section-subtitle">选择适合您的套餐</div>
        </div>
        
        <div class="product-grid">
          <div 
            class="product-card" 
            :class="{ 'selected': selectedProduct?.id === item.id }"
            v-for="item in products" 
            :key="item.id"
            @click="handleProductSelect(item)"
          >
            <div class="product-check" v-if="selectedProduct?.id === item.id">
              <van-icon name="success" size="16" color="#fff" />
            </div>
            <div class="product-amount">￥{{ item.discount }}</div>
            <div class="product-label">售价</div>
          </div>
          
          <!-- 手动输入金额 -->
          <div 
            class="product-card custom-amount-card"
            :class="{ 'selected': customAmount && customAmount > 0 }"
            @click.stop
          >
            <div class="product-check" v-if="customAmount && customAmount > 0">
              <van-icon name="success" size="16" color="#fff" />
            </div>
            <div class="custom-amount-input-wrapper">
              <span class="currency-symbol">￥</span>
              <van-field
                v-model="customAmount"
                type="number"
                placeholder="请输入"
                :min="1"
                :max="9999"
                class="custom-amount-input"
                @input="handleCustomAmountInput"
                @blur="handleCustomAmountBlur"
                @click.stop
              />
            </div>
            <div class="product-label">自定义</div>
          </div>
        </div>
        
        <!-- 余额和到账信息 -->
        <div class="balance-info">
          <div class="balance-item">
            <span class="balance-label">算力余额:</span>
            <div class="balance-value-wrapper">
              <van-icon name="balance-o" size="18" color="#333" />
              <span class="balance-value">{{ form.power }}</span>
            </div>
          </div>
          <div class="balance-item balance-item-right" v-if="selectedProduct || customAmount">
            <span class="balance-label">到账算力:</span>
            <span class="balance-value credit">{{ calculatedPower }}</span>
          </div>
        </div>
        
        <!-- 全局去支付按钮 -->
        <div class="global-pay-section">
          <van-button 
            type="primary" 
            size="large" 
            round
            block
            class="global-pay-btn"
            :disabled="!selectedProduct && !customAmount"
            @click="handleGlobalPay"
          >
            <van-icon name="credit-pay" size="18" color="#fff" />
            <span v-if="selectedProduct">￥{{ selectedProduct.discount }} 确认下单</span>
            <span v-else-if="customAmount">￥{{ customAmount }} 确认下单</span>
            <span v-else>请选择套餐或输入金额</span>
          </van-button>
        </div>
      </div>
    </van-popup>

    <!-- 我的订单列表弹窗 -->
    <van-popup
      v-model:show="showOrderDialog"
      position="bottom"
      :style="{ height: '80%', borderRadius: '20px 20px 0 0' }"
      closeable
      close-icon-position="top-right"
      :close-on-click-overlay="true"
      class="order-popup-wrapper"
    >
      <div class="order-popup">
        <div class="order-popup-header">
          <h3 class="order-popup-title">
            <i class="iconfont icon-order"></i>
            我的订单
          </h3>
          <span class="order-popup-subtitle">充值记录</span>
        </div>
        
        <div class="order-popup-content">
          <van-empty 
            v-if="!orderListLoading && orderList.length === 0" 
            description="暂无订单记录"
            image-size="100"
          />
          <div v-else class="order-list-full">
            <div v-for="item in orderList" :key="item.id" class="order-item">
              <div class="order-left">
                <div class="order-title">{{ item.subject || '充值套餐' }}</div>
                <div class="order-meta">
                  <span class="order-time">{{ formatOrderTime(item.created_at || item.pay_time) }}</span>
                  <span class="order-id">单号：{{ item.order_no || item.orderNo }}</span>
                </div>
              </div>
              <div class="order-right">
                <div class="order-amount">￥{{ Number(item.amount || item.total_amount).toFixed(2) }}</div>
                <div
                  class="order-status"
                  :class="getOrderStatusInfo(item.status).type"
                >
                  {{ getOrderStatusInfo(item.status).text }}
                </div>
              </div>
            </div>
          </div>
          <van-loading v-if="orderListLoading" type="spinner" vertical>加载中...</van-loading>
        </div>
      </div>
    </van-popup>

  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { showFailToast, showLoadingToast, showNotify, showSuccessToast } from "vant";
import { httpGet, httpPost } from "@/utils/http";
import { dateFormat, showLoginDialog } from "@/utils/libs";
import { ElMessage } from "element-plus";
import { checkSession, getSystemInfo } from "@/store/cache";
import { useRouter } from "vue-router";
import { removeUserToken } from "@/store/session";
import { useSharedStore } from "@/store/sharedata";
import QRCode from "qrcode";
import Compressor from "compressorjs";

// 联诚收银台支付配置（真实环境）
const LCSW_BASE_URL = "https://pay.lcsw.cn/lcsw/open/wap/110/pay";
const LCSW_MERCHANT_NO = "856704816000012";
const LCSW_TERMINAL_ID = "19013255";
//const LCSW_NOTIFY_URL = "http://z56ae4b5.natappfree.cc/api/payment/notify/wechat";
const LCSW_NOTIFY_URL = "http://sozm.ai/api/payment/notify/wechat";
// TODO: 将此处替换为联诚支付后台分配的 access_token 密钥
const LCSW_ACCESS_TOKEN = "48bfefa440cf4faa9a95012b2cc583a9";

// 简单 MD5 实现（用于生成 key_sign）
function md5cycle(x, k) {
  let [a, b, c, d] = x;
  function ff(a, b, c, d, x, s, t) {
    a = a + ((b & c) | (~b & d)) + x + t;
    return (((a << s) | (a >>> (32 - s))) + b) | 0;
  }
  function gg(a, b, c, d, x, s, t) {
    a = a + ((b & d) | (c & ~d)) + x + t;
    return (((a << s) | (a >>> (32 - s))) + b) | 0;
  }
  function hh(a, b, c, d, x, s, t) {
    a = a + (b ^ c ^ d) + x + t;
    return (((a << s) | (a >>> (32 - s))) + b) | 0;
  }
  function ii(a, b, c, d, x, s, t) {
    a = a + (c ^ (b | ~d)) + x + t;
    return (((a << s) | (a >>> (32 - s))) + b) | 0;
  }

  a = ff(a, b, c, d, k[0], 7, -680876936);
  d = ff(d, a, b, c, k[1], 12, -389564586);
  c = ff(c, d, a, b, k[2], 17, 606105819);
  b = ff(b, c, d, a, k[3], 22, -1044525330);
  a = ff(a, b, c, d, k[4], 7, -176418897);
  d = ff(d, a, b, c, k[5], 12, 1200080426);
  c = ff(c, d, a, b, k[6], 17, -1473231341);
  b = ff(b, c, d, a, k[7], 22, -45705983);
  a = ff(a, b, c, d, k[8], 7, 1770035416);
  d = ff(d, a, b, c, k[9], 12, -1958414417);
  c = ff(c, d, a, b, k[10], 17, -42063);
  b = ff(b, c, d, a, k[11], 22, -1990404162);
  a = ff(a, b, c, d, k[12], 7, 1804603682);
  d = ff(d, a, b, c, k[13], 12, -40341101);
  c = ff(c, d, a, b, k[14], 17, -1502002290);
  b = ff(b, c, d, a, k[15], 22, 1236535329);

  a = gg(a, b, c, d, k[1], 5, -165796510);
  d = gg(d, a, b, c, k[6], 9, -1069501632);
  c = gg(c, d, a, b, k[11], 14, 643717713);
  b = gg(b, c, d, a, k[0], 20, -373897302);
  a = gg(a, b, c, d, k[5], 5, -701558691);
  d = gg(d, a, b, c, k[10], 9, 38016083);
  c = gg(c, d, a, b, k[15], 14, -660478335);
  b = gg(b, c, d, a, k[4], 20, -405537848);
  a = gg(a, b, c, d, k[9], 5, 568446438);
  d = gg(d, a, b, c, k[14], 9, -1019803690);
  c = gg(c, d, a, b, k[3], 14, -187363961);
  b = gg(b, c, d, a, k[8], 20, 1163531501);
  a = gg(a, b, c, d, k[13], 5, -1444681467);
  d = gg(d, a, b, c, k[2], 9, -51403784);
  c = gg(c, d, a, b, k[7], 14, 1735328473);
  b = gg(b, c, d, a, k[12], 20, -1926607734);

  a = hh(a, b, c, d, k[5], 4, -378558);
  d = hh(d, a, b, c, k[8], 11, -2022574463);
  c = hh(c, d, a, b, k[11], 16, 1839030562);
  b = hh(b, c, d, a, k[14], 23, -35309556);
  a = hh(a, b, c, d, k[1], 4, -1530992060);
  d = hh(d, a, b, c, k[4], 11, 1272893353);
  c = hh(c, d, a, b, k[7], 16, -155497632);
  b = hh(b, c, d, a, k[10], 23, -1094730640);
  a = hh(a, b, c, d, k[13], 4, 681279174);
  d = hh(d, a, b, c, k[0], 11, -358537222);
  c = hh(c, d, a, b, k[3], 16, -722521979);
  b = hh(b, c, d, a, k[6], 23, 76029189);
  a = hh(a, b, c, d, k[9], 4, -640364487);
  d = hh(d, a, b, c, k[12], 11, -421815835);
  c = hh(c, d, a, b, k[15], 16, 530742520);
  b = hh(b, c, d, a, k[2], 23, -995338651);

  a = ii(a, b, c, d, k[0], 6, -198630844);
  d = ii(d, a, b, c, k[7], 10, 1126891415);
  c = ii(c, d, a, b, k[14], 15, -1416354905);
  b = ii(b, c, d, a, k[5], 21, -57434055);
  a = ii(a, b, c, d, k[12], 6, 1700485571);
  d = ii(d, a, b, c, k[3], 10, -1894986606);
  c = ii(c, d, a, b, k[10], 15, -1051523);
  b = ii(b, c, d, a, k[1], 21, -2054922799);
  a = ii(a, b, c, d, k[8], 6, 1873313359);
  d = ii(d, a, b, c, k[15], 10, -30611744);
  c = ii(c, d, a, b, k[6], 15, -1560198380);
  b = ii(b, c, d, a, k[13], 21, 1309151649);
  a = ii(a, b, c, d, k[4], 6, -145523070);
  d = ii(d, a, b, c, k[11], 10, -1120210379);
  c = ii(c, d, a, b, k[2], 15, 718787259);
  b = ii(b, c, d, a, k[9], 21, -343485551);

  x[0] = (x[0] + a) | 0;
  x[1] = (x[1] + b) | 0;
  x[2] = (x[2] + c) | 0;
  x[3] = (x[3] + d) | 0;
}

function md5blk(s) {
  const md5blks = [];
  for (let i = 0; i < 64; i += 4) {
    md5blks[i >> 2] =
      s.charCodeAt(i) +
      (s.charCodeAt(i + 1) << 8) +
      (s.charCodeAt(i + 2) << 16) +
      (s.charCodeAt(i + 3) << 24);
  }
  return md5blks;
}

function md51(s) {
  const n = s.length;
  const state = [1732584193, -271733879, -1732584194, 271733878];
  let i;
  for (i = 64; i <= n; i += 64) {
    md5cycle(state, md5blk(s.substring(i - 64, i)));
  }
  s = s.substring(i - 64);
  const tail = new Array(16).fill(0);
  for (i = 0; i < s.length; i++) {
    tail[i >> 2] |= s.charCodeAt(i) << ((i % 4) << 3);
  }
  tail[i >> 2] |= 0x80 << ((i % 4) << 3);
  if (i > 55) {
    md5cycle(state, tail);
    for (i = 0; i < 16; i++) tail[i] = 0;
  }
  tail[14] = n * 8;
  md5cycle(state, tail);
  return state;
}

function rhex(n) {
  const hexChr = "0123456789abcdef";
  let s = "";
  for (let j = 0; j < 4; j++) {
    s +=
      hexChr.charAt((n >> (j * 8 + 4)) & 0x0f) +
      hexChr.charAt((n >> (j * 8)) & 0x0f);
  }
  return s;
}

function hex(x) {
  for (let i = 0; i < x.length; i++) {
    x[i] = rhex(x[i]);
  }
  return x.join("");
}

function md5(str) {
  return hex(md51(str));
}
const form = ref({
  username: "",
  nickname: "",
  mobile: "",
  avatar: "",
  power: 0,
  lev: 0,
});
const fileList = ref([
  {
    url: "/images/user-info.png",
    message: "上传中...",
  },
]);

const products = ref([]);
const vipMonthPower = ref(0);
const payWays = ref({});
const router = useRouter();
const userId = ref(0);
const isLogin = ref(false);
const showSettings = ref(false);
const showServiceDialog = ref(false);
const showRechargeDialog = ref(false);
const showOrderDialog = ref(false);
const orderListLoading = ref(false);
const selectedProduct = ref(null);
const customAmount = ref(null);
const customAmountCheckTimer = ref(null); // 防抖定时器
const store = useSharedStore();
const stream = ref(store.chatStream);
const dark = ref(store.theme === "dark");
const avatarInputRef = ref(null);

// 推荐码相关
const inviteURL = ref("");
const qrImg = ref("");
const inviteCode = ref("ABC123"); // 硬编码推荐码

// 上传码相关（匿名扫码上传图片）
const uploadURL = ref("");
const uploadQrImg = ref("");
const uploadCode = ref("");

// 邀请统计数据
const inviteStats = ref({
  regNum: 0,    // 直推量
  jhNum: 0,     // 激活量
  hits: 0       // 团队数量
});

// 我的订单列表（当前登录用户，已支付订单）
const orderList = ref([]);

// 订单状态显示文案 & 类型
const getOrderStatusInfo = (status) => {
  const s = Number(status);
  if (s === 2) {
    return { text: "已支付", type: "success" };
  }
  if (s === 1) {
    return { text: "待支付", type: "pending" }; // 已扫码待支付
  }
  if (s === 0) {
    return { text: "未支付", type: "pending" };
  }
  return { text: "未知状态", type: "default" };
};

// 格式化订单时间
const formatOrderTime = (timestamp) => {
  if (!timestamp) return "";
  try {
    return dateFormat(new Date(timestamp * 1000), "YYYY-MM-DD HH:mm");
  } catch (e) {
    return "";
  }
};

const cashierDialog = ref(false);
const cashierLoading = ref(false);
const cashierQRCode = ref("");
const cashierPayUrl = ref("");
const cashierPayWay = ref(null);
const cashierProduct = ref(null);
const cashierOrder = ref({
  orderNo: "",
  amount: 0,
  subject: "",
});

const mockPayDialog = ref(false)
const mockOrder = ref({ orderNo: '', amount: 0, subject: '', productId: 0 })

// 是否启用模拟支付（已切换为真实收银台环境，这里默认关闭）
const useMockPay = ref(false)

const isWechatPay = computed(() => {
  if (!cashierPayWay.value) {
    return false;
  }
  return (
    cashierPayWay.value.pay_way === "wechat" ||
    cashierPayWay.value.pay_type === "wechat"
  );
});

const cashierAmountText = computed(() => {
  const amount = Number(cashierOrder.value.amount || 0);
  return amount.toFixed(2);
});

const cashierProductName = computed(() => {
  if (cashierOrder.value.subject) {
    return cashierOrder.value.subject;
  }
  return cashierProduct.value?.name || "";
});

const cashierPayDisplayName = computed(() => {
  if (!cashierPayWay.value) {
    return "";
  }
  const payType = (cashierPayWay.value.pay_type || cashierPayWay.value.pay_way || "").toString().toLowerCase();
  if (payType.includes("alipay")) {
    return "支付宝";
  }
  if (payType.includes("wechat")) {
    return "微信支付";
  }
  return cashierPayWay.value.name || cashierPayWay.value.title || cashierPayWay.value.pay_type || cashierPayWay.value.pay_way || "";
});

// 计算到账算力
const calculatedPower = computed(() => {
  if (selectedProduct.value) {
    return selectedProduct.value.power > 0 ? selectedProduct.value.power : vipMonthPower.value;
  }
  if (customAmount.value && !isNaN(Number(customAmount.value))) {
    // 算力和金额比例为10:1，10个算力等于1rmb
    return Number(customAmount.value) * 10;
  }
  return 0;
});

const generateOrderNo = () => {
  const now = new Date();
  const pad = (value) => value.toString().padStart(2, "0");
  const randomSuffix = Math.floor(Math.random() * 1000).toString().padStart(3, "0");
  return `SO${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}${pad(
    now.getHours()
  )}${pad(now.getMinutes())}${pad(now.getSeconds())}${randomSuffix}`;
};

// 获取等级文字
const getLevelText = (lev) => {
  const levelMap = {
    0: "普通会员",
    1: "一星会员",
    2: "二星会员",
    3: "三星会员",
    7: "青铜代理",
    8: "白银代理",
    9: "黄金代理",
  };
  return levelMap[lev] || "普通会员";
};

// 生成联诚收银台所需的时间字符串：YYYYMMDDHHmmss
const buildTerminalTime = () => {
  const now = new Date();
  const pad = (v) => v.toString().padStart(2, "0");
  return (
    now.getFullYear().toString() +
    pad(now.getMonth() + 1) +
    pad(now.getDate()) +
    pad(now.getHours()) +
    pad(now.getMinutes()) +
    pad(now.getSeconds())
  );
};

// 构建联诚收银台支付 URL（包含签名）
const buildLcswPayUrl = (product) => {
  const amountYuan = Number(product?.discount ?? product?.price ?? 0);
  const totalFee = Math.round(amountYuan * 100); // 单位：分
  const terminalTrace = generateOrderNo(); // 作为当前系统订单号
  const params = {
    merchant_no: LCSW_MERCHANT_NO,
    terminal_id: LCSW_TERMINAL_ID,
    terminal_trace: terminalTrace,
    terminal_time: buildTerminalTime(),
    total_fee: String(totalFee),
    notify_url: LCSW_NOTIFY_URL,
  };

  const keys = Object.keys(params).sort();
  const string1 = keys.map((k) => `${k}=${params[k]}`).join("&");
  const signStr = `${string1}&access_token=${LCSW_ACCESS_TOKEN}`;
  const keySign = md5(signStr).toUpperCase();

  // 最终 URL：按照文档要求使用字典序参数 + key_sign
  const finalQuery = `${string1}&key_sign=${keySign}`;
  const payUrl = `${LCSW_BASE_URL}?${finalQuery}`;

  return {
    url: payUrl,
    orderNo: terminalTrace,
    amount: amountYuan,
    subject: product?.name || product?.title || "套餐支付",
  };
};

// 打开联诚收银台：先让后台生成订单，再跳转到联诚收银台链接
const openLcswCashier = (product) => {
  if (!isLogin.value) {
    return showLoginDialog(router);
  }
  if (!LCSW_ACCESS_TOKEN || LCSW_ACCESS_TOKEN.startsWith("请替换")) {
    showFailToast("支付配置不完整，请先在前端代码中配置 access_token");
    return;
  }

  // 1. 先在前端生成联诚收银台所需的参数（包括 terminal_trace 作为本系统订单号）
  const payInfo = buildLcswPayUrl(product);

  // 2. 将订单信息先提交到后台，由后台生成/落库订单
  const nowTs = Math.floor(Date.now() / 1000);
  const payload = {
    order_no: payInfo.orderNo,
    amount: payInfo.amount,
    subject: payInfo.subject,
    user_id: userId.value,
    product_id: Number(product?.id || 0),
    pay_time: nowTs,
    pay_method: "lcsw", // 标记为联诚收银台
    status: "unpaid",   // 后台可按需使用
  };

  httpPost("/api/order/mockPaid", payload)
    .then(() => {
      // 3. 后台订单创建成功后，再跳转到联诚收银台链接
      window.location.href = payInfo.url;
    })
    .catch((e) => {
      showFailToast("创建订单失败：" + e.message);
    });
};

const fetchUserProfile = () => {
  httpGet("/api/user/profile")
    .then((res) => {
      if (res.data) {
        form.value = res.data;
        if (res.data.avatar) {
          fileList.value[0].url = res.data.avatar;
        }
      }
    })
    .catch((e) => {
      console.log(e.message);
      showFailToast("获取用户信息失败");
    });
};

onMounted(() => {
  checkSession()
    .then((user) => {
      userId.value = user.id;
      isLogin.value = true;
      fetchUserProfile();
      
      // 用户登录后，获取邀请码并设置分享链接的 meta 标签
      // 这样当用户在微信中分享当前页面时，会使用正确的分享链接
      httpGet("/api/invite/code")
        .then((res) => {
          if (res.data && res.data.code) {
            const shareLink = `${location.protocol}//${location.host}/mobile/home?referral=${res.data.code}`;
            const shareTitle = '邀请好友体验新功能 - So-AI';
            const shareDesc = '每成功邀请一位好友访问，您将获得 +50 积分奖励！';
            const shareImage = `${location.protocol}//${location.host}/images/logo.png`;
            // 设置 og:url 为分享链接，这样微信分享时会使用这个链接
            setWeChatShareMeta(shareTitle, shareDesc, shareImage, shareLink);
          }
        })
        .catch((e) => {
          console.error("获取邀请码失败:", e);
        });
    })
    .catch(() => {
      // 未登录：提示并跳转到登录页
      showFailToast("请先登录");
      router.replace({
        path: "/mobile/login",
        query: { redirect: "/mobile/profile" },
      });
    });

  // 获取产品列表
  httpGet("/api/product/list")
    .then((res) => {
      products.value = res.data;
    })
    .catch((e) => {
      showFailToast("获取产品套餐失败：" + e.message);
    });

  getSystemInfo()
    .then((res) => {
      vipMonthPower.value = res.data["vip_month_power"];
      // 获取客服信息
      if (res.data["wechat_card_url"]) {
        serviceInfo.value.qrcode = res.data["wechat_card_url"];
      }
    })
    .catch((e) => {
      showFailToast("获取系统配置失败：" + e.message);
    });

  httpGet("/api/payment/payWays")
    .then((res) => {
      if (Array.isArray(res.data)) {
        payWays.value = res.data;
      } else if (res.data && typeof res.data === "object") {
        payWays.value = Object.values(res.data);
      } else {
        payWays.value = [];
      }
    })
    .catch((e) => {
      ElMessage.error("获取支付方式失败：" + e.message);
    });

  // 生成推荐码二维码（硬编码）

   httpGet("/api/invite/code")
        .then((res) => {
          const text = `${location.protocol}//${location.host}/mobile/reg?invite_code=${res.data.code}`;

          // 更新邀请统计数据
          if (res.data) {
            inviteStats.value.regNum = res.data["reg_num"] || res.data.regNum || 0;  // 直推量
            inviteStats.value.jhNum = res.data["jh_num"] || res.data.jhNum || res.data["activate_num"] || res.data.activateNum || 0;  // 激活量
            inviteStats.value.hits = res.data["hits"] || res.data.hits || res.data["team_num"] || res.data.teamNum || 0;  // 团队数量
          }
        
          QRCode.toDataURL(text, { width: 400, height: 400, margin: 2 }, (error, url) => {
            if (error) {
              console.error(error);
            } else {
              qrImg.value = url;
            }
          });
          inviteURL.value = text;
        })
        .catch((e) => {
          console.error("获取邀请码信息失败:", e);
        });

  // 获取上传码并生成二维码（用于匿名扫码上传图片）
  httpGet("/api/upload_code/code")
    .then((res) => {
      if (!res.data || !res.data.code) {
        return;
      }
      uploadCode.value = res.data.code;
      const text = `${location.protocol}//${location.host}/mobile/upload?code=${res.data.code}`;
      QRCode.toDataURL(text, { width: 400, height: 400, margin: 2 }, (error, url) => {
        if (error) {
          console.error(error);
        } else {
          uploadQrImg.value = url;
        }
      });
      uploadURL.value = text;
    })
    .catch((e) => {
      console.error("获取上传码失败:", e);
    });

  // 获取当前用户订单列表（只取前 3 条已支付订单，用于首页展示）
  httpGet("/api/order/list", { page: 1, page_size: 3 })
    .then((res) => {
      const data = res && res.data ? res.data : null;

      // 兼容多种后端返回结构：
      // 1) { code, data: { items: [...] } }
      // 2) { code, data: { list: [...] } }
      // 3) { code, data: [...] }
      if (data && Array.isArray(data.items)) {
        orderList.value = data.items;
      } else if (data && Array.isArray(data.list)) {
        orderList.value = data.list;
      } else if (Array.isArray(data)) {
        orderList.value = data;
      } else {
        orderList.value = [];
      }
    })
    .catch((e) => {
      console.error("获取订单列表失败:", e);
    });

  // const inviteLink = `${location.protocol}//${location.host}/mobile/register?invite_code=${inviteCode.value}`;
  // inviteURL.value = inviteLink;


  
  // QRCode.toDataURL(inviteLink, {width: 400, height: 400, margin: 2}, (error, url) => {
  //   if (error) {
  //     console.error("生成二维码失败:", error);
  //   } else {
  //     qrImg.value = url;
  //   }
  //   });
});

// 获取完整订单列表（用于订单弹窗）
const fetchOrderList = async () => {
  orderListLoading.value = true;
  try {
    const res = await httpGet("/api/order/list", { page: 1, page_size: 20 });
    const data = res && res.data ? res.data : null;

    // 兼容多种后端返回结构
    if (data && Array.isArray(data.items)) {
      orderList.value = data.items;
    } else if (data && Array.isArray(data.list)) {
      orderList.value = data.list;
    } else if (Array.isArray(data)) {
      orderList.value = data;
    } else {
      orderList.value = [];
    }
  } catch (e) {
    console.error("获取订单列表失败:", e);
    showFailToast("获取订单列表失败");
    orderList.value = [];
  } finally {
    orderListLoading.value = false;
  }
};

// 点击订单按钮
const handleOrderClick = () => {
  showOrderDialog.value = true;
  fetchOrderList();
};

// 监听充值弹窗，打开时重置选中状态
watch(showRechargeDialog, (visible) => {
  if (visible) {
    selectedProduct.value = null;
    customAmount.value = null;
  }
});

// 监听自定义金额输入，实时验证并提示
watch(customAmount, (newValue, oldValue) => {
  // 清除之前的防抖定时器
  if (customAmountCheckTimer.value) {
    clearTimeout(customAmountCheckTimer.value);
    customAmountCheckTimer.value = null;
  }
  
  // 如果输入为空，直接返回
  if (!newValue || newValue === '' || newValue === null) {
    return;
  }
  
  const numValue = Number(newValue);
  
  // 如果已经是修正后的值（1或9999），避免无限循环
  if (numValue === 1 || numValue === 9999) {
    return;
  }
  
  // 如果小于1，使用防抖延迟提示
  if (isNaN(numValue) || numValue < 1) {
    customAmountCheckTimer.value = setTimeout(() => {
      if (customAmount.value && (isNaN(Number(customAmount.value)) || Number(customAmount.value) < 1)) {
        showFailToast("输入金额不能小于1元");
        customAmount.value = 1;
      }
      customAmountCheckTimer.value = null;
    }, 500);
    return;
  }
  
  // 如果大于9999，立即提示并修正
  if (numValue > 9999) {
    showFailToast("输入金额不能超过9999元");
    customAmount.value = 9999;
  }
});

watch(cashierDialog, (visible) => {
  if (!visible) {
    cashierLoading.value = false;
    cashierQRCode.value = "";
    cashierPayUrl.value = "";
    cashierPayWay.value = null;
    cashierProduct.value = null;
    cashierOrder.value = {
      orderNo: "",
      amount: 0,
      subject: "",
    };
  }
});

// 触发头像上传
const triggerAvatarUpload = () => {
  if (avatarInputRef.value) {
    avatarInputRef.value.click();
  }
};

// 处理头像上传
const handleAvatarUpload = (event) => {
  const file = event.target.files[0];
  if (!file) return;

  if (!file.type.startsWith('image/')) {
    showFailToast('请选择图片文件');
    return;
  }

  showLoadingToast({
    message: '上传中...',
    forbidClick: true,
  });

  // 压缩图片并上传
  new Compressor(file, {
    quality: 0.6,
    maxWidth: 500,
    maxHeight: 500,
    success(result) {
      const formData = new FormData();
      formData.append('file', result, result.name);
      
      // 执行上传操作
      httpPost('/api/upload', formData).then((res) => {
        form.value.avatar = res.data.url;
        fileList.value[0].url = res.data.url;
        
        // 更新用户信息
        httpPost('/api/user/profile/update', form.value).then(() => {
          showSuccessToast('头像上传成功');
        }).catch(() => {
          showFailToast('更新头像失败');
        });
      }).catch((e) => {
        showFailToast('上传失败：' + e.message);
      });
    },
    error(err) {
      console.error('图片压缩失败:', err);
      showFailToast('图片处理失败');
    },
  });

  // 清空input，以便可以重复选择同一文件
  if (avatarInputRef.value) {
    avatarInputRef.value.value = '';
  }
};



// 客服信息
const serviceInfo = ref({
  wechat: "soai_001",
  qq: "3822457531",
  phone: "400-123-4567",
  qrcode: "/images/service-qrcode.png"
});

const beforeClose = (action) => {
  return new Promise((resolve) => {
    resolve(action === "confirm");
  });
};

// 提交修改密码
const updatePass = () => {
  if (pass.value.old === "") {
    return showNotify({ type: "danger", message: "请输入旧密码" });
  }
  if (!pass.value.new || pass.value.new.length < 8) {
    return showNotify({ type: "danger", message: "密码的长度为8-16个字符" });
  }
  if (pass.value.renew !== pass.value.new) {
    return showNotify({ type: "danger", message: "两次输入密码不一致" });
  }
  httpPost("/api/user/password", {
    old_pass: pass.value.old,
    password: pass.value.new,
    repass: pass.value.renew,
  })
    .then(() => {
      showSuccessToast("更新成功！");
      showPasswordDialog.value = false;
    })
    .catch((e) => {
      showFailToast("更新失败，" + e.message);
      showPasswordDialog.value = false;
    });
};

const pay = (product, payWay) => {
  if (!isLogin.value) {
    return showLoginDialog(router);
  }

  // 根据支付方式决定是否走联诚真实收银台
  const payType = (payWay?.pay_type || payWay?.pay_way || "").toString().toLowerCase();
  if (payType.includes("wechat") || payType.includes("wxpay")) {
    // 微信相关支付，直接跳转到联诚收银台（真实环境）
    openLcswCashier(product);
    return;
  }


  cashierProduct.value = product;
  cashierPayWay.value = payWay;
  const payAmount = Number(product.discount ?? product.price ?? 0);
  const subject = product.name || product.title || "套餐支付";
  cashierOrder.value = {
    orderNo: generateOrderNo(),
    amount: payAmount,
    subject,
  };
  cashierDialog.value = true;
  createPaymentOrder();
};

// 处理产品选择
const handleProductSelect = (product) => {
  selectedProduct.value = product;
  customAmount.value = null; // 清除手动输入
};

// 处理手动输入金额（清除产品选择）
const handleCustomAmountInput = (value) => {
  // 清除产品选择
  if (value && value !== '' && value !== null) {
    selectedProduct.value = null;
  }
};

// 处理输入框失焦，确保范围在1-9999（只修正，不提示，因为输入时已提示）
const handleCustomAmountBlur = () => {
  // 清除防抖定时器
  if (customAmountCheckTimer.value) {
    clearTimeout(customAmountCheckTimer.value);
    customAmountCheckTimer.value = null;
  }
  
  if (!customAmount.value) {
    return;
  }
  const numValue = Number(customAmount.value);
  if (isNaN(numValue) || numValue < 1) {
    customAmount.value = 1;
  } else if (numValue > 9999) {
    customAmount.value = 9999;
  }
};

// 全局去支付按钮处理
const handleGlobalPay = () => {
  if (!selectedProduct.value && !customAmount.value) {
    showFailToast("请先选择套餐或输入金额");
    return;
  }

  if (!isLogin.value) {
    return showLoginDialog(router);
  }

  // 创建支付产品对象
  let productToPay;
  if (selectedProduct.value) {
    productToPay = selectedProduct.value;
  } else if (customAmount.value) {
    // 手动输入时创建虚拟产品对象
    productToPay = {
      id: 'custom',
      name: '自定义充值',
      title: '自定义充值',
      price: Number(customAmount.value),
      discount: Number(customAmount.value),
      power: calculatedPower.value
    };
  }

  // 如果有多个支付方式，使用第一个；如果没有，使用模拟支付
  if (Array.isArray(payWays.value) && payWays.value.length > 0) {
    // 优先使用支付宝，如果没有则使用第一个
    const alipayWay = payWays.value.find(way => {
      const payType = (way.pay_type || way.pay_way || "").toString().toLowerCase();
      return payType.includes("alipay");
    });
    const payWay = alipayWay || payWays.value[0];
    pay(productToPay, payWay);
  } else {
    // 没有支付方式，使用模拟支付
    openMockCashier(productToPay);
  }
};

const createPaymentOrder = () => {
  if (!cashierProduct.value || !cashierPayWay.value) {
    return;
  }
  cashierLoading.value = true;
  cashierQRCode.value = "";
  cashierPayUrl.value = "";
  const loadingToast = showLoadingToast({
    message: "正在生成订单...",
    forbidClick: true,
    duration: 0,
  });

  let host = process.env.VUE_APP_API_HOST;
  if (!host) {
    host = `${location.protocol}//${location.host}`;
  }

  httpPost(`${process.env.VUE_APP_API_HOST}/api/payment/doPay`, {
    product_id: cashierProduct.value.id,
    pay_way: cashierPayWay.value.pay_way,
    pay_type: cashierPayWay.value.pay_type,
    user_id: userId.value,
    host: host,
    device: "mobile",
    amount: cashierOrder.value.amount,
    order_no: cashierOrder.value.orderNo,
    subject: cashierOrder.value.subject,
  })
    .then((res) => {
      loadingToast.clear();
      cashierLoading.value = false;
      const payData = res.data;
      if (!payData) {
        showFailToast("未获取到支付信息");
        return;
      }

      let payLink = "";
      if (typeof payData === "string") {
        payLink = payData;
      } else if (typeof payData === "object") {
        cashierOrder.value.orderNo = payData.order_no || payData.orderNo || cashierOrder.value.orderNo;
        if (payData.amount || payData.total_amount) {
          cashierOrder.value.amount = Number(payData.amount || payData.total_amount);
        }
        payLink =
          payData.pay_url ||
          payData.url ||
          payData.h5_url ||
          payData.h5Url ||
          payData.payLink ||
          payData.qr ||
          payData.qrcode ||
          payData.qr_code ||
          payData.code_url ||
          payData.data ||
          "";
      }

      payLink = payLink ? String(payLink) : "";

      if (!payLink) {
        if (isWechatPay.value) {
          showFailToast("未获取到微信支付链接");
        } else {
          showFailToast("未获取到支付链接");
        }
        return;
      }

      cashierPayUrl.value = payLink;
      if (isWechatPay.value) {
        QRCode.toDataURL(payLink, { width: 260, height: 260, margin: 2 }, (error, url) => {
          if (error) {
            console.error(error);
            showFailToast("生成二维码失败");
          } else {
            cashierQRCode.value = url;
          }
        });
      }
    })
    .catch((e) => {
      loadingToast.clear();
      cashierLoading.value = false;
      showFailToast("生成支付订单失败：" + e.message);
    });
};

const refreshCashierOrder = () => {
  if (!cashierProduct.value || !cashierPayWay.value) {
    return;
  }
  cashierOrder.value.orderNo = generateOrderNo();
  createPaymentOrder();
};

const handleOpenPayLink = () => {
  if (!cashierPayUrl.value) {
    showFailToast("支付链接生成中，请稍后");
    return;
  }
  window.location.href = cashierPayUrl.value;
};

const handleCashierSuccess = () => {
  cashierDialog.value = false;
  showSuccessToast("支付结果已提交，正在刷新账户信息");
  fetchUserProfile();
};

const handleCashierClose = () => {
  cashierDialog.value = false;
};

const copyOrderNo = () => {
  if (!cashierOrder.value.orderNo) {
    showNotify({ type: "warning", message: "暂未生成订单号" });
    return;
  }
  if (navigator.clipboard) {
    navigator.clipboard
      .writeText(cashierOrder.value.orderNo)
      .then(() => {
        showSuccessToast("订单号已复制");
      })
      .catch(() => {
        fallbackCopyTextToClipboard(cashierOrder.value.orderNo);
      });
  } else {
    fallbackCopyTextToClipboard(cashierOrder.value.orderNo);
  }
};

const logout = function () {
  httpGet("/api/user/logout")
    .then(() => {
      removeUserToken();
      store.setIsLogin(false);
      router.push("/mobile/login");
    })
    .catch(() => {
      showFailToast("注销失败！");
    });
};

// 复制到剪贴板
const copyToClipboard = (text) => {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(text).then(() => {
      showSuccessToast("复制成功！");
    }).catch(() => {
      fallbackCopyTextToClipboard(text);
    });
  } else {
    fallbackCopyTextToClipboard(text);
  }
};

// 备用复制方法
const fallbackCopyTextToClipboard = (text) => {
  const textArea = document.createElement("textarea");
  textArea.value = text;
  textArea.style.top = "0";
  textArea.style.left = "0";
  textArea.style.position = "fixed";
  document.body.appendChild(textArea);
  textArea.focus();
  textArea.select();
  try {
    document.execCommand('copy');
    showSuccessToast("复制成功！");
  } catch (err) {
    showFailToast("复制失败！");
  }
  document.body.removeChild(textArea);
};

// 拨打电话
const callPhone = (phone) => {
  window.location.href = `tel:${phone}`;
};

// 复制邀请链接
const copyInviteURL = () => {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(inviteURL.value).then(() => {
      showSuccessToast("链接已复制");
    }).catch(() => {
      fallbackCopyTextToClipboard(inviteURL.value);
    });
  } else {
    fallbackCopyTextToClipboard(inviteURL.value);
  }
};

// 复制上传链接
const copyUploadURL = () => {
  if (!uploadURL.value) {
    showFailToast("上传链接生成中，请稍后");
    return;
  }
  if (navigator.clipboard) {
    navigator.clipboard.writeText(uploadURL.value).then(() => {
      showSuccessToast("链接已复制");
    }).catch(() => {
      fallbackCopyTextToClipboard(uploadURL.value);
    });
  } else {
    fallbackCopyTextToClipboard(uploadURL.value);
  }
};


//转账操作

const showPasswordDialog = ref(false)
const pass = ref({
  old: "",
  new: "",
  renew: ""
})

// 分享相关
const showShareOverlay = ref(false)
const shareUrl = ref("")


const transferDialog = ref(false)
const transfer = ref({
  uid: "",
  power: "",
  password: ""
})


const passwordDialog = ref(false)

// 处理转账点击事件，判断权限
const handleTransferClick = () => {
  // 判断是否为代理商（lev > 0）
  if ((form.value.lev || 0) <= 0) {
    showFailToast("代理商才能进行转账操作");
    return;
  }
  // 代理商可以正常打开转账对话框
  transferDialog.value = true;
};

// 处理提现点击事件，判断权限
const handleWithdrawClick = () => {
  // 判断是否为代理商（lev > 0）
  if ((form.value.lev || 0) <= 0) {
    showFailToast("代理商才能进行提现操作");
    return;
  }
  // 代理商可以正常跳转到提现页面
  router.push('/mobile/withdraw');
};

// 处理算力值明细点击事件
const handlePowerDetailClick = () => {
  router.push('/mobile/powerlog');
};

// 处理积分明细点击事件
const handlePointsDetailClick = () => {
  // TODO: 跳转到积分明细页面
  router.push('/mobile/pointslog');
 // console.log('跳转到积分明细页面');
};

const updateTransferBefore = () => {
  if (!transfer.value.uid) {
    return showNotify({ type: "danger", message: "请输入收款人推荐码" });
  }
  if (!transfer.value.power) {
    return showNotify({ type: "danger", message: "请输入算力值" });
  }
  transfer.value.password = "";
  transferDialog.value = false;
  passwordDialog.value = true;
};

const updateTransfer = () => {
  if (!transfer.value.uid) {
    return showNotify({ type: "danger", message: "请输入收款人推荐码" });
  }
  if (!transfer.value.power) {
    return showNotify({ type: "danger", message: "请输入算力值" });
  }
  if (!transfer.value.password) {
    return showNotify({ type: "danger", message: "请输入密码" });
  }

  httpPost("/api/user/transfer", {
    uid: transfer.value.uid,
    power: transfer.value.power,
    password: transfer.value.password,
  })
    .then(() => {
      showSuccessToast("转账成功");
      passwordDialog.value = false;
      transfer.value = {
        uid: "",
        power: "",
        password: "",
      };
      setTimeout(() => location.reload(), 1000);
    })
    .catch((e) => {
      showFailToast("转账失败，" + e.message);
      passwordDialog.value = false;
    });
};

const openMockCashier = (product) => {


  // const amount = Number(product?.discount ?? product?.price ?? 0)
  // const subject = product?.name || product?.title || '套餐支付'
  // const now = new Date()
  // const orderNo = `MOCK${now.getFullYear()}${(now.getMonth()+1).toString().padStart(2,'0')}${now.getDate().toString().padStart(2,'0')}${now.getHours().toString().padStart(2,'0')}${now.getMinutes().toString().padStart(2,'0')}${now.getSeconds().toString().padStart(2,'0')}${Math.floor(Math.random()*1000).toString().padStart(3,'0')}`
  // mockOrder.value = { orderNo, amount, subject, productId: Number(product?.id || 0) }
  // mockPayDialog.value = true

  // 模拟支付按钮现改为直接走真实联诚收银台
  openLcswCashier(product);


}

const confirmMockPaid = () => {
  const payTime = Math.floor(Date.now() / 1000)
  const payload = {
    order_no: mockOrder.value.orderNo,
    amount: mockOrder.value.amount,
    subject: mockOrder.value.subject,
    user_id: userId.value,
    product_id: mockOrder.value.productId,
    pay_time: payTime,
    pay_method: 'mock',
    status: 'paid'
  }
  httpPost('/api/order/mockPaid', payload)
    .then(() => {
      mockPayDialog.value = false
      showSuccessToast('支付成功（模拟）')
      fetchUserProfile()
    })
    .catch((e) => {
      mockPayDialog.value = false
      showFailToast('上报订单失败（模拟）：' + e.message)
    })
}

const getPayButtonLabel = (payWay) => {
  if (!payWay) {
    return "去支付";
  }
  const payType = (payWay.pay_type || payWay.pay_way || "").toString().toLowerCase();
  if (payType.includes("alipay")) {
    return "支付宝";
  }
  if (payType.includes("wechat") || payType.includes("weixin")) {
    return "微信支付";
  }
  return payWay.name || payWay.title || "去支付";
};

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
};

// 分享功能
const share = () => {
  if (!isLogin.value) {
    return showLoginDialog(router);
  }
  
  // 获取邀请码并跳转到分享页面
  httpGet("/api/invite/code")
    .then((res) => {
      if (res.data && res.data.code) {
        // 设置标记，表示这是用户自己点击分享跳转的
        sessionStorage.setItem('showShareGuide', 'true');
        // 跳转到分享页面，带上 referral 参数
        router.push(`/mobile/share?referral=${res.data.code}`);
      } else {
        showFailToast("获取邀请码失败");
      }
    })
    .catch((e) => {
      showFailToast("获取分享链接失败：" + e.message);
    });
};

// 显示分享引导遮罩层
const showShareGuide = () => {
  showShareOverlay.value = true;
};

// 隐藏分享引导遮罩层
const hideShareGuide = () => {
  showShareOverlay.value = false;
};

</script>

<style lang="stylus">
.mobile-user-profile {
  min-height 100vh
  background #f5f5f5
  position relative
  overflow hidden

  .bg-decoration {
    position absolute
    top 0
    left 0
    right 0
    height 300px
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

    // 用户信息卡片
    .user-card {
      background rgba(255, 255, 255, 0.95)
      border-radius 28px
      padding 32px
      margin-bottom 28px
      box-shadow 0 6px 24px rgba(168, 230, 207, 0.12)
      backdrop-filter blur(12px)
      border 1px solid rgba(168, 230, 207, 0.25)
      animation slideInUp 0.6s ease-out

      .avatar-section {
        display flex
        align-items center
        margin-bottom 28px

        .avatar-container {
          position relative
          margin-right 24px
          cursor pointer

          .van-image {
            border-radius 50%
            box-shadow 0 6px 20px rgba(168, 230, 207, 0.25)
            transition transform 0.3s ease
          }

          .avatar-ring {
            position absolute
            top -8px
            left -8px
            right -8px
            bottom -8px
            border-radius 50%
            background transparent
            border 2px solid rgba(168, 230, 207, 0.3)
            z-index -1
          }

          .avatar-upload-overlay {
            position absolute
            top 0
            left 0
            right 0
            bottom 0
            border-radius 50%
            background rgba(0, 0, 0, 0.4)
            display flex
            align-items center
            justify-content center
            opacity 0
            transition opacity 0.3s ease
            z-index 1
          }

          &:hover {
            .van-image {
              transform scale(1.05)
            }

            .avatar-upload-overlay {
              opacity 1
            }
          }

          &:active {
            transform scale(0.98)
          }
        }

        .user-info {
          flex 1

          .username {
            font-size 24px
            font-weight 600
            color #2d5a4a
            margin 0 0 8px 0
          }

          .user-id {
            font-size 15px
            color #7a8b8b
            margin 0
          }

          .user-level {
            margin-top 8px
            display inline-block
            padding 4px 12px
            border-radius 12px
            font-size 12px
            font-weight 500
            transition all 0.3s ease

            .level-text {
              display inline-block
            }

            &.level-0 {
              background linear-gradient(135deg, #e0e0e0 0%, #bdbdbd 100%)
              color #616161
              box-shadow 0 2px 4px rgba(0, 0, 0, 0.1)
            }

            &.level-1 {
              background linear-gradient(135deg, #ffd700 0%, #ffa500 50%, #ff8c00 100%)
              color #fff
              box-shadow 0 3px 12px rgba(255, 215, 0, 0.4), 0 0 20px rgba(255, 165, 0, 0.2)
              border 1px solid rgba(255, 255, 255, 0.3)
              position relative
              overflow hidden
              
              &::before {
                content ''
                position absolute
                top -50%
                left -50%
                width 200%
                height 200%
                background linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.3), transparent)
                animation shine 3s infinite
              }
            }

            &.level-2 {
              background linear-gradient(135deg, #ff6b35 0%, #f7931e 50%, #ff4500 100%)
              color #fff
              box-shadow 0 3px 15px rgba(255, 107, 53, 0.5), 0 0 25px rgba(255, 69, 0, 0.3)
              border 1px solid rgba(255, 255, 255, 0.4)
              position relative
              overflow hidden
              
              &::before {
                content ''
                position absolute
                top -50%
                left -50%
                width 200%
                height 200%
                background linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.4), transparent)
                animation shine 3s infinite
              }
            }

            &.level-3 {
              background linear-gradient(135deg, #ff1493 0%, #ff69b4 50%, #dc143c 100%)
              color #fff
              box-shadow 0 4px 20px rgba(255, 20, 147, 0.6), 0 0 30px rgba(220, 20, 60, 0.4)
              border 1px solid rgba(255, 255, 255, 0.5)
              position relative
              overflow hidden
              
              &::before {
                content ''
                position absolute
                top -50%
                left -50%
                width 200%
                height 200%
                background linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.5), transparent)
                animation shine 2.5s infinite
              }
            }

            &.level-7 {
              background linear-gradient(135deg, #cd7f32 0%, #b87333 50%, #8b4513 100%)
              color #fff
              box-shadow 0 3px 15px rgba(205, 127, 50, 0.5), 0 0 25px rgba(139, 69, 19, 0.3)
              border 1px solid rgba(255, 255, 255, 0.4)
              position relative
              overflow hidden
              
              &::before {
                content ''
                position absolute
                top -50%
                left -50%
                width 200%
                height 200%
                background linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.4), transparent)
                animation shine 3s infinite
              }
            }

            &.level-8 {
              background linear-gradient(135deg, #c0c0c0 0%, #a8a8a8 50%, #808080 100%)
              color #fff
              box-shadow 0 3px 15px rgba(192, 192, 192, 0.5), 0 0 25px rgba(128, 128, 128, 0.3)
              border 1px solid rgba(255, 255, 255, 0.4)
              position relative
              overflow hidden
              
              &::before {
                content ''
                position absolute
                top -50%
                left -50%
                width 200%
                height 200%
                background linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.4), transparent)
                animation shine 2.8s infinite
              }
            }
            
            &.level-9 {
              background linear-gradient(135deg, #FFD700 0%, #FFA500 50%, #DAA520 100%)
              color #fff
              box-shadow 0 4px 20px rgba(255, 215, 0, 0.6), 0 0 30px rgba(218, 165, 32, 0.4)
              border 1px solid rgba(255, 255, 255, 0.5)
              position relative
              overflow hidden
              
              &::before {
                content ''
                position absolute
                top -50%
                left -50%
                width 200%
                height 200%
                background linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.5), transparent)
                animation shine 2.5s infinite
              }
            }
          }
        }
      }

      .stats-section {
        display flex
        align-items center
        justify-content space-around
        padding 24px 0
        background #f0fdf4
        border-radius 20px
        margin-top 24px
        border 1px solid rgba(168, 230, 207, 0.25)

        .stat-item {
          text-align center
          flex 1

          .stat-value {
            font-size 22px
            font-weight 600
            color #2d5a4a
            margin-bottom 8px
          }

          .stat-label-wrapper {
            display flex
            align-items center
            justify-content center
            gap 6px

            .stat-label {
              font-size 14px
              color #7a8b8b
            }

            .stat-icon {
              cursor pointer
              transition all 0.3s ease
              flex-shrink 0
              display inline-block
              
              &:active {
                transform scale(0.9)
                opacity 0.7
              }
              
              &:hover {
                opacity 0.8
              }
            }

            .recharge-btn {
              margin-left 8px
              font-size 12px
              padding 4px 12px
              height 24px
              border-radius 12px
              background-color #07C160
              border-color #07C160
              box-shadow 0 2px 6px rgba(7, 193, 96, 0.3)
              
              &:active {
                transform scale(0.95)
                background-color #06AD56
                border-color #06AD56
              }
              
              &:hover {
                background-color #06AD56
                border-color #06AD56
              }
            }
            
            .withdraw-btn {
              margin-left 8px
              font-size 12px
              padding 4px 12px
              height 24px
              border-radius 12px
              background-color #576B95
              border-color #576B95
              box-shadow 0 2px 6px rgba(87, 107, 149, 0.3)
              
              &:active {
                transform scale(0.95)
                background-color #4A5A7F
                border-color #4A5A7F
              }
              
              &:hover {
                background-color #4A5A7F
                border-color #4A5A7F
              }
            }
          }
        }

        .stat-divider {
          width 1px
          height 60px
          background #a8e6cf
        }
      }

      // 我的订单
      .order-section {
        margin-top 20px
        padding 18px 16px
        border-radius 20px
        background #ffffff
        box-shadow 0 8px 20px rgba(0, 0, 0, 0.03)
        border 1px solid #f0f0f0

        .section-header {
          display flex
          align-items baseline
          justify-content space-between
          margin-bottom 12px

          .section-title {
            display flex
            align-items center
            font-size 16px
            font-weight 600
            color #333

            .iconfont {
              margin-right 6px
              font-size 18px
              color #409EFF
            }
          }

          .section-subtitle {
            font-size 12px
            color #999
          }
        }

        .order-list {
          display flex
          flex-direction column
          gap 10px
        }

        .order-item {
          display flex
          justify-content space-between
          align-items flex-start
          padding 10px 12px
          border-radius 14px
          background #f8fafc
          border 1px solid #eef2f7

          .order-left {
            flex 1
            min-width 0

            .order-title {
              font-size 14px
              font-weight 500
              color #333
              margin-bottom 4px
            }

            .order-meta {
              font-size 11px
              color #999
              display flex
              flex-direction column
              gap 2px

              .order-id {
                word-break break-all
              }
            }
          }

          .order-right {
            text-align right
            margin-left 12px
            white-space nowrap

            .order-amount {
              font-size 16px
              font-weight 600
              color #e74c3c
              margin-bottom 4px
            }

            .order-status {
              font-size 11px
              padding 2px 8px
              border-radius 999px

              &.success {
                background rgba(103, 194, 58, 0.08)
                color #67c23a
              }

              &.pending {
                background rgba(230, 162, 60, 0.08)
                color #e6a23c
              }

              &.default {
                background rgba(144, 147, 153, 0.08)
                color #909399
              }
            }
          }
        }
      }
    }

    // 快速操作菜单
    .quick-actions {
      margin-bottom 28px
      animation slideInUp 0.6s ease-out 0.2s both
      background rgba(255, 255, 255, 0.95)
      border-radius 24px
      padding 20px
      box-shadow 0 6px 24px rgba(168, 230, 207, 0.12)
      backdrop-filter blur(12px)
      border 1px solid rgba(168, 230, 207, 0.25)

      :deep(.van-grid-item__content) {
        background #f0fdf4
        border-radius 16px
        transition all 0.3s ease
        border 1px solid rgba(168, 230, 207, 0.25)
        padding 16px 8px

        &:active {
          transform scale(0.95)
          background #e8f9f0
        }

        &:hover {
          transform translateY(-2px)
          box-shadow 0 4px 12px rgba(168, 230, 207, 0.2)
        }
      }

      :deep(.van-grid-item__icon) {
        margin-bottom 8px
      }

      .grid-text {
        margin-top 8px
        color #2d5a4a
        font-size 14px
        font-weight 500
      }
    }

    // 推荐码区域
    .invite-section {
      animation slideInUp 0.6s ease-out 0.35s both
      margin-bottom 28px

      .section-header {
        text-align center
        margin-bottom 28px

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

      .invite-card {
        background rgba(255, 255, 255, 0.95)
        border-radius 24px
        padding 32px
        box-shadow 0 6px 24px rgba(168, 230, 207, 0.12)
        backdrop-filter blur(12px)
        border 1px solid rgba(168, 230, 207, 0.25)
        transition all 0.3s ease

        &:hover {
          transform translateY(-4px)
          box-shadow 0 12px 40px rgba(168, 230, 207, 0.2)
        }

        .invite-qrcode {
          text-align center
          margin-bottom 24px

          .qrcode-image {
            border-radius 12px
            box-shadow 0 4px 12px rgba(0, 0, 0, 0.1)
            margin-bottom 12px
          }

          .qrcode-tip {
            font-size 14px
            color #7a8b8b
            margin 0
          }
        }

        .invite-url-box {
          .invite-url {
            display flex
            align-items center
            justify-content space-between
            padding 16px
            background #f0fdf4
            border-radius 16px
            border 1px solid rgba(168, 230, 207, 0.25)

            span {
              flex 1
              font-size 14px
              color #2d5a4a
              word-break break-all
              margin-right 12px
            }

            .copy-icon {
              font-size 18px
              color #7fcdcd
              cursor pointer
              transition all 0.3s ease

              &:hover {
                color #5a9fa0
                transform scale(1.1)
              }

              &:active {
                transform scale(0.95)
              }
            }
          }
        }
      }
    }

    // 邀请统计区域
    .invite-stats-section {
      animation slideInUp 0.6s ease-out 0.38s both
      margin-bottom 28px

      .section-header {
        text-align center
        margin-bottom 28px

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
      }

      .stats-grid {
        display grid
        grid-template-columns repeat(3, 1fr)
        gap 16px

        .stat-card {
          background rgba(255, 255, 255, 0.95)
          border-radius 20px
          padding 24px
          text-align center
          box-shadow 0 4px 16px rgba(168, 230, 207, 0.12)
          backdrop-filter blur(12px)
          border 1px solid rgba(168, 230, 207, 0.25)
          transition all 0.3s ease

          &:hover {
            transform translateY(-4px)
            box-shadow 0 8px 24px rgba(168, 230, 207, 0.2)
          }

          .stat-value {
            font-size 32px
            font-weight 600
            color #2d5a4a
            margin-bottom 8px
          }

          .stat-label {
            font-size 14px
            color #7a8b8b
            font-weight 500
          }
        }
      }
    }

  }

  .setting-content {
    padding 16px
  }

  // 客服弹窗样式
  .service-content {
    padding 20px

    .service-section {
      margin-bottom 24px

      &:last-child {
        margin-bottom 0
      }

      .section-title {
        display flex
        align-items center
        gap 8px
        font-size 16px
        font-weight 600
        color #2d5a4a
        margin 0 0 15px 0

        .iconfont {
          font-size 18px
          color #7fcdcd
        }
      }

      .contact-list {
        .contact-item {
          display flex
          justify-content space-between
          align-items center
          padding 12px 0
          border-bottom 1px solid #f0f0f0

          &:last-child {
            border-bottom none
          }

          .contact-label {
            font-size 14px
            color #666
            font-weight 500
          }

          .contact-value {
            display flex
            align-items center
            gap 10px

            span {
              font-size 14px
              color #333
              font-weight 500
            }

            .copy-btn {
              font-size 12px
              padding 4px 8px
              border-radius 4px
            }
          }
        }
      }

      .qrcode-container {
        text-align center
        padding 20px 0

        .qrcode-image {
          border-radius 12px
          box-shadow 0 4px 12px rgba(0, 0, 0, 0.1)
          margin-bottom 12px
        }

        .qrcode-tip {
          font-size 14px
          color #666
          margin 0
        }
      }

      .service-time {
        p {
          margin 0 0 8px 0
          font-size 14px
          color #333

          &:last-child {
            margin-bottom 0
          }
        }

        .time-tip {
          color #999
          font-size 12px
          font-style italic
        }
      }
    }
  }

  .cashier-dialog {
    .van-dialog__header {
      font-weight 600
      font-size 18px
      color #2d5a4a
    }

    .cashier-content {
      padding 12px 16px 20px
      min-width 280px
    }

    .cashier-summary {
      background #f0fdf4
      border-radius 12px
      padding 12px 14px
      margin-bottom 16px

      .summary-row {
        display flex
        justify-content space-between
        align-items center
        font-size 14px
        color #2d5a4a
        margin-bottom 8px

        &:last-child {
          margin-bottom 0
        }

        .label {
          color #7a8b8b
        }

        .value.amount {
          font-size 18px
          font-weight 600
          color #e74c3c
        }

        .order-no {
          color #20a0ff
          cursor pointer
          text-decoration underline

          &.disabled {
            color #9e9e9e
            cursor default
            text-decoration none
          }
        }
      }
    }

    .cashier-tip {
      margin 10px 0
      text-align center
      font-size 13px
      color #666
    }

    .qrcode-wrapper {
      display flex
      flex-direction column
      align-items center
      justify-content center
      gap 12px
    }

    .cashier-pay {
      .van-button {
        margin-top 12px
      }
    }

    .cashier-actions {
      margin-top 12px

      .van-button + .van-button {
        margin-top 8px
      }
    }
  }
}

// 动画定义
@keyframes shine {
  0% {
    transform translateX(-100%) translateY(-100%) rotate(45deg)
  }
  100% {
    transform translateX(100%) translateY(100%) rotate(45deg)
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

@keyframes pulse {
  0% {
    transform scale(1)
  }
  50% {
    transform scale(1.05)
  }
  100% {
    transform scale(1)
  }
}

// 分享遮罩层样式
.share-overlay {
  position fixed
  top 0
  left 0
  right 0
  bottom 0
  background rgba(0, 0, 0, 0.85)
  z-index 9999
  display flex
  align-items flex-start
  justify-content center
  padding-top 96px
  animation fadeIn 0.3s ease-out

  .share-arrow {
    position absolute
    top 10px
    right 20px
    width 60px
    height 60px
    background-image url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="white"><path d="M15 5l-1.41 1.41L18.17 11H2v2h16.17l-4.59 4.59L15 19l7-7-7-7z" transform="rotate(-45 12 12)"/></svg>')
    background-repeat no-repeat
    background-size contain
    animation floatArrow 1.5s infinite ease-in-out
  }

  .share-guide-content {
    text-align center
    color white
    padding 0 32px
    max-width 320px

    .share-guide-title {
      font-size 20px
      font-weight 600
      margin 0 0 16px 0
      color white

      .share-dots {
        font-size 28px
        display inline-block
      }
    }

    .share-guide-text {
      font-size 18px
      line-height 1.6
      margin 0 0 32px 0
      color white

      .share-highlight {
        color #07c160
        font-weight 600
      }
    }

    .share-tip-box {
      background rgba(255, 255, 255, 0.1)
      border-radius 12px
      padding 16px
      font-size 14px
      color rgba(255, 255, 255, 0.9)
      line-height 1.6
    }
  }
}

@keyframes fadeIn {
  from {
    opacity 0
  }
  to {
    opacity 1
  }
}

@keyframes floatArrow {
  0%, 100% {
    transform translateY(0) rotate(10deg)
  }
  50% {
    transform translateY(-10px) rotate(10deg)
  }
}

// 响应式设计
@media (max-width: 480px) {
  .mobile-user-profile {
    .content {
      padding 20px 16px 100px

      .user-card {
        padding 24px
        border-radius 24px

        .avatar-section {
          margin-bottom 24px

          .avatar-container {
            margin-right 20px

            .van-image {
              width 85px
              height 85px
            }
          }

          .user-info {
            .username {
              font-size 22px
            }
          }
        }

        .stats-section {
          padding 20px 0
          margin-top 20px

          .stat-item {
            .stat-value {
              font-size 20px
            }
          }
        }
      }

      .quick-actions {
        padding 16px
        margin-bottom 24px

        :deep(.van-grid-item__content) {
          padding 12px 6px
        }

        .grid-text {
          font-size 13px
        }

        :deep(.van-icon) {
          font-size 24px
        }
      }

      .invite-section {
        .section-header {
          margin-bottom 24px

          .section-title {
            font-size 24px
          }
        }

        .invite-card {
          padding 24px

          .invite-qrcode {
            .qrcode-image {
              width 180px
              height 180px
            }
          }

          .invite-url-box {
            .invite-url {
              span {
                font-size 12px
              }
            }
          }
        }
      }

      .invite-stats-section {
        .section-header {
          margin-bottom 24px

          .section-title {
            font-size 24px
          }
        }

        .stats-grid {
          gap 12px

          .stat-card {
            padding 20px

            .stat-value {
              font-size 28px
            }

            .stat-label {
              font-size 13px
            }
          }
        }
      }

      .recharge-popup {
        padding 20px 16px

        .section-header {
          margin-bottom 24px

          .section-title {
            font-size 24px
          }
        }

        .product-grid {
          gap 20px

          .product-card {
            padding 24px
            border-radius 20px

            .card-header {
              margin-bottom 20px

              .product-name {
                font-size 20px
              }
            }

            .card-content {
              margin-bottom 24px

              .price-section {
                gap 16px
                margin-bottom 16px

                .current-price {
                  font-size 28px
                }

                .original-price {
                  font-size 18px
                }
              }

              .power-info {
                font-size 15px
                padding 10px 14px
              }
            }

            .card-footer {
              .pay-buttons {
                gap 12px

                .pay-btn-wrapper {
                  min-width 120px

                  .pay-btn {
                    height 44px
                    font-size 15px
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}

// 充值弹窗样式（全局样式，因为 van-popup 挂载到 body）
.recharge-popup-wrapper {
  box-shadow 0 20px 60px rgba(0, 0, 0, 0.15) !important
  border none !important
}

.recharge-popup {
  padding 24px 20px
  background #fff
  max-height 85vh
  overflow-y auto
  display flex
  flex-direction column

  .section-header {
    margin-bottom 24px
    padding-bottom 16px
    border-bottom 1px solid #f0f0f0

    .section-title {
      font-size 20px
      font-weight 700
      color #1a1a1a
      margin 0 0 8px 0
      line-height 1.4
    }

    .section-subtitle {
      font-size 13px
      color #8a8a8a
      font-weight 400
      text-align left
      margin 0
    }
  }

  .product-grid {
    display grid
    grid-template-columns repeat(3, 1fr)
    gap 14px
    margin-bottom 24px

    @media (max-width: 480px) {
      gap 10px
    }

    .product-card {
      background #ffffff
      border-radius 12px
      padding 20px 12px
      border 1px solid #e5e7eb
      cursor pointer
      transition all 0.3s ease
      position relative
      text-align center
      min-height 100px
      display flex
      flex-direction column
      justify-content center
      align-items center

      &:hover {
        border-color #409EFF
        box-shadow 0 2px 8px rgba(64, 158, 255, 0.15)
      }

      &.selected {
        background linear-gradient(135deg, #409EFF 0%, #337ECC 100%)
        border-color #409EFF
        border-width 2px
        box-shadow 0 4px 12px rgba(64, 158, 255, 0.3)

        .product-amount {
          color #ffffff
          font-weight 700
        }

        .product-label {
          color rgba(255, 255, 255, 0.9)
        }
      }

      .product-amount {
        font-size 22px
        font-weight 700
        color #1a1a1a
        margin-bottom 6px
        line-height 1.2
      }

      .product-label {
        font-size 12px
        color #999
        font-weight 400
      }

      .product-check {
        position absolute
        top 8px
        right 8px
        width 20px
        height 20px
        background rgba(255, 255, 255, 0.3)
        border-radius 50%
        display none
        align-items center
        justify-content center
        backdrop-filter blur(4px)
      }

      &.selected .product-check {
        display flex
      }
    }

    .custom-amount-card {
      .custom-amount-input-wrapper {
        display flex
        align-items center
        justify-content center
        gap 4px
        margin-bottom 6px

        .currency-symbol {
          font-size 26px
          font-weight 700
          color #999
        }

        .custom-amount-input {
          flex 1
          padding 0
          background transparent
          border none
          text-align center
          max-width 100px

          :deep(.van-field__control) {
            font-size 26px
            font-weight 700
            color #999
            padding 0
            text-align center
          }

          :deep(.van-field__body) {
            padding 0
          }

          :deep(.van-field__placeholder) {
            font-size 16px
            color #999
            font-weight 400
          }
        }
      }

      &.selected {
        .currency-symbol {
          color #ffffff
        }

        .custom-amount-input {
          :deep(.van-field__control) {
            color #ffffff
          }

          :deep(.van-field__placeholder) {
            color rgba(255, 255, 255, 0.7)
          }
        }
      }
    }
  }

  .balance-info {
    padding 16px 0
    margin-bottom 20px
    display flex
    justify-content space-between
    align-items flex-start

    .balance-item {
      display flex
      flex-direction column
      gap 8px
      flex 1

      &.balance-item-right {
        margin-left auto
        flex 0 0 auto
      }

      .balance-label {
        font-size 13px
        color #666
        font-weight 400
      }

      .balance-value-wrapper {
        display flex
        align-items center
        gap 6px

        .balance-value {
          font-size 18px
          font-weight 700
          color #1a1a1a
        }
      }

      .balance-value {
        font-size 18px
        font-weight 700
        color #1a1a1a

        &.credit {
          color #409EFF
          font-size 20px
        }
      }
    }
  }

  .global-pay-section {
    margin-top 0

    .global-pay-btn {
      height 52px
      font-size 16px
      font-weight 600
      background linear-gradient(135deg, #07C160 0%, #06AD56 100%)
      border none
      box-shadow 0 4px 12px rgba(7, 193, 96, 0.3)
      transition all 0.3s ease
      display flex
      align-items center
      justify-content center
      gap 8px

      :deep(.van-icon) {
        margin-right 0
      }

      &:hover:not(:disabled) {
        background linear-gradient(135deg, #06AD56 0%, #059A4A 100%)
        box-shadow 0 6px 16px rgba(7, 193, 96, 0.4)
        transform translateY(-1px)
      }

      &:active:not(:disabled) {
        transform translateY(0) scale(0.98)
      }

      &:disabled {
        opacity 0.5
        cursor not-allowed
        background #cbd5e1
        box-shadow none
        transform none
      }
    }
  }
}

// 订单弹窗样式（全局样式，因为 van-popup 挂载到 body）
.order-popup-wrapper {
  box-shadow 0 -4px 20px rgba(0, 0, 0, 0.1) !important
  border none !important
}

.order-popup {
  height 100%
  display flex
  flex-direction column
  background #fff
  overflow hidden

  .order-popup-header {
    padding 20px 16px 16px
    border-bottom 1px solid #f0f0f0
    flex-shrink 0

    .order-popup-title {
      display flex
      align-items center
      font-size 18px
      font-weight 600
      color #333
      margin 0 0 6px 0
      gap 8px

      .iconfont {
        font-size 20px
        color #409EFF
      }
    }

    .order-popup-subtitle {
      font-size 13px
      color #999
      font-weight 400
    }
  }

  .order-popup-content {
    flex 1
    overflow-y auto
    padding 16px
    padding-bottom 20px

    // 自定义滚动条样式
    &::-webkit-scrollbar {
      width 4px
    }

    &::-webkit-scrollbar-track {
      background transparent
    }

    &::-webkit-scrollbar-thumb {
      background rgba(0, 0, 0, 0.2)
      border-radius 2px

      &:hover {
        background rgba(0, 0, 0, 0.3)
      }
    }

    .order-list-full {
      display flex
      flex-direction column
      gap 12px
    }

    .order-item {
      display flex
      justify-content space-between
      align-items flex-start
      padding 14px 16px
      border-radius 14px
      background #f8fafc
      border 1px solid #eef2f7
      transition all 0.3s ease

      &:hover {
        background #f1f5f9
        border-color #e2e8f0
        transform translateY(-2px)
        box-shadow 0 4px 12px rgba(0, 0, 0, 0.08)
      }

      .order-left {
        flex 1
        min-width 0

        .order-title {
          font-size 15px
          font-weight 500
          color #333
          margin-bottom 6px
        }

        .order-meta {
          font-size 12px
          color #999
          display flex
          flex-direction column
          gap 4px

          .order-time {
            color #666
          }

          .order-id {
            word-break break-all
            color #999
            font-size 11px
          }
        }
      }

      .order-right {
        text-align right
        margin-left 16px
        white-space nowrap

        .order-amount {
          font-size 18px
          font-weight 600
          color #e74c3c
          margin-bottom 6px
        }

        .order-status {
          font-size 12px
          padding 4px 10px
          border-radius 999px
          display inline-block

          &.success {
            background rgba(103, 194, 58, 0.1)
            color #67c23a
          }

          &.pending {
            background rgba(230, 162, 60, 0.1)
            color #e6a23c
          }

          &.default {
            background rgba(144, 147, 153, 0.1)
            color #909399
          }
        }
      }
    }
  }
}

</style>
