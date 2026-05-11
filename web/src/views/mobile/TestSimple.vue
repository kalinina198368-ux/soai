<template>
  <div class="test-page">
    <h1>Sora2 视频生成测试</h1>
    
    <!-- 测试基本组件 -->
    <van-button type="primary" @click="testClick">测试按钮</van-button>
    
    <van-card>
      <template #header>
        <div>测试卡片</div>
      </template>
      <div>卡片内容</div>
    </van-card>
    
    <!-- 测试输入框 -->
    <van-field
      v-model="testInput"
      label="测试输入"
      placeholder="请输入内容"
    />
    
    <!-- 测试进度条 -->
    <van-progress :percentage="50" />
    
    <!-- 测试图标 -->
    <van-icon name="video-o" />
    
    <p>输入内容: {{ testInput }}</p>
  </div>
</template>

<script setup>
import { ref } from "vue";

const testInput = ref("");

const testClick = () => {
  console.log("按钮被点击了");
  alert("按钮工作正常！");
};
</script>

<style scoped>
.test-page {
  padding: 20px;
  background: #f5f5f5;
  min-height: 100vh;
}

h1 {
  color: #333;
  text-align: center;
  margin-bottom: 20px;
}

.van-button {
  margin: 10px 0;
}

.van-card {
  margin: 10px 0;
}

.van-field {
  margin: 10px 0;
}

.van-progress {
  margin: 10px 0;
}

.van-icon {
  font-size: 24px;
  color: #1989fa;
  margin: 10px;
}
</style>
