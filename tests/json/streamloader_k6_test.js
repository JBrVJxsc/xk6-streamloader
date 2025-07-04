import streamloader from 'k6/x/streamloader';
import { check, fail } from 'k6';

export const options = {
    thresholds: {
        // Require 100% of checks to pass
        'checks': ['rate==1.0'],
    },
};

export default function () {
    // JSON Tests (existing)
    
    // 1. Normal case: valid file
    const objects = streamloader.loadJSON('samples.json');
    console.log('objects:', JSON.stringify(objects));
    check(objects, {
        'loaded 2 objects': (s) => s.length === 2,
        'first object is GET': (s) => s[0].method === 'GET',
        'first object URI': (s) => s[0].requestURI === '/foo',
        'first object header': (s) => s[0].headers['A'] === 'B',
        'first object content': (s) => s[0].content === 'abc',
        'second object is POST': (s) => s[1].method === 'POST',
        'second object URI': (s) => s[1].requestURI === '/bar',
        'second object header': (s) => s[1].headers['C'] === 'D',
        'second object content': (s) => s[1].content === 'def',
    });

    // 2. Error case: missing file
    try {
        streamloader.loadJSON('no_such_file.json');
        fail('Expected error for missing file');
    } catch (e) {
        check(e, { 'error for missing file': (err) => String(err).includes('no_such_file') || String(err).includes('no such file') });
    }

    // 3. Error case: invalid JSON
    try {
        streamloader.loadJSON('bad.json');
        fail('Expected error for invalid JSON');
    } catch (e) {
        console.log('Invalid JSON error:', String(e));
        check(e, { 'error for invalid JSON': (err) => String(err).toLowerCase().includes('invalid') });
    }

    // 4. Edge case: empty array
    const emptyObjects = streamloader.loadJSON('empty.json');
    console.log('emptyObjects:', JSON.stringify(emptyObjects));
    check(emptyObjects, { 'empty array returns empty': (s) => Array.isArray(s) && s.length === 0 });

    // 5. Large array
    const largeObjects = streamloader.loadJSON('large.json');
    console.log('largeObjects length:', largeObjects.length);
    console.log('largeObjects[0]:', JSON.stringify(largeObjects[0]));
    console.log('largeObjects[999]:', JSON.stringify(largeObjects[999]));
    check(largeObjects, {
        'large array length': (s) => s.length === 1000,
        'large array first': (s) => s[0].requestURI === '/bulk/0',
        'large array last': (s) => s[999].requestURI === '/bulk/999',
    });

    // 6. Complex JSON structure test
    const complexObjects = streamloader.loadJSON('complex.json');
    console.log('complexObjects length:', complexObjects.length);
    console.log('complexObjects[0] user name:', complexObjects[0].user.name);
    console.log('complexObjects[0] user profile:', JSON.stringify(complexObjects[0].user.profile));
    
    check(complexObjects, {
        'complex array length': (s) => s.length === 3,
        
        // Test first object structure
        'first object has id': (s) => s[0].id === 1,
        'first object user name': (s) => s[0].user.name === 'Alice Johnson',
        'first object user email': (s) => s[0].user.email === 'alice@example.com',
        'first object user age': (s) => s[0].user.profile.age === 30,
        'first object user city': (s) => s[0].user.profile.location.city === 'New York',
        'first object user country': (s) => s[0].user.profile.location.country === 'USA',
        'first object user coordinates lat': (s) => s[0].user.profile.location.coordinates.lat === 40.7128,
        'first object user coordinates lng': (s) => s[0].user.profile.location.coordinates.lng === -74.0060,
        'first object user theme': (s) => s[0].user.profile.preferences.theme === 'dark',
        'first object user email notification': (s) => s[0].user.profile.preferences.notifications.email === true,
        'first object user push notification': (s) => s[0].user.profile.preferences.notifications.push === false,
        'first object user sms notification': (s) => s[0].user.profile.preferences.notifications.sms === true,
        'first object user languages array length': (s) => s[0].user.profile.preferences.languages.length === 3,
        'first object user languages first': (s) => s[0].user.profile.preferences.languages[0] === 'en',
        'first object user languages second': (s) => s[0].user.profile.preferences.languages[1] === 'es',
        'first object user languages third': (s) => s[0].user.profile.preferences.languages[2] === 'fr',
        'first object user roles array length': (s) => s[0].user.roles.length === 2,
        'first object user roles first': (s) => s[0].user.roles[0] === 'admin',
        'first object user roles second': (s) => s[0].user.roles[1] === 'user',
        'first object user created at': (s) => s[0].user.metadata.created_at === '2023-01-15T10:30:00Z',
        'first object user last login': (s) => s[0].user.metadata.last_login === '2024-01-20T14:45:00Z',
        'first object user login count': (s) => s[0].user.metadata.login_count === 156,
        
        // Test orders array
        'first object orders array length': (s) => s[0].orders.length === 1,
        'first object order id': (s) => s[0].orders[0].order_id === 'ORD-001',
        'first object order items length': (s) => s[0].orders[0].items.length === 2,
        'first object first item product id': (s) => s[0].orders[0].items[0].product_id === 'PROD-123',
        'first object first item name': (s) => s[0].orders[0].items[0].name === 'Laptop',
        'first object first item price': (s) => s[0].orders[0].items[0].price === 1299.99,
        'first object first item quantity': (s) => s[0].orders[0].items[0].quantity === 1,
        'first object first item cpu': (s) => s[0].orders[0].items[0].specifications.cpu === 'Intel i7',
        'first object first item ram': (s) => s[0].orders[0].items[0].specifications.ram === '16GB',
        'first object first item storage': (s) => s[0].orders[0].items[0].specifications.storage === '512GB SSD',
        'first object second item product id': (s) => s[0].orders[0].items[1].product_id === 'PROD-456',
        'first object second item name': (s) => s[0].orders[0].items[1].name === 'Mouse',
        'first object second item price': (s) => s[0].orders[0].items[1].price === 29.99,
        'first object second item quantity': (s) => s[0].orders[0].items[1].quantity === 2,
        'first object second item type': (s) => s[0].orders[0].items[1].specifications.type === 'Wireless',
        'first object second item dpi': (s) => s[0].orders[0].items[1].specifications.dpi === 1600,
        
        // Test shipping information
        'first object shipping street': (s) => s[0].orders[0].shipping.address.street === '123 Main St',
        'first object shipping city': (s) => s[0].orders[0].shipping.address.city === 'New York',
        'first object shipping state': (s) => s[0].orders[0].shipping.address.state === 'NY',
        'first object shipping zip': (s) => s[0].orders[0].shipping.address.zip === '10001',
        'first object shipping method': (s) => s[0].orders[0].shipping.method === 'express',
        'first object tracking number': (s) => s[0].orders[0].shipping.tracking.number === 'TRK-789456',
        'first object tracking status': (s) => s[0].orders[0].shipping.tracking.status === 'delivered',
        'first object tracking events length': (s) => s[0].orders[0].shipping.tracking.events.length === 3,
        'first object first tracking event timestamp': (s) => s[0].orders[0].shipping.tracking.events[0].timestamp === '2024-01-18T09:00:00Z',
        'first object first tracking event status': (s) => s[0].orders[0].shipping.tracking.events[0].status === 'shipped',
        'first object first tracking event location': (s) => s[0].orders[0].shipping.tracking.events[0].location === 'Warehouse',
        
        // Test settings
        'first object profile visibility': (s) => s[0].settings.privacy.profile_visibility === 'public',
        'first object analytics sharing': (s) => s[0].settings.privacy.data_sharing.analytics === true,
        'first object marketing sharing': (s) => s[0].settings.privacy.data_sharing.marketing === false,
        'first object third party sharing': (s) => s[0].settings.privacy.data_sharing.third_party === false,
        'first object two factor': (s) => s[0].settings.security.two_factor === true,
        'first object session timeout': (s) => s[0].settings.security.session_timeout === 3600,
        'first object allowed ips length': (s) => s[0].settings.security.allowed_ips.length === 2,
        'first object first allowed ip': (s) => s[0].settings.security.allowed_ips[0] === '192.168.1.1',
        'first object second allowed ip': (s) => s[0].settings.security.allowed_ips[1] === '10.0.0.1',
        
        // Test second object structure
        'second object has id': (s) => s[1].id === 2,
        'second object user name': (s) => s[1].user.name === 'Bob Smith',
        'second object user email': (s) => s[1].user.email === 'bob@example.com',
        'second object user age': (s) => s[1].user.profile.age === 25,
        'second object user city': (s) => s[1].user.profile.location.city === 'Los Angeles',
        'second object user coordinates lat': (s) => s[1].user.profile.location.coordinates.lat === 34.0522,
        'second object user coordinates lng': (s) => s[1].user.profile.location.coordinates.lng === -118.2437,
        'second object user theme': (s) => s[1].user.profile.preferences.theme === 'light',
        'second object user email notification': (s) => s[1].user.profile.preferences.notifications.email === false,
        'second object user push notification': (s) => s[1].user.profile.preferences.notifications.push === true,
        'second object user sms notification': (s) => s[1].user.profile.preferences.notifications.sms === false,
        'second object user languages array length': (s) => s[1].user.profile.preferences.languages.length === 1,
        'second object user languages first': (s) => s[1].user.profile.preferences.languages[0] === 'en',
        'second object user roles array length': (s) => s[1].user.roles.length === 1,
        'second object user roles first': (s) => s[1].user.roles[0] === 'user',
        'second object user login count': (s) => s[1].user.metadata.login_count === 89,
        
        // Test empty orders array
        'second object orders array length': (s) => s[1].orders.length === 0,
        
        // Test second object settings
        'second object profile visibility': (s) => s[1].settings.privacy.profile_visibility === 'private',
        'second object analytics sharing': (s) => s[1].settings.privacy.data_sharing.analytics === false,
        'second object marketing sharing': (s) => s[1].settings.privacy.data_sharing.marketing === false,
        'second object third party sharing': (s) => s[1].settings.privacy.data_sharing.third_party === false,
        'second object two factor': (s) => s[1].settings.security.two_factor === false,
        'second object session timeout': (s) => s[1].settings.security.session_timeout === 1800,
        'second object allowed ips length': (s) => s[1].settings.security.allowed_ips.length === 0,
        
        // Test third object structure
        'third object has id': (s) => s[2].id === 3,
        'third object user name': (s) => s[2].user.name === 'Carol Davis',
        'third object user email': (s) => s[2].user.email === 'carol@example.com',
        'third object user age': (s) => s[2].user.profile.age === 35,
        'third object user city': (s) => s[2].user.profile.location.city === 'Chicago',
        'third object user coordinates lat': (s) => s[2].user.profile.location.coordinates.lat === 41.8781,
        'third object user coordinates lng': (s) => s[2].user.profile.location.coordinates.lng === -87.6298,
        'third object user theme': (s) => s[2].user.profile.preferences.theme === 'auto',
        'third object user email notification': (s) => s[2].user.profile.preferences.notifications.email === true,
        'third object user push notification': (s) => s[2].user.profile.preferences.notifications.push === true,
        'third object user sms notification': (s) => s[2].user.profile.preferences.notifications.sms === true,
        'third object user languages array length': (s) => s[2].user.profile.preferences.languages.length === 2,
        'third object user languages first': (s) => s[2].user.profile.preferences.languages[0] === 'en',
        'third object user languages second': (s) => s[2].user.profile.preferences.languages[1] === 'de',
        'third object user roles array length': (s) => s[2].user.roles.length === 2,
        'third object user roles first': (s) => s[2].user.roles[0] === 'user',
        'third object user roles second': (s) => s[2].user.roles[1] === 'moderator',
        'third object user login count': (s) => s[2].user.metadata.login_count === 234,
        
        // Test third object orders
        'third object orders array length': (s) => s[2].orders.length === 1,
        'third object order id': (s) => s[2].orders[0].order_id === 'ORD-002',
        'third object order items length': (s) => s[2].orders[0].items.length === 1,
        'third object item product id': (s) => s[2].orders[0].items[0].product_id === 'PROD-789',
        'third object item name': (s) => s[2].orders[0].items[0].name === 'Headphones',
        'third object item price': (s) => s[2].orders[0].items[0].price === 199.99,
        'third object item quantity': (s) => s[2].orders[0].items[0].quantity === 1,
        'third object item type': (s) => s[2].orders[0].items[0].specifications.type === 'Bluetooth',
        'third object item battery life': (s) => s[2].orders[0].items[0].specifications.battery_life === '20 hours',
        'third object item noise cancellation': (s) => s[2].orders[0].items[0].specifications.noise_cancellation === true,
        
        // Test third object shipping
        'third object shipping street': (s) => s[2].orders[0].shipping.address.street === '456 Oak Ave',
        'third object shipping city': (s) => s[2].orders[0].shipping.address.city === 'Chicago',
        'third object shipping state': (s) => s[2].orders[0].shipping.address.state === 'IL',
        'third object shipping zip': (s) => s[2].orders[0].shipping.address.zip === '60601',
        'third object shipping method': (s) => s[2].orders[0].shipping.method === 'standard',
        'third object tracking number': (s) => s[2].orders[0].shipping.tracking.number === 'TRK-123789',
        'third object tracking status': (s) => s[2].orders[0].shipping.tracking.status === 'in_transit',
        'third object tracking events length': (s) => s[2].orders[0].shipping.tracking.events.length === 1,
        'third object first tracking event timestamp': (s) => s[2].orders[0].shipping.tracking.events[0].timestamp === '2024-01-20T10:00:00Z',
        'third object first tracking event status': (s) => s[2].orders[0].shipping.tracking.events[0].status === 'shipped',
        'third object first tracking event location': (s) => s[2].orders[0].shipping.tracking.events[0].location === 'Warehouse',
        
        // Test third object settings
        'third object profile visibility': (s) => s[2].settings.privacy.profile_visibility === 'friends',
        'third object analytics sharing': (s) => s[2].settings.privacy.data_sharing.analytics === true,
        'third object marketing sharing': (s) => s[2].settings.privacy.data_sharing.marketing === true,
        'third object third party sharing': (s) => s[2].settings.privacy.data_sharing.third_party === false,
        'third object two factor': (s) => s[2].settings.security.two_factor === true,
        'third object session timeout': (s) => s[2].settings.security.session_timeout === 7200,
        'third object allowed ips length': (s) => s[2].settings.security.allowed_ips.length === 3,
        'third object first allowed ip': (s) => s[2].settings.security.allowed_ips[0] === '192.168.1.100',
        'third object second allowed ip': (s) => s[2].settings.security.allowed_ips[1] === '10.0.0.50',
        'third object third allowed ip': (s) => s[2].settings.security.allowed_ips[2] === '172.16.0.1',
    });

    // 7. Object JSON format test
    const objectObjects = streamloader.loadJSON('object.json');
    console.log('objectObjects keys:', Object.keys(objectObjects));
    console.log('objectObjects.user1:', JSON.stringify(objectObjects.user1));
    check(objectObjects, {
        'object JSON has 3 keys': (s) => Object.keys(s).length === 3,
        'user1 object has correct data': (s) => s.user1 && s.user1.method === 'GET' && s.user1.requestURI === '/user1' && s.user1.content === 'user1_data',
        'user2 object has correct data': (s) => s.user2 && s.user2.method === 'POST' && s.user2.requestURI === '/user2' && s.user2.content === 'user2_data',
        'user3 object has correct data': (s) => s.user3 && s.user3.method === 'PUT' && s.user3.requestURI === '/user3' && s.user3.content === 'user3_data',
        'user1 headers': (s) => s.user1 && s.user1.headers && s.user1.headers.A === 'B',
        'user2 headers': (s) => s.user2 && s.user2.headers && s.user2.headers.C === 'D',
        'user3 headers': (s) => s.user3 && s.user3.headers && s.user3.headers.E === 'F',
    });

    // 8. Recording stats object test
    const recordingStatsObjects = streamloader.loadJSON('recordingstats.json');
    console.log('recordingStatsObjects keys:', Object.keys(recordingStatsObjects));
    check(recordingStatsObjects, {
        'recordingStats has 13 keys': (s) => Object.keys(s).length === 13,
        'recordingId correct': (s) => s.recordingId === '18aebc27-ef24-42a0-aee4-5e01f8ac6049',
        'domain correct': (s) => s.domain === 'EATS_CUSTOMER',
        'ttl correct': (s) => s.ttl === 604800,
        'matchedPercentage correct': (s) => s.matchedPercentage === 86.39,
        'filterStats is array': (s) => Array.isArray(s.filterStats),
        'filterStats length 1': (s) => s.filterStats.length === 1,
        'filterStats[0] domain': (s) => s.filterStats[0].domain === 'EATS',
        'filterStats[0] method': (s) => s.filterStats[0].method === 'GET',
        'filterStats[0] uriRegex': (s) => s.filterStats[0].uriRegex === '/endpoint/store.get_gateway',
        'filterStats[0] sampleCount': (s) => s.filterStats[0].sampleCount === 562,
        'filterStats[0] percentage': (s) => s.filterStats[0].percentage === 3.61,
    });

    // 9. Mixed object types
    const mixedObj = streamloader.loadJSON('mixedtypes.json');
    check(mixedObj, {
        'mixedObj is object': (s) => typeof s === 'object' && !Array.isArray(s),
        'obj.a == 1': (s) => s.obj && s.obj.a === 1,
        'arr[0] == 1': (s) => Array.isArray(s.arr) && s.arr[0] === 1,
        'str is hello': (s) => s.str === 'hello',
        'num is 42': (s) => s.num === 42,
        'bool is true': (s) => s.bool === true,
        'null is null': (s) => s.null === null,
    });

    // 10. Empty object
    const emptyObj = streamloader.loadJSON('emptyobj.json');
    check(emptyObj, {
        'emptyObj is object': (s) => typeof s === 'object' && !Array.isArray(s),
        'emptyObj has 0 keys': (s) => Object.keys(s).length === 0,
    });

    // 11. Deeply nested object
    const deepObj = streamloader.loadJSON('deep.json');
    check(deepObj, {
        'deepObj.a.b.c.d.e == 123': (s) => s.a && s.a.b && s.a.b.c && s.a.b.c.d && s.a.b.c.d.e === 123,
    });

    // 12. Object with special keys
    const specialKeysObj = streamloader.loadJSON('specialkeys.json');
    check(specialKeysObj, {
        'key "1" is 1': (s) => s['1'] === 1,
        'unicode key is yes': (s) => s['üñîçødë'] === 'yes',
        'special key is true': (s) => s['!@#$%^&*()'] === true,
    });

    // 13. NDJSON test
    const ndjsonArr = streamloader.loadJSON('ndjson.ndjson');
    check(ndjsonArr, {
        'ndjson is array': (s) => Array.isArray(s),
        'ndjson length 3': (s) => s.length === 3,
        'ndjson[0].a == 1': (s) => s[0].a === 1,
        'ndjson[2].c[1] == 4': (s) => Array.isArray(s[2].c) && s[2].c[1] === 4,
    });

    // 14. Array of primitives
    const primArr = streamloader.loadJSON('primarr.json');
    check(primArr, {
        'primArr is array': (s) => Array.isArray(s),
        'primArr[0] == 1': (s) => s[0] === 1,
        'primArr[2] == 3': (s) => s[2] === 3,
    });

    // CSV Tests (new)
    
    // 15. Basic CSV test
    const basicCSV = streamloader.loadCSV('basic.csv');
    console.log('basicCSV length:', basicCSV.length);
    console.log('basicCSV[0] (headers):', JSON.stringify(basicCSV[0]));
    console.log('basicCSV[1] (first row):', JSON.stringify(basicCSV[1]));
    check(basicCSV, {
        'CSV has 5 rows (header + 4 data)': (s) => s.length === 5,
        'CSV is array of arrays': (s) => Array.isArray(s) && Array.isArray(s[0]),
        'CSV header has 4 columns': (s) => s[0].length === 4,
        'CSV header name': (s) => s[0][0] === 'name',
        'CSV header age': (s) => s[0][1] === 'age',
        'CSV header city': (s) => s[0][2] === 'city',
        'CSV header active': (s) => s[0][3] === 'active',
        'CSV first data row name': (s) => s[1][0] === 'John Doe',
        'CSV first data row age': (s) => s[1][1] === '30',
        'CSV first data row city': (s) => s[1][2] === 'New York',
        'CSV first data row active': (s) => s[1][3] === 'true',
        'CSV last data row name': (s) => s[4][0] === 'Alice Brown',
        'CSV last data row age': (s) => s[4][1] === '28',
        'CSV last data row city': (s) => s[4][2] === 'Houston',
        'CSV last data row active': (s) => s[4][3] === 'false',
        'malformed CSV returns some records': (csv) => csv.length > 0,
    });

    // 16. Quoted CSV test  
    const quotedCSV = streamloader.loadCSV('quoted.csv');
    console.log('quotedCSV length:', quotedCSV.length);
    console.log('quotedCSV[1] (Widget A):', JSON.stringify(quotedCSV[1]));
    console.log('quotedCSV[2] (Gadget B):', JSON.stringify(quotedCSV[2]));
    check(quotedCSV, {
        'Quoted CSV has 5 rows': (s) => s.length === 5,
        'Quoted CSV header description': (s) => s[0][1] === 'description',
        'Quoted field with comma': (s) => s[1][1] === 'A simple, useful widget',
        'Quoted field with escaped quotes': (s) => s[2][1] === 'Complex gadget with "special" features',
        'Quoted field with newlines': (s) => s[3][1].includes('Contains commas, quotes, and\nnewlines'),
        'Unquoted field': (s) => s[4][1] === 'No quotes needed',
        'Widget A name': (s) => s[1][0] === 'Widget A',
        'Widget A price': (s) => s[1][2] === '19.99',
        'Widget A category': (s) => s[1][3] === 'electronics',
    });

    // 17. Empty CSV test
    const emptyCSV = streamloader.loadCSV('empty.csv');
    console.log('emptyCSV length:', emptyCSV.length);
    check(emptyCSV, {
        'Empty CSV returns empty array': (s) => Array.isArray(s) && s.length === 0,
    });

    // 18. Headers only CSV test
    const headersOnlyCSV = streamloader.loadCSV('headers_only.csv');
    console.log('headersOnlyCSV:', JSON.stringify(headersOnlyCSV));
    check(headersOnlyCSV, {
        'Headers only CSV has 1 row': (s) => s.length === 1,
        'Headers only CSV has 4 columns': (s) => s[0].length === 4,
        'Headers only first column': (s) => s[0][0] === 'id',
        'Headers only second column': (s) => s[0][1] === 'name',
        'Headers only third column': (s) => s[0][2] === 'email',
        'Headers only fourth column': (s) => s[0][3] === 'created_at',
    });

    // 19. Large CSV test
    const largeCSV = streamloader.loadCSV('large.csv');
    console.log('largeCSV length:', largeCSV.length);
    console.log('largeCSV[0] (headers):', JSON.stringify(largeCSV[0]));
    console.log('largeCSV[1] (first data row):', JSON.stringify(largeCSV[1]));
    console.log('largeCSV[10000] (last data row):', JSON.stringify(largeCSV[10000]));
    check(largeCSV, {
        'Large CSV has 10001 rows (header + 10000 data)': (s) => s.length === 10001,
        'Large CSV header has 10 columns': (s) => s[0].length === 10,
        'Large CSV header id': (s) => s[0][0] === 'id',
        'Large CSV header name': (s) => s[0][1] === 'name',
        'Large CSV header email': (s) => s[0][2] === 'email',
        'Large CSV header phone': (s) => s[0][3] === 'phone',
        'Large CSV header age': (s) => s[0][4] === 'age',
        'Large CSV header city': (s) => s[0][5] === 'city',
        'Large CSV header country': (s) => s[0][6] === 'country',
        'Large CSV header department': (s) => s[0][7] === 'department',
        'Large CSV header salary': (s) => s[0][8] === 'salary',
        'Large CSV header active': (s) => s[0][9] === 'active',
        'Large CSV first data row id': (s) => s[1][0] === '1',
        'Large CSV last data row id': (s) => s[10000][0] === '10000',
        'Large CSV all rows have 10 columns': (s) => s.every(row => row.length === 10),
        'Large CSV first data row has email': (s) => s[1][2].includes('@'),
        'Large CSV last data row has email': (s) => s[10000][2].includes('@'),
        'Large CSV first data row age is number': (s) => !isNaN(parseInt(s[1][4])),
        'Large CSV last data row age is number': (s) => !isNaN(parseInt(s[10000][4])),
    });

    // 20. Malformed CSV error test
    try {
        streamloader.loadCSV('malformed.csv');
        fail('Expected error for malformed CSV');
    } catch (e) {
        console.log('Malformed CSV error:', String(e));
        check(e, { 
            'error for malformed CSV': (err) => String(err).toLowerCase().includes('parse') || String(err).toLowerCase().includes('csv')
        });
    }

    // 21. Missing CSV file error test
    try {
        streamloader.loadCSV('nonexistent_file.csv');
        fail('Expected error for missing CSV file');
    } catch (e) {
        console.log('Missing CSV file error:', String(e));
        check(e, { 
            'error for missing CSV file': (err) => String(err).includes('nonexistent_file') || String(err).toLowerCase().includes('open')
        });
    }

    // File Loading Tests
    
    // 1. Normal case: valid file
    const fileContent = streamloader.loadText('test.txt');
    const expectedContent = `This is a test file for the loadText function.
It contains multiple lines.
And some special characters: !@#$%^&*() `;
    check(null, {
        'loaded file content matches expected': () => fileContent === expectedContent,
    });

    // 2. Error case: missing file
    try {
        streamloader.loadText('no_such_file.txt');
        fail('Expected error for missing file');
    } catch (e) {
        check(e, { 'error for missing file': (err) => String(err).includes('no_such_file') || String(err).includes('no such file') });
    }

    // 3. Edge case: empty file
    const emptyContent = streamloader.loadText('empty.txt');
    check(null, {
        'file with single space returns a space': () => emptyContent === ' ',
    });

    console.log('All JSON and CSV tests completed successfully!');

    // 21. All fields quoted
    const csvAllQuoted = streamloader.loadCSV('all_quoted.csv');
    console.log('csvAllQuoted:', JSON.stringify(csvAllQuoted));
    check(csvAllQuoted, {
        'all quoted loaded correctly': (r) => r.length === 3,
        'all quoted header correct': (r) => r[0][0] === 'name' && r[0][1] === 'description' && r[0][2] === 'price',
        'all quoted row 1 correct': (r) => r[1][0] === 'Widget A' && r[1][1] === 'A, simple widget' && r[1][2] === '19.99',
        'all quoted row 2 correct': (r) => r[2][0] === 'Gadget B' && r[2][1] === 'A complex gadget' && r[2][2] === '49.99',
    });

    // 22. CSV with special characters
    const csvSpecial = streamloader.loadCSV('specialchars.csv');
    console.log('csvSpecial raw output:', JSON.stringify(csvSpecial));
    if (csvSpecial.length > 1) {
        console.log('Failing row data:', JSON.stringify(csvSpecial[1]));
    }
    check(csvSpecial, {
        'csv with special characters has 2 rows': (s) => s.length === 2,
        'csv special characters row 1 correct': (s) => s[0][0] === 'header1' && s[0][1] === 'header2',
        'csv special characters row 2 correct': (s) => s[1][0] === 'value with "quotes"' && s[1][1] === 'value with ,comma',
    });

    // 23. Whitespace Handling
    const csvWhitespace = streamloader.loadCSV('quoted.csv');
    console.log('csvWhitespace:', JSON.stringify(csvWhitespace));
    check(csvWhitespace, {
        'csv with whitespace has correct rows': (s) => s.length === 5,
        'csv with whitespace row 1 correct': (s) => s[1][0] === 'Widget A' && s[1][1] === 'A simple, useful widget',
    });
}