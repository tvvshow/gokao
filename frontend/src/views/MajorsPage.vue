<template>
  <div class="majors-page min-h-screen bg-gray-50 dark:bg-gray-900">
    <div class="container-modern py-8">
      <div class="page-header">
        <h1 class="page-title">专业分析</h1>
        <p class="page-subtitle">探索热门专业，了解就业前景，做出明智选择</p>
      </div>

      <!-- 搜索筛选 -->
      <div class="search-section">
        <el-card class="content-card filter-panel">
          <el-form :model="searchForm" label-position="top" class="filter-form">
            <el-row :gutter="16">
              <el-col :xs="24" :sm="12" :lg="8">
                <el-form-item label="专业名称">
                  <el-input
                    v-model="searchForm.name"
                    placeholder="请输入专业名称"
                    clearable
                  />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="12" :lg="8">
                <el-form-item label="学科门类">
                  <el-select
                    v-model="searchForm.category"
                    placeholder="选择学科门类"
                    clearable
                  >
                    <el-option
                      v-for="category in categories"
                      :key="category"
                      :label="category"
                      :value="category"
                    />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="12" :lg="8">
                <el-form-item label="学位类型">
                  <el-select
                    v-model="searchForm.degree"
                    placeholder="选择学位类型"
                    clearable
                  >
                    <el-option label="学士" value="学士" />
                    <el-option label="硕士" value="硕士" />
                    <el-option label="博士" value="博士" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
            <div class="filter-actions">
              <el-button type="primary" @click="handleSearch"
                >搜索专业</el-button
              >
              <el-button @click="handleReset">重置条件</el-button>
            </div>
          </el-form>
        </el-card>
      </div>

      <!-- 专业统计 -->
      <div class="stats-section">
        <el-row :gutter="20">
          <el-col :xs="12" :md="6">
            <div class="stat-card">
              <div class="stat-value">{{ totalMajors }}</div>
              <div class="stat-label">专业总数</div>
            </div>
          </el-col>
          <el-col :xs="12" :md="6">
            <div class="stat-card">
              <div class="stat-value">{{ hotMajors }}</div>
              <div class="stat-label">热门专业</div>
            </div>
          </el-col>
          <el-col :xs="12" :md="6">
            <div class="stat-card">
              <div class="stat-value">{{ avgEmploymentRate }}%</div>
              <div class="stat-label">平均就业率</div>
            </div>
          </el-col>
          <el-col :xs="12" :md="6">
            <div class="stat-card">
              <div class="stat-value">{{ avgSalary }}k</div>
              <div class="stat-label">平均薪资</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 专业列表 -->
      <div class="majors-grid content-card results-panel">
        <div class="results-head" v-if="!loading">
          <span>
            已检索到 <strong>{{ total }}</strong> 个专业
          </span>
        </div>

        <!-- Skeleton loading state -->
        <SkeletonList
          v-if="loading"
          :count="pageSize"
          :col-span="{ xs: 24, sm: 12, md: 8, lg: 6 }"
          :lines="4"
          avatar-size="0"
          :show-header="false"
        />

        <!-- Actual content -->
        <el-row v-else :gutter="20">
          <el-col
            :xs="24"
            :sm="12"
            :md="8"
            :lg="6"
            v-for="major in majors"
            :key="major.id"
          >
            <MajorCard :major="major" @view="viewMajorDetail" />
          </el-col>
        </el-row>

        <el-empty
          v-if="!loading && majors.length === 0"
          description="未找到匹配的专业"
        />
      </div>

      <!-- 分页 -->
      <div class="pagination-wrapper" v-if="total > 0">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import MajorCard from '@/components/MajorCard.vue';
import { SkeletonList } from '@/components/common';
import type { Major } from '@/types/university';
import { api } from '@/api/api-client';
import { ElMessage } from 'element-plus';

const router = useRouter();

const searchForm = reactive({
  name: '',
  category: '',
  degree: '',
});

const loading = ref(false);
const majors = ref<Major[]>([]);
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(12);

// 统计数据
const totalMajors = ref(1542);
const hotMajors = ref(156);
const avgEmploymentRate = ref(87.3);
const avgSalary = ref(8.5);

const categories = ref([
  '哲学',
  '经济学',
  '法学',
  '教育学',
  '文学',
  '历史学',
  '理学',
  '工学',
  '农学',
  '医学',
  '管理学',
  '艺术学',
]);

interface MajorSearchParams {
  page: number;
  page_size: number;
  keyword?: string;
  category?: string;
}

const handleSearch = async () => {
  loading.value = true;
  try {
    const params: MajorSearchParams = {
      page: currentPage.value,
      page_size: pageSize.value, // 后端使用 page_size
    };

    if (searchForm.name) {
      params.keyword = searchForm.name;
    }
    if (searchForm.category) {
      params.category = searchForm.category;
    }

    const response = (await api.majors.list(params)) as {
      data: Major[];
      total?: number;
    };
    majors.value = response.data;
    total.value = response.total || response.data.length;

    // 更新统计数据
    totalMajors.value = response.total || response.data.length;

    ElMessage.success(`找到 ${total.value} 个专业`);
  } catch (error) {
    console.error('获取专业数据失败:', error);
    ElMessage.error('获取专业数据失败');
    // 如果API失败，使用模拟数据作为备选
    majors.value = generateMockMajors();
    total.value = 120;
  } finally {
    loading.value = false;
  }
};

const handleReset = () => {
  Object.assign(searchForm, { name: '', category: '', degree: '' });
  handleSearch();
};

