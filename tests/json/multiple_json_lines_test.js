import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function () {
  // ---------------------------------------------------------
  // Test 1: Multiple JSON Lines (uncompressed)
  // ---------------------------------------------------------
  console.log("Testing writeMultipleJsonLinesToArrayFile...");
  
  // Create multiple batches of JSON lines
  const batch1 = streamloader.objectsToJsonLines([
    { id: 1, name: "Alice", department: "Engineering" },
    { id: 2, name: "Bob", department: "Marketing" }
  ]);
  
  const batch2 = streamloader.objectsToJsonLines([
    { id: 3, name: "Charlie", department: "Sales", details: { age: 30 } },
    { id: 4, name: "Dave", department: "Finance" }
  ]);
  
  // Write multiple JSON line strings to a single array file
  const jsonArrayPath = 'multiple_json_lines_output.json';
  const objectCount = streamloader.writeMultipleJsonLinesToArrayFile([batch1, batch2], jsonArrayPath);
  
  check(objectCount, {
    'writeMultipleJsonLinesToArrayFile returns correct count': count => count === 4
  });
  
  // Read back the file and verify
  const fileContent = streamloader.loadText(jsonArrayPath);
  
  check(fileContent, {
    'Output file is valid JSON': content => {
      try {
        const parsed = JSON.parse(content);
        return Array.isArray(parsed) && parsed.length === 4;
      } catch (e) {
        console.error(`Failed to parse JSON: ${e}`);
        return false;
      }
    },
    'Objects have correct data': content => {
      const parsed = JSON.parse(content);
      return parsed[0].id === 1 && 
             parsed[1].name === "Bob" &&
             parsed[2].department === "Sales" && 
             parsed[3].id === 4;
    }
  });
  
  console.log(`Successfully wrote ${objectCount} objects to ${jsonArrayPath}`);

  // ---------------------------------------------------------
  // Test 2: Multiple Compressed JSON Lines
  // ---------------------------------------------------------
  console.log("Testing writeMultipleCompressedJsonLinesToArrayFile...");
  
  // Create multiple batches of compressed JSON lines
  const compBatch1 = streamloader.objectsToCompressedJsonLines([
    { id: 5, name: "Eve", department: "HR" },
    { id: 6, name: "Frank", department: "Legal" }
  ]);
  
  const compBatch2 = streamloader.objectsToCompressedJsonLines([
    { id: 7, name: "Grace", department: "IT", skills: ["Java", "Python", "Go"] },
    { id: 8, name: "Hank", department: "Operations", location: { city: "New York" } }
  ]);
  
  // Write multiple compressed JSON line strings to a single array file
  const compJsonArrayPath = 'multiple_compressed_json_lines_output.json';
  const compObjectCount = streamloader.writeMultipleCompressedJsonLinesToArrayFile(
    [compBatch1, compBatch2], compJsonArrayPath);
  
  check(compObjectCount, {
    'writeMultipleCompressedJsonLinesToArrayFile returns correct count': count => count === 4
  });
  
  // Read back the file and verify
  const compFileContent = streamloader.loadText(compJsonArrayPath);
  
  check(compFileContent, {
    'Compressed output file is valid JSON': content => {
      try {
        const parsed = JSON.parse(content);
        return Array.isArray(parsed) && parsed.length === 4;
      } catch (e) {
        console.error(`Failed to parse JSON: ${e}`);
        return false;
      }
    },
    'Compressed objects have correct data': content => {
      const parsed = JSON.parse(content);
      return parsed[0].id === 5 && 
             parsed[1].department === "Legal" &&
             Array.isArray(parsed[2].skills) && 
             parsed[3].location.city === "New York";
    }
  });
  
  console.log(`Successfully wrote ${compObjectCount} compressed objects to ${compJsonArrayPath}`);
  
  // ---------------------------------------------------------
  // Test 3: Mixed batch types
  // ---------------------------------------------------------
  console.log("Testing mixed batch types...");
  
  // Create different types of objects in different batches
  const typeBatch1 = streamloader.objectsToCompressedJsonLines([
    { type: "user", id: 101, name: "User 1" },
    { type: "user", id: 102, name: "User 2" }
  ]);
  
  const typeBatch2 = streamloader.objectsToCompressedJsonLines([
    { type: "product", id: "P1", name: "Product 1", price: 19.99 },
    { type: "product", id: "P2", name: "Product 2", price: 29.99 }
  ]);
  
  const typeBatch3 = streamloader.objectsToCompressedJsonLines([
    { type: "order", id: "O1", userId: 101, products: ["P1", "P2"], total: 49.98 }
  ]);
  
  // Write all different batches to a single array file
  const mixedArrayPath = 'mixed_types_output.json';
  const mixedCount = streamloader.writeMultipleCompressedJsonLinesToArrayFile(
    [typeBatch1, typeBatch2, typeBatch3], mixedArrayPath);
  
  check(mixedCount, {
    'Mixed batches returns correct total count': count => count === 5
  });
  
  // Read back the file and verify
  const mixedContent = streamloader.loadText(mixedArrayPath);
  const mixedObjects = JSON.parse(mixedContent);
  
  check(mixedObjects, {
    'Mixed objects file has correct length': objs => objs.length === 5,
    'Contains correct user objects': objs => 
      objs[0].type === "user" && objs[1].type === "user",
    'Contains correct product objects': objs => 
      objs[2].type === "product" && objs[3].type === "product",
    'Contains correct order object': objs => 
      objs[4].type === "order" && Array.isArray(objs[4].products)
  });
  
  console.log(`Successfully wrote ${mixedCount} objects of different types`);

  // ---------------------------------------------------------
  // Test 4: Complete Roundtrip Test
  // ---------------------------------------------------------
  console.log("Performing complete roundtrip test...");
  
  // Original objects (containing various data types and structures)
  const originalObjects = [
    // Simple objects
    { id: 1, name: "Alice", active: true, score: 95.5 },
    { id: 2, name: "Bob", active: false, score: 82.3 },
    
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
    }
  ];
  
  // Split objects into batches
  const batch1Objects = originalObjects.slice(0, 2);
  const batch2Objects = originalObjects.slice(2);
  
  // Step 1: Convert objects to JSON lines
  const roundtripBatch1 = streamloader.objectsToJsonLines(batch1Objects);
  const roundtripBatch2 = streamloader.objectsToJsonLines(batch2Objects);
  
  // Step 2: Write multiple JSON line batches to array file
  const roundtripPath = 'roundtrip_test.json';
  const rtCount = streamloader.writeMultipleJsonLinesToArrayFile(
    [roundtripBatch1, roundtripBatch2], roundtripPath);
  
  // Step 3: Read back and verify
  const rtContent = streamloader.loadText(roundtripPath);
  const rtObjects = JSON.parse(rtContent);
  
  check(rtCount, {
    'Roundtrip object count matches': count => count === originalObjects.length
  });
  
  check(rtObjects, {
    'Simple values preserved': objs => 
      objs[0].id === 1 && 
      objs[0].name === "Alice" && 
      objs[0].active === true,
      
    'Nested structures preserved': objs => 
      objs[2].profile.address.city === "New York" && 
      objs[2].hobbies.length === 3,
      
    'Special characters preserved': objs => 
      objs[3].description.includes('\"quotes\"') && 
      objs[3].html.includes('<br/>'),
      
    'Unicode characters preserved': objs => 
      objs[4].name === "José María Müller" && 
      objs[4].location === "北京市朝阳区" && 
      objs[4].emoji.includes('😊')
  });
  
  console.log("Roundtrip test completed successfully");
  
  // ---------------------------------------------------------
  // Test 5: Complete Compressed Roundtrip Test
  // ---------------------------------------------------------
  console.log("Performing complete compressed roundtrip test...");
  
  // Compress the original objects in batches
  const compRoundtripBatch1 = streamloader.objectsToCompressedJsonLines(batch1Objects);
  const compRoundtripBatch2 = streamloader.objectsToCompressedJsonLines(batch2Objects);
  
  // Write compressed batches to array file
  const compRoundtripPath = 'compressed_roundtrip_test.json';
  const compRtCount = streamloader.writeMultipleCompressedJsonLinesToArrayFile(
    [compRoundtripBatch1, compRoundtripBatch2], compRoundtripPath);
  
  // Read back and verify
  const compRtContent = streamloader.loadText(compRoundtripPath);
  const compRtObjects = JSON.parse(compRtContent);
  
  check(compRtCount, {
    'Compressed roundtrip object count matches': count => count === originalObjects.length
  });
  
  check(compRtObjects, {
    'Compressed roundtrip preserves simple values': objs => 
      objs[0].id === 1 && 
      objs[0].name === "Alice" && 
      objs[0].active === true,
      
    'Compressed roundtrip preserves nested structures': objs => 
      objs[2].profile.address.city === "New York" && 
      objs[2].hobbies.length === 3,
      
    'Compressed roundtrip preserves special characters': objs => 
      objs[3].description.includes('\"quotes\"') && 
      objs[3].html.includes('<br/>'),
      
    'Compressed roundtrip preserves Unicode characters': objs => 
      objs[4].name === "José María Müller" && 
      objs[4].location === "北京市朝阳区" && 
      objs[4].emoji.includes('😊')
  });
  
  console.log("Compressed roundtrip test completed successfully");
  
  // ---------------------------------------------------------
  // Test 6: Compare JSON and original objects
  // ---------------------------------------------------------
  console.log("Comparing outputs from regular and compressed methods...");
  
  // Verify that both methods produce identical results
  check(true, {
    'Both methods produce identical results': () => {
      // Compare the stringified objects for equality
      const regularJson = JSON.stringify(rtObjects);
      const compressedJson = JSON.stringify(compRtObjects);
      return regularJson === compressedJson;
    }
  });
  
  // ---------------------------------------------------------
  // Test 7: MultipleCompressedJsonLinesToObjects
  // ---------------------------------------------------------
  console.log("Testing multipleCompressedJsonLinesToObjects...");
  
  // Use the same compressed batches from the earlier test
  const compressedObjectsResult = streamloader.multipleCompressedJsonLinesToObjects(
    [compBatch1, compBatch2]);
  
  check(compressedObjectsResult, {
    'multipleCompressedJsonLinesToObjects returns correct count': objects => objects.length === 4,
    'multipleCompressedJsonLinesToObjects returns array of objects': objects => Array.isArray(objects),
    'First object has correct structure': objects => 
      objects[0].id === 5 && objects[0].name === "Eve" && objects[0].department === "HR",
    'Second object has correct structure': objects => 
      objects[1].id === 6 && objects[1].name === "Frank" && objects[1].department === "Legal",
    'Third object has nested array': objects => 
      objects[2].id === 7 && Array.isArray(objects[2].skills) && objects[2].skills.length === 3,
    'Fourth object has nested location': objects => 
      objects[3].id === 8 && objects[3].location && objects[3].location.city === "New York"
  });
  
  console.log(`Successfully decompressed and parsed ${compressedObjectsResult.length} objects`);
  
  // Test with mixed batch types (from earlier test)
  const mixedObjectsResult = streamloader.multipleCompressedJsonLinesToObjects(
    [typeBatch1, typeBatch2, typeBatch3]);
  
  check(mixedObjectsResult, {
    'Mixed types returns correct count': objects => objects.length === 5,
    'Contains user objects': objects => 
      objects[0].type === "user" && objects[1].type === "user",
    'Contains product objects': objects => 
      objects[2].type === "product" && objects[3].type === "product",
    'Contains order object': objects => 
      objects[4].type === "order" && Array.isArray(objects[4].products)
  });
  
  console.log(`Successfully processed ${mixedObjectsResult.length} mixed-type objects`);
  
  // Test with empty array
  const emptyResult = streamloader.multipleCompressedJsonLinesToObjects([]);
  check(emptyResult, {
    'Empty array returns empty result': objects => Array.isArray(objects) && objects.length === 0
  });
  
  // Test error handling with invalid data
  try {
    streamloader.multipleCompressedJsonLinesToObjects(["invalid_base64_data!!!"]);
    check(false, { 'Should have thrown error for invalid data': () => false });
  } catch (e) {
    check(true, { 'Correctly throws error for invalid data': () => true });
    console.log("Correctly handled invalid compressed data");
  }
  
  // ---------------------------------------------------------
  // Test 8: Compare multipleCompressedJsonLinesToObjects with file-based method
  // ---------------------------------------------------------
  console.log("Comparing multipleCompressedJsonLinesToObjects with file-based method...");
  
  // Use the objects we got from the new function
  const objectsFromMemory = compressedObjectsResult;
  
  // Compare with objects from the file written earlier
  const objectsFromFile = JSON.parse(compFileContent);
  
  check(true, {
    'Memory and file methods produce identical results': () => {
      if (objectsFromMemory.length !== objectsFromFile.length) {
        console.log(`Length mismatch: memory=${objectsFromMemory.length}, file=${objectsFromFile.length}`);
        return false;
      }
      
      // Helper function to deep compare objects regardless of key order
      function deepEqual(obj1, obj2) {
        if (obj1 === obj2) return true;
        if (obj1 == null || obj2 == null) return false;
        if (typeof obj1 !== typeof obj2) return false;
        
        if (typeof obj1 === 'object') {
          const keys1 = Object.keys(obj1).sort();
          const keys2 = Object.keys(obj2).sort();
          
          if (keys1.length !== keys2.length) return false;
          if (keys1.join(',') !== keys2.join(',')) return false;
          
          for (let key of keys1) {
            if (!deepEqual(obj1[key], obj2[key])) return false;
          }
          return true;
        }
        
        return obj1 === obj2;
      }
      
      for (let i = 0; i < objectsFromMemory.length; i++) {
        if (!deepEqual(objectsFromMemory[i], objectsFromFile[i])) {
          console.log(`Object ${i} content mismatch`);
          console.log(`Memory:`, JSON.stringify(objectsFromMemory[i]));
          console.log(`File:`, JSON.stringify(objectsFromFile[i]));
          return false;
        }
      }
      return true;
    }
  });
  
  console.log("Memory vs file comparison completed successfully!");

  console.log("All tests completed successfully!");
}