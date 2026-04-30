<template>
  <div class="order-history">
    <h2>订单历史</h2>
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="orders.length === 0" class="no-orders">暂无订单记录</div>
    <div v-else>
      <div class="order-filters">
        <select v-model="filterStatus" @change="fetchOrders">
          <option value="">全部状态</option>
          <option value="pending">待支付</option>
          <option value="paid">已支付</option>
          <option value="canceled">已取消</option>
          <option value="expired">已过期</option>
          <option value="refunded">已退款</option>
        </select>
        <input
          v-model="startDate"
          type="date"
          @change="fetchOrders"
          placeholder="开始日期"
        />
        <input
          v-model="endDate"
          type="date"
          @change="fetchOrders"
          placeholder="结束日期"
        />
      </div>
      <div class="orders-list">
        <div v-for="order in orders" :key="order.order_no" class="order-item">
          <div class="order-header">
            <div class="order-no">订单号: {{ order.order_no }}</div>
            <div class="order-status" :class="order.status">
              {{ getStatusText(order.status) }}
            </div>
          </div>
          <div class="order-details">
            <div class="detail-row">
              <span class="label">商品:</span>
              <span class="value">{{ order.subject }}</span>
            </div>
            <div class="detail-row">
              <span class="label">金额:</span>
              <span class="value">¥{{ order.amount }}</span>
            </div>
            <div class="detail-row">
              <span class="label">支付方式:</span>
              <span class="value">{{
                getPaymentChannelText(order.payment_channel)
              }}</span>
            </div>
            <div class="detail-row">
              <span class="label">创建时间:</span>
              <span class="value">{{ formatDate(order.created_at) }}</span>
            </div>
            <div v-if="order.paid_at" class="detail-row">
              <span class="label">支付时间:</span>
              <span class="value">{{ formatDate(order.paid_at) }}</span>
            </div>
            <div v-if="order.expire_time" class="detail-row">
              <span class="label">过期时间:</span>
              <span class="value">{{ formatDate(order.expire_time) }}</span>
            </div>
          </div>
          <div class="order-actions">
            <button
              v-if="order.status === 'pending'"
              @click="cancelOrder(order.order_no)"
              class="cancel-btn"
            >
              取消订单
            </button>
            <button
              v-if="order.status === 'paid'"
              @click="getInvoice(order.order_no)"
              class="invoice-btn"
            >
              查看发票
            </button>
          </div>
        </div>
      </div>
      <div class="pagination">
        <button
          :disabled="currentPage === 1"
          @click="changePage(currentPage - 1)"
        >
          上一页
        </button>
        <span>第 {{ currentPage }} 页，共 {{ totalPages }} 页</span>
        <button
          :disabled="currentPage === totalPages"
          @click="changePage(currentPage + 1)"
        >
          下一页
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { usePaymentStore } from '@/stores/payment';
import type { PaymentOrder } from '@/types/payment';

const paymentStore = usePaymentStore();

const orders = ref<PaymentOrder[]>([]);
const loading = ref<boolean>(false);
const filterStatus = ref<string>('');
const startDate = ref<string>('');
const endDate = ref<string>('');
const currentPage = ref<number>(1);
const pageSize = ref<number>(10);
const total = ref<number>(0);

const totalPages = computed(() => {
  return Math.ceil(total.value / pageSize.value);
});

onMounted(() => {
  fetchOrders();
});

const fetchOrders = async () => {
  loading.value = true;
  try {
    const response = await paymentStore.getOrderList({
      page: currentPage.value,
      page_size: pageSize.value,
      status: filterStatus.value,
      start_time: startDate.value,
      end_time: endDate.value,
    });
    orders.value = response.orders;
    total.value = response.total;
  } catch (error) {
    console.error('获取订单列表失败:', error);
  } finally {
    loading.value = false;
  }
};

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    pending: '待支付',
    paid: '已支付',
    canceled: '已取消',
    cancelled: '已取消',
    expired: '已过期',
    refunded: '已退款',
    refunding: '退款中',
  };
  return statusMap[status] || status;
};

const getPaymentChannelText = (channel: string | undefined) => {
  if (!channel) return '';
  const channelMap: Record<string, string> = {
    alipay: '支付宝',
    wechat: '微信支付',
    unionpay: '银联支付',
  };
  return channelMap[channel] || channel;
};

