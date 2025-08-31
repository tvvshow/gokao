<template>
  <div class="universities-page">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">院校查询</h1>
        <p class="page-subtitle">探索全国优质高等院校，找到最适合你的大学</p>
      </div>

      <!-- 搜索和筛选区域 -->
      <div class="search-section">
        <el-card class="content-card">
          <el-form :model="searchForm" :inline="true" class="search-form">
            <el-form-item label="院校名称">
              <el-input
                v-model="searchForm.name"
                placeholder="请输入院校名称"
                clearable
                style="width: 200px"
                @keyup.enter="handleSearch"
              >
                <template #prefix>
                  <el-icon><search /></el-icon>
                </template>
              </el-input>
            </el-form-item>

            <el-form-item label="所在省份">
              <el-select
                v-model="searchForm.province"
                placeholder="选择省份"
                clearable
                style="width: 150px"
              >
                <el-option
                  v-for="province in provinces"
                  :key="province"
                  :label="province"
                  :value="province"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="院校类型">
              <el-select
                v-model="searchForm.type"
                placeholder="选择类型"
                clearable
                style="width: 150px"
              >
                <el-option label="综合类" value="综合类" />
                <el-option label="理工类" value="理工类" />
                <el-option label="师范类" value="师范类" />
                <el-option label="财经类" value="财经类" />
                <el-option label="医药类" value="医药类" />
                <el-option label="艺术类" value="艺术类" />
              </el-select>
            </el-form-item>

            <el-form-item label="办学层次">
              <el-select
                v-model="searchForm.level"
                placeholder="选择层次"
                clearable
                style="width: 150px"
              >
                <el-option label="985工程" value="985" />
                <el-option label="211工程" value="211" />
                <el-option label="双一流" value="double_first_class" />
                <el-option label="普通本科" value="regular" />
              </el-select>
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="handleSearch" :loading="loading">
                <el-icon><search /></el-icon>
                搜索
              </el-button>
              <el-button @click="handleReset">重置</el-button>
            </el-form-item>
          </el-form>

          <!-- 高级筛选 -->
          <div class="advanced-filter" v-show="showAdvanced">
            <el-divider content-position="left">高级筛选</el-divider>
            <el-row :gutter="20">
              <el-col :span="8">
                <el-form-item label="录取分数线">
                  <el-slider
                    v-model="searchForm.scoreRange"
                    range
                    :min="300"
                    :max="750"
                    :step="10"
                    show-stops
                    :format-tooltip="formatScore"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="院校排名">
                  <el-slider
                    v-model="searchForm.rankRange"
                    range
                    :min="1"
                    :max="1000"
                    :step="10"
                    show-stops
                    :format-tooltip="formatRank"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="学校规模">
                  <el-radio-group v-model="searchForm.scale">
                    <el-radio label="">不限</el-radio>
                    <el-radio label="large">大型</el-radio>
                    <el-radio label="medium">中型</el-radio>
                    <el-radio label="small">小型</el-radio>
                  </el-radio-group>
                </el-form-item>
              </el-col>
            </el-row>
          </div>

          <div class="filter-actions">
            <el-button type="text" @click="showAdvanced = !showAdvanced">
              {{ showAdvanced ? '收起' : '展开' }}高级筛选
              <el-icon>
                <arrow-down v-if="!showAdvanced" />
                <arrow-up v-else />
              </el-icon>
            </el-button>
          </div>
        </el-card>
      </div>

      <!-- 结果统计 -->
      <div class="result-stats" v-if="universities.length > 0">
        <el-alert
          :title="`共找到 ${total} 所院校，当前显示第 ${(currentPage - 1) * pageSize + 1}-${Math.min(currentPage * pageSize, total)} 所`"
          type="info"
          :closable="false"
          show-icon
        />
      </div>

      <!-- 院校列表 -->
      <div class="universities-grid">
        <el-row :gutter="20" v-loading="loading">
          <el-col :xs="24" :sm="12" :md="8" :lg="6" v-for="university in universities" :key="university.id">
            <UniversityCard
              :university="university"
              @view="viewUniversity"
              @compare="addToCompare"
              @favorite="toggleFavorite"
            />
          </el-col>
        </el-row>

        <!-- 空状态 -->
        <el-empty
          v-if="!loading && universities.length === 0"
          description="未找到匹配的院校"
          :image-size="200"
        >
          <el-button type="primary" @click="handleReset">重新搜索</el-button>
        </el-empty>
      </div>

      <!-- 分页 -->
      <div class="pagination-wrapper" v-if="total > 0">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[12, 24, 36, 48]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </div>

    <!-- 对比浮窗 -->
    <div class="compare-panel" v-if="compareList.length > 0">
      <div class="compare-content">
        <div class="compare-header">
          <span>已选择 {{ compareList.length }} 所院校对比</span>
          <el-button type="primary" size="small" @click="showCompareDialog = true">
            对比分析
          </el-button>
          <el-button size="small" @click="clearCompare">清空</el-button>
        </div>
        <div class="compare-list">
          <el-tag
            v-for="university in compareList"
            :key="university.id"
            closable
            @close="removeFromCompare(university.id)"
          >
            {{ university.name }}
          </el-tag>
        </div>
      </div>
    </div>

    <!-- 对比弹窗 -->
    <CompareDialog
      v-model="showCompareDialog"
      :universities="compareList"
      @clear="clearCompare"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Search,
  ArrowDown,
  ArrowUp
} from '@element-plus/icons-vue'
import UniversityCard from '@/components/UniversityCard.vue'
import CompareDialog from '@/components/CompareDialog.vue'
import { universityApi } from '@/api/university'
import type { University } from '@/types/university'

