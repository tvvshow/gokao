export interface User {
  id: string
  username: string
  email: string
  phone?: string
  avatar?: string
  membershipLevel: 'free' | 'basic' | 'premium'
  membershipExpiry?: string
  createdAt: string
  updatedAt: string
}

export interface LoginForm {
  username: string
  password: string
}

export interface RegisterForm {
  username: string
  email: string
  password: string
  phone?: string
}

export interface MembershipInfo {
  level: 'free' | 'basic' | 'premium'
  expiry?: string
  features: string[]
  usageCount: {
    recommendations: number
    searches: number
    analyses: number
  }
  limits: {
    recommendations: number
    searches: number
    analyses: number
  }
}