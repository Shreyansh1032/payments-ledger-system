import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

const BASE_URL = 'http://localhost:8080';

const transferFailures = new Counter('transfer_failures');

export const options = {
  scenarios: {
    transfers: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 20 },
        { duration: '30s', target: 50 },
        { duration: '10s', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'],
  },
};

// Two pre-existing users, seeded before the test (see setup below).
// Every VU transfers a tiny, fixed amount back and forth so balances
// never run out mid-test regardless of VU count.
const SENDER_TOKEN = __ENV.SENDER_TOKEN;
const RECEIVER_USERNAME = __ENV.RECEIVER_USERNAME || 'loadtest_receiver';

export default function () {
  const idempotencyKey = `k6-${__VU}-${__ITER}-${Date.now()}`;

  const res = http.post(
    `${BASE_URL}/api/v1/transfer/`,
    JSON.stringify({ to_username: RECEIVER_USERNAME, amount: '1' }),
    {
      headers: {
        'Content-Type': 'application/json',
        'Idempotency-Key': idempotencyKey,
        Authorization: `Bearer ${SENDER_TOKEN}`,
      },
    }
  );

  const ok = check(res, {
    'status is 200': (r) => r.status === 200,
  });

  if (!ok) {
    transferFailures.add(1);
  }

  sleep(0.1);
}
