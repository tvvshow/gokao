import { ElMessage, ElNotification } from 'element-plus';
import type { App } from 'vue';

// Error types for categorization
export enum ErrorType {
  NETWORK = 'network',
  AUTH = 'auth',
  VALIDATION = 'validation',
  SERVER = 'server',
  CLIENT = 'client',
  UNKNOWN = 'unknown',
}

// Error severity levels
export enum ErrorSeverity {
  INFO = 'info',
  WARNING = 'warning',
  ERROR = 'error',
  CRITICAL = 'critical',
}

// Structured error interface
export interface AppError {
  type: ErrorType;
  severity: ErrorSeverity;
  message: string;
  code?: string;
  details?: string;
  timestamp: Date;
  context?: Record<string, unknown>;
}

// Error messages mapping
const ERROR_MESSAGES: Record<string, string> = {
  // Network errors
  NETWORK_ERROR: '网络连接失败，请检查网络设置',
  TIMEOUT_ERROR: '请求超时，请稍后重试',
  CORS_ERROR: '跨域请求被拒绝',

  // Auth errors
  UNAUTHORIZED: '登录已过期，请重新登录',
  FORBIDDEN: '权限不足，无法访问',
  TOKEN_EXPIRED: '登录凭证已过期',

  // Server errors
  SERVER_ERROR: '服务器内部错误，请稍后重试',
  SERVICE_UNAVAILABLE: '服务暂时不可用',
  BAD_GATEWAY: '网关错误',

  // Validation errors
  VALIDATION_ERROR: '输入数据验证失败',
  INVALID_INPUT: '输入格式不正确',

  // Default
  UNKNOWN_ERROR: '发生未知错误，请稍后重试',
};

// Get user-friendly error message
export function getErrorMessage(error: unknown): string {
  if (typeof error === 'string') {
    return error;
  }

  if (error instanceof Error) {
    // Check for known error codes
    const errorCode = (error as Error & { code?: string }).code;
    if (errorCode && ERROR_MESSAGES[errorCode]) {
      return ERROR_MESSAGES[errorCode];
    }
    return error.message || ERROR_MESSAGES.UNKNOWN_ERROR;
  }

  if (typeof error === 'object' && error !== null) {
    const errorObj = error as Record<string, unknown>;
    if (typeof errorObj.message === 'string') {
      return errorObj.message;
    }
    if (typeof errorObj.msg === 'string') {
      return errorObj.msg;
    }
  }

  return ERROR_MESSAGES.UNKNOWN_ERROR;
}

// Categorize error type
export function categorizeError(error: unknown): ErrorType {
  if (error instanceof TypeError) {
    return ErrorType.CLIENT;
  }

  if (error instanceof Error) {
    const message = error.message.toLowerCase();
    const errorObj = error as Error & { code?: string; status?: number };

    // Network errors
    if (
      message.includes('network') ||
      message.includes('fetch') ||
      errorObj.code === 'ECONNABORTED'
    ) {
      return ErrorType.NETWORK;
    }

    // Auth errors
    if (
      errorObj.status === 401 ||
      errorObj.status === 403 ||
      message.includes('unauthorized') ||
      message.includes('forbidden')
    ) {
      return ErrorType.AUTH;
    }

    // Server errors
    if (errorObj.status && errorObj.status >= 500) {
      return ErrorType.SERVER;
    }

    // Validation errors
    if (
      errorObj.status === 400 ||
      errorObj.status === 422 ||
      message.includes('validation')
    ) {
      return ErrorType.VALIDATION;
    }
  }

  return ErrorType.UNKNOWN;
}

// Show error notification based on severity
export function showError(
  message: string,
  severity: ErrorSeverity = ErrorSeverity.ERROR,
  duration: number = 3000
): void {
  switch (severity) {
    case ErrorSeverity.INFO:
      ElMessage.info({ message, duration });
      break;
    case ErrorSeverity.WARNING:
      ElMessage.warning({ message, duration });
      break;
    case ErrorSeverity.ERROR:
      ElMessage.error({ message, duration });
      break;
    case ErrorSeverity.CRITICAL:
      ElNotification.error({
        title: '严重错误',
        message,
        duration: 0, // Don't auto-close critical errors
      });
      break;
  }
}

// Handle error with unified logic
export function handleError(error: unknown, context?: string): AppError {
  const message = getErrorMessage(error);
  const type = categorizeError(error);

  // Determine severity based on error type
  let severity = ErrorSeverity.ERROR;
  if (type === ErrorType.VALIDATION) {
    severity = ErrorSeverity.WARNING;
  } else if (type === ErrorType.AUTH) {
    severity = ErrorSeverity.WARNING;
  } else if (type === ErrorType.SERVER) {
    severity = ErrorSeverity.CRITICAL;
  }

  const appError: AppError = {
    type,
    severity,
    message,
    timestamp: new Date(),
    context: context ? { source: context } : undefined,
  };

  // Show error to user
  showError(message, severity);

  // Log error for debugging (in development)
  if (import.meta.env.DEV) {
    console.error('[Error Handler]', {
      ...appError,
      originalError: error,
    });
  }

  return appError;
}

// Vue plugin for global error handling
export function setupErrorHandler(app: App): void {
  // Global error handler
  app.config.errorHandler = (error, instance, info) => {
    const componentName = instance?.$options?.name || 'Unknown';
    handleError(error, `Vue Component: ${componentName}, Info: ${info}`);
  };

  // Global warning handler (development only)
  if (import.meta.env.DEV) {
    app.config.warnHandler = (msg, instance, trace) => {
      console.warn('[Vue Warning]', msg, trace);
    };
  }

  // Handle unhandled promise rejections
  window.addEventListener('unhandledrejection', (event) => {
    event.preventDefault();
    handleError(event.reason, 'Unhandled Promise Rejection');
  });

  // Handle global errors
  window.addEventListener('error', (event) => {
    // Ignore script loading errors
    if (event.target && (event.target as HTMLElement).tagName === 'SCRIPT') {
      return;
    }
    handleError(event.error || event.message, 'Global Error');
  });
}

// Export default handler
export default {
  install: setupErrorHandler,
};
