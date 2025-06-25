# xk6-streamloader

A k6 extension for efficiently loading large JSON arrays, newline-delimited JSON (NDJSON), top-level JSON objects, and CSV files from disk, using streaming and minimal intermediate memory.

## Features

- **JSON Support**: Load JSON arrays, NDJSON, and JSON objects
- **CSV Support**: Stream CSV files with incremental parsing
- **Memory Efficient**: Minimal memory footprint with streaming architecture
- **Large File Support**: Handle files of any size without memory spikes
- **Robust Parsing**: Handle quoted fields, special characters, and malformed data gracefully

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

## CSV Features

- **Streaming Parser**: Reads CSV files incrementally, one row at a time
- **Memory Efficient**: Uses 64KB buffered reading to minimize memory usage
- **Robust Parsing**: Handles quoted fields, escaped quotes, newlines in fields, and special characters
- **Error Handling**: Detailed error messages with line numbers for parsing issues
- **Whitespace Handling**: Automatically trims leading whitespace from fields
- **Flexible Format**: Supports files with variable number of columns per row

## Files

- `streamloader.go`: Extension source code with JSON and CSV loading functions
- `streamloader_test.go`: Go unit tests for both JSON and CSV functionality
- `streamloader_k6_test.js`: k6 JS test script for both JSON and CSV functionality
- `Makefile`: Build and test automation
- `generate_large_csv.py`: Script to generate large CSV files for testing

### JSON Test Data Files:
- `samples.json`: Basic JSON array with simple objects
- `complex.json`: Complex nested JSON structures with various data types
- `object.json`: Top-level JSON object with key-value pairs
- `bad.json`: Invalid JSON for error testing
- `empty.json`: Empty JSON array
- `large.json`: Large JSON array with 1000 objects

### CSV Test Data Files:
- `basic.csv`: Basic CSV with headers and mixed data types
- `quoted.csv`: CSV with quoted fields, commas, escaped quotes, and newlines
- `empty.csv`: Empty CSV file for edge case testing
- `headers_only.csv`: CSV with only header row
- `malformed.csv`: Malformed CSV for error testing
- `large.csv`: Large CSV with 10,000 rows for memory efficiency testing

## API Reference

### streamloader.loadJSON(filePath)
- **Parameters**: `filePath` (string) - Path to the JSON file
- **Returns**: Array (for JSON arrays/NDJSON) or Object (for JSON objects)
- **Throws**: Error if file not found or JSON is malformed

### streamloader.loadCSV(filePath)
- **Parameters**: `filePath` (string) - Path to the CSV file
- **Returns**: Array of arrays of strings (`[][]string`)
- **Throws**: Error if file not found or CSV is malformed

## Memory Efficiency

Both JSON and CSV loaders are designed for memory efficiency:

- **Streaming Architecture**: Process data incrementally without loading entire files
- **Buffered I/O**: 64KB buffer size for optimal performance
- **Minimal Memory Footprint**: Consistent memory usage regardless of file size
- **No Memory Spikes**: Avoid large memory allocations during processing
