package blueconfig

import (
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sfi2k7/microweb"
)

// ============================================================================
// Test Helpers
// ============================================================================

func createTestTree(t *testing.T) (*Tree, string) {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	tr, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: dbPath,
		Port:                  0, // Don't start server in tests
		Token:                 "test-token",
	})
	if err != nil {
		t.Fatalf("Failed to create test tree: %v", err)
	}

	return tr, dbPath
}

func cleanup(t *testing.T, tr *Tree) {
	t.Helper()
	if err := tr.Close(); err != nil {
		t.Errorf("Failed to close tree: %v", err)
	}
}

// ============================================================================
// Path Utilities Tests
// ============================================================================

func TestFixPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty path", "", "root"},
		{"root path", "/", "root"},
		{"root string", "root", "root"},
		{"simple path", "hello", "root/hello"},
		{"path with leading slash", "/hello", "root/hello"},
		{"path with trailing slash", "hello/", "root/hello"},
		{"path with double slashes", "hello//world", "root/hello/world"},
		{"complex path", "/hello/world/test/", "root/hello/world/test"},
		{"already prefixed", "root/hello/world", "root/hello/world"},
		{"multiple double slashes", "//hello///world//", "root/hello/world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fixpath(tt.input)
			if result != tt.expected {
				t.Errorf("fixpath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParsePath(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		offset        int
		expectedNode  string
		expectedProp  string
		expectedValue string
		expectError   bool
	}{
		{
			name:          "offset 0",
			path:          "root/hello/world",
			offset:        0,
			expectedNode:  "root/hello/world",
			expectedProp:  "",
			expectedValue: "",
			expectError:   false,
		},
		{
			name:          "offset 1 - extract prop",
			path:          "root/hello/world/name",
			offset:        1,
			expectedNode:  "root/hello/world",
			expectedProp:  "name",
			expectedValue: "",
			expectError:   false,
		},
		{
			name:          "offset 2 - extract prop and value",
			path:          "root/hello/world/name/john",
			offset:        2,
			expectedNode:  "root/hello/world",
			expectedProp:  "name",
			expectedValue: "john",
			expectError:   false,
		},
		{
			name:        "invalid offset - too large",
			path:        "root/hello",
			offset:      5,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, prop, value, err := parsePath(tt.path, tt.offset)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if node != tt.expectedNode {
				t.Errorf("node = %q, want %q", node, tt.expectedNode)
			}
			if prop != tt.expectedProp {
				t.Errorf("prop = %q, want %q", prop, tt.expectedProp)
			}
			if value != tt.expectedValue {
				t.Errorf("value = %q, want %q", value, tt.expectedValue)
			}
		})
	}
}

func TestSanitizeBucketName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no special chars", "hello", "hello"},
		{"with space", "hello world", "hello_world"},
		{"with quotes", `hello"world`, "hello_world"},
		{"with single quotes", "hello'world", "hello_world"},
		{"with backticks", "hello`world", "hello_world"},
		{"multiple special chars", `hello "world" test`, "hello__world__test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeBucketName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeBucketName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Tree Construction Tests
// ============================================================================

func TestNewOrOpenTree(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Create new tree
	tr1, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: dbPath,
		Port:                  9999,
		Token:                 "test-token",
	})
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	if tr1.port != 9999 {
		t.Errorf("port = %d, want 9999", tr1.port)
	}
	if tr1.token != "test-token" {
		t.Errorf("token = %q, want %q", tr1.token, "test-token")
	}

	tr1.Close()

	// Open existing tree
	tr2, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: dbPath,
		Port:                  8888,
		Token:                 "new-token",
	})
	if err != nil {
		t.Fatalf("Failed to open existing tree: %v", err)
	}
	defer tr2.Close()

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestTreeClose(t *testing.T) {
	tr, _ := createTestTree(t)

	err := tr.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Calling close again should still work (bbolt handles this)
	err = tr.Close()
	if err != nil {
		t.Errorf("Second Close() returned error: %v", err)
	}
}

// ============================================================================
// Path Operations Tests
// ============================================================================

