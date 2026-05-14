<template>
  <div class="system-config form" v-loading="loading">
    <el-tabs v-model="activeName" class="sys-tabs">
      <el-tab-pane label="系统配置" name="basic">
        <div class="container">
          <el-form :model="system" label-width="150px" label-position="right" ref="systemFormRef" :rules="rules">
            <el-tabs type="border-card">
              <el-tab-pane label="基础配置">
                <el-form-item label="网站标题" prop="title">
                  <el-input v-model="system['title']" />
                </el-form-item>
                <el-form-item label="控制台标题" prop="admin_title">
                  <el-input v-model="system['admin_title']" />
                </el-form-item>
                <el-form-item label="网站Slogan" prop="slogan">
                  <el-input v-model="system['slogan']" />
                </el-form-item>
                <el-form-item label="网站 LOGO" prop="logo">
                  <el-input v-model="system['logo']" placeholder="网站LOGO图片">
                    <template #append>
                      <el-upload :auto-upload="true" :show-file-list="false" @click="beforeUpload('logo')" :http-request="uploadImg">
                        <el-icon class="uploader-icon">
                          <UploadFilled />
                        </el-icon>
                      </el-upload>
                    </template>
                  </el-input>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      首页背景图
                      <el-tooltip effect="dark" content="网站首页背景图片" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex justify-between w-100">
                    <el-input v-model="system['index_bg_url']" placeholder="网站首页背景图片">
                      <template #append>
                        <el-upload :auto-upload="true" :show-file-list="false" @click="beforeUpload('index_bg_url')" :http-request="uploadImg">
                          <el-icon class="uploader-icon">
                            <UploadFilled />
                          </el-icon>
                        </el-upload>
                      </template>
                    </el-input>
                    <el-button class="ml-1" type="primary" @click="system.index_bg_url = 'https://api.dujin.org/bing/1920.php'">使用动态背景</el-button>
                    <el-button class="ml-1" @click="system.index_bg_url = 'color'">使用纯色背景</el-button>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      首页导航菜单
                      <el-tooltip effect="dark" content="被选中的菜单将会在首页导航栏显示" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-select v-model="system['index_navs']" multiple :filterable="true" placeholder="请选择菜单，多选" style="width: 100%">
                    <el-option v-for="item in menus" :key="item.id" :label="item.name" :value="item.id" />
                  </el-select>
                </el-form-item>

                <!-- <el-form-item label="版权信息" prop="copyright">
                  <el-input v-model="system['copyright']" placeholder="更改此选项需要获取 License 授权" />
                </el-form-item> -->

                <!-- <el-form-item>
                  <template #label>
                    <div class="label-title">
                      开放注册
                      <el-tooltip effect="dark" content="关闭注册之后只能通过管理后台添加用户" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-switch v-model="system['enabled_register']" />
                </el-form-item> -->

                <!-- <el-form-item>
                  <template #label>
                    <div class="label-title">
                      启用验证码
                      <el-tooltip
                        effect="dark"
                        content="启用验证码之后，注册登录都会加载行为验证码，增加安全性。此功能需要购买验证码服务才会生效。"
                        raw-content
                        placement="right"
                      >
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-switch v-model="system['enabled_verify']" />
                </el-form-item> -->

                <el-form-item label="注册方式" prop="register_ways">
                  <el-checkbox-group v-model="system['register_ways']">
                    <el-checkbox value="mobile">手机注册</el-checkbox>
                    <el-checkbox value="email">邮箱注册</el-checkbox>
                    <el-checkbox value="username">用户名注册</el-checkbox>
                  </el-checkbox-group>
                </el-form-item>

                <el-form-item label="邮件域名白名单" prop="register_ways">
                  <items-input v-model:value="system['email_white_list']" />
                </el-form-item>

                <el-form-item label="微信客服二维码" prop="wechat_card_url">
                  <el-input v-model="system['wechat_card_url']" placeholder="微信客服二维码">
                    <template #append>
                      <el-upload :auto-upload="true" :show-file-list="false" @click="beforeUpload('wechat_card_url')" :http-request="uploadImg">
                        <el-icon class="uploader-icon">
                          <UploadFilled />
                        </el-icon>
                      </el-upload>
                    </template>
                  </el-input>
                </el-form-item>
                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      默认翻译模型
                      <el-tooltip effect="dark" content="选择一个默认模型来翻译提示词" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-select v-model.number="system['translate_model_id']" :filterable="true" placeholder="选择一个默认模型来翻译提示词" style="width: 100%">
                    <el-option v-for="item in models" :key="item.id" :label="item.name" :value="item.id" />
                  </el-select>
                </el-form-item>

                <el-form-item label="开启聊天上下文">
                  <el-switch v-model="system['enable_context']" />
                </el-form-item>
                <el-form-item label="会话上下文深度">
                  <div class="tip-input-line">
                    <el-input-number v-model="system['context_deep']" :min="0" :max="10" />
                    <div class="tip">
                      会话上下文深度：在老会话中继续会话，默认加载多少条聊天记录作为上下文。如果设置为 0
                      则不加载聊天记录，仅仅使用当前角色的上下文。该配置参数必须设置需要为偶数。
                    </div>
                  </div>
                </el-form-item>
