import { check, group } from 'k6';
import { debugOptions, debugCsvOptions } from 'k6/x/streamloader';

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
    group('Testing null value parameter passing', function () {
        // Test with null in GroupBy
        group('Null values in GroupBy', function() {
            const optionsWithNullGroupBy = {
                skipHeader: true,
                filters: [],
                transforms: [],
                groupBy: null,
                fields: []
            };
            
            console.log("Testing null GroupBy:");
            debugObject(optionsWithNullGroupBy);
            
            try {
                const result = debugCsvOptions(optionsWithNullGroupBy);
                console.log("Result:");
                debugObject(result);
                
                check(result, {
                    'Null GroupBy handled correctly': (r) => r.groupBy === null
                });
            } catch (e) {
                console.error(`Error in null GroupBy test: ${e.message}`);
                check(null, {
                    'Null GroupBy handled correctly': () => false
                });
            }
        });
        
        // Test with null in filters and transforms
        group('Null values in arrays', function() {
            const optionsWithNullArrays = {
                skipHeader: true,
                filters: null,
                transforms: null,
                fields: null
            };
            
            console.log("Testing null arrays:");
            debugObject(optionsWithNullArrays);
            
            try {
                const result = debugCsvOptions(optionsWithNullArrays);
                console.log("Result:");
                debugObject(result);
                
                // Check that null arrays are handled gracefully (probably converted to empty arrays)
                check(result, {
                    'Null filters handled gracefully': (r) => Array.isArray(r.filters),
                    'Null transforms handled gracefully': (r) => Array.isArray(r.transforms),
                    'Null fields handled gracefully': (r) => Array.isArray(r.fields)
                });
            } catch (e) {
                console.error(`Error in null arrays test: ${e.message}`);
                check(null, {
                    'Null arrays handled gracefully': () => false
                });
            }
        });
    });
}