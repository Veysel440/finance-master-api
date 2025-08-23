import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = { vus: 5, duration: '10s' };

export default function () {
    const res = http.get('http://127.0.0.1:8080/health');
    check(res, { 'status 200': (r) => r.status === 200 });
    sleep(1);
}