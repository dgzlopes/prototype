import http from 'k6/http';
import { sleep } from 'k6';
import { randomString, randomIntBetween } from "https://jslib.k6.io/k6-utils/1.1.0/index.js";

export const options = {
    duration: "1m",
    vus: 10,
    thresholds: {
        http_req_failed: [{ threshold: 'rate<0.01', abortOnFail: true }],   // http errors should be less than 1% 
        http_req_duration: [{ threshold: 'p(99)<1500', abortOnFail: true }], // 99% of requests must complete below 1.5s
    },
};

export default function () {
    var payload = JSON.stringify({
        "cluster": "default",
        "service": "service",
        "id": `service-${randomString(8)}`,
        "envoy_info": {
            "version": randomString(2),
            "state": "LIVE",
            "uptime": randomIntBetween(1,10000).toString(),
            "timestamp": randomIntBetween(100000,999999)
        }
    });
    var params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };
    http.post('http://localhost:10000/api/protod',payload,params);
    sleep(1);
}