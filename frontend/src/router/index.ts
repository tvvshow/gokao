import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/HomePageModern.vue'),
    meta: { title: '首页' }
  },
  {
    path: '/universities',
    name: 'Universities',
    component: () => import('@/views/UniversitiesPageModern.vue'),
    meta: { title: '院校查询' }
  },
  {
    path: '/majors',
    name: 'Majors',
    component: () => import('@/views/MajorsPage.vue'),
    meta: { title: '专业分析' }
  },
  {
    path: '/recommendation',
    name: 'Recommendation',
    component: () => import('@/views/RecommendationPage.vue'),
    meta: { title: '智能推荐' }
  },
  {
    path: '/analysis',
    name: 'Analysis',
    component: () => import('@/views/AnalysisPage.vue'),
    meta: { title: '数据分析' }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/ProfilePage.vue'),
    meta: { title: '个人中心', requiresAuth: true }
  },
  {
    path: '/membership',
    name: 'Membership',
    component: () => import('@/views/MembershipPage.vue'),
    meta: { title: '会员服务' }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/LoginPage.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/RegisterPage.vue'),
    meta: { title: '注册' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0 }
    }
  }
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  document.title = `${to.meta.title} - 高考志愿填报助手`
  
  // 检查是否需要登录
  if (to.meta.requiresAuth) {
    const token = localStorage.getItem('auth_token')
    if (!token) {
      next({
        path: '/login',
        query: { redirect: to.fullPath }
      })
      return
    }
  }
  
  next()
})

export default router