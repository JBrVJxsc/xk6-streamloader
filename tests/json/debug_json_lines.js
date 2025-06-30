import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function() {
    // Test basic conversion
    const objects = [
        { id: 1, name: "Alice" },
        { id: 2, name: "Bob" }
    ];
    
    // To JSON lines
    const jsonLines = streamloader.objectsToJsonLines(objects);
    console.log("JSON Lines:", jsonLines);
    
    // Back to objects
    const parsedObjects = streamloader.jsonLinesToObjects(jsonLines);
    console.log("Parsed Objects:", JSON.stringify(parsedObjects));
    
    // Compare
    check(parsedObjects, {
        'Count matches': (arr) => arr.length === objects.length,
        'First object id matches': (arr) => arr[0].id === objects[0].id,
        'Second object name matches': (arr) => arr[1].name === objects[1].name,
        'Objects match after roundtrip': (arr) => JSON.stringify(arr) === JSON.stringify(objects)
    });
}