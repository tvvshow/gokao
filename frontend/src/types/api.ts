/**
 * API types - consolidated from api.ts and api-params.ts
 */

import type { Recommendation, StudentInfo } from './recommendation';

// Re-export domain entity types from specialized type files
export type {
  University,
  UniversitySearchParams,
  UniversityDetail,
  Major,
  AdmissionData,
} from './university';

// ============ API Infrastructure Types ============

// 统计信息接口
export interface UniversityStatistics {
  total: number;
  count985: number;
  count211: number;
  provinces: string[];
  types: string[];
}

// ============ Request Parameter Types ============

// Generic pagination parameters
export interface PaginationParams {
  page?: number;
  limit?: number;
  pageSize?: number;
  page_size?: number;
}

// University list parameters
export interface UniversityListParams extends PaginationParams {
  name?: string;
  province?: string;
  type?: string;
  level?: string;
  minScore?: number;
  maxScore?: number;
}

// Major list parameters
export interface MajorListParams extends PaginationParams {
  name?: string;
  category?: string;
  universityId?: number;
}

// Admission data parameters
export interface AdmissionListParams extends PaginationParams {
  universityId?: number;
  majorId?: number;
  year?: number;
  province?: string;
}

// Recommendation save data
export interface RecommendationSaveData {
  name: string;
  studentInfo: StudentInfo;
  recommendations: Recommendation[];
}

// User login data
export interface LoginData {
  username: string;
  password: string;
}

// User register data
export interface RegisterData {
  username: string;
  email: string;
  password: string;
  phone?: string;
}

// ============ API Infrastructure Types ============

// Axios request options
export interface RequestOptions {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  params?: Record<string, unknown>;
  data?: unknown;
  headers?: Record<string, string>;
}

// API error response
export interface ApiError {
  code: string;
  message: string;
  status?: number;
  details?: Record<string, unknown>;
}

// Generic API response (unified from api-client.ts and services/api.ts)
export interface ApiResponse<T = unknown> {
  success: boolean;
  code?: number;
  data: T;
  message?: string;
  total?: number;
  timestamp?: string;
  request_id?: string;
}

// ============ UI Component Types ============

// Membership plan type - re-exported from payment.ts
export type { MembershipPlan, MembershipPlanItem } from './payment';

// Compare table row type
export interface CompareTableRow {
  field: string;
  [universityId: string]: string | number | undefined;
}

// Tab change event type
export interface TabChangeEvent {
  paneName: string;
  props?: {
    name?: string;
    label?: string;
  };
}

// Form rule type (Element Plus compatible)
export interface FormRule {
  required?: boolean;
  message?: string;
  trigger?: string | string[];
  type?: string;
  min?: number;
  max?: number;
  validator?: (
    rule: FormRule,
    value: string,
    callback: (error?: Error) => void
  ) => void;
}

// Column configuration for responsive tables
export interface ColumnConfig {
  prop: string;
  label: string;
  width?: number | string;
  minWidth?: number | string;
  fixed?: boolean | 'left' | 'right';
  sortable?: boolean;
  formatter?: (row: unknown, column: unknown, cellValue: unknown) => string;
}

// Home page statistics data
export interface HomeStatistics {
  universityCount: number;
  majorCount: number;
  userCount: number;
  accuracyRate: number;
}

// Statistics API response
export interface StatisticsResponse {
  success: boolean;
  data: HomeStatistics;
  message: string;
}
