<template>
  <div class="min-h-screen">
    <!-- 现代化英雄区域 -->
    <section class="relative overflow-hidden bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50 dark:from-gray-900 dark:via-blue-900 dark:to-indigo-900">
      <div class="absolute inset-0 bg-grid-pattern opacity-5"></div>
      <div class="container-modern py-20 lg:py-32">
        <div class="grid lg:grid-cols-2 gap-12 items-center">
          <div class="space-y-8 animate-fade-in">
            <div class="space-y-4">
              <div class="inline-flex items-center px-4 py-2 bg-primary-100 dark:bg-primary-900 rounded-full text-primary-700 dark:text-primary-300 text-sm font-medium">
                <SparklesIcon class="w-4 h-4 mr-2" />
                AI智能推荐系统
              </div>
              <h1 class="text-4xl lg:text-6xl font-bold text-gradient leading-tight">
                智能高考<br>志愿填报
              </h1>
              <p class="text-xl text-gray-600 dark:text-gray-300 leading-relaxed">
                基于大数据分析和AI算法，为您提供个性化的志愿填报方案，让每一分都发挥最大价值
              </p>
            </div>
            
            <div class="flex flex-col sm:flex-row gap-4">
              <button @click="startRecommendation" class="btn btn-primary text-lg px-8 py-4">
                <SparklesIcon class="w-5 h-5 mr-2" />
                开始智能推荐
              </button>
              <button @click="viewColleges" class="btn btn-secondary text-lg px-8 py-4">
                <BuildingIcon class="w-5 h-5 mr-2" />
                查看院校信息
              </button>
            </div>
            
            <div class="flex items-center space-x-6 text-sm text-gray-500 dark:text-gray-400">
              <div class="flex items-center">
                <CheckCircleIcon class="w-4 h-4 mr-2 text-green-500" />
                2700+ 高校数据
              </div>
              <div class="flex items-center">
                <CheckCircleIcon class="w-4 h-4 mr-2 text-green-500" />
                1400+ 专业信息
              </div>
              <div class="flex items-center">
                <CheckCircleIcon class="w-4 h-4 mr-2 text-green-500" />
                AI智能分析
              </div>
            </div>
          </div>
          
          <div class="relative animate-float">
            <div class="relative z-10">
              <div class="card p-8 bg-gradient-to-br from-white to-blue-50 dark:from-gray-800 dark:to-blue-900 shadow-glow">
                <div class="space-y-6">
                  <div class="flex items-center justify-between">
                    <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">推荐结果预览</h3>
                    <div class="badge badge-success">匹配度 95%</div>
                  </div>
                  <div class="space-y-4">
                    <div class="flex items-center space-x-3 p-3 bg-white dark:bg-gray-700 rounded-lg">
                      <div class="w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center">
                        <BuildingIcon class="w-5 h-5 text-blue-600 dark:text-blue-400" />
                      </div>
                      <div>
                        <div class="font-medium text-gray-900 dark:text-gray-100">清华大学</div>
                        <div class="text-sm text-gray-500 dark:text-gray-400">计算机科学与技术</div>
                      </div>
                      <div class="ml-auto text-green-600 dark:text-green-400 font-medium">冲刺</div>
                    </div>
                    <div class="flex items-center space-x-3 p-3 bg-white dark:bg-gray-700 rounded-lg">
                      <div class="w-10 h-10 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center">
                        <BuildingIcon class="w-5 h-5 text-green-600 dark:text-green-400" />
                      </div>
                      <div>
                        <div class="font-medium text-gray-900 dark:text-gray-100">北京理工大学</div>
                        <div class="text-sm text-gray-500 dark:text-gray-400">软件工程</div>
                      </div>
                      <div class="ml-auto text-blue-600 dark:text-blue-400 font-medium">稳妥</div>
                    </div>
                    <div class="flex items-center space-x-3 p-3 bg-white dark:bg-gray-700 rounded-lg">
                      <div class="w-10 h-10 bg-orange-100 dark:bg-orange-900 rounded-lg flex items-center justify-center">
                        <BuildingIcon class="w-5 h-5 text-orange-600 dark:text-orange-400" />
                      </div>
                      <div>
                        <div class="font-medium text-gray-900 dark:text-gray-100">华北电力大学</div>
                        <div class="text-sm text-gray-500 dark:text-gray-400">电气工程</div>
                      </div>
                      <div class="ml-auto text-orange-600 dark:text-orange-400 font-medium">保底</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div class="absolute -top-4 -right-4 w-72 h-72 bg-gradient-to-br from-primary-400 to-secondary-400 rounded-full opacity-20 blur-3xl"></div>
            <div class="absolute -bottom-4 -left-4 w-72 h-72 bg-gradient-to-br from-accent-400 to-error-400 rounded-full opacity-20 blur-3xl"></div>
          </div>
        </div>
      </div>
    </section>

    <!-- 现代化统计数据 -->
    <section class="py-20 bg-white dark:bg-gray-900">
      <div class="container-modern">
        <div class="grid grid-cols-2 md:grid-cols-4 gap-8">
          <div v-for="stat in stats" :key="stat.label" class="stat-card">
            <div class="stat-icon">
              <component :is="stat.icon" class="w-8 h-8 text-primary-600 dark:text-primary-400" />
            </div>
            <div class="stat-number">{{ stat.value }}</div>
            <div class="stat-label">{{ stat.label }}</div>
          </div>
        </div>
      </div>
    </section>

    <!-- 现代化功能特色 -->
    <section class="py-20 bg-gray-50 dark:bg-gray-800">
      <div class="container-modern">
        <div class="text-center mb-16">
          <h2 class="page-title">核心功能</h2>
          <p class="text-xl text-gray-600 dark:text-gray-300 mt-4">
            全方位的志愿填报解决方案
          </p>
        </div>
        
        <div class="grid-modern">
          <div v-for="feature in features" :key="feature.title" 
               class="card card-interactive p-8 group cursor-pointer"
               @click="$router.push(feature.link)">
            <div class="feature-icon mb-6">
              <div class="w-16 h-16 bg-primary-100 dark:bg-primary-900 rounded-2xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                <component :is="feature.icon" class="w-8 h-8 text-primary-600 dark:text-primary-400" />
              </div>
            </div>
            <h3 class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-3">
              {{ feature.title }}
            </h3>
            <p class="text-gray-600 dark:text-gray-300 mb-6">
              {{ feature.description }}
            </p>
            <div class="flex items-center text-primary-600 dark:text-primary-400 font-medium group-hover:translate-x-2 transition-transform duration-300">
              了解更多
              <ArrowRightIcon class="w-4 h-4 ml-2" />
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 现代化推荐流程 -->
    <section class="py-20 bg-white dark:bg-gray-900">
      <div class="container-modern">
        <div class="text-center mb-16">
          <h2 class="page-title">智能推荐流程</h2>
          <p class="text-xl text-gray-600 dark:text-gray-300 mt-4">
            四步完成个性化志愿方案
          </p>
        </div>
        
        <div class="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
          <div v-for="(step, index) in steps" :key="step.title" class="text-center group">
            <div class="relative mb-6">
              <div class="w-20 h-20 bg-gradient-to-br from-primary-500 to-secondary-500 rounded-full flex items-center justify-center mx-auto group-hover:scale-110 transition-transform duration-300">
                <span class="text-2xl font-bold text-white">{{ index + 1 }}</span>
              </div>
              <div v-if="index < steps.length - 1" class="hidden lg:block absolute top-10 left-full w-full h-0.5 bg-gradient-to-r from-primary-300 to-transparent"></div>
            </div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">
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
    <section class="py-20 bg-gradient-to-br from-primary-600 to-secondary-600">
      <div class="container-modern text-center">
        <div class="max-w-3xl mx-auto space-y-8">
          <h2 class="text-3xl lg:text-4xl font-bold text-white">
            开始您的智能志愿填报之旅
          </h2>
          <p class="text-xl text-primary-100">
            让AI为您的未来保驾护航，每一分都不浪费
          </p>
          <div class="flex flex-col sm:flex-row gap-4 justify-center">
            <button @click="startRecommendation" class="btn bg-white text-primary-600 hover:bg-gray-100 text-lg px-8 py-4">
              立即开始推荐
            </button>
            <button @click="viewDemo" class="btn btn-ghost border-white text-white hover:bg-white hover:text-primary-600 text-lg px-8 py-4">
              查看演示
            </button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  SparklesIcon,
  BuildingIcon,
  CheckCircleIcon,
  ArrowRightIcon,
  UsersIcon,
  BookOpenIcon,
  BarChartIcon,
  CrownIcon
} from 'lucide-vue-next'

