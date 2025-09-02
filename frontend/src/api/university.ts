import { api } from './api-client'
import type { 
  University, 
  UniversitySearchParams, 
  UniversitySearchResponse,
  UniversityDetail,
  AdmissionData
} from '@/types/university'

export const universityApi = {
  // 搜索院校
  search(params: UniversitySearchParams): Promise<{
    success: boolean
    data: UniversitySearchResponse
    message?: string
  }> {
    return api.get('/api/v1/data/universities', { params })
  },

  // 获取院校详情
  getDetail(id: string): Promise<{
    success: boolean
    data: UniversityDetail
    message?: string
  }> {
    return api.get(`/api/v1/data/universities/${id}`)
  },

  // 获取院校统计信息
  getStatistics(): Promise<{
    success: boolean
    data: {
      totalCount: number
      provinceCount: number
      typeCount: Record<string, number>
      levelCount: Record<string, number>
    }
    message?: string
  }> {
    return api.get('/api/v1/data/universities/statistics')
  },

  // 获取热门院校
  getPopular(limit: number = 10): Promise<{
    success: boolean
    data: University[]
    message?: string
  }> {
    return api.get('/api/v1/data/universities/popular', { 
      params: { limit } 
    })
  },

  // 收藏/取消收藏院校
  toggleFavorite(universityId: string): Promise<{
    success: boolean
    data: { isFavorite: boolean }
    message?: string
  }> {
    return api.post(`/api/v1/data/universities/${universityId}/favorite`)
  },

  // 获取收藏的院校
  getFavorites(): Promise<{
    success: boolean
    data: University[]
    message?: string
  }> {
    return api.get('/api/v1/data/universities/favorites')
  },

  // 获取录取数据
  getAdmissionData(universityId: string, params?: {
    years?: number
    province?: string
    scienceType?: string
  }): Promise<{
    success: boolean
    data: AdmissionData[]
    message?: string
  }> {
    return api.get(`/api/v1/data/universities/${universityId}/admission`, { params })
  },

  // 分析录取趋势
  analyzeAdmissionTrend(universityId: string, years: number = 5): Promise<{
    success: boolean
    data: {
      trend: Array<{
        year: number
        minScore: number
        avgScore: number
        difficulty: 'easy' | 'medium' | 'hard'
      }>
      prediction: {
        nextYear: number
        confidence: number
      }
    }
    message?: string
  }> {
    return api.get(`/api/v1/data/universities/${universityId}/admission/analyze`, {
      params: { years }
    })
  },

  // 对比院校
  compare(universityIds: string[]): Promise<{
    success: boolean
    data: {
      universities: UniversityDetail[]
      comparison: {
        scores: Record<string, number>
        ranks: Record<string, number>
        features: Record<string, string[]>
      }
    }
    message?: string
  }> {
    return api.post('/api/v1/data/universities/compare', { universityIds })
  }
}