const handlePageChange = (page: number) => {
  currentPage.value = page;
  handleSearch();
};

const viewMajorDetail = (majorId: string) => {
  router.push(`/majors/${majorId}`);
};

// 生成模拟数据
const generateMockMajors = (): Major[] => {
  const mockMajors = [
    {
      name: '计算机科学与技术',
      category: '工学',
      employmentRate: 95.2,
      averageSalary: 12.5,
    },
    {
      name: '软件工程',
      category: '工学',
      employmentRate: 94.8,
      averageSalary: 13.2,
    },
    {
      name: '人工智能',
      category: '工学',
      employmentRate: 96.1,
      averageSalary: 15.8,
    },
    {
      name: '数据科学与大数据技术',
      category: '工学',
      employmentRate: 93.5,
      averageSalary: 11.8,
    },
    {
      name: '临床医学',
      category: '医学',
      employmentRate: 89.3,
      averageSalary: 9.2,
    },
    {
      name: '金融学',
      category: '经济学',
      employmentRate: 91.7,
      averageSalary: 10.5,
    },
  ];

  return mockMajors.map((major, index) => ({
    id: `major_${index + 1}`,
    name: major.name,
    code: `${index + 1}`.padStart(6, '0'),
    category: major.category,
    degree: '学士',
    duration: 4,
    employmentRate: major.employmentRate,
    averageSalary: major.averageSalary,
    isPopular: major.averageSalary > 12,
    description: `${major.name}是一个充满挑战和机遇的专业...`,
  }));
};

onMounted(() => {
  handleSearch();
});
</script>

<style scoped>
.majors-page {
  min-height: calc(100vh - 160px);
}

.page-header {
  text-align: center;
  margin-bottom: 34px;
}

.page-title {
  font-size: 32px;
  color: #0f172a;
  letter-spacing: -0.02em;
  margin-bottom: 12px;
}

.page-subtitle {
  color: #64748b;
  font-size: 16px;
}

.search-section {
  margin-bottom: 24px;
}

.filter-panel {
  border-radius: 14px;
  border: 1px solid rgb(148 163 184 / 0.24);
  background: linear-gradient(
    180deg,
    rgb(255 255 255 / 0.95),
    rgb(248 250 252 / 0.95)
  );
  box-shadow: 0 16px 30px -28px rgb(14 165 233 / 0.5);
}

.filter-form :deep(.el-form-item__label) {
  margin-bottom: 6px;
  font-size: 0.9375rem;
  font-weight: 600;
  color: #334155;
}

.filter-form :deep(.el-input__wrapper),
.filter-form :deep(.el-select__wrapper) {
  min-height: 46px;
  border: 1px solid #dbe3ef;
  border-radius: 12px;
  box-shadow: none;
}

.filter-form :deep(.el-input__inner),
.filter-form :deep(.el-select__selected-item),
.filter-form :deep(.el-select__placeholder) {
  font-size: 15px;
}

.filter-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: flex-end;
  padding-top: 2px;
}

.filter-actions :deep(.el-button) {
  min-height: 40px;
  border-radius: 10px;
}

.stats-section {
  margin-bottom: 26px;
}

.stat-card {
  text-align: center;
  padding: 20px 12px;
  background: rgb(255 255 255 / 0.94);
  border: 1px solid rgb(148 163 184 / 0.22);
  border-radius: 12px;
  box-shadow: 0 10px 24px -24px rgb(15 23 42 / 0.5);
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #0ea5e9;
  margin-bottom: 8px;
}

.stat-label {
  color: #64748b;
  font-size: 14px;
}

.majors-grid {
  margin-bottom: 30px;
}

.results-panel {
  padding: 18px;
  border-radius: 14px;
  border: 1px solid rgb(148 163 184 / 0.22);
}

.results-head {
  margin: 2px 0 14px;
  font-size: 14px;
  color: #475569;
}

.results-head strong {
  color: #0284c7;
}

.majors-grid .el-col {
  margin-bottom: 20px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
}

.dark .page-title {
  color: #f1f5f9;
}

.dark .page-subtitle {
  color: #94a3b8;
}

.dark .filter-panel,
.dark .results-panel {
  border-color: rgb(71 85 105 / 0.45);
  box-shadow: none;
}

.dark .filter-panel {
  background: linear-gradient(180deg, rgb(31 41 55 / 0.9), rgb(17 24 39 / 0.9));
}

.dark .filter-form :deep(.el-form-item__label) {
  color: #cbd5e1;
}

.dark .filter-form :deep(.el-input__wrapper),
.dark .filter-form :deep(.el-select__wrapper) {
  border-color: #334155;
}

.dark .stat-card {
  background: rgb(31 41 55 / 0.9);
  border-color: rgb(71 85 105 / 0.5);
}

.dark .stat-label,
.dark .results-head {
  color: #94a3b8;
}

.dark .results-head strong {
  color: #67e8f9;
}

@media (max-width: 768px) {
  .page-title {
    font-size: 28px;
  }

  .filter-actions {
    justify-content: stretch;
  }

  .filter-actions :deep(.el-button) {
    flex: 1;
  }

  .filter-form :deep(.el-input__wrapper),
  .filter-form :deep(.el-select__wrapper) {
    min-height: 48px;
  }

  .filter-form :deep(.el-input__inner),
  .filter-form :deep(.el-select__selected-item),
  .filter-form :deep(.el-select__placeholder) {
    font-size: 16px;
  }
}
</style>
