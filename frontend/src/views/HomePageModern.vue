<template>
  <div class="min-h-screen home-shell">
    <!-- 英雄区域 -->
    <section class="hero-section py-16 md:py-24">
      <div class="container-modern">
        <div class="grid lg:grid-cols-2 gap-12 items-center hero-grid">
          <div class="space-y-8 animate-fade-in hero-copy">
            <div class="space-y-4">
              <div class="hero-kicker">
                <SparklesIcon class="w-4 h-4 mr-2" />
                AI智能推荐系统
              </div>
              <h1 class="hero-title">智能高考<br />志愿填报</h1>
              <p
                class="text-lg md:text-xl text-gray-600 dark:text-gray-300 leading-relaxed"
              >
                基于大数据分析和AI算法，为您提供个性化的志愿填报方案，让每一分都发挥最大价值
              </p>
            </div>

            <div class="flex flex-wrap gap-4 hero-cta-row">
              <button
                @click="startRecommendation"
                class="btn text-lg px-8 py-3 hero-cta-primary"
                aria-label="开始智能推荐，获取个性化志愿方案"
              >
                <SparklesIcon class="w-5 h-5 mr-2" aria-hidden="true" />
                开始智能推荐
              </button>
              <button
                @click="viewColleges"
                class="btn text-lg px-8 py-3 hero-cta-secondary"
                aria-label="查看院校信息，浏览全国高校数据"
              >
                <BuildingIcon class="w-5 h-5 mr-2" aria-hidden="true" />
                查看院校信息
              </button>
            </div>

            <div class="hero-proof" role="list" aria-label="平台数据统计">
              <div class="hero-proof-item" role="listitem">
                <CheckCircleIcon
                  class="w-4 h-4 mr-2 text-success-500"
                  aria-hidden="true"
                />
                2700+ 高校数据
              </div>
              <div class="hero-proof-item" role="listitem">
                <CheckCircleIcon
                  class="w-4 h-4 mr-2 text-success-500"
                  aria-hidden="true"
                />
                1400+ 专业信息
              </div>
              <div class="hero-proof-item" role="listitem">
                <CheckCircleIcon
                  class="w-4 h-4 mr-2 text-success-500"
                  aria-hidden="true"
                />
                AI智能分析
              </div>
            </div>
          </div>

          <div class="relative animate-float hero-preview-shell">
            <div class="relative z-10">
              <div class="hero-preview p-8">
                <div class="space-y-6">
                  <div class="flex items-center justify-between">
                    <h3
                      class="text-lg font-semibold text-gray-900 dark:text-gray-100"
                    >
                      推荐结果预览
                    </h3>
                    <div class="badge badge-success">匹配度 95%</div>
                  </div>
                  <div class="space-y-4">
                    <div class="hero-preview__item">
                      <div
                        class="w-10 h-10 bg-primary-100 dark:bg-primary-900 rounded-lg flex items-center justify-center"
                      >
                        <BuildingIcon
                          class="w-5 h-5 text-primary-600 dark:text-primary-400"
                        />
                      </div>
                      <div>
                        <div
                          class="font-medium text-gray-900 dark:text-gray-100"
                        >
                          清华大学
                        </div>
                        <div class="text-sm text-gray-500 dark:text-gray-400">
                          计算机科学与技术
                        </div>
                      </div>
                      <div
                        class="ml-auto hero-preview__tag hero-preview__tag--a"
                      >
                        冲刺
                      </div>
                    </div>
                    <div class="hero-preview__item">
                      <div
                        class="w-10 h-10 bg-success-100 dark:bg-success-900 rounded-lg flex items-center justify-center"
                      >
                        <BuildingIcon
                          class="w-5 h-5 text-success-600 dark:text-success-400"
                        />
                      </div>
                      <div>
                        <div
                          class="font-medium text-gray-900 dark:text-gray-100"
                        >
                          北京理工大学
                        </div>
                        <div class="text-sm text-gray-500 dark:text-gray-400">
                          软件工程
                        </div>
                      </div>
                      <div
                        class="ml-auto hero-preview__tag hero-preview__tag--b"
                      >
                        稳妥
                      </div>
                    </div>
                    <div class="hero-preview__item">
                      <div
                        class="w-10 h-10 bg-warning-100 dark:bg-warning-900 rounded-lg flex items-center justify-center"
                      >
                        <BuildingIcon
                          class="w-5 h-5 text-warning-600 dark:text-warning-400"
                        />
                      </div>
                      <div>
                        <div
                          class="font-medium text-gray-900 dark:text-gray-100"
                        >
                          华北电力大学
                        </div>
                        <div class="text-sm text-gray-500 dark:text-gray-400">
                          电气工程
                        </div>
                      </div>
                      <div
                        class="ml-auto hero-preview__tag hero-preview__tag--c"
                      >
                        保底
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div class="hero-orb hero-orb--right"></div>
            <div class="hero-orb hero-orb--left"></div>
          </div>
        </div>
      </div>
    </section>

    <!-- 统计数据 -->
    <section
      class="py-16 md:py-24 bg-white dark:bg-gray-900"
      aria-labelledby="stats-heading"
    >
      <h2 id="stats-heading" class="sr-only">平台统计数据</h2>
      <div class="container-modern">
        <div class="grid grid-cols-2 md:grid-cols-4 gap-6" role="list">
          <div
            v-for="stat in stats"
            :key="stat.label"
            class="stat-card"
            role="listitem"
          >
            <div class="stat-icon">
              <component
                :is="stat.icon"
                class="w-8 h-8 text-primary-600 dark:text-primary-400"
                aria-hidden="true"
              />
            </div>
            <div
              v-if="isLoadingStats"
              class="stat-skeleton"
              aria-busy="true"
              aria-label="加载中"
            >
              <div class="skeleton-line skeleton-value"></div>
              <div class="skeleton-line skeleton-label"></div>
            </div>
            <template v-else>
              <div
                class="stat-number"
                :aria-label="`${stat.label}：${stat.value}`"
              >
                {{ stat.value }}
              </div>
              <div class="stat-label">{{ stat.label }}</div>
            </template>
          </div>
        </div>
      </div>
    </section>

    <!-- 功能特色 -->
    <section
      class="py-16 md:py-24 bg-gray-50 dark:bg-gray-800"
      aria-labelledby="features-heading"
    >
      <div class="container-modern">
        <div class="text-center mb-16">
          <h2 id="features-heading" class="page-title">核心功能</h2>
          <p class="text-xl text-gray-600 dark:text-gray-300 mt-4">
            全方位的志愿填报解决方案
          </p>
        </div>

        <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-6" role="list">
          <div
            v-for="feature in features"
            :key="feature.title"
            class="card card-hover p-8 group cursor-pointer"
            role="listitem"
            tabindex="0"
            :aria-label="`${feature.title}：${feature.description}，点击了解更多`"
            @click="$router.push(feature.link)"
            @keydown.enter="$router.push(feature.link)"
            @keydown.space.prevent="$router.push(feature.link)"
          >
            <div class="feature-icon mb-6">
              <div
                class="w-16 h-16 bg-primary-100 dark:bg-primary-900 rounded-2xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300"
              >
                <component
                  :is="feature.icon"
                  class="w-8 h-8 text-primary-600 dark:text-primary-400"
                  aria-hidden="true"
                />
              </div>
            </div>
            <h3
              class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-3"
            >
              {{ feature.title }}
            </h3>
            <p class="text-gray-600 dark:text-gray-300 mb-6">
              {{ feature.description }}
            </p>
            <div
              class="flex items-center text-primary-600 dark:text-primary-400 font-medium group-hover:translate-x-2 transition-transform duration-300"
              aria-hidden="true"
            >
              了解更多
              <ArrowRightIcon class="w-4 h-4 ml-2" />
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 推荐流程 -->
    <section class="py-16 md:py-24 bg-white dark:bg-gray-900">
      <div class="container-modern">
        <div class="text-center mb-16">
          <h2 class="page-title">智能推荐流程</h2>
          <p class="text-xl text-gray-600 dark:text-gray-300 mt-4">
            四步完成个性化志愿方案
          </p>
        </div>

        <div class="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
          <div
            v-for="(step, index) in steps"
            :key="step.title"
            class="text-center group"
          >
            <div class="relative mb-6">
              <div
                class="w-20 h-20 bg-gradient-to-br from-primary-500 to-secondary-500 rounded-full flex items-center justify-center mx-auto group-hover:scale-110 transition-transform duration-300"
              >
                <span class="text-2xl font-bold text-white">{{
                  index + 1
                }}</span>
              </div>
              <div
                v-if="index < steps.length - 1"
                class="hidden lg:block absolute top-10 left-full w-full h-0.5 bg-gradient-to-r from-primary-300 to-transparent"
              ></div>
            </div>
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3"
            >
              {{ step.title }}
            </h3>
            <p class="text-gray-600 dark:text-gray-300">
              {{ step.description }}
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA区域 -->
    <section class="py-16 md:py-24 cta-section" aria-labelledby="cta-heading">
      <div class="container-modern text-center">
        <div class="max-w-3xl mx-auto space-y-8">
          <h2
            id="cta-heading"
            class="text-3xl lg:text-4xl font-bold text-white"
          >
            开始您的智能志愿填报之旅
          </h2>
          <p class="text-xl text-cyan-100/90">
            让AI为您的未来保驾护航，每一分都不浪费
          </p>
          <div class="flex flex-wrap justify-center gap-4">
            <button
              @click="startRecommendation"
              class="btn text-lg px-8 py-3 hero-cta-white"
              aria-label="立即开始智能推荐"
            >
              立即开始推荐
            </button>
            <button
              @click="viewDemo"
              class="btn text-lg px-8 py-3 hero-cta-outline"
              aria-label="查看系统演示"
            >
              查看演示
            </button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import {
  SparklesIcon,
  BuildingIcon,
  CheckCircleIcon,
  ArrowRightIcon,
  UsersIcon,
  BookOpenIcon,
  BarChartIcon,
  CrownIcon,
} from 'lucide-vue-next';
import { api } from '@/api/api-client';
import {
  DEFAULT_HOME_STATISTICS,
  HOME_FEATURES,
  RECOMMENDATION_STEPS,
} from '@/config/constants';
import type { HomeStatistics } from '@/types/api';

