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
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300"
              >院校名称</label
            >
            <div class="relative">
              <SearchIcon
                class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400"
              />
              <input
                v-model="searchForm.name"
                type="text"
                placeholder="请输入院校名称"
                class="input pl-10"
                @keyup.enter="() => handleSearch()"
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300"
              >所在省份</label
            >
            <select v-model="searchForm.province" class="input">
              <option value="">选择省份</option>
              <option
                v-for="province in provinces"
                :key="province"
                :value="province"
              >
                {{ province }}
              </option>
            </select>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300"
              >院校类型</label
            >
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
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300"
              >院校层次</label
            >
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
          <button
            @click="() => handleSearch()"
            class="btn btn-primary"
            :disabled="loading"
          >
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

      <div
        v-else-if="universities.length === 0 && hasSearched"
        class="text-center py-12"
      >
        <BuildingIcon
          class="w-16 h-16 text-gray-300 dark:text-gray-600 mx-auto mb-4"
        />
        <p class="text-gray-500 dark:text-gray-400">未找到符合条件的院校</p>
        <button @click="resetSearch" class="btn btn-primary mt-4">
          重新搜索
        </button>
      </div>

      <div v-else class="space-y-6">
        <!-- 结果统计 -->
        <div class="flex items-center justify-between">
          <p class="text-gray-600 dark:text-gray-300">
            找到
            <span class="font-semibold text-primary-600">{{
              universities.length
            }}</span>
            所院校
          </p>
          <div class="flex items-center space-x-2">
            <span class="text-sm text-gray-500 dark:text-gray-400"
              >排序方式:</span
            >
            <select v-model="sortBy" @change="handleSort" class="input text-sm">
              <option value="name">院校名称</option>
              <option value="level">院校层次</option>
              <option value="province">所在省份</option>
            </select>
          </div>
        </div>

        <!-- 院校列表 - 使用虚拟滚动优化大列表性能 -->
        <VirtualList
          v-if="paginatedUniversities.length > 100"
          :items="paginatedUniversities"
          :item-height="220"
          container-height="800px"
          key-field="id"
          aria-label="院校列表"
          class="university-virtual-list"
        >
          <template #default="{ item: university }">
            <div
              class="card card-hover p-6 cursor-pointer university-card-item"
              @click="viewUniversityDetail(university)"
            >
              <div class="flex items-start space-x-4">
                <div
                  class="w-16 h-16 bg-gradient-to-br from-primary-500 to-secondary-500 rounded-xl flex items-center justify-center flex-shrink-0"
                >
                  <BuildingIcon class="w-8 h-8 text-white" />
                </div>

                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between mb-2">
                    <h3
                      class="text-lg font-semibold text-gray-900 dark:text-gray-100 truncate"
                    >
                      {{ university.name }}
                    </h3>
                    <div class="flex space-x-1">
                      <span
                        v-if="university.is985"
                        class="badge badge-error text-xs"
                        >985</span
                      >
                      <span
                        v-if="university.is211"
                        class="badge badge-warning text-xs"
                        >211</span
                      >
                      <span
                        v-if="university.isDoubleFirstClass"
                        class="badge badge-primary text-xs"
                        >双一流</span
                      >
                    </div>
                  </div>

                  <div
                    class="space-y-2 text-sm text-gray-600 dark:text-gray-300"
                  >
                    <div class="flex items-center">
                      <MapPinIcon class="w-4 h-4 mr-2" />
                      {{ university.province }} · {{ university.city }}
                    </div>
                    <div class="flex items-center">
                      <TagIcon class="w-4 h-4 mr-2" />
                      {{ university.type }}
                    </div>
                  </div>

                  <div class="mt-4 flex items-center justify-between">
                    <div
                      class="text-sm text-gray-500 dark:text-gray-400 truncate"
                    >
                      {{ university.description || '暂无简介' }}
                    </div>
                    <ArrowRightIcon
                      class="w-4 h-4 text-gray-400 flex-shrink-0"
                    />
                  </div>
                </div>
              </div>
            </div>
          </template>
        </VirtualList>

        <!-- 少量数据时使用普通列表 -->
        <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div
            v-for="university in paginatedUniversities"
            :key="university.id"
            class="card card-hover p-6 cursor-pointer"
            @click="viewUniversityDetail(university)"
          >
            <div class="flex items-start space-x-4">
              <div
                class="w-16 h-16 bg-gradient-to-br from-primary-500 to-secondary-500 rounded-xl flex items-center justify-center flex-shrink-0"
              >
                <BuildingIcon class="w-8 h-8 text-white" />
              </div>

              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between mb-2">
                  <h3
                    class="text-lg font-semibold text-gray-900 dark:text-gray-100 truncate"
                  >
                    {{ university.name }}
                  </h3>
                  <div class="flex space-x-1">
                    <span
                      v-if="university.is985"
                      class="badge badge-error text-xs"
                      >985</span
                    >
                    <span
                      v-if="university.is211"
                      class="badge badge-warning text-xs"
                      >211</span
                    >
                    <span
                      v-if="university.isDoubleFirstClass"
                      class="badge badge-primary text-xs"
                      >双一流</span
                    >
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
                  <!-- 新增关键信息展示 -->
                  <div v-if="university.rank" class="flex items-center">
                    <span class="w-4 h-4 mr-2 text-center font-bold">🏆</span>
                    全国排名: {{ university.rank }}名
                  </div>
                  <div
                    v-if="university.employmentRate"
                    class="flex items-center"
                  >
                    <span class="w-4 h-4 mr-2 text-center font-bold">📈</span>
                    就业率: {{ university.employmentRate }}%
                  </div>
                  <div
                    v-if="
                      university.strongMajors &&
                      university.strongMajors.length > 0
                    "
                    class="flex items-start"
                  >
                    <span class="w-4 h-4 mr-2 text-center font-bold mt-0.5"
                      >🎯</span
                    >
                    <div>
                      <span class="font-medium">优势专业:</span>
                      <div class="flex flex-wrap gap-1 mt-1">
                        <span
                          v-for="(
                            major, index
                          ) in university.strongMajors.slice(0, 2)"
                          :key="index"
                          class="badge badge-info text-xs"
                        >
                          {{ major }}
                        </span>
                        <span
                          v-if="university.strongMajors.length > 2"
                          class="text-xs text-gray-500"
                        >
                          等{{ university.strongMajors.length - 2 }}个
                        </span>
                      </div>
                    </div>
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
              @click="handlePageChange(currentPage - 1)"
              :disabled="currentPage === 1"
              class="btn btn-secondary"
            >
              <ChevronLeftIcon class="w-4 h-4" />
            </button>

            <span class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300">
              第 {{ currentPage }} 页，共 {{ totalPages }} 页
            </span>

            <button
              @click="handlePageChange(currentPage + 1)"
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
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import {
  SearchIcon,
  RefreshCwIcon,
  BuildingIcon,
  MapPinIcon,
  TagIcon,
  UsersIcon,
  ArrowRightIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
} from 'lucide-vue-next';
import { universityApi } from '@/api/university';
import { VirtualList } from '@/components/common';
import type { University } from '@/types/university';
import { UNIVERSITY_LEVEL_MAP, UNIVERSITY_TYPE_MAP } from '@/types/university';
import type { UniversitySearchParams } from '@/types/api';
import { ElMessage } from 'element-plus';

