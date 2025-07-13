# xk6-streamloader

A k6 extension for efficiently loading large JSON arrays, newline-delimited JSON (NDJSON), top-level JSON objects, and CSV files from disk, using streaming and minimal intermediate memory.

## Features

- **JSON Support**: Load JSON arrays, NDJSON, and JSON objects
- **JSON Utilities**: Convert JavaScript objects to JSON lines and vice versa, stream write JSON arrays to files
- **Compression Support**: Gzip compression for JSON lines to reduce memory footprint and file size
- **CSV Support**: Stream CSV files with incremental parsing
- **Advanced CSV Processing**: Filter, transform, group, and project CSV data in a single pass
- **Memory Efficient**: Minimal memory footprint with streaming architecture
- **Large File Support**: Handle files of any size without memory spikes
- **Bidirectional Conversion**: Convert between objects and JSON lines formats in both directions

## Build

Build k6 with this extension using the provided Makefile:

```sh
make build
```

## Test

Run all tests (Go unit tests + k6 JS tests):

```sh
make test
```

## Usage in k6 script

### JSON Loading

```js
import streamloader from 'k6/x/streamloader';

export default function () {
    // Load objects from a standard JSON array, NDJSON file, or top-level object
    const data = streamloader.loadJSON('samples.json');
    // If samples.json is a JSON array:
    // data is an Array of plain JS objects with the original JSON keys
    // e.g. data[0].requestURI, data[0].headers["A"], etc.

    // If loading a top-level object (object.json):
    // {
    //   "user1": { ... },
    //   "user2": { ... }
    // }
    // The result will be a plain JS object:
    // data.user1, data.user2, etc.
}
```

### JSON Utilities

```js
import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function () {
    // Sample objects
    const objects = [
        { id: 1, name: "Alice", details: { age: 30 } },
        { id: 2, name: "Bob", active: true },
        { id: 3, name: "Charlie", tags: ["admin", "user"] }
    ];
    
    // Convert objects to JSON lines (JSONL format)
    const jsonLines = streamloader.objectsToJsonLines(objects);
    console.log("JSON Lines:", jsonLines);
    // Output: {"id":1,"name":"Alice","details":{"age":30}}
    //         {"id":2,"name":"Bob","active":true}
    //         {"id":3,"name":"Charlie","tags":["admin","user"]}
    
    // Write JSON lines to a JSON array file (streaming)
    const outputFile = "data.json";
    const count = streamloader.writeJsonLinesToArrayFile(jsonLines, outputFile);
    console.log(`Wrote ${count} objects to ${outputFile}`);
    
    // Direct write objects to a JSON array file (convenience method)
    const directFile = "direct.json";
    const directCount = streamloader.writeObjectsToJsonArrayFile(objects, directFile);
    console.log(`Directly wrote ${directCount} objects to ${directFile}`);
    
    // Combine multiple JSON array files into one (streaming)
    const combinedFile = "combined.json";
    const combinedCount = streamloader.combineJsonArrayFiles([outputFile, directFile], combinedFile);
    console.log(`Combined ${combinedCount} objects into ${combinedFile}`);
    
    // Read back the file to verify
    const fileContent = streamloader.loadText(outputFile);
    check(fileContent, {
        'File contains valid JSON array': (content) => {
            try {
                const parsed = JSON.parse(content);
                return Array.isArray(parsed) && parsed.length === objects.length;
            } catch (e) {
                return false;
            }
        }
    });
}
```

### JSON Lines Conversion (Both Directions)

