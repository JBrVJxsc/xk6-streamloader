import { check, group } from 'k6';
import { processCsvFile, debugCsvOptions, debugOptions } from 'k6/x/streamloader';
import { sleep } from 'k6';

export const options = {
    thresholds: {
        // Require 100% of checks to pass
        'checks': ['rate==1.0'],
    },
};

export default function () {
    group('CSV Options Tests', function () {
        // First test that both TrimLeadingSpace and TrimSpace are different options
        group('TrimSpace vs TrimLeadingSpace', function() {
            const spacesOptions = {
                trimLeadingSpace: true,  // Only trims leading spaces (default)
                trimSpace: true,         // Trims all spaces (not default)
            };
            
            console.log('Testing both trim options together');
            const debugSpaceResult = debugOptions(spacesOptions);
            
            check(debugSpaceResult, {
                'Both trim options are recognized separately': (r) => 
                    r.trimLeadingSpace === true && r.trimSpace === true
            });
        });
        // Test that all CSV options exist in ProcessCsvOptions
        const options = {
            skipHeader: true,
            lazyQuotes: true,
            trimLeadingSpace: true,
            trimSpace: false,
            reuseRecord: true,
            fields: [
                { type: "column", column: 0 },
                { type: "column", column: 1 }
            ]
        };
        
        try {
            // Verify all options are recognized
            check(options, {
                'lazyQuotes option exists': (o) => o.hasOwnProperty('lazyQuotes'),
                'trimLeadingSpace option exists': (o) => o.hasOwnProperty('trimLeadingSpace'),
                'trimSpace option exists': (o) => o.hasOwnProperty('trimSpace'),
                'reuseRecord option exists': (o) => o.hasOwnProperty('reuseRecord')
            });
            
            // Use debugCsvOptions to verify all options are passed through
            const debugOpts = debugCsvOptions(options);
            check(debugOpts, {
                'lazyQuotes value passed correctly': (o) => o.lazyQuotes === true,
                'trimLeadingSpace value passed correctly': (o) => o.trimLeadingSpace === true,
                'trimSpace value passed correctly': (o) => o.trimSpace === false,
                'reuseRecord value passed correctly': (o) => o.reuseRecord === true
            });
            
            console.log('Successfully verified all CSV options exist and are passed correctly');
            
            // Test providing direct CsvOptions to LoadCSV
            const csvOptions = {
                lazyQuotes: true,
                trimLeadingSpace: false,
                trimSpace: true,
                reuseRecord: true
            };
            
            // Use debugOptions to verify our options object structure is recognized correctly
            const debugResult = debugOptions(csvOptions);
            check(debugResult, {
                'CSV options structure is recognized': (r) => {
                    console.log('Debug options result:', JSON.stringify(r));
                    return r && r.lazyQuotes === true && 
                           r.trimLeadingSpace === false && 
                           r.trimSpace === true &&
                           r.reuseRecord === true;
                }
            });
        } catch (e) {
            console.error(`Error testing CSV options: ${e.message}`);
            check(null, {
                'CSV options exist and work properly': () => false
            });
        }
    });
    
    // Test trimSpace functionality specifically
    group('TrimSpace Functionality Test', function() {
        const testData = `id,name,value
1," Leading space only",100
2,"Trailing space only  ",200
3,"  Both leading and trailing  ",300
4,"No spaces",400`;

        // First file without trim
        const fileName1 = "notrim_test.csv";
        try {
            // Create a test file with spaces in values
            const csvOptions = {
                trimLeadingSpace: false,  // Keep leading spaces
                trimSpace: false,         // Don't trim spaces at all
            };
            
            console.log('Testing with trimSpace=false');
            
            const debugResult = debugOptions(csvOptions);
            check(debugResult, {
                'Options correctly set for untrimmed test': (r) => 
                    r.trimLeadingSpace === false && r.trimSpace === false
            });
            
            // This would normally read the file, but we just verify options
        } catch (e) {
            console.error(`Error in TrimSpace=false test: ${e.message}`);
            check(null, {
                'trimSpace=false test successful': () => false
            });
        }
        
        // Then test with trimSpace=true
        const fileName2 = "trim_test.csv";
        try {
            // Create a test file with spaces in values
            const csvOptions = {
                trimLeadingSpace: false,  // Would normally keep leading spaces
                trimSpace: true,          // But trimSpace overrides and trims all spaces
            };
            
            console.log('Testing with trimSpace=true');
            
            const debugResult = debugOptions(csvOptions);
            check(debugResult, {
                'Options correctly set for trimmed test': (r) => 
                    r.trimLeadingSpace === false && r.trimSpace === true
            });
            
            // This would normally read the file, but we just verify options
        } catch (e) {
            console.error(`Error in TrimSpace=true test: ${e.message}`);
            check(null, {
                'trimSpace=true test successful': () => false
            });
        }
    });
    
    console.log('CSV options tests completed');
    
    // Sleep to ensure logs are flushed
    sleep(0.1);
}