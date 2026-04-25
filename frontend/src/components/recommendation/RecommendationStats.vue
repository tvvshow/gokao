<template>
  <div class="recommendation-stats">
    <el-row :gutter="20">
      <el-col :span="6">
        <div class="stat-item">
          <div class="stat-value">{{ totalCount }}</div>
          <div class="stat-label">推荐院校</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-item">
          <div class="stat-value">{{ successRate }}%</div>
          <div class="stat-label">预计成功率</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-item">
          <div class="stat-value">{{ riskLevelText }}</div>
          <div class="stat-label">风险等级</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-item">
          <div class="stat-value">{{ matchScore }}</div>
          <div class="stat-label">匹配度</div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { Recommendation } from '@/types/recommendation';

interface Props {
  recommendations: Recommendation[];
  riskTolerance: 'conservative' | 'moderate' | 'aggressive';
}

const props = defineProps<Props>();

// Total count of recommendations
const totalCount = computed(() => props.recommendations.length);

// Calculate average success rate
const successRate = computed(() => {
  if (props.recommendations.length === 0) return 0;
  const avgProbability =
    props.recommendations.reduce(
      (sum, rec) => sum + rec.admissionProbability,
      0
    ) / props.recommendations.length;
  return Math.round(avgProbability);
});

// Map risk tolerance to display text
const riskLevelText = computed(() => {
  const riskMap: Record<string, string> = {
    conservative: '低风险',
    moderate: '中风险',
    aggressive: '高风险',
  };
  return riskMap[props.riskTolerance] || '中风险';
});

// Calculate average match score
const matchScore = computed(() => {
  if (props.recommendations.length === 0) return 0;
  const avgMatch =
    props.recommendations.reduce((sum, rec) => sum + rec.matchScore, 0) /
    props.recommendations.length;
  return Math.round(avgMatch);
});
</script>

<style scoped>
.recommendation-stats {
  margin-bottom: 24px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #667eea;
  margin-bottom: 4px;
}

.stat-label {
  color: #7f8c8d;
  font-size: 14px;
}

@media (max-width: 768px) {
  .recommendation-stats {
    padding: 15px;
  }

  .stat-value {
    font-size: 18px;
  }

  .stat-label {
    font-size: 12px;
  }
}
</style>
