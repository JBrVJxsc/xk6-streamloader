import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function() {
    // Test bidirectional JSON conversion
    const testObjects = [
        { id: 1, name: "Alice", department: "Engineering" },
        { id: 2, name: "Bob", department: "Marketing" },
        {
            id: 3,
            name: "Charlie",
            department: "Sales",
            details: {
                age: 30,
                experience: ["Product", "B2B", "Enterprise"],
                active: true
            }
        }
    ];
    
    console.log("Test objects:", JSON.stringify(testObjects));
    
    // Test regular JSON lines bidirectional conversion
    const jsonLines = streamloader.objectsToJsonLines(testObjects);
    console.log("JSON lines:", jsonLines);
    
    const parsedObjects = streamloader.jsonLinesToObjects(jsonLines);
    console.log("Parsed objects:", JSON.stringify(parsedObjects));
    
    // Debug log for the first object
    if (parsedObjects.length > 0) {
        console.log("First test object:", JSON.stringify(testObjects[0]));
        console.log("First parsed object:", JSON.stringify(parsedObjects[0]));
        
        // Compare properties one by one
        for (const key in testObjects[0]) {
            console.log(`Property ${key}: original=${JSON.stringify(testObjects[0][key])}, parsed=${JSON.stringify(parsedObjects[0][key])}, equal=${JSON.stringify(testObjects[0][key]) === JSON.stringify(parsedObjects[0][key])}`);
        }
    }
    
    // Check with the improved deep comparison
    check(parsedObjects, {
        'Objects match after JSON lines roundtrip (deep check)': (arr) => {
            if (arr.length !== testObjects.length) return false;
            console.log(`Arrays have same length: ${arr.length}`);
            
            for (let i = 0; i < arr.length; i++) {
                const orig = testObjects[i];
                const parsed = arr[i];
                
                // Check each property
                for (const key in orig) {
                    const origVal = JSON.stringify(orig[key]);
                    const parsedVal = JSON.stringify(parsed[key]);
                    
                    if (origVal !== parsedVal) {
                        console.log(`Object ${i} property ${key} mismatch: ${origVal} != ${parsedVal}`);
                        return false;
                    }
                }
                console.log(`Object ${i} properties match`);
            }
            return true;
        }
    });
}