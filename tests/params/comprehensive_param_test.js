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
    const testFilePath = 'comprehensive_param_test.csv';
    
    group('Testing comprehensive parameter structures', function () {
        // Define a minimal options structure first
        const simpleOptions = {
            skipHeader: true
        };
        
        console.log("Testing simple options structure:");
        debugObject(simpleOptions);
        
        try {
            const result = debugCsvOptions(simpleOptions);
            console.log("Simple result:");
            debugObject(result);
            
            check(result, {
                'Simple struct parameter passed correctly': (r) => r.skipHeader === true
            });
        } catch (e) {
            console.error(`Error in simple options test: ${e.message}`);
            check(null, {
                'Simple struct parameter passed correctly': () => false
            });
        }
        
        // Test with most complex possible options structure
        const minValue = 50;
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
                    column: 0
                },
                {
                    type: "fixedValue",
                    column: 1,
                    value: "FIXED"
                },
                {
                    type: "substring",
                    column: 2,
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
        
        console.log("Comprehensive options structure:");
        debugObject(options);
        
        try {
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
        } catch (e) {
            console.error(`Error in comprehensive options test: ${e.message}`);
            check(null, {
                'Complete ProcessCsvOptions correctly passed': () => false
            });
        }
        
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
            
            try {
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
            } catch (e) {
                console.error(`Error in numeric edge cases test: ${e.message}`);
                check(null, {
                    'Numeric edge cases passed correctly': () => false
                });
            }
        });
    });
}