```js
import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function () {
    // Sample objects
    const objects = [
        { id: 1, name: "Alice", details: { age: 30 } },
        { id: 2, name: "Bob", active: true },
        { id: 3, name: "Charlie", tags: ["admin", "user"] }
    ];
    
    // Convert objects to JSON lines
    const jsonLines = streamloader.objectsToJsonLines(objects);
    console.log("JSON Lines:", jsonLines);
    // Output: {"id":1,"name":"Alice","details":{"age":30}}
    //         {"id":2,"name":"Bob","active":true}
    //         {"id":3,"name":"Charlie","tags":["admin","user"]}
    
    // Convert JSON lines back to objects
    const parsedObjects = streamloader.jsonLinesToObjects(jsonLines);
    console.log("Parsed Objects:", JSON.stringify(parsedObjects));
    
    check(parsedObjects, {
        'Objects match after roundtrip': (arr) => 
            JSON.stringify(arr) === JSON.stringify(objects)
    });
    
    // Using compression
    const compressedJsonLines = streamloader.objectsToCompressedJsonLines(objects);
    console.log(`Compressed size: ${compressedJsonLines.length} bytes`);
    
    // Convert compressed JSON lines back to objects
    const decompressedObjects = streamloader.compressedJsonLinesToObjects(compressedJsonLines);
    
    check(decompressedObjects, {
        'Objects match after compression roundtrip': (arr) => 
            JSON.stringify(arr) === JSON.stringify(objects)
    });
    
    // Complete roundtrip example:
    // 1. Objects to JSON lines
    // 2. JSON lines to objects
    // 3. Objects to compressed JSON lines
    // 4. Compressed JSON lines to objects
    const step1 = streamloader.objectsToJsonLines(objects);
    const step2 = streamloader.jsonLinesToObjects(step1);
    const step3 = streamloader.objectsToCompressedJsonLines(step2);
    const step4 = streamloader.compressedJsonLinesToObjects(step3);
    
    check(step4, {
        'Objects preserved through multiple conversions': (arr) => 
            JSON.stringify(arr) === JSON.stringify(objects)
    });
}
```

### Compressed JSON Utilities

```js
import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function () {
    // Sample large dataset
    const objects = [];
    for (let i = 0; i < 1000; i++) {
        objects.push({
            id: i,
            name: `User-${i}`,
            email: `user${i}@example.com`,
            profile: {
                age: 20 + (i % 50),
                active: i % 2 === 0,
                joinDate: new Date().toISOString(),
                preferences: {
                    theme: i % 2 === 0 ? 'light' : 'dark',
                    language: 'en-US'
                }
            },
            tags: [`tag-${i % 10}`, `category-${i % 5}`]
        });
    }
    
    // Convert objects to compressed JSON lines (gzipped and base64-encoded)
    // Optional compression level (0-9): 0=no compression, 1=fastest, 9=best compression
    const compressedJsonLines = streamloader.objectsToCompressedJsonLines(objects, 6);
    console.log(`Compressed size: ${compressedJsonLines.length} bytes`);
    
    // Write compressed JSON lines to a JSON array file
    // This decompresses data and writes it as a standard JSON array
    const outputFile = "data.json";
    const count = streamloader.writeCompressedJsonLinesToArrayFile(compressedJsonLines, outputFile);
    console.log(`Wrote ${count} objects to ${outputFile}`);
    
    // Direct write objects to a JSON array file using compression
    // This is a convenience function that combines compression and file writing
    const directFile = "compressed_direct.json";
    const directCount = streamloader.writeCompressedObjectsToJsonArrayFile(objects, directFile);
    console.log(`Directly wrote ${directCount} objects to ${directFile}`);
    
    // Compare compression levels
    const noCompression = streamloader.objectsToCompressedJsonLines(objects, 0).length;
    const bestSpeed = streamloader.objectsToCompressedJsonLines(objects, 1).length;
    const bestCompression = streamloader.objectsToCompressedJsonLines(objects, 9).length;
    
    console.log(`Compression comparison:
    - No compression: ${noCompression} bytes
    - Best speed: ${bestSpeed} bytes
    - Best compression: ${bestCompression} bytes`);
    
    // Convert compressed JSON lines back to objects
    const decompressedObjects = streamloader.compressedJsonLinesToObjects(compressedJsonLines);
    console.log(`Decompressed ${decompressedObjects.length} objects`);
    
    // Verify that decompressed objects match the original
    check(decompressedObjects.length, {
        'Decompressed all objects correctly': (len) => len === objects.length
    });
    
    // Check a sample object to verify correct decompression
    check(decompressedObjects[42], {
        'Sample object matches original': (obj) => 
            obj.id === objects[42].id && 
            obj.name === objects[42].name
    });
}
```

### Working with Multiple JSON Line Batches

