# xk6-streamloader

A k6 extension for efficiently loading large JSON arrays of HTTP traffic samples from disk, using streaming and minimal memory.

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
    const samples = streamloader.loadSamples('samples.json');
    // Use samples in your test logic
}
```

## Files
- `streamloader.go`: Extension source code
- `streamloader_test.go`: Go unit tests
- `streamloader_k6_test.js`: k6 JS test script
- `samples.json`: Example data for tests
- `Makefile`: Build and test automation