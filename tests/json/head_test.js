import streamloader from 'k6/x/streamloader';
import { check, fail } from 'k6';

export const options = {
    thresholds: {
        // Require 100% of checks to pass
        'checks': ['rate==1.0'],
    },
};

export default function () {
    // Head Tests
    console.log("--- Running Head Tests ---");

    // 1. Read first 2 lines from test.txt
    const head2 = streamloader.head('test.txt', 2);
    console.log("Head(2) from test.txt:", head2);
    check(head2, {
        'head(2) returns 2 lines': (s) => s.split('\n').length === 2,
        'head(2) first line correct': (s) => s.split('\n')[0] === 'This is a test file for the loadFile function.',
        'head(2) second line correct': (s) => s.split('\n')[1] === 'It contains multiple lines.',
    });

    // 2. Read more lines than available
    const head10 = streamloader.head('test.txt', 10);
    console.log("Head(10) from test.txt:", head10);
    check(head10, {
        'head(10) returns all 3 lines': (s) => s.split('\n').length === 3,
        'head(10) content is correct': (s) => s.includes('!@#$%^&*()'),
    });

    // 3. Read 0 lines
    const head0 = streamloader.head('test.txt', 0);
    console.log("Head(0) from test.txt:", head0);
    check(head0, {
        'head(0) returns empty string': (s) => s === '',
    });

    // 4. Read from empty file
    const headEmpty = streamloader.head('empty.csv', 5);
    console.log("Head(5) from empty.csv:", headEmpty);
    console.log("Type of headEmpty:", typeof headEmpty, "Length:", headEmpty ? headEmpty.length : 'N/A');
    check(headEmpty, {
        'head from empty file returns empty string': (s) => !s || s.length === 0,
    });

    // 5. Error case: missing file
    try {
        streamloader.head('no_such_file.txt', 5);
        fail('Expected error for missing file');
    } catch (e) {
        check(e, { 'head error for missing file': (err) => String(err).includes('no_such_file') || String(err).includes('no such file') });
    }
}