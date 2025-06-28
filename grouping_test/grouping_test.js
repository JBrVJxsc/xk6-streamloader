import streamloader from 'k6/x/streamloader';

/**
 * Processes order test data from a CSV file, grouping items by orderId
 * MEMORY-OPTIMIZED VERSION while maintaining array return format
 *
 * @param {Object} options - Configuration options
 * @param {string} options.csvFilePath - Path to the CSV file
 * @param {number|null} options.orderIdColumn - Column index for orderId (0-based), null if no orderId
 * @param {number} options.vendorItemIdColumn - Column index for vendorItemId (0-based)
 * @param {number|null} options.quantityColumn - Column index for quantity (0-based), null if using fixed quantity
 * @param {number|null} options.dateColumn - Column index for date (0-based), null if no date check needed
 * @param {string|null} options.quantity - Fixed quantity value to use if quantityColumn is null
 * @param {boolean} options.checkFreshness - Whether to check if data is fresh (not older than 2 days)
 * @param {number} options.maxDaysOld - Maximum days old the data can be (default: 2)
 * @returns {Array<Array<number>>} - List of flat arrays where each array contains [vendorItemId, quantity, vendorItemId, quantity, ...] for each order group
 */
export function getOrderTestData(options) {
    // Destructure options with defaults
    const {
        csvFilePath,
        orderIdColumn = 0,
        vendorItemIdColumn = 1,
        quantityColumn = 2,
        dateColumn = 3,
        quantity = null,
        checkFreshness = true,
        maxDaysOld = 2,
    } = options;

    try {
        if (!csvFilePath || typeof csvFilePath !== 'string') {
            throw new Error('csvFilePath must be provided as a string');
        }

        // First, check freshness using the last line of the file
        if (checkFreshness && dateColumn !== null) {
            const lastLine = streamloader.tail(csvFilePath, 1);

            if (!lastLine || lastLine.length === 0) {
                throw new Error('Unable to read file or file is empty');
            }

            // Parse the last line to get the date
            const lastLineData = lastLine.trim().split(',');

            if (lastLineData.length > dateColumn) {
                const orderDate = lastLineData[dateColumn];
                // Extract date portion (first 8 chars) from datetime string (YYYYMMDDHHMMSS)
                const orderDateStr = orderDate.substring(0, 8);

                // Calculate minimum allowed date string (YYYYMMDD format)
                const minAllowedDate = getDateNDaysAgo(maxDaysOld);

                // Simple string comparison (since YYYYMMDD format allows lexicographic comparison)
                if (orderDateStr < minAllowedDate) {
                    console.error(`Order data is stale, more than ${maxDaysOld} days old, order date: ${orderDateStr}`);
                    throw new Error('Stale order');
                }
            }
        }

        // Configure streamloader options for CSV processing
        const streamloaderOptions = {
            skipHeader: true,
            filters: [
                // Filter out empty rows by checking if vendorItemId column is not empty
                { type: 'emptyString', column: vendorItemIdColumn },
            ],
            fields: [
                { type: 'column', column: vendorItemIdColumn },
            ],
        };

        // Add quantity field
        if (quantityColumn !== null) {
            streamloaderOptions.fields.push({ type: 'column', column: quantityColumn });
        } else if (quantity !== null) {
            streamloaderOptions.fields.push({ type: 'fixed', value: quantity });
        } else {
            streamloaderOptions.fields.push({ type: 'fixed', value: 1 });
        }

        // Configure groupBy only if orderIdColumn is provided
        if (orderIdColumn !== null) {
            streamloaderOptions.groupBy = { column: orderIdColumn };
        }

        // Process the CSV file using streamloader
        const csvData = streamloader.processCsvFile(csvFilePath, streamloaderOptions);

        if (!csvData || csvData.length === 0) {
            throw new Error('No data found in CSV file or all rows filtered out');
        }

        return csvData;

    } catch (e) {
        console.error(`Failed to get test order data, ${e.message}`);
        throw e; // Re-throw for unit tests to catch
    }
}

/**
 * Helper function to get date string N days ago in YYYYMMDD format
 * @param {number} daysAgo - Number of days ago
 * @returns {string} Date string in YYYYMMDD format
 */
export function getDateNDaysAgo(daysAgo) {
    const date = new Date();
    date.setDate(date.getDate() - daysAgo);
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}${month}${day}`;
}
