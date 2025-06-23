import streamloader from 'k6/x/streamloader';
import { check, fail } from 'k6';

export default function () {
    // 1. Normal case: valid file
    const objects = streamloader.loadJSON('samples.json');
    console.log('objects:', JSON.stringify(objects));
    check(objects, {
        'loaded 2 objects': (s) => s.length === 2,
        'first object is GET': (s) => s[0].method === 'GET',
        'first object URI': (s) => s[0].requestURI === '/foo',
        'first object header': (s) => s[0].headers['A'] === 'B',
        'first object content': (s) => s[0].content === 'abc',
        'second object is POST': (s) => s[1].method === 'POST',
        'second object URI': (s) => s[1].requestURI === '/bar',
        'second object header': (s) => s[1].headers['C'] === 'D',
        'second object content': (s) => s[1].content === 'def',
    });

    // 2. Error case: missing file
    try {
        streamloader.loadJSON('no_such_file.json');
        fail('Expected error for missing file');
    } catch (e) {
        check(e, { 'error for missing file': (err) => String(err).includes('no_such_file') || String(err).includes('no such file') });
    }

    // 3. Error case: invalid JSON
    try {
        streamloader.loadJSON('bad.json');
        fail('Expected error for invalid JSON');
    } catch (e) {
        console.log('Invalid JSON error:', String(e));
        check(e, { 'error for invalid JSON': (err) => String(err).toLowerCase().includes('invalid') });
    }

    // 4. Edge case: empty array
    const emptyObjects = streamloader.loadJSON('empty.json');
    console.log('emptyObjects:', JSON.stringify(emptyObjects));
    check(emptyObjects, { 'empty array returns empty': (s) => Array.isArray(s) && s.length === 0 });

    // 5. Large array
    const largeObjects = streamloader.loadJSON('large.json');
    console.log('largeObjects length:', largeObjects.length);
    console.log('largeObjects[0]:', JSON.stringify(largeObjects[0]));
    console.log('largeObjects[999]:', JSON.stringify(largeObjects[999]));
    check(largeObjects, {
        'large array length': (s) => s.length === 1000,
        'large array first': (s) => s[0].requestURI === '/bulk/0',
        'large array last': (s) => s[999].requestURI === '/bulk/999',
    });
} 