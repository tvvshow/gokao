import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, LoginForm, RegisterForm } from '@/types/user'
import { userApi } from '@/api/user'

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const loading = ref(false)

  const isLoggedIn = computed(() => !!token.value)

  // 初始化，从localStorage恢复登录状态
  const init = () => {
    const savedToken = localStorage.getItem('token')
    const savedUser = localStorage.getItem('user')
    
    if (savedToken && savedUser) {
      token.value = savedToken
      user.value = JSON.parse(savedUser)
    }
  }

  // 登录
  const login = async (loginForm: LoginForm) => {
    loading.value = true
    try {
      const response = await userApi.login(loginForm)
      
      if (response.success) {
        token.value = response.data.token
        user.value = response.data.user
        
        // 保存到localStorage
        localStorage.setItem('token', response.data.token)
        localStorage.setItem('user', JSON.stringify(response.data.user))
        
        return { success: true }
      } else {
        return { success: false, message: response.message }
      }
    } catch (error: any) {
      return { 
        success: false, 
        message: error.message || '登录失败，请稍后重试' 
      }
    } finally {
      loading.value = false
    }
  }

  // 注册
  const register = async (registerForm: RegisterForm) => {
    loading.value = true
    try {
      const response = await userApi.register(registerForm)
      
      if (response.success) {
        return { success: true, message: '注册成功，请登录' }
      } else {
        return { success: false, message: response.message }
      }
    } catch (error: any) {
      return { 
        success: false, 
        message: error.message || '注册失败，请稍后重试' 
      }
    } finally {
      loading.value = false
    }
  }

  // 退出登录
  const logout = () => {
    user.value = null
    token.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  // 更新用户信息
  const updateProfile = async (userData: Partial<User>) => {
    loading.value = true
    try {
      const response = await userApi.updateProfile(userData)
      
      if (response.success && user.value) {
        user.value = { ...user.value, ...response.data }
        localStorage.setItem('user', JSON.stringify(user.value))
        return { success: true }
      } else {
        return { success: false, message: response.message }
      }
    } catch (error: any) {
      return { 
        success: false, 
        message: error.message || '更新失败，请稍后重试' 
      }
    } finally {
      loading.value = false
    }
  }

  // 获取会员信息
  const getMembershipInfo = async () => {
    try {
      const response = await userApi.getMembershipInfo()
      return response
    } catch (error: any) {
      return { 
        success: false, 
        message: error.message || '获取会员信息失败' 
      }
    }
  }

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
    getMembershipInfo
  }
})