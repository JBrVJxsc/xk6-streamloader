import streamloader from 'k6/x/streamloader';
import { check } from 'k6';

export default function() {
    // Print all available functions/properties
    console.log("Available functions/properties in streamloader:");
    for (let prop in streamloader) {
        console.log(`- ${prop}: ${typeof streamloader[prop]}`);
    }
    
    // Check if our new functions are available
    check(streamloader, {
        'jsonLinesToObjects exists': (sl) => typeof sl.jsonLinesToObjects === 'function',
        'compressedJsonLinesToObjects exists': (sl) => typeof sl.compressedJsonLinesToObjects === 'function'
    });
}