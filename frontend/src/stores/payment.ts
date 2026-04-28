import { defineStore } from 'pinia';
import { ref } from 'vue';
import { api } from '@/api/api-client';
import type {
  MembershipStatus,
  MembershipPlan,
  PaymentOrder,
  CreateOrderParams,
} from '@/types/payment';
import { isWrappedResponse } from '@/utils/api-response';
import { loadFromStorage, saveToStorage } from '@/utils/storage';

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

const MEMBERSHIP_KEY = 'payment_membership_status';
const ORDERS_KEY = 'payment_orders';

const DEFAULT_PLANS: MembershipPlan[] = [
  {
    id: 'basic',
    plan_code: 'basic',
    name: '基础会员',
    price: 29,
    duration: '30天',
    duration_days: 30,
    max_queries: 200,
    max_downloads: 20,
    features: ['院校查询增强', '推荐结果保存', '基础导出'],
  },
  {
    id: 'premium',
    plan_code: 'premium',
    name: '高级会员',
    price: 99,
    duration: '90天',
    duration_days: 90,
    max_queries: 1000,
    max_downloads: 100,
    recommended: true,
    features: ['无限制推荐生成', '推荐方案导出', '录取趋势分析'],
  },
];

function getDefaultMembershipStatus(): MembershipStatus {
  return {
    level: 'free',
    isActive: false,
    is_vip: false,
    plan_code: '',
    plan_name: '',
    start_time: null,
    end_time: null,
    remaining_days: 0,
    auto_renew: false,
    used_queries: 0,
    max_queries: 0,
    used_downloads: 0,
    max_downloads: 0,
    features: {},
  };
}

function mapPlanToLevel(planCode: string): MembershipStatus['level'] {
  return planCode === 'premium' ? 'premium' : 'basic';
}

function normalizePlan(raw: Record<string, unknown>): MembershipPlan {
  const planCode = String(raw.plan_code ?? raw.id ?? '');
  const days = Number(raw.duration_days ?? 30) || 30;
  return {
    id: String(raw.id ?? planCode),
    plan_code: planCode,
    name: String(raw.name ?? planCode),
    price: Number(raw.price ?? 0),
    duration: `${days}天`,
    duration_days: days,
    max_queries: Number(raw.max_queries ?? 0),
    max_downloads: Number(raw.max_downloads ?? 0),
    features: Array.isArray(raw.features)
      ? (raw.features as string[])
      : typeof raw.features === 'object' && raw.features !== null
        ? (raw.features as Record<string, boolean>)
        : [],
  };
}

function normalizeMembershipStatus(
  raw: Record<string, unknown>
): MembershipStatus {
  const isVIP = Boolean(raw.is_vip ?? raw.isVIP ?? false);
  const planCode = String(raw.plan_code ?? raw.planCode ?? '');
  const isPremiumLike =
    planCode === 'premium' ||
    planCode === 'enterprise' ||
    planCode === 'ultimate';
  const endTime =
    (raw.end_time as string | undefined) || (raw.endTime as string | undefined);
  const remainingDays = Number(raw.remaining_days ?? raw.remainingDays ?? 0);

  return {
    level: isVIP ? (isPremiumLike ? 'premium' : 'basic') : 'free',
    expiresAt: endTime,
    isActive: isVIP && remainingDays > 0,
    is_vip: isVIP,
    plan_code: planCode,
    plan_name: String(raw.plan_name ?? raw.planName ?? ''),
    start_time:
      (raw.start_time as string | undefined) ||
      (raw.startTime as string | undefined) ||
      null,
    end_time: endTime || null,
    remaining_days: remainingDays,
    auto_renew: Boolean(raw.auto_renew ?? raw.autoRenew ?? false),
    used_queries: Number(raw.used_queries ?? raw.usedQueries ?? 0),
    max_queries: Number(raw.max_queries ?? raw.maxQueries ?? 0),
    used_downloads: Number(raw.used_downloads ?? raw.usedDownloads ?? 0),
    max_downloads: Number(raw.max_downloads ?? raw.maxDownloads ?? 0),
    features:
      typeof raw.features === 'object' && raw.features !== null
        ? (raw.features as Record<string, boolean>)
        : {},
  };
}

