import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 50,
    duration: '1m',
    thresholds: { http_req_duration: ['p(95)<400'] },
};

const BASE = 'http://localhost:8080';
const TOKEN = __ENV.AUTH_TOKEN;

export default function () {
    const payload = JSON.stringify({
        type: 'expense',
        amount: Math.floor(Math.random()*100)+1,
        currency: 'TRY',
        categoryId: 1,
        walletId: 1,
        note: 'k6',
        occurredAt: new Date().toISOString(),
    });
    const res = http.post(`${BASE}/v1/transactions`, payload, {
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${TOKEN}` }
    });
    check(res, { '201': r => r.status === 201 });
    sleep(0.2);
}
