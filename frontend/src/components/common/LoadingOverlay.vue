<template>
  <Transition name="fade">
    <div
      v-if="loading"
      class="loading-overlay"
      :class="{ 'loading-overlay--fullscreen': fullscreen }"
    >
      <div class="loading-content">
        <el-icon class="loading-icon" :size="iconSize">
          <Loading />
        </el-icon>
        <p v-if="text" class="loading-text">{{ text }}</p>
        <el-progress
          v-if="showProgress && progress !== undefined"
          :percentage="progress"
          :stroke-width="6"
          :show-text="showProgressText"
          class="loading-progress"
        />
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { Loading } from '@element-plus/icons-vue';

interface Props {
  loading?: boolean;
  text?: string;
  fullscreen?: boolean;
  iconSize?: number;
  showProgress?: boolean;
  progress?: number;
  showProgressText?: boolean;
}

withDefaults(defineProps<Props>(), {
  loading: false,
  text: '',
  fullscreen: false,
  iconSize: 40,
  showProgress: false,
  progress: undefined,
  showProgressText: true,
});
</script>

<style scoped>
.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  border-radius: inherit;
}

.loading-overlay--fullscreen {
  position: fixed;
  z-index: 2000;
  border-radius: 0;
}

.loading-content {
  text-align: center;
}

.loading-icon {
  color: #0ea5e9;
  animation: rotate 1.5s linear infinite;
}

.loading-text {
  margin-top: 16px;
  color: #606266;
  font-size: 14px;
}

.loading-progress {
  margin-top: 16px;
  width: 200px;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* Transition */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Dark mode */
:deep(.dark) .loading-overlay {
  background: rgba(31, 41, 55, 0.9);
}

:deep(.dark) .loading-text {
  color: #d1d5db;
}
</style>
