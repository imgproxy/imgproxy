import { randomSeed, check } from 'k6';
import http from 'k6/http';
import { SharedArray } from 'k6/data';
import exec from 'k6/execution';
import { Trend } from 'k6/metrics';

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
      unshuffled.push({url, group: e.group})
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

let group_durations = [...new Set(urls.map(url => url.group))]
  .reduce((trends, group) => {
    trends[group] = new Trend(`http_req_duration_${group}`, true);
    return trends;
  }, {});

let group_sizes = [...new Set(urls.map(url => url.group))]
  .reduce((trends, group) => {
    trends[group] = new Trend(`http_res_body_size_${group}`, false);
    return trends;
  }, {});

export default function() {
  const url = urls[exec.scenario.iterationInTest % urls.length]
  const res = http.get(url.url);
  check(res, {
    'is status 200': (r) => r.status === 200,
  });
  group_durations[url.group].add(res.timings.duration);

  const body_size = Math.round(parseInt(res.headers["Content-Length"]) / 10.24) / 100;
  group_sizes[url.group].add(body_size);
}
