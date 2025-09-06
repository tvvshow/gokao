import { ElMessage } from 'element-plus'
<<<<<<< HEAD
import axios, { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
=======
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc

// API基础配置 - 连接到API Gateway
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const API_TIMEOUT = 10000

// 统一API响应格式
interface ApiResponse<T = any> {
  success: boolean
  data: T
  message: string
  total?: number
}

// API客户端类
class ApiClient {
<<<<<<< HEAD
  private axiosInstance: AxiosInstance
  private baseURL: string

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
        const token = localStorage.getItem('auth_token')
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => {
        return Promise.reject(error)
      }
    )

    // 响应拦截器 - 统一错误处理
    this.axiosInstance.interceptors.response.use(
      (response: AxiosResponse<ApiResponse>) => {
        return this.handleResponse(response.data)
      },
      (error) => {
        this.handleError(error)
        return Promise.reject(error)
      }
    )
  }

  private async request<T>(url: string, options: any = {}): Promise<ApiResponse<T>> {
    try {
      const response = await this.axiosInstance.request({
        url,
        ...options,
      })
      return response.data
    } catch (error: any) {
=======
  private baseURL: string
  private timeout: number

  constructor(baseURL: string, timeout: number = 10000) {
    this.baseURL = baseURL
    this.timeout = timeout
  }

  private async request<T>(url: string, options: RequestInit = {}): Promise<ApiResponse<T>> {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), this.timeout)

    try {
      const response = await fetch(`${this.baseURL}${url}`, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
        signal: controller.signal,
      })

      clearTimeout(timeoutId)

      if (!response.ok) {
        this.handleHttpError(response.status)
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return this.handleResponse(data)
    } catch (error: any) {
      clearTimeout(timeoutId)
      this.handleError(error)
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
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
        ElMessage.error('登录已过期，请重新登录')
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
<<<<<<< HEAD
    if (error.code === 'ECONNABORTED') {
      ElMessage.error('请求超时')
    } else if (error.response) {
      // Axios响应错误
      const status = error.response.status
      this.handleHttpError(status)
    } else if (error.request) {
      // 网络错误
      ElMessage.error('网络连接失败')
    } else {
      // 其他错误
      ElMessage.error('请求发生错误')
=======
    if (error.name === 'AbortError') {
      ElMessage.error('请求超时')
    } else if (!error.message.includes('HTTP error')) {
      ElMessage.error('网络连接失败')
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
    }
  }

  async get<T>(url: string, params?: Record<string, any>): Promise<ApiResponse<T>> {
<<<<<<< HEAD
    return this.request<T>(url, { method: 'GET', params })
  }

  async post<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(url, { method: 'POST', data })
  }

  async put<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(url, { method: 'PUT', data })
=======
    const searchParams = params ? new URLSearchParams(params).toString() : ''
    const fullUrl = searchParams ? `${url}?${searchParams}` : url
    return this.request<T>(fullUrl, { method: 'GET' })
  }

  async post<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(url, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  async put<T>(url: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(url, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    })
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
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

export default api
