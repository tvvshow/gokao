<template>
  <div class="membership-page">
    <div class="container">
      <div class="page-header">
        <h1 class="page-title">会员服务</h1>
        <p class="page-subtitle">解锁更多功能，享受专业服务</p>
      </div>

      <!-- 会员套餐 -->
      <div class="plans-section">
        <el-row :gutter="30">
          <el-col :span="8" v-for="plan in membershipPlans" :key="plan.id">
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
              <el-icon v-if="row.free" color="#67c23a"><check /></el-icon>
              <el-icon v-else color="#f56c6c"><close /></el-icon>
            </template>
          </el-table-column>
          <el-table-column prop="basic" label="基础版" align="center">
            <template #default="{ row }">
              <el-icon v-if="row.basic" color="#67c23a"><check /></el-icon>
              <el-icon v-else color="#f56c6c"><close /></el-icon>
            </template>
          </el-table-column>
          <el-table-column prop="premium" label="专业版" align="center">
            <template #default="{ row }">
              <el-icon v-if="row.premium" color="#67c23a"><check /></el-icon>
              <el-icon v-else color="#f56c6c"><close /></el-icon>
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
  padding: 20px 0;
  min-height: calc(100vh - 160px);
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.page-header {
  text-align: center;
  margin-bottom: 50px;
}

.page-title {
  font-size: 28px;
  color: #2c3e50;
  margin-bottom: 12px;
}

.page-subtitle {
  color: #7f8c8d;
  font-size: 16px;
}

.plans-section {
  margin-bottom: 60px;
}

.plan-card {
  height: 100%;
  transition: all 0.3s ease;
  margin-bottom: 20px;
}

.plan-card.featured {
  border: 2px solid #667eea;
}

.plan-card:hover {
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
}

.plan-header {
  text-align: center;
}

.plan-name {
  font-size: 20px;
  color: #2c3e50;
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
  color: #7f8c8d;
}

.amount {
  font-size: 32px;
  font-weight: 700;
  color: #667eea;
  margin: 0 4px;
}

.period {
  font-size: 14px;
  color: #7f8c8d;
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
  color: #2c3e50;
  font-size: 14px;
}

.plan-features li .el-icon {
  color: #67c23a;
  margin-right: 8px;
  flex-shrink: 0;
}

.featured-btn {
  background: linear-gradient(45deg, #667eea, #764ba2);
  border: none;
}

.comparison-section {
  margin-top: 60px;
}

.comparison-section h2 {
  text-align: center;
  margin-bottom: 30px;
  color: #2c3e50;
  font-size: 24px;
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
  .container {
    padding: 0 16px;
  }

  .page-header {
    margin-bottom: 30px;
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
</style>
