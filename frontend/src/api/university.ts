import { api } from './api-client';
import type {
  University,
  UniversitySearchParams,
  UniversitySearchResponse,
  UniversityDetail,
  AdmissionData,
  Major,
} from '@/types/university';
import {
  isWrappedResponse,
  type WrappedResponse,
  unwrapDataOrSelf,
} from '@/utils/api-response';
import { loadFromStorage, saveToStorage } from '@/utils/storage';

const FAVORITE_STORAGE_KEY = 'favorite_university_ids';

type RawUniversityListResponse = {
  universities: unknown[];
  total: number;
  page: number;
  page_size: number;
};

type RawAdmissionListResponse = {
  admission_data: unknown[];
};

function loadFavoriteIds(): string[] {
  return loadFromStorage<string[]>(FAVORITE_STORAGE_KEY, []);
}

function saveFavoriteIds(ids: string[]) {
  saveToStorage(FAVORITE_STORAGE_KEY, ids);
}

function normalizeMajor(raw: Record<string, unknown>): Major {
  return {
    id: String(raw.id ?? ''),
    name: String(raw.name ?? ''),
    code: String(raw.code ?? ''),
    category: String(raw.category ?? ''),
    degree:
      String(raw.degree ?? raw.degree_type ?? raw.degreeType ?? '') || '本科',
    duration: Number(raw.duration ?? 0),
    description: (raw.description as string | undefined) || undefined,
    employmentRate: Number(raw.employment_rate ?? raw.employmentRate ?? 0),
    averageSalary: Number(raw.average_salary ?? raw.averageSalary ?? 0),
    isPopular: Boolean(raw.is_popular ?? raw.isPopular),
  };
}

function normalizeAdmissionData(raw: Record<string, unknown>): AdmissionData {
  return {
    year: Number(raw.year ?? 0),
    province: String(raw.province ?? ''),
    batchType: String(raw.batchType ?? raw.batch ?? ''),
    scienceType: String(raw.scienceType ?? raw.category ?? ''),
    minScore: Number(raw.minScore ?? raw.min_score ?? 0),
    avgScore: Number(raw.avgScore ?? raw.avg_score ?? 0),
    maxScore: Number(raw.maxScore ?? raw.max_score ?? 0),
    minRank: Number(raw.minRank ?? raw.min_rank ?? 0),
    avgRank: Number(raw.avgRank ?? raw.avg_rank ?? 0),
    planCount: Number(raw.planCount ?? raw.planned_count ?? 0),
    admissionCount: Number(raw.admissionCount ?? raw.actual_count ?? 0),
  };
}

function normalizeUniversity(raw: Record<string, unknown>): UniversityDetail {
  const favoriteIds = loadFavoriteIds();
  const level = String(raw.level ?? '');
  const established = raw.established as string | undefined;
  const founded = established ? new Date(established).getFullYear() : undefined;
  const majors = Array.isArray(raw.majors)
    ? raw.majors.map((item) => normalizeMajor(item as Record<string, unknown>))
    : undefined;
  const admissionData = Array.isArray(raw.admission_data)
    ? raw.admission_data.map((item) =>
        normalizeAdmissionData(item as Record<string, unknown>)
      )
    : undefined;

  return {
    id: String(raw.id ?? ''),
    name: String(raw.name ?? ''),
    shortName: (raw.short_name as string | undefined) || (raw.alias as string),
    logo: (raw.logo as string | undefined) || undefined,
    province: String(raw.province ?? ''),
    city: String(raw.city ?? ''),
    type: String(raw.type ?? ''),
    level,
    rank: Number(raw.rank ?? raw.national_rank ?? 0) || undefined,
    founded,
    description: (raw.description as string | undefined) || undefined,
    studentCount:
      Number(raw.student_count ?? raw.studentCount ?? 0) || undefined,
    teacherCount:
      Number(raw.teacher_count ?? raw.teacherCount ?? 0) || undefined,
    majorCount:
      Number(raw.major_count ?? raw.majorCount ?? majors?.length ?? 0) ||
      undefined,
    campusArea: Number(raw.campus_area ?? raw.campusArea ?? 0) || undefined,
    employmentRate:
      Number(raw.employment_rate ?? raw.employmentRate ?? 0) || undefined,
    features: (raw.features as string[] | undefined) || undefined,
    strongMajors: (raw.strong_majors as string[] | undefined) || undefined,
    website: (raw.website as string | undefined) || undefined,
    phone: (raw.phone as string | undefined) || undefined,
    email: (raw.email as string | undefined) || undefined,
    address: (raw.address as string | undefined) || undefined,
    is985: level === '985' || level.includes('985'),
    is211: level === '211' || level.includes('211'),
    isDoubleFirstClass:
      level === 'double_first_class' || level.includes('双一流'),
    isFavorite: favoriteIds.includes(String(raw.id ?? '')),
    createdAt: String(raw.created_at ?? raw.createdAt ?? ''),
    updatedAt: String(raw.updated_at ?? raw.updatedAt ?? ''),
    admissionData,
    majors,
  };
}

function mapScienceType(scienceType?: string): string | undefined {
  if (!scienceType) {
    return undefined;
  }
  if (scienceType === '理科') {
    return 'science';
  }
  if (scienceType === '文科') {
    return 'liberal_arts';
  }
  if (scienceType === '新高考') {
    return 'comprehensive';
  }
  return scienceType;
}

