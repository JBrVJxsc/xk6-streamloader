import streamloader from 'k6/x/streamloader';
import { check } from 'k6';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.2/index.js';

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    'checks': ['rate==1.0'],  // All checks must pass
  },
};

// Generate a test dataset with various characteristics that affect compression
function generateTestDataset(size) {
  console.log(`Generating ${size}-item test dataset...`);
  const dataset = [];
  
  // Add a variety of data types and structures to test compression effectiveness
  for (let i = 0; i < size; i++) {
    const recordType = i % 5; // Create different types of records
    
    switch (recordType) {
      case 0: // Simple record with numbers and text
        dataset.push({
          id: i,
          name: `User-${i}`,
          active: i % 2 === 0,
          score: Math.floor(Math.random() * 100) / 10,
        });
        break;
        
      case 1: // Record with repeated values (good for compression)
        dataset.push({
          id: i,
          category: "standard",
          tags: ["standard", "standard", "standard", "standard"],
          status: "active",
          type: "standard",
          flags: {
            standard: true,
            default: true,
            normal: true
          }
        });
        break;
        
      case 2: // Record with unique values (bad for compression)
        dataset.push({
          id: i,
          uniqueId: `${Math.random().toString(36).substring(2, 15)}-${Math.random().toString(36).substring(2, 15)}`,
          timestamp: new Date().toISOString(),
          randomBytes: Array(20).fill().map(() => Math.floor(Math.random() * 256).toString(16).padStart(2, '0')).join('')
        });
        break;
        
      case 3: // Record with nested structure
        dataset.push({
          id: i,
          user: {
            firstName: `First-${i}`,
            lastName: `Last-${i}`,
            address: {
              street: `${i} Main St`,
              city: i % 5 === 0 ? "New York" : "Los Angeles",
              zip: `${10000 + i}`
            },
            preferences: {
              theme: i % 2 === 0 ? "dark" : "light",
              notifications: true,
              language: "en"
            }
          }
        });
        break;
        
      case 4: // Record with arrays
        const items = [];
        for (let j = 0; j < (i % 10) + 1; j++) {
          items.push({
            itemId: `item-${j}`,
            quantity: j + 1,
            price: (j + 1) * 10.99
          });
        }
        
        dataset.push({
          id: i,
          order: `ORD-${1000 + i}`,
          items: items,
          total: items.reduce((sum, item) => sum + (item.quantity * item.price), 0)
        });
        break;
    }
  }
  
  return dataset;
}

