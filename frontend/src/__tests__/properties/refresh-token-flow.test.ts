/**
 * 契约测试：refresh_token 链路闭环
 *
 * 验证：
 *   1. authEvents 模块的发布/订阅契约（运行时）
 *   2. api-client.handleAuthFailure 通过事件而非直接操作 localStorage（静态）
 *   3. user store 订阅 force-logout 事件并定义 clearAuthState（静态）
 *   4. router guard 通过 store 判定登录态而非裸读 localStorage（静态）
 *
 * 阶段0 高危项：登录未存 refresh_token / 401 不强制下线 / 登出残留状态。
 * 这些点合在一起的根因是 api-client、store、router 三方对认证状态各自为政；
 * 修复策略是事件总线 + store 单一信源。
 */
import { describe, it, expect } from 'vitest';
import * as fs from 'fs';
import * as path from 'path';
import { authEvents } from '@/utils/auth-events';

const repoFile = (rel: string): string =>
  fs.readFileSync(path.resolve(__dirname, '../..', rel), 'utf-8');

describe('refresh_token 链路闭环', () => {
  describe('authEvents 运行时契约', () => {
    it('emit 触发同名 on 订阅者', () => {
      let received: { reason: string; redirect?: string } | null = null;
      const off = authEvents.on('force-logout', (detail) => {
        received = detail;
      });

      authEvents.emit('force-logout', {
        reason: 'refresh-failed',
        redirect: '/profile',
      });

      expect(received).not.toBeNull();
      expect(received!.reason).toBe('refresh-failed');
      expect(received!.redirect).toBe('/profile');

      off();
    });

    it('取消订阅后不再收到事件', () => {
      let count = 0;
      const off = authEvents.on('force-logout', () => {
        count += 1;
      });

      authEvents.emit('force-logout', { reason: 'manual' });
      off();
      authEvents.emit('force-logout', { reason: 'manual' });

      expect(count).toBe(1);
    });

    it('多订阅者并行接收同一事件', () => {
      const calls: string[] = [];
      const off1 = authEvents.on('force-logout', () => calls.push('a'));
      const off2 = authEvents.on('force-logout', () => calls.push('b'));

      authEvents.emit('force-logout', { reason: 'token-expired' });

      expect(calls).toEqual(['a', 'b']);
      off1();
      off2();
    });
  });

  describe('api-client 静态契约', () => {
    const content = repoFile('api/api-client.ts');

    it('handleAuthFailure 通过 authEvents.emit 解耦，不再直接操作业务状态', () => {
      expect(content).toContain("authEvents.emit('force-logout'");
      // 反向：handleAuthFailure 函数体不应直接 router.push 或直接 removeItem
      const handlerMatch = content.match(
        /private handleAuthFailure\(\)[\s\S]*?\n {2}\}/
      );
      expect(handlerMatch).not.toBeNull();
      const body = handlerMatch![0];
      expect(body).not.toMatch(/localStorage\.removeItem/);
      expect(body).not.toMatch(/router\.push/);
    });

    it('仍保留 refresh-token 拦截器主循环', () => {
      expect(content).toContain('refreshToken');
      expect(content).toContain('REFRESH_TOKEN_KEY');
      expect(content).toContain('interceptors.response.use');
    });
  });

  describe('user store 静态契约', () => {
    const content = repoFile('stores/user.ts');

    it('订阅 force-logout 并定义统一清理入口', () => {
      expect(content).toMatch(/authEvents\.on\(\s*['"]force-logout['"]/);
      expect(content).toContain('clearAuthState');
    });

    it('login 持久化 refreshToken（如后端返回）', () => {
      expect(content).toMatch(/REFRESH_TOKEN_KEY[\s\S]{0,200}refreshToken/);
    });

    it('logout 触发 clearAuthState 统一清理', () => {
      const logoutMatch = content.match(
        /const logout = async[\s\S]*?\n {2}\};/
      );
      expect(logoutMatch).not.toBeNull();
      expect(logoutMatch![0]).toContain('clearAuthState');
    });
  });

  describe('router guard 静态契约', () => {
    const content = repoFile('router/index.ts');

    it('requiresAuth 分支以 store.isLoggedIn 为单一信源', () => {
      expect(content).toContain('useUserStore');
      expect(content).toContain('isLoggedIn');
      // 不应再裸读 localStorage 判定登录
      expect(content).not.toMatch(
        /localStorage\.getItem\(['"]auth_token['"]\)/
      );
    });
  });
});
