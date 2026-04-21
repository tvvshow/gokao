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
          <li v-for="(item, index) in normalizedFeatures(plan.features)" :key="index">
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

const normalizedFeatures = (features: string[] | Record<string, boolean>): FeatureItem[] => {
  if (Array.isArray(features)) {
    // 数组格式：直接显示所有特性
    return features.map(name => ({ name, included: true }));
  } else {
    // 对象格式：键为特性名，值为是否包含
    return Object.entries(features).map(([name, included]) => ({ name, included: Boolean(included) }));
  }
};

const createOrder = async () => {
  if (!selectedPlan.value || !selectedPaymentChannel.value) return;

  // TODO: 支付功能开发中
  ElMessage.info('支付功能开发中，敬请期待');
  return;

  loading.value = true;
  try {
    const order = await paymentStore.createOrder({
      plan_code: selectedPlan.value,
      payment_channel: selectedPaymentChannel.value,
    });
    console.log('订单创建成功:', order);
    // 这里可以跳转到支付页面或显示支付二维码
  } catch (error) {
    console.error('创建订单失败:', error);
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.payment-form {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.plan-selection {
  display: flex;
  gap: 20px;
  margin-bottom: 30px;
  flex-wrap: wrap;
}

.plan-card {
  flex: 1;
  min-width: 250px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.plan-card:hover {
  border-color: #409eff;
}

.plan-card.selected {
  border-color: #409eff;
  background-color: #f0f8ff;
}

.plan-card h3 {
  margin-top: 0;
  color: #333;
}

.price {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
  margin: 10px 0;
}

.duration {
  color: #666;
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
  color: #666;
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
  border: 1px solid #dcdfe6;
  background-color: #fff;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.payment-btn:hover {
  border-color: #409eff;
}

.payment-btn.selected {
  border-color: #409eff;
  background-color: #ecf5ff;
  color: #409eff;
}

.pay-button {
  width: 100%;
  padding: 15px;
  background-color: #409eff;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.pay-button:hover:not(:disabled) {
  background-color: #66b1ff;
}

.pay-button:disabled {
  background-color: #a0cfff;
  cursor: not-allowed;
}
</style>
