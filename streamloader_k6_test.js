import streamloader from 'k6/x/streamloader';
import { check, fail } from 'k6';

export default function () {
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
} 