const router = useRouter();

// 响应式数据
const loading = ref(false);
const hasSearched = ref(false);
const currentPage = ref(1);
const pageSize = 10;
const sortBy = ref('');
const sortOrder = ref<'asc' | 'desc'>('asc');
const totalCount = ref(0);

// 搜索表单
const searchForm = ref({
  name: '',
  province: '',
  type: '',
  level: '',
});

// 省份列表
const provinces = ref([
  '北京',
  '上海',
  '天津',
  '重庆',
  '河北',
  '山西',
  '辽宁',
  '吉林',
  '黑龙江',
  '江苏',
  '浙江',
  '安徽',
  '福建',
  '江西',
  '山东',
  '河南',
  '湖北',
  '湖南',
  '广东',
  '广西',
  '海南',
  '四川',
  '贵州',
  '云南',
  '西藏',
  '陕西',
  '甘肃',
  '青海',
  '宁夏',
  '新疆',
  '内蒙古',
  '台湾',
  '香港',
  '澳门',
]);

// 院校数据
const universities = ref<University[]>([]);

// 计算属性

const totalPages = computed(() => Math.ceil(totalCount.value / pageSize));

// 后端已分页，直接使用返回数据
const paginatedUniversities = computed(() => universities.value);

// 方法
const handleSearch = async (resetPage = true) => {
  loading.value = true;
  hasSearched.value = true;
  if (resetPage) {
    currentPage.value = 1;
  }

  try {
    let response;

    // 如果有搜索关键词，使用搜索API；否则使用列表API
    if (searchForm.value.name && searchForm.value.name.trim()) {
      // 构建后端查询参数（映射枚举值）
      const params: Record<string, unknown> = {
        page: currentPage.value,
        page_size: pageSize,
        keyword: searchForm.value.name.trim(),
      };

      // 添加排序参数
      if (sortBy.value) {
        params.sort_by = sortBy.value;
      }
      if (sortOrder.value) {
        params.sort_order = sortOrder.value;
      }

      if (searchForm.value.province) {
        params.province = searchForm.value.province;
      }
      // 映射院校类型（中文→英文）
      if (searchForm.value.type) {
        params.type = UNIVERSITY_TYPE_MAP[searchForm.value.type] || searchForm.value.type;
      }
      // 映射院校层次（中文→英文）
      if (searchForm.value.level) {
        params.level = UNIVERSITY_LEVEL_MAP[searchForm.value.level] || searchForm.value.level;
      }

      response = await universityApi.search(params as UniversitySearchParams);
    } else {
      // 没有搜索关键词，使用列表API
      const params: Record<string, unknown> = {
        page: currentPage.value,
        page_size: pageSize,
      };

      if (searchForm.value.province) {
        params.province = searchForm.value.province;
      }
      if (searchForm.value.type) {
        params.type = UNIVERSITY_TYPE_MAP[searchForm.value.type] || searchForm.value.type;
      }
      if (searchForm.value.level) {
        params.level = UNIVERSITY_LEVEL_MAP[searchForm.value.level] || searchForm.value.level;
      }

      response = await universityApi.list(params);
    }

    if (response.success && response.data) {
      universities.value = response.data.universities || [];
      totalCount.value = response.data.total || 0;
    } else {
      universities.value = [];
      totalCount.value = 0;
    }

    ElMessage.success(`找到 ${totalCount.value} 所院校`);
  } catch (error) {
    console.error('搜索院校失败:', error);
    ElMessage.error('搜索院校失败，请稍后重试');
  } finally {
    loading.value = false;
  }
};

const resetSearch = () => {
  searchForm.value = {
    name: '',
    province: '',
    type: '',
    level: '',
  };
  hasSearched.value = false;
  currentPage.value = 1;
  universities.value = [];
};

const handleSort = () => {
  currentPage.value = 1;
};

const handlePageChange = (newPage: number) => {
  currentPage.value = newPage;
  handleSearch(false);
};

const viewUniversityDetail = (university: University) => {
  router.push(`/universities/${university.id}`);
};

// 生命周期
onMounted(() => {
  // 初始化数据 - 加载热门院校
  handleSearch();
});
</script>

<style scoped>
/* 自定义样式 */
.university-virtual-list {
  border-radius: 8px;
}

.university-card-item {
  margin-bottom: 16px;
  margin-right: 8px;
}
</style>
