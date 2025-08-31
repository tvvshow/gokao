import api from './index'
import type { 
  User, 
  LoginForm, 
  RegisterForm,
  MembershipInfo
} from '@/types/user'

export const userApi = {
  // 用户登录
  login(loginForm: LoginForm): Promise<{
    success: boolean
    data: {
      token: string
      user: User
    }
    message?: string
  }> {
    return api.post('/v1/auth/login', loginForm)
  },

  // 用户注册
  register(registerForm: RegisterForm): Promise<{
    success: boolean
    message?: string
  }> {
    return api.post('/v1/auth/register', registerForm)
  },

  // 获取用户信息
  getProfile(): Promise<{
    success: boolean
    data: User
    message?: string
  }> {
    return api.get('/v1/user/profile')
  },

  // 更新用户信息
  updateProfile(userData: Partial<User>): Promise<{
    success: boolean
    data: User
    message?: string
  }> {
    return api.put('/v1/user/profile', userData)
  },

  // 获取会员信息
  getMembershipInfo(): Promise<{
    success: boolean
    data: MembershipInfo
    message?: string
  }> {
    return api.get('/v1/user/membership')
  },

  // 退出登录
  logout(): Promise<{
    success: boolean
    message?: string
  }> {
    return api.post('/v1/auth/logout')
  }
}