import { check, group } from 'k6';
import { debugCsvOptions } from 'k6/x/streamloader';

// Helper function to print the string representation of an object
function printObject(obj) {
    console.log(JSON.stringify(obj, null, 2));
}

export default function () {
    group('Testing field tag mapping in ProcessCsvOptions', function () {
        // Test each field individually to pinpoint the issue
        
        // Test 1: Test skipHeader field
        const skipHeaderOptions = {
            skipHeader: true
        };
        
        const returnedSkipHeaderOptions = debugCsvOptions(skipHeaderOptions);
        console.log("skipHeader test:");
        printObject(returnedSkipHeaderOptions);
        
        check(returnedSkipHeaderOptions, {
            'skipHeader is preserved correctly': (obj) => obj.skipHeader === true
        });
        
        // Test 2: Test filters field
        const filtersOptions = {
            filters: [
                { type: "regexMatch", column: 1, pattern: "Test" },
                { type: "valueRange", column: 2, min: 100, max: 200 }
            ]
        };
        
        const returnedFiltersOptions = debugCsvOptions(filtersOptions);
        console.log("filters test:");
        printObject(returnedFiltersOptions);
        
        check(returnedFiltersOptions, {
            'filters array exists': (obj) => Array.isArray(obj.filters),
            'filters length is correct': (obj) => obj.filters.length === 2,
            'filter type is preserved': (obj) => obj.filters[0].type === "regexMatch",
            'filter column is preserved': (obj) => obj.filters[0].column === 1,
            'filter pattern field exists': (obj) => 'pattern,omitempty' in obj.filters[0] || 'pattern' in obj.filters[0],
            'filter min field exists': (obj) => 'min,omitempty' in obj.filters[1] || 'min' in obj.filters[1],
            'filter max field exists': (obj) => 'max,omitempty' in obj.filters[1] || 'max' in obj.filters[1]
        });
        
        // Test 3: Test transforms field
        const transformsOptions = {
            transforms: [
                { type: "parseInt", column: 3 },
                { type: "substring", column: 1, start: 0, length: 5 }
            ]
        };
        
        const returnedTransformsOptions = debugCsvOptions(transformsOptions);
        console.log("transforms test:");
        printObject(returnedTransformsOptions);
        
        check(returnedTransformsOptions, {
            'transforms array exists': (obj) => Array.isArray(obj.transforms),
            'transforms length is correct': (obj) => obj.transforms.length === 2,
            'transform type is preserved': (obj) => obj.transforms[0].type === "parseInt",
            'transform column is preserved': (obj) => obj.transforms[0].column === 3,
            'transform value field exists': (obj) => 'value,omitempty' in obj.transforms[0] || 'value' in obj.transforms[0],
            'transform start field exists': (obj) => 'start,omitempty' in obj.transforms[1] || 'start' in obj.transforms[1],
            'transform length field exists': (obj) => 'length,omitempty' in obj.transforms[1] || 'length' in obj.transforms[1]
        });
        
        // Test 4: Test groupBy field
        const groupByOptions = {
            groupBy: {
                column: 4
            }
        };
        
        const returnedGroupByOptions = debugCsvOptions(groupByOptions);
        console.log("groupBy test:");
        printObject(returnedGroupByOptions);
        
        check(returnedGroupByOptions, {
            'groupBy field exists': (obj) => obj.groupBy !== undefined && obj.groupBy !== null || 'groupBy,omitempty' in obj,
            'groupBy properties are preserved': (obj) => {
                if (obj.groupBy) {
                    return obj.groupBy.column === 4;
                }
                return false;
            }
        });
        
        // Test 5: Test fields field
        const fieldsOptions = {
            fields: [
                { type: "column", column: 0 },
                { type: "fixed", value: "TEST" }
            ]
        };
        
        const returnedFieldsOptions = debugCsvOptions(fieldsOptions);
        console.log("fields test:");
        printObject(returnedFieldsOptions);
        
        check(returnedFieldsOptions, {
            'fields array exists': (obj) => Array.isArray(obj.fields),
            'fields length is correct': (obj) => obj.fields.length === 2,
            'field type is preserved': (obj) => obj.fields[0].type === "column",
            'field column is preserved': (obj) => obj.fields[0].column === 0 || obj.fields[0]['column,omitempty'] === 0,
            'field value field exists': (obj) => 'value,omitempty' in obj.fields[1] || 'value' in obj.fields[1] 
        });
        
        // Detailed output from struct fields
        console.log("\nDetailed field output to help diagnose tag mapping issues:\n");
        
        // Create a minimal object with all required fields
        const minimalOptions = {
            skipHeader: true,
            filters: [{ type: "test", column: 1 }],
            transforms: [{ type: "test", column: 2 }],
            groupBy: { column: 3 },
            fields: [{ type: "test" }]
        };
        
        const returnedMinimalOptions = debugCsvOptions(minimalOptions);
        
        // Log each field and its structure
        console.log("Field names in the returned struct:");
        Object.keys(returnedMinimalOptions).forEach(key => {
            console.log(`- ${key}`);
        });
        
        if (returnedMinimalOptions.filters && returnedMinimalOptions.filters.length > 0) {
            console.log("\nFilter struct field names:");
            Object.keys(returnedMinimalOptions.filters[0]).forEach(key => {
                console.log(`- ${key}`);
            });
        }
        
        if (returnedMinimalOptions.transforms && returnedMinimalOptions.transforms.length > 0) {
            console.log("\nTransform struct field names:");
            Object.keys(returnedMinimalOptions.transforms[0]).forEach(key => {
                console.log(`- ${key}`);
            });
        }
        
        if (returnedMinimalOptions.fields && returnedMinimalOptions.fields.length > 0) {
            console.log("\nField struct field names:");
            Object.keys(returnedMinimalOptions.fields[0]).forEach(key => {
                console.log(`- ${key}`);
            });
        }
    });
}