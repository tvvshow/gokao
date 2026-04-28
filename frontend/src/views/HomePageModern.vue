<template>
  <div class="min-h-screen">
    <!-- 英雄区域 -->
    <section class="hero-section">
      <div class="container">
        <div class="grid-responsive lg:grid-cols-2 gap-8xl items-center">
          <div class="space-y-6xl animate-fade-in">
            <div class="space-y-4">
              <div
                class="inline-flex items-center px-4 py-2 bg-color-primary-100 dark:bg-color-primary-900 rounded-full text-color-primary-700 dark:text-color-primary-300 text-sm font-medium"
              >
                <SparklesIcon class="w-4 h-4 mr-2" />
                AI智能推荐系统
              </div>
              <h1 class="hero-section__title text-gradient">
                智能高考<br />志愿填报
              </h1>
              <p class="hero-section__subtitle">
                基于大数据分析和AI算法，为您提供个性化的志愿填报方案，让每一分都发挥最大价值
              </p>
            </div>

            <div class="hero-section__actions">
              <button
                @click="startRecommendation"
                class="btn btn--primary text-lg"
                aria-label="开始智能推荐，获取个性化志愿方案"
              >
                <SparklesIcon class="w-5 h-5 mr-2" aria-hidden="true" />
                开始智能推荐
              </button>
              <button
                @click="viewColleges"
                class="btn btn--secondary text-lg"
                aria-label="查看院校信息，浏览全国高校数据"
              >
                <BuildingIcon class="w-5 h-5 mr-2" aria-hidden="true" />
                查看院校信息
              </button>
            </div>

            <div
              class="flex items-center space-x-6 text-sm text-color-gray-500 dark:text-color-gray-400"
              role="list"
              aria-label="平台数据统计"
            >
              <div class="flex items-center" role="listitem">
                <CheckCircleIcon
                  class="w-4 h-4 mr-2 text-color-success-500"
                  aria-hidden="true"
                />
                2700+ 高校数据
              </div>
              <div class="flex items-center" role="listitem">
                <CheckCircleIcon
                  class="w-4 h-4 mr-2 text-color-success-500"
                  aria-hidden="true"
                />
                1400+ 专业信息
              </div>
              <div class="flex items-center" role="listitem">
                <CheckCircleIcon
                  class="w-4 h-4 mr-2 text-color-success-500"
                  aria-hidden="true"
                />
                AI智能分析
              </div>
            </div>
          </div>

          <div class="relative animate-float">
            <div class="relative z-10">
              <div class="base-card p-8xl">
                <div class="space-y-6xl">
                  <div class="flex items-center justify-between">
                    <h3
                      class="text-lg font-semibold text-color-gray-900 dark:text-color-gray-100"
                    >
                      推荐结果预览
                    </h3>
                    <div class="badge badge-success">匹配度 95%</div>
                  </div>
                  <div class="space-y-4">
                    <div
                      class="flex items-center space-x-3 p-3xl bg-white dark:bg-color-gray-800 rounded-lg"
                    >
                      <div
                        class="w-10 h-10 bg-color-primary-100 dark:bg-color-primary-900 rounded-lg flex items-center justify-center"
                      >
                        <BuildingIcon
                          class="w-5 h-5 text-color-primary-600 dark:text-color-primary-400"
                        />
                      </div>
                      <div>
                        <div
                          class="font-medium text-color-gray-900 dark:text-color-gray-100"
                        >
                          清华大学
                        </div>
                        <div
                          class="text-sm text-color-gray-500 dark:text-color-gray-400"
                        >
                          计算机科学与技术
                        </div>
                      </div>
                      <div
                        class="ml-auto text-color-success-600 dark:text-color-success-400 font-medium"
                      >
                        冲刺
                      </div>
                    </div>
                    <div
                      class="flex items-center space-x-3 p-3xl bg-white dark:bg-color-gray-800 rounded-lg"
                    >
                      <div
                        class="w-10 h-10 bg-color-success-100 dark:bg-color-success-900 rounded-lg flex items-center justify-center"
                      >
                        <BuildingIcon
                          class="w-5 h-5 text-color-success-600 dark:text-color-success-400"
                        />
                      </div>
                      <div>
                        <div
                          class="font-medium text-color-gray-900 dark:text-color-gray-100"
                        >
                          北京理工大学
                        </div>
                        <div
                          class="text-sm text-color-gray-500 dark:text-color-gray-400"
                        >
                          软件工程
                        </div>
                      </div>
                      <div
                        class="ml-auto text-color-primary-600 dark:text-color-primary-400 font-medium"
                      >
                        稳妥
                      </div>
                    </div>
                    <div
                      class="flex items-center space-x-3 p-3xl bg-white dark:bg-color-gray-800 rounded-lg"
                    >
                      <div
                        class="w-10 h-10 bg-color-warning-100 dark:bg-color-warning-900 rounded-lg flex items-center justify-center"
                      >
                        <BuildingIcon
                          class="w-5 h-5 text-color-warning-600 dark:text-color-warning-400"
                        />
                      </div>
                      <div>
                        <div
                          class="font-medium text-color-gray-900 dark:text-color-gray-100"
                        >
                          华北电力大学
                        </div>
                        <div
                          class="text-sm text-color-gray-500 dark:text-color-gray-400"
                        >
                          电气工程
                        </div>
                      </div>
                      <div
                        class="ml-auto text-color-warning-600 dark:text-color-warning-400 font-medium"
                      >
                        保底
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div
              class="absolute -top-4 -right-4 w-72 h-72 bg-gradient-to-br from-color-primary-400 to-color-secondary-400 rounded-full opacity-20 blur-3xl"
            ></div>
            <div
              class="absolute -bottom-4 -left-4 w-72 h-72 bg-gradient-to-br from-color-warning-400 to-color-error-400 rounded-full opacity-20 blur-3xl"
            ></div>
          </div>
        </div>
      </div>
    </section>

    <!-- 统计数据 -->
    <section
      class="py-responsive bg-white dark:bg-color-gray-900"
      aria-labelledby="stats-heading"
    >
      <h2 id="stats-heading" class="sr-only">平台统计数据</h2>
      <div class="container">
        <div class="grid-responsive md:grid-cols-4 gap-6xl" role="list">
          <div
            v-for="stat in stats"
            :key="stat.label"
            class="stat-card"
            role="listitem"
          >
            <div class="stat-icon">
              <component
                :is="stat.icon"
                class="w-8 h-8 text-color-primary-600 dark:text-color-primary-400"
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
      class="py-responsive bg-color-gray-50 dark:bg-color-gray-800"
      aria-labelledby="features-heading"
    >
      <div class="container">
        <div class="text-center mb-16">
          <h2 id="features-heading" class="page-title">核心功能</h2>
          <p class="text-xl text-color-gray-600 dark:text-color-gray-300 mt-4">
            全方位的志愿填报解决方案
          </p>
        </div>

        <div class="grid-responsive" role="list">
          <div
            v-for="feature in features"
            :key="feature.title"
            class="base-card base-card--elevated p-8xl group cursor-pointer"
            role="listitem"
            tabindex="0"
            :aria-label="`${feature.title}：${feature.description}，点击了解更多`"
            @click="$router.push(feature.link)"
            @keydown.enter="$router.push(feature.link)"
            @keydown.space.prevent="$router.push(feature.link)"
          >
            <div class="feature-icon mb-6">
              <div
                class="w-16 h-16 bg-color-primary-100 dark:bg-color-primary-900 rounded-2xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300"
              >
                <component
                  :is="feature.icon"
                  class="w-8 h-8 text-color-primary-600 dark:text-color-primary-400"
                  aria-hidden="true"
                />
              </div>
            </div>
            <h3
              class="text-xl font-semibold text-color-gray-900 dark:text-color-gray-100 mb-3"
            >
              {{ feature.title }}
            </h3>
            <p class="text-color-gray-600 dark:text-color-gray-300 mb-6">
              {{ feature.description }}
            </p>
            <div
              class="flex items-center text-color-primary-600 dark:text-color-primary-400 font-medium group-hover:translate-x-2 transition-transform duration-300"
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
    <section class="py-responsive bg-white dark:bg-color-gray-900">
      <div class="container">
        <div class="text-center mb-16">
          <h2 class="page-title">智能推荐流程</h2>
          <p class="text-xl text-color-gray-600 dark:text-color-gray-300 mt-4">
            四步完成个性化志愿方案
          </p>
        </div>

        <div class="grid md:grid-cols-2 lg:grid-cols-4 gap-6xl">
          <div
            v-for="(step, index) in steps"
            :key="step.title"
            class="text-center group"
          >
            <div class="relative mb-6">
              <div
                class="w-20 h-20 bg-gradient-to-br from-color-primary-500 to-color-secondary-500 rounded-full flex items-center justify-center mx-auto group-hover:scale-110 transition-transform duration-300"
              >
                <span class="text-2xl font-bold text-white">{{
                  index + 1
                }}</span>
              </div>
              <div
                v-if="index < steps.length - 1"
                class="hidden lg:block absolute top-10 left-full w-full h-0.5 bg-gradient-to-r from-color-primary-300 to-transparent"
              ></div>
            </div>
            <h3
              class="text-lg font-semibold text-color-gray-900 dark:text-color-gray-100 mb-3"
            >
              {{ step.title }}
            </h3>
            <p class="text-color-gray-600 dark:text-color-gray-300">
              {{ step.description }}
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA区域 -->
    <section
      class="py-responsive bg-gradient-to-br from-color-primary-600 to-color-secondary-600"
      aria-labelledby="cta-heading"
    >
      <div class="container text-center">
        <div class="max-w-3xl mx-auto space-y-8xl">
          <h2
            id="cta-heading"
            class="text-3xl lg:text-4xl font-bold text-white"
          >
            开始您的智能志愿填报之旅
          </h2>
          <p class="text-xl text-color-primary-100">
            让AI为您的未来保驾护航，每一分都不浪费
          </p>
          <div class="hero-section__actions">
            <button
              @click="startRecommendation"
              class="btn btn--primary text-lg bg-white text-color-primary-600 hover:bg-color-gray-100"
              aria-label="立即开始智能推荐"
            >
              立即开始推荐
            </button>
            <button
              @click="viewDemo"
              class="btn btn--ghost border-white text-white hover:bg-white hover:text-color-primary-600 text-lg"
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
/* 文本渐变效果 */
.text-gradient {
  background: linear-gradient(
    135deg,
    var(--color-primary-600),
    var(--color-secondary-600)
  );
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* 动画效果 */
.feature-icon {
  transition: all var(--transition-normal);
}

/* 骨架屏加载样式 */
.stat-skeleton {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
}

.skeleton-line {
  background: linear-gradient(
    90deg,
    var(--color-gray-200) 25%,
    var(--color-gray-300) 50%,
    var(--color-gray-200) 75%
  );
  background-size: 200% 100%;
  border-radius: var(--radius-sm);
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
  background: linear-gradient(
    90deg,
    var(--color-gray-700) 25%,
    var(--color-gray-600) 50%,
    var(--color-gray-700) 75%
  );
  background-size: 200% 100%;
}

/* 响应式优化 */
@media (max-width: 639px) {
  .hero-section__title {
    font-size: var(--font-size-3xl);
  }

  .hero-section__subtitle {
    font-size: var(--font-size-lg);
  }

  .hero-section__actions {
    flex-direction: column;
  }

  .btn {
    width: 100%;
  }
}

@media (min-width: 1024px) {
  .hero-section__actions {
    justify-content: center;
  }
}
</style>
