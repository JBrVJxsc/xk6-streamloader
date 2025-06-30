import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function () {
    // Test base64 error case
    try {
        const result = streamloader.compressedJsonLinesToObjects('!@#$%^&*()_+');
        console.log("Invalid base64 didn't throw!");
        console.log("Result:", result);
    } catch (error) {
        console.log("Base64 error properly thrown:", error.message);
        check(error, {
            'Throws error on invalid base64': (err) => 
                err.toString().includes('failed to decode base64'),
        });
    }

    // Test gzip error case
    try {
        // Create a base64 string that isn't valid gzip
        // Use a hardcoded base64 string instead of btoa (which isn't available in k6)
        const invalidGzipBase64 = 'VGhpcyBpcyBub3QgZ3ppcCBkYXRh';
        const result = streamloader.compressedJsonLinesToObjects(invalidGzipBase64);
        console.log("Invalid gzip didn't throw!");
        console.log("Result:", result);
    } catch (error) {
        console.log("Gzip error properly thrown:", error.message);
        check(error, {
            'Throws error on invalid gzip': (err) => 
                err.toString().includes('failed to create gzip reader') ||
                err.toString().includes('failed to decompress'),
        });
    }

    // Test roundtrip with simple objects
    const objects = [
        { id: 1, name: "Alice" },
        { id: 2, name: "Bob" }
    ];
    
    // Objects to JSON lines roundtrip
    const jsonLines = streamloader.objectsToJsonLines(objects);
    const parsedObjects = streamloader.jsonLinesToObjects(jsonLines);
    
    console.log("Original:", JSON.stringify(objects));
    console.log("Roundtrip:", JSON.stringify(parsedObjects));
    
    check(parsedObjects, {
        'Objects match after JSON lines roundtrip': (arr) => 
            JSON.stringify(arr) === JSON.stringify(objects),
    });
    
    // Objects to compressed JSON lines roundtrip
    const compressedJsonLines = streamloader.objectsToCompressedJsonLines(objects);
    const decompressedObjects = streamloader.compressedJsonLinesToObjects(compressedJsonLines);
    
    console.log("Compressed roundtrip:", JSON.stringify(decompressedObjects));
    
    check(decompressedObjects, {
        'Objects match after compressed roundtrip': (arr) => 
            JSON.stringify(arr) === JSON.stringify(objects),
    });
}