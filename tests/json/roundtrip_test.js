import streamloader from 'k6/x/streamloader';
import { check, group } from 'k6';
import { deepEqual } from './helper_compare.js';

export default function () {
    group('Bidirectional JSON Conversion Tests', () => {
        // Create test objects
        const testObjects = [
            { id: 1, name: "Alice", department: "Engineering" },
            { id: 2, name: "Bob", department: "Marketing" },
            {
                id: 3,
                name: "Charlie",
                department: "Sales",
                details: {
                    age: 30,
                    experience: ["Product", "B2B", "Enterprise"],
                    active: true
                }
            }
        ];

        // Test regular JSON lines bidirectional conversion
        {
            // Objects to JSON lines
            const jsonLines = streamloader.objectsToJsonLines(testObjects);
            
            // JSON lines back to objects
            const parsedObjects = streamloader.jsonLinesToObjects(jsonLines);
            
            check(parsedObjects, {
                'Objects match after JSON lines roundtrip': (arr) => deepEqual(arr, testObjects)
            });
        }

        // Test compressed JSON lines bidirectional conversion
        {
            // Objects to compressed JSON lines
            const compressedJsonLines = streamloader.objectsToCompressedJsonLines(testObjects);
            
            // Compressed JSON lines back to objects
            const decompressedObjects = streamloader.compressedJsonLinesToObjects(compressedJsonLines);
            
            check(decompressedObjects, {
                'Objects match after compressed JSON lines roundtrip': (arr) => deepEqual(arr, testObjects)
            });
        }

        // Test multi-step conversions (Objects → JSON → Objects → Compressed → Objects)
        {
            // Step 1: Objects to JSON lines
            const step1 = streamloader.objectsToJsonLines(testObjects);
            
            // Step 2: JSON lines back to objects
            const step2 = streamloader.jsonLinesToObjects(step1);
            
            // Step 3: Objects to compressed JSON lines
            const step3 = streamloader.objectsToCompressedJsonLines(step2);
            
            // Step 4: Compressed JSON lines back to objects
            const step4 = streamloader.compressedJsonLinesToObjects(step3);
            
            check(step4, {
                'Objects preserved through multiple conversions': (arr) => deepEqual(arr, testObjects),
                'Array length preserved': (arr) => arr.length === testObjects.length
            });
        }

        // Test with special characters and complex structures
        {
            const specialObjects = [
                {
                    id: 101,
                    description: "Product with \"quotes\" and commas, plus other chars: !@#$%^&*()",
                    html: "<div>Some HTML content with <br/> tags</div>",
                    json: "{\"nested\":\"json string\"}"
                },
                {
                    id: 102,
                    name: "José María Müller",
                    location: "北京市朝阳区",
                    emoji: "😊👍👨‍👩‍👧‍👦" // Emoji and complex Unicode
                },
                {
                    id: 103,
                    nestedArrays: [
                        [1, 2, 3],
                        ["a", "b", "c"],
                        [{ x: 1 }, { y: 2 }]
                    ],
                    deepNesting: {
                        level1: {
                            level2: {
                                level3: {
                                    value: "deep"
                                }
                            }
                        }
                    }
                }
            ];

            // Test normal JSON lines conversion
            const jsonLines = streamloader.objectsToJsonLines(specialObjects);
            const parsedObjects = streamloader.jsonLinesToObjects(jsonLines);
            
            check(parsedObjects, {
                'Special characters preserved in JSON lines roundtrip': (arr) => deepEqual(arr, specialObjects)
            });
            
            // Test compressed conversion
            const compressedJsonLines = streamloader.objectsToCompressedJsonLines(specialObjects);
            const decompressedObjects = streamloader.compressedJsonLinesToObjects(compressedJsonLines);
            
            check(decompressedObjects, {
                'Special characters preserved in compressed roundtrip': (arr) => deepEqual(arr, specialObjects)
            });
        }

        // Test edge cases
        {
            // Empty array
            const emptyArray = [];
            const emptyJsonLines = streamloader.objectsToJsonLines(emptyArray);
            const parsedEmpty = streamloader.jsonLinesToObjects(emptyJsonLines);
            
            check(parsedEmpty, {
                'Empty array roundtrip works': (arr) => 
                    Array.isArray(arr) && arr.length === 0
            });
            
            // Array with null and undefined values (converted to null in JSON)
            const nullishArray = [
                { id: 1, value: null },
                { id: 2, value: 0 },
                { id: 3, value: "" },
                { id: 4, value: false }
            ];
            
            const nullishJsonLines = streamloader.objectsToJsonLines(nullishArray);
            const parsedNullish = streamloader.jsonLinesToObjects(nullishJsonLines);
            
            check(parsedNullish, {
                'Nullish values preserved in roundtrip': (arr) => 
                    arr[0].value === null &&
                    arr[1].value === 0 &&
                    arr[2].value === "" &&
                    arr[3].value === false
            });
        }
    });
}