const formatDate = (date: string | null | undefined) => {
  if (!date) return '';
  return new Date(date).toLocaleString('zh-CN');
};

const cancelOrder = async (orderNo: string | undefined) => {
  if (!orderNo) return;
  try {
    await paymentStore.cancelOrder(orderNo);
    // 重新获取订单列表
    fetchOrders();
  } catch (error) {
    console.error('取消订单失败:', error);
  }
};

const getInvoice = async (orderNo: string | undefined) => {
  if (!orderNo) return;
  try {
    const invoice = await paymentStore.getInvoice(orderNo);
    console.log('发票信息:', invoice);
    // 这里可以显示发票详情或下载发票
  } catch (error) {
    console.error('获取发票失败:', error);
  }
};

const changePage = (page: number) => {
  currentPage.value = page;
  fetchOrders();
};
</script>

<style scoped>
.order-history {
  max-width: 980px;
  margin: 0 auto;
  padding: 1.25rem;
  border: 1px solid #e2e8f0;
  border-radius: 1rem;
  background: linear-gradient(180deg, #fff 0%, #f8fafc 100%);
  box-shadow: 0 10px 30px -24px rgba(15, 23, 42, 0.55);
}

.order-history h2 {
  text-align: center;
  color: #0f172a;
}

.loading,
.no-orders {
  text-align: center;
  padding: 40px;
  color: #475569;
}

.order-filters {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.order-filters select,
.order-filters input {
  padding: 0.625rem 0.75rem;
  border: 1px solid #dbe3ef;
  border-radius: 0.625rem;
}

.orders-list {
  margin-bottom: 20px;
}

.order-item {
  border: 1px solid #e2e8f0;
  border-radius: 0.875rem;
  padding: 20px;
  margin-bottom: 20px;
  background: #fff;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e2e8f0;
}

.order-no {
  font-weight: bold;
  color: #0f172a;
}

.order-status {
  padding: 4px 10px;
  border-radius: 9999px;
  font-size: 12px;
  font-weight: bold;
}

.order-status.pending {
  background-color: #f0f9ff;
  color: #0ea5e9;
}

.order-status.paid {
  background-color: #f0fdf4;
  color: #22c55e;
}

.order-status.canceled {
  background-color: #fef2f2;
  color: #ef4444;
}

.order-status.expired {
  background-color: #f5f5f5;
  color: #94a3b8;
}

.order-status.refunded {
  background-color: #f5f5f5;
  color: #94a3b8;
}

.order-status.refunding {
  background-color: #fffbeb;
  color: #f59e0b;
}

.order-details {
  margin-bottom: 15px;
}

.detail-row {
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

.order-actions {
  display: flex;
  gap: 10px;
}

.cancel-btn,
.invoice-btn {
  padding: 8px 14px;
  border: none;
  border-radius: 0.625rem;
  cursor: pointer;
  font-size: 14px;
}

.cancel-btn {
  background-color: #ef4444;
  color: white;
}

.cancel-btn:hover {
  background-color: #f87171;
}

.invoice-btn {
  background-color: #0ea5e9;
  color: white;
}

.invoice-btn:hover {
  background-color: #38bdf8;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
}

.pagination button {
  padding: 8px 12px;
  border: 1px solid #dbe3ef;
  background-color: #fff;
  border-radius: 0.625rem;
  cursor: pointer;
}

.pagination button:disabled {
  background-color: #f1f5f9;
  color: #c0c4cc;
  cursor: not-allowed;
}

.pagination button:not(:disabled):hover {
  background-color: #e0f2fe;
  color: #0ea5e9;
}

.dark .order-history {
  border-color: #334155;
  background: linear-gradient(180deg, #1f2937 0%, #111827 100%);
}

.dark .order-history h2,
.dark .order-no,
.dark .value {
  color: #f1f5f9;
}

.dark .loading,
.dark .no-orders,
.dark .label {
  color: #94a3b8;
}

.dark .order-filters select,
.dark .order-filters input,
.dark .pagination button {
  border-color: #334155;
  background: #1f2937;
  color: #e2e8f0;
}

.dark .order-item {
  border-color: #334155;
  background: #1f2937;
}

.dark .order-header {
  border-bottom-color: #334155;
}

@media (max-width: 768px) {
  .order-history {
    padding: 1rem;
  }

  .order-header,
  .detail-row,
  .order-actions {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }

  .order-actions button {
    width: 100%;
  }
}
</style>