<!-- 
                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      SD反向提示词
                      <el-tooltip effect="dark" content="Stable-Diffusion 绘画默认反向提示词" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-input type="textarea" :rows="2" v-model="system['sd_neg_prompt']" placeholder="" />
                </el-form-item> -->

                <el-form-item label="会员充值说明" prop="order_pay_timeout">
                  <template #label>
                    <div class="label-title">
                      会员充值说明
                      <el-tooltip effect="dark" content="会员充值页面的充值说明文字" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-input type="textarea" :rows="2" v-model="system['vip_info_text']" placeholder="" />
                </el-form-item>

                <el-form-item label="MJ默认API模式" prop="mj_mode">
                  <el-select v-model="system['mj_mode']" placeholder="请选择模式">
                    <el-option v-for="item in mjModels" :value="item.value" :label="item.name" :key="item.value">{{ item.name }} </el-option>
                  </el-select>
                </el-form-item>
              </el-tab-pane>

              <el-tab-pane label="算力配置">
                 <el-form-item label="注册赠送算力" prop="init_power">
                  <el-input v-model.number="system['init_power']" placeholder="新用户注册赠送算力" />
                </el-form-item>
                <!--
                <el-form-item label="邀请赠送算力" prop="invite_power">
                  <el-input v-model.number="system['invite_power']" placeholder="邀请新用户注册赠送算力" />
                </el-form-item> -->
                <!-- <el-form-item label="VIP每月赠送算力" prop="vip_month_power">
                  <el-input v-model.number="system['vip_month_power']" placeholder="VIP用户每月赠送算力" />
                </el-form-item> -->
                <!-- <el-form-item>
                  <template #label>
                    <div class="label-title">
                      签到赠送算力
                      <el-tooltip effect="dark" content="每日签到赠送算力" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-input v-model.number="system['daily_power']" placeholder="默认值0" />
                </el-form-item> -->


                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      nb算力
                      <el-tooltip effect="dark" content="使用nano banana画一张图消耗算力" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-input v-model.number="system['mj_power']" placeholder="" />
                </el-form-item>


                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      nab2算力
                      <el-tooltip effect="dark" content="使用nano banana2画一张图消耗算力" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-input v-model.number="system['mj_power2']" placeholder="" />
                </el-form-item>


                <!-- <el-form-item>
                  <template #label>
                    <div class="label-title">
                      MJ操作算力
                      <el-tooltip effect="dark" content="放大，变换，重绘操作一次消耗的算力" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-input v-model.number="system['mj_action_power']" placeholder="" />
                </el-form-item> -->
                <!-- <el-form-item label="Stable-Diffusion算力" prop="sd_power">
                  <el-input v-model.number="system['sd_power']" placeholder="使用Stable-Diffusion画一张图消耗算力" />
                </el-form-item> -->
                <!-- <el-form-item label="DALL-E-3算力" prop="dall_power">
                  <template #label>
                    <div class="label-title">
                      MJ操作算力
                      <el-tooltip effect="dark" content="主要用户函数调用 DALL-E-3 进行绘画" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-input v-model.number="system['dall_power']" placeholder="使用DALL-E-3画一张图消耗算力" />
                </el-form-item> -->
                <!-- <el-form-item label="Suno 算力" prop="suno_power">
                  <el-input v-model.number="system['suno_power']" placeholder="使用 Suno 生成一首音乐消耗算力" />
                </el-form-item> -->

                <el-form-item label="sora 算力" prop="luma_power">
                  <el-input v-model.number="system['luma_power']" placeholder="使用 sora  生成一段视频消耗算力" />
                </el-form-item>
                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      高级语音算力
                      <el-tooltip effect="dark" content="使用一次 OpenAI 高级语音对话消耗的算力" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-input v-model.number="system['advance_voice_power']" placeholder="" />
                </el-form-item>
                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      提示词算力
                      <el-tooltip effect="dark" content="生成AI绘图提示词，歌词，视频描述消耗的算力" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <el-input v-model.number="system['prompt_power']" placeholder="" />
                </el-form-item>
              </el-tab-pane>


              <!-- 会员配置 -->

               <el-tab-pane label="会员配置">
                <el-divider content-position="left">升级条件配置</el-divider>
                
                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      1星升级条件
                      <el-tooltip effect="dark" content="达到1星会员所需的累计充值金额" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <!-- <el-input-number v-model.number="system['star1_upgrade_direct_count']" :min="0" placeholder="直推人数" style="width: 150px" />
                    <span>个直推</span> -->
                    <el-input-number v-model.number="system['star1_upgrade_recharge']" :min="0" :precision="2" placeholder="充值金额" style="width: 150px" />
                    <span>元充值</span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      2星升级条件
                      <el-tooltip effect="dark" content="达到2星会员所需的累计充值金额" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <!-- <el-input-number v-model.number="system['star2_upgrade_star1_count']" :min="0" placeholder="一星数量" style="width: 150px" />
                    <span>个一星（直推+间推）</span> -->

                     <el-input-number v-model.number="system['star2_upgrade_recharge']" :min="0" :precision="2" placeholder="充值金额" style="width: 150px" />
                    <span>元充值</span>

                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      3星升级条件
                      <el-tooltip effect="dark" content="达到3星会员所需的累计充值金额" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <!-- <el-input-number v-model.number="system['star3_upgrade_star2_count']" :min="0" placeholder="二星数量" style="width: 150px" />
                    <span>个二星（直推+间推）</span> -->

                    
                     <el-input-number v-model.number="system['star3_upgrade_recharge']" :min="0" :precision="2" placeholder="充值金额" style="width: 150px" />
                    <span>元充值</span>
                  </div>
                </el-form-item>

                <el-divider content-position="left">佣金比例配置</el-divider>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      普通会员佣金
                      <el-tooltip effect="dark" content="普通会员佣金比例（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    直推(%)
                    <el-input-number v-model.number="system['normal_direct_commission']" :min="0" :max="100" :precision="2" placeholder="直推比例" style="width: 150px" />
                    <span></span>
                    间推(%)
                    <el-input-number v-model.number="system['normal_indirect_commission']" :min="0" :max="100" :precision="2" placeholder="间推比例" style="width: 150px" />
                    <span></span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      1星会员佣金
                      <el-tooltip effect="dark" content="1星会员佣金（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    直推(%)
                    <el-input-number v-model.number="system['star1_direct_commission']" :min="0" :max="100" :precision="2" placeholder="直推比例" style="width: 150px" />
                    <span></span>
                     间推(%)
                    <el-input-number v-model.number="system['star1_indirect_commission']" :min="0" :max="100" :precision="2" placeholder="间推比例" style="width: 150px" />
                    <span></span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      2星会员佣金
                      <el-tooltip effect="dark" content="2星会员佣金（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                     直推(%)
                    <el-input-number v-model.number="system['star2_direct_commission']" :min="0" :max="100" :precision="2" placeholder="直推比例" style="width: 150px" />
                    <span></span>
                    间推(%)
                    <el-input-number v-model.number="system['star2_indirect_commission']" :min="0" :max="100" :precision="2" placeholder="间推比例" style="width: 150px" />
                    <span></span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      3星会员佣金
                      <el-tooltip effect="dark" content="3星会员佣金（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    直推(%)
                    <el-input-number v-model.number="system['star3_direct_commission']" :min="0" :max="100" :precision="2" placeholder="直推比例" style="width: 150px" />
                    <span></span>
                    间推(%)
                    <el-input-number v-model.number="system['star3_indirect_commission']" :min="0" :max="100" :precision="2" placeholder="间推比例" style="width: 150px" />
                    <span></span>
                  </div>
                </el-form-item>

                <el-divider content-position="left">折扣配置</el-divider>

                <!-- <el-form-item>
                  <template #label>
                    <div class="label-title">
                      普通会员折扣
                      <el-tooltip effect="dark" content="普通会员下单可享受的折扣比例（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <el-input-number v-model.number="system['normal_discount']" :min="0" :max="100" :precision="2" placeholder="折扣百分比" style="width: 150px" />
                    <span>%</span>
                  </div>
                </el-form-item> -->

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      1星会员折扣
                      <el-tooltip effect="dark" content="一星会员下单可享受的折扣比例（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <el-input-number v-model.number="system['star1_discount']" :min="0" :max="100" :precision="2" placeholder="折扣百分比" style="width: 150px" />
                    <span>%</span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      2星会员折扣
                      <el-tooltip effect="dark" content="二星会员下单可享受的折扣比例（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <el-input-number v-model.number="system['star2_discount']" :min="0" :max="100" :precision="2" placeholder="折扣百分比" style="width: 150px" />
                    <span>%</span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      3星会员折扣
                      <el-tooltip effect="dark" content="三星会员下单可享受的折扣比例（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <el-input-number v-model.number="system['star3_discount']" :min="0" :max="100" :precision="2" placeholder="折扣百分比" style="width: 150px" />
                    <span>%</span>
                  </div>
                </el-form-item>
              </el-tab-pane>



              <el-tab-pane label="代理商配置">
                <el-divider content-position="left">升级条件配置</el-divider>
                
                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      青铜代理条件
                      <el-tooltip effect="dark" content="达到青铜代理条件一次性充值金额" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <el-input-number v-model.number="system['agent_bronze_upgrade_recharge']" :min="0" :precision="2" placeholder="充值金额" style="width: 150px" />
                    <span>元充值</span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      白银代理条件
                      <el-tooltip effect="dark" content="达到白银代理条件一次性充值金额" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                     <el-input-number v-model.number="system['agent_silver_upgrade_recharge']" :min="0" :precision="2" placeholder="充值金额" style="width: 150px" />
                    <span>元充值</span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      黄金代理条件
                      <el-tooltip effect="dark" content="达到黄金代理条件一次性充值金额" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                     <el-input-number v-model.number="system['agent_gold_upgrade_recharge']" :min="0" :precision="2" placeholder="充值金额" style="width: 150px" />
                    <span>元充值</span>
                  </div>
                </el-form-item>

                <el-divider content-position="left">佣金比例配置</el-divider>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      青铜代理佣金
                      <el-tooltip effect="dark" content="青铜代理佣金比例（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    直推(%)
                    <el-input-number v-model.number="system['agent_bronze_direct_commission']" :min="0" :max="100" :precision="2" placeholder="直推比例" style="width: 150px" />
                    <span></span>
                    间推(%)
                    <el-input-number v-model.number="system['agent_bronze_indirect_commission']" :min="0" :max="100" :precision="2" placeholder="间推比例" style="width: 150px" />
                    <span></span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      白银代理佣金
                      <el-tooltip effect="dark" content="白银代理佣金（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    直推(%)
                    <el-input-number v-model.number="system['agent_silver_direct_commission']" :min="0" :max="100" :precision="2" placeholder="直推比例" style="width: 150px" />
                    <span></span>
                     间推(%)
                    <el-input-number v-model.number="system['agent_silver_indirect_commission']" :min="0" :max="100" :precision="2" placeholder="间推比例" style="width: 150px" />
                    <span></span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      黄金代理佣金
                      <el-tooltip effect="dark" content="黄金代理佣金（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                     直推(%)
                    <el-input-number v-model.number="system['agent_gold_direct_commission']" :min="0" :max="100" :precision="2" placeholder="直推比例" style="width: 150px" />
                    <span></span>
                    间推(%)
                    <el-input-number v-model.number="system['agent_gold_indirect_commission']" :min="0" :max="100" :precision="2" placeholder="间推比例" style="width: 150px" />
                    <span></span>
                  </div>
                </el-form-item>


                <el-divider content-position="left">折扣配置</el-divider>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      青铜代理折扣
                      <el-tooltip effect="dark" content="青铜代理下单可享受的折扣比例（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <el-input-number v-model.number="system['agent_bronze_discount']" :min="0" :max="100" :precision="2" placeholder="折扣百分比" style="width: 150px" />
                    <span>%</span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      白银代理折扣
                      <el-tooltip effect="dark" content="白银代理下单可享受的折扣比例（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <el-input-number v-model.number="system['agent_silver_discount']" :min="0" :max="100" :precision="2" placeholder="折扣百分比" style="width: 150px" />
                    <span>%</span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      黄金代理折扣
                      <el-tooltip effect="dark" content="黄金代理下单可享受的折扣比例（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <el-input-number v-model.number="system['agent_gold_discount']" :min="0" :max="100" :precision="2" placeholder="折扣百分比" style="width: 150px" />
                    <span>%</span>
                  </div>
                </el-form-item>

                <el-divider content-position="left">团队奖励配置</el-divider>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      青铜代理团队奖
                      <el-tooltip effect="dark" content="青铜代理对应的团队奖励比例（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <el-input-number v-model.number="system['agent_bronze_team_reward']" :min="0" :max="100" :precision="2" placeholder="团队奖励比例" style="width: 150px" />
                    <span>%</span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      白银代理团队奖
                      <el-tooltip effect="dark" content="白银代理对应的团队奖励比例（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <el-input-number v-model.number="system['agent_silver_team_reward']" :min="0" :max="100" :precision="2" placeholder="团队奖励比例" style="width: 150px" />
                    <span>%</span>
                  </div>
                </el-form-item>

                <el-form-item>
                  <template #label>
                    <div class="label-title">
                      黄金代理团队奖
                      <el-tooltip effect="dark" content="黄金代理对应的团队奖励比例（百分比）" raw-content placement="right">
                        <el-icon>
                          <InfoFilled />
                        </el-icon>
                      </el-tooltip>
                    </div>
                  </template>
                  <div class="d-flex align-center" style="gap: 10px">
                    <el-input-number v-model.number="system['agent_gold_team_reward']" :min="0" :max="100" :precision="2" placeholder="团队奖励比例" style="width: 150px" />
                    <span>%</span>
                  </div>
                </el-form-item>

              </el-tab-pane>



              <!-- <el-divider content-position="left">团队奖励配置</el-divider> -->



            </el-tabs>

            <div style="padding: 10px">
              <el-form-item>
                <el-button type="primary" @click="save('system')">保存</el-button>
              </el-form-item>
            </div>
          </el-form>
        </div>


      </el-tab-pane>
      <el-tab-pane label="公告配置" name="notice">
        <md-editor class="mgb20" v-model="notice" :theme="store.theme" @on-upload-img="onUploadImg" />
        <el-form-item>
          <div style="padding-top: 10px; margin-left: 150px">
            <el-button type="primary" @click="save('notice')">保存</el-button>
          </div>
        </el-form-item>
      </el-tab-pane>
      <el-tab-pane label="思维导图" name="mark_map">
        <md-editor class="mgb20" :theme="store.theme" v-model="system['mark_map_text']" @on-upload-img="onUploadImg" />
        <el-form-item>
          <div style="padding-top: 10px; margin-left: 150px">
            <el-button type="primary" @click="save('system')">保存</el-button>
          </div>
        </el-form-item>
      </el-tab-pane>
      <el-tab-pane label="菜单配置" name="menu">
        <Menu />
      </el-tab-pane>
      <el-tab-pane label="视频数据导入" name="video-import">
        <div class="container" style="padding: 20px">
          <el-form label-width="120px" label-position="right">
            <el-form-item label="JSON 数据">
              <el-input
                v-model="videoJsonData"
                type="textarea"
                :rows="15"
                placeholder="请粘贴从 Postman 获取的 JSON 数据，格式：{&quot;items&quot;:[...],&quot;cursor&quot;:null}"
                style="font-family: monospace;"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="importVideoData" :loading="importing">
                <el-icon v-if="!importing"><Upload /></el-icon>
                导入数据
              </el-button>
              <el-button @click="videoJsonData = ''">清空</el-button>
            </el-form-item>
            <el-form-item v-if="importResult">
              <el-alert
                :title="importResult.title"
                :type="importResult.type"
                :description="importResult.message"
                show-icon
                :closable="true"
                @close="importResult = null"
              />
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>

      <el-tab-pane label="教学 / TTS / 配图" name="teaching">
        <div class="container">
          <el-form :model="system" label-width="180px" label-position="right">
            <el-alert
              type="info"
              :closable="false"
              show-icon
              style="margin-bottom: 16px"
              title="教学配图颗粒度在下方「配图颗粒度」中配置；保存后若 API 未热加载配置，请重启后端服务。"
            />
            <el-divider content-position="left">大模型（与「语言模型」管理同一批模型）</el-divider>
            <el-form-item label="教学文案 / 生图提示词模型">
              <el-select
                v-model.number="system['teaching_script_model_id']"
                :filterable="true"
                clearable
                placeholder="选择用于口播脚本与生图提示词输出的 Chat 模型"
                style="width: 100%"
              >
                <el-option v-for="item in models" :key="item.id" :label="item.name + ' (id=' + item.id + ')'" :value="item.id" />
              </el-select>
              <div class="tip" style="margin-top: 6px">与后台「语言模型」菜单中的模型 ID 一致，使用该模型绑定的 API Key 与网关地址。</div>
            </el-form-item>

            <el-divider content-position="left">配图颗粒度（分镜密度 / 提示词分批）</el-divider>
            <el-form-item label="一节一图（不拆条）">
              <el-switch
                v-model="system['teaching_visual_sparse']"
                :active-value="true"
                :inactive-value="false"
                inline-prompt
                active-text="开"
                inactive-text="关"
              />
              <div class="tip" style="margin-top: 6px">
                <strong>关</strong>：长口播按「每镜最大字数」拆成多条分镜（如 seg-2__1、seg-2__2），多图切换更密。<strong>开</strong>：每个大纲小节只对应一张配图。
              </div>
            </el-form-item>
            <el-form-item label="每镜最大字数">
              <el-input-number v-model="system['teaching_visual_shot_max_runes']" :min="24" :max="120" :step="4" controls-position="right" />
              <span class="tip ml-2">数值越小图越密，建议 36～56；默认 44</span>
            </el-form-item>
            <el-form-item label="全稿最多分镜条数">
              <el-input-number v-model="system['teaching_visual_max_total_shots']" :min="8" :max="48" :step="1" controls-position="right" />
              <span class="tip ml-2">超过后不再为后续口播生成分镜，默认 40</span>
            </el-form-item>
            <el-form-item label="提示词 LLM 每批条数">
              <el-input-number v-model="system['teaching_image_prompts_llm_batch']" :min="4" :max="12" :step="1" controls-position="right" />
              <span class="tip ml-2">每批送给大模型的分镜数，默认 8；偏小更稳、偏大请求次数更少</span>
            </el-form-item>

            <el-divider content-position="left">字节 OpenSpeech TTS（HTTP 单向流式）</el-divider>
            <el-form-item label="X-Api-Key（推荐）">
              <el-input
                v-model="system['tts_api_key']"
                type="password"
                show-password
                placeholder="火山 seed-tts-2.0 等：填控制台下发的 API Key，与 ResourceId 搭配"
              />
              <div class="tip" style="margin-top: 6px">
                若填写此项，请求头仅使用 <code>X-Api-Key</code> + <code>X-Api-Resource-Id</code>，无需再填下方 AppId / AccessKey。
              </div>
            </el-form-item>
            <el-form-item label="TTS 请求地址">
              <el-input
                v-model="system['tts_api_url']"
                placeholder="默认：https://openspeech.bytedance.com/api/v3/tts/unidirectional"
              />
            </el-form-item>
            <el-form-item label="X-Api-App-Id（旧版，可选）">
              <el-input v-model="system['tts_app_id']" placeholder="AppId" />
            </el-form-item>
            <el-form-item label="X-Api-Access-Key（旧版，可选）">
              <el-input v-model="system['tts_access_key']" type="password" show-password placeholder="AccessKey" />
            </el-form-item>
            <el-form-item label="X-Api-Resource-Id">
              <el-input v-model="system['tts_resource_id']" placeholder="例如：seed-tts-2.0" show-password />
            </el-form-item>
            <el-form-item label="发音人 speaker">
              <el-input v-model="system['tts_speaker']" placeholder="例如：zh_female_cancan_mars_bigtts" />
            </el-form-item>
            <el-form-item label="音频格式">
              <el-input v-model="system['tts_audio_format']" placeholder="mp3" />
            </el-form-item>
            <el-form-item label="采样率">
              <el-input-number v-model="system['tts_sample_rate']" :min="8000" :max="48000" :step="1000" />
              <span class="tip ml-2">常用 24000</span>
            </el-form-item>

            <el-divider content-position="left">教学配图（OpenAI 兼容 images/generations）</el-divider>
            <el-form-item label="API 根地址">
              <el-input v-model="system['teaching_image_api_url']" placeholder="例如：https://api.bltcy.ai（不含 /v1）" />
            </el-form-item>
            <el-form-item label="Bearer API Key">
              <el-input v-model="system['teaching_image_api_key']" type="password" show-password placeholder="sk-..." />
            </el-form-item>
            <el-form-item label="生图模型 model">
              <el-input v-model="system['teaching_image_model']" placeholder="例如：gpt-image-2" />
            </el-form-item>
            <el-form-item label="尺寸 size">
              <el-input v-model="system['teaching_image_size']" placeholder="例如：1024x1024" />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="save('system')">保存到系统配置</el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>

    </el-tabs>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from "vue";
