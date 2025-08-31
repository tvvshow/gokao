import api from './index'
import type { 
  StudentInfo, 
  Recommendation,
  RecommendationScheme
} from '@/types/recommendation'

export const recommendationApi = {
  // 生成智能推荐
  generateRecommendations(studentInfo: StudentInfo): Promise<{
    success: boolean
    data: {
      recommendations: Recommendation[]
      analysisReport: string
    }
    message?: string
  }> {
    return api.post('/v1/recommendation/generate', studentInfo)
  },

  // 获取推荐类型
  getRecommendTypes(): Promise<{
    success: boolean
    data: string[]
    message?: string
  }> {
    return api.get('/v1/recommendation/types')
  },

  // 获取风险承受度选项
  getRiskToleranceOptions(): Promise<{
    success: boolean
    data: Array<{
      value: string
      label: string
      description: string
    }>
    message?: string
  }> {
    return api.get('/v1/recommendation/risk-tolerance')
  },

  // 保存推荐方案
  saveScheme(scheme: RecommendationScheme): Promise<{
    success: boolean
    data: { id: string }
    message?: string
  }> {
    return api.post('/v1/recommendation/schemes', scheme)
  },

  // 获取保存的方案
  getSchemes(): Promise<{
    success: boolean
    data: RecommendationScheme[]
    message?: string
  }> {
    return api.get('/v1/recommendation/schemes')
  },

  // 导出推荐报告
  exportReport(recommendations: Recommendation[]): Promise<{
    success: boolean
    data: Blob
    message?: string
  }> {
    return api.post('/v1/recommendation/export', 
      { recommendations }, 
      { responseType: 'blob' }
    )
  }
}