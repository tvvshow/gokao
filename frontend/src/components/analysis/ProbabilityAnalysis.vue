<template>
  <div class="probability-analysis">
    <el-card>
      <template #header>
        <h3>录取概率预测</h3>
      </template>
      <div class="analysis-content">
        <el-form :model="form" :inline="true">
          <el-form-item label="高考分数">
            <el-input v-model="form.score" placeholder="请输入分数" />
          </el-form-item>
          <el-form-item label="目标院校">
            <el-select v-model="form.university" placeholder="选择院校">
              <el-option label="清华大学" value="1" />
              <el-option label="北京大学" value="2" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="analyze">开始分析</el-button>
          </el-form-item>
        </el-form>
        
        <div v-if="result" class="result">
          <el-alert :title="`录取概率：${result.probability}%`" type="success" show-icon />
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'

const form = reactive({
  score: '',
  university: ''
})

const result = ref<any>(null)

const analyze = () => {
  // 模拟分析
  result.value = {
    probability: Math.floor(Math.random() * 100)
  }
}
</script>

<style scoped>
.analysis-content {
  padding: 20px 0;
}

.result {
  margin-top: 20px;
}
</style>