<template>
  <el-dialog
    v-model="dialogVisible"
    title="院校对比分析"
    width="90%"
    top="5vh"
    @close="$emit('update:modelValue', false)"
  >
    <div class="compare-content" v-if="universities.length > 0">
      <!-- 对比表格 -->
      <el-table
        :data="comparisonData"
        border
        style="width: 100%"
        max-height="500"
      >
        <el-table-column prop="field" label="对比项" width="120" fixed="left" />
        <el-table-column
          v-for="university in universities"
          :key="university.id"
          :label="university.name"
          :prop="university.id"
          align="center"
          min-width="150"
        >
          <template #header>
            <div class="university-header">
              <img
                :src="university.logo || '/default-logo.svg'"
                :alt="university.name"
                class="university-logo"
              />
              <div>
                <div class="university-name">{{ university.name }}</div>
                <div class="university-location">{{ university.province }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 雷达图对比 -->
      <div class="radar-chart" style="margin-top: 20px">
        <h3>综合对比雷达图</h3>
        <div id="radarChart" style="height: 400px"></div>
      </div>
    </div>

    <div class="empty-state" v-else>
      <el-empty description="请先选择要对比的院校" />
    </div>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="$emit('clear')">清空对比</el-button>
        <el-button type="primary" @click="exportComparison">导出对比</el-button>
        <el-button @click="dialogVisible = false">关闭</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import type { University } from '@/types/university';

interface Props {
  modelValue: boolean;
  universities: University[];
}

const props = defineProps<Props>();
const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  clear: [];
}>();

const dialogVisible = ref(false);

// 监听 modelValue 变化（immediate确保初始值同步）
watch(
  () => props.modelValue,
  (newVal) => {
    dialogVisible.value = newVal;
  },
  { immediate: true }
);

watch(dialogVisible, (newVal) => {
  emit('update:modelValue', newVal);
});

// 对比数据
const comparisonData = computed(() => {
  if (props.universities.length === 0) return [];

  const fields = [
    { field: '院校类型', key: 'type' },
    { field: '办学层次', key: 'level' },
    { field: '全国排名', key: 'rank' },
    { field: '理科分数线', key: 'minScoreScience' },
    { field: '文科分数线', key: 'minScoreLiberalArts' },
    { field: '在校生人数', key: 'studentCount' },
    { field: '专业数量', key: 'majorCount' },
    { field: '就业率', key: 'employmentRate' },
    { field: '建校年份', key: 'founded' },
  ];

  return fields.map((item) => {
    const row: Record<string, string | number | boolean | undefined> = {
      field: item.field,
    };
    props.universities.forEach((university) => {
      const value = university[item.key as keyof University];
      if (value === undefined || value === null) {
        row[university.id] = '--';
      } else if (Array.isArray(value)) {
        row[university.id] = value.join(', ');
      } else {
        row[university.id] = value as string | number | boolean;
      }
    });
    return row;
  });
});

// 导出对比报告
const exportComparison = () => {
  // 实现导出功能
  console.log('导出对比报告');
};
</script>

<style scoped>
.compare-content {
  max-height: 70vh;
  overflow-y: auto;
}

.university-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.university-logo {
  width: 30px;
  height: 30px;
  border-radius: 4px;
  object-fit: cover;
}

.university-name {
  font-weight: 600;
  font-size: 14px;
}

.university-location {
  font-size: 12px;
  color: #999;
}

.empty-state {
  text-align: center;
  padding: 40px;
}

.radar-chart h3 {
  text-align: center;
  margin-bottom: 20px;
  color: #2c3e50;
}
</style>