import { httpGet, httpPost } from "@/utils/http";
import Compressor from "compressorjs";
import { ElMessage, ElMessageBox } from "element-plus";
import { CloseBold, InfoFilled, Select, UploadFilled, Upload } from "@element-plus/icons-vue";
import MdEditor from "md-editor-v3";
import "md-editor-v3/lib/style.css";
import Menu from "@/views/admin/Menu.vue";
import { copyObj, dateFormat } from "@/utils/libs";
import ItemsInput from "@/components/ui/ItemsInput.vue";
import { useSharedStore } from "@/store/sharedata";

const activeName = ref("basic");
const system = ref({ models: [] });
const configBak = ref({});
const loading = ref(true);
const systemFormRef = ref(null);
const models = ref([]);
const notice = ref("");
const license = ref({ is_active: false });
const menus = ref([]);
const mjModels = ref([
  { name: "慢速（Relax）", value: "relax" },
  { name: "快速（Fast）", value: "fast" },
  { name: "急速（Turbo）", value: "turbo" },
]);
const store = useSharedStore();

// 视频数据导入相关
const videoJsonData = ref("");
const importing = ref(false);
const importResult = ref(null);

const upgradeFields = [
  "star1_upgrade_direct_count",
  "star1_upgrade_recharge",
  "star2_upgrade_star1_count",
  "star2_upgrade_recharge",
  "star3_upgrade_star2_count",
  "star3_upgrade_recharge",
  "agent_bronze_upgrade_recharge",
  "agent_silver_upgrade_recharge",
  "agent_gold_upgrade_recharge",
];

