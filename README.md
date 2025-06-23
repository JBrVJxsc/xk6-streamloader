# xk6-streamloader

A k6 extension for efficiently loading large JSON arrays or newline-delimited JSON (NDJSON) of objects from disk, using streaming and minimal intermediate memory.

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
    // Load objects from a standard JSON array or NDJSON file
    const objects = streamloader.loadJSON('samples.json');
    // objects is an Array of plain JS objects with the original JSON keys
    // e.g. objects[0].requestURI, objects[0].headers["A"], etc.
}
```

## Supported formats

- **JSON array**: a top-level `[...]` containing objects
- **NDJSON**: one JSON object per line, newline-separated

## Files

- `streamloader.go`: Extension source code
- `streamloader_test.go`: Go unit tests
- `streamloader_k6_test.js`: k6 JS test script
- `Makefile`: Build and test automation
- Example test data files (`samples.json`, `bad.json`, `empty.json`, `large.json`) are in the project root
