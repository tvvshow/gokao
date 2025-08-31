<template>
  <div class="recommendation-page">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">AI智能推荐</h1>
        <p class="page-subtitle">基于大数据分析和AI算法，为您量身定制最优志愿方案</p>
      </div>

      <div class="recommendation-content">
        <!-- 左侧：信息录入 -->
        <div class="input-section">
          <el-card class="content-card">
            <template #header>
              <div class="card-header">
                <el-icon><edit /></el-icon>
                <span>填写考生信息</span>
              </div>
            </template>

            <el-form
              ref="formRef"
              :model="studentForm"
              :rules="formRules"
              label-width="100px"
              @submit.prevent="handleRecommend"
            >
              <!-- 基础信息 -->
              <div class="form-section">
                <h3 class="section-title">基础信息</h3>
                <el-row :gutter="20">
                  <el-col :span="12">
                    <el-form-item label="高考分数" prop="score" required>
                      <el-input
                        v-model.number="studentForm.score"
                        type="number"
                        placeholder="请输入高考分数"
                        :min="0"
                        :max="750"
                      >
                        <template #append>分</template>
                      </el-input>
                    </el-form-item>
                  </el-col>
                  <el-col :span="12">
                    <el-form-item label="所在省份" prop="province" required>
                      <el-select v-model="studentForm.province" placeholder="选择省份" style="width: 100%">
                        <el-option
                          v-for="province in provinces"
                          :key="province"
                          :label="province"
                          :value="province"
                        />
                      </el-select>
                    </el-form-item>
                  </el-col>
                </el-row>

                <el-row :gutter="20">
                  <el-col :span="12">
                    <el-form-item label="文理科类" prop="scienceType" required>
                      <el-radio-group v-model="studentForm.scienceType">
                        <el-radio label="理科">理科</el-radio>
                        <el-radio label="文科">文科</el-radio>
                        <el-radio label="新高考">新高考</el-radio>
                      </el-radio-group>
                    </el-form-item>
                  </el-col>
                  <el-col :span="12">
                    <el-form-item label="高考年份" prop="year">
                      <el-select v-model="studentForm.year" placeholder="选择年份" style="width: 100%">
                        <el-option
                          v-for="year in years"
                          :key="year"
                          :label="year"
                          :value="year"
                        />
                      </el-select>
                    </el-form-item>
                  </el-col>
                </el-row>

                <el-form-item label="位次排名" prop="rank">
                  <el-input
                    v-model.number="studentForm.rank"
                    type="number"
                    placeholder="请输入位次排名（可选）"
                  >
                    <template #append>名</template>
                  </el-input>
                </el-form-item>
              </div>

              <!-- 偏好设置 -->
              <div class="form-section">
                <h3 class="section-title">志愿偏好</h3>
                
                <el-form-item label="意向地区">
                  <el-select
                    v-model="studentForm.preferences.regions"
                    multiple
                    placeholder="选择意向地区"
                    style="width: 100%"
                  >
                    <el-option
                      v-for="region in regions"
                      :key="region"
                      :label="region"
                      :value="region"
                    />
                  </el-select>
                </el-form-item>

                <el-form-item label="专业类别">
                  <el-select
                    v-model="studentForm.preferences.majorCategories"
                    multiple
                    placeholder="选择感兴趣的专业类别"
                    style="width: 100%"
                  >
                    <el-option
                      v-for="category in majorCategories"
                      :key="category"
                      :label="category"
                      :value="category"
                    />
                  </el-select>
                </el-form-item>

                <el-form-item label="院校类型">
                  <el-checkbox-group v-model="studentForm.preferences.universityTypes">
                    <el-checkbox label="985工程">985工程</el-checkbox>
                    <el-checkbox label="211工程">211工程</el-checkbox>
                    <el-checkbox label="双一流">双一流</el-checkbox>
                    <el-checkbox label="普通本科">普通本科</el-checkbox>
                  </el-checkbox-group>
                </el-form-item>

                <el-form-item label="风险承受度">
                  <el-radio-group v-model="studentForm.preferences.riskTolerance">
                    <el-radio label="conservative">保守型（冲1保5稳4）</el-radio>
                    <el-radio label="moderate">稳健型（冲2保3稳5）</el-radio>
                    <el-radio label="aggressive">激进型（冲4保2稳4）</el-radio>
                  </el-radio-group>
                </el-form-item>

                <el-form-item label="特殊要求">
                  <el-input
                    v-model="studentForm.preferences.specialRequirements"
                    type="textarea"
                    :rows="3"
                    placeholder="如：不接受中外合作办学、希望在大城市、对某专业有特别偏好等"
                  />
                </el-form-item>
              </div>

              <div class="form-actions">
                <el-button
                  type="primary"
                  size="large"
                  :loading="generating"
                  @click="handleRecommend"
                >
                  <el-icon><magic-stick /></el-icon>
                  生成智能推荐
                </el-button>
                <el-button size="large" @click="handleReset">重置</el-button>
              </div>
            </el-form>
          </el-card>
        </div>

        <!-- 右侧：推荐结果 -->
        <div class="result-section">
          <el-card class="content-card" v-if="recommendations.length > 0">
            <template #header>
              <div class="card-header">
                <el-icon><trophy /></el-icon>
                <span>推荐结果</span>
                <div class="result-actions">
                  <el-button size="small" @click="exportRecommendations">
                    <el-icon><download /></el-icon>
                    导出报告
                  </el-button>
                  <el-button size="small" type="primary" @click="saveRecommendations">
                    <el-icon><collection /></el-icon>
                    保存方案
                  </el-button>
                </div>
              </div>
            </template>

            <!-- 推荐统计 -->
            <div class="recommendation-stats">
              <el-row :gutter="20">
                <el-col :span="6">
                  <div class="stat-item">
                    <div class="stat-value">{{ recommendations.length }}</div>
                    <div class="stat-label">推荐院校</div>
                  </div>
                </el-col>
                <el-col :span="6">
                  <div class="stat-item">
                    <div class="stat-value">{{ getSuccessRate() }}%</div>
                    <div class="stat-label">预计成功率</div>
                  </div>
                </el-col>
                <el-col :span="6">
                  <div class="stat-item">
                    <div class="stat-value">{{ getRiskLevel() }}</div>
                    <div class="stat-label">风险等级</div>
                  </div>
                </el-col>
                <el-col :span="6">
                  <div class="stat-item">
                    <div class="stat-value">{{ getMatchScore() }}</div>
                    <div class="stat-label">匹配度</div>
                  </div>
                </el-col>
              </el-row>
            </div>

            <!-- 分类标签 -->
            <div class="category-tabs">
              <el-tabs v-model="activeCategory" @tab-click="handleCategoryChange">
                <el-tab-pane label="冲一冲" name="aggressive">
                  <el-badge :value="getRecommendationsByType('aggressive').length" class="tab-badge" />
                </el-tab-pane>
                <el-tab-pane label="稳一稳" name="moderate">
                  <el-badge :value="getRecommendationsByType('moderate').length" class="tab-badge" />
                </el-tab-pane>
                <el-tab-pane label="保一保" name="conservative">
                  <el-badge :value="getRecommendationsByType('conservative').length" class="tab-badge" />
                </el-tab-pane>
              </el-tabs>
            </div>

            <!-- 推荐列表 -->
            <div class="recommendations-list">
              <RecommendationCard
                v-for="(recommendation, index) in getCurrentRecommendations()"
                :key="recommendation.id"
                :recommendation="recommendation"
                :index="index + 1"
                @view="viewUniversityDetail"
                @compare="addToCompare"
                @favorite="toggleFavorite"
              />
            </div>
          </el-card>

          <!-- 空状态 -->
          <el-card class="content-card" v-else-if="!generating">
            <div class="empty-state">
              <el-icon size="80"><magic-stick /></el-icon>
              <h3>开始您的志愿推荐</h3>
              <p>填写左侧信息，获取AI智能推荐的志愿方案</p>
            </div>
          </el-card>

          <!-- 加载状态 -->
          <el-card class="content-card" v-else>
            <div class="loading-state">
              <el-icon class="rotating" size="60"><loading /></el-icon>
              <h3>AI正在分析中...</h3>
              <p>正在基于您的信息进行智能匹配，请稍候</p>
              <el-progress :percentage="progressPercentage" :stroke-width="8" />
            </div>
          </el-card>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  Edit,
  MagicStick,
  Trophy,
  Download,
  Collection,
  Loading
} from '@element-plus/icons-vue'
import RecommendationCard from '@/components/RecommendationCard.vue'
import { recommendationApi } from '@/api/recommendation'
import type { StudentInfo, Recommendation } from '@/types/recommendation'

