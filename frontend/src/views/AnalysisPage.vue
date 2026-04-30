<template>
  <div class="analysis-page min-h-screen bg-gray-50 dark:bg-gray-900">
    <div class="container-modern py-8">
      <div class="page-header">
        <h1 class="page-title">数据分析</h1>
        <p class="page-subtitle">深度数据分析，助力科学决策</p>
      </div>

      <div class="coming-soon">
        <el-empty description="数据分析功能正在开发中，敬请期待">
          <el-button type="primary" class="back-btn" @click="goBack"
            >返回首页</el-button
          >
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
  min-height: calc(100vh - 160px);
}

.page-header {
  text-align: center;
  margin-bottom: 34px;
}

.page-title {
  font-size: 2rem;
  color: #0f172a;
  margin-bottom: 0.75rem;
  letter-spacing: -0.02em;
}

.page-subtitle {
  color: #475569;
  font-size: 1rem;
}

.coming-soon {
  background: linear-gradient(180deg, #fff 0%, #f8fafc 100%);
  border: 1px solid #e2e8f0;
  border-radius: 1rem;
  padding: 4rem 1.5rem;
  box-shadow: 0 10px 30px -24px rgba(15, 23, 42, 0.55);
}

.back-btn {
  min-width: 140px;
}

.dark .page-title {
  color: #f1f5f9;
}

.dark .page-subtitle {
  color: #94a3b8;
}

.dark .coming-soon {
  background: linear-gradient(180deg, #1f2937 0%, #111827 100%);
  border-color: #334155;
}

@media (max-width: 768px) {
  .page-title {
    font-size: 1.75rem;
  }

  .coming-soon {
    padding: 3rem 1rem;
  }
}
</style>
