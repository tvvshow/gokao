<template>
  <el-card class="major-card" shadow="hover" @click="$emit('view', major.id)">
    <template #header>
      <div class="card-header">
        <h3 class="major-name">{{ major.name }}</h3>
        <el-tag v-if="major.isPopular" type="danger" size="small">热门</el-tag>
      </div>
    </template>

    <div class="major-info">
      <div class="basic-info">
        <div class="info-item">
          <span class="label">学科门类:</span>
          <span class="value">{{ major.category }}</span>
        </div>
        <div class="info-item">
          <span class="label">学制:</span>
          <span class="value">{{ major.duration }}年</span>
        </div>
        <div class="info-item">
          <span class="label">学位:</span>
          <span class="value">{{ major.degree }}</span>
        </div>
      </div>

      <div class="stats-info">
        <div class="stat-item">
          <div class="stat-value employment">
            {{ major.employmentRate || '--' }}%
          </div>
          <div class="stat-label">就业率</div>
        </div>
        <div class="stat-item">
          <div class="stat-value salary">
            {{ major.averageSalary || '--' }}k
          </div>
          <div class="stat-label">平均薪资</div>
        </div>
      </div>

      <div class="description">
        <p>{{ major.description || '暂无专业描述' }}</p>
      </div>
    </div>

    <template #footer>
      <div class="card-actions">
        <el-button
          type="primary"
          size="small"
          @click.stop="$emit('view', major.id)"
        >
          查看详情
        </el-button>
        <el-button size="small" @click.stop="viewEmploymentData">
          就业分析
        </el-button>
      </div>
    </template>
  </el-card>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import type { Major } from '@/types/university';

interface Props {
  major: Major;
}

const props = defineProps<Props>();

const router = useRouter();

const viewEmploymentData = () => {
  router.push(`/analysis?major=${props.major.id}`);
};
</script>

<style scoped>
.major-card {
  height: 100%;
  cursor: pointer;
  transition: all 0.3s ease;
  border-radius: 8px;
}

.major-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0;
}

.major-name {
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
  margin: 0;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.major-info {
  padding: 0;
}

.basic-info {
  margin-bottom: 16px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
}

.info-item .label {
  color: #7f8c8d;
}

.info-item .value {
  color: #2c3e50;
  font-weight: 500;
}

.stats-info {
  display: flex;
  justify-content: space-around;
  background: #f8f9fa;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 4px;
}

.stat-value.employment {
  color: #67c23a;
}

.stat-value.salary {
  color: #e6a23c;
}

.stat-label {
  font-size: 12px;
  color: #7f8c8d;
}

.description {
  font-size: 14px;
  color: #7f8c8d;
  line-height: 1.5;
}

.description p {
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-actions {
  display: flex;
  justify-content: space-between;
  padding: 0;
}

.card-actions .el-button {
  flex: 1;
  margin: 0 4px;
}

.card-actions .el-button:first-child {
  margin-left: 0;
}

.card-actions .el-button:last-child {
  margin-right: 0;
}
</style>
