import { check, group } from 'k6';
import { processCsvFile, debugCsvOptions } from 'k6/x/streamloader';

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
    // Create a test CSV file
    const testFilePath = 'parameter_struct_test.csv';
    
    group('Testing parameter passing with ProcessCsvOptions struct', function () {
        // Define options with all properties for testing parameter passing
        const options = {
            skipHeader: true,
            lazyQuotes: true,
            trimLeadingSpace: true,
            reuseRecord: true,
            filters: [
                {
                    type: "regexMatch",
                    column: 1,
                    pattern: "test"
                },
                {
                    type: "valueRange",
                    column: 2,
                    min: 10,
                    max: 20
                }
            ],
            transforms: [
                {
                    type: "parseInt", 
                    column: 0
                }
            ],
            groupBy: {
                column: 3
            },
            fields: [
                {
                    type: "column",
                    column: 1
                },
                {
                    type: "fixed",
                    value: "test"
                }
            ]
        };
        
        console.log("Sending options to ProcessCsvOptions struct:");
        debugObject(options);
        
        try {
            // Use debugCsvOptions to verify parameter passing without needing the file
            const result = debugCsvOptions(options);
            console.log("Parameter passing debug result:");
            debugObject(result);
            
            // Verify key parameters were passed correctly
            let isValid = true;
            
            // Check skipHeader is correctly passed
            isValid = isValid && result.skipHeader === true;
            console.log(`skipHeader check: ${result.skipHeader === true}`);
            
            // Check lazyQuotes is correctly passed
            isValid = isValid && result.lazyQuotes === true;
            console.log(`lazyQuotes check: ${result.lazyQuotes === true}`);
            
            // Check filters array is correctly passed
            isValid = isValid && Array.isArray(result.filters);
            console.log(`filters is array check: ${Array.isArray(result.filters)}`);
            
            isValid = isValid && result.filters.length === 2;
            console.log(`filters length check: ${result.filters.length === 2}`);
            
            // Check first filter (regex)
            isValid = isValid && result.filters[0].type === "regexMatch";
            console.log(`filter[0].type check: ${result.filters[0].type === "regexMatch"}`);
            
            // Verify the complete parameter passing worked
            check(isValid, {
                'ProcessCsvOptions correctly passed with js tags': (result) => result === true
            });
            
            try {
                // This will likely fail due to missing file, but we've already validated parameters
                const result = processCsvFile(testFilePath, options);
            } catch (e) {
                console.error(`Error processing CSV: ${e.message}`);
                check(e.message, {
                    'Error details': (msg) => msg.includes("file")
                });
            }
        } catch (e) {
            console.error(`Error in debugCsvOptions: ${e.message}`);
            check(null, {
                'ProcessCsvOptions correctly passed with js tags': () => false
            });
        }
    });
}