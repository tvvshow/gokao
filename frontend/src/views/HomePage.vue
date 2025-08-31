<template>
  <div class="home-page">
    <!-- 英雄区域 -->
    <section class="hero-section">
      <div class="hero-content">
        <h1 class="hero-title">AI智能志愿填报助手</h1>
        <p class="hero-subtitle">基于大数据分析和AI算法，为每位考生量身定制最优志愿方案</p>
        <div class="hero-actions">
          <el-button type="primary" size="large" @click="$router.push('/recommendation')">
            <el-icon><magic-stick /></el-icon>
            开始智能推荐
          </el-button>
          <el-button size="large" @click="$router.push('/universities')">
            <el-icon><search /></el-icon>
            院校查询
          </el-button>
        </div>
        
        <!-- 快速查询 -->
        <div class="quick-search">
          <h3>快速查询</h3>
          <el-row :gutter="20">
            <el-col :span="8">
              <el-input
                v-model="quickSearchForm.score"
                placeholder="请输入分数"
                type="number"
                size="large"
              >
                <template #prepend>分数</template>
              </el-input>
            </el-col>
            <el-col :span="8">
              <el-select
                v-model="quickSearchForm.province"
                placeholder="选择省份"
                size="large"
                style="width: 100%"
              >
                <el-option label="北京" value="北京" />
                <el-option label="上海" value="上海" />
                <el-option label="广东" value="广东" />
                <el-option label="江苏" value="江苏" />
                <el-option label="浙江" value="浙江" />
              </el-select>
            </el-col>
            <el-col :span="8">
              <el-button
                type="primary"
                size="large"
                :loading="searching"
                @click="handleQuickSearch"
                style="width: 100%"
              >
                快速匹配
              </el-button>
            </el-col>
          </el-row>
        </div>
      </div>
    </section>

    <!-- 统计数据 -->
    <section class="stats-section">
      <div class="container">
        <el-row :gutter="30">
          <el-col :xs="12" :sm="6" v-for="stat in stats" :key="stat.label">
            <div class="stat-card">
              <div class="stat-icon">
                <el-icon><component :is="stat.icon" /></el-icon>
              </div>
              <div class="stat-number">{{ stat.value }}</div>
              <div class="stat-label">{{ stat.label }}</div>
            </div>
          </el-col>
        </el-row>
      </div>
    </section>

    <!-- 功能特色 -->
    <section class="features-section">
      <div class="container">
        <h2 class="section-title">核心功能</h2>
        <el-row :gutter="30">
          <el-col :xs="24" :sm="12" :md="8" v-for="feature in features" :key="feature.title">
            <div class="feature-card" @click="$router.push(feature.link)">
              <div class="feature-icon">
                <el-icon><component :is="feature.icon" /></el-icon>
              </div>
              <h3>{{ feature.title }}</h3>
              <p>{{ feature.description }}</p>
              <el-button type="primary" text>了解更多 →</el-button>
            </div>
          </el-col>
        </el-row>
      </div>
    </section>

    <!-- 热门院校 -->
    <section class="popular-section">
      <div class="container">
        <h2 class="section-title">热门院校</h2>
        <el-row :gutter="20">
          <el-col :xs="12" :sm="8" :md="6" v-for="university in popularUniversities" :key="university.id">
            <el-card class="university-card" shadow="hover" @click="viewUniversity(university.id)">
              <template #header>
                <div class="university-header">
                  <img :src="university.logo" :alt="university.name" class="university-logo" />
                  <div class="university-rank">#{{ university.rank }}</div>
                </div>
              </template>
              <div class="university-name">{{ university.name }}</div>
              <div class="university-location">{{ university.location }}</div>
              <div class="university-score">录取分数: {{ university.minScore }}+</div>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  MagicStick,
  Search,
  School,
  PieChart,
  Trophy,
  UserFilled,
  DataAnalysis,
  Star,
  TrendCharts,
  Location
} from '@element-plus/icons-vue'

const router = useRouter()

// 快速搜索表单
const quickSearchForm = reactive({
  score: '',
  province: ''
})

const searching = ref(false)

// 统计数据
const stats = ref([
  { label: '合作院校', value: '2,856', icon: School },
  { label: '专业数据', value: '15,248', icon: PieChart },
  { label: '成功案例', value: '186,429', icon: Trophy },
  { label: '注册用户', value: '458,932', icon: UserFilled }
])

// 功能特色
const features = ref([
  {
    title: '智能推荐',
    description: 'AI算法分析历史数据，为您推荐最适合的院校和专业组合',
    icon: MagicStick,
    link: '/recommendation'
  },
  {
    title: '院校查询',
    description: '全面的院校信息库，支持多维度筛选和对比分析',
    icon: School,
    link: '/universities'
  },
  {
    title: '专业分析',
    description: '详细的专业信息、就业前景和薪资水平分析',
    icon: PieChart,
    link: '/majors'
  },
  {
    title: '数据分析',
    description: '历史录取数据分析，录取概率预测和风险评估',
    icon: DataAnalysis,
    link: '/analysis'
  },
  {
    title: '趋势预测',
    description: '基于历史数据分析未来录取趋势和分数变化',
    icon: TrendCharts,
    link: '/analysis'
  },
  {
    title: '个性定制',
    description: '根据个人兴趣和职业规划定制专属志愿方案',
    icon: Star,
    link: '/recommendation'
  }
])

