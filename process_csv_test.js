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
    group('ProcessCsvFile tests', function () {
        group('Basic processing', function () {
            const result = processCsvFile(filename, { skipHeader: true });
            check(result, {
                'Basic: row count is 5': (r) => r.length === 5,
                'Basic: first row is correct': (r) => deepEqual(r[0], ['1', 'alpha', '100', 'A']),
            });
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
                'Filter: result is correct': (r) => deepEqual(r[0], ['3', 'charlie', '300', 'A']),
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
                'Transform: row count is 5': (r) => r.length === 5,
                'Transform: values are correct': (r) =>
                    deepEqual(r, [
                        ['1', 'processed', 'A'],
                        ['2', 'processed', 'B'],
                        ['3', 'processed', 'A'],
                        ['4', 'processed', 'C'],
                        ['5', 'processed', 'B'],
                    ]),
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
                'Grouping: group count is 3': (r) => r.length === 3,
                'Grouping: group content is correct': (r) => {
                    const groups = {};
                    r.forEach(group => {
                        // Heuristic to identify groups based on flattened structure
                        if (group.includes('alpha') && group.includes('charlie')) groups['A'] = group;
                        else if (group.includes('bravo') && group.includes('')) groups['B'] = group;
                        else if (group.includes('delta')) groups['C'] = group;
                    });
                    return (
                        deepEqual(groups['A'], ['1', 'alpha', '3', 'charlie']) &&
                        deepEqual(groups['B'], ['2', 'bravo', '5', '']) &&
                        deepEqual(groups['C'], ['4', 'delta'])
                    );
                },
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
                'Projection: row count is 5': (r) => r.length === 5,
                'Projection: values are correct': (r) =>
                    deepEqual(r, [
                        ['alpha', 'static', 'A'],
                        ['bravo', 'static', 'B'],
                        ['charlie', 'static', 'A'],
                        ['delta', 'static', 'C'],
                        ['', 'static', 'B'],
                    ]),
            });
        });
    });
} 