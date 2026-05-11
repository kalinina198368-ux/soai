<template>
  <div class="custom-scroll">
    <div class="page-invitation">
      <div class="inner">
        <div class="page-header">
          <template v-if="isEditingBrand">
            <input
              v-model="form.div"
              class="brand-title brand-title-input"
              @blur="saveBrandEdit"
              @keyup.enter="saveBrandEdit"
              maxlength="30"
              autofocus
              placeholder="请输入品牌"
            />
          </template>
          <template v-else>
            <h2
              :class="{'brand-title': form.div}"
              @click="startEditBrand"
              style="cursor: pointer;"
              title="点击编辑品牌"
            >
              {{ form.div ? form.div : '个人中心' }}
            </h2>
          </template>
        </div>

        <van-form >
          <van-cell-group inset v-model="form" class="user-info-group"  >
            <van-field class="avatar-field">
              <template #label>
                <div class="field-label">
                  <!-- <span>头像</span> -->
                </div>
              </template>
              <template #input>
                <van-uploader
                  v-model="fileList"
                  reupload
                  max-count="1"
                  :deletable="false"
                  :after-read="afterRead"
                  class="avatar-uploader"
                />
              </template>
            </van-field>



            <van-field>
              <template #label>
                <span style="font-size: 16px;">用户名</span>
              </template>
              <template #input>
                <div>
                  <span style="font-size: 20px">{{ form.username }}</span>
                  <span v-if="form.lev" style="margin-left:10px; color:#FF9800; font-weight:bold; font-size:18px; background:#fffbe6; padding:2px 10px; border-radius:12px; border:1px solid #FFD700; box-shadow:0 0 6px #FFD700; vertical-align:middle;">
                    {{ form.lev }}钻
                  </span>
                </div>
              </template>
            </van-field>

            

            <van-field>
              <template #label>
                <span style="font-size: 16px;">邮 箱</span>
              </template>
              <template #input>
                <div v-if="!form.mail">
                  <a href="javascript:void(0);" @click="showMailDialog = true" style="font-size: 16px;">绑定</a>
                </div>
                <div v-else>
                  <span style="font-size: 20px">{{ hiddenMail }}</span>
                </div>
              </template>
            </van-field>

            <van-field>
              <template #label>
                <span style="font-size: 16px;">转账码</span>
              </template>
              <template #input>
                <span style="font-size: 20px">{{userS.zzcode }}</span>
              </template>
            </van-field>

            <van-field>
              <template #label>
                <span style="font-size: 16px;">算力余额</span>
              </template>
              <template #input>
                <span style="font-size: 20px; color: green;">{{ form.power }}</span>
              </template>
            </van-field>

            <van-field>
              <template #label>
                <span style="font-size: 16px;">Q2余额</span>
              </template>
              <template #input>
                <span style="font-size: 20px; color: green;">{{ form.coin }}</span>
              </template>
            </van-field>

          

            <van-field>
              <template #label>
                <span style="font-size: 16px;">BSC地址</span>
              </template>
              <template #input>
                <div>
                  <van-field
                    v-model="form.eth_address"
                    placeholder="请输入BSC收款地址"
                    @blur="updateEthAddress"
                    style="padding: 0;"
                  />
                </div>
              </template>
            </van-field>

            

            <div class="quick-actions">
              <van-grid :column-num="4" :border="false" :gutter="16" :clickable="true">
                <van-grid-item @click="router.push('/mobile/powerlog')" class="grid-item">
                  <template #icon>
                    <van-icon name="balance-list-o" size="28" color="#409EFF" />
                  </template>
                  <template #text><span class="grid-text">算力</span></template>
                </van-grid-item>


                <van-grid-item @click="router.push('/mobile/q2log')" class="grid-item">
                  <template #icon>
                    <van-icon name="balance-list-o" size="28" color="#409EFF" />
                  </template>
                  <template #text><span class="grid-text">Q2</span></template>
                </van-grid-item>
                
                <van-grid-item @click="transferDialog = true" class="grid-item">
                  <template #icon>
                    <van-icon name="exchange" size="28" color="#67C23A" />
                  </template>
                  <template #text><span class="grid-text">转账</span></template>
                </van-grid-item>
