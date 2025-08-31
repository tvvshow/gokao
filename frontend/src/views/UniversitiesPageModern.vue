<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <div class="container-modern py-8">
      <!-- 页面头部 -->
      <div class="text-center mb-12">
        <h1 class="page-title">院校查询</h1>
        <p class="text-xl text-gray-600 dark:text-gray-300 mt-4">
          探索全国2700+优质高等院校，找到最适合你的大学
        </p>
      </div>

      <!-- 搜索和筛选区域 -->
      <div class="card p-6 mb-8">
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300">院校名称</label>
            <div class="relative">
              <SearchIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
              <input
                v-model="searchForm.name"
                type="text"
                placeholder="请输入院校名称"
                class="input pl-10"
                @keyup.enter="handleSearch"
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300">所在省份</label>
            <select v-model="searchForm.province" class="input">
              <option value="">选择省份</option>
              <option v-for="province in provinces" :key="province" :value="province">
                {{ province }}
              </option>
            </select>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300">院校类型</label>
            <select v-model="searchForm.type" class="input">
              <option value="">选择类型</option>
              <option value="综合类">综合类</option>
              <option value="理工类">理工类</option>
              <option value="师范类">师范类</option>
              <option value="财经类">财经类</option>
              <option value="医药类">医药类</option>
              <option value="艺术类">艺术类</option>
            </select>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300">院校层次</label>
            <select v-model="searchForm.level" class="input">
              <option value="">选择层次</option>
              <option value="985">985工程</option>
              <option value="211">211工程</option>
              <option value="双一流">双一流</option>
              <option value="本科">普通本科</option>
            </select>
          </div>
        </div>

        <div class="flex justify-center mt-6 space-x-4">
          <button @click="handleSearch" class="btn btn-primary" :disabled="loading">
            <SearchIcon class="w-4 h-4 mr-2" />
            {{ loading ? '搜索中...' : '搜索院校' }}
          </button>
          <button @click="resetSearch" class="btn btn-secondary">
            <RefreshCwIcon class="w-4 h-4 mr-2" />
            重置条件
          </button>
        </div>
      </div>

      <!-- 搜索结果 -->
      <div v-if="loading" class="text-center py-12">
        <div class="loading-spinner mx-auto mb-4"></div>
        <p class="text-gray-500 dark:text-gray-400">正在搜索院校信息...</p>
      </div>

      <div v-else-if="universities.length === 0 && hasSearched" class="text-center py-12">
        <BuildingIcon class="w-16 h-16 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
        <p class="text-gray-500 dark:text-gray-400">未找到符合条件的院校</p>
        <button @click="resetSearch" class="btn btn-primary mt-4">
          重新搜索
        </button>
      </div>

      <div v-else class="space-y-6">
        <!-- 结果统计 -->
        <div class="flex items-center justify-between">
          <p class="text-gray-600 dark:text-gray-300">
            找到 <span class="font-semibold text-primary-600">{{ universities.length }}</span> 所院校
          </p>
          <div class="flex items-center space-x-2">
            <span class="text-sm text-gray-500 dark:text-gray-400">排序方式:</span>
            <select v-model="sortBy" @change="handleSort" class="input text-sm">
              <option value="name">院校名称</option>
              <option value="level">院校层次</option>
              <option value="province">所在省份</option>
            </select>
          </div>
        </div>

        <!-- 院校列表 -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div
            v-for="university in paginatedUniversities"
            :key="university.id"
            class="card card-hover p-6 cursor-pointer"
            @click="viewUniversityDetail(university)"
          >
            <div class="flex items-start space-x-4">
              <div class="w-16 h-16 bg-gradient-to-br from-primary-500 to-secondary-500 rounded-xl flex items-center justify-center flex-shrink-0">
                <BuildingIcon class="w-8 h-8 text-white" />
              </div>
              
              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between mb-2">
                  <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 truncate">
                    {{ university.name }}
                  </h3>
                  <div class="flex space-x-1">
                    <span v-if="university.is985" class="badge badge-error text-xs">985</span>
                    <span v-if="university.is211" class="badge badge-warning text-xs">211</span>
                    <span v-if="university.isDoubleFirstClass" class="badge badge-primary text-xs">双一流</span>
                  </div>
                </div>
                
                <div class="space-y-2 text-sm text-gray-600 dark:text-gray-300">
                  <div class="flex items-center">
                    <MapPinIcon class="w-4 h-4 mr-2" />
                    {{ university.province }} · {{ university.city }}
                  </div>
                  <div class="flex items-center">
                    <TagIcon class="w-4 h-4 mr-2" />
                    {{ university.type }}
                  </div>
                  <div class="flex items-center">
                    <UsersIcon class="w-4 h-4 mr-2" />
                    在校生 {{ university.studentCount || '未知' }} 人
                  </div>
                </div>
                
                <div class="mt-4 flex items-center justify-between">
                  <div class="text-sm text-gray-500 dark:text-gray-400">
                    {{ university.description || '暂无简介' }}
                  </div>
                  <ArrowRightIcon class="w-4 h-4 text-gray-400" />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 分页 -->
        <div v-if="totalPages > 1" class="flex justify-center mt-8">
          <div class="flex items-center space-x-2">
            <button
              @click="currentPage = Math.max(1, currentPage - 1)"
              :disabled="currentPage === 1"
              class="btn btn-secondary"
            >
              <ChevronLeftIcon class="w-4 h-4" />
            </button>
            
            <span class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300">
              第 {{ currentPage }} 页，共 {{ totalPages }} 页
            </span>
            
            <button
              @click="currentPage = Math.min(totalPages, currentPage + 1)"
              :disabled="currentPage === totalPages"
              class="btn btn-secondary"
            >
              <ChevronRightIcon class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  SearchIcon,
  RefreshCwIcon,
  BuildingIcon,
  MapPinIcon,
  TagIcon,
  UsersIcon,
  ArrowRightIcon,
  ChevronLeftIcon,
  ChevronRightIcon
} from 'lucide-vue-next'

