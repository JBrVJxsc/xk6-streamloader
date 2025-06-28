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
        // Test that all CSV options exist in ProcessCsvOptions
        const options = {
            skipHeader: true,
            lazyQuotes: true,
            trimLeadingSpace: true,
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
                'reuseRecord option exists': (o) => o.hasOwnProperty('reuseRecord')
            });
            
            // Use debugCsvOptions to verify all options are passed through
            const debugOpts = debugCsvOptions(options);
            check(debugOpts, {
                'lazyQuotes value passed correctly': (o) => o.lazyQuotes === true,
                'trimLeadingSpace value passed correctly': (o) => o.trimLeadingSpace === true,
                'reuseRecord value passed correctly': (o) => o.reuseRecord === true
            });
            
            console.log('Successfully verified all CSV options exist and are passed correctly');
            
            // Test providing direct CsvOptions to LoadCSV
            const csvOptions = {
                lazyQuotes: true,
                trimLeadingSpace: false,
                reuseRecord: true
            };
            
            // Use debugOptions to verify our options object structure is recognized correctly
            const debugResult = debugOptions(csvOptions);
            check(debugResult, {
                'CSV options structure is recognized': (r) => {
                    console.log('Debug options result:', JSON.stringify(r));
                    return r && r.lazyQuotes === true && 
                           r.trimLeadingSpace === false && 
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
    
    console.log('CSV options tests completed');
    
    // Sleep to ensure logs are flushed
    sleep(0.1);
}