<template>
  <div class="membership-page min-h-screen bg-gray-50 dark:bg-gray-900">
    <div class="container-modern py-8">
      <div class="page-header">
        <h1 class="page-title">会员服务</h1>
        <p class="page-subtitle">解锁更多功能，享受专业服务</p>
      </div>

      <!-- 会员套餐 -->
      <div class="plans-section">
        <el-row :gutter="24">
          <el-col
            v-for="plan in membershipPlans"
            :key="plan.id"
            :xs="24"
            :md="12"
            :lg="8"
          >
            <el-card class="plan-card" :class="{ featured: plan.featured }">
              <template #header>
                <div class="plan-header">
                  <h3 class="plan-name">{{ plan.name }}</h3>
                  <div class="plan-price">
                    <span class="currency">¥</span>
                    <span class="amount">{{ plan.price }}</span>
                    <span class="period">/{{ plan.period }}</span>
                  </div>
                </div>
              </template>

              <div class="plan-features">
                <ul>
                  <li
                    v-for="feature in normalizePlanFeatures(plan.features)"
                    :key="feature"
                  >
                    <el-icon><check /></el-icon>
                    <span>{{ feature }}</span>
                  </li>
                </ul>
              </div>

              <template #footer>
                <el-button
                  type="primary"
                  :class="{ 'featured-btn': plan.featured }"
                  @click="selectPlan(plan)"
                  style="width: 100%"
                >
                  {{ plan.buttonText }}
                </el-button>
              </template>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <!-- 功能对比 -->
      <div class="comparison-section">
        <h2>功能对比</h2>
        <el-table :data="featureComparison" border>
          <el-table-column prop="feature" label="功能" width="200" />
          <el-table-column prop="free" label="免费版" align="center">
            <template #default="{ row }">
              <el-icon v-if="row.free" color="#22c55e"><check /></el-icon>
              <el-icon v-else color="#ef4444"><close /></el-icon>
            </template>
          </el-table-column>
          <el-table-column prop="basic" label="基础版" align="center">
            <template #default="{ row }">
              <el-icon v-if="row.basic" color="#22c55e"><check /></el-icon>
              <el-icon v-else color="#ef4444"><close /></el-icon>
            </template>
          </el-table-column>
          <el-table-column prop="premium" label="专业版" align="center">
            <template #default="{ row }">
              <el-icon v-if="row.premium" color="#22c55e"><check /></el-icon>
              <el-icon v-else color="#ef4444"><close /></el-icon>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { Check, Close } from '@element-plus/icons-vue';
import { usePaymentStore } from '@/stores/payment';
import type { MembershipPlanItem } from '@/types/payment';

const paymentStore = usePaymentStore();
const membershipPlans = ref<MembershipPlanItem[]>([
  {
    id: 'free',
    name: '免费版',
    price: 0,
    period: '永久',
    featured: false,
    buttonText: '当前版本',
    features: ['基础院校查询', '简单数据分析', '每日10次查询', '基础推荐功能'],
  },
]);

// 功能对比
const featureComparison = ref([
  { feature: '院校查询', free: true, basic: true, premium: true },
  { feature: '专业分析', free: true, basic: true, premium: true },
  { feature: 'AI智能推荐', free: false, basic: true, premium: true },
  { feature: '历史趋势分析', free: false, basic: true, premium: true },
  { feature: '就业数据报告', free: false, basic: true, premium: true },
  { feature: '一对一咨询', free: false, basic: false, premium: true },
  { feature: '定制化报告', free: false, basic: false, premium: true },
]);

const selectPlan = (plan: MembershipPlanItem) => {
  if (plan.id === 'free') {
    ElMessage.info('您当前使用的是免费版');
    return;
  }

  // TODO: 支付功能开发中
  ElMessage.info('支付功能开发中，敬请期待');
  // 这里可以跳转到支付页面
};

const normalizePlanFeatures = (
  features: MembershipPlanItem['features']
): string[] => {
  if (Array.isArray(features)) {
    return features;
  }
  return Object.keys(features || {});
};