const router = useRouter()

// 响应式数据
const loading = ref(false)
const hasSearched = ref(false)
const currentPage = ref(1)
const pageSize = 10
const sortBy = ref('name')

// 搜索表单
const searchForm = ref({
  name: '',
  province: '',
  type: '',
  level: ''
})

// 省份列表
const provinces = ref([
  '北京', '上海', '天津', '重庆', '河北', '山西', '辽宁', '吉林', '黑龙江',
  '江苏', '浙江', '安徽', '福建', '江西', '山东', '河南', '湖北', '湖南',
  '广东', '广西', '海南', '四川', '贵州', '云南', '西藏', '陕西', '甘肃',
  '青海', '宁夏', '新疆', '内蒙古', '台湾', '香港', '澳门'
])

// 院校数据
const universities = ref([
  {
    id: 1,
    name: '清华大学',
    province: '北京',
    city: '北京',
    type: '综合类',
    is985: true,
    is211: true,
    isDoubleFirstClass: true,
    studentCount: 50000,
    description: '中国顶尖综合性研究型大学'
  },
  {
    id: 2,
    name: '北京大学',
    province: '北京',
    city: '北京',
    type: '综合类',
    is985: true,
    is211: true,
    isDoubleFirstClass: true,
    studentCount: 45000,
    description: '中国最高学府之一'
  },
  {
    id: 3,
    name: '复旦大学',
    province: '上海',
    city: '上海',
    type: '综合类',
    is985: true,
    is211: true,
    isDoubleFirstClass: true,
    studentCount: 32000,
    description: '享誉海内外的综合性研究型大学'
  }
])

// 计算属性
const filteredUniversities = computed(() => {
  let result = [...universities.value]
  
  if (searchForm.value.name) {
    result = result.filter(u => u.name.includes(searchForm.value.name))
  }
  
  if (searchForm.value.province) {
    result = result.filter(u => u.province === searchForm.value.province)
  }
  
  if (searchForm.value.type) {
    result = result.filter(u => u.type === searchForm.value.type)
  }
  
  if (searchForm.value.level) {
    const level = searchForm.value.level
    result = result.filter(u => {
      if (level === '985') return u.is985
      if (level === '211') return u.is211
      if (level === '双一流') return u.isDoubleFirstClass
      return true
    })
  }
  
  return result
})

const sortedUniversities = computed(() => {
  const result = [...filteredUniversities.value]
  
  result.sort((a, b) => {
    switch (sortBy.value) {
      case 'name':
        return a.name.localeCompare(b.name)
      case 'level':
        if (a.is985 && !b.is985) return -1
        if (!a.is985 && b.is985) return 1
        if (a.is211 && !b.is211) return -1
        if (!a.is211 && b.is211) return 1
        return 0
      case 'province':
        return a.province.localeCompare(b.province)
      default:
        return 0
    }
  })
  
  return result
})

const totalPages = computed(() => Math.ceil(sortedUniversities.value.length / pageSize))

const paginatedUniversities = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  const end = start + pageSize
  return sortedUniversities.value.slice(start, end)
})

// 方法
const handleSearch = async () => {
  loading.value = true
  hasSearched.value = true
  currentPage.value = 1
  
  // 模拟API调用
  await new Promise(resolve => setTimeout(resolve, 1000))
  
  loading.value = false
}

const resetSearch = () => {
  searchForm.value = {
    name: '',
    province: '',
    type: '',
    level: ''
  }
  hasSearched.value = false
  currentPage.value = 1
}

const handleSort = () => {
  currentPage.value = 1
}

const viewUniversityDetail = (university: any) => {
  router.push(`/universities/${university.id}`)
}

// 生命周期
onMounted(() => {
  // 初始化数据
})
</script>

<style scoped>
/* 自定义样式 */
</style>
