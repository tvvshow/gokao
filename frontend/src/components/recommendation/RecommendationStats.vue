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
  border: 1px solid #e2e8f0;
  background: linear-gradient(180deg, #fff 0%, #f8fafc 100%);
  border-radius: 0.875rem;
  box-shadow: 0 10px 30px -24px rgba(15, 23, 42, 0.55);
}

.stat-item {
  text-align: center;
  padding: 0.5rem 0.25rem;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #0ea5e9;
  margin-bottom: 4px;
}

.stat-label {
  color: #64748b;
  font-size: 14px;
}

.dark .recommendation-stats {
  border-color: #334155;
  background: linear-gradient(180deg, #1f2937 0%, #111827 100%);
}

.dark .stat-value {
  color: #67e8f9;
}

.dark .stat-label {
  color: #94a3b8;
}

@media (max-width: 768px) {
  .recommendation-stats {
    padding: 15px;
  }

  .recommendation-stats :deep(.el-col) {
    margin-bottom: 0.625rem;
  }

  .stat-value {
    font-size: 18px;
  }

  .stat-label {
    font-size: 12px;
  }
}
</style>