```js
import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function () {
    // Scenario: You're processing data in batches (e.g., from an API that returns paginated results)
    // or parallel processing where different workers generate separate batches of data.
    
    // Create multiple batches of objects
    const batch1 = [
        { id: 1, name: "Alice", department: "Engineering" },
        { id: 2, name: "Bob", department: "Marketing" }
    ];
    
    const batch2 = [
        { id: 3, name: "Charlie", department: "Sales", details: { age: 30 } },
        { id: 4, name: "Dave", department: "Finance" }
    ];
    
    // Approach 1: Using uncompressed JSON lines
    // Step 1: Convert each batch to JSON lines
    const jsonLines1 = streamloader.objectsToJsonLines(batch1);
    const jsonLines2 = streamloader.objectsToJsonLines(batch2);
    
    // Step 2: Write multiple JSON line strings to a single JSON array file
    const combinedPath = "combined_uncompressed.json";
    const combinedCount = streamloader.writeMultipleJsonLinesToArrayFile(
        [jsonLines1, jsonLines2], combinedPath);
    console.log(`Wrote ${combinedCount} objects to ${combinedPath}`);
    // Output: [{"id":1,"name":"Alice","department":"Engineering"},{"id":2,"name":"Bob","department":"Marketing"},
    //         {"id":3,"name":"Charlie","department":"Sales","details":{"age":30}},{"id":4,"name":"Dave","department":"Finance"}]
    
    // Approach 2: Using compression for more efficient storage/transfer
    // Step 1: Convert and compress each batch
    const compressedLines1 = streamloader.objectsToCompressedJsonLines(batch1);
    const compressedLines2 = streamloader.objectsToCompressedJsonLines(batch2);
    
    // Step 2: Write multiple compressed JSON line strings to a single JSON array file
    // The function handles decompression and writes a standard JSON array file
    const compressedPath = "combined_from_compressed.json";
    const compressedCount = streamloader.writeMultipleCompressedJsonLinesToArrayFile(
        [compressedLines1, compressedLines2], compressedPath);
    console.log(`Wrote ${compressedCount} objects to ${compressedPath}`);
    
    // Both approaches produce identical JSON files
    const combined1 = JSON.parse(streamloader.loadText(combinedPath));
    const combined2 = JSON.parse(streamloader.loadText(compressedPath));
    
    check(combined1.length, {
        'Both approaches produced the same number of objects': val => val === combined2.length
    });
    
    // Optional: You can specify buffer size for better performance with very large files
    // The buffer size controls how much data is held in memory during file operations
    const largeBufferSize = 256 * 1024; // 256 KB buffer (default is 64 KB)
    
    streamloader.writeMultipleJsonLinesToArrayFile(
        [jsonLines1, jsonLines2], "large_buffer_example.json", largeBufferSize);
    
    // Real-world use case: Process data in chunks to avoid memory spikes
    // This is especially useful for very large datasets that can't fit in memory at once
    function processLargeDatasetInChunks(chunkSize, totalRecords) {
        const compressedBatches = [];
        
        // Simulate processing data in chunks
        for (let offset = 0; offset < totalRecords; offset += chunkSize) {
            // In real code, you would fetch/process each chunk of data here
            const chunk = generateDataChunk(offset, Math.min(chunkSize, totalRecords - offset));
            
            // Compress each chunk to save memory
            const compressedChunk = streamloader.objectsToCompressedJsonLines(chunk);
            compressedBatches.push(compressedChunk);
            
            console.log(`Processed chunk ${offset/chunkSize + 1}, total chunks: ${Math.ceil(totalRecords/chunkSize)}`);
        }
        
        // Combine all chunks into a single file without using much memory
        return streamloader.writeMultipleCompressedJsonLinesToArrayFile(
            compressedBatches, "large_dataset_output.json");
    }
    
    // Approach 3: Convert compressed batches directly to JavaScript objects in memory
    // This is useful when you want to work with the objects directly without writing to a file
    const objectsFromCompressed = streamloader.multipleCompressedJsonLinesToObjects(
        [compressedLines1, compressedLines2]);
    
    console.log(`Decompressed and parsed ${objectsFromCompressed.length} objects into memory`);
    
    // You can now work with the objects directly
    objectsFromCompressed.forEach((obj, index) => {
        console.log(`Object ${index}: ${JSON.stringify(obj)}`);
    });
    
    // Verify the objects match what we expect
    check(objectsFromCompressed, {
        'Correct number of objects decompressed': objects => objects.length === 4,
        'First object has correct structure': objects => 
            objects[0].id === 1 && objects[0].name === "Alice",
        'Last object has correct structure': objects => 
            objects[3].id === 4 && objects[3].name === "Dave"
    });
    
    // Use case: Process compressed data from different sources
    function processCompressedDataFromMultipleSources(compressedSources) {
        // Each source provides compressed JSON lines data
        const allObjects = streamloader.multipleCompressedJsonLinesToObjects(compressedSources);
        
        // Process the combined objects as needed
        const processed = allObjects.map(obj => ({
            ...obj,
            processed: true,
            timestamp: new Date().toISOString()
        }));
        
        return processed;
    }
    
    // Helper to generate sample data
    function generateDataChunk(startId, count) {
        const chunk = [];
        for (let i = 0; i < count; i++) {
            chunk.push({
                id: startId + i,
                name: `User-${startId + i}`,
                timestamp: new Date().toISOString()
            });
        }
        return chunk;
    }
    
    // Example: Process 10,000 records in chunks of 2,500
    // This allows processing very large datasets with minimal memory usage
    const totalCount = processLargeDatasetInChunks(2500, 10000);
    console.log(`Processed and wrote ${totalCount} total records`);
    
    // Alternative approach: Convert compressed batches back to objects for processing
    const decompressed1 = streamloader.compressedJsonLinesToObjects(compressedLines1);
    const decompressed2 = streamloader.compressedJsonLinesToObjects(compressedLines2);
    
    // Now you can work with the decompressed objects directly
    console.log(`Decompressed batch 1: ${decompressed1.length} objects`);
    console.log(`Decompressed batch 2: ${decompressed2.length} objects`);
    
    // ---------------------------------------------------------
    // Approach 4: Weighted sampling for dataset balancing
    // ---------------------------------------------------------
    
    // Use case: Balance datasets by controlling representation of each batch
    // This is useful for machine learning datasets, A/B testing, or statistical sampling
    
    console.log("Weighted sampling examples:");
    
    // Example data with different batch sizes
    const smallBatch = [{id: 1, category: "A"}, {id: 2, category: "A"}]; // 2 objects
    const largeBatch = [ // 5 objects
        {id: 3, category: "B"}, {id: 4, category: "B"}, {id: 5, category: "B"},
        {id: 6, category: "B"}, {id: 7, category: "B"}
    ];
    
    const compressedSmall = streamloader.objectsToCompressedJsonLines(smallBatch);
    const compressedLarge = streamloader.objectsToCompressedJsonLines(largeBatch);
    
    // Scenario 1: Equal representation (3 objects from each category)
    // Each entry is [array of compressed strings, weight]
    const balancedWeights = [
        [[compressedSmall], 3], // Group with 2 objects -> 3 objects: [A1, A2, A1]
        [[compressedLarge], 3]  // Group with 5 objects -> 3 objects: [B3, B4, B5]
    ];
    
    const balancedFile = "balanced_dataset.json";
    const balancedCount = streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
        balancedWeights, balancedFile);
    
    console.log(`Created balanced dataset with ${balancedCount} objects (${balancedCount/2} from each category)`);
    
    // Scenario 2: Proportional sampling (1:2 ratio)
    const proportionalWeights = [
        [[compressedSmall], 2], // Group with 2 objects -> 2 objects: keep both
        [[compressedLarge], 4]  // Group with 5 objects -> 4 objects: keep first 4
    ];
    
    const proportionalFile = "proportional_dataset.json";
    const proportionalCount = streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
        proportionalWeights, proportionalFile);
    
    console.log(`Created proportional dataset with ${proportionalCount} objects (1:2 ratio)`);
    
    // Scenario 3: Oversampling minority class
    const oversampledWeights = [
        [[compressedSmall], 6], // Group with 2 objects -> 6 objects: [A1, A2, A1, A2, A1, A2]
        [[compressedLarge], 5]  // Group with 5 objects -> 5 objects: keep all
    ];
    
    const oversampledFile = "oversampled_dataset.json";
    const oversampledCount = streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
        oversampledWeights, oversampledFile);
    
    console.log(`Created oversampled dataset with ${oversampledCount} objects`);
    
    // Verify the weighted sampling results
    const balancedData = JSON.parse(streamloader.loadText(balancedFile));
    const categoryACounts = balancedData.filter(obj => obj.category === "A").length;
    const categoryBCounts = balancedData.filter(obj => obj.category === "B").length;
    
    console.log(`Balanced dataset: Category A: ${categoryACounts}, Category B: ${categoryBCounts}`);
    
    // Real-world use case: Dataset preparation for ML training
    function prepareTrainingDataset(rawBatches, targetSamplesPerClass) {
        const weightedBatches = rawBatches.map(([data, className]) => {
            const compressed = streamloader.objectsToCompressedJsonLines(data);
            return [[compressed], targetSamplesPerClass]; // Wrap compressed in array
        });
        
        const outputFile = "ml_training_dataset.json";
        const totalSamples = streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(
            weightedBatches, outputFile);
        
        console.log(`Prepared ML dataset: ${totalSamples} samples (${targetSamplesPerClass} per class)`);
        return outputFile;
    }
    
    // Example usage for ML dataset preparation
    const rawDataBatches = [
        [smallBatch, "classA"],
        [largeBatch, "classB"]
    ];
    
    const mlDatasetFile = prepareTrainingDataset(rawDataBatches, 4);
    console.log(`ML training dataset ready: ${mlDatasetFile}`);
}
```

