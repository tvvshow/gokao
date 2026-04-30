<template>
  <el-card
    class="skeleton-card"
    :class="{ 'skeleton-card--animated': animated }"
  >
    <template #header v-if="showHeader">
      <div class="skeleton-header">
        <div
          class="skeleton-avatar"
          :style="{ width: avatarSize, height: avatarSize }"
        ></div>
        <div class="skeleton-header-content">
          <div class="skeleton-line skeleton-line--title"></div>
          <div class="skeleton-line skeleton-line--subtitle"></div>
        </div>
      </div>
    </template>

    <div class="skeleton-body">
      <div
        v-for="n in lines"
        :key="n"
        class="skeleton-line"
        :style="{ width: getLineWidth(n) }"
      ></div>
    </div>

    <template #footer v-if="showFooter">
      <div class="skeleton-footer">
        <div class="skeleton-button"></div>
        <div class="skeleton-button"></div>
      </div>
    </template>
  </el-card>
</template>

<script setup lang="ts">
interface Props {
  lines?: number;
  showHeader?: boolean;
  showFooter?: boolean;
  avatarSize?: string;
  animated?: boolean;
}

withDefaults(defineProps<Props>(), {
  lines: 3,
  showHeader: true,
  showFooter: true,
  avatarSize: '50px',
  animated: true,
});

// Generate varying line widths for more natural look
const getLineWidth = (index: number): string => {
  const widths = ['100%', '85%', '70%', '90%', '60%'];
  return widths[(index - 1) % widths.length];
};
</script>

<style scoped>
.skeleton-card {
  height: 100%;
  border-radius: 12px;
  overflow: hidden;
}

.skeleton-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.skeleton-avatar {
  border-radius: 8px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e2e8f0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  flex-shrink: 0;
}

.skeleton-header-content {
  flex: 1;
  min-width: 0;
}

.skeleton-line {
  height: 16px;
  border-radius: 0.5rem;
  background: linear-gradient(90deg, #f0f0f0 25%, #e2e8f0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  margin-bottom: 12px;
}

.skeleton-line:last-child {
  margin-bottom: 0;
}

.skeleton-line--title {
  height: 20px;
  width: 60%;
  margin-bottom: 8px;
}

.skeleton-line--subtitle {
  height: 14px;
  width: 40%;
  margin-bottom: 0;
}

.skeleton-body {
  padding: 8px 0;
}

.skeleton-footer {
  display: flex;
  justify-content: space-between;
  gap: 8px;
}

.skeleton-button {
  flex: 1;
  height: 32px;
  border-radius: 0.5rem;
  background: linear-gradient(90deg, #f0f0f0 25%, #e2e8f0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
}

/* Animation */
.skeleton-card--animated .skeleton-avatar,
.skeleton-card--animated .skeleton-line,
.skeleton-card--animated .skeleton-button {
  animation: skeleton-loading 1.5s infinite;
}

@keyframes skeleton-loading {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

/* Dark mode support */
:deep(.dark) .skeleton-avatar,
:deep(.dark) .skeleton-line,
:deep(.dark) .skeleton-button {
  background: linear-gradient(90deg, #374151 25%, #4b5563 50%, #374151 75%);
  background-size: 200% 100%;
}
</style>
