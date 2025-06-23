import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function () {
    // Prepare a test file (in real use, ensure samples.json exists)
    // For demo, we assume samples.json is present
    const filePath = 'samples.json';
    const samples = streamloader.loadSamples(filePath);

    check(samples, {
        'loaded 2 samples': (s) => s.length === 2,
        'first sample is GET': (s) => s[0].method === 'GET',
        'second sample is POST': (s) => s[1].method === 'POST',
    });
} 