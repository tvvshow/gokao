<template>
  <div id="app" class="min-h-screen flex flex-col">
    <!-- 现代化导航栏 -->
    <AppHeader />

    <!-- 主内容区域 -->
    <main class="flex-1 container-modern py-8">
      <router-view v-slot="{ Component }">
        <transition
          name="page"
          mode="out-in"
          enter-active-class="animate-fade-in"
          leave-active-class="animate-fade-out"
        >
          <component :is="Component" />
        </transition>
      </router-view>
    </main>

    <!-- 现代化页脚 -->
    <AppFooter />

    <!-- 暗色模式切换按钮 -->
    <ThemeToggle />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useDark, useToggle } from '@vueuse/core'
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
import ThemeToggle from '@/components/ThemeToggle.vue'

// 暗色模式支持
const isDark = useDark()
const toggleDark = useToggle(isDark)

// 初始化主题
onMounted(() => {
  // 应用主题类到 html 元素
  document.documentElement.classList.toggle('dark', isDark.value)
})
</script>

<style scoped>
/* 页面过渡动画 */
.page-enter-active,
.page-leave-active {
  transition: all 0.3s ease-in-out;
}

.page-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.page-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

.animate-fade-out {
  animation: fadeOut 0.3s ease-in-out;
}

@keyframes fadeOut {
  from {
    opacity: 1;
  }
  to {
    opacity: 0;
  }
}
</style>