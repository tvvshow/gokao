import { ElMessage } from 'element-plus'
import axios from 'axios'
import type { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import router from '@/router'

// API基础配置 - 连接到API Gateway
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const API_TIMEOUT = 10000

// Token存储键名
const TOKEN_KEY = 'auth_token'
const REFRESH_TOKEN_KEY = 'refresh_token'

// 统一API响应格式
interface ApiResponse<T = any> {
  success: boolean
  data: T
  message: string
  total?: number
}

// Token刷新响应格式
interface RefreshTokenResponse {
  success: boolean
  data: {
    token: string
    refreshToken?: string
  }
  message?: string
}

// API客户端类
class ApiClient {
  private axiosInstance: AxiosInstance
  private baseURL: string
  private isRefreshing: boolean = false
  private refreshSubscribers: Array<(token: string) => void> = []

  constructor(baseURL: string, timeout: number = 10000) {
    this.baseURL = baseURL
    this.axiosInstance = axios.create({
      baseURL,
      timeout,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // 请求拦截器 - 添加Bearer Token
    this.axiosInstance.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        const token = localStorage.getItem(TOKEN_KEY)
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => {
        return Promise.reject(error)
      }
    )

    // 响应拦截器 - 统一错误处理和Token刷新
    this.axiosInstance.interceptors.response.use(
      (response: AxiosResponse<ApiResponse>) => {
        return response
      },
      async (error) => {
        const originalRequest = error.config

        // 处理401错误 - Token过期
        if (error.response?.status === 401 && !originalRequest._retry) {
          // 如果是刷新Token的请求失败，直接跳转登录
          if (originalRequest.url?.includes('/auth/refresh')) {
            this.handleAuthFailure()
            return Promise.reject(error)
          }

          // 尝试刷新Token
          if (!this.isRefreshing) {
            this.isRefreshing = true
            originalRequest._retry = true

            try {
              const newToken = await this.refreshToken()
              if (newToken) {
                // 通知所有等待的请求
                this.onRefreshed(newToken)
                // 重试原始请求
                originalRequest.headers.Authorization = `Bearer ${newToken}`
                return this.axiosInstance(originalRequest)
              }
            } catch (refreshError) {
              this.handleAuthFailure()
              return Promise.reject(refreshError)
            } finally {
              this.isRefreshing = false
            }
          } else {
            // 等待Token刷新完成
            return new Promise((resolve) => {
              this.subscribeTokenRefresh((token: string) => {
                originalRequest.headers.Authorization = `Bearer ${token}`
                resolve(this.axiosInstance(originalRequest))
              })
            })
          }
        }

        this.handleError(error)
        return Promise.reject(error)
      }
    )
  }

  // 刷新Token
  private async refreshToken(): Promise<string | null> {
    const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)
    if (!refreshToken) {
      return null
    }

    try {
      const response = await axios.post<RefreshTokenResponse>(
        `${this.baseURL}/api/v1/users/auth/refresh`,
        { refreshToken },
        { headers: { 'Content-Type': 'application/json' } }
      )

      if (response.data.success && response.data.data.token) {
        const newToken = response.data.data.token
        localStorage.setItem(TOKEN_KEY, newToken)
        
        // 如果返回了新的refreshToken，也更新
        if (response.data.data.refreshToken) {
          localStorage.setItem(REFRESH_TOKEN_KEY, response.data.data.refreshToken)
        }
        
        return newToken
      }
      return null
    } catch {
      return null
    }
  }

  // 订阅Token刷新
  private subscribeTokenRefresh(callback: (token: string) => void) {
    this.refreshSubscribers.push(callback)
  }

  // 通知所有订阅者Token已刷新
  private onRefreshed(token: string) {
    this.refreshSubscribers.forEach((callback) => callback(token))
    this.refreshSubscribers = []
  }

  // 处理认证失败 - 清除Token并跳转登录页
  private handleAuthFailure() {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    ElMessage.error('登录已过期，请重新登录')
    
    // 跳转到登录页，保存当前路径用于登录后跳回
    const currentPath = router.currentRoute.value.fullPath
    if (currentPath !== '/login') {
      router.push({
        path: '/login',
        query: { redirect: currentPath }
      })
    }
  }

  private async request<T>(url: string, options: any = {}): Promise<ApiResponse<T>> {
    try {
      const response = await this.axiosInstance.request({
        url,
        ...options,
      })
      return this.handleResponse(response.data)
    } catch (error: any) {
      throw error
    }
  }

  private handleResponse<T>(data: any): ApiResponse<T> {
    if (data.success === false) {
      ElMessage.error(data.message || '请求失败')
      throw new Error(data.message || '请求失败')
    }
    return data
  }

  private handleHttpError(status: number) {
    switch (status) {
      case 401:
        // 401已在拦截器中处理，这里不重复提示
        break
      case 403:
        ElMessage.error('权限不足')
        break
      case 404:
        ElMessage.error('资源不存在')
        break
      case 500:
        ElMessage.error('服务器内部错误')
        break
      default:
        ElMessage.error(`请求失败 (${status})`)
    }
  }

  private handleError(error: any) {
    if (error.code === 'ECONNABORTED') {
      ElMessage.error('请求超时')
    } else if (error.response) {
      // Axios响应错误
      const status = error.response.status
      // 401已在拦截器中处理
      if (status !== 401) {
        this.handleHttpError(status)
      }
    } else if (error.request) {
      // 网络错误
      ElMessage.error('网络连接失败')
    } else {
      // 其他错误
      ElMessage.error('请求发生错误')
    }
  }

  async get<T>(url: string, params?: Record<string, any>): Promise<ApiResponse<T>> {
    return this.request<T>(url, { method: 'GET', params })
  }

  async post<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(url, { method: 'POST', data })
  }

  async put<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(url, { method: 'PUT', data })
  }

  async delete<T>(url: string): Promise<ApiResponse<T>> {
    return this.request<T>(url, { method: 'DELETE' })
  }
}