const router = useRouter()
const formRef = ref<FormInstance>()

// 表单数据
const studentForm = reactive<StudentInfo>({
  score: null,
  province: '',
  scienceType: '理科',
  year: new Date().getFullYear(),
  rank: null,
  preferences: {
    regions: [],
    majorCategories: [],
    universityTypes: [],
    riskTolerance: 'moderate',
    specialRequirements: ''
  }
})

// 表单验证规则
const formRules: FormRules = {
  score: [
    { required: true, message: '请输入高考分数', trigger: 'blur' },
    { type: 'number', min: 0, max: 750, message: '分数应在0-750之间', trigger: 'blur' }
  ],
  province: [
    { required: true, message: '请选择所在省份', trigger: 'change' }
  ],
  scienceType: [
    { required: true, message: '请选择文理科类', trigger: 'change' }
  ]
}

// 状态数据
const generating = ref(false)
const progressPercentage = ref(0)
const recommendations = ref<Recommendation[]>([])
const activeCategory = ref('moderate')

// 基础数据
const provinces = ref([
  '北京', '上海', '天津', '重庆', '河北', '山西', '辽宁', '吉林', '黑龙江',
  '江苏', '浙江', '安徽', '福建', '江西', '山东', '河南', '湖北', '湖南',
  '广东', '海南', '四川', '贵州', '云南', '陕西', '甘肃', '青海',
  '内蒙古', '广西', '西藏', '宁夏', '新疆'
])

