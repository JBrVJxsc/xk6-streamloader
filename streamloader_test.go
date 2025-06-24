package streamloader

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestLoadJSON_ArrayFormat(t *testing.T) {
	jsonData := `[
	  {"method": "GET", "requestURI": "/foo", "headers": {"A": "B"}, "content": "abc"},
	  {"method": "POST", "requestURI": "/bar", "headers": {"C": "D"}, "content": "def"}
	]`

	tmpfile, err := os.CreateTemp("", "objects-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(arr))
	}
	objects := make([]map[string]any, len(arr))
	for i, v := range arr {
		obj, ok := v.(map[string]any)
		if !ok {
			t.Fatalf("expected element %d to be object, got %T", i, v)
		}
		objects[i] = obj
	}

	// First object
	o0 := objects[0]
	if m, ok := o0["method"].(string); !ok || m != "GET" {
		t.Errorf("expected method GET, got %v", o0["method"])
	}
	if uri, ok := o0["requestURI"].(string); !ok || uri != "/foo" {
		t.Errorf("expected requestURI /foo, got %v", o0["requestURI"])
	}
	h, ok := o0["headers"].(map[string]interface{})
	if !ok || h["A"].(string) != "B" {
		t.Errorf("expected header A:B, got %v", o0["headers"])
	}
	if c, ok := o0["content"].(string); !ok || c != "abc" {
		t.Errorf("expected content abc, got %v", o0["content"])
	}

	// Second object
	o1 := objects[1]
	if m, ok := o1["method"].(string); !ok || m != "POST" {
		t.Errorf("expected method POST, got %v", o1["method"])
	}
	if uri, ok := o1["requestURI"].(string); !ok || uri != "/bar" {
		t.Errorf("expected requestURI /bar, got %v", o1["requestURI"])
	}
	h2, ok := o1["headers"].(map[string]interface{})
	if !ok || h2["C"].(string) != "D" {
		t.Errorf("expected header C:D, got %v", o1["headers"])
	}
	if c, ok := o1["content"].(string); !ok || c != "def" {
		t.Errorf("expected content def, got %v", o1["content"])
	}
}

func TestLoadJSON_EmptyArray(t *testing.T) {
	jsonData := `[]`
	tmpfile, err := os.CreateTemp("", "objects-empty-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if len(arr) != 0 {
		t.Errorf("expected 0 objects, got %d", len(arr))
	}
}