const commissionFields = [
  "normal_direct_commission",
  "normal_indirect_commission",
  "star1_direct_commission",
  "star1_indirect_commission",
  "star2_direct_commission",
  "star2_indirect_commission",
  "star3_direct_commission",
  "star3_indirect_commission",
  "agent_bronze_direct_commission",
  "agent_bronze_indirect_commission",
  "agent_silver_direct_commission",
  "agent_silver_indirect_commission",
  "agent_gold_direct_commission",
  "agent_gold_indirect_commission",
];

const discountFields = [
  "normal_discount",
  "star1_discount",
  "star2_discount",
  "star3_discount",
  "agent_bronze_discount",
  "agent_silver_discount",
  "agent_gold_discount",
];

const agentTeamRewardFields = [
  "agent_bronze_team_reward",
  "agent_silver_team_reward",
  "agent_gold_team_reward",
];

const initTeachingVisualDefaults = (s) => {
  if (typeof s !== "object" || !s) return;
  if (s.teaching_visual_sparse === undefined || s.teaching_visual_sparse === null) s.teaching_visual_sparse = false;
  if (s.teaching_visual_shot_max_runes === undefined || s.teaching_visual_shot_max_runes === null) {
    s.teaching_visual_shot_max_runes = 44;
  }
  if (s.teaching_visual_max_total_shots === undefined || s.teaching_visual_max_total_shots === null) {
    s.teaching_visual_max_total_shots = 40;
  }
  if (s.teaching_image_prompts_llm_batch === undefined || s.teaching_image_prompts_llm_batch === null) {
    s.teaching_image_prompts_llm_batch = 8;
  }
};

