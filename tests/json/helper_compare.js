/**
 * Deep equality comparison that ignores property order
 * @param {any} a First value to compare
 * @param {any} b Second value to compare
 * @returns {boolean} True if the values are deeply equal
 */
export function deepEqual(a, b) {
    // If they're strictly equal, return true
    if (a === b) return true;
    
    // Check if either is null/undefined
    if (a == null || b == null) return a === b;
    
    // Check object types
    if (typeof a !== 'object' || typeof b !== 'object') return a === b;
    
    // Handle arrays
    if (Array.isArray(a) && Array.isArray(b)) {
        if (a.length !== b.length) return false;
        for (let i = 0; i < a.length; i++) {
            if (!deepEqual(a[i], b[i])) return false;
        }
        return true;
    }
    
    // If one is array and the other isn't
    if (Array.isArray(a) !== Array.isArray(b)) return false;
    
    // Compare object properties, ignoring order
    const aKeys = Object.keys(a).sort();
    const bKeys = Object.keys(b).sort();
    
    if (aKeys.length !== bKeys.length) return false;
    
    // Check if all keys match
    for (let i = 0; i < aKeys.length; i++) {
        if (aKeys[i] !== bKeys[i]) return false;
    }
    
    // Check all property values
    for (const key of aKeys) {
        if (!deepEqual(a[key], b[key])) return false;
    }
    
    return true;
}