const router = useRouter();

// Statistics data - loaded from API with fallback to defaults
const statistics = ref<HomeStatistics>({
  universityCount: DEFAULT_HOME_STATISTICS.universityCount,
  majorCount: DEFAULT_HOME_STATISTICS.majorCount,
  userCount: DEFAULT_HOME_STATISTICS.userCount,
  accuracyRate: DEFAULT_HOME_STATISTICS.accuracyRate,
});

const isLoadingStats = ref(false);

// Computed stats for display
const stats = computed(() => [
  {
    icon: BuildingIcon,
    value: `${statistics.value.universityCount}+`,
    label: '合作高校',
  },
  {
    icon: BookOpenIcon,
    value: `${statistics.value.majorCount}+`,
    label: '专业数据',
  },
  {
    icon: UsersIcon,
    value: formatUserCount(statistics.value.userCount),
    label: '服务学生',
  },
  {
    icon: BarChartIcon,
    value: `${statistics.value.accuracyRate}%`,
    label: '推荐准确率',
  },
]);

// Format user count for display (e.g., 500000 -> "50万+")
function formatUserCount(count: number): string {
  if (count >= 10000) {
    return `${Math.floor(count / 10000)}万+`;
  }
  return `${count}+`;
}

// Feature list from config
const features = ref(
  HOME_FEATURES.map((feature) => ({
    ...feature,
    icon: getFeatureIcon(feature.id),
  }))
);

