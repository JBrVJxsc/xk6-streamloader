import streamloader from 'k6/x/streamloader';
import { check } from 'k6';
import { b64decode } from 'k6/encoding';

// Test data
const testObjects = [
  { id: 1, name: "Alice" },
  { id: 2, name: "Bob" },
  { id: 3, name: "Charlie", details: { age: 30, city: "New York" } }
];

// Output file paths for tests
const jsonLinesPath = 'test_output.jsonl';
const jsonArrayPath = 'test_output.json';
const roundTripPath = 'test_roundtrip.json';

export default function () {
  // Test objectsToJsonLines
  const jsonLines = streamloader.objectsToJsonLines(testObjects);
  console.log(`Generated JSON lines: ${jsonLines}`);
  
  check(jsonLines, {
    'ObjectsToJsonLines generates valid output': (lines) => {
      // Lines should contain each object serialized on a separate line
      const linesArr = lines.split('\n');
      return linesArr.length === testObjects.length &&
        JSON.parse(linesArr[0]).id === 1 &&
        JSON.parse(linesArr[1]).id === 2;
    }
  });

  // Test WriteJsonLinesToArrayFile
  const objectCount = streamloader.writeJsonLinesToArrayFile(jsonLines, jsonArrayPath);
  console.log(`Wrote ${objectCount} objects to ${jsonArrayPath}`);
  
  check(objectCount, {
    'writeJsonLinesToArrayFile returns correct object count': (count) => count === testObjects.length
  });
  
  // Read the generated file to verify its contents
  const fileContent = streamloader.loadText(jsonArrayPath);
  check(fileContent, {
    'Generated JSON array file is valid JSON': (content) => {
      try {
        const parsed = JSON.parse(content);
        return Array.isArray(parsed) && parsed.length === testObjects.length;
      } catch (e) {
        console.error(`Failed to parse JSON: ${e}`);
        return false;
      }
    }
  });

  // Test WriteObjectsToJsonArrayFile (direct write)
  const directWriteCount = streamloader.writeObjectsToJsonArrayFile(testObjects, roundTripPath);
  console.log(`Directly wrote ${directWriteCount} objects to ${roundTripPath}`);
  
  check(directWriteCount, {
    'writeObjectsToJsonArrayFile returns correct object count': (count) => count === testObjects.length
  });
  
  // Read the generated file to verify its contents
  const roundTripContent = streamloader.loadText(roundTripPath);
  check(roundTripContent, {
    'Round-trip JSON array file is valid': (content) => {
      try {
        const parsed = JSON.parse(content);
        return Array.isArray(parsed) && 
               parsed.length === testObjects.length &&
               parsed[0].id === testObjects[0].id &&
               parsed[1].name === testObjects[1].name;
      } catch (e) {
        console.error(`Failed to parse JSON: ${e}`);
        return false;
      }
    }
  });

  // Test with special characters and complex structures
  const specialObjects = [
    { id: 1, html: "<div>Some HTML content</div>", quotes: "Text with \"quotes\" and commas," },
    { id: 2, nested: { array: [1, 2, 3], object: { key: "value" } } },
    { id: 3, unicode: "こんにちは", emoji: "😊👍" }
  ];
  
  const specialJsonLines = streamloader.objectsToJsonLines(specialObjects);
  const specialPath = 'special_test.json';
  const specialCount = streamloader.writeJsonLinesToArrayFile(specialJsonLines, specialPath);
  
  check(specialCount, {
    'Special characters are handled correctly': (count) => count === specialObjects.length
  });
  
  // Read and verify the special characters file
  const specialContent = streamloader.loadText(specialPath);
  check(specialContent, {
    'Special characters are preserved': (content) => {
      try {
        const parsed = JSON.parse(content);
        return parsed[0].html.includes('<div>') && 
               parsed[0].quotes.includes('\"quotes\"') &&
               parsed[2].unicode === "こんにちは" &&
               parsed[2].emoji === "😊👍";
      } catch (e) {
        console.error(`Failed to parse special JSON: ${e}`);
        return false;
      }
    }
  });
}