const regions = ref([
  '华北地区', '东北地区', '华东地区', '华中地区', '华南地区', '西南地区', '西北地区'
])

const majorCategories = ref([
  '哲学', '经济学', '法学', '教育学', '文学', '历史学', '理学', '工学', 
  '农学', '医学', '管理学', '艺术学'
])

const years = computed(() => {
  const currentYear = new Date().getFullYear()
  return Array.from({ length: 5 }, (_, i) => currentYear - i)
})

// 生成推荐
const handleRecommend = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    
    generating.value = true
    progressPercentage.value = 0

    // 模拟进度
    const progressInterval = setInterval(() => {
      progressPercentage.value += Math.random() * 15
      if (progressPercentage.value >= 90) {
        clearInterval(progressInterval)
      }
    }, 200)

    try {
      const response = await recommendationApi.generateRecommendations(studentForm)
      
      if (response.success) {
        recommendations.value = response.data.recommendations
        progressPercentage.value = 100
        ElMessage.success('推荐生成成功')
      } else {
        ElMessage.error(response.message || '推荐生成失败')
      }
    } catch (error) {
      ElMessage.error('推荐生成失败，请稍后重试')
    } finally {
      clearInterval(progressInterval)
      setTimeout(() => {
        generating.value = false
        progressPercentage.value = 0
      }, 500)
    }
  } catch {
    ElMessage.warning('请完善必填信息')
  }
}

// 重置表单
const handleReset = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  recommendations.value = []
}

// 获取当前分类的推荐
const getCurrentRecommendations = () => {
  return getRecommendationsByType(activeCategory.value)
}

const getRecommendationsByType = (type: string) => {
  return recommendations.value.filter(rec => rec.type === type)
}

// 统计函数
const getSuccessRate = () => {
  if (recommendations.value.length === 0) return 0
  const avgProbability = recommendations.value.reduce((sum, rec) => sum + rec.admissionProbability, 0) / recommendations.value.length
  return Math.round(avgProbability)
}

const getRiskLevel = () => {
  const riskMap = {
    conservative: '低风险',
    moderate: '中风险',
    aggressive: '高风险'
  }
  return riskMap[studentForm.preferences.riskTolerance] || '中风险'
}

const getMatchScore = () => {
  if (recommendations.value.length === 0) return 0
  const avgMatch = recommendations.value.reduce((sum, rec) => sum + rec.matchScore, 0) / recommendations.value.length
  return Math.round(avgMatch)
}

// 处理分类切换
const handleCategoryChange = (tab: any) => {
  activeCategory.value = tab.paneName
}