const router = useRouter()

// 统计数据
const stats = ref([
  { icon: BuildingIcon, value: '2700+', label: '合作高校' },
  { icon: BookOpenIcon, value: '1400+', label: '专业数据' },
  { icon: UsersIcon, value: '50万+', label: '服务学生' },
  { icon: BarChartIcon, value: '95%', label: '推荐准确率' }
])

// 功能特色
const features = ref([
  {
    icon: SparklesIcon,
    title: 'AI智能推荐',
    description: '基于大数据分析，为您量身定制最优志愿方案',
    link: '/recommendation'
  },
  {
    icon: BuildingIcon,
    title: '院校查询',
    description: '2700+高校详细信息，历年分数线一目了然',
    link: '/universities'
  },
  {
    icon: BookOpenIcon,
    title: '专业分析',
    description: '1400+专业深度解析，就业前景全面了解',
    link: '/majors'
  },
  {
    icon: BarChartIcon,
    title: '数据分析',
    description: '多维度数据可视化，科学决策有依据',
    link: '/analysis'
  },
  {
    icon: CrownIcon,
    title: '会员服务',
    description: '专属顾问一对一指导，VIP专享服务',
    link: '/membership'
  },
  {
    icon: UsersIcon,
    title: '模拟填报',
    description: '真实模拟填报环境，提前体验填报流程',
    link: '/simulation'
  }
])

// 推荐流程
const steps = ref([
  {
    title: '输入成绩',
    description: '填写高考成绩和基本信息'
  },
  {
    title: 'AI分析',
    description: '智能算法分析匹配度'
  },
  {
    title: '生成方案',
    description: '个性化志愿填报方案'
  },
  {
    title: '优化调整',
    description: '根据偏好微调完善'
  }
])

// 方法
const startRecommendation = () => {
  router.push('/recommendation')
}

const viewColleges = () => {
  router.push('/universities')
}

const viewDemo = () => {
  // 演示功能
  console.log('查看演示')
}
</script>

<style scoped>
.bg-grid-pattern {
  background-image: radial-gradient(circle, rgba(0, 0, 0, 0.1) 1px, transparent 1px);
  background-size: 20px 20px;
}

.feature-icon {
  @apply transition-all duration-300;
}
</style>
