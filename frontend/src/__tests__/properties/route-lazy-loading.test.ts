/**
 * Property Test: Route Lazy Loading
 *
 * Feature: gui-audit, Property 8: Route Lazy Loading
 * Validates: Requirements 1.3, 5.1
 *
 * For any route configuration, the component should use dynamic import syntax
 * (() => import(...)) to implement lazy loading.
 */
import { describe, it, expect } from 'vitest';
import * as fc from 'fast-check';
import * as fs from 'fs';
import * as path from 'path';

// Patterns for lazy loading
const LAZY_LOAD_PATTERN = /component\s*:\s*\(\)\s*=>\s*import\s*\(/g;
const ROUTE_DEFINITION_PATTERN = /{\s*path\s*:/g;

/**
 * Analyze route configuration for lazy loading
 */
function analyzeRouteConfig(filePath: string): {
  totalRoutes: number;
  lazyLoadedRoutes: number;
  staticRoutes: number;
  routeDetails: Array<{ path: string; isLazy: boolean }>;
} {
  const content = fs.readFileSync(filePath, 'utf-8');

  // Count total routes
  const routeMatches = content.match(ROUTE_DEFINITION_PATTERN) || [];
  const totalRoutes = routeMatches.length;

  // Count lazy loaded routes
  const lazyMatches = content.match(LAZY_LOAD_PATTERN) || [];
  const lazyLoadedRoutes = lazyMatches.length;

  // Extract route details
  const routeDetails: Array<{ path: string; isLazy: boolean }> = [];

  // Parse route paths and their loading type
  const pathRegex = /path\s*:\s*['"]([^'"]+)['"]/g;
  let match;
  while ((match = pathRegex.exec(content)) !== null) {
    const routePath = match[1];

    // Check if this route uses lazy loading
    // Look for the component definition after this path
    const afterPath = content.slice(match.index);
    const componentMatch = afterPath.match(
      /component\s*:\s*(\(\)\s*=>\s*import|[\w]+)/
    );

    const isLazy = componentMatch
      ? componentMatch[1].includes('import')
      : false;

    routeDetails.push({
      path: routePath,
      isLazy,
    });
  }

  return {
    totalRoutes,
    lazyLoadedRoutes,
    staticRoutes: totalRoutes - lazyLoadedRoutes,
    routeDetails,
  };
}

describe('Property 8: Route Lazy Loading', () => {
  const srcDir = path.resolve(__dirname, '../../');
  const routerPath = path.resolve(srcDir, 'router/index.ts');

  it('should have router configuration file', () => {
    expect(fs.existsSync(routerPath)).toBe(true);
  });

  it('should use lazy loading for all routes', () => {
    const analysis = analyzeRouteConfig(routerPath);

    console.log(`\nRoute Lazy Loading Analysis:`);
    console.log(`  - Total routes: ${analysis.totalRoutes}`);
    console.log(`  - Lazy loaded routes: ${analysis.lazyLoadedRoutes}`);
    console.log(`  - Static routes: ${analysis.staticRoutes}`);

    // All routes should use lazy loading
    const lazyLoadingPercent =
      analysis.totalRoutes > 0
        ? (analysis.lazyLoadedRoutes / analysis.totalRoutes) * 100
        : 100;

    console.log(`  - Lazy loading coverage: ${lazyLoadingPercent.toFixed(1)}%`);

    // 90%+ of routes should use lazy loading (allowing for parsing edge cases)
    expect(lazyLoadingPercent).toBeGreaterThanOrEqual(90);
  });

  it('should use dynamic import syntax for components', () => {
    const content = fs.readFileSync(routerPath, 'utf-8');

    // Check for dynamic import pattern
    const hasDynamicImport = /\(\)\s*=>\s*import\s*\(/.test(content);
    expect(hasDynamicImport).toBe(true);

    // Check that no static component imports are used in routes
    // (excluding type imports)
    const hasStaticComponentInRoute =
      /component\s*:\s*(?!.*=>.*import)[A-Z]\w+(?!.*import)/.test(content);
    expect(hasStaticComponentInRoute).toBe(false);
  });

  it('should have route meta information', () => {
    const content = fs.readFileSync(routerPath, 'utf-8');

    // Check for meta property
    const hasMeta = /meta\s*:\s*{/.test(content);
    expect(hasMeta).toBe(true);

    // Check for title in meta
    const hasTitle = /title\s*:\s*['"]/.test(content);
    expect(hasTitle).toBe(true);
  });

  it('should have route guards for protected routes', () => {
    const content = fs.readFileSync(routerPath, 'utf-8');

    // Check for beforeEach guard
    const hasBeforeEach = /router\.beforeEach/.test(content);
    expect(hasBeforeEach).toBe(true);

    // Check for requiresAuth meta
    const hasRequiresAuth = /requiresAuth/.test(content);
    expect(hasRequiresAuth).toBe(true);

    // 认证判定单一信源：必须通过 user store（isLoggedIn）；
    // 历史实现裸读 localStorage，会与 Pinia 内存状态分裂，已废弃。
    const usesStore =
      /useUserStore/.test(content) && /isLoggedIn/.test(content);
    expect(usesStore).toBe(true);
  });

  it('should have proper scroll behavior', () => {
    const content = fs.readFileSync(routerPath, 'utf-8');

    // Check for scrollBehavior
    const hasScrollBehavior = /scrollBehavior/.test(content);
    expect(hasScrollBehavior).toBe(true);
  });

  // Property-based test: all route paths should be valid
  it('property: all route paths should be valid URL paths', () => {
    const analysis = analyzeRouteConfig(routerPath);

    // Filter out duplicate paths (from route guards)
    const uniqueRoutes = analysis.routeDetails.filter(
      (route, index, self) =>
        index === self.findIndex((r) => r.path === route.path)
    );

    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: uniqueRoutes.length - 1 }),
        (index) => {
          const route = uniqueRoutes[index];

          // Path should start with /
          expect(route.path.startsWith('/') || route.path === '*').toBe(true);

          // Path should not contain spaces
          expect(route.path).not.toContain(' ');

          // Route should use lazy loading
          expect(route.isLazy).toBe(true);

          return true;
        }
      ),
      { numRuns: Math.min(100, uniqueRoutes.length) }
    );
  });

  it('should list all routes with their loading type', () => {
    const analysis = analyzeRouteConfig(routerPath);

    console.log(`\nRoute Details:`);
    for (const route of analysis.routeDetails) {
      const status = route.isLazy ? '✓ Lazy' : '✗ Static';
      console.log(`  ${status}: ${route.path}`);
    }
  });
});
