import { api } from './api-client';
import type {
  User,
  LoginForm,
  RegisterForm,
  MembershipInfo,
} from '@/types/user';
import {
  isWrappedResponse,
  normalizeMessageResponse,
  type WrappedResponse,
  unwrapDataOrSelf,
} from '@/utils/api-response';

type RawLoginResponse = {
  access_token?: string;
  refresh_token?: string;
  user?: Record<string, unknown>;
  message?: string;
};

function normalizeUser(raw: Record<string, unknown>): User {
  return {
    id: String(raw.id ?? ''),
    username: String(raw.username ?? ''),
    email: String(raw.email ?? ''),
    phone: (raw.phone as string | undefined) || undefined,
    avatar: (raw.avatar as string | undefined) || undefined,
    membershipLevel: String(
      raw.membershipLevel ?? raw.membership_level ?? 'free'
    ) as User['membershipLevel'],
    membershipExpiry:
      (raw.membershipExpiry as string | undefined) ||
      (raw.membership_expiry as string | undefined),
    createdAt: String(raw.createdAt ?? raw.created_at ?? ''),
    updatedAt: String(raw.updatedAt ?? raw.updated_at ?? ''),
  };
}

function normalizeMembershipInfo(raw: Record<string, unknown>): MembershipInfo {
  const level = String(
    raw.level ?? raw.membership_level ?? raw.membershipLevel ?? 'free'
  ) as MembershipInfo['level'];
  return {
    level,
    expiry:
      (raw.expiry as string | undefined) ||
      (raw.membership_expiry as string | undefined),
    features: Array.isArray(raw.features) ? (raw.features as string[]) : [],
    usageCount: {
      recommendations: Number(raw.recommendations ?? 0),
      searches: Number(raw.searches ?? 0),
      analyses: Number(raw.analyses ?? 0),
    },
    limits: {
      recommendations: Number(raw.max_devices ?? raw.recommendation_limit ?? 0),
      searches: Number(raw.search_limit ?? 0),
      analyses: Number(raw.analysis_limit ?? 0),
    },
  };
}

export const userApi = {
  async login(loginForm: LoginForm): Promise<{
    success: boolean;
    data: {
      token: string;
      refreshToken?: string;
      user: User;
    };
    message?: string;
  }> {
    const response = (await api.post(
      '/api/v1/users/auth/login',
      loginForm
    )) as unknown;

    if (isWrappedResponse(response)) {
      const wrapped = response as WrappedResponse<{
        token: string;
        refreshToken?: string;
        user: User;
      }>;
      return wrapped;
    }

    const raw = response as RawLoginResponse;
    return {
      success: true,
      data: {
        token: raw.access_token || '',
        refreshToken: raw.refresh_token,
        user: normalizeUser((raw.user || {}) as Record<string, unknown>),
      },
      message: raw.message,
    };
  },

  async register(registerForm: RegisterForm): Promise<{
    success: boolean;
    message?: string;
  }> {
    const response = (await api.post(
      '/api/v1/users/auth/register',
      registerForm
    )) as unknown;

    return normalizeMessageResponse(response, '注册成功');
  },

  async getProfile(): Promise<{
    success: boolean;
    data: User;
    message?: string;
  }> {
    const response = (await api.get('/api/v1/users/profile')) as unknown;

    const raw = unwrapDataOrSelf<Record<string, unknown> | User>(
      response as WrappedResponse<Record<string, unknown> | User> | User
    );
    return {
      success: true,
      data: normalizeUser(raw as Record<string, unknown>),
    };
  },

  async updateProfile(userData: Partial<User>): Promise<{
    success: boolean;
    data: User;
    message?: string;
  }> {
    const response = (await api.put(
      '/api/v1/users/profile',
      userData
    )) as unknown;

    const raw = unwrapDataOrSelf<Record<string, unknown> | User>(
      response as WrappedResponse<Record<string, unknown> | User> | User
    );
    return {
      success: true,
      data: normalizeUser(raw as Record<string, unknown>),
    };
  },

  async getMembershipInfo(): Promise<{
    success: boolean;
    data: MembershipInfo;
    message?: string;
  }> {
    const response = (await api.get('/api/v1/users/membership')) as unknown;

    const raw = unwrapDataOrSelf<Record<string, unknown> | MembershipInfo>(
      response as WrappedResponse<Record<string, unknown> | MembershipInfo> | MembershipInfo
    );
    return {
      success: true,
      data: normalizeMembershipInfo(raw as Record<string, unknown>),
    };
  },

  async logout(): Promise<{
    success: boolean;
    message?: string;
  }> {
    const refreshToken = localStorage.getItem('refresh_token');
    const response = (await api.post('/api/v1/users/auth/logout', {
      refresh_token: refreshToken || undefined,
    })) as unknown;

    return normalizeMessageResponse(response, '退出成功');
  },
};