### Text File Loading

```js
import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function () {
    // Load an entire text file as a single string
    const textContent = streamloader.loadText('path/to/your/file.txt');
    
    check(textContent, {
        'file is not empty': (content) => content.length > 0,
    });
}
```

### File Previewing

```js
import streamloader from 'k6/x/streamloader';

export default function () {
    // Load the first 10 lines of a file
    const first10Lines = streamloader.head('path/to/your/large_file.txt', 10);
    console.log(`First 10 lines: ${first10Lines}`);
    
    // Load the last 10 lines of a file
    const last10Lines = streamloader.tail('path/to/your/large_file.txt', 10);
    console.log(`Last 10 lines: ${last10Lines}`);
}
```

### CSV Loading

```js
import streamloader from 'k6/x/streamloader';

export default function () {
    // Load CSV file as array of arrays of strings
    const csvData = streamloader.loadCSV('data.csv');
    
    // csvData[0] contains the first row (typically headers) as array of strings
    // csvData[1] contains the second row as array of strings, etc.
    
    console.log('Headers:', csvData[0]);
    console.log('First data row:', csvData[1]);
    
    // Process each row
    csvData.forEach((row, index) => {
        if (index === 0) {
            console.log('Headers:', row);
        } else {
            console.log(`Row ${index}:`, row);
        }
    });
    
    // Example accessing specific fields (assuming headers: name,age,city,active)
    for (let i = 1; i < csvData.length; i++) {
        const [name, age, city, active] = csvData[i];
        console.log(`Person: ${name}, Age: ${age}, City: ${city}, Active: ${active}`);
    }
}
```

