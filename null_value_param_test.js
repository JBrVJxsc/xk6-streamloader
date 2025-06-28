import { check, group } from 'k6';
import { debugOptions, debugCsvOptions } from 'k6/x/streamloader';

// Helper function to print the string representation of an object
function debugObject(obj) {
    console.log(JSON.stringify(obj, null, 2));
}

export default function () {
    group('Null/undefined value parameter passing tests', function () {
        
        // Test 1: Optional fields with null/undefined values
        group('Optional fields with null/undefined', function() {
            const testOptions = {
                skipHeader: true,
                filters: [
                    {
                        // Required fields
                        type: "valueRange",
                        column: 2,
                        
                        // Optional fields with explicit null
                        pattern: null,
                        min: null,
                        max: null
                    },
                    {
                        // Required fields
                        type: "regexMatch",
                        column: 1,
                        
                        // Optional field with value
                        pattern: "test",
                        
                        // Leave other optional fields undefined
                    }
                ],
                transforms: [
                    {
                        type: "substring",
                        column: 1,
                        start: 0,
                        
                        // Optional field with null
                        length: null,
                        
                        // Optional field with undefined
                        value: undefined
                    }
                ]
            };
            
            console.log("Testing options with null/undefined fields:");
            debugObject(testOptions);
            
            const result = debugCsvOptions(testOptions);
            console.log("Result:");
            debugObject(result);
            
            check(result, {
                'Required fields preserved': (r) => 
                    r.skipHeader === true && 
                    r.filters[0].type === "valueRange" && 
                    r.filters[0].column === 2,
                'Explicit null values preserved or defaulted correctly': (r) => 
                    r.filters[0].pattern === "" && 
                    r.filters[0].min === null && 
                    r.filters[0].max === null,
                'Pattern value preserved when set': (r) => 
                    r.filters[1].pattern === "test",
                'Optional length field preserved as null when explicitly set': (r) =>
                    r.transforms[0].length === null,
                'Undefined value handled appropriately': (r) =>
                    r.transforms[0].value === null || r.transforms[0].value === undefined
            });
        });
        
        // Test 2: Empty arrays and objects
        group('Empty arrays and objects', function() {
            const emptyCollections = {
                filters: [],
                transforms: [],
                emptyObj: {},
                nested: {
                    emptyArray: [],
                    emptyObject: {}
                }
            };
            
            console.log("Testing empty arrays and objects:");
            debugObject(emptyCollections);
            
            const result = debugOptions(emptyCollections);
            console.log("Result:");
            debugObject(result);
            
            check(result, {
                'Empty filters array preserved': (r) => Array.isArray(r.filters) && r.filters.length === 0,
                'Empty transforms array preserved': (r) => Array.isArray(r.transforms) && r.transforms.length === 0,
                'Empty object preserved': (r) => typeof r.emptyObj === 'object' && Object.keys(r.emptyObj).length === 0,
                'Nested empty array preserved': (r) => Array.isArray(r.nested.emptyArray) && r.nested.emptyArray.length === 0,
                'Nested empty object preserved': (r) => typeof r.nested.emptyObject === 'object' && Object.keys(r.nested.emptyObject).length === 0
            });
        });
        
        // Test 3: Fields with explicit null vs undefined
        group('Explicit null vs undefined', function() {
            // Create an object with some fields explicit null and others undefined
            const options = {
                explicit_null: null,
                // implicit_undefined is not defined
            };
            
            // Verify undefined is not included in JSON representation
            const jsonStr = JSON.stringify(options);
            console.log("JSON string of object with null and undefined fields:");
            console.log(jsonStr);
            
            const hasUndefinedInJson = jsonStr.includes("undefined");
            console.log(`JSON includes 'undefined': ${hasUndefinedInJson}`);
            
            const result = debugOptions(options);
            console.log("Result from debugOptions:");
            debugObject(result);
            
            check(result, {
                'Explicit null preserved': (r) => r.explicit_null === null,
                'Undefined field properly handled': (r) => !('implicit_undefined' in r),
                'JSON did not contain undefined': () => !hasUndefinedInJson
            });
        });
        
        // Test 4: ProcessCsvOptions with null GroupBy
        group('ProcessCsvOptions with null GroupBy', function() {
            const optionsWithNullGroupBy = {
                skipHeader: true,
                filters: [],
                transforms: [],
                groupBy: null,
                fields: []
            };
            
            console.log("Testing options with null GroupBy:");
            debugObject(optionsWithNullGroupBy);
            
            const result = debugCsvOptions(optionsWithNullGroupBy);
            console.log("Result:");
            debugObject(result);
            
            check(result, {
                'Null GroupBy handled correctly': (r) => r.groupBy === null
            });
        });
    });
}