const ensureFields = (target, fields, defaultValue = 0) => {
  fields.forEach((field) => {
    if (target[field] === undefined || target[field] === null) {
      target[field] = defaultValue;
    }
  });
};

onMounted(() => {
  // 加载系统配置
  httpGet("/api/admin/config/get?key=system")
    .then((res) => {
      system.value = res.data;
      
      // 初始化配置字段（如果不存在，设置默认值）
      ensureFields(system.value, upgradeFields);
      ensureFields(system.value, commissionFields);
      ensureFields(system.value, discountFields);
      ensureFields(system.value, agentTeamRewardFields);
      initTeachingVisualDefaults(system.value);

      configBak.value = copyObj(system.value);
    })
    .catch((e) => {
      ElMessage.error("加载系统配置失败: " + e.message);
    });
  // 加载聊天配置
  httpGet("/api/admin/config/get?key=notice")
    .then((res) => {
      notice.value = res.data["content"];
    })
    .catch((e) => {
      ElMessage.error("公告信息失败: " + e.message);
    });

  httpGet("/api/admin/model/list")
    .then((res) => {
      models.value = res.data;
      loading.value = false;
    })
    .catch((e) => {
      ElMessage.error("获取模型失败：" + e.message);
    });

  httpGet("/api/admin/menu/list")
    .then((res) => {
      menus.value = res.data;
    })
    .catch((e) => {
      ElMessage.error("获取模型失败：" + e.message);
    });

  fetchLicense();
});

