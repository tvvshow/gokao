<template>
  <el-card
    class="recommendation-card"
    shadow="hover"
    :aria-label="`推荐院校：${recommendation.university.name}，${getTypeLabel(recommendation.type)}，录取概率${recommendation.admissionProbability}%`"
  >
    <template #header>
      <div class="card-header">
        <div
          class="rank-badge"
          :class="recommendation.type"
          role="status"
          :aria-label="`排名第${index}，${getTypeLabel(recommendation.type)}`"
        >
          <span class="rank-number" aria-hidden="true">#{{ index }}</span>
          <span class="rank-type" aria-hidden="true">{{
            getTypeLabel(recommendation.type)
          }}</span>
        </div>
        <div
          class="probability"
          role="status"
          :aria-label="`录取概率${recommendation.admissionProbability}%`"
        >
          <el-progress
            type="circle"
            :percentage="recommendation.admissionProbability"
            :width="50"
            :stroke-width="6"
            :color="getProgressColor(recommendation.admissionProbability)"
            aria-hidden="true"
          />
          <span class="probability-text" aria-hidden="true">录取概率</span>
        </div>
      </div>
    </template>

    <div class="university-info">
      <div class="university-basic">
        <img
          :src="recommendation.university.logo || '/default-logo.svg'"
          :alt="`${recommendation.university.name}校徽`"
          class="university-logo"
        />
        <div class="university-details">
          <h3 class="university-name">{{ recommendation.university.name }}</h3>
          <div class="university-meta" role="list" aria-label="院校标签">
            <el-tag
              size="small"
              :type="getTagType(recommendation.university.level)"
              role="listitem"
            >
              {{ recommendation.university.level }}
            </el-tag>
            <el-tag size="small" type="info" role="listitem">{{
              recommendation.university.type
            }}</el-tag>
            <el-tag
              size="small"
              :type="getRiskTagType(recommendation.riskLevel)"
              role="listitem"
            >
              {{ getRiskLabel(recommendation.riskLevel) }}
            </el-tag>
          </div>
          <div class="location">
            <el-icon aria-hidden="true"><location /></el-icon>
            <span
              >{{ recommendation.university.province }}
              {{ recommendation.university.city }}</span
            >
          </div>
        </div>
      </div>

      <div class="recommendation-details">
        <div class="match-score">
          <span class="label">匹配度</span>
          <el-rate
            v-model="matchStars"
            disabled
            show-score
            text-color="#f59e0b"
            :max="5"
          />
        </div>

        <div class="score-info">
          <div class="score-item">
            <span class="label">历年分数线</span>
            <div class="score-range">
              <span class="min-score">{{ getMinScore() }}</span>
              <span class="separator">-</span>
              <span class="max-score">{{ getMaxScore() }}</span>
            </div>
          </div>
        </div>

        <div class="recommend-reason">
          <h4>推荐理由</h4>
          <p>{{ recommendation.recommendReason }}</p>
        </div>

        <div
          class="suggested-majors"
          v-if="recommendation.suggestedMajors.length > 0"
        >
          <h4>推荐专业</h4>
          <div class="majors-list">
            <el-tag
              v-for="major in recommendation.suggestedMajors.slice(0, 3)"
              :key="major.id"
              size="small"
              effect="plain"
            >
              {{ major.name }} ({{ major.probability }}%)
            </el-tag>
            <span
              v-if="recommendation.suggestedMajors.length > 3"
              class="more-majors"
            >
              +{{ recommendation.suggestedMajors.length - 3 }}个专业
            </span>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="card-actions" role="group" aria-label="院校操作">
        <el-button
          size="small"
          @click="$emit('view', recommendation.university.id)"
          :aria-label="`查看${recommendation.university.name}详情`"
        >
          查看详情
        </el-button>
        <el-button
          size="small"
          @click="$emit('compare', recommendation)"
          :aria-label="`将${recommendation.university.name}加入对比`"
        >
          加入对比
        </el-button>
        <el-button
          size="small"
          :type="recommendation.university.isFavorite ? 'danger' : 'default'"
          @click="$emit('favorite', recommendation)"
          :aria-label="
            recommendation.university.isFavorite
              ? `取消收藏${recommendation.university.name}`
              : `收藏${recommendation.university.name}`
          "
        >
          {{ recommendation.university.isFavorite ? '已收藏' : '收藏' }}
        </el-button>
      </div>
    </template>
  </el-card>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Location } from '@element-plus/icons-vue';
import type { Recommendation } from '@/types/recommendation';