function normalizeOrder(raw: Record<string, unknown>): PaymentOrder {
  const orderNo = String(raw.order_no ?? raw.orderNo ?? '');
  const createdAt =
    (raw.created_at as string | undefined) ||
    (raw.createdAt as string | undefined) ||
    new Date().toISOString();
  const paidAt =
    (raw.paid_at as string | undefined) || (raw.paidAt as string | undefined);

  return {
    id: String((raw.id ?? orderNo) || `order_${Date.now()}`),
    order_no: orderNo || String(raw.id ?? ''),
    planId: String(raw.plan_id ?? raw.planId ?? ''),
    planName: String(raw.plan_name ?? raw.planName ?? raw.subject ?? ''),
    amount: Number(raw.amount ?? 0),
    status: String(raw.status ?? 'pending') as PaymentOrder['status'],
    paymentMethod: String(raw.payment_method ?? raw.payment_channel ?? ''),
    payment_channel: String(raw.payment_channel ?? raw.paymentMethod ?? ''),
    subject: String(raw.subject ?? ''),
    createdAt,
    created_at: createdAt,
    paidAt,
    paid_at: paidAt,
    expire_time:
      (raw.expire_time as string | undefined) ||
      (raw.expired_at as string | undefined),
  };
}

export const usePaymentStore = defineStore('payment', () => {
  const membershipStatus = ref<MembershipStatus>(getDefaultMembershipStatus());
  const plans = ref<MembershipPlan[]>([...DEFAULT_PLANS]);
  const orders = ref<PaymentOrder[]>([]);
  const loading = ref(false);

  function syncFromStorage() {
    membershipStatus.value = loadFromStorage(
      MEMBERSHIP_KEY,
      getDefaultMembershipStatus()
    );
    orders.value = loadFromStorage(ORDERS_KEY, [] as PaymentOrder[]);
  }

  async function fetchMembershipStatus(): Promise<MembershipStatus> {
    loading.value = true;
    try {
      try {
        const response = (await api.get(
          '/api/v1/payments/membership/status'
        )) as
          | {
              success: boolean;
              data: Record<string, unknown>;
              message?: string;
            }
          | Record<string, unknown>;
        const raw = isWrappedResponse<Record<string, unknown>>(response)
          ? response.data
          : response;
        const nextStatus = normalizeMembershipStatus(raw);
        membershipStatus.value = nextStatus;
        saveToStorage(MEMBERSHIP_KEY, nextStatus);
        return membershipStatus.value;
      } catch {
        // ignore and fallback to local cache
      }

      syncFromStorage();
      return membershipStatus.value;
    } finally {
      loading.value = false;
    }
  }

  async function getMembershipStatus(): Promise<MembershipStatus> {
    return fetchMembershipStatus();
  }

  async function fetchPlans(): Promise<MembershipPlan[]> {
    loading.value = true;
    try {
      try {
        const response = (await api.get(
          '/api/v1/payments/membership/plans'
        )) as
          | { success: boolean; data: unknown[]; message?: string }
          | unknown[];

        const rawPlans = isWrappedResponse<unknown[]>(response)
          ? response.data
          : response;
        if (Array.isArray(rawPlans) && rawPlans.length > 0) {
          plans.value = rawPlans.map((item) =>
            normalizePlan(item as Record<string, unknown>)
          );
          return plans.value;
        }
      } catch {
        // ignore and fallback to defaults
      }

      plans.value = [...DEFAULT_PLANS];
      return plans.value;
    } finally {
      loading.value = false;
    }
  }

  async function getMembershipPlans(): Promise<MembershipPlan[]> {
    return fetchPlans();
  }

  async function fetchOrders(): Promise<PaymentOrder[]> {
    loading.value = true;
    try {
      try {
        const response = (await api.get('/api/v1/payments', {
          page: 1,
          limit: 50,
        })) as
          | { payments?: unknown[]; total?: number }
          | {
              success: boolean;
              data: { payments?: unknown[]; total?: number };
            };
        const rawPayments = isWrappedResponse<{ payments?: unknown[] }>(
          response
        )
          ? response.data?.payments
          : (response as { payments?: unknown[] }).payments;

        if (Array.isArray(rawPayments)) {
          const nextOrders = rawPayments.map((item) =>
            normalizeOrder(item as Record<string, unknown>)
          );
          orders.value = nextOrders;
          saveToStorage(ORDERS_KEY, nextOrders);
          return orders.value;
        }
      } catch {
        // ignore and fallback to local cache
      }

      syncFromStorage();
      return orders.value;
    } finally {
      loading.value = false;
    }
  }

  async function getOrderList(
    params?: OrderListParams
  ): Promise<OrderListResponse> {
    loading.value = true;
    try {
      await fetchOrders();

      let filtered = [...orders.value];
      if (params?.status) {
        filtered = filtered.filter((item) => item.status === params.status);
      }
      if (params?.start_time) {
        filtered = filtered.filter(
          (item) =>
            new Date(item.createdAt) >= new Date(params.start_time as string)
        );
      }
      if (params?.end_time) {
        filtered = filtered.filter(
          (item) =>
            new Date(item.createdAt) <= new Date(params.end_time as string)
        );
      }

      const page = params?.page || 1;
      const pageSize = params?.page_size || 10;
      const start = (page - 1) * pageSize;
      const paged = filtered.slice(start, start + pageSize);

      return { orders: paged, total: filtered.length };
    } finally {
      loading.value = false;
    }
  }

  async function createOrder(params: string | CreateOrderParams) {
    loading.value = true;
    try {
      syncFromStorage();
      const planCode = typeof params === 'string' ? params : params.plan_code;
      const paymentChannel =
        typeof params === 'string' ? 'alipay' : params.payment_channel;
      const plan = plans.value.find(
        (item) => item.plan_code === planCode || item.id === planCode
      );
      if (!plan) {
        return { success: false, message: '会员套餐不存在' };
      }

      const now = new Date();
      const expiresAt = new Date(now);
      expiresAt.setDate(now.getDate() + (plan.duration_days || 30));

      const order: PaymentOrder = {
        id: `order_${Date.now()}`,
        order_no: `NO${Date.now()}`,
        planId: plan.id,
        planName: plan.name,
        subject: `${plan.name}开通`,
        amount: plan.price,
        status: 'paid',
        paymentMethod: paymentChannel,
        payment_channel: paymentChannel,
        createdAt: now.toISOString(),
        created_at: now.toISOString(),
        paidAt: now.toISOString(),
        paid_at: now.toISOString(),
      };

      orders.value = [order, ...orders.value];
      saveToStorage(ORDERS_KEY, orders.value);

      membershipStatus.value = {
        level: mapPlanToLevel(planCode),
        expiresAt: expiresAt.toISOString(),
        isActive: true,
        is_vip: true,
        plan_code: planCode,
        plan_name: plan.name,
        start_time: now.toISOString(),
        end_time: expiresAt.toISOString(),
        remaining_days: plan.duration_days || 30,
        auto_renew: false,
        used_queries: 0,
        max_queries: plan.max_queries || 0,
        used_downloads: 0,
        max_downloads: plan.max_downloads || 0,
        features: Array.isArray(plan.features)
          ? Object.fromEntries(plan.features.map((item) => [item, true]))
          : plan.features,
      };
      saveToStorage(MEMBERSHIP_KEY, membershipStatus.value);

      return { success: true, orderId: order.order_no };
    } finally {
      loading.value = false;
    }
  }

  async function cancelOrder(orderId: string) {
    loading.value = true;
    try {
      syncFromStorage();
      orders.value = orders.value.map((item) =>
        item.order_no === orderId || item.id === orderId
          ? { ...item, status: 'cancelled' }
          : item
      );
      saveToStorage(ORDERS_KEY, orders.value);
      return { success: true };
    } finally {
      loading.value = false;
    }
  }

  async function cancelMembership() {
    loading.value = true;
    try {
      membershipStatus.value = getDefaultMembershipStatus();
      saveToStorage(MEMBERSHIP_KEY, membershipStatus.value);
      return { success: true };
    } finally {
      loading.value = false;
    }
  }

  async function getInvoice(orderId: string) {
    loading.value = true;
    try {
      const order = orders.value.find(
        (item) => item.order_no === orderId || item.id === orderId
      );
      return {
        success: true,
        url: order ? `invoice://${order.order_no}` : '',
      };
    } finally {
      loading.value = false;
    }
  }

  syncFromStorage();

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
