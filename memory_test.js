import streamloader from 'k6/x/streamloader';
import { check } from 'k6';
import { scenario } from 'k6/execution';

const FILE_PATH = 'large_file.txt';

export const options = {
  scenarios: {
    k6_open: {
      exec: 'test_k6_open',
      executor: 'per-vu-iterations',
      vus: 1,
      iterations: 1,
    },
    streamloader_loadfile: {
      exec: 'test_streamloader_loadfile',
      executor: 'per-vu-iterations',
      vus: 1,
      iterations: 1,
      startTime: '10s', // Stagger the start time
    },
  },
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'count'],
};

// To run this test and compare memory usage:
// 1. Run the test: ./k6 run memory_test.js
// 2. Observe the memory usage reported in the k6 summary for each scenario.
//    Look for the 'peak_rss' metric, which indicates the peak resident memory usage.
//    A lower 'peak_rss' for the 'streamloader_loadfile' scenario would indicate better memory efficiency.

export function test_k6_open() {
  console.log(`[${scenario.name}] Running test with k6's built-in open()`);
  const data = open(FILE_PATH);
  check(data, {
    '[k6_open] file content is not empty': (d) => d.length > 0,
  });
  console.log(`[${scenario.name}] Read ${data.length} bytes.`);
}

export function test_streamloader_loadfile() {
  console.log(`[${scenario.name}] Running test with streamloader.loadFile()`);
  const data = streamloader.loadFile(FILE_PATH);
  check(data, {
    '[streamloader_loadfile] file content is not empty': (d) => d.length > 0,
  });
  console.log(`[${scenario.name}] Read ${data.length} bytes.`);
} 