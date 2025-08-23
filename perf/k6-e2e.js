import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = { vus: 1, iterations: 1 };

const BASE = __ENV.API_BASE || 'http://127.0.0.1:8080';

export default function () {
    
    const email = `u${Date.now()}@ex.com`;
    let res = http.post(`${BASE}/v1/auth/register`, JSON.stringify({name:"u", email, password:"pass1234"}), { headers: { 'Content-Type':'application/json' }});
    check(res, { 'register 201': r => r.status === 201 });


    res = http.post(`${BASE}/v1/auth/login`, JSON.stringify({email, password:"pass1234", deviceId:"ci", deviceName:"k6"}), { headers: { 'Content-Type':'application/json' }});
    check(res, { 'login 200': r => r.status === 200 });
    const token = res.json('token');

    const h = { headers: { Authorization: `Bearer ${token}`, 'Content-Type':'application/json' } };


    res = http.post(`${BASE}/v1/wallets`, JSON.stringify({name:"Cüzdan", currency:"TRY"}), h);
    check(res, { 'wallet 201': r => r.status === 201 });
    const walletId = res.json('id');


    res = http.post(`${BASE}/v1/categories`, JSON.stringify({name:"Yemek", type:"expense"}), h);
    check(res, { 'cat 201': r => r.status === 201 });
    const catId = res.json('id');


    const txPayload = { type:"expense", amount: 100.5, currency:"TRY", note:"k6", walletId, categoryId: catId, occurredAt: new Date().toISOString() };
    res = http.post(`${BASE}/v1/transactions`, JSON.stringify(txPayload), h);
    check(res, { 'tx 201': r => r.status === 201 });
    const txId = res.json('id');


    res = http.get(`${BASE}/v1/transactions/${txId}`, h);
    check(res, { 'tx get 200': r => r.status === 200 });


    res = http.get(`${BASE}/v1/transactions?page=1&size=10&q=k6`, h);
    check(res, { 'tx list 200': r => r.status === 200 });


    const from = new Date(Date.now()-24*3600*1000).toISOString();
    const to = new Date(Date.now()+24*3600*1000).toISOString();
    res = http.get(`${BASE}/v1/transactions/summary?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`, h);
    check(res, { 'summary 200': r => r.status === 200 });

    sleep(1);
}