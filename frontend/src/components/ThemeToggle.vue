<template>
  <div class="fixed bottom-6 right-6 z-50">
    <button
      @click="toggleTheme"
      class="btn-theme-toggle"
      :class="{ 'dark': isDark }"
      :title="isDark ? '切换到亮色模式' : '切换到暗色模式'"
    >
      <transition name="icon" mode="out-in">
        <SunIcon v-if="isDark" key="sun" class="w-5 h-5" />
        <MoonIcon v-else key="moon" class="w-5 h-5" />
      </transition>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useDark, useToggle } from '@vueuse/core'
import { SunIcon, MoonIcon } from 'lucide-vue-next'

// 暗色模式状态
const isDark = useDark()
const toggleTheme = useToggle(isDark)
</script>

<style scoped>
.btn-theme-toggle {
  @apply w-12 h-12 rounded-full shadow-large;
  @apply bg-white dark:bg-gray-800;
  @apply text-gray-700 dark:text-gray-300;
  @apply border border-gray-200 dark:border-gray-700;
  @apply flex items-center justify-center;
  @apply transition-all duration-300 ease-in-out;
  @apply hover:scale-110 hover:shadow-glow;
  @apply focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2;
  @apply backdrop-blur-sm;
}

.btn-theme-toggle:hover {
  @apply bg-gray-50 dark:bg-gray-700;
  @apply transform rotate-12;
}

.btn-theme-toggle.dark {
  @apply shadow-glow;
}

/* 图标过渡动画 */
.icon-enter-active,
.icon-leave-active {
  transition: all 0.3s ease-in-out;
}

.icon-enter-from {
  opacity: 0;
  transform: rotate(-90deg) scale(0.8);
}

.icon-leave-to {
  opacity: 0;
  transform: rotate(90deg) scale(0.8);
}
</style>
