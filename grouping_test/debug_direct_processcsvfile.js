import { check } from 'k6';
import { processCsvFile } from 'k6/x/streamloader';

export default function () {
    // Define the path to our test CSV
    const filePath = '../advanced_process.csv';
    
    console.log(`Processing file: ${filePath}`);
    
    // First test: just load the raw data
    const basicOptions = {
        skipHeader: true
    };
    
    try {
        const result = processCsvFile(filePath, basicOptions);
        console.log(`Successfully loaded CSV with ${result.length} rows`);
        
        if (result.length > 0) {
            console.log("First row:", JSON.stringify(result[0]));
        }
    } catch (e) {
        console.error(`Error loading CSV: ${e.message}`);
    }
    
    // Test with explicit struct fields
    console.log("\nTesting with explicit fields in options:");
    
    const explicitOptions = {
        skipHeader: true,
        filters: [],
        transforms: [],
        fields: [
            { type: "column", column: 0 },  // id
            { type: "column", column: 1 },  // name
            { type: "column", column: 4 }   // category
        ]
    };
    
    try {
        const result = processCsvFile(filePath, explicitOptions);
        console.log(`Result with explicit fields has ${result.length} rows`);
        
        if (result.length > 0) {
            console.log("First row:", JSON.stringify(result[0]));
            console.log(`ID: ${result[0][0]}`);
            console.log(`Name: ${result[0][1]}`);
            console.log(`Category: ${result[0][2]}`);
        }
        
        // Try to check each category
        console.log("\nCategories in result:");
        const categories = result.map(row => row[2]);
        const uniqueCategories = [...new Set(categories)];
        console.log("Unique categories:", uniqueCategories.join(", "));
    } catch (e) {
        console.error(`Error with explicit fields: ${e.message}`);
    }
}