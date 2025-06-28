import { check } from 'k6';
import { processCsvFile } from 'k6/x/streamloader';

export default function () {
    // Create a test file path
    const testFilePath = 'advanced_process.csv';
    
    // Define minimal options to just read the CSV
    const options = {
        skipHeader: true,
        fields: []  // no field projection, get raw rows
    };
    
    try {
        // Process CSV file with minimal options
        const result = processCsvFile(testFilePath, options);
        
        // Log the result structure
        console.log(`CSV file has ${result.length} rows`);
        if (result.length > 0) {
            console.log("First row content:", JSON.stringify(result[0]));
            console.log("First row length:", result[0].length);
            
            // Output each field with its index
            for (let i = 0; i < result[0].length; i++) {
                console.log(`Field ${i}: ${result[0][i]}`);
            }
        }
        
        // Check if we're getting the expected data
        if (result.length > 0) {
            const row = result[0];
            const id = row[0];        // Should be "1001"
            const name = row[1];      // Should be "Desktop Computer"
            const price = row[2];     // Should be "999.99"
            const quantity = row[3];  // Should be "5"
            const category = row[4];  // Should be "Electronics"
            
            console.log(`\nChecking first row values:`);
            console.log(`ID: ${id}`);
            console.log(`Name: ${name}`);
            console.log(`Price: ${price}`);
            console.log(`Quantity: ${quantity}`);
            console.log(`Category: ${category}`);
        }
        
    } catch (e) {
        console.error(`Error processing CSV: ${e.message}`);
    }
    
    // Try the regex filtering again with explicit columns
    const regexOptions = {
        skipHeader: true,
        filters: [
            {
                type: "regexMatch",
                column: 4,  // category column
                pattern: "^(Audio|Mobile)$"
            }
        ],
        fields: [
            { type: "column", column: 0 },  // id
            { type: "column", column: 1 },  // name
            { type: "column", column: 4 }   // category
        ]
    };
    
    console.log("\nTrying regex filter again with explicit column references:");
    
    try {
        const result = processCsvFile(testFilePath, regexOptions);
        console.log(`Filtered result has ${result.length} rows`);
        
        if (result.length > 0) {
            console.log("First row after filtering:", JSON.stringify(result[0]));
            
            // Check categories directly
            const categories = result.map(row => row[2]);
            console.log("All categories in result:", categories);
            
            const uniqueCategories = [...new Set(categories)];
            console.log("Unique categories:", uniqueCategories);
            
            // Check if we have Audio or Mobile
            const hasAudio = categories.includes("Audio");
            const hasMobile = categories.includes("Mobile");
            
            console.log(`Contains Audio category: ${hasAudio}`);
            console.log(`Contains Mobile category: ${hasMobile}`);
        }
    } catch (e) {
        console.error(`Error with regex filtering: ${e.message}`);
    }
}