func TestCreatePath(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	tests := []struct {
		name string
		path string
	}{
		{"simple path", "/apps/blue"},
		{"nested path", "/apps/blue/smartq/web"},
		{"deep nesting", "/level1/level2/level3/level4/level5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tr.CreatePath(tt.path)
			if err != nil {
				t.Errorf("CreatePath(%q) returned error: %v", tt.path, err)
			}

			// Verify path exists by trying to get nodes
			_, err = tr.GetNodesInPath(tt.path)
			if err != nil {
				t.Errorf("Path %q was not created successfully: %v", tt.path, err)
			}
		})
	}
}

func TestGetNodesInPath(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	// Create test structure
	tr.CreatePath("/cities/usa/ny")
	tr.CreatePath("/cities/usa/ca")
	tr.CreatePath("/cities/usa/tx")
	tr.CreatePath("/cities/uk/london")

	tests := []struct {
		name          string
		path          string
		expectedNodes []string
	}{
		{
			name:          "root cities",
			path:          "/cities",
			expectedNodes: []string{"usa", "uk"},
		},
		{
			name:          "usa cities",
			path:          "/cities/usa",
			expectedNodes: []string{"ny", "ca", "tx"},
		},
		{
			name:          "uk cities",
			path:          "/cities/uk",
			expectedNodes: []string{"london"},
		},
		{
			name:          "leaf node",
			path:          "/cities/usa/ny",
			expectedNodes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := tr.GetNodesInPath(tt.path)
			if err != nil {
				t.Errorf("GetNodesInPath(%q) returned error: %v", tt.path, err)
				return
			}

			if len(nodes) != len(tt.expectedNodes) {
				t.Errorf("GetNodesInPath(%q) returned %d nodes, want %d", tt.path, len(nodes), len(tt.expectedNodes))
				return
			}

			// Convert to map for easier comparison
			nodeMap := make(map[string]bool)
			for _, node := range nodes {
				nodeMap[node] = true
			}

			for _, expected := range tt.expectedNodes {
				if !nodeMap[expected] {
					t.Errorf("Expected node %q not found in result", expected)
				}
			}
		})
	}
}

func TestDeleteNode(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	// Setup test data
	tr.CreatePath("/apps/blue/service1")
	tr.CreatePath("/apps/blue/service2")
	tr.SetValue("/apps/blue/service1/port", "8080")

	t.Run("delete leaf node", func(t *testing.T) {
		err := tr.DeleteNode("/apps/blue/service2", false)
		if err != nil {
			t.Errorf("DeleteNode returned error: %v", err)
		}

		// Verify it's deleted
		nodes, _ := tr.GetNodesInPath("/apps/blue")
		for _, node := range nodes {
			if node == "service2" {
				t.Error("service2 should have been deleted")
			}
		}
	})

	t.Run("delete node with children without force", func(t *testing.T) {
		tr.CreatePath("/apps/blue/service3/nested")
		err := tr.DeleteNode("/apps/blue/service3", false)
		if err == nil {
			t.Error("Expected error when deleting node with children without force")
		}
	})

	t.Run("delete node with children with force", func(t *testing.T) {
		tr.CreatePath("/apps/blue/service4/nested")
		err := tr.DeleteNode("/apps/blue/service4", true)
		if err != nil {
			t.Errorf("DeleteNode with force returned error: %v", err)
		}

		// Verify it's deleted
		nodes, _ := tr.GetNodesInPath("/apps/blue")
		for _, node := range nodes {
			if node == "service4" {
				t.Error("service4 should have been deleted")
			}
		}
	})

	t.Run("cannot delete root", func(t *testing.T) {
		err := tr.DeleteNode("/root", false)
		if err == nil {
			t.Error("Expected error when deleting root node")
		}
	})
}

// ============================================================================
// Property Operations Tests
// ============================================================================

