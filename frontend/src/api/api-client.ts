import { ElMessage } from 'element-plus';
import axios from 'axios';
import type {
  AxiosInstance,
  AxiosResponse,
  InternalAxiosRequestConfig,
  AxiosError,
} from 'axios';
import router from '@/router';
import type {
  ApiResponse,
  RequestOptions,
  UniversityListParams,
  UniversitySearchParams,
  MajorListParams,
  AdmissionListParams,
} from '@/types/api';
import { isWrappedResponse, type WrappedResponse } from '@/utils/api-response';

// API基础配置 - 生产环境使用相对路径，开发环境使用localhost
// 使用 ?? 空值合并运算符，只有undefined/null时才使用默认值，空字符串不会被替换
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '';
const API_TIMEOUT = 10000;

// Token存储键名
const TOKEN_KEY = 'auth_token';
const REFRESH_TOKEN_KEY = 'refresh_token';

// Token刷新响应格式
type RefreshTokenResponse = WrappedResponse<{
  token: string;
  refreshToken?: string;
}>;

interface RawRefreshTokenResponse {
  access_token?: string;
  refresh_token?: string;
}

// API客户端类
class ApiClient {
  private axiosInstance: AxiosInstance;
  private baseURL: string;
  private isRefreshing: boolean = false;
  private refreshSubscribers: Array<(token: string) => void> = [];