<!--                 
                <van-grid-item @click="startScan" class="grid-item">
                  <template #icon>
                    <van-icon name="scan" size="28" color="#FF6B35" />
                  </template>
                  <template #text><span class="grid-text">扫一扫</span></template>
                </van-grid-item>
                
                <van-grid-item @click="showManualTransfer" class="grid-item">
                  <template #icon>
                    <van-icon name="edit" size="28" color="#9C27B0" />
                  </template>
                  <template #text><span class="grid-text">输入转账码</span></template>
                </van-grid-item> -->

                <van-grid-item @click="getFuli" class="grid-item">
                  <template #icon>
                    <van-icon name="sign" size="28" color="#E6A23C" />
                  </template>
                  <template #text><span class="grid-text">签到</span></template>
                </van-grid-item>

                <van-grid-item @click="router.push('/mobile/trade')" class="grid-item">
                  <template #icon>
                    <van-icon name="orders-o" size="28" color="#F56C6C" />
                  </template>
                  <template #text><span class="grid-text">交易</span></template>
                </van-grid-item>

                <van-grid-item @click="router.push('/mobile/contact')" class="grid-item">
                  <template #icon>
                    <van-icon name="chat-o" size="28" color="#409EFF" />
                  </template>
                  <template #text><span class="grid-text">客服</span></template>
                </van-grid-item>
                                       
                        
              
                <van-grid-item @click="showPasswordDialog = true" class="grid-item">
                  <template #icon>
                    <van-icon name="edit" size="28" color="#909399" />
                  </template>
                  <template #text><span class="grid-text">设置</span></template>
                </van-grid-item>


                <van-grid-item @click="showNodeDialog = true" class="grid-item">
                  <template #icon>
                    <van-icon name="checked" size="28" color="#909399" />
                  </template>
                  <template #text><span class="grid-text">节点</span></template>
                </van-grid-item>

               

                <van-grid-item @click="logout" class="grid-item">
                  <template #icon>
                    <van-icon name="close" size="28" color="#606266" />
                  </template>
                  <template #text><span class="grid-text">退出</span></template>
                </van-grid-item>

              </van-grid>

            
            </div>

                  <!-- <div class="bottom-actions">
          <van-button round block type="primary" @click="showPasswordDialog = true" class="action-btn">
            修改密码
          </van-button>
          <van-button round block class="action-btn logout-btn" @click="logout">
            退出登录
          </van-button>
        </div> -->

            <div class="share-box">
              <div class="invite-qrcode">
                <h2>我的推广码</h2>
                <el-image :src="qrImg"/>
              </div>

              <div class="invite-url">
                <span>{{ inviteURL }}</span>
                <el-icon 
                  class="copy-icon" 
                  @click="copyInviteURL"
                  title="复制链接">
                  <DocumentCopy />
                </el-icon>
              </div>
            </div>

            <!-- 添加转账码二维码展示 
            <div class="share-box">
              <div class="invite-qrcode">
                <h2>我的收款码</h2>
                <el-image :src="transferQrImg"/>
              </div>

              <div class="invite-url">
                <span>{{ userS.zzcode }}</span>
                <el-icon 
                  class="copy-icon" 
                  @click="copyTransferCode"
                  title="复制转账码">
                  <DocumentCopy />
                </el-icon>
              </div>
            </div>-->

            <div class="invite-stats">
              <el-row :gutter="20">
                <el-col :span="8">
                  <div class="item-box blue">
                    <el-row :gutter="10">
                      <el-col :span="24">
                        <div class="item-info">
                          <div class="num">{{ regNum }}</div>
                          <div class="text">直推量</div>
                        </div>
                      </el-col>
                    </el-row>
                  </div>
                </el-col>

                <el-col :span="8">
                  <div class="item-box blue">
                    <el-row :gutter="10">
                      <el-col :span="24">
                        <div class="item-info">
                          <div class="num">{{ jhNum }}</div>
                          <div class="text">激活量</div>
                        </div>
                      </el-col>
                    </el-row>
                  </div>
                </el-col>

                <el-col :span="8">
                  <div class="item-box blue">
                    <el-row :gutter="10">
                      <el-col :span="24">
                        <div class="item-info">
                          <div class="num">{{ hits }}</div>
                          <div class="text">团队数量</div>
                        </div>
                      </el-col>
                    </el-row>
                  </div>
                </el-col>


                
              </el-row>
            </div>

            <h2>直推用户 </h2>
            <h2> <span style="font-size:16px;color:#FFD700;">(总业绩: {{ totalYeji }}，有效总业绩: {{ validYeji }})</span></h2>
            <div class="invite-logs">
              <invite-list v-if="isLogin"/>
            </div>

            

            


          </van-cell-group>
        </van-form>

        <!-- <div class="bottom-actions">
          <van-button round block type="primary" @click="showPasswordDialog = true" class="action-btn">
            修改密码
          </van-button>
          <van-button round block class="action-btn logout-btn" @click="logout">
            退出登录
          </van-button>
        </div> -->

        <div class="bottom-spacing"></div>
      </div>
    </div>
  </div>

  <login-dialog :show="showLoginDialog" @hide="showLoginDialog =  false" @success="initData"/>

  <van-dialog v-model:show="showPasswordDialog" title="修改密码" show-cancel-button
              @confirm="updatePass"
              @cancel="showPasswordDialog = false"
              :before-close="beforeClose">
    <van-form>
      <van-cell2-group inset>
        <van-field
            v-model="pass.old"
            placeholder="旧密码"
        />
        <van-field
            v-model="pass.new"
            type="password"
            placeholder="新密码"
        />
        <van-field
            v-model="pass.renew"
            type="password"
            placeholder="确认密码"
        />
      </van-cell2-group>
    </van-form>
  </van-dialog>


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
            placeholder="算力值"
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




   <!-- 弹出公告对话框 -->
    <van-dialog
          v-model:show="showAnnouncement"
          class="announcement-dialog"
          title="📢 公告"
          show-cancel-button
          cancel-button-text="不再显示"
          confirm-button-text="我知道了"
          @confirm="handleConfirm"
          @cancel="handleDoNotShow"
          width="80%"
          style="--van-dialog-title-font-size: 22px; --van-dialog-message-padding: 20px 15px;"
        >
          <div class="announcement-content">
        
            <div v-html="announcements"></div>
          </div>
        </van-dialog>



        <!-- 邮箱绑定弹出框 -->
    <van-dialog v-model:show="showMailDialog" title="绑定邮箱(用于密码找回)" show-cancel-button
              @confirm="updateMail"
              @cancel="showMailDialog = false "
              :before-close="beforeClose">
    <van-form>
      <van-cell-group inset>
        <van-field
            v-model="bind.mail"
            placeholder="绑定的邮箱地址"
        />

     
      </van-cell-group>
    </van-form>
  </van-dialog>

  <!-- 扫码弹出框 -->
  <van-dialog v-model:show="showScanDialog" title="扫一扫" show-cancel-button
              @confirm="handleScanResult"
              @cancel="closeScan"
              confirm-button-text="手动输入"
              cancel-button-text="关闭"
              :before-close="beforeClose">
    <div class="scan-container">
      <video ref="videoRef" autoplay playsinline class="scan-video"></video>
      <div class="scan-overlay">
        <div class="scan-box">
          <div class="scan-line"></div>
        </div>
        <p class="scan-tip">将二维码放入框内，即可自动扫描</p>
        <div class="scan-actions">
          <van-button size="small" type="primary" @click="switchCamera" v-if="hasMultipleCameras">
            切换摄像头
          </van-button>
          <van-button size="small" @click="showManualInput = true">
            手动输入
          </van-button>
        </div>
      </div>
    </div>
  </van-dialog>

  <!-- 手动输入弹出框 -->
  <van-dialog v-model:show="showManualInput" title="手动输入" show-cancel-button
              @confirm="handleManualInput"
              @cancel="showManualInput = false">
    <van-field
      v-model="manualInputValue"
      placeholder="请输入用户推荐码或转账码"
      clearable
    />
  </van-dialog>

  <!-- 节点信息弹出框 -->
  <van-dialog v-model:show="showNodeDialog" title="节点信息" show-cancel-button
              @confirm="showNodeDialog = false"
              @cancel="showNodeDialog = false">
    <div class="node-info">
      <p>当前节点状态：正常</p>
      <!-- <p>节点地址：{{ nodeAddress }}</p> -->
      <p>连接状态：已连接</p>
    </div>
  </van-dialog>