// Get icon component for feature
function getFeatureIcon(featureId: string) {
  const iconMap: Record<string, typeof SparklesIcon> = {
    'ai-recommendation': SparklesIcon,
    'university-search': BuildingIcon,
    'major-analysis': BookOpenIcon,
    'data-analysis': BarChartIcon,
    membership: CrownIcon,
    simulation: UsersIcon,
  };
  return iconMap[featureId] || SparklesIcon;
}

// Recommendation steps from config
const steps = ref([...RECOMMENDATION_STEPS]);

// Fetch statistics from API
async function fetchStatistics() {
  isLoadingStats.value = true;
  try {
    const response = await api.universities.statistics();
    if (response.success && response.data) {
      // Map API response to our statistics format
      const data = response.data as Record<string, unknown>;
      statistics.value = {
        universityCount:
          (data.total as number) || DEFAULT_HOME_STATISTICS.universityCount,
        majorCount:
          (data.majorCount as number) || DEFAULT_HOME_STATISTICS.majorCount,
        userCount:
          (data.userCount as number) || DEFAULT_HOME_STATISTICS.userCount,
        accuracyRate:
          (data.accuracyRate as number) || DEFAULT_HOME_STATISTICS.accuracyRate,
      };
    }
  } catch (error) {
    // Use default values on error
    console.warn('Failed to fetch statistics, using defaults:', error);
  } finally {
    isLoadingStats.value = false;
  }
}