interface Props {
  recommendation: Recommendation;
  index: number;
}

const props = defineProps<Props>();

// 计算匹配度星级
const matchStars = computed(() => {
  return Math.round(props.recommendation.matchScore / 20); // 转换为1-5星
});

// 获取类型标签
const getTypeLabel = (type: string) => {
  const labels = {
    aggressive: '冲一冲',
    moderate: '稳一稳',
    conservative: '保一保',
  };
  return labels[type as keyof typeof labels] || type;
};

// 获取进度条颜色
const getProgressColor = (percentage: number) => {
  if (percentage >= 80) return '#22c55e';
  if (percentage >= 60) return '#f59e0b';
  if (percentage >= 40) return '#ef4444';
  return '#94a3b8';
};

// 获取标签类型
const getTagType = (level: string) => {
  const types: Record<string, string> = {
    '985工程': 'danger',
    '211工程': 'warning',
    双一流: 'success',
    普通本科: 'info',
  };
  return types[level] || 'info';
};

// 获取风险标签类型
const getRiskTagType = (risk: string) => {
  const types = {
    low: 'success',
    medium: 'warning',
    high: 'danger',
  };
  return types[risk as keyof typeof types] || 'info';
};

// 获取风险标签
const getRiskLabel = (risk: string) => {
  const labels = {
    low: '低风险',
    medium: '中风险',
    high: '高风险',
  };
  return labels[risk as keyof typeof labels] || risk;
};

// 获取最低分数
const getMinScore = () => {
  if (props.recommendation.historicalData.length === 0) return '--';
  return Math.min(
    ...props.recommendation.historicalData.map((d) => d.minScore)
  );
};

// 获取最高分数
const getMaxScore = () => {
  if (props.recommendation.historicalData.length === 0) return '--';
  return Math.max(
    ...props.recommendation.historicalData.map((d) => d.maxScore)
  );
};
</script>

<style scoped>
.recommendation-card {
  margin-bottom: 18px;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid rgb(148 163 184 / 0.24);
  box-shadow: 0 18px 36px -30px rgb(15 23 42 / 0.55);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0;
}

.rank-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 999px;
  font-weight: 600;
}

.rank-badge.aggressive {
  background: #fff2e8;
  color: #f59e0b;
}

.rank-badge.moderate {
  background: #f0f9ff;
  color: #0ea5e9;
}

.rank-badge.conservative {
  background: #f0f9f0;
  color: #22c55e;
}

.rank-number {
  font-size: 18px;
}

.rank-type {
  font-size: 12px;
}

.probability {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.probability-text {
  font-size: 11px;
  letter-spacing: 0.03em;
  color: #475569;
}

.university-info {
  padding: 2px 0;
}

.university-basic {
  display: flex;
  gap: 13px;
  margin-bottom: 14px;
}

.university-logo {
  width: 64px;
  height: 64px;
  border-radius: 10px;
  object-fit: cover;
  border: 1px solid #e2e8f0;
}

.university-details {
  flex: 1;
}

.university-name {
  font-size: 19px;
  font-weight: 600;
  color: #0f172a;
  margin: 0 0 8px 0;
  letter-spacing: -0.01em;
}

.university-meta {
  display: flex;
  gap: 6px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.location {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #64748b;
  font-size: 13px;
}

.recommendation-details {
  space-y: 16px;
}

.match-score {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.match-score .label {
  font-weight: 500;
  color: #0f172a;
}

.score-info {
  background: #f8fafc;
  padding: 13px;
  border-radius: 10px;
  margin-bottom: 16px;
  border: 1px solid rgb(148 163 184 / 0.2);
}

.score-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.score-range {
  font-weight: 600;
  color: #ef4444;
}

.separator {
  margin: 0 4px;
  color: #94a3b8;
}

.recommend-reason,
.suggested-majors {
  margin-bottom: 16px;
}

.recommend-reason h4,
.suggested-majors h4 {
  font-size: 14px;
  color: #0f172a;
  margin: 0 0 8px 0;
}

.recommend-reason p {
  color: #64748b;
  font-size: 14px;
  line-height: 1.6;
  margin: 0;
}

.majors-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.more-majors {
  font-size: 12px;
  color: #94a3b8;
  background: #f1f5f9;
  padding: 2px 6px;
  border-radius: 999px;
}

.card-actions {
  display: flex;
  justify-content: space-between;
  padding: 0;
  gap: 8px;
}

.card-actions .el-button {
  flex: 1;
  margin: 0;
  border-radius: 10px;
  min-height: 38px;
}
</style>