</template>

<script setup>
import {onMounted, ref,computed, nextTick} from "vue"
import QRCode from "qrcode";
import jsQR from "jsqr";
import {httpGet, httpPost} from "@/utils/http";
import {ElMessage} from "element-plus";
import Clipboard from "clipboard";
import InviteList from "@/components/InviteList.vue";
import {checkSession} from "@/action/session";
import LoginDialog from "@/components/LoginDialog.vue";
import {showConfirmDialog,showFailToast, showNotify, showSuccessToast} from "vant";
import {removeUserToken} from "@/store/session";
import Compressor from 'compressorjs';
import {useRouter} from "vue-router";
import { DocumentCopy } from '@element-plus/icons-vue';


const inviteURL = ref("")
const qrImg = ref("")
const transferQrImg = ref("")
const inviteChatCalls = ref(0)
const inviteImgCalls = ref(0)
const hits = ref(0)
const regNum = ref(0)
const jhNum = ref(0)
const rate = ref(0)
const isLogin = ref(false)
const showLoginDialog = ref(false)
const router = useRouter()
const zzcode=ref("")
const totalYeji = ref(0)
const validYeji = ref(0)



const latexPlugin = require('markdown-it-latex2img')
const mathjaxPlugin = require('markdown-it-mathjax')
const md = require('markdown-it')({
  breaks: true,
  html: true,
  linkify: true,
  typographer: true,
  highlight: function (str, lang) {
    const codeIndex = parseInt(Date.now()) + Math.floor(Math.random() * 10000000)
    // 显示复制代码按钮
    const copyBtn = `<span class="copy-code-btn" data-clipboard-action="copy" data-clipboard-target="#copy-target-${codeIndex}">复制</span>
<textarea style="position: absolute;top: -9999px;left: -9999px;z-index: -9999;" id="copy-target-${codeIndex}">${str.replace(/<\/textarea>/g, '&lt;/textarea>')}</textarea>`
    
    // 处理代码高亮
    const preCode = md.utils.escapeHtml(str)
    // 将代码包裹在 pre 中
    return `<pre class="code-container"><code class="language-${lang} hljs">${preCode}</code>${copyBtn}</pre>`
  }
});

md.use(latexPlugin)
md.use(mathjaxPlugin)

const fileList = ref([
  {
    url: '',
    message: '上传中...',
  }
]);

//邮箱隐藏

    // 计算属性：隐藏部分邮箱信息
    const hiddenMail = computed(() => {
      const mail = form.value.mail;
      if (!mail) return ''; // 如果邮箱为空

      // 分割邮箱地址的用户名和域名部分
      const [localPart, domain] = mail.split('@');

      // 只显示最后两位字符，其他部分用星号代替
      const hiddenLocalPart = localPart.slice(-2).padStart(localPart.length, '*');

      return `${hiddenLocalPart}@${domain}`;
    });





const showDialog = ref(false); // 控制弹出框
const newMail = ref(''); // 存储输入的邮箱

const submitEmail = () => {
  if (newMail.value) {
    form.mail = newMail.value; // 更新邮箱
    showDialog.value = false; // 关闭弹出框
    newMail.value = ''; // 清空输入框
  }
};


const form = ref({
  username: '',
  nickname: '',
  mobile: '',
  avatar: '',
  calls: 0,
  tokens: 0,
  eth_address: ''
})

const userS = ref({
  zzcode:''
})





