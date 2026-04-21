import type { University } from './university';

export interface StudentInfo {
  score: number | null;
  province: string;
  scienceType: '理科' | '文科' | '新高考';
  year: number;
  rank?: number | null;
  preferences: StudentPreferences;
}

export interface StudentPreferences {
  regions: string[];
  majorCategories: string[];
  universityTypes: string[];
  riskTolerance: 'conservative' | 'moderate' | 'aggressive';
  specialRequirements?: string;
}

// 后端返回的原始推荐类型（中文）
export type RecommendationTypeRaw = '冲刺' | '稳妥' | '保底';

// 前端使用的推荐类型（英文枚举）
export type RecommendationType = 'aggressive' | 'moderate' | 'conservative';

export interface RecommendationRaw {
  id: string;
  university: University;
  type: RecommendationTypeRaw; // 后端返回的中文类型
  admissionProbability: number; // 录取概率 0-100
  matchScore: number; // 匹配度 0-100
  recommendReason: string;
  riskLevel: 'low' | 'medium' | 'high';
  suggestedMajors: Array<{
    id: string;
    name: string;
    probability: number;
  }>;
  historicalData: {
    minScore: number;
    avgScore: number;
    maxScore: number;
    year: number;
  }[];
}

export interface Recommendation {
  id: string;
  university: University;
  type: RecommendationType; // 冲、稳、保
  admissionProbability: number; // 录取概率 0-100
  matchScore: number; // 匹配度 0-100
  recommendReason: string;
  riskLevel: 'low' | 'medium' | 'high';
  suggestedMajors: Array<{
    id: string;
    name: string;
    probability: number;
  }>;
  historicalData: {
    minScore: number;
    avgScore: number;
    maxScore: number;
    year: number;
  }[];
}

export interface RecommendationScheme {
  id?: string;
  name: string;
  studentInfo: StudentInfo;
  recommendations: Recommendation[];
  createdAt?: string;
  updatedAt?: string;
}
