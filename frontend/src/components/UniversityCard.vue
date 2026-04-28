<template>
  <el-card
    class="university-card"
    shadow="hover"
    @click="$emit('view', university.id)"
  >
    <template #header>
      <div class="card-header">
        <div class="university-basic">
          <img
            :src="university.logo || '/default-logo.svg'"
            :alt="university.name"
            class="university-logo"
            @error="handleImageError"
          />
          <div class="university-info">
            <h3 class="university-name">{{ university.name }}</h3>
            <div class="university-meta">
              <el-tag size="small" :type="getTagType(university.level)">
                {{ university.level }}
              </el-tag>
              <el-tag size="small" type="info">{{ university.type }}</el-tag>
            </div>
          </div>
        </div>
        <div class="card-actions" @click.stop>
          <el-tooltip content="收藏" placement="top">
            <el-button
              circle
              size="small"
              :type="university.isFavorite ? 'danger' : 'default'"
              @click="$emit('favorite', university)"
            >
              <el-icon>
                <star-filled v-if="university.isFavorite" />
                <star v-else />
              </el-icon>
            </el-button>
          </el-tooltip>
          <el-tooltip content="加入对比" placement="top">
            <el-button
              circle
              size="small"
              type="primary"
              @click="$emit('compare', university)"
            >
              <el-icon><scale-to-original /></el-icon>
            </el-button>
          </el-tooltip>
        </div>
      </div>
    </template>

    <div class="card-content">
      <div class="location-rank">
        <div class="location">
          <el-icon><location /></el-icon>
          <span>{{ university.province }} {{ university.city }}</span>
        </div>
        <div class="rank" v-if="university.rank">
          <el-icon><trophy /></el-icon>
          <span>全国排名 #{{ university.rank }}</span>
        </div>
      </div>

      <div class="score-info">
        <div class="score-item">
          <span class="label">理科分数线</span>
          <span class="value science">{{
            university.minScoreScience || '--'
          }}</span>
        </div>
        <div class="score-item">
          <span class="label">文科分数线</span>
          <span class="value liberal">{{
            university.minScoreLiberalArts || '--'
          }}</span>
        </div>
      </div>

      <div class="features">
        <el-tag
          v-for="feature in university.features?.slice(0, 3)"
          :key="feature"
          size="small"
          effect="plain"
        >
          {{ feature }}
        </el-tag>
        <span
          v-if="university.features && university.features.length > 3"
          class="more-features"
        >
          +{{ university.features.length - 3 }}
        </span>
      </div>

      <div class="stats">
        <div class="stat-item">
          <div class="stat-value">{{ university.studentCount || '--' }}</div>
          <div class="stat-label">在校生</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ university.majorCount || '--' }}</div>
          <div class="stat-label">专业数</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ university.employmentRate || '--' }}%</div>
          <div class="stat-label">就业率</div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="card-footer">
        <el-button
          type="primary"
          size="small"
          @click.stop="$emit('view', university.id)"
        >
          查看详情
        </el-button>
        <el-button size="small" @click.stop="viewAdmissionData">
          录取数据
        </el-button>
      </div>
    </template>
  </el-card>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import {
  Star,
  StarFilled,
  ScaleToOriginal,
  Location,
  Trophy,
} from '@element-plus/icons-vue';
import type { University } from '@/types/university';

interface Props {
  university: University;
}

const props = defineProps<Props>();

const router = useRouter();

// 获取标签类型
const getTagType = (level: string) => {
  const levelTypes: Record<string, string> = {
    '985工程': 'danger',
    '211工程': 'warning',
    双一流: 'success',
    普通本科: 'info',
  };
  return levelTypes[level] || 'info';
};

// 处理图片加载错误
const handleImageError = (event: Event) => {
  const target = event.target as HTMLImageElement;
  target.src = '/default-logo.svg';
};

// 查看录取数据
const viewAdmissionData = () => {
  router.push(`/analysis?university=${props.university.id}`);
};
</script>

<style scoped>
.university-card {
  height: 100%;
  cursor: pointer;
  transition: all 0.3s ease;
  border-radius: 12px;
  overflow: hidden;
}

.university-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 0;
}

.university-basic {
  display: flex;
  align-items: center;
  flex: 1;
}

.university-logo {
  width: 50px;
  height: 50px;
  border-radius: 8px;
  object-fit: cover;
  margin-right: 12px;
  border: 1px solid #ebeef5;
}

.university-info {
  flex: 1;
  min-width: 0;
}

.university-name {
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
  margin: 0 0 8px 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.university-meta {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.card-actions {
  display: flex;
  gap: 6px;
}

.card-content {
  padding: 0;
}

.location-rank {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  font-size: 14px;
  color: #7f8c8d;
}

.location,
.rank {
  display: flex;
  align-items: center;
  gap: 4px;
}

.score-info {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
}

.score-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.score-item .label {
  font-size: 12px;
  color: #7f8c8d;
  margin-bottom: 4px;
}

.score-item .value {
  font-size: 18px;
  font-weight: 600;
}

.score-item .value.science {
  color: #e74c3c;
}

.score-item .value.liberal {
  color: #3498db;
}

.features {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 16px;
}

.more-features {
  font-size: 12px;
  color: #7f8c8d;
  background: #f8f9fa;
  padding: 2px 6px;
  border-radius: 4px;
}

.stats {
  display: flex;
  justify-content: space-around;
  background: #f8f9fa;
  border-radius: 8px;
  padding: 12px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #7f8c8d;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  padding: 0;
}

.card-footer .el-button {
  flex: 1;
  margin: 0 4px;
}

.card-footer .el-button:first-child {
  margin-left: 0;
}

.card-footer .el-button:last-child {
  margin-right: 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .university-name {
    font-size: 14px;
  }

  .location-rank {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .score-info {
    flex-direction: column;
    gap: 8px;
  }

  .score-item {
    flex-direction: row;
    justify-content: space-between;
  }
}
</style>