onMounted(() => {
  initData()
  checkShowDialog();
  fetchAnnouncements(); // 获取公告数据
  fetchTotalYeji();


  // 复制链接
  const clipboard = new Clipboard('.copy-link');
  clipboard.on('success', () => {
    ElMessage.success('复制链接成功！');
  })


  // const clipboard2 = new Clipboard('.copy-code');
  // clipboard2.on('success', () => {
  //   ElMessage.success('复制邀请码成功！');
  // })


  clipboard.on('error', () => {
    ElMessage.error('复制失败！');
  })


  httpGet("/api/invite/code").then(res => {
      const text = `${location.protocol}//${location.host}/register?invite_code=${res.data.code}`
      hits.value = res.data["hits"]
      regNum.value = res.data["reg_num"]
      jhNum.value = res.data["jh_num"]//激活数量

      userS.value.zzcode=res.data.code;
          //赋值
      if (hits.value > 0) {
        rate.value = ((regNum.value / hits.value) * 100).toFixed(2)
      }
      QRCode.toDataURL(text, {width: 400, height: 400, margin: 2}, (error, url) => {
        if (error) {
          console.error(error)
        } else {
          qrImg.value = url;
        
        }
      });
      
      // 生成转账码二维码
      const transferData = JSON.stringify({
        type: 'transfer',
        uid: res.data.code,
        code: res.data.code
      });
      QRCode.toDataURL(transferData, {width: 400, height: 400, margin: 2}, (error, url) => {
        if (error) {
          console.error(error)
        } else {
          transferQrImg.value = url;
        }
      });
      
      inviteURL.value = text
    }).catch(e => {
      ElMessage.error("获取邀请码失败：" + e.message)
    })



})

const initData = () => {
  checkSession().then(() => {
    isLogin.value = true



    httpGet('/api/user/profile').then(res => {
      form.value = res.data
      fileList.value[0].url = form.value.avatar
    }).catch((e) => {
      console.log(e.message)
      showFailToast('获取用户信息失败')
    });



    httpGet("/api/invite/code").then(res => {
      const text = `${location.protocol}//${location.host}/register?invite_code=${res.data.code}`
      hits.value = res.data["hits"]
      regNum.value = res.data["reg_num"]
      form.value.userS=res.data.code;
          //赋值
      if (hits.value > 0) {
        rate.value = ((regNum.value / hits.value) * 100).toFixed(2)
      }
      QRCode.toDataURL(text, {width: 400, height: 400, margin: 2}, (error, url) => {
        if (error) {
          console.error(error)
        } else {
          qrImg.value = url;
        
        }
      });
      
      // 生成转账码二维码
      const transferData = JSON.stringify({
        type: 'transfer',
        uid: res.data.code,
        code: res.data.code
      });
      QRCode.toDataURL(transferData, {width: 400, height: 400, margin: 2}, (error, url) => {
        if (error) {
          console.error(error)
        } else {
          transferQrImg.value = url;
        }
      });
      
      inviteURL.value = text
    }).catch(e => {
     // ElMessage.error("获取邀请码失败：" + e.message)
    })

    httpGet("/api/config/get?key=system").then(res => {
      inviteChatCalls.value = res.data["invite_chat_calls"]
      inviteImgCalls.value = res.data["invite_img_calls"]
    }).catch(e => {
      //ElMessage.error("获取系统配置失败：" + e.message)
    })
  }).catch(() => {
    showLoginDialog.value = true
  });
}




const afterRead = (file) => {
  file.status = 'uploading';
  file.message = '上传中...';
  // 压图片并上传
  new Compressor(file.file, {
    quality: 0.6,
    success(result) {
      const formData = new FormData();
      formData.append('file', result, result.name);
      // 执行上传操作
      httpPost('/api/upload', formData).then((res) => {
        form.value.avatar = res.data.url
        file.status = 'success'
        httpPost('/api/user/profile/update', form.value).then(() => {
          showSuccessToast('上传成功')
        }).catch(() => {
          showFailToast('上传失败')
        })
      }).catch((e) => {
        showNotify({type: 'danger', message: '上传失败：' + e.message})
      })
    },
    error(err) {
      console.log(err.message);
    },
  });
}

const showPasswordDialog = ref(false)
const pass = ref({
  old: "",
  new: "",
  renew: ""
})


const transferDialog = ref(false)
const transfer = ref({
  uid: "",
  power: "",
  password: ""
})


const passwordDialog = ref(false)


const showMailDialog = ref(false)
const bind = ref({
  mail: ""
 
})

const beforeClose = (action) => {
  new Promise((resolve) => {
    resolve(action === 'confirm');
  });
}


// 提交修改密码
const updatePass = () => {
  if (pass.value.old === '') {
    return showNotify({type: "danger", message: "请输入密码"})
  }
  if (!pass.value.new || pass.value.new.length < 8) {
    return showNotify({type: "danger", message: "密码的长度为8-16个字符"})
  }
  if (pass.value.renew !== pass.value.new) {
    return showNotify({type: "danger", message: "两次输入密码不一致"})
  }
  httpPost('/api/user/password', {
    old_pass: pass.value.old,
    password: pass.value.new,
    repass: pass.value.renew
  }).then(() => {
    showSuccessToast("更新成功！")
    showPasswordDialog.value = false
  }).catch((e) => {
    showFailToast('更新失败，' + e.message)
    showPasswordDialog.value = false
  })
}


// 提交转账之前 输入密码
const updateTransferBefore = () => {


    // 首先关闭转账对话框，然后弹出密码输入对话框
    transferDialog.value = false;
    passwordDialog.value = true;


}