### Advanced CSV Processing

For more advanced CSV processing, you can use `processCsvFile` to filter, transform, group, and project data in a single pass. This is highly memory-efficient for large datasets.

```js
import streamloader from 'k6/x/streamloader';

export default function () {
    const options = {
        skipHeader: true,
        filters: [
            { type: 'emptyString', column: 1 },
            { type: 'regexMatch', column: 3, pattern: '^[A-C]$' },
            { type: 'valueRange', column: 2, min: 200, max: 350 },
        ],
        transforms: [
            { type: 'parseInt', column: 2 },
            { type: 'fixedValue', column: 1, value: 'processed' },
            { type: 'substring', column: 3, start: 0, length: 1 },
        ],
        groupBy: { column: 3 },
        fields: [
            { type: 'column', column: 0 },
            { type: 'column', column: 1 },
            { type: 'fixed', value: 'constant' },
        ],
    };
    const result = streamloader.processCsvFile('data.csv', options);
    // result contains the processed and grouped data
}
```

### CSV Loading with Options

```js
import streamloader from 'k6/x/streamloader';

export default function () {
    // With detailed options
    const options = {
        lazyQuotes: true,          // Allow unescaped quotes in quoted fields
        trimLeadingSpace: true,    // Remove whitespace at the beginning of fields
        trimSpace: false,          // Don't trim all whitespace (leading and trailing)
        reuseRecord: true          // Reuse record memory for better performance
    };
    const csvData = streamloader.loadCSV('data.csv', options);

    // With simple boolean for backward compatibility (lazy quotes)
    const csvDataLazy = streamloader.loadCSV('data.csv', true);
    
    // With default settings (all options true except trimSpace)
    const csvDataDefault = streamloader.loadCSV('data.csv');
}
```