func TestSetAndGetValue(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	tr.CreatePath("/config/app")

	tests := []struct {
		name  string
		path  string
		value string
	}{
		{"simple value", "/config/app/port", "8080"},
		{"string value", "/config/app/name", "MyApp"},
		{"numeric value", "/config/app/timeout", "30"},
		{"complex value", "/config/app/url", "https://example.com/api/v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set value
			err := tr.SetValue(tt.path, tt.value)
			if err != nil {
				t.Errorf("SetValue(%q, %q) returned error: %v", tt.path, tt.value, err)
				return
			}

			// Get value
			result, err := tr.GetValue(tt.path)
			if err != nil {
				t.Errorf("GetValue(%q) returned error: %v", tt.path, err)
				return
			}

			if result != tt.value {
				t.Errorf("GetValue(%q) = %q, want %q", tt.path, result, tt.value)
			}
		})
	}
}

func TestSetValues(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	tr.CreatePath("/person/john")

	values := map[string]interface{}{
		"name":  "John Doe",
		"age":   30,
		"email": "john@example.com",
		"city":  "New York",
	}

	err := tr.SetValues("/person/john", values)
	if err != nil {
		t.Fatalf("SetValues returned error: %v", err)
	}

	// Verify all values were set
	for key, expectedValue := range values {
		result, err := tr.GetValue(fmt.Sprintf("/person/john/%s", key))
		if err != nil {
			t.Errorf("GetValue for %q returned error: %v", key, err)
			continue
		}

		expectedStr := fmt.Sprintf("%v", expectedValue)
		if result != expectedStr {
			t.Errorf("Value for %q = %q, want %q", key, result, expectedStr)
		}
	}
}

func TestCreateNodeWithProps(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	// Create node with properties in single operation
	values := map[string]interface{}{
		"name":    "Alice",
		"age":     25,
		"city":    "Boston",
		"active":  true,
		"balance": 100.50,
	}

	err := tr.CreateNodeWithProps("/users/alice", values)
	if err != nil {
		t.Fatalf("CreateNodeWithProps returned error: %v", err)
	}

	// Verify all values were set
	for key, expectedValue := range values {
		result, err := tr.GetValue(fmt.Sprintf("/users/alice/%s", key))
		if err != nil {
			t.Errorf("GetValue for %q returned error: %v", key, err)
			continue
		}

		expectedStr := fmt.Sprintf("%v", expectedValue)
		if result != expectedStr {
			t.Errorf("Value for %q = %q, want %q", key, result, expectedStr)
		}
	}

	// Verify the node exists in the tree
	nodes, err := tr.GetNodesInPath("/users")
	if err != nil {
		t.Fatalf("GetNodesInPath returned error: %v", err)
	}

	if len(nodes) != 1 || nodes[0] != "alice" {
		t.Errorf("Expected node 'alice' under /users, got %v", nodes)
	}
}

func TestSetNodeWithProps(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	// SetNodeWithProps should work same as CreateNodeWithProps
	values := map[string]interface{}{
		"host": "localhost",
		"port": 5432,
		"db":   "testdb",
	}

	err := tr.SetNodeWithProps("/config/database", values)
	if err != nil {
		t.Fatalf("SetNodeWithProps returned error: %v", err)
	}

	// Verify values
	result, err := tr.GetAllPropsWithValues("/config/database")
	if err != nil {
		t.Fatalf("GetAllPropsWithValues returned error: %v", err)
	}

	for key, expectedValue := range values {
		expectedStr := fmt.Sprintf("%v", expectedValue)
		if result[key] != expectedStr {
			t.Errorf("Value for %q = %q, want %q", key, result[key], expectedStr)
		}
	}
}

