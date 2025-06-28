import { check, group } from 'k6';
import { processCsvFile, debugCsvOptions } from 'k6/x/streamloader';

export default function () {
    group('Testing js struct tags for parameter passing', function() {
        // Create a test file path
        const testFilePath = 'parameter_struct_test.csv';
        
        // Create an Order object that matches the struct in the Go code
        const order = {
            id: "12345",
            details: {
                name: "Test Product",
                note: "Testing nested structs",
                more_details: {
                    age: 30
                }
            }
        };
        
        // Log the object for debugging
        console.log("Order object:", JSON.stringify(order, null, 2));
        
        // Define ProcessCsvOptions with standard parameters
        const options = {
            skipHeader: true,
            filters: [
                {
                    type: "emptyString",
                    column: 1
                }
            ],
            transforms: [
                {
                    type: "parseInt",
                    column: 3
                }
            ],
            groupBy: {
                column: 4
            },
            fields: [
                {
                    type: "column",
                    column: 0
                },
                {
                    type: "fixed",
                    value: "FIXED"
                }
            ]
        };
        
        // Log options for verification
        console.log("ProcessCsvOptions object:", JSON.stringify(options, null, 2));
        
        try {
            // Use debugCsvOptions to verify parameter passing without needing the file
            const debugResult = debugCsvOptions(options);
            
            console.log("Parameter passing debug result:");
            console.log(JSON.stringify(debugResult, null, 2));
            
            // Check that options were correctly passed
            check(debugResult, {
                'ProcessCsvOptions correctly passed with js tags': (result) => 
                    result.skipHeader === true && 
                    Array.isArray(result.filters) && 
                    result.filters.length === 1 &&
                    result.filters[0].type === "emptyString" && 
                    result.filters[0].column === 1 &&
                    Array.isArray(result.transforms) &&
                    result.transforms[0].type === "parseInt" &&
                    result.transforms[0].column === 3 &&
                    result.groupBy && result.groupBy.column === 4 &&
                    result.fields[1].value === "FIXED"
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