// 提交转账
const updateTransfer = () => {
  if (transfer.value.uid === '') {
    return showNotify({type: "danger", message: "请输入转账人id"})
  }
  if (transfer.value.power === '') {
    return showNotify({type: "danger", message: "请输入算力值"})
  }
  if (transfer.value.password === '') {
    return showNotify({type: "danger", message: "请输入密码"})
  }
  
  httpPost('/api/user/transfer', {
    uid: transfer.value.uid,
    power: transfer.value.power,
    password: transfer.value.password,
  }).then(() => {
    showSuccessToast("转账成功")
    passwordDialog.value = false
    // 清空转账表单
    transfer.value = {
      uid: "",
      power: "",
      password: ""
    }
    // 刷新用户信息
    setTimeout(() => location.reload(), 1000)
  }).catch((e) => {
    showFailToast('转账失败，' + e.message)
    passwordDialog.value = false
  })
}


// 提交绑定邮箱操作
const updateMail = () => {

if (bind.value.mail === '') {
  return showNotify({type: "danger", message: "请输入需要绑定的邮箱地址"})
}

    //绑定邮箱操作
    httpPost('/api/user/updateMail', {
      mail: bind.value.mail
    }).then(() => {
      showSuccessToast("绑定成功")
      showMailDialog.value = false
      setTimeout(() => location.reload(), 800)
    }).catch((e) => {
      showFailToast('绑定失败，' + e.message)
      showMailDialog.value = false
    })

    //  showSuccessToast("绑定邮箱模拟成功");
      // showMailDialog.value = false
    }



// 领取每日免费福利
const getFuli = () => {
  showConfirmDialog({
    title: '领取福利',
    message:
        '确定领取吗?',
  }).then(() => {

        httpPost("/api/user/getFuli", {
     //   product_id: item.id,
       // user_id: loginUser.value.id
      }).then(() => {
        showSuccessToast("今日免费算力领取成功....")
        //location.reload()
        setTimeout(() => location.reload(), 3000)
      }).catch(e => {
        showFailToast("签到失败" + e.message)
        //location.reload()
         setTimeout(() => location.reload(), 3000)
      })
  }).catch(() => {


  });
}




const logout = function () {
  httpGet('/api/user/logout').then(() => {
    removeUserToken();
    router.push('/login');
  }).catch(() => {
    ElMessage.error('注销失败！');
  })
}



const showAnnouncement = ref(false);
const announcements =  ref("")

// 检查 localStorage 中是否有不再显示的标记
const checkShowDialog = () => {
  const doNotShow = localStorage.getItem('doNotShowAnnouncement');
  //const isError=isBk.value; 

 
  if (!doNotShow ) {
    showAnnouncement.value = true;
  }
};

// 处理 "我知道了" 按钮点击事件
const handleConfirm = () => {
  showAnnouncement.value = false;
};

// 处理 "不再显示" 按钮点击事件
const handleDoNotShow = () => {
  localStorage.setItem('doNotShowAnnouncement', 'true'); // 设置不再显示的标记
  showAnnouncement.value = false;
};


// 模拟用户重新登录时清除 localStorage 中的标记
// 你可以根据项目实际逻辑来调整这个部分，比如在用户登录时清除标记
const clearDoNotShowFlagOnLogin = () => {
  localStorage.removeItem('doNotShowAnnouncement');
};


const fetchAnnouncements = function () {


  // 获取系统公告
  httpGet("/api/config/get?key=notice").then(res => {
  try {
    console.log("Response data:", res.data); // 打印整个响应数据
    const doNotShow = localStorage.getItem('doNotShowAnnouncement');
    if (res.data.content) {
      //announcements.value = res.data.content;
      announcements.value = md.render(res.data['content']);
      if(doNotShow){
        showAnnouncement.value = false; // 无公告时不显示对话框
      }else {

        showAnnouncement.value = true; // 有公告时显示对话框
      }

    
      
    } else if (res.data.updated !== undefined && res.code === 0) {
    // 确保是没有内容的响应
     showAnnouncement.value = false; // 无公告时不显示对话框
  }
    

  //announcements.value = md.render(res.data['content']);

    //announcements.value = md.render(res.data['content']);
    
    // // 判断 announcements 是否为空或 null
    // console.log("Content:", res.data['content']);
    // if (!res.data['content']) {
    //   showAnnouncement.value = false;
    // } else {
    //   showAnnouncement.value = true;
    // }
  } catch (e) {
    showAnnouncement.value = false;
   
    console.error("Error in processing response:", e); // 捕获异常并打印错误
  }
}).catch(e => {
  ElMessage.error("获取系统配置失败：" + e.message);
  console.error("HTTP request error:", e); // 打印请求错误
});




};



// const copyTs = () => {

//   const clipboard2 = new Clipboard('.copy-code');
//   clipboard2.on('success', () => {
//     ElMessage.success('复制邀请码成功！');
//   })
// }

// 复制链接
const copyInviteURL = () => {
  const clipboard = new Clipboard('.copy-icon', {
    text: () => inviteURL.value
  });

  clipboard.on('success', () => {
    ElMessage.success('链接已复制');
    clipboard.destroy();
  });

  clipboard.on('error', () => {
    ElMessage.error('复制失败，请手动复制');
    clipboard.destroy();
  });
};

// 复制转账码
const copyTransferCode = () => {
  const clipboard = new Clipboard('.copy-icon', {
    text: () => userS.value.zzcode
  });

  clipboard.on('success', () => {
    ElMessage.success('转账码已复制');
    clipboard.destroy();
  });

  clipboard.on('error', () => {
    ElMessage.error('复制失败，请手动复制');
    clipboard.destroy();
  });
};

const updateEthAddress = () => {
  httpPost('/api/user/bindEthAddress', {
    eth_address: form.value.eth_address
  }).then(() => {
    showSuccessToast("BSC地址更新成功")
  }).catch((e) => {
    showFailToast('更新失败，' + e.message)
  })
}

