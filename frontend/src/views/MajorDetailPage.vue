<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <div class="container-modern py-8">
      <!-- 顶部导航栏 -->
      <div class="flex items-center justify-between mb-6">
        <button class="btn btn-secondary" @click="goBack">
          <ArrowLeftIcon class="w-4 h-4 mr-2" />
          返回列表
        </button>
        <div class="flex space-x-2">
          <button
            class="btn btn-outline"
            :class="{ 'btn-active': isFavorite }"
            @click="toggleFavorite"
            title="收藏专业"
          >
            <HeartIcon
              class="w-4 h-4"
              :class="{ 'fill-current text-red-500': isFavorite }"
            />
            {{ isFavorite ? '已收藏' : '收藏' }}
          </button>
          <button class="btn btn-primary" @click="addToCompare">
            <ScaleIcon class="w-4 h-4 mr-2" />
            加入对比
          </button>
        </div>
      </div>

      <!-- 加载态 -->
      <div v-if="loading" class="loading-container">
        <div class="loading-spinner mx-auto mb-4"></div>
        <p class="text-gray-500 dark:text-gray-400">加载专业详情中...</p>
      </div>

      <!-- 错误态 -->
      <div v-else-if="error" class="error-container">
        <div class="text-red-500 mb-4">{{ error }}</div>
        <button class="btn btn-primary" @click="fetchDetail">
          <RefreshCwIcon class="w-4 h-4 mr-2" />
          重试
        </button>
      </div>

      <!-- 空态 -->
      <div v-else-if="!major" class="empty-container">
        <p class="text-gray-500 dark:text-gray-400 mb-4">未找到专业信息</p>
        <button class="btn btn-primary" @click="goBack">返回列表</button>
      </div>

      <!-- 内容区域 -->
      <template v-else>
        <!-- 专业头部卡片 -->
        <div class="header-card">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-3 mb-3">
                <h1 class="page-title mb-0">{{ major.name }}</h1>
                <span
                  v-if="major.isPopular"
                  class="badge badge-popular"
                >
                  <TrendingUpIcon class="w-3 h-3" />
                  热门
                </span>
              </div>
              <p v-if="major.description" class="text-gray-200">
                {{ major.description }}
              </p>
              <div class="flex items-center gap-4 mt-3 text-sm text-gray-300">
                <span v-if="major.code" class="flex items-center">
                  <HashIcon class="w-4 h-4 mr-1" />
                  专业代码: {{ major.code }}
                </span>
                <span v-if="major.category" class="flex items-center">
                  <FolderIcon class="w-4 h-4 mr-1" />
                  {{ major.category }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- 主要内容网格 -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mt-6">
          <!-- 左侧内容 (占2列) -->
          <div class="lg:col-span-2 space-y-6">
            <!-- 基础信息 -->
            <div class="card">
              <h2 class="section-title">
                <InfoIcon class="w-5 h-5 mr-2" />
                基础信息
              </h2>
              <div class="info-grid">
                <div class="info-row">
                  <span class="label">所属院校</span>
                  <span class="value">{{ majorUniversityName || '-' }}</span>
                </div>
                <div class="info-row">
                  <span class="label">专业类别</span>
                  <span class="value">{{ major.category || '-' }}</span>
                </div>
                <div class="info-row">
                  <span class="label">学位类型</span>
                  <span class="value">{{ major.degree || '-' }}</span>
                </div>
                <div class="info-row">
                  <span class="label">学制年限</span>
                  <span class="value">{{
                    major.duration ? `${major.duration}年` : '-'
                  }}</span>
                </div>
                <div class="info-row">
                  <span class="label">是否热门</span>
                  <span class="value">{{
                    major.isPopular ? '是' : '否'
                  }}</span>
                </div>
              </div>
            </div>

            <!-- 就业与薪酬 -->
            <div
              v-if="hasEmploymentData"
              class="card employment-card"
            >
              <h2 class="section-title">
                <BriefcaseIcon class="w-5 h-5 mr-2" />
                就业与薪酬
              </h2>
              <div class="employment-grid">
                <div class="employment-item">
                  <div class="employment-icon employment-rate">
                    <TrendingUpIcon class="w-6 h-6" />
                  </div>
                  <div class="employment-content">
                    <div class="employment-value">
                      {{
                        major.employmentRate
                          ? `${major.employmentRate}%`
                          : '-'
                      }}
                    </div>
                    <div class="employment-label">就业率</div>
                    <div
                      v-if="major.employmentRate"
                      class="employment-bar"
                    >
                      <div
                        class="employment-bar-fill"
                        :style="{ width: `${major.employmentRate}%` }"
                      ></div>
                    </div>
                  </div>
                </div>
                <div class="employment-item">
                  <div class="employment-icon employment-salary">
                    <DollarSignIcon class="w-6 h-6" />
                  </div>
                  <div class="employment-content">
                    <div class="employment-value">
                      {{
                        major.averageSalary
                          ? `${major.averageSalary}K`
                          : '-'
                      }}
                    </div>
                    <div class="employment-label">平均年薪</div>
                    <div class="employment-desc">
                      {{ major.averageSalary ? '毕业首年' : '-' }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 专业介绍 -->
            <div v-if="major.description" class="card">
              <h2 class="section-title">
                <BookOpenIcon class="w-5 h-5 mr-2" />
                专业介绍
              </h2>
              <div class="prose prose-sm max-w-none">
                <p>{{ major.description }}</p>
              </div>
            </div>

            <!-- 相关专业推荐 -->
            <div class="card">
              <div class="flex items-center justify-between mb-4">
                <h2 class="section-title mb-0">
                  <SparklesIcon class="w-5 h-5 mr-2" />
                  相关专业
                </h2>
              </div>
              <div class="related-majors">
                <router-link
                  v-for="relatedMajor in relatedMajors"
                  :key="relatedMajor.id"
                  :to="`/majors/${relatedMajor.id}`"
                  class="related-major-card"
                >
                  <div class="related-major-name">
                    {{ relatedMajor.name }}
                  </div>
                  <div class="related-major-meta">
                    {{ relatedMajor.category }}
                  </div>
                </router-link>
              </div>
            </div>
          </div>

          <!-- 右侧边栏 -->
          <div class="space-y-6">
            <!-- 快速统计 -->
            <div class="card stats-card">
              <h2 class="section-title">专业数据</h2>
              <div class="quick-stats">
                <div class="quick-stat">
                  <div class="stat-icon stat-degree">
                    <GraduationCapIcon class="w-5 h-5" />
                  </div>
                  <div class="stat-content">
                    <div class="stat-value">
                      {{ major.degree || '-' }}
                    </div>
                    <div class="stat-label">学位类型</div>
                  </div>
                </div>
                <div class="quick-stat">
                  <div class="stat-icon stat-duration">
                    <ClockIcon class="w-5 h-5" />
                  </div>
                  <div class="stat-content">
                    <div class="stat-value">
                      {{ major.duration ? `${major.duration}年` : '-' }}
                    </div>
                    <div class="stat-label">学制</div>
                  </div>
                </div>
                <div
                  v-if="major.employmentRate"
                  class="quick-stat"
                >
                  <div class="stat-icon stat-employment">
                    <BriefcaseIcon class="w-5 h-5" />
                  </div>
                  <div class="stat-content">
                    <div class="stat-value">
                      {{ major.employmentRate }}%
                    </div>
                    <div class="stat-label">就业率</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 开设院校 -->
            <div class="card">
              <h2 class="section-title">开设院校</h2>
              <div class="universities-list">
                <div
                  v-for="uni in offeringUniversities"
                  :key="uni.id"
                  class="university-item"
                >
                  <div class="university-item-content">
                    <div class="university-item-name">
                      {{ uni.name }}
                    </div>
                    <div class="university-item-meta">
                      {{ uni.province }} · {{ uni.level }}
                    </div>
                  </div>
                  <router-link
                    :to="`/universities/${uni.id}`"
                    class="university-item-link"
                  >
                    <ArrowRightIcon class="w-4 h-4" />
                  </router-link>
                </div>
              </div>
            </div>

            <!-- 操作按钮 -->
            <div class="card action-card">
              <h2 class="section-title">快捷操作</h2>
              <div class="action-buttons">
                <button class="action-btn" @click="addToCompare">
                  <ScaleIcon class="w-5 h-5" />
                  <span>添加到对比</span>
                </button>
                <button class="action-btn" @click="viewMajorsGuide">
                  <BookOpenIcon class="w-5 h-5" />
                  <span>报考指南</span>
                </button>
                <button class="action-btn" @click="shareMajor">
                  <ShareIcon class="w-5 h-5" />
                  <span>分享专业</span>
                </button>
              </div>
            </div>

            <!-- 报考提示 -->
            <div class="card tip-card">
              <h2 class="section-title">
                <LightbulbIcon class="w-5 h-5 mr-2" />
                报考提示
              </h2>
              <div class="tip-content">
                <p>结合个人兴趣与职业规划选择专业，更多数据将持续补充。</p>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { api } from '@/api/api-client';
import { ElMessage } from 'element-plus';
import {
  ArrowLeftIcon,
  HeartIcon,
  ScaleIcon,
  TrendingUpIcon,
  HashIcon,
  FolderIcon,
  InfoIcon,
  BriefcaseIcon,
  DollarSignIcon,
  BookOpenIcon,
  SparklesIcon,
  GraduationCapIcon,
  ClockIcon,
  ShareIcon,
  ArrowRightIcon,
  RefreshCwIcon,
  LightbulbIcon,
} from 'lucide-vue-next';
import type { Major } from '@/types/university';

const route = useRoute();
const router = useRouter();

const major = ref<Major & { university?: { name?: string; id?: string } } | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);
const isFavorite = ref(false);

// 模拟相关专业数据
const relatedMajors = computed(() => {
  if (!major.value) return [];
  const category = major.value.category;
  const mockMajors: Major[] = [
    { id: '1', name: '计算机科学与技术', code: '080901', category, degree: '工学学士', duration: 4 },
    { id: '2', name: '软件工程', code: '080902', category, degree: '工学学士', duration: 4 },
    { id: '3', name: '网络工程', code: '080903', category, degree: '工学学士', duration: 4 },
    { id: '4', name: '信息安全', code: '080904', category, degree: '工学学士', duration: 4 },
  ];
  return mockMajors.filter(m => m.category === category).slice(0, 4);
});

// 模拟开设院校数据
const offeringUniversities = computed(() => {
  const mockUnis = [
    { id: '1', name: '清华大学', province: '北京', level: '985' },
    { id: '2', name: '北京大学', province: '北京', level: '985' },
    { id: '3', name: '浙江大学', province: '浙江', level: '985' },
  ];
  return mockUnis;
});

const majorUniversityName = computed(() => {
  const data = major.value;
  return data?.university?.name || '';
});

const hasEmploymentData = computed(() => {
  const m = major.value;
  return m && (m.employmentRate || m.averageSalary);
});

const fetchDetail = async () => {
  const id = route.params.id as string | undefined;
  if (!id) {
    error.value = '专业ID缺失';
    major.value = null;
    return;
  }
  loading.value = true;
  error.value = null;
  try {
    const response = await api.get<Major>(`/api/v1/data/majors/${id}`);
    if (response.success) {
      major.value = response.data;
      isFavorite.value = false;
    } else {
      major.value = null;
      error.value = response.message || '获取专业详情失败';
    }
  } catch (err) {
    major.value = null;
    error.value = err instanceof Error ? err.message : '获取专业详情失败';
  } finally {
    loading.value = false;
  }
};

const goBack = () => {
  router.push('/majors');
};

const toggleFavorite = () => {
  isFavorite.value = !isFavorite.value;
  ElMessage.success(isFavorite.value ? '已添加到收藏' : '已取消收藏');
};

const addToCompare = () => {
  if (!major.value) return;
  ElMessage.success(`已添加 ${major.value.name} 到对比列表`);
};

const viewMajorsGuide = () => {
  ElMessage.info('报考指南功能开发中');
};

const shareMajor = () => {
  const url = window.location.href;
  navigator.clipboard?.writeText(url);
  ElMessage.success('链接已复制到剪贴板');
};

onMounted(fetchDetail);
watch(() => route.params.id, fetchDetail);
</script>

<style scoped>
/* 页面头部 - 使用绿色渐变 */
.header-card {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  padding: 2rem;
  border-radius: 1rem;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1);
}