const fetchLicense = () => {
  httpGet("/api/admin/config/license")
    .then((res) => {
      license.value = res.data;
    })
    .catch((e) => {
      ElMessage.error("获取 License 失败：" + e.message);
    });
};

const rules = reactive({
  title: [{ required: true, message: "请输入网站标题", trigger: "blur" }],
  admin_title: [{ required: true, message: "请输入控制台标题", trigger: "blur" }],
  init_chat_calls: [{ required: true, message: "请输入赠送对话次数", trigger: "blur" }],
  user_img_calls: [{ required: true, message: "请输入赠送绘图次数", trigger: "blur" }],
});
const save = function (key) {
  if (key === "system") {
    systemFormRef.value.validate((valid) => {
      if (valid) {
        // 确保所有分销配置字段都被包含，即使值为 undefined 也设置为默认值
        const configToSave = { ...system.value };

        ensureFields(configToSave, upgradeFields);
        ensureFields(configToSave, commissionFields);
        ensureFields(configToSave, discountFields);
        ensureFields(configToSave, agentTeamRewardFields);
        initTeachingVisualDefaults(configToSave);

        httpPost("/api/admin/config/update", { key: key, config: configToSave, config_bak: configBak.value })
          .then(() => {
            ElMessage.success("操作成功！");
            // 保存成功后更新备份
            configBak.value = copyObj(configToSave);
          })
          .catch((e) => {
            ElMessage.error("操作失败：" + e.message);
          });
      }
    });
  } else if (key === "notice") {
    httpPost("/api/admin/config/update", { key: key, config: { content: notice.value, updated: true } })
      .then(() => {
        ElMessage.success("操作成功！");
      })
      .catch((e) => {
        ElMessage.error("操作失败：" + e.message);
      });
  }
};