func TestLoadJSON_InvalidJSON(t *testing.T) {
	jsonData := `{not valid json`
	tmpfile, err := os.CreateTemp("", "objects-bad-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	_, err = loader.LoadJSON(tmpfile.Name())
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestLoadJSON_MissingFile(t *testing.T) {
	loader := StreamLoader{}
	_, err := loader.LoadJSON("no_such_file.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadJSON_LargeArray(t *testing.T) {
	// Generate large JSON
	var b strings.Builder
	b.WriteString("[")
	for i := 0; i < 1000; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"method":"GET","requestURI":"/bulk/` + strconv.Itoa(i) + `","headers":{"X":"` + strconv.Itoa(i) + `"},"content":"` + strconv.Itoa(i) + `"}`)
	}
	b.WriteString("]")

	tmpfile, err := os.CreateTemp("", "objects-large-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(b.String())); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if len(arr) != 1000 {
		t.Errorf("expected 1000 objects, got %d", len(arr))
	}
	objects := make([]map[string]any, len(arr))
	for i, v := range arr {
		obj, ok := v.(map[string]any)
		if !ok {
			t.Fatalf("expected element %d to be object, got %T", i, v)
		}
		objects[i] = obj
	}
	if uri, ok := objects[0]["requestURI"].(string); !ok || uri != "/bulk/0" {
		t.Errorf("unexpected first object URI: %v", objects[0]["requestURI"])
	}
	if uri, ok := objects[999]["requestURI"].(string); !ok || uri != "/bulk/999" {
		t.Errorf("unexpected last object URI: %v", objects[999]["requestURI"])
	}
}

func TestLoadJSON_WrongFieldName(t *testing.T) {
	jsonData := `[{"method": "GET", "request_uri": "/foo", "headers": {"A": "B"}, "content": "abc"}]`
	tmpfile, err := os.CreateTemp("", "objects-wrongfield-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 object, got %d", len(arr))
	}
	obj, ok := arr[0].(map[string]any)
	if !ok {
		t.Fatalf("expected object, got %T", arr[0])
	}
	if _, ok := obj["requestURI"]; ok {
		t.Errorf("did not expect requestURI key, but found %v", obj["requestURI"])
	}
	if v, ok := obj["request_uri"]; !ok || v != "/foo" {
		t.Errorf("expected request_uri key with value /foo, got %v", v)
	}
}

func TestLoadJSON_ComplexStructure(t *testing.T) {
	jsonData := `[
		{
			"id": 1,
			"user": {
				"name": "Alice Johnson",
				"email": "alice@example.com",
				"profile": {
					"age": 30,
					"location": {
						"city": "New York",
						"country": "USA",
						"coordinates": {
							"lat": 40.7128,
							"lng": -74.0060
						}
					},
					"preferences": {
						"theme": "dark",
						"notifications": {
							"email": true,
							"push": false,
							"sms": true
						},
						"languages": ["en", "es", "fr"]
					}
				},
				"roles": ["admin", "user"],
				"metadata": {
					"created_at": "2023-01-15T10:30:00Z",
					"last_login": "2024-01-20T14:45:00Z",
					"login_count": 156
				}
			},
			"orders": [
				{
					"order_id": "ORD-001",
					"items": [
						{
							"product_id": "PROD-123",
							"name": "Laptop",
							"price": 1299.99,
							"quantity": 1,
							"specifications": {
								"cpu": "Intel i7",
								"ram": "16GB",
								"storage": "512GB SSD"
							}
						}
					],
					"shipping": {
						"address": {
							"street": "123 Main St",
							"city": "New York",
							"state": "NY",
							"zip": "10001"
						},
						"method": "express",
						"tracking": {
							"number": "TRK-789456",
							"status": "delivered",
							"events": [
								{
									"timestamp": "2024-01-18T09:00:00Z",
									"status": "shipped",
									"location": "Warehouse"
								}
							]
						}
					}
				}
			],
			"settings": {
				"privacy": {
					"profile_visibility": "public",
					"data_sharing": {
						"analytics": true,
						"marketing": false,
						"third_party": false
					}
				},
				"security": {
					"two_factor": true,
					"session_timeout": 3600,
					"allowed_ips": ["192.168.1.1", "10.0.0.1"]
				}
			}
		},
		{
			"id": 2,
			"user": {
				"name": "Bob Smith",
				"email": "bob@example.com",
				"profile": {
					"age": 25,
					"location": {
						"city": "Los Angeles",
						"country": "USA",
						"coordinates": {
							"lat": 34.0522,
							"lng": -118.2437
						}
					},
					"preferences": {
						"theme": "light",
						"notifications": {
							"email": false,
							"push": true,
							"sms": false
						},
						"languages": ["en"]
					}
				},
				"roles": ["user"],
				"metadata": {
					"created_at": "2023-06-10T08:15:00Z",
					"last_login": "2024-01-21T11:20:00Z",
					"login_count": 89
				}
			},
			"orders": [],
			"settings": {
				"privacy": {
					"profile_visibility": "private",
					"data_sharing": {
						"analytics": false,
						"marketing": false,
						"third_party": false
					}
				},
				"security": {
					"two_factor": false,
					"session_timeout": 1800,
					"allowed_ips": []
				}
			}
		}
	]`

	tmpfile, err := os.CreateTemp("", "objects-complex-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(arr))
	}
	objects := make([]map[string]any, len(arr))
	for i, v := range arr {
		obj, ok := v.(map[string]any)
		if !ok {
			t.Fatalf("expected element %d to be object, got %T", i, v)
		}
		objects[i] = obj
	}

	// Test first object with complex nested structure
	o0 := objects[0]

	// Test top-level fields
	if id, ok := o0["id"].(float64); !ok || id != 1 {
		t.Errorf("expected id 1, got %v", o0["id"])
	}

	// Test nested user object
	user, ok := o0["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected user to be a map, got %T", o0["user"])
	}
	if name, ok := user["name"].(string); !ok || name != "Alice Johnson" {
		t.Errorf("expected user name Alice Johnson, got %v", user["name"])
	}
	if email, ok := user["email"].(string); !ok || email != "alice@example.com" {
		t.Errorf("expected user email alice@example.com, got %v", user["email"])
	}

	// Test deeply nested profile structure
	profile, ok := user["profile"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected profile to be a map, got %T", user["profile"])
	}
	if age, ok := profile["age"].(float64); !ok || age != 30 {
		t.Errorf("expected age 30, got %v", profile["age"])
	}

	// Test location coordinates
	location, ok := profile["location"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected location to be a map, got %T", profile["location"])
	}
	coordinates, ok := location["coordinates"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected coordinates to be a map, got %T", location["coordinates"])
	}
	if lat, ok := coordinates["lat"].(float64); !ok || lat != 40.7128 {
		t.Errorf("expected lat 40.7128, got %v", coordinates["lat"])
	}
	if lng, ok := coordinates["lng"].(float64); !ok || lng != -74.0060 {
		t.Errorf("expected lng -74.0060, got %v", coordinates["lng"])
	}

	// Test preferences with boolean values
	preferences, ok := profile["preferences"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected preferences to be a map, got %T", profile["preferences"])
	}
	notifications, ok := preferences["notifications"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected notifications to be a map, got %T", preferences["notifications"])
	}
	if email, ok := notifications["email"].(bool); !ok || !email {
		t.Errorf("expected email notification true, got %v", notifications["email"])
	}
	if push, ok := notifications["push"].(bool); !ok || push {
		t.Errorf("expected push notification false, got %v", notifications["push"])
	}

	// Test array of strings
	languages, ok := preferences["languages"].([]interface{})
	if !ok {
		t.Fatalf("expected languages to be an array, got %T", preferences["languages"])
	}
	if len(languages) != 3 {
		t.Errorf("expected 3 languages, got %d", len(languages))
	}
	if lang0, ok := languages[0].(string); !ok || lang0 != "en" {
		t.Errorf("expected first language en, got %v", languages[0])
	}

	// Test array of strings in roles
	roles, ok := user["roles"].([]interface{})
	if !ok {
		t.Fatalf("expected roles to be an array, got %T", user["roles"])
	}
	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}
	if role0, ok := roles[0].(string); !ok || role0 != "admin" {
		t.Errorf("expected first role admin, got %v", roles[0])
	}

	// Test orders array with complex nested objects
	orders, ok := o0["orders"].([]interface{})
	if !ok {
		t.Fatalf("expected orders to be an array, got %T", o0["orders"])
	}
	if len(orders) != 1 {
		t.Errorf("expected 1 order, got %d", len(orders))
	}

	order, ok := orders[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected order to be a map, got %T", orders[0])
	}
	if orderID, ok := order["order_id"].(string); !ok || orderID != "ORD-001" {
		t.Errorf("expected order_id ORD-001, got %v", order["order_id"])
	}

	// Test items array within order
	items, ok := order["items"].([]interface{})
	if !ok {
		t.Fatalf("expected items to be an array, got %T", order["items"])
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}

	item, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected item to be a map, got %T", items[0])
	}
	if price, ok := item["price"].(float64); !ok || price != 1299.99 {
		t.Errorf("expected price 1299.99, got %v", item["price"])
	}

	// Test second object with empty orders array
	o1 := objects[1]
	if id, ok := o1["id"].(float64); !ok || id != 2 {
		t.Errorf("expected id 2, got %v", o1["id"])
	}

	user1, ok := o1["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected user to be a map, got %T", o1["user"])
	}
	if name, ok := user1["name"].(string); !ok || name != "Bob Smith" {
		t.Errorf("expected user name Bob Smith, got %v", user1["name"])
	}

	// Test empty orders array
	orders1, ok := o1["orders"].([]interface{})
	if !ok {
		t.Fatalf("expected orders to be an array, got %T", o1["orders"])
	}
	if len(orders1) != 0 {
		t.Errorf("expected 0 orders, got %d", len(orders1))
	}

	// Test settings with nested boolean values
	settings1, ok := o1["settings"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected settings to be a map, got %T", o1["settings"])
	}
	privacy1, ok := settings1["privacy"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected privacy to be a map, got %T", settings1["privacy"])
	}
	dataSharing1, ok := privacy1["data_sharing"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data_sharing to be a map, got %T", privacy1["data_sharing"])
	}
	if analytics, ok := dataSharing1["analytics"].(bool); !ok || analytics {
		t.Errorf("expected analytics false, got %v", dataSharing1["analytics"])
	}
}

func TestLoadJSON_ObjectFormat(t *testing.T) {
	jsonData := `{
		"user1": {"method": "GET", "requestURI": "/user1", "headers": {"A": "B"}, "content": "user1_data"},
		"user2": {"method": "POST", "requestURI": "/user2", "headers": {"C": "D"}, "content": "user2_data"},
		"user3": {"method": "PUT", "requestURI": "/user3", "headers": {"E": "F"}, "content": "user3_data"}
	}`

	tmpfile, err := os.CreateTemp("", "objects-object-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	objects, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected object, got %T", result)
	}

	if len(objects) != 3 {
		t.Fatalf("expected 3 objects, got %d", len(objects))
	}

	// Test user1
	o0, ok := objects["user1"].(map[string]any)
	if !ok {
		t.Fatalf("expected user1 to be a map, got %T", objects["user1"])
	}
	if method, ok := o0["method"].(string); !ok || method != "GET" {
		t.Errorf("expected method GET, got %v", o0["method"])
	}
	if uri, ok := o0["requestURI"].(string); !ok || uri != "/user1" {
		t.Errorf("expected requestURI /user1, got %v", o0["requestURI"])
	}
	h, ok := o0["headers"].(map[string]any)
	if !ok || h["A"].(string) != "B" {
		t.Errorf("expected header A:B, got %v", o0["headers"])
	}
	if content, ok := o0["content"].(string); !ok || content != "user1_data" {
		t.Errorf("expected content user1_data, got %v", o0["content"])
	}

	// Test user2
	o1, ok := objects["user2"].(map[string]any)
	if !ok {
		t.Fatalf("expected user2 to be a map, got %T", objects["user2"])
	}
	if method, ok := o1["method"].(string); !ok || method != "POST" {
		t.Errorf("expected method POST, got %v", o1["method"])
	}
	if uri, ok := o1["requestURI"].(string); !ok || uri != "/user2" {
		t.Errorf("expected requestURI /user2, got %v", o1["requestURI"])
	}
	h2, ok := o1["headers"].(map[string]any)
	if !ok || h2["C"].(string) != "D" {
		t.Errorf("expected header C:D, got %v", o1["headers"])
	}
	if content, ok := o1["content"].(string); !ok || content != "user2_data" {
		t.Errorf("expected content user2_data, got %v", o1["content"])
	}

	// Test user3
	o2, ok := objects["user3"].(map[string]any)
	if !ok {
		t.Fatalf("expected user3 to be a map, got %T", objects["user3"])
	}
	if method, ok := o2["method"].(string); !ok || method != "PUT" {
		t.Errorf("expected method PUT, got %v", o2["method"])
	}
	if uri, ok := o2["requestURI"].(string); !ok || uri != "/user3" {
		t.Errorf("expected requestURI /user3, got %v", o2["requestURI"])
	}
	h3, ok := o2["headers"].(map[string]any)
	if !ok || h3["E"].(string) != "F" {
		t.Errorf("expected header E:F, got %v", o2["headers"])
	}
	if content, ok := o2["content"].(string); !ok || content != "user3_data" {
		t.Errorf("expected content user3_data, got %v", o2["content"])
	}
}

func TestLoadJSON_ObjectFormatWithNonObjectValues(t *testing.T) {
	jsonData := `{
		"string_value": "hello world",
		"number_value": 42,
		"boolean_value": true,
		"null_value": null,
		"array_value": [1, 2, 3],
		"object_value": {"nested": "data"}
	}`

	tmpfile, err := os.CreateTemp("", "objects-mixed-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	objects, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected object, got %T", result)
	}

	if len(objects) != 6 {
		t.Fatalf("expected 6 objects, got %d", len(objects))
	}

	// Test string value
	if v, ok := objects["string_value"].(string); !ok || v != "hello world" {
		t.Errorf("expected string_value to be 'hello world', got %v", objects["string_value"])
	}
	// Test number value
	if v, ok := objects["number_value"].(float64); !ok || v != 42 {
		t.Errorf("expected number_value to be 42, got %v", objects["number_value"])
	}
	// Test boolean value
	if v, ok := objects["boolean_value"].(bool); !ok || !v {
		t.Errorf("expected boolean_value to be true, got %v", objects["boolean_value"])
	}
	// Test null value
	if objects["null_value"] != nil {
		t.Errorf("expected null_value to be nil, got %v", objects["null_value"])
	}
	// Test array value
	if arr, ok := objects["array_value"].([]interface{}); !ok || len(arr) != 3 || arr[0].(float64) != 1 || arr[1].(float64) != 2 || arr[2].(float64) != 3 {
		t.Errorf("expected array_value to be [1,2,3], got %v", objects["array_value"])
	}
	// Test object value
	if obj, ok := objects["object_value"].(map[string]any); !ok || obj["nested"] != "data" {
		t.Errorf("expected object_value.nested to be 'data', got %v", objects["object_value"])
	}
}

func TestLoadJSON_RecordingStatsFormat(t *testing.T) {
	jsonData := `{
    "recordingId": "18aebc27-ef24-42a0-aee4-5e01f8ac6049",
    "domain": "EATS_CUSTOMER",
    "startTime": "2025-06-22T06:54:03.749Z",
    "endTime": "2025-06-22T06:56:04.299Z",
    "expireAt": "2025-06-29T06:56:04.299Z",
    "ttl": 604800,
    "totalMatchedCount": 15550,
    "totalUnmatchedCount": 2450,
    "totalProcessedCount": 18000,
    "totalDataSize": 57911089,
    "matchedPercentage": 86.39,
    "unmatchedPercentage": 13.61,
    "filterStats": [
      {
        "domain": "EATS",
        "method": "GET",
        "uriRegex": "/endpoint/store.get_gateway",
        "sampleCount": 562,
        "count": 562,
        "dataSize": 2185148,
        "percentage": 3.61,
        "weight": 361
      }
    ]
  }`

	tmpfile, err := os.CreateTemp("", "objects-recordingstats-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	objects, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected object, got %T", result)
	}

	if len(objects) != 13 {
		t.Fatalf("expected 13 objects, got %d", len(objects))
	}

	tests := []struct {
		key  string
		want interface{}
	}{
		{"recordingId", "18aebc27-ef24-42a0-aee4-5e01f8ac6049"},
		{"domain", "EATS_CUSTOMER"},
		{"startTime", "2025-06-22T06:54:03.749Z"},
		{"endTime", "2025-06-22T06:56:04.299Z"},
		{"expireAt", "2025-06-29T06:56:04.299Z"},
		{"ttl", 604800.0},
		{"totalMatchedCount", 15550.0},
		{"totalUnmatchedCount", 2450.0},
		{"totalProcessedCount", 18000.0},
		{"totalDataSize", 57911089.0},
		{"matchedPercentage", 86.39},
		{"unmatchedPercentage", 13.61},
	}

	for _, tt := range tests {
		v, ok := objects[tt.key]
		if !ok {
			t.Errorf("object with key %q not found", tt.key)
			continue
		}
		if v != tt.want {
			t.Errorf("for key %q, expected value %v, got %v", tt.key, tt.want, v)
		}
	}

	// Check filterStats
	fs, ok := objects["filterStats"].([]interface{})
	if !ok || len(fs) != 1 {
		t.Fatalf("expected filterStats to be array of length 1, got %v", objects["filterStats"])
	}
	fs0, ok := fs[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected filterStats[0] to be a map, got %T", fs[0])
	}
	if v, ok := fs0["domain"].(string); !ok || v != "EATS" {
		t.Errorf("expected filterStats[0].domain EATS, got %v", fs0["domain"])
	}
	if v, ok := fs0["method"].(string); !ok || v != "GET" {
		t.Errorf("expected filterStats[0].method GET, got %v", fs0["method"])
	}
	if v, ok := fs0["uriRegex"].(string); !ok || v != "/endpoint/store.get_gateway" {
		t.Errorf("expected filterStats[0].uriRegex /endpoint/store.get_gateway, got %v", fs0["uriRegex"])
	}
	if v, ok := fs0["sampleCount"].(float64); !ok || v != 562 {
		t.Errorf("expected filterStats[0].sampleCount 562, got %v", fs0["sampleCount"])
	}
	if v, ok := fs0["percentage"].(float64); !ok || v != 3.61 {
		t.Errorf("expected filterStats[0].percentage 3.61, got %v", fs0["percentage"])
	}
}

func TestLoadJSON_MixedObjectTypes(t *testing.T) {
	jsonData := `{
		"obj": {"a": 1, "b": [2,3]},
		"arr": [1,2,3],
		"str": "hello",
		"num": 42,
		"bool": true,
		"null": null
	}`
	tmpfile, err := os.CreateTemp("", "objects-mixedtypes-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	obj, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected object, got %T", result)
	}
	if v, ok := obj["obj"].(map[string]any); !ok || v["a"].(float64) != 1 {
		t.Errorf("expected obj.a == 1, got %v", obj["obj"])
	}
	if arr, ok := obj["arr"].([]interface{}); !ok || len(arr) != 3 || arr[0].(float64) != 1 {
		t.Errorf("expected arr [1,2,3], got %v", obj["arr"])
	}
	if s, ok := obj["str"].(string); !ok || s != "hello" {
		t.Errorf("expected str 'hello', got %v", obj["str"])
	}
	if n, ok := obj["num"].(float64); !ok || n != 42 {
		t.Errorf("expected num 42, got %v", obj["num"])
	}
	if b, ok := obj["bool"].(bool); !ok || !b {
		t.Errorf("expected bool true, got %v", obj["bool"])
	}
	if obj["null"] != nil {
		t.Errorf("expected null to be nil, got %v", obj["null"])
	}
}

func TestLoadJSON_EmptyObject(t *testing.T) {
	jsonData := `{}`
	tmpfile, err := os.CreateTemp("", "objects-emptyobj-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	obj, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected object, got %T", result)
	}
	if len(obj) != 0 {
		t.Errorf("expected empty object, got %v", obj)
	}
}

func TestLoadJSON_DeeplyNestedObject(t *testing.T) {
	jsonData := `{"a":{"b":{"c":{"d":{"e":123}}}}}`
	tmpfile, err := os.CreateTemp("", "objects-deep-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	obj, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected object, got %T", result)
	}
	if v := obj["a"].(map[string]any)["b"].(map[string]any)["c"].(map[string]any)["d"].(map[string]any)["e"]; v != 123.0 {
		t.Errorf("expected deeply nested value 123, got %v", v)
	}
}

func TestLoadJSON_ObjectWithSpecialKeys(t *testing.T) {
	jsonData := `{"1":1,"üñîçødë":"yes","!@#$%^&*()":true}`
	tmpfile, err := os.CreateTemp("", "objects-specialkeys-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	obj, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected object, got %T", result)
	}
	if obj["1"].(float64) != 1 {
		t.Errorf("expected key '1' to be 1, got %v", obj["1"])
	}
	if obj["üñîçødë"].(string) != "yes" {
		t.Errorf("expected unicode key to be 'yes', got %v", obj["üñîçødë"])
	}
	if obj["!@#$%^&*()"].(bool) != true {
		t.Errorf("expected special key to be true, got %v", obj["!@#$%^&*()"])
	}
}

func TestLoadJSON_NDJSON(t *testing.T) {
	ndjson := `{"a":1}
{"b":2}
{"c":[3,4]}
`
	tmpfile, err := os.CreateTemp("", "objects-ndjson-*.ndjson")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(ndjson)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	var arr []map[string]any
	switch v := result.(type) {
	case []map[string]any:
		arr = v
	case []interface{}:
		arr = make([]map[string]any, len(v))
		for i, elem := range v {
			m, ok := elem.(map[string]any)
			if !ok {
				t.Fatalf("expected map at %d, got %T", i, elem)
			}
			arr[i] = m
		}
	default:
		t.Fatalf("expected array, got %T", result)
	}
	if len(arr) != 3 {
		t.Errorf("expected 3 NDJSON objects, got %d", len(arr))
	}
	if arr[0]["a"].(float64) != 1 {
		t.Errorf("expected first NDJSON object a==1, got %v", arr[0])
	}
	if arr[2]["c"].([]interface{})[1].(float64) != 4 {
		t.Errorf("expected third NDJSON object c[1]==4, got %v", arr[2])
	}
}

func TestLoadJSON_ArrayOfPrimitives(t *testing.T) {
	jsonData := `[1,2,3]`
	tmpfile, err := os.CreateTemp("", "objects-primarr-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	result, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array of primitives, got %T", result)
	}
	if len(arr) != 3 || arr[0].(float64) != 1 {
		t.Errorf("expected [1,2,3], got %v", arr)
	}
}