  constructor(baseURL: string, timeout: number = 10000) {
    this.baseURL = baseURL;
    this.axiosInstance = axios.create({
      baseURL,
      timeout,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // 请求拦截器 - 添加Bearer Token
    this.axiosInstance.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        const token = localStorage.getItem(TOKEN_KEY);
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // 响应拦截器 - 统一错误处理和Token刷新
    this.axiosInstance.interceptors.response.use(
      (response: AxiosResponse<ApiResponse>) => {
        return response;
      },
      async (error) => {
        const originalRequest = error.config;

        // 处理401错误 - Token过期
        if (error.response?.status === 401 && !originalRequest._retry) {
          // 如果是刷新Token的请求失败，直接跳转登录
          if (originalRequest.url?.includes('/auth/refresh')) {
            this.handleAuthFailure();
            return Promise.reject(error);
          }

          // 尝试刷新Token
          if (!this.isRefreshing) {
            this.isRefreshing = true;
            originalRequest._retry = true;

            try {
              const newToken = await this.refreshToken();
              if (newToken) {
                // 通知所有等待的请求
                this.onRefreshed(newToken);
                // 重试原始请求
                originalRequest.headers.Authorization = `Bearer ${newToken}`;
                return this.axiosInstance(originalRequest);
              }
              this.handleAuthFailure();
              return Promise.reject(error);
            } catch (refreshError) {
              this.handleAuthFailure();
              return Promise.reject(refreshError);
            } finally {
              this.isRefreshing = false;
            }
          } else {
            // 等待Token刷新完成
            return new Promise((resolve) => {
              this.subscribeTokenRefresh((token: string) => {
                originalRequest.headers.Authorization = `Bearer ${token}`;
                resolve(this.axiosInstance(originalRequest));
              });
            });
          }
        }

        this.handleError(error);
        return Promise.reject(error);
      }
    );
  }

  // 刷新Token
  private async refreshToken(): Promise<string | null> {
    const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY);
    if (!refreshToken) {
      return null;
    }

    try {
      const response = await axios.post<
        RefreshTokenResponse | RawRefreshTokenResponse
      >(
        `${this.baseURL}/api/v1/users/auth/refresh`,
        { refresh_token: refreshToken },
        { headers: { 'Content-Type': 'application/json' } }
      );

      const responseData = response.data;
      const wrappedToken = isWrappedResponse<{
        token: string;
        refreshToken?: string;
      }>(responseData)
        ? responseData.data?.token
        : undefined;
      const wrappedRefreshToken = isWrappedResponse<{
        token: string;
        refreshToken?: string;
      }>(responseData)
        ? responseData.data?.refreshToken
        : undefined;
      const rawToken =
        'access_token' in responseData ? responseData.access_token : undefined;
      const rawRefreshToken =
        'refresh_token' in responseData
          ? responseData.refresh_token
          : undefined;
      const newToken = wrappedToken || rawToken;

      if (newToken) {
        localStorage.setItem(TOKEN_KEY, newToken);

        const newRefreshToken = wrappedRefreshToken || rawRefreshToken;
        if (newRefreshToken) {
          localStorage.setItem(REFRESH_TOKEN_KEY, newRefreshToken);
        }

        return newToken;
      }
      return null;
    } catch {
      return null;
    }
  }

  // 订阅Token刷新
  private subscribeTokenRefresh(callback: (token: string) => void) {
    this.refreshSubscribers.push(callback);
  }

  // 通知所有订阅者Token已刷新
  private onRefreshed(token: string) {
    this.refreshSubscribers.forEach((callback) => callback(token));
    this.refreshSubscribers = [];
  }

  // 处理认证失败 - 清除Token并跳转登录页
  private handleAuthFailure() {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    ElMessage.error('登录已过期，请重新登录');

    // 跳转到登录页，保存当前路径用于登录后跳回
    const currentPath = router.currentRoute.value.fullPath;
    if (currentPath !== '/login') {
      router.push({
        path: '/login',
        query: { redirect: currentPath },
      });
    }
  }

  private async request<T>(
    url: string,
    options: Partial<RequestOptions> = {}
  ): Promise<ApiResponse<T>> {
    const response = await this.axiosInstance.request({
      url,
      ...options,
    });
    return this.handleResponse(response.data);
  }

  private handleResponse<T>(data: ApiResponse<T>): ApiResponse<T> {
    if (data.success === false) {
      ElMessage.error(data.message || '请求失败');
      throw new Error(data.message || '请求失败');
    }
    return data;
  }

  private handleHttpError(status: number) {
    switch (status) {
      case 401:
        // 401已在拦截器中处理，这里不重复提示
        break;
      case 403:
        ElMessage.error('权限不足');
        break;
      case 404:
        ElMessage.error('资源不存在');
        break;
      case 500:
        ElMessage.error('服务器内部错误');
        break;
      default:
        ElMessage.error(`请求失败 (${status})`);
    }
  }

  private handleError(error: AxiosError) {
    if (error.code === 'ECONNABORTED') {
      ElMessage.error('请求超时');
    } else if (error.response) {
      // Axios response error
      const status = error.response.status;
      // 401 is handled in interceptor
      if (status !== 401) {
        this.handleHttpError(status);
      }
    } else if (error.request) {
      // Network error
      ElMessage.error('网络连接失败');
    } else {
      // Other errors
      ElMessage.error('请求发生错误');
    }
  }

  async get<T>(
    url: string,
    params?: Record<string, unknown>
  ): Promise<ApiResponse<T>> {
    return this.request<T>(url, { method: 'GET', params });
  }

  async post<T>(url: string, data?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>(url, { method: 'POST', data });
  }

  async put<T>(url: string, data?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>(url, { method: 'PUT', data });
  }

  async delete<T>(url: string): Promise<ApiResponse<T>> {
    return this.request<T>(url, { method: 'DELETE' });
  }

  // Download method for blob responses (PDFs, etc.)
  async download<T = Blob>(url: string, data?: unknown): Promise<T> {
    const response = await this.axiosInstance.request({
      url,
      method: 'POST',
      data,
      responseType: 'blob',
    });
    return response.data as T;
  }
}

// 创建API客户端实例
const apiClient = new ApiClient(API_BASE_URL, API_TIMEOUT);

// Export API methods
export const api = {
  // HTTP methods
  get: <T>(url: string, params?: Record<string, unknown>) =>
    apiClient.get<T>(url, params),
  post: <T>(url: string, data?: unknown) => apiClient.post<T>(url, data),
  put: <T>(url: string, data?: unknown) => apiClient.put<T>(url, data),
  delete: <T>(url: string) => apiClient.delete<T>(url),
  download: <T = Blob>(url: string, data?: unknown) =>
    apiClient.download<T>(url, data),

  // API routes - 统一路径：/api/v1/{service}/...
  universities: {
    list: (params?: UniversityListParams) =>
      apiClient.get(
        '/api/v1/data/universities',
        params as Record<string, unknown>
      ),
    get: (id: number) => apiClient.get(`/api/v1/data/universities/${id}`),
    search: (params: UniversitySearchParams) =>
      apiClient.get(
        '/api/v1/data/universities/search',
        params as Record<string, unknown>
      ),
    statistics: () => apiClient.get('/api/v1/data/universities/statistics'),
  },

  // Major related API
  majors: {
    list: (params?: MajorListParams) =>
      apiClient.get('/api/v1/data/majors', params as Record<string, unknown>),
    get: (id: number) => apiClient.get(`/api/v1/data/majors/${id}`),
  },

  // Admission data API
  admission: {
    list: (params?: AdmissionListParams) =>
      apiClient.get(
        '/api/v1/data/admission/data',
        params as Record<string, unknown>
      ),
  },

  // Health check
  health: () => apiClient.get('/api/v1/data/health'),
};

// 导出Token管理工具
export const tokenManager = {
  setToken(token: string, refreshToken?: string) {
    localStorage.setItem(TOKEN_KEY, token);
    if (refreshToken) {
      localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
    }
  },

  getToken(): string | null {
    return localStorage.getItem(TOKEN_KEY);
  },

  clearTokens() {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
  },

  isAuthenticated(): boolean {
    return !!localStorage.getItem(TOKEN_KEY);
  },
};

export default api;