// 激活授权
const licenseKey = ref("");
const active = () => {
  if (licenseKey.value === "") {
    return ElMessage.error("请输入授权码");
  }
  httpPost("/api/admin/config/active", { license: licenseKey.value })
    .then((res) => {
      ElMessage.success("授权成功，机器编码为：" + res.data);
      fetchLicense();
    })
    .catch((e) => {
      ElMessage.error(e.message);
    });
};

const configKey = ref("");
const beforeUpload = (key) => {
  configKey.value = key;
};

// 图片上传
const uploadImg = (file) => {
  // 压缩图片并上传
  new Compressor(file.file, {
    quality: 0.6,
    success(result) {
      const formData = new FormData();
      formData.append("file", result, result.name);
      // 执行上传操作
      httpPost("/api/admin/upload", formData)
        .then((res) => {
          system.value[configKey.value] = res.data.url;
          ElMessage.success("上传成功");
        })
        .catch((e) => {
          ElMessage.error("上传失败:" + e.message);
        });
    },
    error(e) {
      ElMessage.error("上传失败:" + e.message);
    },
  });
};

// 编辑期文件上传处理
const onUploadImg = (files, callback) => {
  Promise.all(
    files.map((file) => {
      return new Promise((rev, rej) => {
        const formData = new FormData();
        formData.append("file", file, file.name);
        // 执行上传操作
        httpPost("/api/admin/upload", formData)
          .then((res) => rev(res))
          .catch((error) => rej(error));
      });
    })
  )
    .then((res) => {
      ElMessage.success({ message: "上传成功", duration: 500 });
      callback(res.map((item) => item.data.url));
    })
    .catch((e) => {
      ElMessage.error("图片上传失败:" + e.message);
    });
};

