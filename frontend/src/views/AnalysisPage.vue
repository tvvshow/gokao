<template>
  <div class="analysis-page">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">数据分析</h1>
        <p class="page-subtitle">深度数据分析，助力科学决策</p>
      </div>

      <!-- 分析类型选择 -->
      <div class="analysis-tabs">
        <el-tabs v-model="activeTab" @tab-click="handleTabChange">
          <el-tab-pane label="录取概率预测" name="probability">
            <ProbabilityAnalysis />
          </el-tab-pane>
          <el-tab-pane label="历史分数线趋势" name="trend">
            <TrendAnalysis />
          </el-tab-pane>
          <el-tab-pane label="专业就业分析" name="employment">
            <EmploymentAnalysis />
          </el-tab-pane>
          <el-tab-pane label="院校对比分析" name="comparison">
            <ComparisonAnalysis />
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ProbabilityAnalysis from '@/components/analysis/ProbabilityAnalysis.vue'
import TrendAnalysis from '@/components/analysis/TrendAnalysis.vue'
import EmploymentAnalysis from '@/components/analysis/EmploymentAnalysis.vue'
import ComparisonAnalysis from '@/components/analysis/ComparisonAnalysis.vue'

const activeTab = ref('probability')

const handleTabChange = (tab: any) => {
  activeTab.value = tab.paneName
}
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

.analysis-tabs {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
</style>