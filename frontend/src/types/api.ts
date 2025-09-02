// 高校信息接口
export interface University {
  id: number
  code: string
  name: string
  province: string
  city: string
  type: string
  level: string
  national_rank: number
  is_active: boolean
  website?: string
  description?: string
}

// 专业信息接口
export interface Major {
  id: number
  university_id: number
  code: string
  name: string
  category: string
  is_active: boolean
}

// 录取数据接口
export interface AdmissionData {
  id: number
  university_id: number
  major_id: number
  year: number
  province: string
  min_score: number
  avg_score: number
  max_score: number
  min_rank: number
  avg_rank: number
  max_rank: number
}

// 搜索参数接口
export interface UniversitySearchParams {
  name?: string
  province?: string
  type?: string
  level?: string
  page?: number
  limit?: number
}

// 统计信息接口
export interface UniversityStatistics {
  total: number
  985_count: number
  211_count: number
  provinces: string[]
  types: string[]
}

// API响应格式
export interface ApiResponse<T = any> {
  success: boolean
  data: T
  message: string
  total?: number
}
