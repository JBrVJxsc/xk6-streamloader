import streamloader from 'k6/x/streamloader';
import { check, fail } from 'k6';

export default function () {
    // 1. Normal case: valid file
    const samples = streamloader.loadSamples('samples.json');
    console.log('samples:', JSON.stringify(samples));
    check(samples, {
        'loaded 2 samples': (s) => s.length === 2,
        'first sample is GET': (s) => s[0].method === 'GET',
        'first sample URI': (s) => s[0].requestURI === '/foo',
        'first sample header': (s) => s[0].headers['A'] === 'B',
        'first sample content': (s) => s[0].content === 'abc',
        'second sample is POST': (s) => s[1].method === 'POST',
        'second sample URI': (s) => s[1].requestURI === '/bar',
        'second sample header': (s) => s[1].headers['C'] === 'D',
        'second sample content': (s) => s[1].content === 'def',
    });

    // 2. Error case: missing file
    try {
        streamloader.loadSamples('no_such_file.json');
        fail('Expected error for missing file');
    } catch (e) {
        check(e, { 'error for missing file': (err) => String(err).includes('no_such_file') || String(err).includes('no such file') });
    }

    // 3. Error case: invalid JSON
    try {
        streamloader.loadSamples('bad.json');
        fail('Expected error for invalid JSON');
    } catch (e) {
        console.log('Invalid JSON error:', String(e));
        check(e, { 'error for invalid JSON': (err) => String(err).toLowerCase().includes('invalid') });
    }

    // 4. Edge case: empty array
    const emptySamples = streamloader.loadSamples('empty.json');
    console.log('emptySamples:', JSON.stringify(emptySamples));
    check(emptySamples, { 'empty array returns empty': (s) => Array.isArray(s) && s.length === 0 });

    // 5. Large array
    const largeSamples = streamloader.loadSamples('large.json');
    console.log('largeSamples length:', largeSamples.length);
    console.log('largeSamples[0]:', JSON.stringify(largeSamples[0]));
    console.log('largeSamples[999]:', JSON.stringify(largeSamples[999]));
    check(largeSamples, {
        'large array length': (s) => s.length === 1000,
        'large array first': (s) => s[0].requestURI === '/bulk/0',
        'large array last': (s) => s[999].requestURI === '/bulk/999',
    });
} 