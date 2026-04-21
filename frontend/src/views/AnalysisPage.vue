<template>
  <div class="analysis-page">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">数据分析</h1>
        <p class="page-subtitle">深度数据分析，助力科学决策</p>
      </div>

      <div class="coming-soon">
        <el-empty description="数据分析功能正在开发中，敬请期待">
          <el-button type="primary" @click="goBack">返回首页</el-button>
        </el-empty>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';

const router = useRouter();

const goBack = () => {
  router.push('/');
};
</script>

<style scoped>
.analysis-page {
  padding: 20px 0;
  min-height: calc(100vh - 160px);
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.page-header {
  text-align: center;
  margin-bottom: 40px;
}

.page-title {
  font-size: 32px;
  color: #2c3e50;
  margin-bottom: 12px;
}

.page-subtitle {
  color: #7f8c8d;
  font-size: 16px;
}

.coming-soon {
  background: white;
  border-radius: 8px;
  padding: 60px 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
</style>
