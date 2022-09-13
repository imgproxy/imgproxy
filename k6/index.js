import { randomSeed, check } from 'k6';
import http from 'k6/http';
import { SharedArray } from 'k6/data';
import exec from 'k6/execution';

export let options = {
  discardResponseBodies: true,
  noConnectionReuse: false,

  vus: 20,
  iterations: 10000,
};

randomSeed(42)

const urls = new SharedArray('urls', function () {
  const urls_path = __ENV.URLS_PATH || './urls.json'
  let data = JSON.parse(open(urls_path));

  const groups = (__ENV.URL_GROUPS || "").split(",").filter((g) => g != "")
  if (groups.length > 0) {
    data = data.filter((d) => groups.includes(d.group))
  }

  const url_prefix = __ENV.URL_PREFIX || "http://localhost:8082/unsafe"

  let unshuffled = [];
  data.forEach((e) => {
    let url = url_prefix + e.url
    let weight = e.weight || 1

    for (var i = 0; i < weight; i++) {
      unshuffled.push(url)
    }
  })

  let shuffled = unshuffled
    .map(value => ({ value, sort: Math.random() }))
    .sort((a, b) => a.sort - b.sort)
    .map(({ value }) => value)

  return shuffled;
});

if (urls.length == 0) {
  throw "URLs list is empty"
}

export default function() {
  const res = http.get(urls[exec.scenario.iterationInTest % urls.length]);
  check(res, {
    'is status 200': (r) => r.status === 200,
  });
}