.header-card .page-title {
  color: white;
  margin: 0;
}

.header-card .text-gray-200 {
  color: rgba(255, 255, 255, 0.9);
}

.header-card .text-gray-300 {
  color: rgba(255, 255, 255, 0.8);
}

/* 徽章样式 */
.badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 600;
}

.badge-popular {
  background: rgba(255, 255, 255, 0.2);
  color: #fef08a;
}

/* 信息网格 */
.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 0;
  border-bottom: 1px dashed rgba(255, 255, 255, 0.2);
}

.info-row:last-child {
  border-bottom: none;
}

.label {
  color: #6b7280;
  font-size: 0.875rem;
}

.dark .label {
  color: #9ca3af;
}

.value {
  color: #1f2937;
  font-weight: 500;
}

.dark .value {
  color: #f3f4f6;
}

/* 就业数据卡片 */
.employment-card {
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
}

.dark .employment-card {
  background: linear-gradient(135deg, #14532d 0%, #166534 100%);
}

.employment-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.employment-item {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
}

.employment-icon {
  width: 50px;
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.75rem;
  color: white;
  flex-shrink: 0;
}

.employment-rate {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
}

.employment-salary {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
}

.employment-content {
  flex: 1;
}

.employment-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1f2937;
}

.dark .employment-value {
  color: #f3f4f6;
}