function normalizeUniversityList(
  response: WrappedResponse<RawUniversityListResponse>
): {
  success: boolean;
  data: UniversitySearchResponse;
  message?: string;
} {
  return {
    success: response.success,
    message: response.message,
    data: {
      universities: response.data.universities.map((item) =>
        normalizeUniversity(item as Record<string, unknown>)
      ),
      total: response.data.total,
      page: response.data.page,
      pageSize: response.data.page_size,
    },
  };
}

export const universityApi = {
  async list(params?: {
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
    const response = (await api.get(
      '/api/v1/data/universities',
      params || {}
    )) as WrappedResponse<RawUniversityListResponse>;
    return normalizeUniversityList(response);
  },

  async search(params: UniversitySearchParams): Promise<{
    success: boolean;
    data: UniversitySearchResponse;
    message?: string;
  }> {
    const searchParams: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(params)) {
      if (key === 'keyword' && value) {
        searchParams.q = value;
      } else if (key === 'pageSize' && value) {
        searchParams.page_size = value;
      } else if (value !== undefined && value !== '') {
        searchParams[key] = value;
      }
    }

    const response = (await api.get(
      '/api/v1/data/universities/search',
      searchParams
    )) as WrappedResponse<RawUniversityListResponse>;
    return normalizeUniversityList(response);
  },

  async getDetail(id: string): Promise<{
    success: boolean;
    data: UniversityDetail;
    message?: string;
  }> {
    const response = (await api.get(`/api/v1/data/universities/${id}`)) as
      | WrappedResponse<Record<string, unknown>>
      | Record<string, unknown>;

    const raw = unwrapDataOrSelf<Record<string, unknown>>(response);
    return {
      success: true,
      data: normalizeUniversity(raw as Record<string, unknown>),
      message: isWrappedResponse<Record<string, unknown>>(response)
        ? response.message
        : undefined,
    };
  },

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

  async getPopular(limit: number = 10): Promise<{
    success: boolean;
    data: University[];
    message?: string;
  }> {
    const response = await universityApi.list({ page: 1, page_size: limit });
    return {
      success: response.success,
      message: response.message,
      data: response.data.universities.slice(0, limit),
    };
  },

  async toggleFavorite(universityId: string): Promise<{
    success: boolean;
    data: { isFavorite: boolean };
    message?: string;
  }> {
    const ids = loadFavoriteIds();
    const exists = ids.includes(universityId);
    const nextIds = exists
      ? ids.filter((id) => id !== universityId)
      : [...ids, universityId];
    saveFavoriteIds(nextIds);

    return {
      success: true,
      data: { isFavorite: !exists },
      message: !exists ? '收藏成功' : '已取消收藏',
    };
  },

  async getFavorites(): Promise<{
    success: boolean;
    data: University[];
    message?: string;
  }> {
    const ids = loadFavoriteIds();
    const universities = await Promise.all(
      ids.map(async (id) => {
        try {
          const response = await universityApi.getDetail(id);
          return response.data;
        } catch {
          return null;
        }
      })
    );

    return {
      success: true,
      data: universities.filter((item): item is University => item !== null),
    };
  },

  async getAdmissionData(
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
    const response = (await api.get('/api/v1/data/admission/data', {
      university_id: universityId,
      province: params?.province,
      category: mapScienceType(params?.scienceType),
      page_size: params?.years ? Math.max(params.years, 10) : 20,
    })) as WrappedResponse<RawAdmissionListResponse>;

    return {
      success: response.success,
      message: response.message,
      data: response.data.admission_data.map((item) =>
        normalizeAdmissionData(item as Record<string, unknown>)
      ),
    };
  },

  async analyzeAdmissionTrend(
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
    const response = await universityApi.getAdmissionData(universityId, {
      years,
    });
    const trend: Array<{
      year: number;
      minScore: number;
      avgScore: number;
      difficulty: 'easy' | 'medium' | 'hard';
    }> = response.data
      .sort((a, b) => a.year - b.year)
      .slice(-years)
      .map((item) => ({
        year: item.year,
        minScore: item.minScore,
        avgScore: item.avgScore,
        difficulty:
          item.minScore >= item.avgScore
            ? 'hard'
            : item.avgScore - item.minScore > 20
              ? 'medium'
              : 'easy',
      }));

    const lastScore = trend[trend.length - 1]?.avgScore ?? 0;
    const previousScore = trend[trend.length - 2]?.avgScore ?? lastScore;

    return {
      success: true,
      data: {
        trend,
        prediction: {
          nextYear: lastScore + (lastScore - previousScore),
          confidence: trend.length > 1 ? 0.7 : 0.4,
        },
      },
    };
  },

  async compare(universityIds: string[]): Promise<{
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
    const universities = await Promise.all(
      universityIds.map(async (id) => (await universityApi.getDetail(id)).data)
    );

    return {
      success: true,
      data: {
        universities,
        comparison: {
          scores: Object.fromEntries(
            universities.map((item) => [item.id, item.avgScoreScience || 0])
          ),
          ranks: Object.fromEntries(
            universities.map((item) => [item.id, item.rank || 0])
          ),
          features: Object.fromEntries(
            universities.map((item) => [item.id, item.features || []])
          ),
        },
      },
    };
  },
};
