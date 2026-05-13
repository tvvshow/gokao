// L-11 主入口压测脚本：覆盖 auth / data 浏览 / 推荐三条核心路径。
//
// SLA 目标（与 CLAUDE.md §6 对齐）：
//   - API Gateway P95 < 100ms（auth、universities、majors 路径）
//   - Recommendation P95 < 500ms
//   - 错误率 < 1%
//   - 高峰期 10 万并发（生产规模），本脚本默认 200 VU stage，CI 用 50 VU smoke
//
// 用法：
//   # 本地起服后跑全套
//   k6 run --env API_BASE_URL=http://localhost:8080 tests/performance/load-test.js
//
//   # 跑 CI smoke 档（更少 VU、更短时长）
//   k6 run --env API_BASE_URL=$STAGING_API_URL --env PROFILE=smoke tests/performance/load-test.js
//
//   # 仅跑推荐路径
//   k6 run --env API_BASE_URL=... --env SCENARIO=recommend tests/performance/load-test.js

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const BASE = __ENV.API_BASE_URL || 'http://localhost:8080';
const TEST_EMAIL = __ENV.TEST_USER_EMAIL || '';
const TEST_PASSWORD = __ENV.TEST_USER_PASSWORD || '';
const PROFILE = __ENV.PROFILE || 'full';        // full | smoke
const SCENARIO = __ENV.SCENARIO || 'all';        // all | auth | browse | recommend

// 业务级别指标（除 k6 内置 http_req_duration 外另外打）
const recommendDuration = new Trend('recommend_duration', true);
const browseDuration = new Trend('browse_duration', true);
const authDuration = new Trend('auth_duration', true);
const errorRate = new Rate('business_errors');

const profileTable = {
  smoke: {
    auth:       { exec: 'authScenario',      executor: 'constant-vus', vus: 5,  duration: '30s' },
    browse:     { exec: 'browseScenario',    executor: 'constant-vus', vus: 10, duration: '30s' },
    recommend:  { exec: 'recommendScenario', executor: 'constant-vus', vus: 3,  duration: '30s' },
  },
  full: {
    auth: {
      exec: 'authScenario',
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 20 },   // warmup
        { duration: '1m',  target: 50 },   // peak
        { duration: '30s', target: 0 },    // cooldown
      ],
      gracefulRampDown: '15s',
    },
    browse: {
      exec: 'browseScenario',
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '2m',  target: 200 },
        { duration: '30s', target: 0 },
      ],
      gracefulRampDown: '15s',
    },
    recommend: {
      exec: 'recommendScenario',
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 10 },
        { duration: '2m',  target: 50 },   // recommend 是贵接口，限 50 VU
        { duration: '30s', target: 0 },
      ],
      gracefulRampDown: '15s',
    },
  },
};

function buildScenarios() {
  const p = profileTable[PROFILE];
  if (!p) throw new Error(`unknown PROFILE: ${PROFILE}`);
  if (SCENARIO === 'all') return p;
  if (!p[SCENARIO]) throw new Error(`unknown SCENARIO: ${SCENARIO}`);
  return { [SCENARIO]: p[SCENARIO] };
}

export const options = {
  scenarios: buildScenarios(),
  thresholds: {
    // SLA gate：超阈值 k6 退出码非零，CI 标红
    'http_req_failed':                       ['rate<0.01'],
    'business_errors':                       ['rate<0.01'],
    'http_req_duration{scenario:auth}':      ['p(95)<200'],   // 含 bcrypt + DB
    'http_req_duration{scenario:browse}':    ['p(95)<150'],
    'recommend_duration':                    ['p(95)<500'],   // 项目 SLA 硬目标
    'browse_duration':                       ['p(95)<150'],
    'auth_duration':                         ['p(95)<200'],
  },
  noConnectionReuse: false,
  userAgent: 'k6-load-test/L-11',
};