// 热门院校
const popularUniversities = ref([
  {
    id: '1',
    name: '清华大学',
    location: '北京',
    rank: 1,
    minScore: 690,
    logo: '/logos/tsinghua.png'
  },
  {
    id: '2',
    name: '北京大学',
    location: '北京',
    rank: 2,
    minScore: 685,
    logo: '/logos/pku.png'
  },
  {
    id: '3',
    name: '复旦大学',
    location: '上海',
    rank: 3,
    minScore: 680,
    logo: '/logos/fudan.png'
  },
  {
    id: '4',
    name: '上海交通大学',
    location: '上海',
    rank: 4,
    minScore: 678,
    logo: '/logos/sjtu.png'
  }
])

// 快速搜索
const handleQuickSearch = async () => {
  if (!quickSearchForm.score || !quickSearchForm.province) {
    ElMessage.warning('请输入分数和选择省份')
    return
  }

  searching.value = true
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    router.push({
      path: '/recommendation',
      query: {
        score: quickSearchForm.score,
        province: quickSearchForm.province
      }
    })
  } catch (error) {
    ElMessage.error('查询失败，请稍后重试')
  } finally {
    searching.value = false
  }
}

// 查看院校详情
const viewUniversity = (id: string) => {
  router.push(`/universities/${id}`)
}

onMounted(() => {
  // 可以在这里加载实际的统计数据
})
</script>

<style scoped>
.home-page {
  min-height: 100vh;
}

/* 英雄区域 */
.hero-section {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 120px 0 80px;
  text-align: center;
}

.hero-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.hero-title {
  font-size: 48px;
  font-weight: 700;
  margin-bottom: 20px;
  background: linear-gradient(45deg, #fff, #ffd700);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-subtitle {
  font-size: 20px;
  margin-bottom: 40px;
  opacity: 0.9;
  line-height: 1.6;
}

.hero-actions {
  margin-bottom: 60px;
}

.hero-actions .el-button {
  margin: 0 10px;
  padding: 16px 32px;
  font-size: 16px;
  border-radius: 25px;
}

.quick-search {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 30px;
  text-align: left;
  max-width: 800px;
  margin: 0 auto;
}

.quick-search h3 {
  text-align: center;
  margin-bottom: 24px;
  font-size: 24px;
}

/* 统计区域 */
.stats-section {
  padding: 80px 0;
  background: #f8f9fa;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.stat-card {
  text-align: center;
  padding: 30px 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-5px);
}

.stat-icon {
  font-size: 36px;
  color: #667eea;
  margin-bottom: 16px;
}

.stat-number {
  font-size: 32px;
  font-weight: 700;
  color: #2c3e50;
  margin-bottom: 8px;
}

.stat-label {
  color: #7f8c8d;
  font-size: 16px;
}

/* 功能特色 */
.features-section {
  padding: 80px 0;
  background: white;
}

.section-title {
  text-align: center;
  font-size: 36px;
  color: #2c3e50;
  margin-bottom: 60px;
  position: relative;
}

.section-title::after {
  content: '';
  position: absolute;
  bottom: -10px;
  left: 50%;
  transform: translateX(-50%);
  width: 60px;
  height: 4px;
  background: linear-gradient(45deg, #667eea, #764ba2);
  border-radius: 2px;
}

.feature-card {
  text-align: center;
  padding: 40px 20px;
  border-radius: 12px;
  background: #f8f9fa;
  transition: all 0.3s ease;
  cursor: pointer;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.feature-card:hover {
  background: white;
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
  transform: translateY(-5px);
}

.feature-icon {
  font-size: 48px;
  color: #667eea;
  margin-bottom: 20px;
}

.feature-card h3 {
  font-size: 20px;
  color: #2c3e50;
  margin-bottom: 16px;
}

.feature-card p {
  color: #7f8c8d;
  line-height: 1.6;
  margin-bottom: 20px;
  flex: 1;
}

/* 热门院校 */
.popular-section {
  padding: 80px 0;
  background: #f8f9fa;
}

.university-card {
  cursor: pointer;
  transition: all 0.3s ease;
  height: 100%;
}

.university-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
}

.university-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.university-logo {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  object-fit: cover;
}

.university-rank {
  background: #667eea;
  color: white;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.university-name {
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
  margin-bottom: 8px;
}

.university-location {
  color: #7f8c8d;
  font-size: 14px;
  margin-bottom: 8px;
}

.university-score {
  color: #e74c3c;
  font-size: 14px;
  font-weight: 500;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .hero-title {
    font-size: 32px;
  }
  
  .hero-subtitle {
    font-size: 16px;
  }
  
  .hero-actions .el-button {
    display: block;
    margin: 10px auto;
    width: 200px;
  }
  
  .quick-search {
    padding: 20px;
  }
  
  .section-title {
    font-size: 28px;
  }
  
  .feature-card {
    margin-bottom: 20px;
  }
}
</style>