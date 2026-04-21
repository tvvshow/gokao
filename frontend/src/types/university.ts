export interface University {
  id: string;
  name: string;
  shortName?: string;
  logo?: string;
  province: string;
  city: string;
  type: string; // 院校类型：综合类、理工类、师范类等
  level: string; // 办学层次：985工程、211工程、双一流、普通本科
  rank?: number; // 全国排名
  founded?: number; // 建校年份
  description?: string; // 学校简介

  // 标签属性
  is985?: boolean;
  is211?: boolean;
  isDoubleFirstClass?: boolean;

  // 录取分数线
  minScoreScience?: number; // 理科最低分
  minScoreLiberalArts?: number; // 文科最低分
  avgScoreScience?: number; // 理科平均分
  avgScoreLiberalArts?: number; // 文科平均分

  // 基本信息
  studentCount?: number; // 在校生人数
  teacherCount?: number; // 教师人数
  majorCount?: number; // 专业数量
  campusArea?: number; // 校园面积
  employmentRate?: number; // 就业率

  // 特色标签
  features?: string[]; // 特色标签
  strongMajors?: string[]; // 优势专业

  // 联系信息
  website?: string;
  phone?: string;
  email?: string;
  address?: string;

  // 状态
  isFavorite?: boolean; // 是否收藏

  createdAt: string;
  updatedAt: string;
}

export interface UniversitySearchParams {
  name?: string;
  keyword?: string;
  province?: string;
  type?: string;
  level?: string;
  scoreRange?: [number, number];
  rankRange?: [number, number];
  scale?: string;
  page?: number;
  pageSize?: number;
  page_size?: number;
  limit?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

// Backend enum value mappings
export const UNIVERSITY_LEVEL_MAP: Record<string, string> = {
  '985': '985',
  '211': '211',
  '双一流': 'double_first_class',
  '本科': 'ordinary',
};

export const UNIVERSITY_TYPE_MAP: Record<string, string> = {
  '综合类': 'comprehensive',
  '理工类': 'science',
  '师范类': 'teacher',
  '财经类': 'finance',
  '医药类': 'medical',
  '艺术类': 'art',
  '农林类': 'agriculture',
  '政法类': 'law',
  '语言类': 'language',
  '体育类': 'sports',
};

export interface UniversitySearchResponse {
  universities: University[];
  total: number;
  page: number;
  pageSize: number;
}

export interface AdmissionData {
  year: number;
  province: string;
  batchType: string; // 批次类型：本科一批、本科二批等
  scienceType: string; // 科类：理科、文科
  minScore: number;
  avgScore: number;
  maxScore: number;
  minRank: number;
  avgRank: number;
  planCount: number; // 招生计划数
  admissionCount: number; // 实际录取数
}

export interface UniversityDetail extends University {
  description?: string; // 学校简介
  history?: string; // 学校历史
  campus?: string; // 校园介绍
  admissionData?: AdmissionData[]; // 历年录取数据
  majors?: Major[]; // 开设专业
}

export interface Major {
  id: string;
  name: string;
  code: string;
  category: string; // 学科门类
  degree: string; // 学位类型：学士、硕士、博士
  duration: number; // 学制年限
  description?: string;
  employmentRate?: number; // 就业率
  averageSalary?: number; // 平均薪资
  isPopular?: boolean; // 是否热门专业
}
