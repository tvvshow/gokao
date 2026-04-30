<template>
  <div class="membership-status">
    <div v-if="membershipStatus.is_vip" class="vip-status">
      <h2>会员状态</h2>
      <div class="status-card">
        <div class="status-header">
          <span class="vip-badge">VIP</span>
          <span class="plan-name">{{ membershipStatus.plan_name }}</span>
        </div>
        <div class="status-details">
          <div class="detail-item">
            <span class="label">到期时间:</span>
            <span class="value">{{
              formatDate(membershipStatus.end_time)
            }}</span>
          </div>
          <div class="detail-item">
            <span class="label">剩余天数:</span>
            <span class="value">{{ membershipStatus.remaining_days }}天</span>
          </div>
          <div class="detail-item">
            <span class="label">自动续费:</span>
            <span class="value">{{
              membershipStatus.auto_renew ? '是' : '否'
            }}</span>
          </div>
        </div>
        <div class="usage-stats">
          <div class="stat-item">
            <div class="stat-label">查询次数</div>
            <div class="stat-value">
              {{ membershipStatus.used_queries }} /
              <span v-if="membershipStatus.max_queries === -1">无限制</span>
              <span v-else>{{ membershipStatus.max_queries }}</span>
            </div>
            <div class="stat-bar">
              <div
                class="stat-progress"
                :style="{
                  width:
                    calculatePercentage(
                      membershipStatus.used_queries,
                      membershipStatus.max_queries
                    ) + '%',
                }"
              ></div>
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-label">下载次数</div>
            <div class="stat-value">
              {{ membershipStatus.used_downloads }} /
              <span v-if="membershipStatus.max_downloads === -1">无限制</span>
              <span v-else>{{ membershipStatus.max_downloads }}</span>
            </div>
            <div class="stat-bar">
              <div
                class="stat-progress"
                :style="{
                  width:
                    calculatePercentage(
                      membershipStatus.used_downloads,
                      membershipStatus.max_downloads
                    ) + '%',
                }"
              ></div>
            </div>
          </div>
        </div>
        <div class="actions">
          <button @click="renewMembership" class="renew-btn">续费会员</button>
          <button @click="cancelMembership" class="cancel-btn">取消会员</button>
        </div>
      </div>
    </div>
    <div v-else class="non-vip-status">
      <h2>您还不是VIP会员</h2>
      <p>成为VIP会员可以享受更多功能和服务</p>
      <button @click="goToPayment" class="become-vip-btn">立即开通</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { usePaymentStore } from '@/stores/payment';
import type { MembershipStatus } from '@/types/payment';

const router = useRouter();
const paymentStore = usePaymentStore();

const membershipStatus = ref<MembershipStatus>({
  level: 'free',
  isActive: false,
  is_vip: false,
  plan_code: '',
  plan_name: '',
  start_time: null,
  end_time: null,
  remaining_days: 0,
  used_queries: 0,
  max_queries: 0,
  used_downloads: 0,
  max_downloads: 0,
  features: {},
  auto_renew: false,
});

onMounted(async () => {
  try {
    membershipStatus.value = await paymentStore.getMembershipStatus();
  } catch (error) {
    console.error('获取会员状态失败:', error);
  }
});

const formatDate = (date: string | null | undefined) => {
  if (!date) return '';
  return new Date(date).toLocaleDateString('zh-CN');
};

const calculatePercentage = (
  used: number | undefined,
  max: number | undefined
) => {
  if (!used || !max || max === -1 || max === 0) return 0;
  return Math.min(100, (used / max) * 100);
};

const renewMembership = async () => {
  try {
    // 这里可以打开续费对话框或跳转到支付页面
    console.log('续费会员');
  } catch (error) {
    console.error('续费会员失败:', error);
  }
};

const cancelMembership = async () => {
  try {
    await paymentStore.cancelMembership();
    // 重新获取会员状态
    membershipStatus.value = await paymentStore.getMembershipStatus();
  } catch (error) {
    console.error('取消会员失败:', error);
  }
};

const goToPayment = () => {
  router.push('/profile');
};
</script>

<style scoped>
.membership-status {
  max-width: 760px;
  margin: 0 auto;
  padding: 1.25rem;
  border: 1px solid #e2e8f0;
  border-radius: 1rem;
  background: linear-gradient(180deg, #fff 0%, #f8fafc 100%);
  box-shadow: 0 10px 30px -24px rgba(15, 23, 42, 0.55);
}

.vip-status h2,
.non-vip-status h2 {
  text-align: center;
  color: #0f172a;
}

.status-card {
  border: 1px solid #e2e8f0;
  border-radius: 0.875rem;
  padding: 20px;
  background-color: #fff;
}

.status-header {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.vip-badge {
  background-color: #fbbf24;
  color: #0f172a;
  padding: 5px 10px;
  border-radius: 9999px;
  font-weight: bold;
  margin-right: 10px;
}

.plan-name {
  font-size: 18px;
  font-weight: bold;
  color: #0f172a;
}

.status-details {
  margin-bottom: 20px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  padding: 5px 0;
}

.label {
  color: #475569;
}

.value {
  font-weight: bold;
}

.usage-stats {
  margin-bottom: 20px;
}

.stat-item {
  margin-bottom: 15px;
}

.stat-label {
  font-size: 14px;
  color: #475569;
  margin-bottom: 5px;
}

.stat-value {
  font-weight: bold;
  margin-bottom: 5px;
}

.stat-bar {
  height: 8px;
  background-color: #e2e8f0;
  border-radius: 9999px;
  overflow: hidden;
}

.stat-progress {
  height: 100%;
  background-color: #0ea5e9;
  transition: width 0.3s ease;
}

.actions {
  display: flex;
  gap: 10px;
}

.renew-btn,
.cancel-btn {
  flex: 1;
  padding: 10px;
  border: none;
  border-radius: 0.625rem;
  cursor: pointer;
  font-weight: bold;
}

.renew-btn {
  background-color: #0ea5e9;
  color: white;
}

.renew-btn:hover {
  background-color: #38bdf8;
}

.cancel-btn {
  background-color: #ef4444;
  color: white;
}

.cancel-btn:hover {
  background-color: #f87171;
}

.non-vip-status {
  text-align: center;
  padding: 40px 20px;
}

.non-vip-status p {
  color: #475569;
  margin: 20px 0;
}

.become-vip-btn {
  padding: 12px 30px;
  background-color: #0ea5e9;
  color: white;
  border: none;
  border-radius: 0.625rem;
  font-size: 16px;
  cursor: pointer;
  font-weight: bold;
}

.become-vip-btn:hover {
  background-color: #38bdf8;
}

.dark .membership-status {
  border-color: #334155;
  background: linear-gradient(180deg, #1f2937 0%, #111827 100%);
}

.dark .vip-status h2,
.dark .non-vip-status h2,
.dark .plan-name,
.dark .value,
.dark .stat-value {
  color: #f1f5f9;
}

.dark .status-card {
  border-color: #334155;
  background: #1f2937;
}

.dark .label,
.dark .stat-label,
.dark .non-vip-status p {
  color: #94a3b8;
}

.dark .stat-bar {
  background: #334155;
}

@media (max-width: 768px) {
  .membership-status {
    padding: 1rem;
  }

  .status-header,
  .detail-item,
  .actions {
    flex-direction: column;
    align-items: flex-start;
  }

  .actions {
    gap: 0.75rem;
  }

  .renew-btn,
  .cancel-btn {
    width: 100%;
  }
}
</style>
