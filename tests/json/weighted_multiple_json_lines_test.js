import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function () {
  console.log("Testing writeWeightedMultipleCompressedJsonLinesToArrayFile...");

  // ---------------------------------------------------------
  // Test 1: Basic weighted functionality
  // ---------------------------------------------------------
  console.log("Test 1: Basic weighted functionality");
  
  // Create test data
  const batch1 = [
    { id: 1, name: "Alice", type: "user" },
    { id: 2, name: "Bob", type: "user" }
  ];
  
  const batch2 = [
    { id: 3, name: "Charlie", type: "admin" },
    { id: 4, name: "Dave", type: "admin" },
    { id: 5, name: "Eve", type: "admin" },
    { id: 6, name: "Frank", type: "admin" },
    { id: 7, name: "Grace", type: "admin" }
  ];
  
  const batch3 = [
    { id: 8, name: "Henry", type: "superuser" }
  ];
  
  // Compress the batches
  const compBatch1 = streamloader.objectsToCompressedJsonLines(batch1);
  const compBatch2 = streamloader.objectsToCompressedJsonLines(batch2);
  const compBatch3 = streamloader.objectsToCompressedJsonLines(batch3);
  
  console.log("Created compressed batches for testing");
  
  // ---------------------------------------------------------
  // Test 2: Equal weight (count == weight)
  // ---------------------------------------------------------
  console.log("Test 2: Equal weight (count == weight)");
  
  const equalWeightBatches = [
    [[compBatch1], 2] // Group with 2 objects, weight 2 -> keep all
  ];
  
  const equalWeightFile = "test_equal_weight.json";
  const equalWeightCount = streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
    equalWeightBatches, equalWeightFile);
  
  const equalWeightContent = JSON.parse(streamloader.loadText(equalWeightFile));
  
  check(equalWeightCount, {
    'Equal weight returns correct count': count => count === 2
  });
  
  check(equalWeightContent, {
    'Equal weight file has correct length': content => content.length === 2,
    'Equal weight preserves Alice': content => content[0].name === "Alice",
    'Equal weight preserves Bob': content => content[1].name === "Bob"
  });
  
  console.log(`Equal weight test: wrote ${equalWeightCount} objects`);
  
  // ---------------------------------------------------------
  // Test 3: Oversample (count > weight)
  // ---------------------------------------------------------
  console.log("Test 3: Oversample (count > weight)");
  
  const oversampleBatches = [
    [[compBatch2], 3] // Group with 5 objects, weight 3 -> slice to first 3
  ];
  
  const oversampleFile = "test_oversample.json";
  const oversampleCount = streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
    oversampleBatches, oversampleFile);
  
  const oversampleContent = JSON.parse(streamloader.loadText(oversampleFile));
  
  check(oversampleCount, {
    'Oversample returns correct count': count => count === 3
  });
  
  check(oversampleContent, {
    'Oversample file has correct length': content => content.length === 3,
    'Oversample keeps first object': content => content[0].name === "Charlie",
    'Oversample keeps second object': content => content[1].name === "Dave",
    'Oversample keeps third object': content => content[2].name === "Eve"
  });
  
  console.log(`Oversample test: wrote ${oversampleCount} objects`);
  
  // ---------------------------------------------------------
  // Test 4: Undersample (count < weight)
  // ---------------------------------------------------------
  console.log("Test 4: Undersample (count < weight)");
  
  const undersampleBatches = [
    [[compBatch3], 4] // Group with 1 object, weight 4 -> duplicate to 4
  ];
  
  const undersampleFile = "test_undersample.json";
  const undersampleCount = streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
    undersampleBatches, undersampleFile);
  
  const undersampleContent = JSON.parse(streamloader.loadText(undersampleFile));
  
  check(undersampleCount, {
    'Undersample returns correct count': count => count === 4
  });
  
  check(undersampleContent, {
    'Undersample file has correct length': content => content.length === 4,
    'Undersample duplicates correctly': content => 
      content.every(obj => obj.name === "Henry" && obj.id === 8)
  });
  
  console.log(`Undersample test: wrote ${undersampleCount} objects`);
  
  // ---------------------------------------------------------
  // Test 5: Complex duplication pattern
  // ---------------------------------------------------------
  console.log("Test 5: Complex duplication pattern");
  
  const complexBatches = [
    [[compBatch1], 5] // Group with 2 objects [Alice, Bob], weight 5 -> [Alice, Bob, Alice, Bob, Alice]
  ];
  
  const complexFile = "test_complex_pattern.json";
  const complexCount = streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
    complexBatches, complexFile);
  
  const complexContent = JSON.parse(streamloader.loadText(complexFile));
  
  check(complexCount, {
    'Complex pattern returns correct count': count => count === 5
  });
  
  check(complexContent, {
    'Complex pattern file has correct length': content => content.length === 5,
    'Complex pattern first object': content => content[0].name === "Alice",
    'Complex pattern second object': content => content[1].name === "Bob", 
    'Complex pattern third object': content => content[2].name === "Alice",
    'Complex pattern fourth object': content => content[3].name === "Bob",
    'Complex pattern fifth object': content => content[4].name === "Alice"
  });
  
  console.log(`Complex pattern test: wrote ${complexCount} objects`);
  
  // ---------------------------------------------------------
  // Test 6: Multiple weighted batches
  // ---------------------------------------------------------
  console.log("Test 6: Multiple weighted batches");
  
  const multiBatches = [
    [[compBatch1], 1], // Group with 2 objects -> 1 object (Alice)
    [[compBatch2], 2], // Group with 5 objects -> 2 objects (Charlie, Dave)
    [[compBatch3], 3]  // Group with 1 object -> 3 objects (Henry, Henry, Henry)
  ];
  
  const multiFile = "test_multiple_weighted.json";
  const multiCount = streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
    multiBatches, multiFile);
  
  const multiContent = JSON.parse(streamloader.loadText(multiFile));
  
  check(multiCount, {
    'Multiple batches return correct total count': count => count === 6
  });
  
  check(multiContent, {
    'Multiple batches file has correct length': content => content.length === 6,
    'Multiple batches first is Alice': content => content[0].name === "Alice",
    'Multiple batches second is Charlie': content => content[1].name === "Charlie",
    'Multiple batches third is Dave': content => content[2].name === "Dave",
    'Multiple batches fourth is Henry': content => content[3].name === "Henry",
    'Multiple batches fifth is Henry': content => content[4].name === "Henry",
    'Multiple batches sixth is Henry': content => content[5].name === "Henry"
  });
  
  console.log(`Multiple weighted batches test: wrote ${multiCount} objects`);
  
  // ---------------------------------------------------------
  // Test 7: Zero and negative weights
  // ---------------------------------------------------------
  console.log("Test 7: Zero and negative weights");
  
  const skipBatches = [
    [[compBatch1], 0],  // Should be skipped
    [[compBatch2], -1], // Should be skipped  
    [[compBatch3], 2]   // Should produce 2 Henry objects
  ];
  
  const skipFile = "test_skip_weights.json";
  const skipCount = streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
    skipBatches, skipFile);
  
  const skipContent = JSON.parse(streamloader.loadText(skipFile));
  
  check(skipCount, {
    'Skip invalid weights returns correct count': count => count === 2
  });
  
  check(skipContent, {
    'Skip invalid weights file has correct length': content => content.length === 2,
    'Skip invalid weights both are Henry': content =>
      content.every(obj => obj.name === "Henry" && obj.id === 8)
  });
  
  console.log(`Skip invalid weights test: wrote ${skipCount} objects`);
  
  // ---------------------------------------------------------
  // Test 8: Empty array
  // ---------------------------------------------------------
  console.log("Test 8: Empty array");
  
  const emptyFile = "test_empty_weighted.json";
  const emptyCount = streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
    [], emptyFile);
  
  const emptyContent = streamloader.loadText(emptyFile);
  
  check(emptyCount, {
    'Empty array returns zero count': count => count === 0
  });
  
  check(emptyContent, {
    'Empty array creates empty JSON array': content => content === "[]"
  });
  
  console.log(`Empty array test: wrote ${emptyCount} objects`);
  
  // ---------------------------------------------------------
  // Test 9: Error handling
  // ---------------------------------------------------------
  console.log("Test 9: Error handling");
  
  // Test invalid compressed data
  try {
    streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
      [[["invalid-base64-data!!!"], 2]], "test_error.json");
    check(false, { 'Should have thrown error for invalid data': () => false });
  } catch (e) {
    check(true, { 'Correctly throws error for invalid compressed data': () => true });
    console.log("Correctly handled invalid compressed data");
  }
  
  // Test invalid weight type
  try {
    streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
      [[[compBatch1], "invalid"]], "test_error2.json");
    check(false, { 'Should have thrown error for invalid weight type': () => false });
  } catch (e) {
    check(true, { 'Correctly throws error for invalid weight type': () => true });
    console.log("Correctly handled invalid weight type");
  }
  
  // ---------------------------------------------------------
  // Test 10: Compare with regular function
  // ---------------------------------------------------------
  console.log("Test 10: Compare with regular function");
  
  // Use equal weights to match regular function behavior
  const regularBatches = [compBatch1, compBatch2];
  const weightedEqualBatches = [
    [[compBatch1], 2], // Keep all 2 objects from group
    [[compBatch2], 5]  // Keep all 5 objects from group
  ];
  
  const regularFile = "test_regular_comparison.json";
  const weightedFile = "test_weighted_comparison.json";
  
  const regularCount = streamloader.writeMultipleCompressedJsonLinesToArrayFile(
    regularBatches, regularFile);
  const weightedCount = streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
    weightedEqualBatches, weightedFile);
  
  const regularContent = JSON.parse(streamloader.loadText(regularFile));
  const weightedContent = JSON.parse(streamloader.loadText(weightedFile));
  
  check(true, {
    'Regular and weighted produce same count when weights equal actual counts': () =>
      regularCount === weightedCount,
    'Regular and weighted produce same content length': () =>
      regularContent.length === weightedContent.length,
    'Regular and weighted produce equivalent content': () => {
      if (regularContent.length !== weightedContent.length) return false;
      
      // Since order might be slightly different due to processing, check names match
      const regularNames = regularContent.map(obj => obj.name).sort();
      const weightedNames = weightedContent.map(obj => obj.name).sort();
      
      return regularNames.join(',') === weightedNames.join(',');
    }
  });
  
  console.log(`Comparison test: regular=${regularCount}, weighted=${weightedCount}`);
  
  console.log("All weighted function tests completed successfully!");
}
