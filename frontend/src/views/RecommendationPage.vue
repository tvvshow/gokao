<template>
  <div class="recommendation-page">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">{{ pageTitle }}</h1>
        <p class="page-subtitle">
          {{ pageSubtitle }}
        </p>
      </div>

      <div class="recommendation-content">
        <!-- 左侧：信息录入 -->
        <div class="input-section">
          <StudentInfoForm
            ref="studentFormRef"
            :student-info="studentForm"
            :loading="generating"
            @submit="handleRecommend"
            @reset="handleReset"
          />
        </div>

        <!-- 右侧：推荐结果 -->
        <div class="result-section">
          <RecommendationResults
            :recommendations="recommendations"
            :loading="generating"
            :progress="progressPercentage"
            :risk-tolerance="studentForm.preferences.riskTolerance"
            @export="exportRecommendations"
            @save="saveRecommendations"
            @view="viewUniversityDetail"
            @compare="addToCompare"
            @favorite="toggleFavorite"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  StudentInfoForm,
  RecommendationResults,
} from '@/components/recommendation';
import { useRecommendationStore } from '@/stores/recommendation';
import { recommendationApi } from '@/api/recommendation';
import type { StudentInfo, Recommendation } from '@/types/recommendation';

const router = useRouter();
const route = useRoute();
const recommendationStore = useRecommendationStore();

// 根据路由名称动态显示标题
const pageTitle = computed(() =>
  route.name === 'Simulation' ? '模拟填报' : 'AI智能推荐'
);
const pageSubtitle = computed(() =>
  route.name === 'Simulation'
    ? '真实模拟填报环境，提前体验志愿填报流程'
    : '基于大数据分析和AI算法，为您量身定制最优志愿方案'
);
const studentFormRef = ref<InstanceType<typeof StudentInfoForm>>();

// Form data
const studentForm = reactive<StudentInfo>({
  score: null,
  province: '',
  scienceType: '理科',
  year: new Date().getFullYear(),
  rank: null,
  preferences: {
    regions: [],
    majorCategories: [],
    universityTypes: [],
    riskTolerance: 'moderate',
    specialRequirements: '',
  },
});

// State
const generating = ref(false);
const progressPercentage = ref(0);

// Computed
const recommendations = computed(() => recommendationStore.recommendations);

// Generate recommendations
const handleRecommend = async () => {
  generating.value = true;
  progressPercentage.value = 0;

  // Simulate progress
  const progressInterval = setInterval(() => {
    progressPercentage.value += Math.random() * 15;
    if (progressPercentage.value >= 90) {
      clearInterval(progressInterval);
    }
  }, 200);

  try {
    await recommendationStore.generateRecommendations(studentForm);
    progressPercentage.value = 100;
    ElMessage.success('推荐生成成功');
  } catch (error) {
    ElMessage.error(
      error instanceof Error ? error.message : '推荐生成失败，请稍后重试'
    );
  } finally {
    clearInterval(progressInterval);
    setTimeout(() => {
      generating.value = false;
      progressPercentage.value = 0;
    }, 500);
  }
};

// Reset form
const handleReset = () => {
  recommendationStore.clearRecommendations();
};

// View university detail
const viewUniversityDetail = (universityId: string) => {
  router.push(`/universities/${universityId}`);
};

// Add to compare
const addToCompare = (recommendation: Recommendation) => {
  ElMessage.success(`已添加 ${recommendation.university.name} 到对比列表`);
};

// Toggle favorite
const toggleFavorite = (recommendation: Recommendation) => {
  recommendation.university.isFavorite = !recommendation.university.isFavorite;
  ElMessage.success(
    recommendation.university.isFavorite ? '已收藏' : '已取消收藏'
  );
};

// Export recommendations report
const exportRecommendations = async () => {
  try {
    const blob = await recommendationApi.exportReport(recommendations.value);
    // 直接返回Blob，创建下载链接
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `志愿推荐报告_${studentForm.score}分_${new Date().toLocaleDateString().replace(/\//g, '-')}.pdf`;
    link.click();
    window.URL.revokeObjectURL(url);
    ElMessage.success('报告导出成功');
  } catch {
    ElMessage.error('导出失败，请稍后重试');
  }
};

// Save recommendations scheme
const saveRecommendations = async () => {
  try {
    const { value: schemeName } = await ElMessageBox.prompt(
      '请输入方案名称',
      '保存方案',
      {
        confirmButtonText: '保存',
        cancelButtonText: '取消',
        inputValue: `${studentForm.score}分志愿方案`,
        inputValidator: (value) => {
          if (!value?.trim()) {
            return '请输入方案名称';
          }
          return true;
        },
      }
    );

    const response = await recommendationApi.saveScheme({
      name: schemeName,
      studentInfo: studentForm,
      recommendations: recommendations.value,
    });

    if (response.success) {
      ElMessage.success('方案保存成功');
    }
  } catch {
    // User cancelled
  }
};

onMounted(() => {
  // Restore data from route query
  const query = router.currentRoute.value.query;
  if (query.score) {
    studentForm.score = Number(query.score);
  }
  if (query.province) {
    studentForm.province = query.province as string;
  }
});
</script>

<style scoped>
.recommendation-page {
  padding: 20px 0;
  min-height: calc(100vh - 160px);
}

.container {
  max-width: 1400px;
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

.recommendation-content {
  display: grid;
  grid-template-columns: minmax(450px, 500px) 1fr;
  gap: 30px;
  align-items: start;
}

.input-section {
  position: sticky;
  top: 100px;
}

.result-section {
  min-height: 600px;
}

/* Responsive design */
@media (max-width: 1200px) {
  .recommendation-content {
    grid-template-columns: 1fr;
    gap: 20px;
  }

  .input-section {
    position: static;
  }
}

@media (max-width: 768px) {
  .container {
    padding: 0 10px;
  }

  .page-title {
    font-size: 24px;
  }
}
</style>
