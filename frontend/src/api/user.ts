import { api } from './api-client';
import type {
  User,
  LoginForm,
  RegisterForm,
  MembershipInfo,
} from '@/types/user';

export const userApi = {
  // 用户登录
  login(loginForm: LoginForm): Promise<{
    success: boolean;
    data: {
      token: string;
      refreshToken?: string;
      user: User;
    };
    message?: string;
  }> {
    return api.post('/api/user/v1/users/auth/login', loginForm);
  },

  // 用户注册
  register(registerForm: RegisterForm): Promise<{
    success: boolean;
    message?: string;
  }> {
    return api.post('/api/user/v1/users/auth/register', registerForm);
  },

  // 获取用户信息
  getProfile(): Promise<{
    success: boolean;
    data: User;
    message?: string;
  }> {
    return api.get('/api/user/v1/users/profile');
  },

  // 更新用户信息
  updateProfile(userData: Partial<User>): Promise<{
    success: boolean;
    data: User;
    message?: string;
  }> {
    return api.put('/api/user/v1/users/profile', userData);
  },

  // 获取会员信息
  getMembershipInfo(): Promise<{
    success: boolean;
    data: MembershipInfo;
    message?: string;
  }> {
    return api.get('/api/user/v1/users/membership');
  },

  // 退出登录
  logout(): Promise<{
    success: boolean;
    message?: string;
  }> {
    return api.post('/api/user/v1/users/auth/logout');
  },
};
