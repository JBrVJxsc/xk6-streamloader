import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

// Test data
const testObjects = [
  { id: 1, name: "Alice" },
  { id: 2, name: "Bob" },
  { id: 3, name: "Charlie", details: { age: 30, city: "New York" } }
];

// Output file paths for tests
const compressedOutputPath = 'compressed_output.json';
const directCompressedPath = 'direct_compressed.json';

export default function () {
  // Test compression with default compression level
  console.log("Testing objectsToCompressedJsonLines with default compression...");
  const compressedJsonLines = streamloader.objectsToCompressedJsonLines(testObjects);
  
  check(compressedJsonLines, {
    'Compressed JSON lines is a non-empty string': (data) => 
      typeof data === 'string' && data.length > 0,
    
    'Compressed data appears to be base64-encoded': (data) => {
      // Simple heuristic: base64 only contains a limited character set
      const base64Regex = /^[A-Za-z0-9+/=]+$/;
      return base64Regex.test(data);
    }
  });
  
  console.log(`Generated compressed JSON lines (first 50 chars): ${compressedJsonLines.substring(0, 50)}...`);
  
  // Test with specific compression level
  console.log("Testing objectsToCompressedJsonLines with best compression...");
  const bestCompressed = streamloader.objectsToCompressedJsonLines(testObjects, 9); // Best compression
  
  check(bestCompressed, {
    'Best compression generates valid data': (data) => 
      typeof data === 'string' && data.length > 0
  });
  
  // Test no compression
  console.log("Testing objectsToCompressedJsonLines with no compression...");
  const noCompression = streamloader.objectsToCompressedJsonLines(testObjects, 0); // No compression
  
  check(noCompression, {
    'No compression generates valid data': (data) => 
      typeof data === 'string' && data.length > 0
  });
  
  console.log(`Comparing compression levels:
    - Default: ${compressedJsonLines.length} bytes
    - Best: ${bestCompressed.length} bytes
    - None: ${noCompression.length} bytes`);
  
  // Test WriteCompressedJsonLinesToArrayFile
  console.log(`Testing writeCompressedJsonLinesToArrayFile to ${compressedOutputPath}...`);
  const count = streamloader.writeCompressedJsonLinesToArrayFile(compressedJsonLines, compressedOutputPath);
  
  check(count, {
    'writeCompressedJsonLinesToArrayFile returns correct count': (c) => c === testObjects.length
  });
  
  // Read the output file to verify
  const fileContent = streamloader.loadText(compressedOutputPath);
  check(fileContent, {
    'Output file is valid JSON': (content) => {
      try {
        const parsed = JSON.parse(content);
        return Array.isArray(parsed) && parsed.length === testObjects.length;
      } catch (e) {
        console.error(`Failed to parse JSON: ${e}`);
        return false;
      }
    },
    
    'Output contains original data': (content) => {
      const parsed = JSON.parse(content);
      return parsed[0].id === 1 && 
             parsed[1].name === "Bob" &&
             parsed[2].details.city === "New York";
    }
  });
  
  // Test WriteCompressedObjectsToJsonArrayFile (direct compressed write)
  console.log(`Testing writeCompressedObjectsToJsonArrayFile to ${directCompressedPath}...`);
  const directCount = streamloader.writeCompressedObjectsToJsonArrayFile(testObjects, directCompressedPath);
  
  check(directCount, {
    'writeCompressedObjectsToJsonArrayFile returns correct count': (c) => c === testObjects.length
  });
  
  // Read the direct output file to verify
  const directContent = streamloader.loadText(directCompressedPath);
  check(directContent, {
    'Direct compressed output is valid JSON': (content) => {
      try {
        const parsed = JSON.parse(content);
        return Array.isArray(parsed) && parsed.length === testObjects.length;
      } catch (e) {
        console.error(`Failed to parse JSON: ${e}`);
        return false;
      }
    }
  });
  
  // Compare two-step and direct compression results
  check(true, {
    'Two-step and direct compression produce equivalent results': () => {
      const parsed1 = JSON.parse(fileContent);
      const parsed2 = JSON.parse(directContent);
      
      if (parsed1.length !== parsed2.length) return false;
      
      // Simple comparison of stringified objects
      for (let i = 0; i < parsed1.length; i++) {
        if (JSON.stringify(parsed1[i]) !== JSON.stringify(parsed2[i])) {
          return false;
        }
      }
      
      return true;
    }
  });
  
  // Test with special characters and structures
  console.log("Testing compression with special characters and complex structures...");
  
  const specialObjects = [
    { 
      id: 1, 
      html: "<div>Some HTML content</div>", 
      quotes: "Text with \"quotes\" and commas,",
      unicode: "こんにちは",
      emoji: "😊👍"
    },
    { 
      id: 2, 
      nested: { 
        array: [1, 2, 3], 
        object: { key: "value" } 
      }
    }
  ];
  
  const specialPath = 'special_compressed.json';
  const specialCompressed = streamloader.objectsToCompressedJsonLines(specialObjects);
  const specialCount = streamloader.writeCompressedJsonLinesToArrayFile(specialCompressed, specialPath);
  
  check(specialCount, {
    'Compression handles special characters correctly': (count) => count === specialObjects.length
  });
  
  // Verify special characters were preserved
  const specialContent = streamloader.loadText(specialPath);
  check(specialContent, {
    'Special characters are preserved in compressed output': (content) => {
      try {
        const parsed = JSON.parse(content);
        return parsed[0].html.includes('<div>') && 
               parsed[0].quotes.includes('\"quotes\"') &&
               parsed[0].unicode === "こんにちは" &&
               parsed[0].emoji === "😊👍" &&
               Array.isArray(parsed[1].nested.array) &&
               parsed[1].nested.array.length === 3;
      } catch (e) {
        console.error(`Failed to parse special JSON: ${e}`);
        return false;
      }
    }
  });
  
  console.log("All compression tests completed successfully");
}