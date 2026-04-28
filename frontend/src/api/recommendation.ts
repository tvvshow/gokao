import { api } from './api-client';
import type {
  StudentInfo,
  Recommendation,
  RecommendationRaw,
  RecommendationScheme,
} from '@/types/recommendation';
import { loadFromStorage, saveToStorage } from '@/utils/storage';

const SCHEME_STORAGE_KEY = 'recommendation_schemes';

type RecommendationSchemeRecord = RecommendationScheme & { id: string };

function loadSchemes(): RecommendationSchemeRecord[] {
  return loadFromStorage<RecommendationSchemeRecord[]>(SCHEME_STORAGE_KEY, []);
}

function saveSchemes(schemes: RecommendationSchemeRecord[]) {
  saveToStorage(SCHEME_STORAGE_KEY, schemes);
}

function createSchemeId() {
  return `scheme_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
}

export const recommendationApi = {
  generateRecommendations(studentInfo: StudentInfo): Promise<{
    success: boolean;
    data: {
      recommendations: RecommendationRaw[];
      analysisReport: string;
    };
    message?: string;
  }> {
    return api.post('/api/v1/recommendations/generate', studentInfo);
  },

  getRecommendTypes(): Promise<{
    success: boolean;
    data: Record<string, string>;
    message?: string;
  }> {
    return api.get('/api/v1/data/algorithm/recommend-types');
  },

  getRiskToleranceOptions(): Promise<{
    success: boolean;
    data: Record<string, string>;
    message?: string;
  }> {
    return api.get('/api/v1/data/algorithm/risk-tolerance');
  },

  async saveScheme(scheme: RecommendationScheme): Promise<{
    success: boolean;
    data: { id: string };
    message?: string;
  }> {
    const schemes = loadSchemes();
    const id = scheme.id || createSchemeId();
    const now = new Date().toISOString();
    const nextScheme: RecommendationSchemeRecord = {
      ...scheme,
      id,
      createdAt: scheme.createdAt || now,
      updatedAt: now,
    };

    const nextSchemes = [
      ...schemes.filter((item) => item.id !== id),
      nextScheme,
    ];
    saveSchemes(nextSchemes);

    return {
      success: true,
      data: { id },
      message: '方案保存成功',
    };
  },

  async getSchemes(): Promise<{
    success: boolean;
    data: RecommendationScheme[];
    message?: string;
  }> {
    return {
      success: true,
      data: loadSchemes(),
    };
  },

  async exportReport(recommendations: Recommendation[]): Promise<Blob> {
    const lines = recommendations.map((item, index) =>
      [
        `#${index + 1} ${item.university.name}`,
        `类型: ${item.type}`,
        `录取概率: ${item.admissionProbability}%`,
        `匹配度: ${item.matchScore}%`,
        `推荐理由: ${item.recommendReason}`,
      ].join('\n')
    );

    return new Blob([lines.join('\n\n')], {
      type: 'text/plain;charset=utf-8',
    });
  },
};
