import { api } from './api-client';
import type {
  University,
  UniversitySearchParams,
  UniversitySearchResponse,
  UniversityDetail,
  AdmissionData,
} from '@/types/university';

export const universityApi = {
  // 获取院校列表
  list(params?: {
    page?: number;
    page_size?: number;
    province?: string;
    type?: string;
    level?: string;
  }): Promise<{
    success: boolean;
    data: UniversitySearchResponse;
    message?: string;
  }> {
    return api.get('/api/v1/data/universities', params as Record<string, unknown> || {});
  },

  // 搜索院校
  search(params: UniversitySearchParams): Promise<{
    success: boolean;
    data: UniversitySearchResponse;
    message?: string;
  }> {
    // 后端使用 'q' 参数而不是 'keyword'，使用 'page_size' 而不是 'pageSize'
    const searchParams: Record<string, unknown> = {};
    // 复制所有参数，只重命名特定的
    for (const [key, value] of Object.entries(params)) {
      if (key === 'keyword' && value) {
        searchParams.q = value;
      } else if (key === 'pageSize' && value) {
        searchParams.page_size = value;
      } else if (value !== undefined && value !== '') {
        searchParams[key] = value;
      }
    }
    return api.get('/api/v1/data/universities/search', searchParams);
  },

  // 获取院校详情
  getDetail(id: string): Promise<{
    success: boolean;
    data: UniversityDetail;
    message?: string;
  }> {
    return api.get(`/api/v1/data/universities/${id}`);
  },

  // 获取院校统计信息
  getStatistics(): Promise<{
    success: boolean;
    data: {
      totalCount: number;
      provinceCount: number;
      typeCount: Record<string, number>;
      levelCount: Record<string, number>;
    };
    message?: string;
  }> {
    return api.get('/api/v1/data/universities/statistics');
  },

  // 获取热门院校
  getPopular(limit: number = 10): Promise<{
    success: boolean;
    data: University[];
    message?: string;
  }> {
    return api.get('/api/v1/data/universities/popular', { limit });
  },

  // 收藏/取消收藏院校
  toggleFavorite(universityId: string): Promise<{
    success: boolean;
    data: { isFavorite: boolean };
    message?: string;
  }> {
    return api.post(`/api/v1/data/universities/${universityId}/favorite`);
  },

  // 获取收藏的院校
  getFavorites(): Promise<{
    success: boolean;
    data: University[];
    message?: string;
  }> {
    return api.get('/api/v1/data/universities/favorites');
  },

  // 获取录取数据
  getAdmissionData(
    universityId: string,
    params?: {
      years?: number;
      province?: string;
      scienceType?: string;
    }
  ): Promise<{
    success: boolean;
    data: AdmissionData[];
    message?: string;
  }> {
    return api.get(`/api/v1/data/universities/${universityId}/admission`, params as Record<string, unknown>);
  },

  // 分析录取趋势
  analyzeAdmissionTrend(
    universityId: string,
    years: number = 5
  ): Promise<{
    success: boolean;
    data: {
      trend: Array<{
        year: number;
        minScore: number;
        avgScore: number;
        difficulty: 'easy' | 'medium' | 'hard';
      }>;
      prediction: {
        nextYear: number;
        confidence: number;
      };
    };
    message?: string;
  }> {
    return api.get(
      `/api/v1/data/universities/${universityId}/admission/analyze`,
      { years }
    );
  },

  // 对比院校
  compare(universityIds: string[]): Promise<{
    success: boolean;
    data: {
      universities: UniversityDetail[];
      comparison: {
        scores: Record<string, number>;
        ranks: Record<string, number>;
        features: Record<string, string[]>;
      };
    };
    message?: string;
  }> {
    return api.post('/api/v1/data/universities/compare', { universityIds });
  },
};
