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
- **Robust Parsing**: Handle quoted fields, special characters, and malformed data gracefully
- **Edge Case Handling**: Graceful handling of Unicode characters, multi-line fields, inconsistent columns, and more

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

Run only Go unit tests:

```sh
make test-go
```

Run only k6 JS tests (requires built k6 binary):

```sh
make test-k6
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
    
    // Read back the files to verify
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

    console.log(`Text content: ${textContent}`);
}
```

### Head (First N Lines)

```js
import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function () {
    // Load the first 10 lines of a file
    const first10Lines = streamloader.head('path/to/your/large_file.txt', 10);
    
    check(first10Lines, {
        'head is not empty': (content) => content.length > 0,
    });

    console.log(`First 10 lines: ${first10Lines}`);
}
```

### Tail (Last N Lines)

```js
import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function () {
    // Load the last 10 lines of a file
    const last10Lines = streamloader.tail('path/to/your/large_file.txt', 10);
    
    check(last10Lines, {
        'tail is not empty': (content) => content.length > 0,
    });

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

#### ProcessCsvFile Options

- **skipHeader** (boolean): Whether to skip the first row as header (default: true)
- **filters** (array): Rules to drop unwanted rows
  - `{ type: "emptyString", column: N }` - Drop rows with empty value in column N
  - `{ type: "regexMatch", column: N, pattern: "regex" }` - Keep only rows where column N matches regex
  - `{ type: "valueRange", column: N, min: X, max: Y }` - Keep only rows where column N is between X and Y
- **transforms** (array): Modify values in-place
  - `{ type: "parseInt", column: N }` - Convert string to integer
  - `{ type: "fixedValue", column: N, value: V }` - Replace with constant value
  - `{ type: "substring", column: N, start: S, length: L }` - Extract substring from column
- **groupBy** (object): Optional grouping
  - `{ column: N }` - Group results by column N
- **fields** (array): Output column selection and projection
  - `{ type: "column", column: N }` - Select column N from input
  - `{ type: "fixed", value: V }` - Output constant value

## Supported formats

### JSON Formats
- **JSON array**: a top-level `[...]` containing objects (returns an array)
- **NDJSON**: one JSON object per line, newline-separated (returns an array)
- **JSON object**: a top-level `{...}` with key-value pairs (returns a plain object)

### CSV Format
- **CSV files**: Comma-separated values with streaming support (returns array of arrays of strings)
- **Quoted fields**: Handles fields with commas, quotes, and newlines
- **Variable columns**: Supports CSV files with inconsistent column counts
- **Unicode support**: Handles special characters and international text
- **Large files**: Memory-efficient processing of files with thousands of rows
- **Empty fields**: Properly handles missing values and empty fields
- **Multi-line fields**: Supports values spanning multiple lines within quotes

## CSV Features

- **Streaming Parser**: Reads CSV files incrementally, one row at a time
- **Memory Efficient**: Uses 64KB buffered reading to minimize memory usage
- **Robust Parsing**: Handles quoted fields, escaped quotes, newlines in fields, and special characters
- **Error Handling**: Detailed error messages with line numbers for parsing issues
- **Whitespace Handling**: Automatically trims leading whitespace from fields
- **Flexible Format**: Supports files with variable number of columns per row
- **Advanced Processing**: Filter, transform, group, and project data in a single pass

## Files

- `streamloader.go`: Extension source code with JSON and CSV loading functions
- `streamloader_test.go`: Go unit tests for both JSON and CSV functionality
- `jsonl_utils_test.go`: Go unit tests for JSON utilities (JSONL and JSON array)
- `compressed_jsonl_test.go`: Go unit tests for compressed JSON utilities
- `streamloader_k6_test.js`: k6 JS test script for both JSON and CSV functionality
- `process_csv_test.js`: Basic k6 JS test script for the ProcessCsvFile function
- `advanced_process_csv_test.js`: Advanced k6 JS tests for ProcessCsvFile with complex configurations
- `edge_case_csv_test.js`: k6 JS tests specifically for CSV edge cases
- `head_test.js`: k6 JS test script for the Head function
- `tail_test.js`: k6 JS test script for the Tail function
- `json_utils_test.js`: k6 JS test script for JSON utilities conversion functions
- `json_roundtrip_test.js`: k6 JS test script for end-to-end JSON utilities testing
- `json_memory_test.js`: k6 JS test script for memory efficiency of JSON utilities
- `compressed_json_test.js`: k6 JS test script for compressed JSON utilities
- `compression_performance_test.js`: k6 JS test script for comparing compression levels and performance
- `Makefile`: Build and test automation
- `generate_large_csv.py`: Script to generate large CSV files for testing
- `generate_large_file.py`: Script to generate large text files for testing
- `generate_large_json.py`: Script to generate large JSON files for testing

### File Test Data Files:
- `test.txt`: Basic text file for `loadFile` testing.
- `empty.txt`: Empty text file for `loadFile` testing.
- `large_file.txt`: Large text file for memory efficiency testing.

### JSON Test Data Files:
- `samples.json`: Basic JSON array with simple objects
- `complex.json`: Complex nested JSON structures with various data types
- `object.json`: Top-level JSON object with key-value pairs
- `bad.json`: Invalid JSON for error testing
- `empty.json`: Empty JSON array
- `large.json`: Large JSON array for memory efficiency testing

### CSV Test Data Files:
- `basic.csv`: Basic CSV with headers and mixed data types
- `quoted.csv`: CSV with quoted fields, commas, escaped quotes, and newlines
- `empty.csv`: Empty CSV file for edge case testing
- `headers_only.csv`: CSV with only header row
- `malformed.csv`: Malformed CSV for error testing
- `large.csv`: Large CSV for memory efficiency testing
- `advanced_process.csv`: CSV file for testing ProcessCsvFile advanced features
- `edge_case_test.csv`: CSV file with various edge cases (Unicode, multi-line fields, etc.)
- `specialchars.csv`: CSV with special characters for robust parsing testing

## API Reference

### streamloader.loadJSON(filePath)
- **Parameters**: `filePath` (string) - Path to the JSON file
- **Returns**: Array (for JSON arrays/NDJSON) or Object (for JSON objects)
- **Throws**: Error if file not found or JSON is malformed

### streamloader.loadFile(filePath)
- **Parameters**: `filePath` (string) - Path to the file
- **Returns**: String containing the entire file content
- **Throws**: Error if file not found or cannot be read

### streamloader.objectsToJsonLines(objects)
- **Parameters**: `objects` (array) - Array of JavaScript objects to convert to JSON lines
- **Returns**: String containing the JSONL representation of the objects
- **Throws**: Error if any object cannot be serialized to JSON
- **Description**: Converts a slice of JavaScript objects into JSONL format (one JSON object per line)

### streamloader.objectsToCompressedJsonLines(objects, [compressionLevel])
- **Parameters**: 
  - `objects` (array) - Array of JavaScript objects to convert to compressed JSON lines
  - `compressionLevel` (int, optional) - Compression level from 0-9 (0=no compression, 1=best speed, 9=best compression, default: -1 which is a compromise between speed and size)
- **Returns**: Base64-encoded string containing the gzip-compressed JSONL data
- **Throws**: Error if any object cannot be serialized to JSON or compression fails
- **Description**: Converts objects to JSON lines, compresses with gzip, and base64-encodes the result

### streamloader.writeJsonLinesToArrayFile(jsonLines, outputFilePath, [bufferSize])
- **Parameters**: 
  - `jsonLines` (string) - JSONL-formatted data with one JSON object per line
  - `outputFilePath` (string) - Path where the JSON array file will be written
  - `bufferSize` (int, optional) - Buffer size in bytes (default: 64KB)
- **Returns**: Number of objects written to the file
- **Throws**: Error if writing fails or JSON is invalid
- **Description**: Takes JSONL data and streams it to a file as a JSON array, minimizing memory usage

### streamloader.writeCompressedJsonLinesToArrayFile(compressedJsonLines, outputFilePath, [bufferSize])
- **Parameters**:
  - `compressedJsonLines` (string) - Base64-encoded, gzip-compressed JSONL data
  - `outputFilePath` (string) - Path where the JSON array file will be written
  - `bufferSize` (int, optional) - Buffer size in bytes (default: 64KB)
- **Returns**: Number of objects written to the file
- **Throws**: Error if decompression fails, writing fails, or JSON is invalid
- **Description**: Decompresses the JSONL data and streams it to a file as a JSON array

### streamloader.writeObjectsToJsonArrayFile(objects, outputFilePath, [bufferSize])
- **Parameters**:
  - `objects` (array) - Array of JavaScript objects to write to the file
  - `outputFilePath` (string) - Path where the JSON array file will be written
  - `bufferSize` (int, optional) - Buffer size in bytes (default: 64KB)
- **Returns**: Number of objects written to the file
- **Throws**: Error if writing fails or JSON is invalid
- **Description**: Convenience function that combines objectsToJsonLines and writeJsonLinesToArrayFile

### streamloader.writeCompressedObjectsToJsonArrayFile(objects, outputFilePath, [compressionLevel])
- **Parameters**:
  - `objects` (array) - Array of JavaScript objects to write to the file
  - `outputFilePath` (string) - Path where the JSON array file will be written
  - `compressionLevel` (int, optional) - Compression level from 0-9 (0=no compression, 1=best speed, 9=best compression, default: -1)
- **Returns**: Number of objects written to the file
- **Throws**: Error if compression fails, writing fails, or JSON is invalid
- **Description**: Compresses objects and streams them to a JSON array file with minimal memory usage

### streamloader.combineJsonArrayFiles(inputFilePaths, outputFilePath, [bufferSize])
- **Parameters**:
  - `inputFilePaths` (array) - Array of paths to JSON array files to combine
  - `outputFilePath` (string) - Path where the combined JSON array file will be written
  - `bufferSize` (int, optional) - Buffer size in bytes (default: 64KB)
- **Returns**: Total number of objects written to the file
- **Throws**: Error if reading or writing fails, or JSON is invalid
- **Description**: Combines multiple JSON array files into a single JSON array file, streaming for memory efficiency

### streamloader.head(filePath, n)
- **Parameters**: 
  - `filePath` (string) - Path to the file
  - `n` (int) - Number of lines to read from the beginning of the file
- **Returns**: String containing the first `n` lines of the file
- **Throws**: Error if file not found or cannot be read

### streamloader.tail(filePath, n)
- **Parameters**:
  - `filePath` (string) - Path to the file
  - `n` (int) - Number of lines to read from the end of the file
- **Returns**: String containing the last `n` lines of the file
- **Throws**: Error if file not found or cannot be read

### streamloader.loadCSV(filePath)
- **Parameters**: `filePath` (string) - Path to the CSV file
- **Returns**: Array of arrays of strings (`[][]string`)
- **Throws**: Error if file not found or CSV is malformed

### streamloader.processCsvFile(filePath, options)
- **Parameters**:
  - `filePath` (string) - Path to the CSV file
  - `options` (object) - Configuration object for processing CSV data:
    - `skipHeader` (boolean) - Whether to skip the first row as header
    - `filters` (array) - Row filtering rules (emptyString, regexMatch, valueRange)
    - `transforms` (array) - Value transformation rules (parseInt, fixedValue, substring)
    - `groupBy` (object) - Optional grouping configuration
    - `fields` (array) - Projection field configurations
- **Returns**: Array of arrays containing processed data, with grouping if specified
- **Throws**: Error if file not found, CSV is malformed, or options contain invalid configurations

## Memory Efficiency

Both JSON and CSV loaders are designed for memory efficiency:

- **Streaming Architecture**: Process data incrementally without loading entire files
- **Buffered I/O**: 64KB buffer size for optimal performance
- **Compression Support**: Gzip compression to reduce memory footprint and file size
- **Minimal Memory Footprint**: Consistent memory usage regardless of file size
- **No Memory Spikes**: Avoid large memory allocations during processing
- **Configurable Compression Levels**: Balance between speed and memory usage
