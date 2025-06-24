# xk6-streamloader

A k6 extension for efficiently loading large JSON arrays, newline-delimited JSON (NDJSON), or top-level JSON objects from disk, using streaming and minimal intermediate memory.

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

## Supported formats

- **JSON array**: a top-level `[...]` containing objects (returns an array)
- **NDJSON**: one JSON object per line, newline-separated (returns an array)
- **JSON object**: a top-level `{...}` with key-value pairs (returns a plain object)

## Files

- `streamloader.go`: Extension source code
- `streamloader_test.go`: Go unit tests (now check for plain object for top-level objects)
- `streamloader_k6_test.js`: k6 JS test script (now checks for plain object for top-level objects)
- `Makefile`: Build and test automation
- Example test data files:
  - `samples.json`: Basic JSON array with simple objects
  - `complex.json`: Complex nested JSON structures with various data types
  - `object.json`: Top-level JSON object with key-value pairs
  - `bad.json`: Invalid JSON for error testing
  - `empty.json`: Empty JSON array
  - `large.json`: Large JSON array with 1000 objects