// Methods
const startRecommendation = () => {
  router.push('/recommendation');
};

const viewColleges = () => {
  router.push('/universities');
};

const viewDemo = () => {
  // Demo functionality
  router.push('/recommendation?demo=true');
};

// Fetch statistics on mount
onMounted(() => {
  fetchStatistics();
});
</script>

<style scoped>
.home-shell {
  position: relative;
}

/* 主视觉区域 */
.hero-section {
  background:
    radial-gradient(
      1200px circle at 10% 15%,
      rgb(34 211 238 / 0.16),
      transparent 45%
    ),
    radial-gradient(
      1000px circle at 85% 0%,
      rgb(56 189 248 / 0.18),
      transparent 40%
    ),
    linear-gradient(165deg, #f8fafc 0%, #eef9ff 42%, #f0f9ff 100%);
}

.dark .hero-section {
  background:
    radial-gradient(
      900px circle at 15% 20%,
      rgb(6 182 212 / 0.2),
      transparent 46%
    ),
    radial-gradient(
      820px circle at 85% 0%,
      rgb(37 99 235 / 0.18),
      transparent 45%
    ),
    linear-gradient(170deg, #020617 0%, #0b1328 50%, #0f172a 100%);
}

.hero-grid {
  align-items: stretch;
}

.hero-copy {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.hero-kicker {
  display: inline-flex;
  align-items: center;
  padding: 0.52rem 1rem;
  border-radius: 999px;
  border: 1px solid rgb(14 165 233 / 0.3);
  background: linear-gradient(
    120deg,
    rgb(224 242 254 / 0.95),
    rgb(204 251 241 / 0.95)
  );
  color: #0369a1;
  font-size: 0.84rem;
  font-weight: 700;
  letter-spacing: 0.03em;
}

.dark .hero-kicker {
  border-color: rgb(56 189 248 / 0.3);
  background: linear-gradient(
    120deg,
    rgb(8 47 73 / 0.82),
    rgb(17 94 89 / 0.82)
  );
  color: #67e8f9;
}

.hero-title {
  background: linear-gradient(115deg, #0f172a 0%, #0f766e 38%, #0ea5e9 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-size: clamp(2.5rem, 6vw, 4.25rem);
  line-height: 1.1;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.dark .hero-title {
  background: linear-gradient(120deg, #e2e8f0 0%, #67e8f9 46%, #38bdf8 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-cta-row {
  margin-top: 0.3rem;
}

.hero-cta-primary {
  border: none;
  color: #f8fafc;
  background: linear-gradient(135deg, #0284c7 0%, #0f766e 100%);
  box-shadow: 0 16px 34px -22px rgb(2 132 199 / 0.85);
}

.hero-cta-primary:hover {
  filter: brightness(1.06);
}

.hero-cta-secondary {
  color: #0f172a;
  border: 1px solid rgb(148 163 184 / 0.38);
  background: rgb(255 255 255 / 0.88);
}

.dark .hero-cta-secondary {
  color: #e2e8f0;
  border-color: rgb(71 85 105 / 0.7);
  background: rgb(15 23 42 / 0.85);
}

.hero-preview {
  border-radius: 1.25rem;
  border: 1px solid rgb(148 163 184 / 0.3);
  background: rgb(255 255 255 / 0.86);
  box-shadow: 0 24px 64px -36px rgb(2 132 199 / 0.5);
  backdrop-filter: blur(6px);
}

.dark .hero-preview {
  border-color: rgb(71 85 105 / 0.5);
  background: rgb(15 23 42 / 0.8);
  box-shadow: 0 24px 64px -32px rgb(15 23 42 / 0.85);
}

.hero-preview-shell {
  display: flex;
  align-items: center;
}

.hero-preview__item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 0.75rem;
  background: rgb(255 255 255 / 0.82);
  border: 1px solid rgb(226 232 240 / 0.8);
}

.hero-preview__tag {
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 700;
}

.hero-preview__tag--a {
  color: #14532d;
  background: #dcfce7;
}

.hero-preview__tag--b {
  color: #075985;
  background: #e0f2fe;
}

.hero-preview__tag--c {
  color: #92400e;
  background: #fef3c7;
}

.dark .hero-preview__item {
  background: rgb(30 41 59 / 0.88);
  border-color: rgb(71 85 105 / 0.55);
}

.hero-proof {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
}

.hero-proof-item {
  display: inline-flex;
  align-items: center;
  border: 1px solid rgb(148 163 184 / 0.24);
  border-radius: 999px;
  padding: 0.46rem 0.9rem;
  background: rgb(248 250 252 / 0.74);
  color: #475569;
  font-size: 0.82rem;
}

.dark .hero-proof-item {
  border-color: rgb(71 85 105 / 0.62);
  background: rgb(30 41 59 / 0.72);
  color: #cbd5e1;
}

.hero-orb {
  position: absolute;
  width: 18rem;
  height: 18rem;
  border-radius: 9999px;
  opacity: 0.24;
  filter: blur(42px);
  pointer-events: none;
}

.hero-orb--right {
  top: -1.1rem;
  right: -1rem;
  background: linear-gradient(145deg, #22d3ee, #0ea5e9);
}

.hero-orb--left {
  left: -1rem;
  bottom: -1rem;
  background: linear-gradient(145deg, #38bdf8, #2563eb);
}

.cta-section {
  background: linear-gradient(130deg, #0f172a 0%, #0f766e 48%, #0ea5e9 100%);
}

.hero-cta-white {
  border: none;
  color: #0f172a;
  background: #f8fafc;
  font-weight: 700;
}

.hero-cta-white:hover {
  background: #e2e8f0;
}

.hero-cta-outline {
  border: 1px solid rgb(226 232 240 / 0.84);
  background: rgb(15 23 42 / 0.15);
  color: #f8fafc;
  font-weight: 700;
}

.hero-cta-outline:hover {
  background: rgb(248 250 252 / 0.92);
  color: #0f172a;
}

/* 动画效果 */
.feature-icon {
  transition: all 300ms ease;
}

/* 骨架屏加载样式 */
.stat-skeleton {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.skeleton-line {
  background: linear-gradient(90deg, #e5e7eb 25%, #d1d5db 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  border-radius: 0.25rem;
  animation: skeleton-loading 1.5s infinite;
}

.skeleton-value {
  width: 80px;
  height: 32px;
}

.skeleton-label {
  width: 60px;
  height: 16px;
}

@keyframes skeleton-loading {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

/* 暗色模式骨架屏 */
.dark .skeleton-line {
  background: linear-gradient(90deg, #374151 25%, #4b5563 50%, #374151 75%);
  background-size: 200% 100%;
}
</style>
