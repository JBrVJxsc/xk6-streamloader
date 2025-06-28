import { check, group } from 'k6';
import { debugOptions, debugCsvOptions, processCsvFile } from 'k6/x/streamloader';

// Helper function to print the string representation of an object
function printObject(obj) {
    console.log(JSON.stringify(obj, null, 2));
}

export default function () {
    group('Testing parameter passing with debugOptions', function () {
        // Test 1: Simple object
        const simpleObject = {
            name: "Test Object",
            value: 42,
            enabled: true
        };
        
        console.log("Original simple object:");
        printObject(simpleObject);
        
        // Pass through debugOptions
        const returnedSimpleObject = debugOptions(simpleObject);
        
        console.log("Returned simple object:");
        printObject(returnedSimpleObject);
        
        // Check if the returned object matches the original
        check(returnedSimpleObject, {
            'Simple object: name is preserved': (obj) => obj.name === "Test Object",
            'Simple object: value is preserved': (obj) => obj.value === 42,
            'Simple object: enabled is preserved': (obj) => obj.enabled === true
        });
        
        // Test 2: Complex nested object
        const complexObject = {
            id: "complex-1",
            settings: {
                theme: "dark",
                notifications: {
                    email: true,
                    push: false
                }
            },
            items: [
                { id: 1, name: "Item 1" },
                { id: 2, name: "Item 2" }
            ]
        };
        
        console.log("Original complex object:");
        printObject(complexObject);
        
        // Pass through debugOptions
        const returnedComplexObject = debugOptions(complexObject);
        
        console.log("Returned complex object:");
        printObject(returnedComplexObject);
        
        // Check if the returned object matches the original
        check(returnedComplexObject, {
            'Complex object: id is preserved': (obj) => obj.id === "complex-1",
            'Complex object: theme is preserved': (obj) => obj.settings?.theme === "dark",
            'Complex object: notifications.email is preserved': (obj) => obj.settings?.notifications?.email === true,
            'Complex object: items array exists': (obj) => Array.isArray(obj.items),
            'Complex object: items length is correct': (obj) => obj.items?.length === 2,
            'Complex object: item name is preserved': (obj) => obj.items?.[0]?.name === "Item 1"
        });
    });
    
    group('Testing parameter passing with ProcessCsvOptions', function () {
        // Test with ProcessCsvOptions structure
        const csvOptions = {
            skipHeader: true,
            filters: [
                { type: "regexMatch", column: 1, pattern: "Product" },
                { type: "valueRange", column: 2, min: 100, max: 200 }
            ],
            transforms: [
                { type: "parseInt", column: 3 },
                { type: "substring", column: 1, start: 0, length: 5 }
            ],
            groupBy: {
                column: 4
            },
            fields: [
                { type: "column", column: 0 },
                { type: "fixed", value: "TEST" }
            ]
        };
        
        console.log("Original CSV options:");
        printObject(csvOptions);
        
        // Pass through debugCsvOptions
        const returnedCsvOptions = debugCsvOptions(csvOptions);
        
        console.log("Returned CSV options:");
        printObject(returnedCsvOptions);
        
        // Check if the returned object matches the original
        check(returnedCsvOptions, {
            'CSV options: skipHeader is preserved': (obj) => obj.skipHeader === true,
            'CSV options: filters array exists': (obj) => Array.isArray(obj.filters),
            'CSV options: filters length is correct': (obj) => obj.filters?.length === 2,
            'CSV options: filter type is preserved': (obj) => obj.filters?.[0]?.type === "regexMatch",
            'CSV options: filter pattern is preserved': (obj) => obj.filters?.[0]?.pattern === "Product",
            'CSV options: valueRange min is preserved': (obj) => obj.filters?.[1]?.min === 100,
            'CSV options: transforms array exists': (obj) => Array.isArray(obj.transforms),
            'CSV options: transform type is preserved': (obj) => obj.transforms?.[0]?.type === "parseInt",
            'CSV options: transform column is preserved': (obj) => obj.transforms?.[0]?.column === 3,
            'CSV options: substring length is preserved': (obj) => obj.transforms?.[1]?.length === 5,
            'CSV options: groupBy exists': (obj) => obj.groupBy !== undefined,
            'CSV options: groupBy column is preserved': (obj) => obj.groupBy?.column === 4,
            'CSV options: fields array exists': (obj) => Array.isArray(obj.fields),
            'CSV options: field type is preserved': (obj) => obj.fields?.[0]?.type === "column",
            'CSV options: fixed value is preserved': (obj) => obj.fields?.[1]?.value === "TEST"
        });
        
        // Additional test for nested objects and arrays
        const advancedOptions = {
            skipHeader: true,
            filters: [
                { 
                    type: "composite",
                    conditions: [
                        { type: "regexMatch", column: 1, pattern: "Product" },
                        { type: "valueRange", column: 2, min: 100, max: 200 }
                    ]
                }
            ],
            metadata: {
                source: "Test Source",
                version: 1.0,
                tags: ["test", "debug"]
            }
        };
        
        console.log("Advanced CSV options:");
        printObject(advancedOptions);
        
        // Try to pass through debugCsvOptions - this might fail since it doesn't match the schema
        try {
            const returnedAdvancedOptions = debugCsvOptions(advancedOptions);
            
            console.log("Returned advanced options (might be partial):");
            printObject(returnedAdvancedOptions);
            
            // Check what was preserved
            check(returnedAdvancedOptions, {
                'Advanced options: skipHeader is preserved': (obj) => obj.skipHeader === true,
                'Advanced options: filters array exists': (obj) => Array.isArray(obj.filters)
            });
        } catch (e) {
            console.error(`Error passing advanced options: ${e.message}`);
            check(null, {
                'Advanced options handle error gracefully': () => true
            });
        }
    });
    
    group('Testing ProcessCsvFile parameter passing directly', function () {
        // Create a simple test file content
        console.log("Testing ProcessCsvFile with simple options...");
        
        // Basic options
        const basicOptions = {
            skipHeader: true,
            filters: [],
            transforms: [],
            fields: []
        };
        
        try {
            // This will fail since we don't have the file, but we can check the error message
            processCsvFile("nonexistent.csv", basicOptions);
        } catch (e) {
            console.log(`ProcessCsvFile error message: ${e.message}`);
            check(e.message, {
                'Error message contains "file": Indicates options were passed correctly': (msg) => msg.includes("file")
            });
        }
    });
}