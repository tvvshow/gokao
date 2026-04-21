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
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.order-history h2 {
  text-align: center;
  color: #333;
}

.loading,
.no-orders {
  text-align: center;
  padding: 40px;
  color: #666;
}

.order-filters {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.order-filters select,
.order-filters input {
  padding: 8px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.orders-list {
  margin-bottom: 20px;
}

.order-item {
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  background-color: #fff;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e0e0e0;
}

.order-no {
  font-weight: bold;
  color: #333;
}

.order-status {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: bold;
}

.order-status.pending {
  background-color: #f0f8ff;
  color: #409eff;
}

.order-status.paid {
  background-color: #f0fff0;
  color: #67c23a;
}

.order-status.canceled {
  background-color: #fff0f0;
  color: #f56c6c;
}

.order-status.expired {
  background-color: #f5f5f5;
  color: #909399;
}

.order-status.refunded {
  background-color: #f5f5f5;
  color: #909399;
}

.order-status.refunding {
  background-color: #fdf6ec;
  color: #e6a23c;
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
  color: #666;
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
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.cancel-btn {
  background-color: #f56c6c;
  color: white;
}

.cancel-btn:hover {
  background-color: #f78989;
}

.invoice-btn {
  background-color: #409eff;
  color: white;
}

.invoice-btn:hover {
  background-color: #66b1ff;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
}

.pagination button {
  padding: 8px 12px;
  border: 1px solid #dcdfe6;
  background-color: #fff;
  border-radius: 4px;
  cursor: pointer;
}

.pagination button:disabled {
  background-color: #f5f7fa;
  color: #c0c4cc;
  cursor: not-allowed;
}

.pagination button:not(:disabled):hover {
  background-color: #ecf5ff;
  color: #409eff;
}
</style>
