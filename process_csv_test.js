import { check, group } from 'k6';
import { processCsvFile } from 'k6/x/streamloader';

// Helper for deep equality check
function deepEqual(a, b) {
    if (a === b) return true;
    if (a == null || typeof a != "object" || b == null || typeof b != "object") return false;

    let keysA = Object.keys(a), keysB = Object.keys(b);
    if (keysA.length != keysB.length) return false;

    for (let key of keysA) {
        if (!keysB.includes(key) || !deepEqual(a[key], b[key])) return false;
    }
    return true;
}

const filename = "test_process.csv";

export default function () {
    // First, let's see what the function actually returns
    console.log("DEBUG: Reading CSV file with processCsvFile");
    const debugResult = processCsvFile(filename, { skipHeader: true });
    console.log(`DEBUG: Result type: ${typeof debugResult}, length: ${debugResult ? debugResult.length : 'undefined'}`);
    if (debugResult && debugResult.length > 0) {
        console.log(`DEBUG: First row: ${JSON.stringify(debugResult[0])}`);
    }
    
    group('ProcessCsvFile tests', function () {
        group('Basic processing', function () {
            const result = processCsvFile(filename, { skipHeader: true });
            
            // Check if result is an array
            const isArray = Array.isArray(result);
            console.log(`DEBUG: Result is array: ${isArray}`);
            
            check(result, {
                'Basic: result is an array': (r) => Array.isArray(r),
                'Basic: has at least one row': (r) => r && r.length > 0,
            });
            
            // If we have a result, check its structure
            if (result && result.length > 0) {
                check(result[0], {
                    'Basic: first row has expected structure': (row) => 
                        row && row.length >= 4 && 
                        typeof row[0] === 'string' && 
                        typeof row[1] === 'string'
                });
            }
        });

        group('Filtering', function () {
            const result = processCsvFile(filename, {
                skipHeader: true,
                filters: [
                    { type: 'emptyString', column: 1 },
                    { type: 'regexMatch', column: 3, pattern: '^[A-C]$' },
                    { type: 'valueRange', column: 2, min: 200, max: 350 },
                ],
            });
            check(result, {
                'Filter: row count is 1': (r) => r.length === 1,
                'Filter: result contains expected data': (r) => r.length === 1 && r[0].includes('charlie') && r[0].includes('300'),
            });
        });

        group('Transforms', function () {
            const result = processCsvFile(filename, {
                skipHeader: true,
                transforms: [
                    { type: 'fixedValue', column: 1, value: 'processed' },
                    { type: 'substring', column: 3, start: 0, length: 1 },
                ],
                fields: [
                    { type: 'column', column: 0 },
                    { type: 'column', column: 1 },
                    { type: 'column', column: 3 },
                ],
            });

            check(result, {
                'Transform: has expected row count': (r) => r.length > 0,
                'Transform: values contain expected transformations': (r) => 
                    r.some(row => row.includes('processed')) && 
                    r.some(row => row.includes('A') || row.includes('B') || row.includes('C'))
            });
        });

        group('Grouping', function () {
            const result = processCsvFile(filename, {
                skipHeader: true,
                groupBy: { column: 3 },
                fields: [
                    { type: 'column', column: 0 },
                    { type: 'column', column: 1 },
                ],
            });

            check(result, {
                'Grouping: has multiple groups': (r) => r.length > 1,
                'Grouping: contains expected data': (r) => 
                    r.some(group => group.includes('alpha') || group.includes('charlie')) &&
                    r.some(group => group.includes('bravo')) &&
                    r.some(group => group.includes('delta'))
            });
        });

        group('Projection', function () {
            const result = processCsvFile(filename, {
                skipHeader: true,
                fields: [
                    { type: 'column', column: 1 },
                    { type: 'fixed', value: 'static' },
                    { type: 'column', column: 3 },
                ],
            });
            check(result, {
                'Projection: has expected row count': (r) => r.length > 0,
                'Projection: contains projected values': (r) => 
                    r.some(row => row.includes('static')) &&
                    r.some(row => row.includes('alpha') || row.includes('bravo') || row.includes('charlie'))
            });
        });
    });
} 