## API Reference

### JSON Functions

#### streamloader.loadJSON(filePath)
- **Parameters**: `filePath` (string) - Path to the JSON file
- **Returns**: Array (for JSON arrays/NDJSON) or Object (for JSON objects)
- **Throws**: Error if file not found or JSON is malformed

#### streamloader.objectsToJsonLines(objects)
- **Parameters**: `objects` (array) - Array of JavaScript objects to convert to JSON lines
- **Returns**: String containing the JSONL representation of the objects
- **Throws**: Error if any object cannot be serialized to JSON

#### streamloader.objectsToCompressedJsonLines(objects, [compressionLevel])
- **Parameters**: 
  - `objects` (array) - Array of JavaScript objects to convert to compressed JSON lines
  - `compressionLevel` (int, optional) - Compression level from 0-9 (0=no compression, 1=best speed, 9=best compression, default: -1)
- **Returns**: Base64-encoded string containing the gzip-compressed JSONL data

#### streamloader.jsonLinesToObjects(jsonLines)
- **Parameters**: `jsonLines` (string) - A string containing JSONL-formatted data, with one JSON object per line
- **Returns**: Array of parsed JavaScript objects
- **Throws**: Error if any line contains invalid JSON

#### streamloader.compressedJsonLinesToObjects(compressedJsonLines)
- **Parameters**: `compressedJsonLines` (string) - A base64-encoded string containing gzip-compressed JSONL data
- **Returns**: Array of parsed JavaScript objects
- **Throws**: Error if decompression fails or any line contains invalid JSON

#### streamloader.writeJsonLinesToArrayFile(jsonLines, outputFilePath, [bufferSize])
- **Parameters**: 
  - `jsonLines` (string) - JSONL-formatted data with one JSON object per line
  - `outputFilePath` (string) - Path where the JSON array file will be written
  - `bufferSize` (int, optional) - Buffer size in bytes (default: 64KB)
- **Returns**: Number of objects written to the file

#### streamloader.writeCompressedJsonLinesToArrayFile(compressedJsonLines, outputFilePath, [bufferSize])
- **Parameters**:
  - `compressedJsonLines` (string) - Base64-encoded, gzip-compressed JSONL data
  - `outputFilePath` (string) - Path where the JSON array file will be written
  - `bufferSize` (int, optional) - Buffer size in bytes (default: 64KB)
- **Returns**: Number of objects written to the file

#### streamloader.writeObjectsToJsonArrayFile(objects, outputFilePath, [bufferSize])
- **Parameters**:
  - `objects` (array) - Array of JavaScript objects to write to the file
  - `outputFilePath` (string) - Path where the JSON array file will be written
  - `bufferSize` (int, optional) - Buffer size in bytes (default: 64KB)
- **Returns**: Number of objects written to the file

#### streamloader.writeCompressedObjectsToJsonArrayFile(objects, outputFilePath, [compressionLevel])
- **Parameters**:
  - `objects` (array) - Array of JavaScript objects to write to the file
  - `outputFilePath` (string) - Path where the JSON array file will be written
  - `compressionLevel` (int, optional) - Compression level from 0-9 (0=no compression, 1=best speed, 9=best compression, default: -1)
- **Returns**: Number of objects written to the file

#### streamloader.combineJsonArrayFiles(inputFilePaths, outputFilePath, [bufferSize])
- **Parameters**:
  - `inputFilePaths` (array) - Array of paths to JSON array files to combine
  - `outputFilePath` (string) - Path where the combined JSON array file will be written
  - `bufferSize` (int, optional) - Buffer size in bytes (default: 64KB)
- **Returns**: Total number of objects written to the file

