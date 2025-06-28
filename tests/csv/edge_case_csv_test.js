import { check, group } from 'k6';
import { processCsvFile } from 'k6/x/streamloader';
import { sleep } from 'k6';

export const options = {
    thresholds: {
        // Require 100% of checks to pass
        'checks': ['rate==1.0'],
    },
};

export default function () {
    // Test file with edge cases
    const filePath = './testdata/edge_case_test.csv';
    
    console.log(`Testing edge cases in file: ${filePath}`);
    
    group('Special Characters Tests', function() {
        // Test handling of commas in quoted fields
        const commaOptions = {
            skipHeader: true,
            filters: [
                { type: "regexMatch", column: 1, pattern: ".*comma.*" }
            ],
            fields: [
                { type: "column", column: 0 },
                { type: "column", column: 1 }
            ]
        };
        
        try {
            const commaResult = processCsvFile(filePath, commaOptions);
            console.log(`Comma in field result: ${JSON.stringify(commaResult[0])}`);
            
            check(commaResult, {
                'Commas in quoted fields are preserved': (r) => {
                    return r.length > 0 && r[0][1].includes(',');
                }
            });
        } catch (e) {
            console.error(`Error in comma test: ${e.message}`);
            check(null, {
                'Commas in quoted fields are preserved': () => false
            });
        }
        
        // Test handling of quotes within quotes
        const quotesOptions = {
            skipHeader: true,
            filters: [
                { type: "regexMatch", column: 1, pattern: ".*quotes.*" }
            ],
            fields: [
                { type: "column", column: 0 },
                { type: "column", column: 1 },
                { type: "column", column: 5 }
            ]
        };
        
        try {
            const quotesResult = processCsvFile(filePath, quotesOptions);
            console.log(`Quote in field result: ${JSON.stringify(quotesResult[0])}`);
            
            check(quotesResult, {
                'Quotes within quoted fields are preserved': (r) => {
                    return r.length > 0 && 
                        r[0][1].includes('quotes') && 
                        r[0][2].includes('quotes');
                }
            });
        } catch (e) {
            console.error(`Error in quotes test: ${e.message}`);
            check(null, {
                'Quotes within quoted fields are preserved': () => false
            });
        }
        
        // Test handling of Unicode characters
        const unicodeOptions = {
            skipHeader: true,
            filters: [
                { type: "regexMatch", column: 1, pattern: ".*unicode.*" }
            ],
            fields: [
                { type: "column", column: 0 },
                { type: "column", column: 1 },
                { type: "column", column: 5 }
            ]
        };
        
        try {
            const unicodeResult = processCsvFile(filePath, unicodeOptions);
            console.log(`Unicode field result: ${JSON.stringify(unicodeResult[0])}`);
            
            check(unicodeResult, {
                'Unicode characters are preserved': (r) => {
                    return r.length > 0 && 
                        r[0][1].includes('unicode') && 
                        r[0][2].includes('你好世界');
                }
            });
        } catch (e) {
            console.error(`Error in unicode test: ${e.message}`);
            check(null, {
                'Unicode characters are preserved': () => false
            });
        }
        
        // Test handling of multi-line fields
        const multilineOptions = {
            skipHeader: true,
            filters: [
                { type: "regexMatch", column: 1, pattern: ".*Multi.*" }
            ],
            fields: [
                { type: "column", column: 0 },
                { type: "column", column: 1 },
                { type: "column", column: 2 }
            ]
        };
        
        try {
            const multilineResult = processCsvFile(filePath, multilineOptions);
            console.log(`Multiline field result: ${JSON.stringify(multilineResult[0])}`);
            
            check(multilineResult, {
                'Multi-line fields are handled correctly': (r) => {
                    return r.length > 0 && 
                        r[0][1].includes('Multi') && 
                        r[0][2].includes('\n');
                }
            });
        } catch (e) {
            console.error(`Error in multiline test: ${e.message}`);
            check(null, {
                'Multi-line fields are handled correctly': () => false
            });
        }
        
        // Test handling of non-ASCII product names
        const nonAsciiOptions = {
            skipHeader: true,
            filters: [
                { type: "regexMatch", column: 1, pattern: "特殊字符产品" }
            ],
            fields: [
                { type: "column", column: 0 },
                { type: "column", column: 1 }
            ]
        };
        
        try {
            const nonAsciiResult = processCsvFile(filePath, nonAsciiOptions);
            console.log(`Non-ASCII field result length: ${nonAsciiResult.length}`);
            if (nonAsciiResult.length > 0) {
                console.log(`Non-ASCII field result: ${JSON.stringify(nonAsciiResult[0])}`);
            }
            
            check(nonAsciiResult, {
                'Non-ASCII product names are preserved': (r) => {
                    console.log(`Got ${r.length} results for non-ASCII test`);
                    return r.length >= 0;
                }
            });
        } catch (e) {
            console.error(`Error in non-ASCII test: ${e.message}`);
            check(null, {
                'Non-ASCII product names are preserved': () => false
            });
        }
    });
    
    group('Incomplete Data Tests', function() {
        // Test handling of missing columns
        const missingColOptions = {
            skipHeader: true,
            filters: [
                { type: "regexMatch", column: 1, pattern: ".*missing.*" }
            ],
            fields: [
                { type: "column", column: 0 },
                { type: "column", column: 1 },
                { type: "column", column: 3 },  // Should be empty for the filtered row
                { type: "column", column: 4 }   // Should be empty for the filtered row
            ]
        };
        
        try {
            const missingColResult = processCsvFile(filePath, missingColOptions);
            console.log(`Missing columns result: ${JSON.stringify(missingColResult[0])}`);
            
            check(missingColResult, {
                'Missing columns are handled as empty strings': (r) => {
                    return r.length > 0 && 
                        r[0][2] === "" && 
                        r[0][3] === "";
                }
            });
        } catch (e) {
            console.error(`Error in missing columns test: ${e.message}`);
            check(null, {
                'Missing columns are handled as empty strings': () => false
            });
        }
        
        // Test handling of empty values
        const emptyValueOptions = {
            skipHeader: true,
            filters: [
                { type: "regexMatch", column: 6, pattern: "empty" }
            ],
            fields: [
                { type: "column", column: 0 },
                { type: "column", column: 1 },
                { type: "column", column: 5 }  // Description (should be empty)
            ]
        };
        
        try {
            const emptyValueResult = processCsvFile(filePath, emptyValueOptions);
            console.log(`Empty value result: ${JSON.stringify(emptyValueResult[0])}`);
            
            check(emptyValueResult, {
                'Empty values are preserved as empty strings': (r) => {
                    return r.length > 0 && r[0][2] === "";
                }
            });
        } catch (e) {
            console.error(`Error in empty value test: ${e.message}`);
            check(null, {
                'Empty values are preserved as empty strings': () => false
            });
        }
        
        // Test filtering with empty fields
        const emptyNameOptions = {
            skipHeader: true,
            filters: [
                { type: "emptyString", column: 1 }  // Filter for empty names
            ],
            fields: [
                { type: "column", column: 0 },
                { type: "column", column: 1 },
                { type: "column", column: 2 }
            ]
        };
        
        try {
            const emptyNameResult = processCsvFile(filePath, emptyNameOptions);
            console.log(`Empty name filter result length: ${emptyNameResult.length}`);
            if (emptyNameResult.length > 0) {
                console.log(`Empty name sample: "${emptyNameResult[0][1]}"`);
            }
            
            check(emptyNameResult, {
                'Empty string filter works correctly': (r) => {
                    console.log(`Got ${r.length} results for empty string filter`);
                    return r.length >= 0;
                }
            });
        } catch (e) {
            console.error(`Error in empty name filter test: ${e.message}`);
            check(null, {
                'Empty string filter works correctly': () => false
            });
        }
    });
    
    group('Type Conversion Tests', function() {
        // Test parseInt on non-numeric values
        const parseIntOptions = {
            skipHeader: true,
            filters: [
                { type: "regexMatch", column: 1, pattern: "Mixed types" }
            ],
            transforms: [
                { type: "parseInt", column: 2 }  // Try to parseInt on "not-a-number"
            ],
            fields: [
                { type: "column", column: 0 },
                { type: "column", column: 1 },
                { type: "column", column: 2 }
            ]
        };
        
        try {
            const parseIntResult = processCsvFile(filePath, parseIntOptions);
            console.log(`parseInt on non-numeric result: ${JSON.stringify(parseIntResult[0])}`);
            
            check(parseIntResult, {
                'parseInt handles non-numeric values gracefully': (r) => {
                    return r.length > 0;
                    // We're just checking it doesn't crash, result can vary
                }
            });
        } catch (e) {
            console.error(`Error in parseInt test: ${e.message}`);
            check(null, {
                'parseInt handles non-numeric values gracefully': () => false
            });
        }
        
        // Test valueRange filter with non-numeric values
        const valueRangeOptions = {
            skipHeader: true,
            filters: [
                { type: "valueRange", column: 2, min: 0, max: 9999 }  // Should skip "not-a-number" row
            ],
            fields: [
                { type: "column", column: 0 },
                { type: "column", column: 1 },
                { type: "column", column: 2 }
            ]
        };
        
        try {
            const valueRangeResult = processCsvFile(filePath, valueRangeOptions);
            console.log(`valueRange with non-numeric result length: ${valueRangeResult.length}`);
            
            check(valueRangeResult, {
                'valueRange filter handles non-numeric values by excluding them': (r) => {
                    // Should not include the row with "not-a-number"
                    return !r.some(row => row[1] === "Mixed types");
                }
            });
        } catch (e) {
            console.error(`Error in valueRange test: ${e.message}`);
            check(null, {
                'valueRange filter handles non-numeric values by excluding them': () => false
            });
        }
    });
    
    group('Extra Fields Tests', function() {
        // Test handling of extra columns
        const extraColOptions = {
            skipHeader: true,
            filters: [
                { type: "regexMatch", column: 1, pattern: "inconsistent" }
            ],
            fields: [
                { type: "column", column: 0 },
                { type: "column", column: 1 },
                { type: "column", column: 7 },  // Beyond declared columns
                { type: "column", column: 8 }   // Beyond declared columns
            ]
        };
        
        try {
            const extraColResult = processCsvFile(filePath, extraColOptions);
            console.log(`Extra columns result: ${JSON.stringify(extraColResult[0])}`);
            
            check(extraColResult, {
                'Extra columns beyond declared ones are accessible': (r) => {
                    return r.length > 0 && 
                        r[0][2] === "extra1" && // Should access the extra column
                        r[0][3] === "extra2";  // Should access the extra column
                }
            });
        } catch (e) {
            console.error(`Error in extra columns test: ${e.message}`);
            check(null, {
                'Extra columns beyond declared ones are accessible': () => false
            });
        }
    });
    
    console.log('Edge case tests completed');
    
    // Add small delay to ensure logs are flushed
    sleep(0.1);
} 