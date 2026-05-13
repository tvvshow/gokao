import { createRouter, createWebHistory } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';
import { useUserStore } from '@/stores/user';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/HomePageModern.vue'),
    meta: { title: '首页' },
  },
  {
    path: '/universities',
    name: 'Universities',
    component: () => import('@/views/UniversitiesPageModern.vue'),
    meta: { title: '院校查询' },
  },
  {
    path: '/universities/:id',
    name: 'UniversityDetail',
    component: () => import('@/views/UniversityDetailPage.vue'),
    meta: { title: '院校详情' },
  },
  {
    path: '/majors',
    name: 'Majors',
    component: () => import('@/views/MajorsPage.vue'),
    meta: { title: '专业分析' },
  },
  {
    path: '/majors/:id',
    name: 'MajorDetail',
    component: () => import('@/views/MajorDetailPage.vue'),
    meta: { title: '专业详情' },
  },
  {
    path: '/recommendation',
    name: 'Recommendation',
    component: () => import('@/views/RecommendationPage.vue'),
    meta: { title: '智能推荐' },
  },
  {
    path: '/simulation',
    name: 'Simulation',
    component: () => import('@/views/RecommendationPage.vue'),
    meta: { title: '模拟填报' },
  },
  {
    path: '/analysis',
    name: 'Analysis',
    component: () => import('@/views/AnalysisPage.vue'),
    meta: { title: '数据分析' },
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/ProfilePage.vue'),
    meta: { title: '个人中心', requiresAuth: true },
  },
  // 暂时隐藏会员服务，支付功能待开发
  // {
  //   path: '/membership',
  //   name: 'Membership',
  //   component: () => import('@/views/MembershipPage.vue'),
  //   meta: { title: '会员服务' },
  // },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/LoginPage.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/LoginPage.vue'),
    meta: { title: '注册' },
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition;
    } else {
      return { top: 0 };
    }
  },
});

// 路由守卫
router.beforeEach((to, _from, next) => {
  // 设置页面标题
  document.title = `${to.meta.title} - 高考志愿填报助手`;

  // 检查是否需要登录：以 Pinia store 为单一信源，避免 store 内存与 localStorage 不一致。
  if (to.meta.requiresAuth) {
    const userStore = useUserStore();
    if (!userStore.isLoggedIn) {
      next({
        path: '/login',
        query: { redirect: to.fullPath },
      });
      return;
    }
  }

  next();
});

export default router;