export default function() {
  // Create datasets of various sizes
  const smallDataset = generateTestDataset(10);
  const mediumDataset = generateTestDataset(100);
  const largeDataset = generateTestDataset(1000);
  
  // Test with different compression levels
  const compressionLevels = [
    { name: "Default", level: undefined },
    { name: "No Compression", level: 0 },
    { name: "Best Speed", level: 1 },
    { name: "Best Compression", level: 9 }
  ];
  
  const results = {
    small: {},
    medium: {},
    large: {}
  };
  
  // Small dataset tests
  console.log("Testing compression on small dataset (10 items)...");
  const smallJsonl = streamloader.objectsToJsonLines(smallDataset);
  results.small.uncompressed = smallJsonl.length;
  
  for (const comp of compressionLevels) {
    const start = new Date();
    const compressedData = comp.level !== undefined 
      ? streamloader.objectsToCompressedJsonLines(smallDataset, comp.level) 
      : streamloader.objectsToCompressedJsonLines(smallDataset);
    const end = new Date();
    
    results.small[comp.name] = {
      size: compressedData.length,
      time: end - start,
      ratio: compressedData.length / smallJsonl.length
    };
    
    check(compressedData, {
      [`Small dataset - ${comp.name} compression produces valid output`]: (data) => 
        typeof data === 'string' && data.length > 0
    });
  }
  
  // Medium dataset tests
  console.log("Testing compression on medium dataset (100 items)...");
  const mediumJsonl = streamloader.objectsToJsonLines(mediumDataset);
  results.medium.uncompressed = mediumJsonl.length;
  
  for (const comp of compressionLevels) {
    const start = new Date();
    const compressedData = comp.level !== undefined 
      ? streamloader.objectsToCompressedJsonLines(mediumDataset, comp.level) 
      : streamloader.objectsToCompressedJsonLines(mediumDataset);
    const end = new Date();
    
    results.medium[comp.name] = {
      size: compressedData.length,
      time: end - start,
      ratio: compressedData.length / mediumJsonl.length
    };
    
    check(compressedData, {
      [`Medium dataset - ${comp.name} compression produces valid output`]: (data) => 
        typeof data === 'string' && data.length > 0
    });
    
    // Test the file write for medium dataset with default compression
    if (comp.name === "Default") {
      const mediumFilePath = "medium_compressed.json";
      const writeStart = new Date();
      const count = streamloader.writeCompressedJsonLinesToArrayFile(compressedData, mediumFilePath);
      const writeEnd = new Date();
      
      check(count, {
        'Medium dataset - Decompression and file write works correctly': (c) => c === mediumDataset.length
      });
      
      console.log(`Medium dataset write time: ${writeEnd - writeStart}ms`);
    }
  }
  
  // Large dataset tests
  console.log("Testing compression on large dataset (1000 items)...");
  const largeJsonl = streamloader.objectsToJsonLines(largeDataset);
  results.large.uncompressed = largeJsonl.length;
  
  for (const comp of compressionLevels) {
    const start = new Date();
    const compressedData = comp.level !== undefined 
      ? streamloader.objectsToCompressedJsonLines(largeDataset, comp.level) 
      : streamloader.objectsToCompressedJsonLines(largeDataset);
    const end = new Date();
    
    results.large[comp.name] = {
      size: compressedData.length,
      time: end - start,
      ratio: compressedData.length / largeJsonl.length
    };
    
    check(compressedData, {
      [`Large dataset - ${comp.name} compression produces valid output`]: (data) => 
        typeof data === 'string' && data.length > 0
    });
  }
  
  // Print results
  console.log("\n===== Compression Results =====");
  
  Object.entries(results).forEach(([datasetName, datasetResults]) => {
    console.log(`\n${datasetName.toUpperCase()} DATASET (Uncompressed: ${datasetResults.uncompressed} bytes):`);
    
    Object.entries(datasetResults).forEach(([levelName, metrics]) => {
      if (levelName === 'uncompressed') return;
      
      console.log(`  ${levelName}: 
    Size: ${metrics.size} bytes (${Math.round(metrics.ratio * 10000) / 100}% of original)
    Time: ${metrics.time}ms`);
    });
  });
  
  // Direct vs two-step comparison
  console.log("\n===== Direct vs Two-Step Write Comparison =====");
  
  // Two-step approach
  const twoStepStart = new Date();
  const compressedData = streamloader.objectsToCompressedJsonLines(mediumDataset);
  const twoStepPath = "two_step_result.json";
  streamloader.writeCompressedJsonLinesToArrayFile(compressedData, twoStepPath);
  const twoStepEnd = new Date();
  
  // Direct approach
  const directStart = new Date();
  const directPath = "direct_result.json";
  streamloader.writeCompressedObjectsToJsonArrayFile(mediumDataset, directPath);
  const directEnd = new Date();
  
  console.log(`  Two-step approach time: ${twoStepEnd - twoStepStart}ms`);
  console.log(`  Direct approach time: ${directEnd - directStart}ms`);
  
  // Verify results are the same
  const twoStepContent = streamloader.loadText(twoStepPath);
  const directContent = streamloader.loadText(directPath);
  
  check(twoStepContent, {
    'Two-step and direct approaches produce identical results': () => twoStepContent === directContent
  });
}

// Custom summary reporting
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
  };
}