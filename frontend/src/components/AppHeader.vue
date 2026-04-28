<template>
  <header class="nav-modern">
    <div class="container-modern">
      <div class="flex-between py-4">
        <!-- Logo 区域 -->
        <div class="flex items-center space-x-4">
          <router-link to="/" class="flex items-center space-x-3 group">
            <div class="logo-container">
              <GraduationCapIcon
                class="w-8 h-8 text-primary-600 group-hover:text-primary-700 transition-colors"
              />
            </div>
            <div class="flex flex-col">
              <span class="text-xl font-bold text-gradient">高考志愿填报</span>
              <span class="text-xs text-gray-500 dark:text-gray-400"
                >智能推荐系统</span
              >
            </div>
          </router-link>
        </div>

        <!-- 导航菜单 -->
        <nav class="hidden md:flex items-center space-x-1" aria-label="主导航">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="nav-link-modern"
            :class="{ active: $route.path === item.path }"
            :aria-current="$route.path === item.path ? 'page' : undefined"
          >
            <component :is="item.icon" class="w-4 h-4" aria-hidden="true" />
            <span>{{ item.name }}</span>
          </router-link>
        </nav>

        <!-- 用户区域 -->
        <div class="flex items-center space-x-4">
          <!-- 通知按钮 -->
          <button
            class="btn-icon"
            title="通知"
            aria-label="查看通知，有3条未读消息"
          >
            <BellIcon class="w-5 h-5" aria-hidden="true" />
            <span class="notification-badge" aria-hidden="true">3</span>
          </button>

          <!-- 用户菜单 -->
          <template v-if="userStore.isLoggedIn">
            <el-dropdown @command="handleUserCommand" trigger="click">
              <div class="user-avatar-container">
                <el-avatar
                  :size="36"
                  :src="userStore.user?.avatar"
                  class="ring-2 ring-primary-200 dark:ring-primary-800"
                >
                  {{ userStore.user?.username?.charAt(0) }}
                </el-avatar>
                <div class="status-indicator"></div>
              </div>
              <template #dropdown>
                <el-dropdown-menu class="user-dropdown">
                  <div class="user-info-header">
                    <el-avatar :size="48" :src="userStore.user?.avatar">
                      {{ userStore.user?.username?.charAt(0) }}
                    </el-avatar>
                    <div class="ml-3">
                      <div class="font-medium text-gray-900 dark:text-gray-100">
                        {{ userStore.user?.username }}
                      </div>
                      <div class="text-sm text-gray-500 dark:text-gray-400">
                        {{ userStore.user?.email || '用户' }}
                      </div>
                    </div>
                  </div>
                  <el-dropdown-item command="profile" class="dropdown-item">
                    <UserIcon class="w-4 h-4" />
                    个人中心
                  </el-dropdown-item>
                  <!-- 暂时隐藏会员中心，支付功能待开发
                  <el-dropdown-item command="membership" class="dropdown-item">
                    <CrownIcon class="w-4 h-4" />
                    会员中心
                  </el-dropdown-item>
                  -->
                  <el-dropdown-item command="settings" class="dropdown-item">
                    <SettingsIcon class="w-4 h-4" />
                    设置
                  </el-dropdown-item>
                  <el-dropdown-item
                    command="logout"
                    divided
                    class="dropdown-item text-red-600"
                  >
                    <LogOutIcon class="w-4 h-4" />
                    退出登录
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
          <template v-else>
            <router-link to="/login" class="btn btn-ghost">登录</router-link>
            <router-link to="/register" class="btn btn-primary"
              >注册</router-link
            >
          </template>

          <!-- 移动端菜单按钮 -->
          <button
            @click="toggleMobileMenu"
            class="md:hidden btn-icon"
            :aria-label="mobileMenuOpen ? '关闭导航菜单' : '打开导航菜单'"
            :aria-expanded="mobileMenuOpen"
          >
            <MenuIcon
              v-if="!mobileMenuOpen"
              class="w-5 h-5"
              aria-hidden="true"
            />
            <XIcon v-else class="w-5 h-5" aria-hidden="true" />
          </button>
        </div>
      </div>

      <!-- 移动端导航菜单 -->
      <transition name="mobile-menu">
        <div
          v-if="mobileMenuOpen"
          class="md:hidden py-4 border-t border-gray-200 dark:border-gray-700"
        >
          <nav class="space-y-2" aria-label="移动端导航">
            <router-link
              v-for="item in navItems"
              :key="item.path"
              :to="item.path"
              class="mobile-nav-link"
              :aria-current="$route.path === item.path ? 'page' : undefined"
              @click="closeMobileMenu"
            >
              <component :is="item.icon" class="w-5 h-5" aria-hidden="true" />
              <span>{{ item.name }}</span>
            </router-link>
          </nav>
        </div>
      </transition>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '@/stores/user';
