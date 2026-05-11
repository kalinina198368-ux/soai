<template>
  <div class="test-sora-page">
    <div class="header">
      <h1>Sora2 视频生成测试页面</h1>
      <p>这是一个简化的测试页面，用于验证Sora2功能</p>
    </div>
    
    <div class="content">
      <van-card>
        <template #header>
          <div class="card-header">
            <van-icon name="video-o" />
            <span>视频生成</span>
          </div>
        </template>
        
        <van-field
          v-model="prompt"
          type="textarea"
          label="视频描述"
          placeholder="请输入您想要生成的视频内容..."
          rows="4"
        />
        
        <div class="btn-container">
          <van-button 
            type="primary" 
            size="large" 
            block
            @click="generateVideo"
            :loading="isGenerating"
          >
            {{ isGenerating ? '生成中...' : '生成视频' }}
          </van-button>
        </div>
      </van-card>
      
      <div v-if="result" class="result">
        <h3>生成结果：</h3>
        <p>{{ result }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { showToast } from 'vant'

const prompt = ref('')
const isGenerating = ref(false)
const result = ref('')

const generateVideo = () => {
  if (!prompt.value.trim()) {
    showToast.fail('请输入视频描述')
    return
  }
  
  isGenerating.value = true
  
  // 模拟生成过程
  setTimeout(() => {
    result.value = `已生成视频：${prompt.value}`
    isGenerating.value = false
    showToast.success('视频生成成功！')
  }, 2000)
}
</script>

<style scoped>
.test-sora-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.header {
  text-align: center;
  color: white;
  margin-bottom: 30px;
}

.header h1 {
  font-size: 24px;
  margin-bottom: 10px;
}

.content {
  max-width: 600px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.btn-container {
  margin-top: 20px;
}

.result {
  margin-top: 20px;
  padding: 15px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.result h3 {
  margin: 0 0 10px 0;
  color: #333;
}

.result p {
  margin: 0;
  color: #666;
}
</style>