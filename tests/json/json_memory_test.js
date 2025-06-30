import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export const options = {
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'count'],
};

// Generate a large dataset
function generateLargeDataset(size) {
  console.log(`Generating dataset of ${size} objects...`);
  const objects = [];
  for (let i = 0; i < size; i++) {
    objects.push({
      id: i,
      value: `Value-${i}`,
      timestamp: new Date().toISOString(),
      details: {
        category: i % 5,
        tags: [`tag-${i % 10}`, `tag-${i % 7}`],
        active: i % 2 === 0
      }
    });
  }
  return objects;
}

export default function () {
  // Test with a medium-sized dataset (adjust based on your memory constraints)
  const dataSize = 10000;
  const dataset = generateLargeDataset(dataSize);
  
  console.log('Testing ObjectsToJsonLines memory efficiency');
  const jsonLines = streamloader.ObjectsToJsonLines(dataset);
  
  check(jsonLines, {
    'ObjectsToJsonLines handles large dataset': (lines) => {
      return lines.split('\n').length === dataSize;
    }
  });

  const outputPath = 'large_dataset.json';
  console.log(`Testing WriteJsonLinesToArrayFile memory efficiency, writing to ${outputPath}`);
  const count = streamloader.WriteJsonLinesToArrayFile(jsonLines, outputPath);
  
  check(count, {
    'WriteJsonLinesToArrayFile processes all objects': (c) => c === dataSize
  });
  
  // Test direct write for comparison
  const directOutputPath = 'large_dataset_direct.json';
  console.log(`Testing WriteObjectsToJsonArrayFile memory efficiency, writing to ${directOutputPath}`);
  const directCount = streamloader.WriteObjectsToJsonArrayFile(dataset, directOutputPath);
  
  check(directCount, {
    'WriteObjectsToJsonArrayFile processes all objects': (c) => c === dataSize
  });
  
  // Test combining multiple files
  const files = [outputPath, directOutputPath];
  const combinedPath = 'combined_dataset.json';
  console.log(`Testing CombineJsonArrayFiles, combining ${files.join(', ')} into ${combinedPath}`);
  const combinedCount = streamloader.CombineJsonArrayFiles(files, combinedPath);
  
  check(combinedCount, {
    'CombineJsonArrayFiles combines all objects': (c) => c === dataSize * 2
  });
}