// 添加扫码相关的变量和方法
const showScanDialog = ref(false)
const showNodeDialog = ref(false)
const showManualInput = ref(false)
const manualInputValue = ref('')
const nodeAddress = ref('node.example.com:8080')
const videoRef = ref(null)
const scanStream = ref(null)
const scanResult = ref('')
const hasMultipleCameras = ref(false)
const currentFacingMode = ref('environment')

// 开始扫码
const startScan = async () => {
  try {
    // 检查是否为HTTPS环境（本地开发环境除外）
    if (location.protocol !== 'https:' && location.hostname !== 'localhost' && location.hostname !== '127.0.0.1') {
      showFailToast('摄像头功能需要在HTTPS环境下使用')
      return
    }

    // 检查浏览器是否支持媒体设备API
    if (!navigator.mediaDevices) {
      showFailToast('您的浏览器不支持摄像头功能，请使用现代浏览器')
      return
    }

    // 检查是否支持getUserMedia
    if (!navigator.mediaDevices.getUserMedia) {
      showFailToast('您的浏览器版本过低，请升级浏览器后重试')
      return
    }

    // 检查可用摄像头数量
    try {
      const devices = await navigator.mediaDevices.enumerateDevices()
      const videoDevices = devices.filter(device => device.kind === 'videoinput')
      
      if (videoDevices.length === 0) {
        showFailToast('未检测到摄像头设备')
        return
      }
      
      hasMultipleCameras.value = videoDevices.length > 1
    } catch (error) {
      console.warn('无法枚举设备:', error)
      // 继续尝试启动摄像头，可能是权限问题
    }

    showScanDialog.value = true
    
    // 等待对话框显示后再启动摄像头
    await nextTick()
    
    await startCamera()
    
  } catch (error) {
    console.error('启动扫码失败:', error)
    showFailToast('启动扫码功能失败，请检查设备权限')
    showScanDialog.value = false
  }
}

// 启动摄像头
const startCamera = async () => {
  try {
    // 请求摄像头权限
    const constraints = {
      video: { 
        facingMode: currentFacingMode.value,
        width: { ideal: 1280, max: 1920 },
        height: { ideal: 720, max: 1080 }
      }
    }
    
    const stream = await navigator.mediaDevices.getUserMedia(constraints)
    
    scanStream.value = stream
    if (videoRef.value) {
      videoRef.value.srcObject = stream
    }
    
    // 开始扫码检测
    startQRCodeDetection()
    
  } catch (error) {
    console.error('启动摄像头失败:', error)
    
    // 根据错误类型提供具体的错误信息
    let errorMessage = '无法启动摄像头'
    
    if (error.name === 'NotAllowedError') {
      errorMessage = '摄像头权限被拒绝，请在浏览器设置中允许摄像头访问'
    } else if (error.name === 'NotFoundError') {
      errorMessage = '未找到摄像头设备'
    } else if (error.name === 'NotReadableError') {
      errorMessage = '摄像头被其他应用占用，请关闭其他应用后重试'
    } else if (error.name === 'OverconstrainedError') {
      errorMessage = '摄像头不支持所需的分辨率，正在尝试降低要求'
      
      // 尝试使用更低的约束条件
      try {
        const fallbackConstraints = {
          video: { 
            facingMode: currentFacingMode.value,
            width: { ideal: 640 },
            height: { ideal: 480 }
          }
        }
        const fallbackStream = await navigator.mediaDevices.getUserMedia(fallbackConstraints)
        scanStream.value = fallbackStream
        if (videoRef.value) {
          videoRef.value.srcObject = fallbackStream
        }
        startQRCodeDetection()
        return
      } catch (fallbackError) {
        errorMessage = '摄像头启动失败，请检查设备是否正常'
      }
    } else if (error.name === 'NotSupportedError') {
      errorMessage = '您的设备不支持摄像头功能'
    } else if (error.name === 'SecurityError') {
      errorMessage = '安全限制：请确保在HTTPS环境下使用'
    }
    
    showFailToast(errorMessage)
    throw error
  }
}

// 切换摄像头
const switchCamera = async () => {
  if (scanStream.value) {
    scanStream.value.getTracks().forEach(track => track.stop())
  }
  
  currentFacingMode.value = currentFacingMode.value === 'environment' ? 'user' : 'environment'
  
  try {
    await startCamera()
  } catch (error) {
    // 如果切换失败，尝试恢复原来的摄像头
    currentFacingMode.value = currentFacingMode.value === 'environment' ? 'user' : 'environment'
    try {
      await startCamera()
    } catch (fallbackError) {
      showFailToast('切换摄像头失败')
    }
  }
}

// 处理手动输入
const handleManualInput = () => {
  if (manualInputValue.value.trim()) {
    handleScanSuccess(manualInputValue.value.trim())
    showManualInput.value = false
    manualInputValue.value = ''
  } else {
    showFailToast('请输入有效内容')
  }
}

// 二维码检测
const startQRCodeDetection = () => {
  if (!videoRef.value) return
  
  const canvas = document.createElement('canvas')
  const context = canvas.getContext('2d')
  
  const detectQRCode = () => {
    if (!showScanDialog.value || !videoRef.value) return
    
    const video = videoRef.value
    if (video.readyState === video.HAVE_ENOUGH_DATA) {
      canvas.width = video.videoWidth
      canvas.height = video.videoHeight
      context.drawImage(video, 0, 0, canvas.width, canvas.height)
      
      try {
        // 使用 jsQR 进行二维码识别
        const imageData = context.getImageData(0, 0, canvas.width, canvas.height)
        const code = jsQR(imageData.data, imageData.width, imageData.height)
        
        if (code) {
          scanResult.value = code.data
          handleScanSuccess(code.data)
          return
        }
        
      } catch (error) {
        console.error('二维码识别失败:', error)
      }
    }
    
    // 继续检测
    requestAnimationFrame(detectQRCode)
  }
  
  detectQRCode()
}

