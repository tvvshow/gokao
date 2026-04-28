export type WrappedResponse<T> = {
  success: boolean;
  data: T;
  message?: string;
};

export function isWrappedResponse<T>(
  value: unknown
): value is WrappedResponse<T> {
  return (
    typeof value === 'object' &&
    value !== null &&
    'success' in value &&
    'data' in value
  );
}

export function normalizeMessageResponse(
  response: unknown,
  fallbackMessage: string
): { success: boolean; message?: string } {
  if (isWrappedResponse<unknown>(response)) {
    return {
      success: response.success,
      message: response.message,
    };
  }

  const raw = (response || {}) as {
    message?: string;
  };
  return {
    success: true,
    message: raw.message || fallbackMessage,
  };
}

export function unwrapDataOrSelf<T>(response: WrappedResponse<T> | T): T {
  if (isWrappedResponse<T>(response)) {
    return response.data;
  }
  return response;
}
