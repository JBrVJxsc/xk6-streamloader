import streamloader from 'k6/x/streamloader';
import { check, group } from 'k6';
import { deepEqual } from './helper_compare.js';

export default function () {
    group('JsonLinesToObjects Function Tests', () => {
        // Test with empty string
        {
            const result = streamloader.jsonLinesToObjects('');
            check(result, {
                'Empty string returns empty array': (arr) => Array.isArray(arr) && arr.length === 0,
            });
        }

        // Test with single object
        {
            const jsonLines = '{"id":1,"name":"Alice"}';
            const result = streamloader.jsonLinesToObjects(jsonLines);
            check(result, {
                'Single object parsed correctly': (arr) => 
                    Array.isArray(arr) && 
                    arr.length === 1 &&
                    arr[0].id === 1 && 
                    arr[0].name === 'Alice',
            });
        }

        // Test with multiple objects
        {
            const jsonLines = '{"id":1,"name":"Alice"}\n{"id":2,"name":"Bob"}\n{"id":3,"name":"Charlie"}';
            const result = streamloader.jsonLinesToObjects(jsonLines);
            check(result, {
                'Multiple objects parsed correctly': (arr) => 
                    Array.isArray(arr) && 
                    arr.length === 3 &&
                    arr[0].name === 'Alice' && 
                    arr[1].name === 'Bob' && 
                    arr[2].name === 'Charlie',
            });
        }

        // Test with blank lines
        {
            const jsonLines = '{"id":1,"name":"Alice"}\n\n{"id":2,"name":"Bob"}\n\n';
            const result = streamloader.jsonLinesToObjects(jsonLines);
            check(result, {
                'Handles blank lines correctly': (arr) => 
                    Array.isArray(arr) && 
                    arr.length === 2 &&
                    arr[0].name === 'Alice' && 
                    arr[1].name === 'Bob',
            });
        }

        // Test with complex objects
        {
            const jsonLines = '{"id":1,"name":"Alice","details":{"age":30,"city":"New York"}}\n' +
                             '{"id":2,"name":"Bob","skills":["Java","Python"]}';
            const result = streamloader.jsonLinesToObjects(jsonLines);
            check(result, {
                'Complex objects parsed correctly': (arr) => 
                    Array.isArray(arr) && 
                    arr.length === 2 &&
                    arr[0].details.age === 30 &&
                    arr[1].skills.length === 2,
            });
        }

        // Test invalid JSON
        try {
            const jsonLines = '{"id":1,"name":"Alice"}\n{this-is-not-json}';
            streamloader.jsonLinesToObjects(jsonLines);
            check(null, {
                'Should throw error on invalid JSON': () => false,
            });
        } catch (error) {
            check(error, {
                'Throws error on invalid JSON': (err) => err.toString().includes('invalid JSON'),
            });
        }
    });

    group('CompressedJsonLinesToObjects Function Tests', () => {
        // Create test objects for compression
        const testObjects = [
            { id: 1, name: 'Alice' },
            { id: 2, name: 'Bob' },
            { id: 3, name: 'Charlie', details: { age: 30 } }
        ];

        // Compress the objects to get compressed JSON lines
        const compressedJsonLines = streamloader.objectsToCompressedJsonLines(testObjects);

        // Test decompression
        {
            const result = streamloader.compressedJsonLinesToObjects(compressedJsonLines);
            check(result, {
                'Compressed data decompressed correctly': (arr) => 
                    Array.isArray(arr) && 
                    arr.length === 3 &&
                    arr[0].id === 1 && 
                    arr[1].id === 2 && 
                    arr[2].id === 3 &&
                    arr[2].details.age === 30,
            });
        }

        // Test with invalid base64
        try {
            streamloader.compressedJsonLinesToObjects('!@#$%^&*()_+');
            check(null, {
                'Should throw error on invalid base64': () => false,
            });
        } catch (error) {
            check(error, {
                'Throws error on invalid base64': (err) => 
                    err.toString().includes('failed to decode base64'),
            });
        }

        // Test with valid base64 but invalid gzip
        try {
            // Create a base64 string that isn't valid gzip
            const invalidGzipBase64 = 'VGhpcyBpcyBub3QgZ3ppcCBkYXRh'; // Base64 for 'This is not gzip data'
            streamloader.compressedJsonLinesToObjects(invalidGzipBase64);
            check(null, {
                'Should throw error on invalid gzip': () => false,
            });
        } catch (error) {
            check(error, {
                'Throws error on invalid gzip': (err) => 
                    err.toString().includes('failed to create gzip reader') ||
                    err.toString().includes('failed to decompress'),
            });
        }
    });

    group('Roundtrip Tests', () => {
        // Test roundtrip from objects to JSON lines and back
        {
            const originalObjects = [
                { id: 1, name: 'Alice' },
                { id: 2, name: 'Bob' },
                { 
                    id: 3, 
                    name: 'Charlie',
                    details: { age: 30, city: 'New York' },
                    skills: ['JavaScript', 'Go', 'Python']
                }
            ];

            const jsonLines = streamloader.objectsToJsonLines(originalObjects);
            const parsedObjects = streamloader.jsonLinesToObjects(jsonLines);

            check(parsedObjects, {
                'Objects match after JSON lines roundtrip': (arr) => deepEqual(arr, originalObjects),
            });

            // Also test compressed roundtrip
            const compressedJsonLines = streamloader.objectsToCompressedJsonLines(originalObjects);
            const decompressedObjects = streamloader.compressedJsonLinesToObjects(compressedJsonLines);

            check(decompressedObjects, {
                'Objects match after compressed roundtrip': (arr) => deepEqual(arr, originalObjects),
            });
        }

        // Test objects with special characters
        {
            const specialObjects = [
                { 
                    id: 1,
                    description: 'Product with "quotes" and commas, plus other chars: !@#$%^&*()',
                    html: '<div>Some HTML content with <br/> tags</div>',
                    json: '{"nested":"json string"}'
                },
                {
                    id: 2,
                    name: 'José María Müller',
                    location: '北京市朝阳区',
                    emoji: '😊👍👨‍👩‍👧‍👦' // Emoji and complex Unicode
                }
            ];

            const jsonLines = streamloader.objectsToJsonLines(specialObjects);
            const parsedObjects = streamloader.jsonLinesToObjects(jsonLines);

            check(parsedObjects, {
                'Special characters preserved in roundtrip': (arr) => {
                    return arr[0].description.includes('"quotes"') && 
                           arr[1].name === 'José María Müller' &&
                           arr[1].location === '北京市朝阳区';
                }
            });
        }

        // Test multi-step roundtrip
        {
            const originalObjects = [
                { id: 1, name: 'Alice' },
                { id: 2, name: 'Bob' }
            ];

            // Objects → JSON lines → objects → compressed JSON lines → objects
            const jsonLines = streamloader.objectsToJsonLines(originalObjects);
            const firstParsed = streamloader.jsonLinesToObjects(jsonLines);
            const compressedJsonLines = streamloader.objectsToCompressedJsonLines(firstParsed);
            const finalObjects = streamloader.compressedJsonLinesToObjects(compressedJsonLines);

            check(finalObjects, {
                'Objects match after multi-step roundtrip': (arr) => 
                    deepEqual(arr, originalObjects),
            });
        }
    });

    group('Integration with Other Functions', () => {
        // Test integration with file writing and reading
        {
            const testObjects = [
                { id: 1, name: 'Alice' },
                { id: 2, name: 'Bob' }
            ];

            // Write objects to a file
            const jsonLines = streamloader.objectsToJsonLines(testObjects);
            const outputFile = 'test_reverse_jsonl.json';
            const count = streamloader.writeJsonLinesToArrayFile(jsonLines, outputFile);

            check(count, {
                'Wrote correct number of objects': (c) => c === 2,
            });

            // Read the file back
            const fileContent = streamloader.loadText(outputFile);
            const fileObjects = JSON.parse(fileContent);

            check(fileObjects, {
                'Objects read back from file correctly': (arr) => 
                    arr.length === 2 &&
                    arr[0].id === 1 && 
                    arr[1].id === 2,
            });

            // Now test with compressed functions
            const compressedJsonLines = streamloader.objectsToCompressedJsonLines(testObjects);
            const compressedFile = 'test_reverse_compressed.json';
            const compressedCount = streamloader.writeCompressedJsonLinesToArrayFile(compressedJsonLines, compressedFile);

            check(compressedCount, {
                'Wrote correct number of compressed objects': (c) => c === 2,
            });

            // Read the file back
            const compressedFileContent = streamloader.loadText(compressedFile);
            const compressedFileObjects = JSON.parse(compressedFileContent);

            check(compressedFileObjects, {
                'Compressed objects read back from file correctly': (arr) => 
                    arr.length === 2 &&
                    arr[0].id === 1 && 
                    arr[1].id === 2,
            });
        }
    });
}