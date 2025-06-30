import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

// Test data with various data types and structures
const testObjects = [
  // Simple objects with primitive types
  { id: 1, name: "Alice", age: 30, active: true, score: 95.5 },
  { id: 2, name: "Bob", age: 25, active: false, score: 82.3 },
  
  // Objects with nested structures
  { 
    id: 3, 
    name: "Charlie",
    profile: {
      age: 40,
      address: {
        street: "123 Main St",
        city: "New York",
        zip: "10001"
      },
      phones: ["555-1234", "555-5678"]
    },
    hobbies: ["reading", "cycling", "photography"]
  },
  
  // Objects with special characters
  {
    id: 4,
    description: "Product with \"quotes\" and commas, plus other chars: !@#$%^&*()",
    html: "<div>Some HTML content with <br/> tags</div>",
    json: "{\"nested\":\"json string\"}",
  },
  
  // Objects with non-ASCII characters
  {
    id: 5,
    name: "José María Müller",
    location: "北京市朝阳区",
    emoji: "😊👍👨‍👩‍👧‍👦" // Emoji and complex Unicode
  },
  
  // Objects with null values and empty arrays/objects
  {
    id: 6,
    description: null,
    emptyArray: [],
    emptyObject: {},
    metadata: null
  }
];

export default function () {
  // File paths for the round-trip test
  const jsonlPath = 'roundtrip_test.jsonl';
  const jsonArrayPath = 'roundtrip_test.json';
  
  console.log("Starting JSON utilities round-trip test");
  
  // Step 1: Convert objects to JSON lines
  const jsonLines = streamloader.objectsToJsonLines(testObjects);
  check(jsonLines, {
    'objectsToJsonLines generates output': (lines) => lines.length > 0
  });
  
  // Skip writing the JSONL to a file - just use it directly
  
  // Step 2: Write JSON lines to JSON array file
  const objectCount = streamloader.writeJsonLinesToArrayFile(jsonLines, jsonArrayPath);
  console.log(`Wrote ${objectCount} objects to ${jsonArrayPath}`);
  
  check(objectCount, {
    'writeJsonLinesToArrayFile processed correct number of objects': (count) => count === testObjects.length
  });
  
  // Step 3: Load the JSON array back and verify data integrity
  const loadedJson = streamloader.loadJSON(jsonArrayPath);
  check(loadedJson, {
    'Loaded JSON array has correct structure': (json) => Array.isArray(json) && json.length === testObjects.length,
    
    'Simple values are preserved': (json) => {
      return json[0].id === 1 && 
             json[0].name === "Alice" && 
             json[0].active === true &&
             json[0].score === 95.5;
    },
    
    'Nested objects are preserved': (json) => {
      return json[2].profile.address.city === "New York" && 
             Array.isArray(json[2].hobbies) && 
             json[2].hobbies.length === 3;
    },
    
    'Special characters are preserved': (json) => {
      return json[3].description.includes('\"quotes\"') && 
             json[3].html.includes('<div>') &&
             json[3].html.includes('<br/>');
    },
    
    'Unicode characters are preserved': (json) => {
      return json[4].name.includes('José') && 
             json[4].location.includes('北京') &&
             json[4].emoji.includes('😊');
    },
    
    'Null values and empty structures are preserved': (json) => {
      return json[5].description === null && 
             Array.isArray(json[5].emptyArray) &&
             json[5].emptyArray.length === 0 &&
             Object.keys(json[5].emptyObject).length === 0;
    }
  });
  
  // Test the combined workflow in one step
  const directJsonPath = 'direct_write_test.json';
  const directCount = streamloader.writeObjectsToJsonArrayFile(testObjects, directJsonPath);
  
  check(directCount, {
    'writeObjectsToJsonArrayFile processed correct number of objects': (count) => count === testObjects.length
  });
  
  // Compare the two output files to ensure they're identical
  const arrayContent = streamloader.loadText(jsonArrayPath);
  const directContent = streamloader.loadText(directJsonPath);
  
  check(arrayContent, {
    'Two-step and direct conversion produce identical results': (content) => {
      // Note: We compare the parsed objects rather than raw JSON strings
      // because JSON.stringify might format the same objects differently
      const parsed1 = JSON.parse(content);
      const parsed2 = JSON.parse(directContent);
      
      // Simple check - compare length
      if (parsed1.length !== parsed2.length) return false;
      
      // Compare each object with its counterpart
      for (let i = 0; i < parsed1.length; i++) {
        const obj1Json = JSON.stringify(parsed1[i]);
        const obj2Json = JSON.stringify(parsed2[i]);
        if (obj1Json !== obj2Json) return false;
      }
      
      return true;
    }
  });
  
  console.log("Round-trip test completed successfully");
}