import {
  GraduationCapIcon,
  HomeIcon,
  BuildingIcon,
  BookOpenIcon,
  SparklesIcon,
  BarChartIcon,
  BellIcon,
  UserIcon,
  SettingsIcon,
  LogOutIcon,
  MenuIcon,
  XIcon,
} from 'lucide-vue-next';

const router = useRouter();
const userStore = useUserStore();

// 移动端菜单状态
const mobileMenuOpen = ref(false);

// 导航菜单项
const navItems = [
  { path: '/', name: '首页', icon: HomeIcon },
  { path: '/universities', name: '院校查询', icon: BuildingIcon },
  { path: '/majors', name: '专业分析', icon: BookOpenIcon },
  { path: '/recommendation', name: '智能推荐', icon: SparklesIcon },
  { path: '/analysis', name: '数据分析', icon: BarChartIcon },
  // { path: '/membership', name: '会员服务', icon: CrownIcon }, // 暂时隐藏，支付功能待开发
];

const toggleMobileMenu = () => {
  mobileMenuOpen.value = !mobileMenuOpen.value;
};

const closeMobileMenu = () => {
  mobileMenuOpen.value = false;
};

const handleUserCommand = (command: string) => {
  switch (command) {
    case 'profile':
      router.push('/profile');
      break;
    // case 'membership':  // 暂时隐藏，支付功能待开发
    //   router.push('/membership');
    //   break;
    case 'settings':
      router.push('/profile');
      break;
    case 'logout':
      userStore.logout();
      router.push('/');
      break;
  }
};
</script>

<style scoped>
/* 现代化导航链接 */
.nav-link-modern {
  @apply flex items-center space-x-2 px-4 py-2 rounded-lg text-sm font-medium;
  @apply text-gray-700 dark:text-gray-300;
  @apply hover:bg-gray-100 dark:hover:bg-gray-800;
  @apply transition-all duration-200 ease-in-out;
  @apply focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2;
}

.nav-link-modern.active {
  @apply bg-primary-100 dark:bg-primary-900;
  @apply text-primary-700 dark:text-primary-300;
  @apply shadow-soft;
}

.nav-link-modern:hover {
  @apply transform scale-105;
}

/* 图标按钮 */
.btn-icon {
  @apply relative p-2 rounded-lg text-gray-600 dark:text-gray-400;
  @apply hover:bg-gray-100 dark:hover:bg-gray-800;
  @apply focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2;
  @apply transition-all duration-200 ease-in-out;
}

.btn-icon:hover {
  @apply text-gray-900 dark:text-gray-100;
}

/* 通知徽章 */
.notification-badge {
  @apply absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs;
  @apply rounded-full flex items-center justify-center;
  @apply animate-pulse;
}

/* 用户头像容器 */
.user-avatar-container {
  @apply relative cursor-pointer;
}

.status-indicator {
  @apply absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-green-400 rounded-full;
  @apply border-2 border-white dark:border-gray-900;
}

/* 用户下拉菜单 */
.user-dropdown {
  @apply min-w-64;
}

.user-info-header {
  @apply flex items-center p-4 border-b border-gray-200 dark:border-gray-700;
  @apply bg-gray-50 dark:bg-gray-800;
}

.dropdown-item {
  @apply flex items-center space-x-2 px-4 py-2;
  @apply hover:bg-gray-50 dark:hover:bg-gray-800;
  @apply transition-colors duration-150;
}

/* 移动端导航链接 */
.mobile-nav-link {
  @apply flex items-center space-x-3 px-4 py-3 rounded-lg;
  @apply text-gray-700 dark:text-gray-300;
  @apply hover:bg-gray-100 dark:hover:bg-gray-800;
  @apply transition-all duration-200 ease-in-out;
}

/* 移动端菜单过渡动画 */
.mobile-menu-enter-active,
.mobile-menu-leave-active {
  transition: all 0.3s ease-in-out;
}

.mobile-menu-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.mobile-menu-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* Logo 容器 */
.logo-container {
  @apply p-2 rounded-lg bg-primary-50 dark:bg-primary-900/20;
  @apply group-hover:bg-primary-100 dark:group-hover:bg-primary-900/30;
  @apply transition-colors duration-200;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .container-modern {
    @apply px-4;
  }
}

/* Element Plus 样式覆盖 */
:deep(.el-dropdown-menu) {
  @apply border-0 shadow-large rounded-xl;
  @apply bg-white/95 dark:bg-gray-800/95 backdrop-blur-sm;
}

:deep(.el-dropdown-menu__item) {
  @apply text-gray-700 dark:text-gray-300;
  @apply hover:bg-gray-50 dark:hover:bg-gray-700;
}

:deep(.el-dropdown-menu__item:hover) {
  @apply text-gray-900 dark:text-gray-100;
}

:deep(.el-dropdown-menu__item.is-divided) {
  @apply border-t border-gray-200 dark:border-gray-700;
}
</style>