func TestBatchCreateNodes(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	// Create multiple nodes in single transaction
	nodes := map[string]map[string]interface{}{
		"/users/bob": {
			"name":  "Bob",
			"age":   30,
			"email": "bob@example.com",
		},
		"/users/carol": {
			"name":  "Carol",
			"age":   28,
			"email": "carol@example.com",
		},
		"/config/app": {
			"port":    8080,
			"host":    "0.0.0.0",
			"timeout": 30,
		},
	}

	err := tr.BatchCreateNodes(nodes)
	if err != nil {
		t.Fatalf("BatchCreateNodes returned error: %v", err)
	}

	// Verify all nodes were created
	for path, expectedValues := range nodes {
		result, err := tr.GetAllPropsWithValues(path)
		if err != nil {
			t.Errorf("GetAllPropsWithValues for %q returned error: %v", path, err)
			continue
		}

		for key, expectedValue := range expectedValues {
			expectedStr := fmt.Sprintf("%v", expectedValue)
			if result[key] != expectedStr {
				t.Errorf("Value at %s/%s = %q, want %q", path, key, result[key], expectedStr)
			}
		}
	}

	// Verify nodes exist in hierarchy
	users, err := tr.GetNodesInPath("/users")
	if err != nil {
		t.Fatalf("GetNodesInPath(/users) returned error: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestGetAllProps(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	tr.CreatePath("/config/db")
	tr.SetValue("/config/db/host", "localhost")
	tr.SetValue("/config/db/port", "5432")
	tr.SetValue("/config/db/database", "mydb")

	props, err := tr.GetAllProps("/config/db")
	if err != nil {
		t.Fatalf("GetAllProps returned error: %v", err)
	}

	expectedProps := map[string]bool{
		"host":     true,
		"port":     true,
		"database": true,
	}

	if len(props) != len(expectedProps) {
		t.Errorf("GetAllProps returned %d props, want %d", len(props), len(expectedProps))
	}

	for _, prop := range props {
		if !expectedProps[prop] {
			t.Errorf("Unexpected prop: %q", prop)
		}
	}
}

func TestGetAllPropsWithValues(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	tr.CreatePath("/server/web")
	expected := map[string]string{
		"host": "0.0.0.0",
		"port": "8080",
		"ssl":  "true",
	}

	for k, v := range expected {
		tr.SetValue(fmt.Sprintf("/server/web/%s", k), v)
	}

	result, err := tr.GetAllPropsWithValues("/server/web")
	if err != nil {
		t.Fatalf("GetAllPropsWithValues returned error: %v", err)
	}

	if len(result) != len(expected) {
		t.Errorf("Got %d props, want %d", len(result), len(expected))
	}

	for k, v := range expected {
		if result[k] != v {
			t.Errorf("Value for %q = %q, want %q", k, result[k], v)
		}
	}
}

func TestDeleteValue(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	tr.CreatePath("/config/app")
	tr.SetValue("/config/app/port", "8080")
	tr.SetValue("/config/app/host", "localhost")

	// Delete one value
	err := tr.DeleteValue("/config/app", "port")
	if err != nil {
		t.Fatalf("DeleteValue returned error: %v", err)
	}

	// Verify it's deleted
	hasValue, err := tr.HasValue("/config/app", "port")
	if err != nil {
		t.Fatalf("HasValue returned error: %v", err)
	}
	if hasValue {
		t.Error("Value should have been deleted")
	}

	// Verify other value still exists
	hasValue, err = tr.HasValue("/config/app", "host")
	if err != nil {
		t.Fatalf("HasValue returned error: %v", err)
	}
	if !hasValue {
		t.Error("Other value should still exist")
	}
}

func TestHasValue(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	tr.CreatePath("/config/test")
	tr.SetValue("/config/test/exists", "yes")

	tests := []struct {
		name     string
		path     string
		prop     string
		expected bool
	}{
		{"existing property", "/config/test", "exists", true},
		{"non-existing property", "/config/test", "notexists", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tr.HasValue(tt.path, tt.prop)
			if err != nil {
				t.Errorf("HasValue returned error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("HasValue(%q, %q) = %v, want %v", tt.path, tt.prop, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// HTTP Handler Tests
// ============================================================================

func setupTestServer(t *testing.T) (*Tree, *microweb.Router) {
	t.Helper()

	tr, _ := createTestTree(t)

	// Setup test data
	tr.CreatePath("/test/node1")
	tr.CreatePath("/test/node2")
	tr.SetValue("/test/node1/key1", "value1")
	tr.SetValue("/test/node1/key2", "value2")

	web := microweb.New()
	web.Use(tr.authMiddleware)
	web.Get("/", tr.handleGetRequest)
	web.Post("/", tr.handlePostRequest)

	return tr, web
}

func TestAuthMiddleware(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	tests := []struct {
		name           string
		path           string
		token          string
		expectAuth     bool
		skipTokenCheck bool
	}{
		{"valid token", "/test", "test-token", true, false},
		{"invalid token", "/test", "wrong-token", false, false},
		{"public path", "/public/test", "", true, true},
		{"no token", "/test", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path+"?token="+tt.token, nil)
			w := httptest.NewRecorder()

			c := &microweb.Context{
				R: req,
				W: w,
			}

			result := tr.authMiddleware(c)

			if result != tt.expectAuth {
				t.Errorf("authMiddleware() = %v, want %v", result, tt.expectAuth)
			}
		})
	}
}



// ============================================================================
// Integration Tests
// ============================================================================

func TestIntegrationFullWorkflow(t *testing.T) {
	tr, _ := createTestTree(t)
	defer cleanup(t, tr)

	// Create a complete hierarchical structure
	paths := []string{
		"/apps/web/frontend",
		"/apps/web/backend",
		"/apps/mobile/ios",
		"/apps/mobile/android",
	}

	for _, path := range paths {
		if err := tr.CreatePath(path); err != nil {
			t.Fatalf("Failed to create path %q: %v", path, err)
		}
	}

	// Set configuration values
	configs := map[string]map[string]string{
		"/apps/web/frontend": {
			"port":    "3000",
			"host":    "localhost",
			"workers": "4",
		},
		"/apps/web/backend": {
			"port":    "8080",
			"host":    "0.0.0.0",
			"workers": "8",
		},
	}

	for path, values := range configs {
		for k, v := range values {
			fullPath := fmt.Sprintf("%s/%s", path, k)
			if err := tr.SetValue(fullPath, v); err != nil {
				t.Fatalf("Failed to set value at %q: %v", fullPath, err)
			}
		}
	}

	// Verify we can navigate the hierarchy
	appNodes, err := tr.GetNodesInPath("/apps")
	if err != nil {
		t.Fatalf("Failed to get app nodes: %v", err)
	}
	if len(appNodes) != 2 {
		t.Errorf("Expected 2 app categories, got %d", len(appNodes))
	}

	// Verify we can retrieve all values
	frontendValues, err := tr.GetAllPropsWithValues("/apps/web/frontend")
	if err != nil {
		t.Fatalf("Failed to get frontend values: %v", err)
	}
	if len(frontendValues) != 3 {
		t.Errorf("Expected 3 frontend values, got %d", len(frontendValues))
	}

	// Test updating a value
	if err := tr.SetValue("/apps/web/frontend/port", "3001"); err != nil {
		t.Fatalf("Failed to update port: %v", err)
	}

	newPort, err := tr.GetValue("/apps/web/frontend/port")
	if err != nil {
		t.Fatalf("Failed to get updated port: %v", err)
	}
	if newPort != "3001" {
		t.Errorf("Port = %q, want %q", newPort, "3001")
	}

	// Test deleting a property
	if err := tr.DeleteValue("/apps/web/frontend", "workers"); err != nil {
		t.Fatalf("Failed to delete workers: %v", err)
	}

	hasWorkers, err := tr.HasValue("/apps/web/frontend", "workers")
	if err != nil {
		t.Fatalf("Failed to check workers: %v", err)
	}
	if hasWorkers {
		t.Error("Workers property should have been deleted")
	}
}

func TestPersistence(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "persist.db")

	// Create tree and add data
	tr1, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: dbPath,
		Port:                  0,
		Token:                 "test",
	})
	if err != nil {
		t.Fatalf("Failed to create first tree: %v", err)
	}

	tr1.CreatePath("/persistent/data")
	tr1.SetValue("/persistent/data/key", "value123")
	tr1.Close()

	// Give it a moment to ensure file is written
	time.Sleep(100 * time.Millisecond)

	// Reopen tree and verify data persisted
	tr2, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: dbPath,
		Port:                  0,
		Token:                 "test",
	})
	if err != nil {
		t.Fatalf("Failed to reopen tree: %v", err)
	}
	defer tr2.Close()

	value, err := tr2.GetValue("/persistent/data/key")
	if err != nil {
		t.Fatalf("Failed to get persisted value: %v", err)
	}

	if value != "value123" {
		t.Errorf("Persisted value = %q, want %q", value, "value123")
	}
}
