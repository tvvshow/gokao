// Payment related types

export interface MembershipStatus {
  level: 'free' | 'basic' | 'premium';
  expiresAt?: string;
  isActive: boolean;
  // Extended properties for component compatibility
  is_vip?: boolean;
  plan_code?: string;
  plan_name?: string;
  start_time?: string | null;
  end_time?: string | null;
  remaining_days?: number;
  auto_renew?: boolean;
  used_queries?: number;
  max_queries?: number;
  used_downloads?: number;
  max_downloads?: number;
  features?: Record<string, boolean>;
}

export interface MembershipPlan {
  id: string;
  name: string;
  price: number;
  duration?: string;
  features: string[] | Record<string, boolean>;
  recommended?: boolean;
  // Extended properties
  plan_code?: string;
  duration_days?: number;
  max_queries?: number;
  max_downloads?: number;
  // UI properties
  period?: string;
  featured?: boolean;
  buttonText?: string;
  originalPrice?: number;
}

// Alias for component compatibility
export type MembershipPlanItem = MembershipPlan & {
  period?: string;
  featured?: boolean;
  buttonText?: string;
  originalPrice?: number;
};

export interface PaymentOrder {
  id: string;
  planId: string;
  planName: string;
  amount: number;
  status: 'pending' | 'paid' | 'cancelled' | 'refunded';
  paymentMethod?: string;
  createdAt: string;
  paidAt?: string;
  // Extended properties
  order_no?: string;
  subject?: string;
  payment_channel?: string;
  created_at?: string;
  paid_at?: string;
  expire_time?: string;
}

export interface PaymentResult {
  success: boolean;
  orderId?: string;
  message?: string;
}

export interface CreateOrderParams {
  plan_code: string;
  payment_channel: string;
}
