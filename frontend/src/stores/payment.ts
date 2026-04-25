import { defineStore } from 'pinia';
import { ref } from 'vue';
import type {
  MembershipStatus,
  MembershipPlan,
  PaymentOrder,
  CreateOrderParams,
} from '@/types/payment';

export interface OrderListParams {
  page?: number;
  page_size?: number;
  status?: string;
  start_time?: string;
  end_time?: string;
}

export interface OrderListResponse {
  orders: PaymentOrder[];
  total: number;
}

export const usePaymentStore = defineStore('payment', () => {
  const membershipStatus = ref<MembershipStatus>({
    level: 'free',
    isActive: false,
    is_vip: false,
  });

  const plans = ref<MembershipPlan[]>([]);
  const orders = ref<PaymentOrder[]>([]);
  const loading = ref(false);

  async function fetchMembershipStatus(): Promise<MembershipStatus> {
    loading.value = true;
    try {
      // TODO: Implement API call
      const status: MembershipStatus = {
        level: 'free',
        isActive: false,
        is_vip: false,
      };
      membershipStatus.value = status;
      return status;
    } finally {
      loading.value = false;
    }
  }

  // Alias for component compatibility
  async function getMembershipStatus(): Promise<MembershipStatus> {
    return fetchMembershipStatus();
  }

  async function fetchPlans(): Promise<MembershipPlan[]> {
    loading.value = true;
    try {
      // TODO: Implement API call
      plans.value = [];
      return plans.value;
    } finally {
      loading.value = false;
    }
  }

  // Alias for component compatibility
  async function getMembershipPlans(): Promise<MembershipPlan[]> {
    return fetchPlans();
  }

  async function fetchOrders(): Promise<PaymentOrder[]> {
    loading.value = true;
    try {
      // TODO: Implement API call
      orders.value = [];
      return orders.value;
    } finally {
      loading.value = false;
    }
  }

  // Alias for component compatibility with params
  async function getOrderList(params?: OrderListParams): Promise<OrderListResponse> {
    loading.value = true;
    try {
      // TODO: Implement API call with params
      console.log('Fetching orders with params:', params);
      orders.value = [];
      return { orders: orders.value, total: 0 };
    } finally {
      loading.value = false;
    }
  }

  async function createOrder(params: string | CreateOrderParams) {
    loading.value = true;
    try {
      const planId = typeof params === 'string' ? params : params.plan_code;
      // TODO: Implement API call
      console.log('Creating order for plan:', planId);
      return { success: true, orderId: 'mock-order-id' };
    } finally {
      loading.value = false;
    }
  }

  async function cancelOrder(orderId: string) {
    loading.value = true;
    try {
      // TODO: Implement API call
      console.log('Cancelling order:', orderId);
      return { success: true };
    } finally {
      loading.value = false;
    }
  }

  async function cancelMembership() {
    loading.value = true;
    try {
      // TODO: Implement API call
      console.log('Cancelling membership');
      return { success: true };
    } finally {
      loading.value = false;
    }
  }

  async function getInvoice(orderId: string) {
    loading.value = true;
    try {
      // TODO: Implement API call
      console.log('Getting invoice for order:', orderId);
      return { success: true, url: '' };
    } finally {
      loading.value = false;
    }
  }

  return {
    membershipStatus,
    plans,
    orders,
    loading,
    fetchMembershipStatus,
    getMembershipStatus,
    fetchPlans,
    getMembershipPlans,
    fetchOrders,
    getOrderList,
    createOrder,
    cancelOrder,
    cancelMembership,
    getInvoice,
  };
});
