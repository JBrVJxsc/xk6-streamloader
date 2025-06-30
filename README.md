# xk6-streamloader

A k6 extension for efficiently loading large JSON arrays, newline-delimited JSON (NDJSON), top-level JSON objects, and CSV files from disk, using streaming and minimal intermediate memory.

## Features

- **JSON Support**: Load JSON arrays, NDJSON, and JSON objects
- **JSON Utilities**: Convert JavaScript objects to JSON lines and stream write JSON arrays to files
- **Compression Support**: Gzip compression for JSON lines to reduce memory footprint and file size
- **CSV Support**: Stream CSV files with incremental parsing
- **Advanced CSV Processing**: Filter, transform, group, and project CSV data in a single pass
- **Memory Efficient**: Minimal memory footprint with streaming architecture
- **Large File Support**: Handle files of any size without memory spikes

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
- **Minimal Memory Footprint**: Consistent memory usage regardless of file size