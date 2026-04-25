<template>
  <div class="skeleton-list">
    <el-row :gutter="gutter">
      <el-col
        v-for="n in count"
        :key="n"
        :xs="colSpan.xs"
        :sm="colSpan.sm"
        :md="colSpan.md"
        :lg="colSpan.lg"
        class="skeleton-col"
      >
        <SkeletonCard
          :lines="lines"
          :show-header="showHeader"
          :show-footer="showFooter"
          :avatar-size="avatarSize"
          :animated="animated"
        />
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import SkeletonCard from './SkeletonCard.vue';

interface ColSpan {
  xs?: number;
  sm?: number;
  md?: number;
  lg?: number;
}

interface Props {
  count?: number;
  gutter?: number;
  colSpan?: ColSpan;
  lines?: number;
  showHeader?: boolean;
  showFooter?: boolean;
  avatarSize?: string;
  animated?: boolean;
}

withDefaults(defineProps<Props>(), {
  count: 4,
  gutter: 20,
  colSpan: () => ({ xs: 24, sm: 12, md: 8, lg: 6 }),
  lines: 3,
  showHeader: true,
  showFooter: true,
  avatarSize: '50px',
  animated: true,
});
</script>

<style scoped>
.skeleton-list {
  min-height: 200px;
}

.skeleton-col {
  margin-bottom: 20px;
}
</style>
