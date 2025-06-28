import { check, group } from 'k6';
import { debugCsvOptions } from 'k6/x/streamloader';

// Helper function to print the string representation of an object
function debugObject(obj) {
    console.log(JSON.stringify(obj, null, 2));
}

export default function () {
    group('Struct tags field mapping tests', function () {
        
        // Test 1: Basic required fields only (no optional fields)
        group('Required fields only', function() {
            const requiredOnlyOptions = {
                skipHeader: true,
                filters: [
                    {
                        type: "emptyString",
                        column: 0
                    }
                ],
                transforms: [
                    {
                        type: "parseInt", 
                        column: 1
                    }
                ],
                fields: [
                    {
                        type: "column",
                        column: 0
                    }
                ]
            };
            
            console.log("Testing options with required fields only:");
            debugObject(requiredOnlyOptions);
            
            const result = debugCsvOptions(requiredOnlyOptions);
            console.log("Result:");
            debugObject(result);
            
            // Verify field names in the returned struct
            console.log("Filter struct field names:");
            const filterFields = Object.keys(result.filters[0]);
            filterFields.forEach(field => {
                console.log(`- ${field}: ${result.filters[0][field]}`);
            });
            
            console.log("Transform struct field names:");
            const transformFields = Object.keys(result.transforms[0]);
            transformFields.forEach(field => {
                console.log(`- ${field}: ${result.transforms[0][field]}`);
            });
            
            check(result, {
                'Required fields correctly mapped': (r) => 
                    r.skipHeader === true && 
                    r.filters[0].type === "emptyString" && 
                    r.filters[0].column === 0 &&
                    r.transforms[0].type === "parseInt" && 
                    r.transforms[0].column === 1 &&
                    r.fields[0].type === "column" && 
                    r.fields[0].column === 0,
                'Optional fields present but with default/null values': (r) => 
                    r.filters[0].pattern !== undefined && 
                    r.filters[0].min !== undefined && 
                    r.filters[0].max !== undefined && 
                    r.transforms[0].value !== undefined && 
                    r.transforms[0].start !== undefined && 
                    r.transforms[0].length !== undefined
            });
        });
        
        // Test 2: All field combinations
        group('All field combinations', function() {
            // Create separate objects for each struct type to ensure all fields are handled correctly
            
            // FilterConfig variations
            const allFilterVariations = {
                filters: [
                    {
                        // Complete filter with all fields
                        type: "valueRange",
                        column: 0,
                        pattern: "not used for valueRange but included",
                        min: 10,
                        max: 100
                    },
                    {
                        // Regex filter with pattern but no min/max
                        type: "regexMatch",
                        column: 1,
                        pattern: "pattern.*"
                    },
                    {
                        // Empty string filter - only required fields
                        type: "emptyString",
                        column: 2
                    }
                ]
            };
            
            // TransformConfig variations
            const allTransformVariations = {
                transforms: [
                    {
                        // Complete transform with all fields
                        type: "substring",
                        column: 0,
                        value: "unused for substring",
                        start: 0,
                        length: 5
                    },
                    {
                        // ParseInt - only required fields
                        type: "parseInt",
                        column: 1
                    },
                    {
                        // FixedValue with value
                        type: "fixedValue",
                        column: 2,
                        value: "FIXED"
                    }
                ]
            };
            
            // FieldConfig variations
            const allFieldVariations = {
                fields: [
                    {
                        // Column type
                        type: "column",
                        column: 0
                    },
                    {
                        // Fixed type with value
                        type: "fixed",
                        value: "TEST"
                    }
                ]
            };
            
            // Combined variations
            const combinedOptions = {
                skipHeader: true,
                filters: allFilterVariations.filters,
                transforms: allTransformVariations.transforms,
                groupBy: { column: 3 },
                fields: allFieldVariations.fields
            };
            
            console.log("Testing all field combinations:");
            debugObject(combinedOptions);
            
            const result = debugCsvOptions(combinedOptions);
            console.log("Result:");
            debugObject(result);
            
            check(result, {
                // Filter checks
                'ValueRange filter fields mapped correctly': (r) => 
                    r.filters[0].type === "valueRange" && 
                    r.filters[0].column === 0 && 
                    r.filters[0].pattern === "not used for valueRange but included" && 
                    Number(r.filters[0].min) === 10 && 
                    Number(r.filters[0].max) === 100,
                
                'RegexMatch filter fields mapped correctly': (r) => 
                    r.filters[1].type === "regexMatch" && 
                    r.filters[1].column === 1 && 
                    r.filters[1].pattern === "pattern.*",
                
                'EmptyString filter fields mapped correctly': (r) => 
                    r.filters[2].type === "emptyString" && 
                    r.filters[2].column === 2,
                
                // Transform checks
                'Substring transform fields mapped correctly': (r) => 
                    r.transforms[0].type === "substring" && 
                    r.transforms[0].column === 0 && 
                    r.transforms[0].value === "unused for substring" && 
                    r.transforms[0].start === 0 && 
                    Number(r.transforms[0].length) === 5,
                
                'ParseInt transform fields mapped correctly': (r) => 
                    r.transforms[1].type === "parseInt" && 
                    r.transforms[1].column === 1,
                
                'FixedValue transform fields mapped correctly': (r) => 
                    r.transforms[2].type === "fixedValue" && 
                    r.transforms[2].column === 2 && 
                    r.transforms[2].value === "FIXED",
                
                // Field checks
                'Column field mapped correctly': (r) => 
                    r.fields[0].type === "column" && 
                    r.fields[0].column === 0,
                
                'Fixed field mapped correctly': (r) => 
                    r.fields[1].type === "fixed" && 
                    r.fields[1].value === "TEST",
                
                // GroupBy check
                'GroupBy mapped correctly': (r) => r.groupBy && r.groupBy.column === 3
            });
        });
        
        // Test 3: Field tag name check
        group('Field tag name check', function() {
            // Create minimal options to check field names in the returned struct
            const minimalOptions = {
                skipHeader: true,
                filters: [{ type: "test", column: 1 }],
                transforms: [{ type: "test", column: 2 }],
                groupBy: { column: 3 },
                fields: [{ type: "test" }]
            };
            
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
        });
    });
}