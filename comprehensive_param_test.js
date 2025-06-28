import { check, group } from 'k6';
import { debugOptions, debugCsvOptions } from 'k6/x/streamloader';

// Helper function to print the string representation of an object
function debugObject(obj) {
    console.log(JSON.stringify(obj, null, 2));
}

export default function () {
    group('Comprehensive parameter passing tests', function () {
        
        // Test 1: Basic data types
        group('Basic data types', function() {
            const basicTypes = {
                stringValue: "test string",
                intValue: 42,
                floatValue: 3.14159,
                boolValue: true,
                nullValue: null,
                undefinedCheck: undefined
            };
            
            console.log("Testing basic data types:");
            debugObject(basicTypes);
            
            const result = debugOptions(basicTypes);
            console.log("Result:");
            debugObject(result);
            
            check(result, {
                'String value passed correctly': (r) => r.stringValue === "test string",
                'Integer value passed correctly': (r) => Number(r.intValue) === 42,
                'Float value passed correctly': (r) => Math.abs(Number(r.floatValue) - 3.14159) < 0.0001,
                'Boolean value passed correctly': (r) => r.boolValue === true,
                'Null value preserved': (r) => r.nullValue === null,
                'Undefined handled appropriately': (r) => r.undefinedCheck === undefined || r.undefinedCheck === null
            });
        });
        
        // Test 2: Arrays of different types
        group('Arrays', function() {
            const arrays = {
                emptyArray: [],
                stringArray: ["one", "two", "three"],
                numberArray: [1, 2, 3, 4, 5],
                mixedArray: ["string", 123, true, null, { key: "value" }],
                nestedArray: [
                    [1, 2, 3],
                    [4, 5, 6],
                    ["a", "b", "c"]
                ]
            };
            
            console.log("Testing arrays:");
            debugObject(arrays);
            
            const result = debugOptions(arrays);
            console.log("Result:");
            debugObject(result);
            
            check(result, {
                'Empty array preserved': (r) => Array.isArray(r.emptyArray) && r.emptyArray.length === 0,
                'String array correct length': (r) => Array.isArray(r.stringArray) && r.stringArray.length === 3,
                'String array values preserved': (r) => r.stringArray[0] === "one" && r.stringArray[2] === "three",
                'Number array correct length': (r) => Array.isArray(r.numberArray) && r.numberArray.length === 5,
                'Number array values preserved': (r) => Number(r.numberArray[0]) === 1 && Number(r.numberArray[4]) === 5,
                'Mixed array length preserved': (r) => Array.isArray(r.mixedArray) && r.mixedArray.length === 5,
                'Mixed array primitive values preserved': (r) => 
                    r.mixedArray[0] === "string" && 
                    Number(r.mixedArray[1]) === 123 && 
                    r.mixedArray[2] === true && 
                    r.mixedArray[3] === null,
                'Mixed array object preserved': (r) => r.mixedArray[4] && r.mixedArray[4].key === "value",
                'Nested array structure preserved': (r) => 
                    Array.isArray(r.nestedArray) && 
                    r.nestedArray.length === 3 &&
                    Array.isArray(r.nestedArray[0]) && 
                    r.nestedArray[0].length === 3
            });
        });
        
        // Test 3: Complex nested objects
        group('Complex nested objects', function() {
            const complexObject = {
                id: "complex-1",
                metadata: {
                    created: "2025-01-15T12:00:00Z",
                    modified: "2025-06-27T10:30:00Z",
                    tags: ["test", "parameter", "passing"],
                    config: {
                        isEnabled: true,
                        timeout: 30000,
                        retries: 3,
                        options: {
                            caching: false,
                            logging: {
                                level: "info",
                                format: "json"
                            }
                        }
                    }
                },
                data: [
                    {
                        name: "Item 1",
                        value: 100,
                        properties: {
                            color: "red",
                            size: "medium"
                        }
                    },
                    {
                        name: "Item 2",
                        value: 200,
                        properties: {
                            color: "blue",
                            size: "large"
                        }
                    }
                ]
            };
            
            console.log("Testing complex nested objects:");
            debugObject(complexObject);
            
            const result = debugOptions(complexObject);
            console.log("Result:");
            debugObject(result);
            
            check(result, {
                'Top level property preserved': (r) => r.id === "complex-1",
                'First level nested object preserved': (r) => !!r.metadata && typeof r.metadata === 'object',
                'Second level nested property preserved': (r) => r.metadata.created === "2025-01-15T12:00:00Z",
                'Nested array preserved': (r) => Array.isArray(r.metadata.tags) && r.metadata.tags.length === 3,
                'Nested array values preserved': (r) => r.metadata.tags[1] === "parameter",
                'Third level nested object preserved': (r) => !!r.metadata.config && typeof r.metadata.config === 'object',
                'Third level nested property preserved': (r) => r.metadata.config.timeout === 30000,
                'Fourth level nested object preserved': (r) => !!r.metadata.config.options && typeof r.metadata.config.options === 'object',
                'Fifth level nested object preserved': (r) => !!r.metadata.config.options.logging && typeof r.metadata.config.options.logging === 'object',
                'Fifth level nested property preserved': (r) => r.metadata.config.options.logging.format === "json",
                'Complex array preserved': (r) => Array.isArray(r.data) && r.data.length === 2,
                'Object in array preserved': (r) => !!r.data[0] && r.data[0].name === "Item 1",
                'Nested object in array preserved': (r) => !!r.data[1].properties && r.data[1].properties.color === "blue"
            });
        });
        
        // Test 4: Edge cases
        group('Edge cases', function() {
            const edgeCases = {
                emptyString: "",
                zeroValue: 0,
                negativeValue: -10,
                maxIntValue: Number.MAX_SAFE_INTEGER,
                minIntValue: Number.MIN_SAFE_INTEGER,
                infinityValue: Infinity,
                nanValue: NaN,
                boolFalse: false,
                emptyObject: {},
                specialChars: "!@#$%^&*()_+{}|:\"<>?[]\\;',./~`",
                unicodeChars: "你好世界 مرحبا بالعالم こんにちは世界",
                longString: "a".repeat(10000)
            };
            
            console.log("Testing edge cases:");
            // Skip logging the full object due to long string
            console.log("(Edge case object contains very long string, not printing full object)");
            
            const result = debugOptions(edgeCases);
            console.log("Result contains edge cases, not printing full result");
            
            check(result, {
                'Empty string preserved': (r) => r.emptyString === "",
                'Zero value preserved': (r) => Number(r.zeroValue) === 0,
                'Negative value preserved': (r) => Number(r.negativeValue) === -10,
                'MAX_SAFE_INTEGER handled correctly': (r) => Math.abs(Number(r.maxIntValue) - Number.MAX_SAFE_INTEGER) < 1,
                'MIN_SAFE_INTEGER handled correctly': (r) => Math.abs(Number(r.minIntValue) - Number.MIN_SAFE_INTEGER) < 1,
                'Boolean false preserved': (r) => r.boolFalse === false,
                'Empty object preserved': (r) => typeof r.emptyObject === 'object' && Object.keys(r.emptyObject).length === 0,
                'Special characters preserved': (r) => r.specialChars === "!@#$%^&*()_+{}|:\"<>?[]\\;',./~`",
                'Unicode characters preserved': (r) => r.unicodeChars === "你好世界 مرحبا بالعالم こんにちは世界",
                'Long string handled correctly': (r) => r.longString && r.longString.length === 10000
            });
        });
        
        // Test 5: ProcessCsvOptions with all possible parameters
        group('Complete ProcessCsvOptions', function() {
            const minValue = 10;
            const maxValue = 200;
            const options = {
                skipHeader: true,
                filters: [
                    {
                        type: "emptyString",
                        column: 0
                    },
                    {
                        type: "regexMatch",
                        column: 1,
                        pattern: "^Product.*"
                    },
                    {
                        type: "valueRange",
                        column: 2,
                        min: minValue,
                        max: maxValue
                    }
                ],
                transforms: [
                    {
                        type: "parseInt",
                        column: 3
                    },
                    {
                        type: "fixedValue",
                        column: 4,
                        value: "FIXED"
                    },
                    {
                        type: "substring",
                        column: 1,
                        start: 0,
                        length: 5
                    }
                ],
                groupBy: {
                    column: 5
                },
                fields: [
                    {
                        type: "column",
                        column: 0
                    },
                    {
                        type: "fixed",
                        value: "TEST"
                    }
                ]
            };
            
            console.log("Testing complete ProcessCsvOptions:");
            debugObject(options);
            
            const result = debugCsvOptions(options);
            console.log("Result:");
            debugObject(result);
            
            let isValid = true;
            
            // Check skipHeader is correctly passed
            isValid = isValid && result.skipHeader === true;
            console.log(`skipHeader check: ${result.skipHeader === true}`);
            
            // Check filters are correctly passed
            isValid = isValid && Array.isArray(result.filters);
            console.log(`filters is array: ${Array.isArray(result.filters)}`);
            
            isValid = isValid && result.filters.length === 3;
            console.log(`filters length = 3: ${result.filters.length === 3}`);
            
            // Check first filter (emptyString)
            isValid = isValid && result.filters[0].type === "emptyString";
            isValid = isValid && result.filters[0].column === 0;
            console.log(`filter[0] type and column: ${result.filters[0].type === "emptyString" && result.filters[0].column === 0}`);
            
            // Check second filter (regexMatch)
            isValid = isValid && result.filters[1].type === "regexMatch";
            isValid = isValid && result.filters[1].column === 1;
            isValid = isValid && result.filters[1].pattern === "^Product.*";
            console.log(`filter[1] regex pattern: ${result.filters[1].pattern === "^Product.*" ? "correct" : result.filters[1].pattern}`);
            
            // Check third filter (valueRange)
            isValid = isValid && result.filters[2].type === "valueRange";
            isValid = isValid && result.filters[2].column === 2;
            isValid = isValid && Number(result.filters[2].min) === minValue;
            isValid = isValid && Number(result.filters[2].max) === maxValue;
            console.log(`filter[2] min/max values: min=${result.filters[2].min} (${minValue}), max=${result.filters[2].max} (${maxValue})`);
            
            // Check transforms
            isValid = isValid && Array.isArray(result.transforms);
            isValid = isValid && result.transforms.length === 3;
            
            // Check each transform
            isValid = isValid && result.transforms[0].type === "parseInt";
            isValid = isValid && result.transforms[1].type === "fixedValue";
            isValid = isValid && result.transforms[1].value === "FIXED";
            isValid = isValid && result.transforms[2].type === "substring";
            isValid = isValid && result.transforms[2].start === 0;
            isValid = isValid && Number(result.transforms[2].length) === 5;
            
            // Check groupBy
            isValid = isValid && result.groupBy && result.groupBy.column === 5;
            console.log(`groupBy column = 5: ${result.groupBy && result.groupBy.column === 5}`);
            
            // Check fields
            isValid = isValid && Array.isArray(result.fields);
            isValid = isValid && result.fields.length === 2;
            isValid = isValid && result.fields[0].type === "column";
            isValid = isValid && result.fields[0].column === 0;
            isValid = isValid && result.fields[1].type === "fixed";
            isValid = isValid && result.fields[1].value === "TEST";
            
            console.log(`ProcessCsvOptions validation: ${isValid ? "PASSED" : "FAILED"}`);
            
            check(isValid, {
                'Complete ProcessCsvOptions correctly passed': () => isValid
            });
        });
        
        // Test 6: Special case - large numeric values and floating points
        group('Numeric edge cases', function() {
            const numericEdgeCases = {
                filters: [
                    {
                        type: "valueRange",
                        column: 0,
                        min: 0.000000001,
                        max: 9999999999.999999
                    },
                    {
                        type: "valueRange",
                        column: 1,
                        min: Number.MAX_SAFE_INTEGER - 10,
                        max: Number.MAX_SAFE_INTEGER
                    }
                ]
            };
            
            console.log("Testing numeric edge cases:");
            debugObject(numericEdgeCases);
            
            const result = debugCsvOptions(numericEdgeCases);
            console.log("Result:");
            debugObject(result);
            
            // For floating point, check that they're very close
            const isSmallFloatCorrect = Math.abs(Number(result.filters[0].min) - 0.000000001) < 0.0000000001;
            const isLargeFloatCorrect = Math.abs(Number(result.filters[0].max) - 9999999999.999999) < 0.0001;
            const isLargeIntCorrect = Number(result.filters[1].max) > Number.MAX_SAFE_INTEGER - 100;
            
            console.log(`Small float check: ${isSmallFloatCorrect} (${result.filters[0].min})`);
            console.log(`Large float check: ${isLargeFloatCorrect} (${result.filters[0].max})`);
            console.log(`Large int check: ${isLargeIntCorrect} (${result.filters[1].max})`);
            
            check(result, {
                'Very small float preserved': () => isSmallFloatCorrect,
                'Very large float preserved': () => isLargeFloatCorrect,
                'Very large int handled correctly': () => isLargeIntCorrect
            });
        });
    });
}