// 处理扫码成功
const handleScanSuccess = (result) => {
  scanResult.value = result
  closeScan()
  
  try {
    // 尝试解析JSON格式的二维码内容
    const qrData = JSON.parse(result)
    
    if (qrData.type === 'transfer' && (qrData.uid || qrData.code)) {
      // 如果是转账二维码，自动填充转账信息
      transfer.value.uid = qrData.uid || qrData.code
      transferDialog.value = true
      showSuccessToast('扫码成功，已自动填充收款人信息')
    } else if (qrData.uid || qrData.code) {
      // 如果是用户二维码，填充用户ID或推荐码
      transfer.value.uid = qrData.uid || qrData.code
      transferDialog.value = true
      showSuccessToast('扫码成功，已自动填充用户信息')
    } else {
      showFailToast('无效的二维码格式')
    }
  } catch (error) {
    // 如果不是JSON格式，尝试直接作为转账码使用
    if (result && result.trim()) {
      transfer.value.uid = result.trim()
      transferDialog.value = true
      showSuccessToast('扫码成功，已自动填充转账码')
    } else {
      showFailToast('无效的二维码内容')
    }
  }
}

// 处理扫码结果
const handleScanResult = () => {
  // 点击确认按钮时显示手动输入
  showManualInput.value = true
}

// 关闭扫码
const closeScan = () => {
  showScanDialog.value = false
  
  // 停止摄像头
  if (scanStream.value) {
    scanStream.value.getTracks().forEach(track => track.stop())
    scanStream.value = null
  }
  
  if (videoRef.value) {
    videoRef.value.srcObject = null
  }
  
  scanResult.value = ''
}

// 显示手动输入转账码对话框
const showManualTransfer = () => {
  showManualInput.value = true
  manualInputValue.value = ''
}

const fetchTotalYeji = async () => {
  try {
    // 先获取总条数
    const res = await httpPost('/api/invite/list', { page: 1, page_size: 1 });
    const total = res.data?.total || 0;
    if (total === 0) {
      totalYeji.value = 0;
      validYeji.value = 0;
      return;
    }
    // 再获取所有数据
    const resAll = await httpPost('/api/invite/list', { page: 1, page_size: total });
    const items = resAll.data?.items || [];
    const yejiArr = items.map(item => Number(item.yeji) || 0);
    const sum = yejiArr.reduce((sum, v) => sum + v, 0);
    const max = yejiArr.length > 0 ? Math.max(...yejiArr) : 0;
    totalYeji.value = sum;
    validYeji.value = sum - max;
  } catch (e) {
    totalYeji.value = 0;
    validYeji.value = 0;
  }
};

const isEditingBrand = ref(false);

function startEditBrand() {
  isEditingBrand.value = true;
}
function saveBrandEdit() {
  isEditingBrand.value = false;
  if (!form.value.div) {
    form.value.div = '';
  }

  // 绑定品牌，保存到后端
  httpPost('/api/user/saveDiv', {
    div: form.value.div
  }).then(() => {
    showSuccessToast("品牌修改成功");
    setTimeout(() => location.reload(), 800);
  }).catch((e) => {
    showFailToast('品牌修改失败，' + e.message);
  });
}

</script>

<style lang="stylus" scoped>
@import "@/assets/css/custom-scroll.styl"


.quick-actions {
  margin: 20px 16px;
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.05);
}

.grid-item {
  padding: px 8px;
}

.grid-text {
  margin-top: 8px;
  color: #ffffff;
  font-size: 14px;
  font-weight: 500;
}

:deep(.van-grid-item__content) {
  background: #f8f9fa;
  border-radius: 8px;
  transition: all 0.3s ease;
}

:deep(.van-grid-item__content:active) {
  background: #f0f2f5;
  transform: scale(0.98);
}

.gray-button {
  background-color: #808080;
  color: white;
}


.page-invitation .inner {
  max-width: 1000px;
  width: 100%;
}

.announcement-content {
  max-height: 300px;
  overflow-y: auto;
  padding-right: 10px;
}

.announcement-content p {
  margin: 10px 0;
  font-size: 16px;
  color: #666;
}



.van-field__label,
.van-tag,
.van-button {
  font-size: 18px; /* 修改字体大小 */
}


.van-cell {
  position: relative;
  display: flex;
  box-sizing: border-box;
  width: 100%;
  padding: var(--van-cell-vertical-padding) var(--van-cell-horizontal-padding);
  overflow: hidden;
  color: var(--van-cell-text-color);
  font-size: var(--van-cell-font-size);
  line-height: var(--van-cell-line-height);
}

.van-field__control {
  font-size: 18px; /* 修改输入框字体大小 */
}

.button-row {
  display: flex;
  gap: 10px; /* 可选：按钮之间间距 */
}

.page-invitation {
  display: flex;
  justify-content: center;
  background-color: #282c34;
  overflow-x: hidden;
  overflow-y: visible;
  
  .inner {
    display: flex;
    flex-flow: column;
    max-width: 1000px;
    width: 100%;
    color: #e1e1e1;
  }
}

