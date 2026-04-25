<template>
  <div class="error-boundary">
    <slot v-if="!hasError" />
    <div v-else class="error-fallback">
      <el-result icon="error" :title="errorTitle" :sub-title="errorMessage">
        <template #extra>
          <el-button type="primary" @click="handleRetry">
            <el-icon><RefreshRight /></el-icon>
            重试
          </el-button>
          <el-button @click="handleGoHome">
            <el-icon><HomeFilled /></el-icon>
            返回首页
          </el-button>
        </template>
      </el-result>
      <div v-if="showDetails && errorDetails" class="error-details">
        <el-collapse>
          <el-collapse-item title="错误详情" name="details">
            <pre class="error-stack">{{ errorDetails }}</pre>
          </el-collapse-item>
        </el-collapse>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onErrorCaptured } from 'vue';
import { useRouter } from 'vue-router';
import { RefreshRight, HomeFilled } from '@element-plus/icons-vue';

interface Props {
  fallbackTitle?: string;
  fallbackMessage?: string;
  showDetails?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  fallbackTitle: '页面出错了',
  fallbackMessage: '抱歉，页面加载时发生了错误',
  showDetails: false,
});

const emit = defineEmits<{
  error: [error: Error];
  retry: [];
}>();
const router = useRouter();

const hasError = ref(false);
const errorTitle = ref(props.fallbackTitle);
const errorMessage = ref(props.fallbackMessage);
const errorDetails = ref<string | null>(null);

// Capture errors from child components
onErrorCaptured((error: Error, instance, info) => {
  hasError.value = true;
  errorTitle.value = props.fallbackTitle;
  errorMessage.value = error.message || props.fallbackMessage;
  errorDetails.value = `${error.stack}\n\nComponent: ${instance?.$options?.name || 'Unknown'}\nInfo: ${info}`;

  emit('error', error);

  // Prevent error from propagating
  return false;
});

const handleRetry = () => {
  hasError.value = false;
  errorDetails.value = null;
  emit('retry');
};

const handleGoHome = () => {
  hasError.value = false;
  router.push('/');
};

// Expose reset method for parent components
defineExpose({
  reset: () => {
    hasError.value = false;
    errorDetails.value = null;
  },
});
</script>

<style scoped>
.error-boundary {
  min-height: 200px;
}

.error-fallback {
  padding: 40px 20px;
  text-align: center;
}

.error-details {
  max-width: 600px;
  margin: 20px auto 0;
  text-align: left;
}

.error-stack {
  background: #f5f7fa;
  padding: 16px;
  border-radius: 4px;
  font-size: 12px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
  color: #606266;
}

/* Dark mode */
:deep(.dark) .error-stack {
  background: #374151;
  color: #d1d5db;
}
</style>
