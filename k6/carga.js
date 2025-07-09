import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  vus: 50, // 50 usuarios simultáneos
  duration: '20s',
};

export default function () {
  let res = http.get('http://localhost:8080/');
  check(res, { 'status es 200': (r) => r.status === 200 });

  let resLibros = http.get('http://localhost:8080/libros');
  check(resLibros, { 'Libros carga OK': (r) => r.status === 200 });

  sleep(1);
}
