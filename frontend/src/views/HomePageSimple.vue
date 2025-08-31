<template>
  <div class="home-page">
    <!-- 英雄区域 -->
    <section class="hero-section">
      <div class="hero-content">
        <h1 class="hero-title">🎓 AI智能志愿填报助手</h1>
        <p class="hero-subtitle">基于大数据分析和AI算法，为每位考生量身定制最优志愿方案</p>
        <div class="hero-actions">
          <el-button type="primary" size="large" @click="$router.push('/recommendation')">
            开始智能推荐
          </el-button>
          <el-button size="large" @click="$router.push('/universities')">
            院校查询
          </el-button>
        </div>
      </div>
    </section>

    <!-- 快速查询 -->
    <section class="quick-search-section">
      <div class="container">
        <el-card class="quick-search-card">
          <h3>🔍 快速查询</h3>
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
        </el-card>
      </div>
    </section>

    <!-- 统计数据 -->
    <section class="stats-section">
      <div class="container">
        <el-row :gutter="30">
          <el-col :xs="12" :sm="6">
            <div class="stat-card">
              <div class="stat-number">2,856</div>
              <div class="stat-label">🏫 合作院校</div>
            </div>
          </el-col>
          <el-col :xs="12" :sm="6">
            <div class="stat-card">
              <div class="stat-number">15,248</div>
              <div class="stat-label">📚 专业数据</div>
            </div>
          </el-col>
          <el-col :xs="12" :sm="6">
            <div class="stat-card">
              <div class="stat-number">186,429</div>
              <div class="stat-label">🏆 成功案例</div>
            </div>
          </el-col>
          <el-col :xs="12" :sm="6">
            <div class="stat-card">
              <div class="stat-number">458,932</div>
              <div class="stat-label">👥 注册用户</div>
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
          <el-col :xs="24" :sm="12" :md="8">
            <div class="feature-card" @click="$router.push('/recommendation')">
              <div class="feature-icon">🤖</div>
              <h3>智能推荐</h3>
              <p>AI算法分析历史数据，为您推荐最适合的院校和专业组合</p>
              <el-button type="primary" text>了解更多 →</el-button>
            </div>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <div class="feature-card" @click="$router.push('/universities')">
              <div class="feature-icon">🏫</div>
              <h3>院校查询</h3>
              <p>全面的院校信息库，支持多维度筛选和对比分析</p>
              <el-button type="primary" text>了解更多 →</el-button>
            </div>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <div class="feature-card" @click="$router.push('/majors')">
              <div class="feature-icon">📊</div>
              <h3>专业分析</h3>
              <p>详细的专业信息、就业前景和薪资水平分析</p>
              <el-button type="primary" text>了解更多 →</el-button>
            </div>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <div class="feature-card" @click="$router.push('/analysis')">
              <div class="feature-icon">📈</div>
              <h3>数据分析</h3>
              <p>历史录取数据分析，录取概率预测和风险评估</p>
              <el-button type="primary" text>了解更多 →</el-button>
            </div>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <div class="feature-card" @click="$router.push('/analysis')">
              <div class="feature-icon">📉</div>
              <h3>趋势预测</h3>
              <p>基于历史数据分析未来录取趋势和分数变化</p>
              <el-button type="primary" text>了解更多 →</el-button>
            </div>
          </el-col>
          <el-col :xs="24" :sm="12" :md="8">
            <div class="feature-card" @click="$router.push('/recommendation')">
              <div class="feature-icon">⭐</div>
              <h3>个性定制</h3>
              <p>根据个人兴趣和职业规划定制专属志愿方案</p>
              <el-button type="primary" text>了解更多 →</el-button>
            </div>
          </el-col>
        </el-row>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const router = useRouter()

// 快速搜索表单
const quickSearchForm = reactive({
  score: '',
  province: ''
})

const searching = ref(false)

// 快速搜索
const handleQuickSearch = async () => {
  if (!quickSearchForm.score || !quickSearchForm.province) {
    ElMessage.warning('请输入分数和选择省份')
    return
  }

  searching.value = true
  try {
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

/* 快速查询 */
.quick-search-section {
  padding: 60px 0;
  background: #f8f9fa;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.quick-search-card {
  text-align: center;
  padding: 30px;
}

.quick-search-card h3 {
  font-size: 24px;
  margin-bottom: 24px;
  color: #2c3e50;
}

/* 统计区域 */
.stats-section {
  padding: 80px 0;
  background: white;
}

.stat-card {
  text-align: center;
  padding: 30px 20px;
  background: #f8f9fa;
  border-radius: 12px;
  transition: transform 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-5px);
}

.stat-number {
  font-size: 32px;
  font-weight: 700;
  color: #667eea;
  margin-bottom: 8px;
}

.stat-label {
  color: #7f8c8d;
  font-size: 16px;
}

/* 功能特色 */
.features-section {
  padding: 80px 0;
  background: #f8f9fa;
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
  background: white;
  transition: all 0.3s ease;
  cursor: pointer;
  height: 100%;
  display: flex;
  flex-direction: column;
  margin-bottom: 20px;
}

.feature-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}

.feature-icon {
  font-size: 48px;
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
  
  .quick-search-card {
    padding: 20px;
  }
  
  .section-title {
    font-size: 28px;
  }
}
</style>