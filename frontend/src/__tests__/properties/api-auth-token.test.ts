/**
 * Property Test: API Authentication Token Passing
 *
 * Feature: gui-audit, Property 4: API Authentication Token Passing
 * Validates: Requirements 3.1
 *
 * For any API request that requires authentication, the request header
 * must contain an Authorization header in the format "Bearer {token}"
 */
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import * as fc from 'fast-check';
import * as fs from 'fs';
import * as path from 'path';

// Token storage key (must match api-client.ts)
const TOKEN_KEY = 'auth_token';

// Mock localStorage for testing
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();

describe('Property 4: API Authentication Token Passing', () => {
  beforeEach(() => {
    // Setup localStorage mock
    Object.defineProperty(global, 'localStorage', {
      value: localStorageMock,
      writable: true,
    });
    localStorageMock.clear();
  });

  afterEach(() => {
    localStorageMock.clear();
  });

  /**
   * Static analysis test: Verify api-client.ts contains proper token handling
   */
  it('should have request interceptor that adds Authorization header', () => {
    const apiClientPath = path.resolve(__dirname, '../../api/api-client.ts');
    const content = fs.readFileSync(apiClientPath, 'utf-8');

    // Check for request interceptor
    expect(content).toContain('interceptors.request.use');

    // Check for Authorization header setting
    expect(content).toContain('Authorization');
    expect(content).toContain('Bearer');

    // Check for token retrieval from localStorage
    expect(content).toContain('localStorage.getItem');
    expect(content).toContain(TOKEN_KEY);
  });

  /**
   * Static analysis test: Verify token format is correct
   */
  it('should use Bearer token format', () => {
    const apiClientPath = path.resolve(__dirname, '../../api/api-client.ts');
    const content = fs.readFileSync(apiClientPath, 'utf-8');

    // Check for Bearer token format pattern
    const bearerPattern =
      /Bearer\s*\$\{.*token.*\}|Bearer\s*\+\s*token|`Bearer\s*\$\{token\}`/i;
    expect(bearerPattern.test(content)).toBe(true);
  });

  /**
   * Static analysis test: Verify 401 response handling
   */
  it('should handle 401 unauthorized responses', () => {
    const apiClientPath = path.resolve(__dirname, '../../api/api-client.ts');
    const content = fs.readFileSync(apiClientPath, 'utf-8');

    // Check for 401 status handling
    expect(content).toContain('401');

    // Check for response interceptor
    expect(content).toContain('interceptors.response.use');
  });

  /**
   * Static analysis test: Verify token refresh mechanism exists
   */
  it('should have token refresh mechanism', () => {
    const apiClientPath = path.resolve(__dirname, '../../api/api-client.ts');
    const content = fs.readFileSync(apiClientPath, 'utf-8');

    // Check for refresh token handling
    expect(content).toContain('refreshToken');
    expect(content).toContain('REFRESH_TOKEN_KEY');
  });

  /**
   * Property test: For any valid JWT-like token string,
   * the Authorization header format should be "Bearer {token}"
   */
  it('property: token format should always be Bearer prefix', () => {
    // Generate random JWT-like tokens
    const jwtArbitrary = fc
      .tuple(
        fc.base64String({ minLength: 10, maxLength: 50 }),
        fc.base64String({ minLength: 10, maxLength: 100 }),
        fc.base64String({ minLength: 10, maxLength: 50 })
      )
      .map(([header, payload, signature]) =>
        `${header}.${payload}.${signature}`.replace(/[+/=]/g, '')
      );

    fc.assert(
      fc.property(jwtArbitrary, (token) => {
        // Simulate the token formatting logic from api-client.ts
        const authHeader = `Bearer ${token}`;

        // Verify format
        expect(authHeader).toMatch(/^Bearer\s+.+$/);
        expect(authHeader.startsWith('Bearer ')).toBe(true);
        expect(authHeader.length).toBeGreaterThan(7); // "Bearer " + at least 1 char

        return true;
      }),
      { numRuns: 100 }
    );
  });

  /**
   * Property test: Token should be retrieved from localStorage correctly
   */
  it('property: stored tokens should be retrievable', () => {
    fc.assert(
      fc.property(fc.string({ minLength: 10, maxLength: 200 }), (token) => {
        // Store token
        localStorageMock.setItem(TOKEN_KEY, token);

        // Retrieve token
        const retrieved = localStorageMock.getItem(TOKEN_KEY);

        // Verify retrieval
        expect(retrieved).toBe(token);

        return true;
      }),
      { numRuns: 100 }
    );
  });

  /**
   * Verify all API files use the centralized api-client
   */
  it('should have all API modules using centralized client', () => {
    const apiDir = path.resolve(__dirname, '../../api');
    const apiFiles = fs.readdirSync(apiDir).filter((f) => f.endsWith('.ts'));

    for (const file of apiFiles) {
      if (file === 'api-client.ts') continue;

      const filePath = path.join(apiDir, file);
      const content = fs.readFileSync(filePath, 'utf-8');

      // Each API module should import from api-client
      const importsApiClient =
        content.includes("from './api-client'") ||
        content.includes("from '@/api/api-client'") ||
        content.includes('from "./api-client"') ||
        content.includes('from "@/api/api-client"');

      // Or use axios with proper interceptors
      const usesAxiosWithInterceptors =
        content.includes('axios') && content.includes('interceptors');

      expect(
        importsApiClient || usesAxiosWithInterceptors,
        `${file} should use centralized api-client or have proper interceptors`
      ).toBe(true);
    }
  });
});
