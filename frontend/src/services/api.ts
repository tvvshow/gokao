import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'

// API配置
const API_CONFIG = {
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  timeout: 30000,
  withCredentials: true
}

// 创建axios实例
const apiClient: AxiosInstance = axios.create(API_CONFIG)

// 请求拦截器
apiClient.interceptors.request.use(
  (config: AxiosRequestConfig) => {
    // 添加认证token
    const token = localStorage.getItem('access_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // 添加请求ID
    const requestId = generateRequestId()
    if (config.headers) {
      config.headers['X-Request-ID'] = requestId
    }

    // 添加时间戳
    if (config.headers) {
      config.headers['X-Timestamp'] = Date.now().toString()
    }

    console.log(`[API Request] ${config.method?.toUpperCase()} ${config.url}`, {
      headers: config.headers,
      data: config.data,
      params: config.params
    })

    return config
  },
  (error) => {
    console.error('[API Request Error]', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    console.log(`[API Response] ${response.config.method?.toUpperCase()} ${response.config.url}`, {
      status: response.status,
      data: response.data,
      headers: response.headers
    })

    return response
  },
  (error) => {
    console.error('[API Response Error]', error)

    // 处理不同类型的错误
    if (error.response) {
      const { status, data } = error.response
      
      switch (status) {
        case 401:
          // 未授权，清除token并跳转到登录页
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
          window.location.href = '/login'
          ElMessage.error('登录已过期，请重新登录')
          break
        
        case 403:
          ElMessage.error('权限不足，无法访问该资源')
          break
        
        case 404:
          ElMessage.error('请求的资源不存在')
          break
        
        case 429:
          ElMessage.error('请求过于频繁，请稍后再试')
          break
        
        case 500:
          ElMessage.error('服务器内部错误，请稍后再试')
          break
        
        default:
          const message = data?.message || `请求失败 (${status})`
          ElMessage.error(message)
      }
    } else if (error.request) {
      // 网络错误
      ElMessage.error('网络连接失败，请检查网络设置')
    } else {
      // 其他错误
      ElMessage.error('请求配置错误')
    }

    return Promise.reject(error)
  }
)

// API接口定义
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
  timestamp: string
  request_id: string
}

export interface PaginationParams {
  page?: number
  page_size?: number
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export interface PaginationResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// 用户相关API
export interface LoginRequest {
  username: string
  password: string
  remember_me?: boolean
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: UserInfo
}

export interface UserInfo {
  id: string
  username: string
  email: string
  phone?: string
  role: string
  avatar?: string
  created_at: string
  updated_at: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  phone?: string
  verification_code?: string
}

// 院校相关API
export interface University {
  id: string
  name: string
  province: string
  city: string
  type: string
  level: string
  is_985: boolean
  is_211: boolean
  is_double_first_class: boolean
  description?: string
  website?: string
  phone?: string
  address?: string
  logo_url?: string
  created_at: string
  updated_at: string
}

export interface UniversitySearchParams extends PaginationParams {
  name?: string
  province?: string
  city?: string
  type?: string
  level?: string
  is_985?: boolean
  is_211?: boolean
  is_double_first_class?: boolean
}

// 专业相关API
export interface Major {
  id: string
  name: string
  code: string
  category: string
  description?: string
  employment_rate?: number
  average_salary?: number
  created_at: string
  updated_at: string
}

// 推荐相关API
export interface RecommendationRequest {
  user_id: string
  exam_scores: {
    total: number
    chinese: number
    math: number
    english: number
    comprehensive: number
  }
  exam_ranking?: number
  preferences: {
    provinces?: string[]
    cities?: string[]
    university_types?: string[]
    majors?: string[]
    min_score?: number
    max_score?: number
  }
  risk_tolerance: 'conservative' | 'balanced' | 'aggressive'
}

export interface RecommendationResponse {
  schemes: RecommendationScheme[]
  analysis: {
    score_analysis: string
    ranking_analysis: string
    recommendations: string[]
  }
  generated_at: string
}

export interface RecommendationScheme {
  type: 'conservative' | 'balanced' | 'aggressive'
  name: string
  description: string
  universities: RecommendedUniversity[]
  success_probability: number
  risk_level: string
}

export interface RecommendedUniversity {
  university: University
  major: Major
  admission_probability: number
  min_score: number
  avg_score: number
  ranking_requirement: number
  risk_warning?: string
}

// 支付相关API
export interface PaymentRequest {
  amount: number
  currency: string
  description: string
  payment_method: 'wechat_pay' | 'alipay' | 'alipay_qr'
  return_url?: string
  extra?: Record<string, any>
}

export interface PaymentResponse {
  order_id: string
  payment_method: string
  status: string
  amount: number
  currency: string
  payment_url: string
  expires_at: string
  created_at: string
  extra?: Record<string, any>
}

// API方法
class ApiService {
  // 用户认证
  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await apiClient.post<ApiResponse<LoginResponse>>('/users/login', data)
    return response.data.data
  }

  async register(data: RegisterRequest): Promise<UserInfo> {
    const response = await apiClient.post<ApiResponse<UserInfo>>('/users/register', data)
    return response.data.data
  }

  async logout(): Promise<void> {
    await apiClient.post('/users/logout')
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  async getCurrentUser(): Promise<UserInfo> {
    const response = await apiClient.get<ApiResponse<UserInfo>>('/users/me')
    return response.data.data
  }

  async refreshToken(): Promise<LoginResponse> {
    const refreshToken = localStorage.getItem('refresh_token')
    const response = await apiClient.post<ApiResponse<LoginResponse>>('/users/refresh', {
      refresh_token: refreshToken
    })
    return response.data.data
  }

  // 院校查询
  async searchUniversities(params: UniversitySearchParams): Promise<PaginationResponse<University>> {
    const response = await apiClient.get<ApiResponse<PaginationResponse<University>>>('/data/universities', {
      params
    })
    return response.data.data
  }

  async getUniversity(id: string): Promise<University> {
    const response = await apiClient.get<ApiResponse<University>>(`/data/universities/${id}`)
    return response.data.data
  }

  async getUniversityMajors(universityId: string): Promise<Major[]> {
    const response = await apiClient.get<ApiResponse<Major[]>>(`/data/universities/${universityId}/majors`)
    return response.data.data
  }

  // 专业查询
  async searchMajors(params: PaginationParams & { name?: string; category?: string }): Promise<PaginationResponse<Major>> {
    const response = await apiClient.get<ApiResponse<PaginationResponse<Major>>>('/data/majors', {
      params
    })
    return response.data.data
  }

  async getMajor(id: string): Promise<Major> {
    const response = await apiClient.get<ApiResponse<Major>>(`/data/majors/${id}`)
    return response.data.data
  }

  // 智能推荐
  async getRecommendations(data: RecommendationRequest): Promise<RecommendationResponse> {
    const response = await apiClient.post<ApiResponse<RecommendationResponse>>('/recommendations/generate', data)
    return response.data.data
  }

  async saveRecommendation(data: any): Promise<void> {
    await apiClient.post('/recommendations/save', data)
  }

  async getUserRecommendations(): Promise<RecommendationResponse[]> {
    const response = await apiClient.get<ApiResponse<RecommendationResponse[]>>('/recommendations/history')
    return response.data.data
  }

  // 支付相关
  async createPayment(data: PaymentRequest): Promise<PaymentResponse> {
    const response = await apiClient.post<ApiResponse<PaymentResponse>>('/payments/create', data)
    return response.data.data
  }

  async getPaymentStatus(orderId: string): Promise<PaymentResponse> {
    const response = await apiClient.get<ApiResponse<PaymentResponse>>(`/payments/${orderId}`)
    return response.data.data
  }

  // 健康检查
  async healthCheck(): Promise<any> {
    const response = await apiClient.get('/health')
    return response.data
  }
}

// 工具函数
function generateRequestId(): string {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
}

// 导出API服务实例
export const api = new ApiService()
export default api
