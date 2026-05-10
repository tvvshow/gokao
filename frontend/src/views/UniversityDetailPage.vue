<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <div class="container-modern py-8">
      <!-- 顶部导航栏 -->
      <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
        <button class="btn btn-secondary" @click="goBack">
          <ArrowLeftIcon class="w-4 h-4 mr-2" />
          返回列表
        </button>
        <div class="flex flex-wrap gap-2 md:justify-end">
          <button
            class="btn btn-outline"
            :class="{ 'btn-active': isFavorite }"
            @click="toggleFavorite"
            title="收藏院校"
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
        <p class="text-gray-500 dark:text-gray-400">加载院校详情中...</p>
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
      <div v-else-if="!university" class="empty-container">
        <p class="text-gray-500 dark:text-gray-400 mb-4">未找到院校信息</p>
        <button class="btn btn-primary" @click="goBack">返回列表</button>
      </div>

      <!-- 内容区域 -->
      <template v-else>
        <!-- 院校头部卡片 -->
        <div class="header-card">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-3 mb-3">
                <h1 class="page-title mb-0">{{ university.name }}</h1>
                <div class="flex space-x-1">
                  <span v-if="university.is985" class="badge badge-985"
                    >985</span
                  >
                  <span v-if="university.is211" class="badge badge-211"
                    >211</span
                  >
                  <span
                    v-if="university.isDoubleFirstClass"
                    class="badge badge-double"
                  >
                    双一流
                  </span>
                </div>
              </div>
              <p
                v-if="university.description"
                class="text-gray-600 dark:text-gray-300"
              >
                {{ university.description }}
              </p>
              <div class="flex items-center gap-4 mt-3 text-sm text-gray-500">
                <span v-if="university.province" class="flex items-center">
                  <MapPinIcon class="w-4 h-4 mr-1" />
                  {{ university.province }} {{ university.city }}
                </span>
                <span v-if="university.type" class="flex items-center">
                  <BuildingIcon class="w-4 h-4 mr-1" />
                  {{ university.type }}
                </span>
                <span v-if="university.rank" class="flex items-center">
                  <AwardIcon class="w-4 h-4 mr-1" />
                  排名第{{ university.rank }}名
                </span>
              </div>
            </div>
            <div v-if="university.logo" class="university-logo">
              <img :src="university.logo" :alt="university.name" />
            </div>
          </div>
        </div>

        <!-- 主要内容网格 -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mt-6">
          <!-- 左侧内容 (占2列) -->
          <div class="lg:col-span-2 space-y-6">
            <!-- 录取分数线趋势 -->
            <div v-if="hasScoreData" class="card score-card">
              <h2 class="section-title">
                <TrendingUpIcon class="w-5 h-5 mr-2" />
                录取分数线
              </h2>
              <div class="score-grid">
                <div class="score-item science">
                  <div class="score-label">理科</div>
                  <div class="score-values">
                    <div class="score-value">
                      <span class="score-label-small">最低</span>
                      <span class="score-number">{{
                        university.minScoreScience || '-'
                      }}</span>
                    </div>
                    <div class="score-value">
                      <span class="score-label-small">平均</span>
                      <span class="score-number score-avg">{{
                        university.avgScoreScience || '-'
                      }}</span>
                    </div>
                  </div>
                </div>
                <div class="score-item liberal">
                  <div class="score-label">文科</div>
                  <div class="score-values">
                    <div class="score-value">
                      <span class="score-label-small">最低</span>
                      <span class="score-number">{{
                        university.minScoreLiberalArts || '-'
                      }}</span>
                    </div>
                    <div class="score-value">
                      <span class="score-label-small">平均</span>
                      <span class="score-number score-avg">{{
                        university.avgScoreLiberalArts || '-'
                      }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <!-- 历年录取数据表格 -->
              <div
                v-if="university.admissionData?.length"
                class="admission-table-container"
              >
                <table class="admission-table">
                  <thead>
                    <tr>
                      <th>年份</th>
                      <th>省份</th>
                      <th>批次</th>
                      <th>科类</th>
                      <th>最低分</th>
                      <th>平均分</th>
                      <th>最低位次</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="(data, index) in university.admissionData.slice(
                        0,
                        5
                      )"
                      :key="index"
                    >
                      <td>{{ data.year }}</td>
                      <td>{{ data.province }}</td>
                      <td>{{ data.batchType }}</td>
                      <td>{{ data.scienceType }}</td>
                      <td>{{ data.minScore }}</td>
                      <td>{{ data.avgScore }}</td>
                      <td>{{ data.minRank }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- 基础信息 -->
            <div class="card">
              <h2 class="section-title">
                <InfoIcon class="w-5 h-5 mr-2" />
                基础信息
              </h2>
              <div class="info-grid">
                <div class="info-row">
                  <span class="label">院校类型</span>
                  <span class="value">{{ university.type || '-' }}</span>
                </div>
                <div class="info-row">
                  <span class="label">院校层次</span>
                  <span class="value">{{ university.level || '-' }}</span>
                </div>
                <div class="info-row">
                  <span class="label">建校年份</span>
                  <span class="value">{{
                    university.founded ? `${university.founded}年` : '-'
                  }}</span>
                </div>
                <div class="info-row">
                  <span class="label">在校生</span>
                  <span class="value">{{
                    university.studentCount
                      ? formatNumber(university.studentCount)
                      : '-'
                  }}</span>
                </div>
                <div class="info-row">
                  <span class="label">教师人数</span>
                  <span class="value">{{
                    university.teacherCount
                      ? formatNumber(university.teacherCount)
                      : '-'
                  }}</span>
                </div>
                <div class="info-row">
                  <span class="label">就业率</span>
                  <span class="value">{{
                    university.employmentRate
                      ? `${university.employmentRate}%`
                      : '-'
                  }}</span>
                </div>
                <div class="info-row">
                  <span class="label">校园面积</span>
                  <span class="value">{{
                    university.campusArea ? `${university.campusArea}亩` : '-'
                  }}</span>
                </div>
                <div class="info-row">
                  <span class="label">专业数量</span>
                  <span class="value">{{
                    university.majorCount ? `${university.majorCount}个` : '-'
                  }}</span>
                </div>
              </div>

              <!-- 联系方式 -->
              <div
                class="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700"
              >
                <h3
                  class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3"
                >
                  联系方式
                </h3>
                <div class="contact-grid">
                  <a
                    v-if="university.website"
                    :href="university.website"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="contact-link"
                  >
                    <GlobeIcon class="w-4 h-4" />
                    <span>{{ university.website }}</span>
                  </a>
                  <div v-if="university.phone" class="contact-item">
                    <PhoneIcon class="w-4 h-4" />
                    <span>{{ university.phone }}</span>
                  </div>
                  <div v-if="university.email" class="contact-item">
                    <MailIcon class="w-4 h-4" />
                    <span>{{ university.email }}</span>
                  </div>
                  <div v-if="university.address" class="contact-item">
                    <MapPinIcon class="w-4 h-4" />
                    <span>{{ university.address }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- 优势专业 -->
            <div v-if="university.strongMajors?.length" class="card">
              <h2 class="section-title">
                <StarIcon class="w-5 h-5 mr-2" />
                优势专业
              </h2>
              <div class="tags-container">
                <router-link
                  v-for="major in university.strongMajors"
                  :key="major"
                  :to="`/majors/${major}`"
                  class="tag tag-major"
                >
                  <BookOpenIcon class="w-3 h-3" />
                  {{ major }}
                </router-link>
              </div>
            </div>

            <!-- 开设专业 -->
            <div v-if="university.majors?.length" class="card">
              <div class="flex items-center justify-between mb-4">
                <h2 class="section-title mb-0">
                  <BookOpenIcon class="w-5 h-5 mr-2" />
                  开设专业
                </h2>
                <router-link
                  to="/majors"
                  class="text-sm text-primary-600 hover:underline"
                >
                  查看全部 →
                </router-link>
              </div>
              <div class="major-grid">
                <router-link
                  v-for="major in university.majors.slice(0, 12)"
                  :key="major.id"
                  :to="`/majors/${major.id}`"
                  class="major-card"
                >
                  <div class="major-name">{{ major.name }}</div>
                  <div class="major-meta">
                    {{ major.category || '-' }}
                    <span v-if="major.degree">· {{ major.degree }}</span>
                  </div>
                </router-link>
              </div>
            </div>
          </div>

          <!-- 右侧边栏 -->
          <div class="space-y-6">
            <!-- 快速统计 -->
            <div class="card stats-card">
              <h2 class="section-title">关键数据</h2>
              <div class="quick-stats">
                <div class="quick-stat">
                  <div class="stat-icon stat-rank">
                    <AwardIcon class="w-5 h-5" />
                  </div>
                  <div class="stat-content">
                    <div class="stat-value">
                      {{ university.rank ?? '-' }}
                    </div>
                    <div class="stat-label">全国排名</div>
                  </div>
                </div>
                <div class="quick-stat">
                  <div class="stat-icon stat-employment">
                    <BriefcaseIcon class="w-5 h-5" />
                  </div>
                  <div class="stat-content">
                    <div class="stat-value">
                      {{
                        university.employmentRate
                          ? `${university.employmentRate}%`
                          : '-'
                      }}
                    </div>
                    <div class="stat-label">就业率</div>
                  </div>
                </div>
                <div class="quick-stat">
                  <div class="stat-icon stat-majors">
                    <BookOpenIcon class="w-5 h-5" />
                  </div>
                  <div class="stat-content">
                    <div class="stat-value">
                      {{ university.majorCount || '-' }}
                    </div>
                    <div class="stat-label">专业数量</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 特色标签 -->
            <div v-if="university.features?.length" class="card">
              <h2 class="section-title">特色标签</h2>
              <div class="tags-container">
                <span
                  v-for="feature in university.features"
                  :key="feature"
                  class="tag tag-feature"
                >
                  {{ feature }}
                </span>
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
                <button class="action-btn" @click="generateReport">
                  <FileTextIcon class="w-5 h-5" />
                  <span>生成报告</span>
                </button>
                <button class="action-btn" @click="shareUniversity">
                  <ShareIcon class="w-5 h-5" />
                  <span>分享院校</span>
                </button>
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
import { universityApi } from '@/api/university';
import { ElMessage } from 'element-plus';
import {
  ArrowLeftIcon,
  HeartIcon,
  ScaleIcon,
  MapPinIcon,
  BuildingIcon,
  AwardIcon,
  TrendingUpIcon,
  InfoIcon,
  GlobeIcon,
  PhoneIcon,
  MailIcon,
  StarIcon,
  BookOpenIcon,
  BriefcaseIcon,
  FileTextIcon,
  ShareIcon,
  RefreshCwIcon,
} from 'lucide-vue-next';
import type { UniversityDetail } from '@/types/university';

const route = useRoute();
const router = useRouter();

const university = ref<UniversityDetail | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);
const isFavorite = ref(false);

const hasScoreData = computed(() => {
  const u = university.value;
  return (
    u &&
    (u.minScoreScience ||
      u.minScoreLiberalArts ||
      u.avgScoreScience ||
      u.avgScoreLiberalArts)
  );
});

const fetchDetail = async () => {
  const id = route.params.id as string | undefined;
  if (!id) {
    error.value = '院校ID缺失';
    university.value = null;
    return;
  }
  loading.value = true;
  error.value = null;
  try {
    const response = await universityApi.getDetail(id);
    if (response.success) {
      university.value = response.data;
      // 检查是否已收藏
      isFavorite.value = response.data.isFavorite || false;
    } else {
      university.value = null;
      error.value = response.message || '获取院校详情失败';
    }
  } catch (err) {
    university.value = null;
    error.value = err instanceof Error ? err.message : '获取院校详情失败';
  } finally {
    loading.value = false;
  }
};

const goBack = () => {
  router.push('/universities');
};

const toggleFavorite = () => {
  isFavorite.value = !isFavorite.value;
  ElMessage.success(isFavorite.value ? '已添加到收藏' : '已取消收藏');
};

const addToCompare = () => {
  if (!university.value) return;
  ElMessage.success(`已添加 ${university.value.name} 到对比列表`);
};

const generateReport = () => {
  ElMessage.info('报告生成功能开发中');
};

const shareUniversity = () => {
  const url = window.location.href;
  navigator.clipboard?.writeText(url);
  ElMessage.success('链接已复制到剪贴板');
};

const formatNumber = (num: number) => {
  if (num >= 10000) {
    return `${(num / 10000).toFixed(1)}万`;
  }
  return num.toString();
};

onMounted(fetchDetail);
watch(() => route.params.id, fetchDetail);
</script>

<style scoped>
.header-card {
  background:
    radial-gradient(
      circle at top right,
      rgba(125, 211, 252, 0.28),
      transparent 45%
    ),
    linear-gradient(135deg, #0369a1 0%, #0f766e 55%, #164e63 100%);
  color: #f8fafc;
  padding: 2rem;
  border-radius: 1rem;
  border: 1px solid rgba(148, 163, 184, 0.28);
  box-shadow:
    0 12px 32px -14px rgba(2, 132, 199, 0.55),
    0 18px 40px -26px rgba(15, 23, 42, 0.72);
}

.header-card .page-title {
  color: #fff;
  margin: 0;
}

.header-card .text-gray-600 {
  color: rgba(240, 249, 255, 0.9);
}

.header-card .text-gray-500 {
  color: rgba(224, 242, 254, 0.78);
}

.university-logo {
  width: 80px;
  height: 80px;
  background: rgba(248, 250, 252, 0.95);
  border-radius: 1rem;
  border: 1px solid rgba(186, 230, 253, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  box-shadow: 0 8px 18px -12px rgba(15, 23, 42, 0.9);
}

.university-logo img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.badge {
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  border: 1px solid transparent;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.01em;
}

.badge-985 {
  background: rgba(254, 226, 226, 0.22);
  border-color: rgba(252, 165, 165, 0.4);
  color: #fee2e2;
}

.badge-211 {
  background: rgba(254, 243, 199, 0.22);
  border-color: rgba(252, 211, 77, 0.45);
  color: #fef3c7;
}

.badge-double {
  background: rgba(224, 242, 254, 0.25);
  border-color: rgba(125, 211, 252, 0.48);
  color: #e0f2fe;
}

.score-card {
  background: linear-gradient(180deg, #f8fcff 0%, #f0f9ff 100%);
}

.dark .score-card {
  background: linear-gradient(180deg, #1f2937 0%, #111827 100%);
}

.score-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 1rem;
}

.score-item {
  background: rgba(255, 255, 255, 0.92);
  padding: 1rem;
  border-radius: 0.875rem;
  border: 1px solid #e2e8f0;
  box-shadow: 0 8px 20px -18px rgba(15, 23, 42, 0.6);
}

.dark .score-item {
  background: rgba(31, 41, 55, 0.92);
  border-color: #374151;
}

.score-item.science {
  border-left: 4px solid #0284c7;
}

.score-item.liberal {
  border-left: 4px solid #14b8a6;
}

.score-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #64748b;
  margin-bottom: 0.5rem;
}

.dark .score-label {
  color: #94a3b8;
}

.score-values {
  display: flex;
  gap: 1rem;
}

.score-value {
  flex: 1;
}

.score-label-small {
  font-size: 0.75rem;
  color: #94a3b8;
}

.score-number {
  font-size: 1.25rem;
  font-weight: 700;
  color: #0f172a;
}

.dark .score-number {
  color: #f1f5f9;
}

.score-avg {
  color: #0284c7;
}

.admission-table-container {
  overflow-x: auto;
  margin-top: 1rem;
  border-radius: 0.75rem;
  border: 1px solid #e2e8f0;
}

.dark .admission-table-container {
  border-color: #334155;
}

.admission-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.admission-table th,
.admission-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

.dark .admission-table th,
.dark .admission-table td {
  border-bottom-color: #334155;
}

.admission-table th {
  background: #f8fafc;
  font-weight: 600;
  color: #334155;
}

.dark .admission-table th {
  background: #1f2937;
  color: #e2e8f0;
}

.admission-table tbody tr:hover {
  background: #f0f9ff;
}

.dark .admission-table tbody tr:hover {
  background: rgba(30, 41, 59, 0.8);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.625rem 0;
  border-bottom: 1px dashed #d1d9e6;
}

.dark .info-row {
  border-bottom-color: #374151;
}

.info-row:last-child {
  border-bottom: none;
}

.label {
  color: #64748b;
  font-size: 0.875rem;
}

.dark .label {
  color: #94a3b8;
}

.value {
  color: #0f172a;
  font-weight: 600;
}

.dark .value {
  color: #f1f5f9;
}

.contact-grid {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.contact-link,
.contact-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #475569;
  font-size: 0.875rem;
}

.contact-link:hover {
  color: #0284c7;
}

.dark .contact-link,
.dark .contact-item {
  color: #94a3b8;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.tag {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.375rem 0.75rem;
  border-radius: 9999px;
  border: 1px solid transparent;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.tag-major {
  background: #ecfeff;
  border-color: #bae6fd;
  color: #0e7490;
}

.dark .tag-major {
  background: rgba(14, 116, 144, 0.22);
  border-color: rgba(125, 211, 252, 0.4);
  color: #67e8f9;
}

.tag-major:hover {
  background: #cffafe;
}

.tag-feature {
  background: #f0f9ff;
  border-color: #bae6fd;
  color: #0369a1;
}

.dark .tag-feature {
  background: rgba(3, 105, 161, 0.22);
  border-color: rgba(56, 189, 248, 0.35);
  color: #7dd3fc;
}

.major-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 0.75rem;
}

.major-card {
  padding: 0.875rem;
  border-radius: 0.75rem;
  border: 1px solid #e2e8f0;
  background: linear-gradient(180deg, #fff 0%, #f8fafc 100%);
  transition: all 0.2s;
  text-decoration: none;
  color: inherit;
}

.dark .major-card {
  background: linear-gradient(180deg, #1f2937 0%, #111827 100%);
  border-color: #334155;
}

.major-card:hover {
  border-color: #7dd3fc;
  background: #f0f9ff;
  transform: translateY(-2px);
}

.dark .major-card:hover {
  border-color: #22d3ee;
  background: #0f172a;
}

.major-name {
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 0.25rem;
}

.dark .major-name {
  color: #f1f5f9;
}

.major-meta {
  font-size: 0.75rem;
  color: #64748b;
}

.dark .major-meta {
  color: #94a3b8;
}

.stats-card {
  background: linear-gradient(180deg, #fff 0%, #f8fafc 100%);
}

.dark .stats-card {
  background: linear-gradient(180deg, #1f2937 0%, #111827 100%);
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
  padding: 0.875rem;
  background: rgba(255, 255, 255, 0.88);
  border: 1px solid #e2e8f0;
  border-radius: 0.75rem;
  box-shadow: 0 8px 20px -18px rgba(15, 23, 42, 0.7);
}

.dark .quick-stat {
  background: rgba(31, 41, 55, 0.92);
  border-color: #334155;
}

.stat-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.625rem;
  color: white;
}

.stat-rank {
  background: linear-gradient(135deg, #f59e0b 0%, #f97316 100%);
}

.stat-employment {
  background: linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%);
}

.stat-majors {
  background: linear-gradient(135deg, #14b8a6 0%, #0d9488 100%);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 700;
  color: #0f172a;
}

.dark .stat-value {
  color: #f1f5f9;
}

.stat-label {
  font-size: 0.75rem;
  color: #64748b;
}

.dark .stat-label {
  color: #94a3b8;
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.8125rem;
  border: 1px solid #dbe3ef;
  border-radius: 0.75rem;
  background: #fff;
  color: #334155;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
  cursor: pointer;
}

.dark .action-btn {
  background: #1f2937;
  border-color: #374151;
  color: #e2e8f0;
}

.action-btn:hover {
  background: #f0f9ff;
  border-color: #38bdf8;
  color: #0369a1;
}

.dark .action-btn:hover {
  background: #0f172a;
  border-color: #22d3ee;
  color: #67e8f9;
}

.btn-active {
  background: #fee2e2 !important;
  color: #dc2626 !important;
  border-color: #dc2626 !important;
}

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

@media (max-width: 768px) {
  .header-card {
    padding: 1.375rem;
  }

  .university-logo {
    display: none;
  }

  .score-grid,
  .info-grid {
    grid-template-columns: 1fr;
  }
}
</style>
