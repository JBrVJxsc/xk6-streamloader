import streamloader from 'k6/x/streamloader';
import { check, fail } from 'k6';

export default function () {
    // Tail Tests
    console.log("--- Running Tail Tests ---");

    // 1. Read last 2 lines from test.txt
    const tail2 = streamloader.tail('test.txt', 2);
    console.log("Tail(2) from test.txt:", tail2);
    check(tail2, {
        'tail(2) returns 2 lines': (s) => s.split('\n').length === 2,
        'tail(2) first line correct': (s) => s.split('\n')[0] === 'It contains multiple lines.',
        'tail(2) second line correct': (s) => s.split('\n')[1] === 'And some special characters: !@#$%^&*() ',
    });

    // 2. Read more lines than available
    const tail10 = streamloader.tail('test.txt', 10);
    console.log("Tail(10) from test.txt:", tail10);
    check(tail10, {
        'tail(10) returns all 3 lines': (s) => s.split('\n').length === 3,
        'tail(10) content is correct': (s) => s.includes('This is a test file'),
    });

    // 3. Read 0 lines
    const tail0 = streamloader.tail('test.txt', 0);
    console.log("Tail(0) from test.txt:", tail0);
    check(tail0, {
        'tail(0) returns empty string': (s) => s === '',
    });

    // 4. Read from empty file
    const tailEmpty = streamloader.tail('empty.csv', 5);
    console.log("Tail(5) from empty.csv:", tailEmpty);
    check(tailEmpty, {
        'tail from empty file returns empty string': (s) => !s || s.length === 0,
    });

    // 5. Error case: missing file
    try {
        streamloader.tail('no_such_file.txt', 5);
        fail('Expected error for missing file');
    } catch (e) {
        check(e, { 'tail error for missing file': (err) => String(err).includes('no_such_file') || String(err).includes('no such file') });
    }
} 