// 查看院校详情
const viewUniversityDetail = (universityId: string) => {
  router.push(`/universities/${universityId}`)
}

// 添加到对比
const addToCompare = (recommendation: Recommendation) => {
  // 实现对比功能
  ElMessage.success(`已添加 ${recommendation.university.name} 到对比列表`)
}

// 收藏切换
const toggleFavorite = (recommendation: Recommendation) => {
  recommendation.university.isFavorite = !recommendation.university.isFavorite
  ElMessage.success(recommendation.university.isFavorite ? '已收藏' : '已取消收藏')
}

// 导出推荐报告
const exportRecommendations = async () => {
  try {
    const response = await recommendationApi.exportReport(recommendations.value)
    if (response.success) {
      // 下载文件
      const blob = new Blob([response.data], { type: 'application/pdf' })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `志愿推荐报告_${studentForm.score}分_${new Date().toLocaleDateString()}.pdf`
      link.click()
      window.URL.revokeObjectURL(url)
      ElMessage.success('报告导出成功')
    }
  } catch (error) {
    ElMessage.error('导出失败，请稍后重试')
  }
}

// 保存推荐方案
const saveRecommendations = async () => {
  try {
    const { value: schemeName } = await ElMessageBox.prompt('请输入方案名称', '保存方案', {
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValue: `${studentForm.score}分志愿方案`,
      inputValidator: (value) => {
        if (!value?.trim()) {
          return '请输入方案名称'
        }
        return true
      }
    })

    const response = await recommendationApi.saveScheme({
      name: schemeName,
      studentInfo: studentForm,
      recommendations: recommendations.value
    })

    if (response.success) {
      ElMessage.success('方案保存成功')
    }
  } catch {
    // 用户取消
  }
}

onMounted(() => {
  // 从路由参数中恢复数据
  const query = router.currentRoute.value.query
  if (query.score) {
    studentForm.score = Number(query.score)
  }
  if (query.province) {
    studentForm.province = query.province as string
  }
})
</script>

<style scoped>
.recommendation-page {
  padding: 20px 0;
  min-height: calc(100vh - 160px);
}

.container {
  max-width: 1400px;
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

.recommendation-content {
  display: grid;
  grid-template-columns: 400px 1fr;
  gap: 30px;
  align-items: start;
}

.input-section {
  position: sticky;
  top: 100px;
}

.result-section {
  min-height: 600px;
}

.card-header {
  display: flex;
  align-items: center;
  font-weight: 600;
  color: #2c3e50;
}

.card-header .el-icon {
  margin-right: 8px;
  color: #667eea;
}

.result-actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

.form-section {
  margin-bottom: 30px;
}

.section-title {
  font-size: 16px;
  color: #2c3e50;
  margin-bottom: 20px;
  padding-bottom: 8px;
  border-bottom: 2px solid #667eea;
}

.form-actions {
  text-align: center;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}

.form-actions .el-button {
  margin: 0 8px;
  padding: 12px 24px;
}

.recommendation-stats {
  margin-bottom: 24px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #667eea;
  margin-bottom: 4px;
}

.stat-label {
  color: #7f8c8d;
  font-size: 14px;
}

.category-tabs {
  margin-bottom: 24px;
}

.tab-badge {
  margin-left: 8px;
}

.recommendations-list {
  max-height: 600px;
  overflow-y: auto;
}

.empty-state,
.loading-state {
  text-align: center;
  padding: 60px 20px;
  color: #7f8c8d;
}

.empty-state .el-icon,
.loading-state .el-icon {
  color: #667eea;
  margin-bottom: 20px;
}

.empty-state h3,
.loading-state h3 {
  margin-bottom: 12px;
  color: #2c3e50;
}

.loading-state .el-progress {
  margin-top: 20px;
  max-width: 300px;
  margin-left: auto;
  margin-right: auto;
}

.rotating {
  animation: rotate 2s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .recommendation-content {
    grid-template-columns: 1fr;
    gap: 20px;
  }
  
  .input-section {
    position: static;
  }
}

@media (max-width: 768px) {
  .container {
    padding: 0 10px;
  }
  
  .page-title {
    font-size: 24px;
  }
  
  .form-actions .el-button {
    display: block;
    width: 100%;
    margin: 8px 0;
  }
}
</style>