const router = useRouter()

// 搜索表单
const searchForm = reactive({
  name: '',
  province: '',
  type: '',
  level: '',
  scoreRange: [400, 700],
  rankRange: [1, 500],
  scale: ''
})

// 状态数据
const loading = ref(false)
const showAdvanced = ref(false)
const universities = ref<University[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(12)
const compareList = ref<University[]>([])
const showCompareDialog = ref(false)

// 省份列表
const provinces = ref([
  '北京', '上海', '天津', '重庆', '河北', '山西', '辽宁', '吉林', '黑龙江',
  '江苏', '浙江', '安徽', '福建', '江西', '山东', '河南', '湖北', '湖南',
  '广东', '海南', '四川', '贵州', '云南', '陕西', '甘肃', '青海', '台湾',
  '内蒙古', '广西', '西藏', '宁夏', '新疆', '香港', '澳门'
])

// 搜索院校
const handleSearch = async () => {
  loading.value = true
  currentPage.value = 1

  try {
    const response = await universityApi.search({
      ...searchForm,
      page: currentPage.value,
      pageSize: pageSize.value
    })

    if (response.success) {
      universities.value = response.data.universities
      total.value = response.data.total
    } else {
      ElMessage.error(response.message || '搜索失败')
    }
  } catch (error) {
    ElMessage.error('搜索失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

// 重置搜索
const handleReset = () => {
  Object.assign(searchForm, {
    name: '',
    province: '',
    type: '',
    level: '',
    scoreRange: [400, 700],
    rankRange: [1, 500],
    scale: ''
  })
  handleSearch()
}

// 分页处理
const handleSizeChange = (size: number) => {
  pageSize.value = size
  handleSearch()
}

const handleCurrentChange = (page: number) => {
  currentPage.value = page
  handleSearch()
}

// 格式化函数
const formatScore = (value: number) => `${value}分`
const formatRank = (value: number) => `第${value}名`

// 查看院校详情
const viewUniversity = (id: string) => {
  router.push(`/universities/${id}`)
}

// 对比功能
const addToCompare = (university: University) => {
  if (compareList.value.length >= 4) {
    ElMessage.warning('最多只能对比4所院校')
    return
  }

  if (compareList.value.find(u => u.id === university.id)) {
    ElMessage.warning('该院校已在对比列表中')
    return
  }

  compareList.value.push(university)
  ElMessage.success(`已添加 ${university.name} 到对比列表`)
}

const removeFromCompare = (id: string) => {
  const index = compareList.value.findIndex(u => u.id === id)
  if (index > -1) {
    compareList.value.splice(index, 1)
  }
}

const clearCompare = () => {
  compareList.value = []
}

// 收藏功能
const toggleFavorite = async (university: University) => {
  try {
    const response = await universityApi.toggleFavorite(university.id)
    if (response.success) {
      university.isFavorite = !university.isFavorite
      ElMessage.success(university.isFavorite ? '已收藏' : '已取消收藏')
    }
  } catch (error) {
    ElMessage.error('操作失败，请稍后重试')
  }
}

// 页面加载时搜索
onMounted(() => {
  handleSearch()
})

// 监听搜索表单变化（防抖）
let searchTimer: NodeJS.Timeout
watch(
  () => searchForm.name,
  () => {
    clearTimeout(searchTimer)
    searchTimer = setTimeout(() => {
      if (searchForm.name) {
        handleSearch()
      }
    }, 500)
  }
)
</script>

<style scoped>
.universities-page {
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

.search-form {
  margin-bottom: 0;
}

.search-form .el-form-item {
  margin-bottom: 16px;
}

.advanced-filter {
  margin-top: 20px;
}

.filter-actions {
  text-align: center;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
}

.result-stats {
  margin-bottom: 20px;
}

.universities-grid {
  min-height: 400px;
}

.universities-grid .el-col {
  margin-bottom: 20px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 40px;
}

/* 对比面板 */
.compare-panel {
  position: fixed;
  bottom: 20px;
  right: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  border: 1px solid #ebeef5;
  z-index: 1000;
  max-width: 400px;
}

.compare-content {
  padding: 16px;
}

.compare-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  font-weight: 500;
}

.compare-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .search-form {
    display: block;
  }

  .search-form .el-form-item {
    display: block;
    margin-bottom: 16px;
  }

  .search-form .el-input,
  .search-form .el-select {
    width: 100% !important;
  }

  .compare-panel {
    left: 20px;
    right: 20px;
    max-width: none;
  }

  .compare-header {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
}
</style>