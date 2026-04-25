<template>
  <el-card
    class="content-card"
    v-if="recommendations.length > 0"
    aria-labelledby="results-heading"
  >
    <template #header>
      <div class="card-header">
        <el-icon aria-hidden="true"><trophy /></el-icon>
        <span id="results-heading">推荐结果</span>
        <div class="result-actions" role="group" aria-label="结果操作">
          <el-button
            size="small"
            @click="handleExport"
            aria-label="导出推荐报告为PDF"
          >
            <el-icon aria-hidden="true"><download /></el-icon>
            导出报告
          </el-button>
          <el-button
            size="small"
            type="primary"
            @click="handleSave"
            aria-label="保存当前推荐方案"
          >
            <el-icon aria-hidden="true"><collection /></el-icon>
            保存方案
          </el-button>
        </div>
      </div>
    </template>

    <!-- 推荐统计 -->
    <RecommendationStats
      :recommendations="recommendations"
      :risk-tolerance="riskTolerance"
    />

    <!-- 分类标签 -->
    <div class="category-tabs" role="tablist" aria-label="推荐分类">
      <el-tabs v-model="activeCategory" @tab-click="handleCategoryChange">
        <el-tab-pane
          label="冲一冲"
          name="aggressive"
          :aria-label="`冲一冲，共${aggressiveCount}所院校`"
        >
          <el-badge
            :value="aggressiveCount"
            class="tab-badge"
            aria-hidden="true"
          />
        </el-tab-pane>
        <el-tab-pane
          label="稳一稳"
          name="moderate"
          :aria-label="`稳一稳，共${moderateCount}所院校`"
        >
          <el-badge
            :value="moderateCount"
            class="tab-badge"
            aria-hidden="true"
          />
        </el-tab-pane>
        <el-tab-pane
          label="保一保"
          name="conservative"
          :aria-label="`保一保，共${conservativeCount}所院校`"
        >
          <el-badge
            :value="conservativeCount"
            class="tab-badge"
            aria-hidden="true"
          />
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 推荐列表 - 使用虚拟滚动优化大列表性能 -->
    <VirtualList
      v-if="currentRecommendations.length > 100"
      :items="currentRecommendations"
      :item-height="280"
      container-height="600px"
      key-field="id"
      aria-label="推荐院校列表"
    >
      <template #default="{ item, index }">
        <RecommendationCard
          :recommendation="item"
          :index="index + 1"
          @view="handleView"
          @compare="handleCompare"
          @favorite="handleFavorite"
        />
      </template>
    </VirtualList>

    <!-- 少量数据时使用普通列表 -->
    <div
      v-else
      class="recommendations-list"
      role="list"
      aria-label="推荐院校列表"
    >
      <RecommendationCard
        v-for="(recommendation, index) in currentRecommendations"
        :key="recommendation.id"
        :recommendation="recommendation"
        :index="index + 1"
        role="listitem"
        @view="handleView"
        @compare="handleCompare"
        @favorite="handleFavorite"
      />
    </div>
  </el-card>

  <!-- 空状态 -->
  <el-card
    class="content-card"
    v-else-if="!loading"
    aria-labelledby="empty-heading"
  >
    <div class="empty-state" role="status">
      <el-icon size="80" aria-hidden="true"><magic-stick /></el-icon>
      <h3 id="empty-heading">开始您的志愿推荐</h3>
      <p>填写左侧信息，获取AI智能推荐的志愿方案</p>
    </div>
  </el-card>

  <!-- 加载状态 -->
  <el-card class="content-card" v-else aria-labelledby="loading-heading">
    <div
      class="loading-state"
      role="status"
      aria-live="polite"
      aria-busy="true"
    >
      <el-icon class="rotating" size="60" aria-hidden="true"
        ><loading
      /></el-icon>
      <h3 id="loading-heading">AI正在分析中...</h3>
      <p>正在基于您的信息进行智能匹配，请稍候</p>
      <el-progress
        :percentage="progress"
        :stroke-width="8"
        :aria-label="`分析进度${progress}%`"
      />
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import {
  Trophy,
  Download,
  Collection,
  MagicStick,
  Loading,
} from '@element-plus/icons-vue';
import RecommendationCard from '@/components/RecommendationCard.vue';
import RecommendationStats from './RecommendationStats.vue';
import { VirtualList } from '@/components/common';
import type { Recommendation } from '@/types/recommendation';

interface Props {
  recommendations: Recommendation[];
  loading?: boolean;
  progress?: number;
  riskTolerance: 'conservative' | 'moderate' | 'aggressive';
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  progress: 0,
});

const emit = defineEmits<{
  export: [];
  save: [];
  view: [universityId: string];
  compare: [recommendation: Recommendation];
  favorite: [recommendation: Recommendation];
}>();

const activeCategory = ref('moderate');

// Filter recommendations by type
const getRecommendationsByType = (type: string) => {
  return props.recommendations.filter((rec) => rec.type === type);
};

// Computed counts for each category
const aggressiveCount = computed(
  () => getRecommendationsByType('aggressive').length
);
const moderateCount = computed(
  () => getRecommendationsByType('moderate').length
);
const conservativeCount = computed(
  () => getRecommendationsByType('conservative').length
);

// Current recommendations based on active category
const currentRecommendations = computed(() => {
  return getRecommendationsByType(activeCategory.value);
});

// Handle category tab change
const handleCategoryChange = (tab: { paneName: string }) => {
  activeCategory.value = tab.paneName;
};

// Event handlers
const handleExport = () => emit('export');
const handleSave = () => emit('save');
const handleView = (universityId: string) => emit('view', universityId);
const handleCompare = (recommendation: Recommendation) =>
  emit('compare', recommendation);
const handleFavorite = (recommendation: Recommendation) =>
  emit('favorite', recommendation);
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  font-weight: 600;
  color: #2c3e50;
}

.card-header .el-icon {
  margin-right: 8px;
  color: #667eea;
}

.result-actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

.category-tabs {
  margin-bottom: 24px;
}

.tab-badge {
  margin-left: 8px;
}

.recommendations-list {
  max-height: 600px;
  overflow-y: auto;
}

.empty-state,
.loading-state {
  text-align: center;
  padding: 60px 20px;
  color: #7f8c8d;
}

.empty-state .el-icon,
.loading-state .el-icon {
  color: #667eea;
  margin-bottom: 20px;
}

.empty-state h3,
.loading-state h3 {
  margin-bottom: 12px;
  color: #2c3e50;
}

.loading-state .el-progress {
  margin-top: 20px;
  max-width: 300px;
  margin-left: auto;
  margin-right: auto;
}

.rotating {
  animation: rotate 2s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
