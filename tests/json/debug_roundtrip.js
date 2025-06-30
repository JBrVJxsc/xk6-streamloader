import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function () {
    const originalObjects = [
        { id: 1, name: 'Alice' },
        { id: 2, name: 'Bob' }
    ];

    // Log original objects for inspection
    console.log("Original objects:", JSON.stringify(originalObjects));
    console.log("Original objects type:", typeof originalObjects);
    console.log("originalObjects instanceof Array:", originalObjects instanceof Array);
    console.log("Array.isArray(originalObjects):", Array.isArray(originalObjects));
    
    // Convert to JSON lines
    const jsonLines = streamloader.objectsToJsonLines(originalObjects);
    console.log("JSON lines:", jsonLines);
    
    // Convert back to objects
    const parsedObjects = streamloader.jsonLinesToObjects(jsonLines);
    console.log("Parsed objects:", JSON.stringify(parsedObjects));
    console.log("Parsed objects type:", typeof parsedObjects);
    console.log("parsedObjects instanceof Array:", parsedObjects instanceof Array);
    console.log("Array.isArray(parsedObjects):", Array.isArray(parsedObjects));
    
    // Check string representations
    const originalJson = JSON.stringify(originalObjects);
    const parsedJson = JSON.stringify(parsedObjects);
    console.log("Original JSON:", originalJson);
    console.log("Parsed JSON:", parsedJson);
    console.log("Strings equal:", originalJson === parsedJson);
    
    // Check property by property
    if (parsedObjects.length === originalObjects.length) {
        for (let i = 0; i < originalObjects.length; i++) {
            const originalObj = originalObjects[i];
            const parsedObj = parsedObjects[i];
            console.log(`Object ${i} id equal:`, originalObj.id === parsedObj.id);
            console.log(`Object ${i} name equal:`, originalObj.name === parsedObj.name);
        }
    }
    
    // Final check with loose comparison
    check(parsedObjects, {
        'Objects match after JSON lines roundtrip (loose check)': (arr) => {
            if (!arr || !Array.isArray(arr) || arr.length !== originalObjects.length) {
                return false;
            }
            for (let i = 0; i < originalObjects.length; i++) {
                const orig = originalObjects[i];
                const parsed = arr[i];
                if (orig.id !== parsed.id || orig.name !== parsed.name) {
                    return false;
                }
            }
            return true;
        }
    });
}