// setup 一次性登录拿 token；recommend / authed-path 共用。
// 没配 TEST_USER_EMAIL 时返回空 token，下游 recommend 场景会 SKIP。
export function setup() {
  if (!TEST_EMAIL || !TEST_PASSWORD) {
    console.warn('TEST_USER_EMAIL/PASSWORD not set; recommend scenario will be skipped');
    return { token: '', authed: false };
  }
  const res = http.post(
    `${BASE}/api/v1/users/auth/login`,
    JSON.stringify({ email: TEST_EMAIL, password: TEST_PASSWORD }),
    { headers: { 'Content-Type': 'application/json' }, tags: { setup: 'true' } },
  );
  if (res.status !== 200) {
    console.error(`setup login failed: ${res.status} ${res.body}`);
    return { token: '', authed: false };
  }
  const body = res.json();
  const token = (body && body.data && (body.data.token || body.data.access_token)) || '';
  return { token, authed: !!token };
}

export function authScenario() {
  // 反复登录场景：模拟登录峰值。每个 VU 独立轮次。
  group('auth login', () => {
    const start = Date.now();
    const res = http.post(
      `${BASE}/api/v1/users/auth/login`,
      JSON.stringify({
        // 用错口令测限流路径——不应让正确账号被锁；后端有 bcrypt cost 决定耗时
        email: `loadtest+${__VU}@gaokao.dev`,
        password: 'invalid-password-for-load-test',
      }),
      { headers: { 'Content-Type': 'application/json' } },
    );
    authDuration.add(Date.now() - start);
    const ok = check(res, {
      'auth status 4xx (expected for invalid creds)': (r) => r.status >= 400 && r.status < 500,
    });
    errorRate.add(!ok);
  });
  sleep(0.2 + Math.random() * 0.3);
}

export function browseScenario() {
  // 浏览路径：universities + majors 分页 + 搜索 keyword
  group('browse universities', () => {
    const start = Date.now();
    const page = (__ITER % 5) + 1;
    const res = http.get(
      `${BASE}/api/v1/data/universities?page=${page}&page_size=20`,
      { tags: { scenario: 'browse' } },
    );
    browseDuration.add(Date.now() - start);
    const ok = check(res, {
      'universities 200': (r) => r.status === 200,
    });
    errorRate.add(!ok);
  });

  group('browse majors', () => {
    const start = Date.now();
    const keywords = ['计算机', '金融', '医学', '工程', '法律'];
    const kw = keywords[__ITER % keywords.length];
    const res = http.get(
      `${BASE}/api/v1/data/majors?keyword=${encodeURIComponent(kw)}&page=1&page_size=20`,
      { tags: { scenario: 'browse' } },
    );
    browseDuration.add(Date.now() - start);
    const ok = check(res, {
      'majors 200': (r) => r.status === 200,
    });
    errorRate.add(!ok);
  });

  sleep(0.5 + Math.random() * 0.5);
}

export function recommendScenario(data) {
  if (!data.authed) {
    // 无 token → 该场景静默跳过，不计错。
    sleep(1);
    return;
  }
  group('recommendations', () => {
    const start = Date.now();
    const res = http.post(
      `${BASE}/api/v1/recommendations`,
      JSON.stringify({
        score: 580 + Math.floor(Math.random() * 60),
        rank: 30000 + Math.floor(Math.random() * 10000),
        province: '北京',
        category: 'science',
        preferences: {
          location_pref: ['北京', '上海'],
          major_pref: ['计算机', '电子'],
          strategy: 'balanced',
        },
      }),
      {
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${data.token}`,
        },
        tags: { scenario: 'recommend' },
      },
    );
    recommendDuration.add(Date.now() - start);
    const ok = check(res, {
      'recommend 200': (r) => r.status === 200,
      'recommend has results': (r) => {
        try {
          const j = r.json();
          return j && j.data && Array.isArray(j.data.recommendations);
        } catch (_) {
          return false;
        }
      },
    });
    errorRate.add(!ok);
  });
  sleep(1 + Math.random() * 1);
}

// k6 默认 default()：本脚本完全由 scenarios 驱动，不用 default。
export default function () {
  // no-op；scenarios 接管
}