.bottom-actions {
  margin: 20px 15px;
  padding: 15px;
  background: rgba(255,255,255,0.05);
  border-radius: 12px;
  
  .action-btn {
    margin-bottom: 15px;
    height: 44px;
    font-size: 16px;
    
    &:last-child {
      margin-bottom: 0;
    }
  }
  
  .logout-btn {
    background: #808080;
    border: none;
    
    &:active {
      opacity: 0.8;
    }
  }
}

// 添加底部间距，为底部导航栏留出空间
.bottom-spacing {
  height: 60px; // 根据底部导航栏的高度调整
}

.page-header
  text-align center
  padding 20px 0
  margin-bottom 20px
  
  h2
    font-size 24px
    color #ffffff
    margin 0

.user-info-group
  background rgba(255,255,255,0.05)
  border-radius 12px
  padding 1px
  margin-bottom 20px

.field-label
  font-size 16px
  color #ffffff
  
.field-value
  font-size 18px
  color #ffffff

.avatar-uploader
  :deep(.van-uploader__upload)
    width 80px
    height 80px
    border-radius 50%

.quick-actions
  margin 20px 0
  padding 15px
  background rgba(255,255,255,0.05)
  border-radius 12px
  
  :deep(.van-grid-item__content)
    background none
    padding 16px 8px
    
  :deep(.van-icon)
    color #4080ff
    
  :deep(.van-grid-item__text)
    color #ffffff
    margin-top 8px
    font-size 14px

.stats-cards
  margin 20px 0
  
  .stat-card
    background rgba(64,128,255,0.1)
    padding 20px
    border-radius 12px
    text-align center
    
    .stat-value
      font-size 24px
      color #4080ff
      font-weight bold
      
    .stat-label
      font-size 14px
      color #ffffff
      margin-top 8px

.page-invitation {
  display: flex;
  justify-content: center;
  background-color: #282c34;
  overflow-x hidden
  overflow-y visible


  

  .inner {
    display flex
    flex-flow column
    max-width 1000px
    width 100%
    color #e1e1e1

    h2 {
      color #ffffff;
    }

    .share-box {
      .info {
        line-height 1.5
        border 1px solid #444444
        border-radius 10px
        padding 10px

        strong {
          color #f56c6c
        }
      }

      .invite-qrcode {
        padding 50px
        text-align center
      }

      .invite-url {
        padding 15px
        display flex
        justify-content space-between
        border 1px solid #444444
        border-radius 10px

        span {
          position relative
          font-family 'Microsoft YaHei', '微软雅黑', Arial, sans-serif
          top 5px
        }
      }
    }

    .invite-stats {
      padding 30px 10px

      .item-box {
        border-radius 10px
        padding 0 10px

        .el-col {
          height 140px
          display flex
          align-items center
          justify-content center

          .iconfont {
            font-size 60px
          }

          .item-info {
            font-size 18px

            .text, .num {
              padding 3px 0
              text-align center
            }

            .num {
              font-size 40px
            }
          }
        }
      }

      .yellow {
        background-color #ffeecc
        color #D68F00
      }

      .blue {
        background-color #D6E4FF
        color #1062FE
      }

      .green {
        background-color #E7F8EB
        color #2D9F46
      }
    }


    .invite-logs {
      padding-bottom 20px
    }
  }

}

.copy-icon {
  cursor: pointer;
  color: #f4f4f5;
  margin-left: 8px;
  font-size: 16px;
  
  &:hover {
    color: #409EFF;
  }
}

// 扫码相关样式
.scan-container {
  position: relative;
  width: 100%;
  height: 300px;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
}

.scan-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.scan-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.scan-box {
  width: 200px;
  height: 200px;
  border: 2px solid #fff;
  border-radius: 8px;
  position: relative;
  
  &::before,
  &::after {
    content: '';
    position: absolute;
    width: 20px;
    height: 20px;
    border: 3px solid #409EFF;
  }
  
  &::before {
    top: -3px;
    left: -3px;
    border-right: none;
    border-bottom: none;
  }
  
  &::after {
    bottom: -3px;
    right: -3px;
    border-left: none;
    border-top: none;
  }
}

.scan-line {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 2px;
  background: linear-gradient(90deg, transparent, #409EFF, transparent);
  animation: scanAnimation 2s linear infinite;
}

@keyframes scanAnimation {
  0% {
    top: 0;
    opacity: 1;
  }
  50% {
    opacity: 1;
  }
  100% {
    top: 100%;
    opacity: 0;
  }
}

.scan-tip {
  color: #fff;
  margin-top: 20px;
  font-size: 14px;
  text-align: center;
}

.scan-actions {
  margin-top: 15px;
  display: flex;
  gap: 10px;
  
  .van-button {
    background: rgba(255, 255, 255, 0.2);
    border: 1px solid rgba(255, 255, 255, 0.3);
    color: #fff;
    
    &:hover {
      background: rgba(255, 255, 255, 0.3);
    }
  }
}

.node-info {
  padding: 20px;
  
  p {
    margin: 10px 0;
    font-size: 16px;
    color: #666;
  }
}

.brand-title {
  color: #95d475 !important;
  font-size: 28px;
  font-weight: 800;
  letter-spacing: 2px;
}

.brand-title-input {
  background: transparent;
  border: none;
  outline: none;
  color: #b1b3b8 !important;
  font-size: 28px;
  font-weight: 800;
  text-align: center;
  width: 100%;
  letter-spacing: 2px;
  padding: 0;
}

</style>