const fixData = () => {
  ElMessageBox.confirm("在修复数据前，请先备份好数据库，以免数据丢失！是否继续操作?", "警告", {
    confirmButtonText: "确定",
    cancelButtonText: "取消",
    type: "warning",
  }).then(() => {
    loading.value = true;
    httpGet("/api/admin/config/fixData")
      .then(() => {
        ElMessage.success("数据修复成功");
        loading.value = false;
      })
      .catch((e) => {
        loading.value = false;
        ElMessage.error("数据修复失败：" + e.message);
      });
  });
};

// 导入视频数据
const importVideoData = () => {
  if (!videoJsonData.value || videoJsonData.value.trim() === "") {
    ElMessage.warning("请输入 JSON 数据");
    return;
  }

  // 验证 JSON 格式
  try {
    JSON.parse(videoJsonData.value);
  } catch (e) {
    ElMessage.error("JSON 格式错误，请检查数据格式");
    return;
  }

  importing.value = true;
  importResult.value = null;

  httpPost("/api/admin/video-square/import", {
    json_data: videoJsonData.value,
  })
    .then((res) => {
      importing.value = false;
      importResult.value = {
        type: "success",
        title: "导入成功",
        message: res.message || "数据导入成功！",
      };
      ElMessage.success("数据导入成功");
      // 可以选择清空输入框
      // videoJsonData.value = "";
    })
    .catch((e) => {
      importing.value = false;
      importResult.value = {
        type: "error",
        title: "导入失败",
        message: e.message || "数据导入失败，请检查 JSON 格式和网络连接",
      };
      ElMessage.error("数据导入失败：" + e.message);
    });
};
</script>

<style lang="stylus" scoped>
@import "@/assets/css/admin/form.styl"
@import "@/assets/css/main.styl"
.system-config {
  display flex
  justify-content center

  .sys-tabs {
    width 100%
    background-color var(--el-bg-color)
    padding 10px 20px 40px 20px
    //border: 1px solid var(--el-border-color);
  }
}
</style>
