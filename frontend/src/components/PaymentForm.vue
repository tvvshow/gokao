<template>
  <div class="payment-form">
    <h2>选择会员套餐</h2>
    <div class="plan-selection">
      <div
        v-for="plan in plans"
        :key="plan.plan_code"
        :class="['plan-card', { selected: selectedPlan === plan.plan_code }]"
        @click="selectPlan(plan.plan_code)"
      >
        <h3>{{ plan.name }}</h3>
        <div class="price">¥{{ plan.price }}</div>
        <div class="duration">{{ plan.duration_days }}天</div>
        <ul class="features">
          <li
            v-for="(item, index) in normalizedFeatures(plan.features)"
            :key="index"
          >
            <span v-if="item.included">✓</span>
            <span v-else>✗</span>
            {{ item.name }}
          </li>
        </ul>
        <div class="limits">
          <div>
            查询次数:
            {{ plan.max_queries === -1 ? '无限制' : plan.max_queries }}
          </div>
          <div>
            下载次数:
            {{ plan.max_downloads === -1 ? '无限制' : plan.max_downloads }}
          </div>
        </div>
      </div>
    </div>

    <div v-if="selectedPlan" class="payment-method">
      <h3>选择支付方式</h3>
      <div class="payment-options">
        <button
          v-for="channel in paymentChannels"
          :key="channel.value"
          :class="[
            'payment-btn',
            { selected: selectedPaymentChannel === channel.value },
          ]"
          @click="selectPaymentChannel(channel.value)"
        >
          {{ channel.label }}
        </button>
      </div>
    </div>

    <button
      v-if="selectedPlan && selectedPaymentChannel"
      class="pay-button"
      @click="createOrder"
      :disabled="loading"
    >
      {{ loading ? '处理中...' : '立即支付' }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { usePaymentStore } from '@/stores/payment';
import type { MembershipPlan } from '@/types/payment';

interface PaymentChannel {
  value: string;
  label: string;
}

const paymentStore = usePaymentStore();

const plans = ref<MembershipPlan[]>([]);
const selectedPlan = ref<string>('');
const selectedPaymentChannel = ref<string>('');
const loading = ref<boolean>(false);

const paymentChannels: PaymentChannel[] = [
  { value: 'alipay', label: '支付宝' },
  { value: 'wechat', label: '微信支付' },
  { value: 'unionpay', label: '银联支付' },
];

onMounted(async () => {
  try {
    plans.value = await paymentStore.getMembershipPlans();
  } catch (error) {
    console.error('获取会员套餐失败:', error);
  }
});

const selectPlan = (planCode: string | undefined) => {
  if (planCode) {
    selectedPlan.value = planCode;
  }
};

const selectPaymentChannel = (channel: string) => {
  selectedPaymentChannel.value = channel;
};

// 统一处理 features 格式（支持数组或对象）
interface FeatureItem {
  name: string;
  included: boolean;
}

const normalizedFeatures = (
  features: string[] | Record<string, boolean>
): FeatureItem[] => {
  if (Array.isArray(features)) {
    // 数组格式：直接显示所有特性
    return features.map((name) => ({ name, included: true }));
  } else {
    // 对象格式：键为特性名，值为是否包含
    return Object.entries(features).map(([name, included]) => ({
      name,
      included: Boolean(included),
    }));
  }
};

const createOrder = async () => {
  if (!selectedPlan.value || !selectedPaymentChannel.value) return;

  const response = await paymentStore.createOrder({
    plan_code: selectedPlan.value,
    payment_channel: selectedPaymentChannel.value,
  });
  if (response.success) {
    ElMessage.success('订单已创建');
  } else {
    ElMessage.error(response.message || '创建订单失败');
  }
};
</script>

<style scoped>
.payment-form {
  max-width: 980px;
  margin: 0 auto;
  padding: 1.25rem;
  border: 1px solid #e2e8f0;
  border-radius: 1rem;
  background: linear-gradient(180deg, #fff 0%, #f8fafc 100%);
  box-shadow: 0 10px 30px -24px rgba(15, 23, 42, 0.55);
}

.plan-selection {
  display: flex;
  gap: 1rem;
  margin-bottom: 30px;
  flex-wrap: wrap;
}

.plan-card {
  flex: 1;
  min-width: 250px;
  border: 1px solid #dbe3ef;
  border-radius: 0.875rem;
  padding: 1rem;
  cursor: pointer;
  transition: all 0.3s ease;
  background: #fff;
}

.plan-card:hover {
  border-color: #38bdf8;
  transform: translateY(-2px);
}

.plan-card.selected {
  border-color: #0ea5e9;
  background-color: #f0f9ff;
  box-shadow: 0 10px 24px -20px rgba(2, 132, 199, 0.6);
}

.plan-card h3 {
  margin-top: 0;
  color: #0f172a;
}

.price {
  font-size: 24px;
  font-weight: bold;
  color: #0ea5e9;
  margin: 10px 0;
}

.duration {
  color: #475569;
  margin-bottom: 15px;
}

.features {
  list-style: none;
  padding: 0;
  margin: 15px 0;
}

.features li {
  padding: 5px 0;
}

.features li span {
  margin-right: 5px;
}

.limits {
  font-size: 14px;
  color: #475569;
  margin-top: 10px;
}

.payment-method {
  margin-bottom: 30px;
}

.payment-options {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.payment-btn {
  padding: 10px 20px;
  border: 1px solid #dbe3ef;
  background-color: #fff;
  border-radius: 0.625rem;
  cursor: pointer;
  transition: all 0.3s ease;
}

.payment-btn:hover {
  border-color: #0ea5e9;
}

.payment-btn.selected {
  border-color: #0ea5e9;
  background-color: #e0f2fe;
  color: #0369a1;
}

.pay-button {
  width: 100%;
  padding: 0.875rem 1rem;
  background-color: #0ea5e9;
  color: white;
  border: none;
  border-radius: 0.75rem;
  font-size: 16px;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.pay-button:hover:not(:disabled) {
  background-color: #38bdf8;
}

.pay-button:disabled {
  background-color: #7dd3fc;
  cursor: not-allowed;
}

.dark .payment-form {
  border-color: #334155;
  background: linear-gradient(180deg, #1f2937 0%, #111827 100%);
}

.dark .plan-card {
  border-color: #334155;
  background: #1f2937;
}

.dark .plan-card h3 {
  color: #f1f5f9;
}

.dark .duration,
.dark .limits,
.dark .payment-method h3 {
  color: #94a3b8;
}

.dark .plan-card.selected {
  border-color: #22d3ee;
  background: rgba(14, 116, 144, 0.22);
}

.dark .payment-btn {
  border-color: #334155;
  background: #1f2937;
  color: #e2e8f0;
}

.dark .payment-btn.selected {
  border-color: #22d3ee;
  background: rgba(14, 116, 144, 0.24);
  color: #67e8f9;
}

@media (max-width: 768px) {
  .payment-form {
    padding: 1rem;
  }

  .plan-card {
    min-width: 100%;
  }
}
</style>
