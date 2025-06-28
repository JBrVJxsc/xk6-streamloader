import { check, group } from 'k6';
import { processCsvFile } from 'k6/x/streamloader';

export default function () {
    const ordersFilePath = 'orders.csv';
    const productsFilePath = 'products.csv';

    group('CSV Transform Tests', function() {
        // Test 1: Convert quantity to integers
        const intTransform = processCsvFile(ordersFilePath, {
            skipHeader: true,
            filters: [],
            transforms: [
                {
                    type: "parseInt",
                    column: 2  // qty column
                }
            ],
            fields: []
        });
        
        // Debug the transform result
        console.log("Int transform result:", JSON.stringify(intTransform));
        
        check(intTransform, {
            'Quantity transformed to integers': (data) => 
                Array.isArray(data) && 
                data.length === 8 &&  // Updated to match 8 rows in orders.csv
                data.every(row => !isNaN(parseInt(row[2]))),
            'Values correctly transformed': (data) => {
                const qtySum = data.reduce((sum, row) => sum + parseInt(row[2]), 0);
                return qtySum === 18;  // Total of all quantities (1+1+3+4+5+1+1+2)
            }
        });
        
        // Test 2: Format ordered date (substring transform)
        const dateTransform = processCsvFile(ordersFilePath, {
            skipHeader: true,
            filters: [],
            transforms: [
                {
                    type: "substring",
                    column: 3,  // orderedAt column
                    start: 0,
                    length: 8  // Just get YYYYMMDD part
                }
            ],
            fields: []
        });
        
        // Debug the date transform result
        console.log("Date transform result:", JSON.stringify(dateTransform));
        
        check(dateTransform, {
            'Date format transformed correctly': (data) => 
                Array.isArray(data) && 
                data.length === 8 &&
                data.every(row => typeof row[3] === 'string' && row[3].startsWith("20250626")),
        });
        
        // Test 3: Multiple transforms - convert price to int and standardize product IDs
        const multipleTransforms = processCsvFile(productsFilePath, {
            skipHeader: true,
            filters: [],
            transforms: [
                {
                    type: "parseInt",
                    column: 3  // price column
                },
                {
                    type: "substring",
                    column: 0,  // productId column
                    start: 8,  // Start after "product_"
                    length: 1
                },
                {
                    type: "fixedValue",
                    column: 1,  // itemId column
                    value: "ITEM-1"
                }
            ],
            fields: []
        });
        
        check(multipleTransforms, {
            'Prices converted to integers': (data) => 
                Array.isArray(data) && 
                data.length === 5 &&
                data.every(row => !isNaN(parseInt(row[3]))),
            'Product IDs transformed to just the number': (data) =>
                data.every(row => row[0] === "1"),
            'Item IDs standardized': (data) =>
                data.every(row => row[1] === "ITEM-1"),
        });
    });

    group('CSV Fields and Projection Tests', function() {
        // Test 4: Basic field selection
        const basicProjection = processCsvFile(ordersFilePath, {
            skipHeader: true,
            filters: [],
            transforms: [],
            fields: [
                {
                    type: "column",
                    column: 0  // orderId
                },
                {
                    type: "column",
                    column: 2  // qty
                }
            ]
        });
        
        // Debug the projection result
        console.log("Basic projection result:", JSON.stringify(basicProjection));
        
        check(basicProjection, {
            'Basic projection includes correct columns': (data) => 
                Array.isArray(data) && 
                data.length === 8 &&
                data.every(row => row.length === 2),
            'First column is orderId': (data) =>
                data[0][0] === "order_0" && data[6][0] === "order_4",
            'Second column is quantity': (data) =>
                data[0][1] === "1" && data[4][1] === "5",
        });
        
        // Test 5: Projection with transform
        const transformedProjection = processCsvFile(ordersFilePath, {
            skipHeader: true,
            filters: [],
            transforms: [
                {
                    type: "parseInt",
                    column: 2  // qty column
                }
            ],
            fields: [
                {
                    type: "column",
                    column: 0  // orderId
                },
                {
                    type: "column",
                    column: 2  // qty (now as integer)
                }
            ]
        });
        
        // Debug transformed projection result
        console.log("Transformed projection result:", JSON.stringify(transformedProjection));
        
        check(transformedProjection, {
            'Projection includes transformed quantities': (data) => 
                Array.isArray(data) && 
                data.length === 8 &&
                // Check if there's a row with qty=3 (order_1,vi_2,3)
                data.some(row => row.length === 2 && parseInt(row[1]) === 3),
        });
        
        // Test 6: Fixed value fields with column fields
        const mixedFieldTypes = processCsvFile(productsFilePath, {
            skipHeader: true,
            filters: [],
            transforms: [],
            fields: [
                {
                    type: "column",
                    column: 2  // vendorItemId
                },
                {
                    type: "column",
                    column: 3  // price
                },
                {
                    type: "fixed",
                    value: "USD"  // Adding currency
                }
            ]
        });
        
        check(mixedFieldTypes, {
            'Mixed field types projection works': (data) => 
                Array.isArray(data) && 
                data.length === 5 &&
                data.every(row => row.length === 3),
            'First column is vendorItemId': (data) =>
                data[0][0] === "vi_1" && data[4][0] === "vi_5",
            'Second column is price': (data) =>
                data[0][1] === "1" && data[4][1] === "5",
            'Third column is fixed value': (data) =>
                data.every(row => row[2] === "USD"),
        });
        
        // Test 7: Combining filters, transforms and projections
        const combinedOperations = processCsvFile(productsFilePath, {
            skipHeader: true,
            filters: [
                {
                    type: "valueRange",
                    column: 3,  // price
                    min: 3,
                    max: 999
                }
            ],
            transforms: [
                {
                    type: "parseInt",
                    column: 3  // price to int
                }
            ],
            fields: [
                {
                    type: "column",
                    column: 2  // vendorItemId
                },
                {
                    type: "column",
                    column: 3  // price (as int)
                },
                {
                    type: "fixed",
                    value: "High-Price"  // Category
                }
            ]
        });
        
        check(combinedOperations, {
            'Combined operations work correctly': (data) => 
                Array.isArray(data) && 
                data.length === 3 && // Only items with price >= 3
                data.every(row => row.length === 3),
            'Contains correct vendorItemIds': (data) =>
                data.some(row => row[0] === "vi_3") &&
                data.some(row => row[0] === "vi_4") &&
                data.some(row => row[0] === "vi_5"),
            'Prices correctly transformed': (data) =>
                data.every(row => !isNaN(parseInt(row[1])) && parseInt(row[1]) >= 3),
            'Fixed value added': (data) =>
                data.every(row => row[2] === "High-Price"),
        });
    });

    group('CSV GroupBy Tests', function() {
        // Test 8: Group orders by orderId
        const orderGrouping = processCsvFile(ordersFilePath, {
            skipHeader: true,
            filters: [],
            transforms: [],
            groupBy: {
                column: 0  // orderId
            },
            fields: [
                {
                    type: "column",
                    column: 1  // vendorItemId
                },
                {
                    type: "column",
                    column: 2  // qty
                }
            ]
        });
        
        // Debug the order grouping result
        console.log("Order grouping result:", JSON.stringify(orderGrouping));
        
        check(orderGrouping, {
            'Orders grouped correctly': (data) => 
                Array.isArray(data) && 
                data.length === 6, // 6 unique order IDs now with order_5
            'Multiple items for order_1 grouped': (data) => {
                // Find the row with both vi_1 and vi_2
                const order1Row = data.find(row => 
                    row.includes("vi_1") && row.includes("vi_2"));
                return order1Row !== undefined;
            },
            'Multiple items for order_2 grouped': (data) => {
                // Find the row with both instances of vi_3
                const order2Row = data.find(row => 
                    row.filter(cell => cell === "vi_3").length === 2);
                return order2Row !== undefined;
            },
        });
        
        // Test 9: Group products by vendorItemId with price transform
        const productGrouping = processCsvFile(productsFilePath, {
            skipHeader: true,
            filters: [],
            transforms: [
                {
                    type: "parseInt",
                    column: 3  // price
                }
            ],
            groupBy: {
                column: 2  // vendorItemId
            },
            fields: [
                {
                    type: "column",
                    column: 3  // price (as int)
                }
            ]
        });
        
        check(productGrouping, {
            'Products grouped correctly': (data) => 
                Array.isArray(data) && 
                data.length === 5, // 5 unique vendorItemIds
            'Each group has one row per vendorItemId': (data) =>
                data.every(row => row.length === 1)
        });
        
        // Test 10: Complex grouping with filters and transforms
        const complexGrouping = processCsvFile(ordersFilePath, {
            skipHeader: true,
            filters: [
                {
                    type: "valueRange",
                    column: 2,  // qty
                    min: 2,
                    max: 999
                }
            ],
            transforms: [
                {
                    type: "parseInt",
                    column: 2  // qty to int
                },
                {
                    type: "substring",
                    column: 3,  // orderedAt
                    start: 8,
                    length: 6  // Get HHMMSS part
                }
            ],
            groupBy: {
                column: 0  // orderId
            },
            fields: [
                {
                    type: "column",
                    column: 1  // vendorItemId
                },
                {
                    type: "column", 
                    column: 2  // qty (as int)
                },
                {
                    type: "column",
                    column: 3  // time portion of orderedAt
                }
            ]
        });
        
        // Debug the complex grouping result
        console.log("Complex grouping result:", JSON.stringify(complexGrouping));
        
        check(complexGrouping, {
            'Complex grouping works correctly': (data) => 
                Array.isArray(data) && 
                data.length === 3, // order_1, order_2, and order_5 have qty >= 2
            'Order_1 group contains items with qty >= 2': (data) => {
                const order1Row = data.find(row => row.includes("vi_2"));
                return order1Row !== undefined && order1Row.some(cell => !isNaN(parseInt(cell)) && parseInt(cell) >= 2);
            },
            'Order_2 group contains items with qty >= 2': (data) => {
                const order2Row = data.find(row => row.includes("vi_3"));
                return order2Row !== undefined && order2Row.some(cell => !isNaN(parseInt(cell)) && parseInt(cell) >= 2);
            },
        });
    });
}