.employment-label {
  font-size: 0.875rem;
  color: #6b7280;
  margin-top: 0.25rem;
}

.dark .employment-label {
  color: #9ca3af;
}

.employment-bar {
  height: 6px;
  background: #e5e7eb;
  border-radius: 3px;
  margin-top: 0.5rem;
  overflow: hidden;
}

.dark .employment-bar {
  background: #374151;
}

.employment-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6 0%, #60a5fa 100%);
  border-radius: 3px;
  transition: width 0.5s ease-out;
}

.employment-desc {
  font-size: 0.75rem;
  color: #9ca3af;
  margin-top: 0.25rem;
}

/* 专业介绍 */
.prose {
  color: #374151;
  line-height: 1.75;
}

.dark .prose {
  color: #d1d5db;
}

/* 相关专业 */
.related-majors {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 0.75rem;
}

.related-major-card {
  padding: 1rem;
  border-radius: 0.5rem;
  background: #f9fafb;
  transition: all 0.2s;
  text-decoration: none;
  color: inherit;
}

.dark .related-major-card {
  background: #374151;
}

.related-major-card:hover {
  background: #dcfce7;
  transform: translateY(-2px);
}

.dark .related-major-card:hover {
  background: #166534;
}

.related-major-name {
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 0.25rem;
}

