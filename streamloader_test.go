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
	objects, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	if len(objects) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(objects))
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
	objects, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	if len(objects) != 0 {
		t.Errorf("expected 0 objects, got %d", len(objects))
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
	objects, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	if len(objects) != 1000 {
		t.Errorf("expected 1000 objects, got %d", len(objects))
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
	objects, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	if _, ok := objects[0]["requestURI"]; ok {
		t.Errorf("did not expect requestURI key, but found %v", objects[0]["requestURI"])
	}
	if v, ok := objects[0]["request_uri"]; !ok || v != "/foo" {
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
	objects, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	if len(objects) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(objects))
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
	objects, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	if len(objects) != 3 {
		t.Fatalf("expected 3 objects, got %d", len(objects))
	}

	// Test first object (user1)
	o0 := objects[0]
	if key, ok := o0["_key"].(string); !ok || key != "user1" {
		t.Errorf("expected _key user1, got %v", o0["_key"])
	}
	if method, ok := o0["method"].(string); !ok || method != "GET" {
		t.Errorf("expected method GET, got %v", o0["method"])
	}
	if uri, ok := o0["requestURI"].(string); !ok || uri != "/user1" {
		t.Errorf("expected requestURI /user1, got %v", o0["requestURI"])
	}
	h, ok := o0["headers"].(map[string]interface{})
	if !ok || h["A"].(string) != "B" {
		t.Errorf("expected header A:B, got %v", o0["headers"])
	}
	if content, ok := o0["content"].(string); !ok || content != "user1_data" {
		t.Errorf("expected content user1_data, got %v", o0["content"])
	}

	// Test second object (user2)
	o1 := objects[1]
	if key, ok := o1["_key"].(string); !ok || key != "user2" {
		t.Errorf("expected _key user2, got %v", o1["_key"])
	}
	if method, ok := o1["method"].(string); !ok || method != "POST" {
		t.Errorf("expected method POST, got %v", o1["method"])
	}
	if uri, ok := o1["requestURI"].(string); !ok || uri != "/user2" {
		t.Errorf("expected requestURI /user2, got %v", o1["requestURI"])
	}
	h2, ok := o1["headers"].(map[string]interface{})
	if !ok || h2["C"].(string) != "D" {
		t.Errorf("expected header C:D, got %v", o1["headers"])
	}
	if content, ok := o1["content"].(string); !ok || content != "user2_data" {
		t.Errorf("expected content user2_data, got %v", o1["content"])
	}

	// Test third object (user3)
	o2 := objects[2]
	if key, ok := o2["_key"].(string); !ok || key != "user3" {
		t.Errorf("expected _key user3, got %v", o2["_key"])
	}
	if method, ok := o2["method"].(string); !ok || method != "PUT" {
		t.Errorf("expected method PUT, got %v", o2["method"])
	}
	if uri, ok := o2["requestURI"].(string); !ok || uri != "/user3" {
		t.Errorf("expected requestURI /user3, got %v", o2["requestURI"])
	}
	h3, ok := o2["headers"].(map[string]interface{})
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
	objects, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	if len(objects) != 6 {
		t.Fatalf("expected 6 objects, got %d", len(objects))
	}

	// Test string value
	foundString := false
	for _, obj := range objects {
		if key, ok := obj["_key"].(string); ok && key == "string_value" {
			if value, ok := obj["_value"].(string); !ok || value != "hello world" {
				t.Errorf("expected _value hello world, got %v", obj["_value"])
			}
			foundString = true
			break
		}
	}
	if !foundString {
		t.Error("did not find string_value object")
	}

	// Test number value
	foundNumber := false
	for _, obj := range objects {
		if key, ok := obj["_key"].(string); ok && key == "number_value" {
			if value, ok := obj["_value"].(float64); !ok || value != 42 {
				t.Errorf("expected _value 42, got %v", obj["_value"])
			}
			foundNumber = true
			break
		}
	}
	if !foundNumber {
		t.Error("did not find number_value object")
	}

	// Test boolean value
	foundBoolean := false
	for _, obj := range objects {
		if key, ok := obj["_key"].(string); ok && key == "boolean_value" {
			if value, ok := obj["_value"].(bool); !ok || !value {
				t.Errorf("expected _value true, got %v", obj["_value"])
			}
			foundBoolean = true
			break
		}
	}
	if !foundBoolean {
		t.Error("did not find boolean_value object")
	}

	// Test object value (should not have _value field)
	foundObject := false
	for _, obj := range objects {
		if key, ok := obj["_key"].(string); ok && key == "object_value" {
			if _, hasValue := obj["_value"]; hasValue {
				t.Error("object_value should not have _value field")
			}
			if nested, ok := obj["nested"].(string); !ok || nested != "data" {
				t.Errorf("expected nested data, got %v", obj["nested"])
			}
			foundObject = true
			break
		}
	}
	if !foundObject {
		t.Error("did not find object_value object")
	}
}