// 创建API客户端实例
const apiClient = new ApiClient(API_BASE_URL, API_TIMEOUT)

// 导出API方法
export const api = {
  // HTTP方法
  get: <T>(url: string, params?: Record<string, any>) => apiClient.get<T>(url, params),
  post: <T>(url: string, data?: any) => apiClient.post<T>(url, data),
  put: <T>(url: string, data?: any) => apiClient.put<T>(url, data),
  delete: <T>(url: string) => apiClient.delete<T>(url),

  // API路由配置 - 通过API Gateway代理
  universities: {
    list: (params?: any) => apiClient.get('/api/v1/data/universities', params),
    get: (id: number) => apiClient.get(`/api/v1/data/universities/${id}`),
    search: (params: any) => apiClient.get('/api/v1/data/universities/search', params),
    statistics: () => apiClient.get('/api/v1/data/universities/statistics'),
  },

  // 专业相关API
  majors: {
    list: (params?: any) => apiClient.get('/api/v1/data/majors', params),
    get: (id: number) => apiClient.get(`/api/v1/data/majors/${id}`),
  },

  // 录取数据API
  admission: {
    list: (params?: any) => apiClient.get('/api/v1/data/admission', params),
  },

  // 健康检查
  health: () => apiClient.get('/api/v1/data/health'),
}

// 导出Token管理工具
export const tokenManager = {
  setToken(token: string, refreshToken?: string) {
    localStorage.setItem(TOKEN_KEY, token)
    if (refreshToken) {
      localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
    }
  },
  
  getToken(): string | null {
    return localStorage.getItem(TOKEN_KEY)
  },
  
  clearTokens() {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
  },
  
  isAuthenticated(): boolean {
    return !!localStorage.getItem(TOKEN_KEY)
  }
}

export default api
