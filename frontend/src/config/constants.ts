/**
 * Application constants configuration
 * Centralized location for all static data used across the application
 */

// List of Chinese provinces
export const PROVINCES: string[] = [
  '北京',
  '上海',
  '天津',
  '重庆',
  '河北',
  '山西',
  '辽宁',
  '吉林',
  '黑龙江',
  '江苏',
  '浙江',
  '安徽',
  '福建',
  '江西',
  '山东',
  '河南',
  '湖北',
  '湖南',
  '广东',
  '海南',
  '四川',
  '贵州',
  '云南',
  '陕西',
  '甘肃',
  '青海',
  '内蒙古',
  '广西',
  '西藏',
  '宁夏',
  '新疆',
];

// Geographic regions of China
export const REGIONS: string[] = [
  '华北地区',
  '东北地区',
  '华东地区',
  '华中地区',
  '华南地区',
  '西南地区',
  '西北地区',
];

// Major categories based on Chinese education system
export const MAJOR_CATEGORIES: string[] = [
  '哲学',
  '经济学',
  '法学',
  '教育学',
  '文学',
  '历史学',
  '理学',
  '工学',
  '农学',
  '医学',
  '管理学',
  '艺术学',
];

// University types
export const UNIVERSITY_TYPES: string[] = [
  '985工程',
  '211工程',
  '双一流',
  '普通本科',
  '师范类',
  '财经类',
  '理工类',
  '医药类',
  '农林类',
  '艺术类',
  '体育类',
  '民族类',
];

// Risk tolerance options
export const RISK_TOLERANCE_OPTIONS = [
  { value: 'conservative', label: '保守型（冲1保5稳4）' },
  { value: 'moderate', label: '稳健型（冲2保3稳5）' },
  { value: 'aggressive', label: '激进型（冲4保2稳4）' },
] as const;

// Science type options
export const SCIENCE_TYPE_OPTIONS = [
  { value: '理科', label: '理科' },
  { value: '文科', label: '文科' },
  { value: '新高考', label: '新高考' },
] as const;

// Score range
export const SCORE_RANGE = {
  min: 0,
  max: 750,
} as const;

// Risk level mapping
export const RISK_LEVEL_MAP: Record<string, string> = {
  conservative: '低风险',
  moderate: '中风险',
  aggressive: '高风险',
};

// Default home page statistics (fallback when API is unavailable)
export const DEFAULT_HOME_STATISTICS = {
  universityCount: 2700,
  majorCount: 1400,
  userCount: 500000,
  accuracyRate: 95,
} as const;

// Feature list for home page
export const HOME_FEATURES = [
  {
    id: 'ai-recommendation',
    title: 'AI智能推荐',
    description: '基于大数据分析，为您量身定制最优志愿方案',
    link: '/recommendation',
  },
  {
    id: 'university-search',
    title: '院校查询',
    description: '2700+高校详细信息，历年分数线一目了然',
    link: '/universities',
  },
  {
    id: 'major-analysis',
    title: '专业分析',
    description: '1400+专业深度解析，就业前景全面了解',
    link: '/majors',
  },
  {
    id: 'data-analysis',
    title: '数据分析',
    description: '多维度数据可视化，科学决策有依据',
    link: '/analysis',
  },
  {
    id: 'membership',
    title: '会员服务',
    description: '专属顾问一对一指导，VIP专享服务',
    link: '/membership',
  },
  {
    id: 'simulation',
    title: '模拟填报',
    description: '真实模拟填报环境，提前体验填报流程',
    link: '/simulation',
  },
] as const;

// Recommendation process steps
export const RECOMMENDATION_STEPS = [
  {
    title: '输入成绩',
    description: '填写高考成绩和基本信息',
  },
  {
    title: 'AI分析',
    description: '智能算法分析匹配度',
  },
  {
    title: '生成方案',
    description: '个性化志愿填报方案',
  },
  {
    title: '优化调整',
    description: '根据偏好微调完善',
  },
] as const;
