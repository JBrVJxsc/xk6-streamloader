import { check, group } from 'k6';
import { debugCsvOptions } from 'k6/x/streamloader';

export const options = {
    thresholds: {
        // Require 100% of checks to pass
        'checks': ['rate==1.0'],
    },
};

// Helper function to print the string representation of an object
function debugObject(obj) {
    console.log(JSON.stringify(obj, null, 2));
}

export default function () {
    group('Testing struct tags in Go to JS conversion', function () {
        // Test 1: JSON struct tag names
        group('JSON struct tag names', function() {
            // Create minimal options
            const minimalOptions = {
                skipHeader: true,
                filters: [{ type: "test", column: 1 }],
                transforms: [{ type: "test", column: 2 }],
                groupBy: { column: 3 },
                fields: [{ type: "test" }]
            };
            
            console.log("Testing minimal options structure:");
            debugObject(minimalOptions);
            
            try {
                const result = debugCsvOptions(minimalOptions);
                console.log("Result:");
                debugObject(result);
                
                check(result, {
                    'skipHeader field exists': (r) => r.hasOwnProperty('skipHeader'),
                    'filters field exists': (r) => r.hasOwnProperty('filters'),
                    'transforms field exists': (r) => r.hasOwnProperty('transforms'),
                    'groupBy field exists': (r) => r.hasOwnProperty('groupBy'),
                    'fields field exists': (r) => r.hasOwnProperty('fields')
                });
            } catch (e) {
                console.error(`Error in JSON struct tag test: ${e.message}`);
                check(null, {
                    'JSON struct tags correctly mapped': () => false
                });
            }
        });
        
        // Test 2: Field tag name check
        group('Field tag name check', function() {
            // Create minimal options to check field names in the returned struct
            const minimalOptions = {
                skipHeader: true,
                filters: [{ type: "test", column: 1 }],
                transforms: [{ type: "test", column: 2 }],
                groupBy: { column: 3 },
                fields: [{ type: "test" }]
            };
            
            try {
                const result = debugCsvOptions(minimalOptions);
                
                // Extract field names from returned objects
                const structFieldNames = Object.keys(result);
                const filterFieldNames = Object.keys(result.filters[0]);
                const transformFieldNames = Object.keys(result.transforms[0]);
                const fieldFieldNames = Object.keys(result.fields[0]);
                const groupByFieldNames = Object.keys(result.groupBy);
                
                console.log("\nField names in the returned structs:");
                
                console.log("ProcessCsvOptions struct fields:");
                structFieldNames.forEach(name => console.log(`- ${name}`));
                
                console.log("\nFilterConfig struct fields:");
                filterFieldNames.forEach(name => console.log(`- ${name}`));
                
                console.log("\nTransformConfig struct fields:");
                transformFieldNames.forEach(name => console.log(`- ${name}`));
                
                console.log("\nFieldConfig struct fields:");
                fieldFieldNames.forEach(name => console.log(`- ${name}`));
                
                console.log("\nGroupByConfig struct fields:");
                groupByFieldNames.forEach(name => console.log(`- ${name}`));
                
                // Check field names don't contain "omitempty" suffix
                const noFieldsWithOmitempty = 
                    !structFieldNames.some(name => name.includes("omitempty")) &&
                    !filterFieldNames.some(name => name.includes("omitempty")) &&
                    !transformFieldNames.some(name => name.includes("omitempty")) &&
                    !fieldFieldNames.some(name => name.includes("omitempty")) &&
                    !groupByFieldNames.some(name => name.includes("omitempty"));
                
                console.log(`\nNo fields with 'omitempty' suffix: ${noFieldsWithOmitempty}`);
                
                check(result, {
                    'No fields contain omitempty suffix in js tags': () => noFieldsWithOmitempty,
                    'All expected fields present in FilterConfig': () => 
                        filterFieldNames.includes("type") && 
                        filterFieldNames.includes("column") && 
                        filterFieldNames.includes("pattern") && 
                        filterFieldNames.includes("min") && 
                        filterFieldNames.includes("max"),
                    'All expected fields present in TransformConfig': () => 
                        transformFieldNames.includes("type") && 
                        transformFieldNames.includes("column") && 
                        transformFieldNames.includes("value") && 
                        transformFieldNames.includes("start") && 
                        transformFieldNames.includes("length"),
                    'All expected fields present in FieldConfig': () => 
                        fieldFieldNames.includes("type") && 
                        fieldFieldNames.includes("column") && 
                        fieldFieldNames.includes("value"),
                    'All expected fields present in GroupByConfig': () => 
                        groupByFieldNames.includes("column")
                });
            } catch (e) {
                console.error(`Error in field tag name check: ${e.message}`);
                check(null, {
                    'Field tag names correctly mapped': () => false
                });
            }
        });
    });
}