import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: Number(__ENV.K6_VUS || 10),
  duration: __ENV.K6_DURATION || '30s',
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

const targets = [
  { name: 'api-gateway', baseUrl: __ENV.API_GATEWAY_BASE_URL || 'http://localhost:8080', path: '/health' },
  { name: 'user-service', baseUrl: __ENV.USER_SERVICE_BASE_URL || 'http://localhost:8083', path: '/health' },
  { name: 'data-service', baseUrl: __ENV.DATA_SERVICE_BASE_URL || 'http://localhost:8082', path: '/health' },
  { name: 'recommendation-service', baseUrl: __ENV.RECOMMENDATION_SERVICE_BASE_URL || 'http://localhost:8084', path: '/health' },
  { name: 'payment-service', baseUrl: __ENV.PAYMENT_SERVICE_BASE_URL || 'http://localhost:8085', path: '/health' },
];

export default function () {
  for (const target of targets) {
    const res = http.get(`${target.baseUrl}${target.path}`, {
      tags: { service: target.name, path: target.path },
      headers: {
        'X-Request-ID': `k6-${target.name}-${__VU}-${__ITER}`,
        'X-Trace-ID': `k6-trace-${target.name}-${__VU}-${__ITER}`,
      },
    });

    check(res, {
      [`${target.name} status is 200`]: (r) => r.status === 200,
    });
  }

  sleep(1);
}
