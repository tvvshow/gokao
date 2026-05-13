import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { ElMessage } from 'element-plus';
import type { User, LoginForm, RegisterForm } from '@/types/user';
import { userApi } from '@/api/user';
import { authEvents, type ForceLogoutDetail } from '@/utils/auth-events';
import router from '@/router';

const TOKEN_KEY = 'auth_token';
const REFRESH_TOKEN_KEY = 'refresh_token';
const USER_KEY = 'user';

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null);
  const token = ref<string | null>(null);
  const loading = ref(false);

  const isLoggedIn = computed(() => !!token.value);

  // 单一清理入口：内存 state + localStorage 三件套同时清，避免任一侧滞留。
  const clearAuthState = () => {
    user.value = null;
    token.value = null;
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  };

  // 被动下线（api-client 401 + 刷新失败时由事件触发）。
  // 主动退出（用户点按钮）走 logout()；两者共用 clearAuthState。
  const handleForceLogout = (detail: ForceLogoutDetail) => {
    clearAuthState();
    ElMessage.error('登录已过期，请重新登录');
    const currentPath = router.currentRoute.value.fullPath;
    if (currentPath !== '/login') {
      router.push({
        path: '/login',
        query: detail.redirect ? { redirect: detail.redirect } : {},
      });
    }
  };

  let unsubscribeForceLogout: (() => void) | null = null;

  // 初始化，从localStorage恢复登录状态，并订阅强制下线事件
  const init = () => {
    const savedToken = localStorage.getItem(TOKEN_KEY);
    const savedUser = localStorage.getItem(USER_KEY);

    if (savedToken && savedUser) {
      token.value = savedToken;
      try {
        user.value = JSON.parse(savedUser);
      } catch {
        // 持久化数据损坏：直接清理，避免半状态。
        clearAuthState();
      }
    }

    // 幂等订阅，防止热更新/重复 init 重复挂监听。
    if (!unsubscribeForceLogout) {
      unsubscribeForceLogout = authEvents.on('force-logout', handleForceLogout);
    }
  };

  // 登录
  const login = async (loginForm: LoginForm) => {
    loading.value = true;
    try {
      const response = await userApi.login(loginForm);

      if (response.success) {
        token.value = response.data.token;
        user.value = response.data.user;

        localStorage.setItem(TOKEN_KEY, response.data.token);
        localStorage.setItem(USER_KEY, JSON.stringify(response.data.user));

        if (response.data.refreshToken) {
          localStorage.setItem(REFRESH_TOKEN_KEY, response.data.refreshToken);
        }

        return { success: true };
      } else {
        return { success: false, message: response.message };
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : '登录失败，请稍后重试';
      return {
        success: false,
        message: errorMessage,
      };
    } finally {
      loading.value = false;
    }
  };

  // 注册
  const register = async (registerForm: RegisterForm) => {
    loading.value = true;
    try {
      const response = await userApi.register(registerForm);

      if (response.success) {
        return { success: true, message: '注册成功，请登录' };
      } else {
        return { success: false, message: response.message };
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : '注册失败，请稍后重试';
      return {
        success: false,
        message: errorMessage,
      };
    } finally {
      loading.value = false;
    }
  };

  // 主动退出登录：先尝试通知后端撤销 refresh_token，再清本地。
  const logout = async () => {
    try {
      await userApi.logout();
    } catch {
      // 后端失败不阻塞本地清理；refresh_token 已过期是预期路径之一。
    }
    clearAuthState();
  };

  // 更新用户信息
  const updateProfile = async (userData: Partial<User>) => {
    loading.value = true;
    try {
      const response = await userApi.updateProfile(userData);

      if (response.success && user.value) {
        user.value = { ...user.value, ...response.data };
        localStorage.setItem(USER_KEY, JSON.stringify(user.value));
        return { success: true };
      } else {
        return { success: false, message: response.message };
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : '更新失败，请稍后重试';
      return {
        success: false,
        message: errorMessage,
      };
    } finally {
      loading.value = false;
    }
  };

  // 获取会员信息
  const getMembershipInfo = async () => {
    try {
      const response = await userApi.getMembershipInfo();
      return response;
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : '获取会员信息失败';
      return {
        success: false,
        message: errorMessage,
      };
    }
  };

  return {
    user,
    token,
    loading,
    isLoggedIn,
    init,
    login,
    register,
    logout,
    updateProfile,
    getMembershipInfo,
  };
});