onMounted(async () => {
  try {
    const plans = await paymentStore.getMembershipPlans();
    const dynamicPlans = plans.map((plan) => {
      const days = plan.duration_days || 30;
      return {
        ...plan,
        period: `${days}天`,
        featured: plan.plan_code === 'premium' || plan.recommended,
        buttonText: '立即购买',
        features: Array.isArray(plan.features)
          ? plan.features
          : Object.keys(plan.features || {}),
      } as MembershipPlanItem;
    });
    membershipPlans.value = [membershipPlans.value[0], ...dynamicPlans];
  } catch {
    // keep fallback plans
  }
});
</script>

<style scoped>
.membership-page {
  min-height: calc(100vh - 160px);
}

.page-header {
  text-align: center;
  margin-bottom: 40px;
}

.page-title {
  font-size: 2rem;
  color: #0f172a;
  margin-bottom: 0.75rem;
  letter-spacing: -0.02em;
}

.page-subtitle {
  color: #475569;
  font-size: 1rem;
}

.plans-section {
  margin-bottom: 48px;
}

.plan-card {
  height: 100%;
  transition: all 0.3s ease;
  border-radius: 1rem;
  border: 1px solid #e2e8f0;
  background: linear-gradient(180deg, #fff 0%, #f8fafc 100%);
  box-shadow: 0 10px 30px -24px rgba(15, 23, 42, 0.55);
}

.plan-card.featured {
  border: 2px solid #0ea5e9;
  box-shadow:
    0 14px 30px -22px rgba(2, 132, 199, 0.58),
    0 16px 30px -24px rgba(15, 23, 42, 0.6);
}

.plan-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 28px -22px rgba(15, 23, 42, 0.7);
}

.plan-header {
  text-align: center;
}

.plan-name {
  font-size: 20px;
  color: #0f172a;
  margin-bottom: 16px;
}

.plan-price {
  display: flex;
  align-items: baseline;
  justify-content: center;
  margin-bottom: 16px;
}

.currency {
  font-size: 16px;
  color: #64748b;
}

.amount {
  font-size: 32px;
  font-weight: 700;
  color: #0ea5e9;
  margin: 0 4px;
}

.period {
  font-size: 14px;
  color: #64748b;
}

.plan-features ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.plan-features li {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  color: #0f172a;
  font-size: 14px;
}

.plan-features li .el-icon {
  color: #22c55e;
  margin-right: 8px;
  flex-shrink: 0;
}

.featured-btn {
  background: linear-gradient(45deg, #0ea5e9, #0f766e);
  border: none;
}

.comparison-section {
  margin-top: 42px;
}

.comparison-section h2 {
  text-align: center;
  margin-bottom: 22px;
  color: #0f172a;
  font-size: 1.5rem;
}

.comparison-section :deep(.el-table) {
  border-radius: 0.875rem;
  overflow: hidden;
  border: 1px solid #dbe3ef;
}

.comparison-section :deep(.el-table th.el-table__cell) {
  background: #f8fafc;
}

/* Responsive design */
@media (min-width: 768px) {
  .page-title {
    font-size: 32px;
  }

  .plan-name {
    font-size: 24px;
  }

  .amount {
    font-size: 36px;
  }

  .plan-card.featured {
    transform: scale(1.05);
  }

  .comparison-section h2 {
    font-size: 28px;
  }
}

@media (max-width: 991px) {
  .plans-section :deep(.el-col) {
    width: 100%;
    max-width: 100%;
    flex: 0 0 100%;
    margin-bottom: 20px;
  }

  .plan-card.featured {
    transform: none;
  }
}

@media (max-width: 767px) {
  .page-header {
    margin-bottom: 28px;
  }

  .plans-section {
    margin-bottom: 40px;
  }

  .comparison-section {
    margin-top: 40px;
    overflow-x: auto;
  }

  .comparison-section :deep(.el-table) {
    min-width: 500px;
  }
}

.dark .page-title {
  color: #f1f5f9;
}

.dark .page-subtitle {
  color: #94a3b8;
}

.dark .plan-card {
  border-color: #334155;
  background: linear-gradient(180deg, #1f2937 0%, #111827 100%);
}

.dark .plan-name,
.dark .plan-features li,
.dark .comparison-section h2 {
  color: #f1f5f9;
}

.dark .currency,
.dark .period {
  color: #94a3b8;
}

.dark .comparison-section :deep(.el-table) {
  border-color: #334155;
}

.dark .comparison-section :deep(.el-table th.el-table__cell) {
  background: #1f2937;
}
</style>
