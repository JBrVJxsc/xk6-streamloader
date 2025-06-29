import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

const FILE_PATH = 'large_file.txt';

// To run this test:
// 1. Run the test: ./k6 run memory_test_streamloader.js
// 2. Observe the 'peak_rss' in the summary.
// This script measures the memory usage of k6 when loading a large file
// using the streamloader.loadText() function in the VU context.

export const options = {
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'count'],
};

export default function () {
  console.log(`Running test with streamloader.loadText()`);
  const data = streamloader.loadText(FILE_PATH);
  check(data, {
    '[streamloader_loadText] file content is not empty': (d) => d.length > 0,
  });
  console.log(`Read ${data.length} bytes.`);
} 