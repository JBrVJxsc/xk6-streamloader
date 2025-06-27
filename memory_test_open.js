import { check } from 'k6';

const FILE_PATH = 'large_file.txt';

// To run this test:
// 1. Run the test: ./k6 run memory_test_open.js
// 2. Observe the 'peak_rss' in the summary.
// This script measures the memory usage of k6 when loading a large file
// using the built-in open() function in the init context.

console.log(`Reading file with open() in init context: ${FILE_PATH}`);
const data = open(FILE_PATH);

export const options = {
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'count'],
};

export default function () {
  check(data, {
    '[k6_open] file content is not empty': (d) => d.length > 0,
  });
  // The primary memory allocation happens in the init context.
  // This VU function just checks the data to ensure it was loaded.
} 