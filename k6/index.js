import { randomSeed } from 'k6';
import http from 'k6/http';
import { SharedArray } from 'k6/data';
import exec from 'k6/execution';

const URL_PREFIX = "http://localhost:8082/unsafe"

export let options = {
  discardResponseBodies: true,
  noConnectionReuse: false,

  vus: 20,
  iterations: 10000,
};

randomSeed(42)

const urls = new SharedArray('urls', function () {
  const data = JSON.parse(open('./urls.json'));

  let unshuffled = [];
  data.forEach((e) => {
    let url = URL_PREFIX + e.url
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

export default function() {
  http.get(urls[exec.scenario.iterationInTest % urls.length])
}
