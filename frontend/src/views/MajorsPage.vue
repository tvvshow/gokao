<template>
  <div class="majors-page">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">专业分析</h1>
        <p class="page-subtitle">探索热门专业，了解就业前景，做出明智选择</p>
      </div>

      <!-- 搜索筛选 -->
      <div class="search-section">
        <el-card class="content-card">
          <el-form :model="searchForm" :inline="true">
            <el-form-item label="专业名称">
              <el-input
                v-model="searchForm.name"
                placeholder="请输入专业名称"
                clearable
                style="width: 200px"
              />
            </el-form-item>
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
            <el-form-item>
              <el-button type="primary" @click="handleSearch">搜索</el-button>
              <el-button @click="handleReset">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </div>

      <!-- 专业统计 -->
      <div class="stats-section">
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-value">{{ totalMajors }}</div>
              <div class="stat-label">专业总数</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-value">{{ hotMajors }}</div>
              <div class="stat-label">热门专业</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-value">{{ avgEmploymentRate }}%</div>
              <div class="stat-label">平均就业率</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-value">{{ avgSalary }}k</div>
              <div class="stat-label">平均薪资</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 专业列表 -->
      <div class="majors-grid">
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
  padding: 20px 0;
  min-height: calc(100vh - 160px);
}

.container {
  max-width: 1200px;
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

.search-section {
  margin-bottom: 24px;
}

.stats-section {
  margin-bottom: 30px;
}

.stat-card {
  text-align: center;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #667eea;
  margin-bottom: 8px;
}

.stat-label {
  color: #7f8c8d;
  font-size: 14px;
}

.majors-grid {
  margin-bottom: 30px;
}

.majors-grid .el-col {
  margin-bottom: 20px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
}
</style>