#### streamloader.writeMultipleJsonLinesToArrayFile(jsonLinesArray, outputFilePath, [bufferSize])
- **Parameters**:
  - `jsonLinesArray` (array) - Array of strings containing JSONL-formatted data
  - `outputFilePath` (string) - Path where the JSON array file will be written
  - `bufferSize` (int, optional) - Buffer size in bytes (default: 64KB)
- **Returns**: Total number of objects written to the file

#### streamloader.writeMultipleCompressedJsonLinesToArrayFile(compressedJsonLinesArray, outputFilePath, [bufferSize])
- **Parameters**:
  - `compressedJsonLinesArray` (array) - Array of base64-encoded, gzip-compressed JSONL strings
  - `outputFilePath` (string) - Path where the JSON array file will be written
  - `bufferSize` (int, optional) - Buffer size in bytes (default: 64KB)
- **Returns**: Total number of objects written to the file

#### streamloader.multipleCompressedJsonLinesToObjects(compressedJsonLinesArray)
- **Parameters**:
  - `compressedJsonLinesArray` (array) - Array of base64-encoded, gzip-compressed JSONL strings
- **Returns**: Array of parsed JavaScript objects from all compressed batches
- **Throws**: Error if decompression fails or any line contains invalid JSON

#### streamloader.writeWeightedMultipleCompressedJsonLinesToArrayFile(weightedMultipleCompressedJsonLinesArray, outputFilePath, [bufferSize])
- **Parameters**:
  - `weightedMultipleCompressedJsonLinesArray` (array) - Array of [multipleCompressedJsonLines, weight] pairs where:
    - `multipleCompressedJsonLines` (array) - array of base64-encoded, gzip-compressed JSONL strings
    - `weight` (number) - target number of objects from this batch group
      - If actual count == weight: keep all objects
      - If actual count > weight: slice to keep only `weight` objects  
      - If actual count < weight: duplicate objects cyclically until count == weight
  - `outputFilePath` (string) - Path where the JSON array file will be written
  - `bufferSize` (int, optional) - Buffer size in bytes (default: 64KB)
- **Returns**: Total number of objects written to the file
- **Throws**: Error if file writing fails, invalid weights, or decompression fails

### File Functions

#### streamloader.loadText(filePath)
- **Parameters**: `filePath` (string) - Path to the file
- **Returns**: String containing the entire file content
- **Throws**: Error if file not found or cannot be read

#### streamloader.head(filePath, n)
- **Parameters**: 
  - `filePath` (string) - Path to the file
  - `n` (int) - Number of lines to read from the beginning of the file
- **Returns**: String containing the first `n` lines of the file

#### streamloader.tail(filePath, n)
- **Parameters**:
  - `filePath` (string) - Path to the file
  - `n` (int) - Number of lines to read from the end of the file
- **Returns**: String containing the last `n` lines of the file

### CSV Functions

#### streamloader.loadCSV(filePath, options)
- **Parameters**: 
  - `filePath` (string) - Path to the CSV file
  - `options` (object or boolean, optional) - CSV parsing options or boolean for lazyQuotes
- **Returns**: Array of arrays of strings (`[][]string`)
- **Throws**: Error if file not found or CSV is malformed

#### streamloader.processCsvFile(filePath, options)
- **Parameters**:
  - `filePath` (string) - Path to the CSV file
  - `options` (object) - Configuration object for processing CSV data:
    - `skipHeader` (boolean) - Whether to skip the first row as header
    - `filters` (array) - Row filtering rules (emptyString, regexMatch, valueRange)
    - `transforms` (array) - Value transformation rules (parseInt, fixedValue, substring)
    - `groupBy` (object) - Optional grouping configuration
    - `fields` (array) - Projection field configurations
- **Returns**: Array of arrays containing processed data, with grouping if specified

## Memory Efficiency

Both JSON and CSV loaders are designed for memory efficiency:

- **Streaming Architecture**: Process data incrementally without loading entire files
- **Buffered I/O**: 64KB buffer size for optimal performance
- **Compression Support**: Gzip compression to reduce memory footprint and file size
- **Bidirectional Conversion**: Convert between objects and JSON lines formats in both directions
- **Minimal Memory Footprint**: Consistent memory usage regardless of file size