.dark .related-major-name {
  color: #f3f4f6;
}

.related-major-meta {
  font-size: 0.75rem;
  color: #6b7280;
}

.dark .related-major-meta {
  color: #9ca3af;
}

/* 统计卡片 */
.stats-card {
  background: linear-gradient(180deg, #ffffff 0%, #f9fafb 100%);
}

.dark .stats-card {
  background: linear-gradient(180deg, #374151 0%, #1f2937 100%);
}

.quick-stats {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.quick-stat {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem;
  background: white;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.dark .quick-stat {
  background: #4b5563;
}

.stat-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  color: white;
}

.stat-degree {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
}

.stat-duration {
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
}

.stat-employment {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 700;
  color: #1f2937;
}

.dark .stat-value {
  color: #f3f4f6;
}

.stat-label {
  font-size: 0.75rem;
  color: #6b7280;
}

.dark .stat-label {
  color: #9ca3af;
}

/* 开设院校列表 */
.universities-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.university-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem;
  border-radius: 0.5rem;
  background: #f9fafb;
  transition: all 0.2s;
}

.dark .university-item {
  background: #374151;
}

.university-item:hover {
  background: #ede9fe;
}

.dark .university-item:hover {
  background: #4c1d95;
}

.university-item-content {
  flex: 1;
}

.university-item-name {
  font-weight: 600;
  color: #1f2937;
}

.dark .university-item-name {
  color: #f3f4f6;
}

.university-item-meta {
  font-size: 0.75rem;
  color: #6b7280;
}

.dark .university-item-meta {
  color: #9ca3af;
}

.university-item-link {
  padding: 0.5rem;
  color: #6d28d9;
  transition: all 0.2s;
}

.university-item-link:hover {
  color: #5b21b6;
  transform: translateX(2px);
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  background: white;
  color: #374151;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
  cursor: pointer;
}

.dark .action-btn {
  background: #4b5563;
  border-color: #6b7280;
  color: #f3f4f6;
}

.action-btn:hover {
  background: #f9fafb;
  border-color: #10b981;
  color: #10b981;
}

.dark .action-btn:hover {
  background: #6b7280;
  border-color: #34d399;
  color: #34d399;
}

/* 提示卡片 */
.tip-card {
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
}

.dark .tip-card {
  background: linear-gradient(135deg, #78350f 0%, #92400e 100%);
}

.tip-card .section-title {
  color: #92400e;
}

.dark .tip-card .section-title {
  color: #fef08a;
}

.tip-content {
  color: #78350f;
  font-size: 0.875rem;
  line-height: 1.6;
}

.dark .tip-content {
  color: #fef3c7;
}

/* 按钮样式 */
.btn-active {
  background: #fee2e2 !important;
  color: #dc2626 !important;
  border-color: #dc2626 !important;
}

/* 动画 */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card {
  animation: fadeIn 0.3s ease-out;
}

/* 加载和错误状态 */
.loading-container,
.error-container,
.empty-container {
  text-align: center;
  padding: 3rem 0;
}

.error-container .text-red-500 {
  font-size: 1.125rem;
  margin-bottom: 1rem;
}

/* 响应式 */
@media (max-width: 768px) {
  .header-card {
    padding: 1.5rem;
  }

  .employment-grid {
    grid-template-columns: 1fr;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }
}
</style>
