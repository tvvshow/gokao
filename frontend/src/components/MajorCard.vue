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
  border-radius: 14px;
  border: 1px solid rgb(148 163 184 / 0.22);
  box-shadow: 0 16px 30px -28px rgb(15 23 42 / 0.55);
}

.major-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 22px 40px -32px rgb(14 165 233 / 0.72);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0;
}

.major-name {
  font-size: 17px;
  font-weight: 600;
  color: #0f172a;
  margin: 0;
  letter-spacing: -0.01em;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.major-info {
  padding: 0;
}

.basic-info {
  margin-bottom: 14px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 9px;
  font-size: 13px;
}

.info-item .label {
  color: #64748b;
}

.info-item .value {
  color: #0f172a;
  font-weight: 500;
}

.stats-info {
  display: flex;
  justify-content: space-around;
  background: #f8fafc;
  border-radius: 10px;
  padding: 13px;
  margin-bottom: 14px;
  border: 1px solid rgb(148 163 184 / 0.2);
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 21px;
  font-weight: 700;
  margin-bottom: 4px;
}

.stat-value.employment {
  color: #22c55e;
}

.stat-value.salary {
  color: #f59e0b;
}

.stat-label {
  font-size: 12px;
  color: #64748b;
}

.description {
  font-size: 13px;
  color: #64748b;
  line-height: 1.6;
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
  gap: 8px;
}

.card-actions .el-button {
  flex: 1;
  margin: 0;
  min-height: 38px;
  border-radius: 10px;
}

@media (max-width: 768px) {
  .major-name {
    font-size: 15px;
  }
}
</style>
