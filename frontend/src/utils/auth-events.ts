// 解耦：api-client 拦截器在认证失败时不直接操作业务状态/路由，
// 而是 emit 强制下线事件；user store 是唯一负责清理 state + localStorage + 跳转的位置。
// 这样 router guard、组件都只需读 store.isLoggedIn，认证状态单一信源。

export type ForceLogoutReason = 'token-expired' | 'refresh-failed' | 'manual';

export interface ForceLogoutDetail {
  reason: ForceLogoutReason;
  // 触发下线时所处页面（用于登录后回跳）；空表示不需回跳。
  redirect?: string;
}

type EventMap = {
  'force-logout': ForceLogoutDetail;
};

type Handler<K extends keyof EventMap> = (detail: EventMap[K]) => void;

const target = new EventTarget();

export const authEvents = {
  emit<K extends keyof EventMap>(name: K, detail: EventMap[K]): void {
    target.dispatchEvent(new CustomEvent(name, { detail }));
  },

  on<K extends keyof EventMap>(name: K, handler: Handler<K>): () => void {
    const listener = (event: Event): void => {
      handler((event as CustomEvent<EventMap[K]>).detail);
    };
    target.addEventListener(name, listener);
    return () => target.removeEventListener(name, listener);
  },
};
