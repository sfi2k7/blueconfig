package blueconfig

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sfi2k7/blueconfig/models"
)

// ============================================================================
// Test Helpers
// ============================================================================

func setupDatabaseTest(t *testing.T) *Tree {
	dbPath := "./test_db_" + t.Name()
	os.RemoveAll(dbPath)

	tree, err := NewOrOpenTree(TreeOptions{
		StorageLocationOnDisk: dbPath,
	})
	if err != nil {
		t.Fatalf("Failed to create tree: %v", err)
	}

	return tree
}

func teardownDatabaseTest(t *testing.T, tree *Tree) {
	if tree != nil {
		dbPath := tree.diskpath
		tree.Close()
		os.RemoveAll(dbPath)
	}
}

// ============================================================================
// Database Operation Tests
// ============================================================================

func TestCreateDatabase(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	err := tree.CreateDatabase("root/mydb", nil)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Verify database type
	isDB, err := tree.IsDatabase("root/mydb")
	if err != nil {
		t.Fatalf("Failed to check database: %v", err)
	}
	if !isDB {
		t.Error("Node should be a database")
	}

	// Verify metadata
	dbType, _ := tree.GetValue("root/mydb/__type")
	if dbType != TypeDatabase {
		t.Errorf("Expected type %s, got %s", TypeDatabase, dbType)
	}

	tableCount, _ := tree.GetValue("root/mydb/__table_count")
	if tableCount != "0" {
		t.Errorf("Expected table count 0, got %s", tableCount)
	}
}

func TestCreateDatabaseWithMetadata(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	metadata := map[string]interface{}{
		"description": "Test database",
		"owner":       "admin",
	}

	err := tree.CreateDatabase("root/mydb", metadata)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Verify custom metadata
	desc, _ := tree.GetValue("root/mydb/description")
	if desc != "Test database" {
		t.Errorf("Expected description 'Test database', got %s", desc)
	}

	owner, _ := tree.GetValue("root/mydb/owner")
	if owner != "admin" {
		t.Errorf("Expected owner 'admin', got %s", owner)
	}
}

func TestListDatabases(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create multiple databases
	tree.CreateDatabase("root/db1", nil)
	tree.CreateDatabase("root/db2", nil)
	tree.CreateDatabase("root/db3", nil)

	// Create non-database node
	tree.CreatePath("root/notdb")
	tree.SetValue("root/notdb/__type", "other")

	databases, err := tree.ListDatabases("root")
	if err != nil {
		t.Fatalf("Failed to list databases: %v", err)
	}

	if len(databases) != 3 {
		t.Errorf("Expected 3 databases, got %d", len(databases))
	}

	// Verify expected databases
	expected := map[string]bool{"db1": true, "db2": true, "db3": true}
	for _, db := range databases {
		if !expected[db] {
			t.Errorf("Unexpected database: %s", db)
		}
	}
}

func TestGetDatabaseInfo(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	err := tree.CreateDatabase("root/mydb", nil)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	info, err := tree.GetDatabaseInfo("root/mydb")
	if err != nil {
		t.Fatalf("Failed to get database info: %v", err)
	}

	if info.Path != "root/mydb" {
		t.Errorf("Expected path 'root/mydb', got %s", info.Path)
	}

	if info.Type != TypeDatabase {
		t.Errorf("Expected type %s, got %s", TypeDatabase, info.Type)
	}

	if info.TableCount != 0 {
		t.Errorf("Expected table count 0, got %d", info.TableCount)
	}

	if info.ViewCount != 0 {
		t.Errorf("Expected view count 0, got %d", info.ViewCount)
	}
}

func TestDeleteDatabaseWithoutForce(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	err := tree.CreateDatabase("root/mydb", nil)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Create a table
	err = tree.CreateTable("root/mydb", "users")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Try to delete without force
	err = tree.DeleteDatabase("root/mydb", false)
	if err == nil {
		t.Error("Expected error when deleting database with tables without force")
	}
}

func TestDeleteDatabaseWithForce(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	err := tree.CreateDatabase("root/mydb", nil)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Create a table
	err = tree.CreateTable("root/mydb", "users")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Delete with force
	err = tree.DeleteDatabase("root/mydb", true)
	if err != nil {
		t.Errorf("Failed to delete database with force: %v", err)
	}

	// Verify deletion
	isDB, _ := tree.IsDatabase("root/mydb")
	if isDB {
		t.Error("Database should be deleted")
	}
}

func TestIsDatabaseOnNonDatabase(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create regular node
	tree.CreatePath("root/regular")

	isDB, err := tree.IsDatabase("root/regular")
	if err != nil {
		t.Fatalf("Failed to check database: %v", err)
	}

	if isDB {
		t.Error("Regular node should not be a database")
	}
}

// ============================================================================
// Table Operation Tests
// ============================================================================

func TestCreateTable(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create database first
	err := tree.CreateDatabase("root/mydb", nil)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Create table
	err = tree.CreateTable("root/mydb", "users")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Verify table type
	isTable, err := tree.IsTable("root/mydb/users")
	if err != nil {
		t.Fatalf("Failed to check table: %v", err)
	}
	if !isTable {
		t.Error("Node should be a table")
	}

	// Verify metadata
	tableType, _ := tree.GetValue("root/mydb/users/__type")
	if tableType != TypeTable {
		t.Errorf("Expected type %s, got %s", TypeTable, tableType)
	}

	tableName, _ := tree.GetValue("root/mydb/users/__name")
	if tableName != "users" {
		t.Errorf("Expected name 'users', got %s", tableName)
	}

	hasSchema, _ := tree.GetValue("root/mydb/users/__has_schema")
	if hasSchema != "false" {
		t.Errorf("Expected has_schema false, got %s", hasSchema)
	}
}

func TestCreateTableInNonDatabase(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create regular node
	tree.CreatePath("root/regular")

	// Try to create table
	err := tree.CreateTable("root/regular", "users")
	if err == nil {
		t.Error("Expected error when creating table in non-database")
	}
}

func TestListTables(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create database
	tree.CreateDatabase("root/mydb", nil)

	// Create multiple tables
	tree.CreateTable("root/mydb", "users")
	tree.CreateTable("root/mydb", "products")
	tree.CreateTable("root/mydb", "orders")

	tables, err := tree.ListTables("root/mydb")
	if err != nil {
		t.Fatalf("Failed to list tables: %v", err)
	}

	if len(tables) != 3 {
		t.Errorf("Expected 3 tables, got %d", len(tables))
	}

	// Verify expected tables
	expected := map[string]bool{"users": true, "products": true, "orders": true}
	for _, table := range tables {
		if !expected[table] {
			t.Errorf("Unexpected table: %s", table)
		}
	}
}

func TestGetTableInfo(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	info, err := tree.GetTableInfo("root/mydb/users")
	if err != nil {
		t.Fatalf("Failed to get table info: %v", err)
	}

	if info.Path != "root/mydb/users" {
		t.Errorf("Expected path 'root/mydb/users', got %s", info.Path)
	}

	if info.Name != "users" {
		t.Errorf("Expected name 'users', got %s", info.Name)
	}

	if info.Type != TypeTable {
		t.Errorf("Expected type %s, got %s", TypeTable, info.Type)
	}

	if info.RowCount != 0 {
		t.Errorf("Expected row count 0, got %d", info.RowCount)
	}

	if info.HasSchema {
		t.Error("Expected HasSchema false")
	}
}

func TestDeleteTable(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	err := tree.DeleteTable("root/mydb", "users", false)
	if err != nil {
		t.Errorf("Failed to delete table: %v", err)
	}

	// Verify deletion
	isTable, _ := tree.IsTable("root/mydb/users")
	if isTable {
		t.Error("Table should be deleted")
	}

	// Verify database table count updated
	tableCount, _ := tree.GetValue("root/mydb/__table_count")
	if tableCount != "0" {
		t.Errorf("Expected table count 0, got %s", tableCount)
	}
}

func TestIsTableOnNonTable(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreatePath("root/regular")

	isTable, err := tree.IsTable("root/regular")
	if err != nil {
		t.Fatalf("Failed to check table: %v", err)
	}

	if isTable {
		t.Error("Regular node should not be a table")
	}
}

// ============================================================================
// Dynamic Schema Tests
// ============================================================================

func TestInferSchemaFromRow(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create row with various types
	rowData := map[string]interface{}{
		"id":     123,
		"name":   "John Doe",
		"age":    30,
		"active": true,
		"score":  95.5,
	}

	row := models.NewRow(rowData)

	schema, err := tree.InferSchemaFromRow(row)
	if err != nil {
		t.Fatalf("Failed to infer schema: %v", err)
	}

	if schema.Version != 1 {
		t.Errorf("Expected version 1, got %d", schema.Version)
	}

	// Check field types
	if schema.Fields["id"] != "int" {
		t.Errorf("Expected id type 'int', got %s", schema.Fields["id"])
	}

	if schema.Fields["name"] != "string" {
		t.Errorf("Expected name type 'string', got %s", schema.Fields["name"])
	}

	if schema.Fields["age"] != "int" {
		t.Errorf("Expected age type 'int', got %s", schema.Fields["age"])
	}

	if schema.Fields["active"] != "bool" {
		t.Errorf("Expected active type 'bool', got %s", schema.Fields["active"])
	}

	if schema.Fields["score"] != "float" {
		t.Errorf("Expected score type 'float', got %s", schema.Fields["score"])
	}

	if len(schema.FlatFields) != 5 {
		t.Errorf("Expected 5 flat fields, got %d", len(schema.FlatFields))
	}
}

func TestInferSchemaFromNestedRow(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create row with nested objects
	rowData := map[string]interface{}{
		"id":   1,
		"name": "John",
		"user": map[string]interface{}{
			"email": "john@example.com",
			"address": map[string]interface{}{
				"city":  "NYC",
				"zip":   "10001",
				"state": "NY",
			},
		},
	}

	// Use NewRow which auto-flattens
	row := models.NewRow(rowData)

	schema, err := tree.InferSchemaFromRow(row)
	if err != nil {
		t.Fatalf("Failed to infer schema: %v", err)
	}

	// Check flattened fields exist
	expectedFields := []string{
		"id", "name", "user.email", "user.address.city", "user.address.zip", "user.address.state",
	}

	for _, field := range expectedFields {
		if _, exists := schema.Fields[field]; !exists {
			t.Errorf("Expected field %s to exist in schema", field)
		}
	}

	if len(schema.FlatFields) != 6 {
		t.Errorf("Expected 6 flat fields, got %d", len(schema.FlatFields))
	}
}

func TestGetTableSchemaNoSchema(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	schema, err := tree.GetTableSchema("root/mydb/users")
	if err != nil {
		t.Fatalf("Failed to get schema: %v", err)
	}

	if schema != nil {
		t.Error("Expected nil schema for new table")
	}
}

func TestSaveAndGetTableSchema(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Create schema
	rowData := map[string]interface{}{
		"id":   1,
		"name": "Test",
		"age":  25,
	}
	row := models.NewRow(rowData)
	schema, _ := tree.InferSchemaFromRow(row)

	// Save schema
	err := tree.saveSchemaToStorage("root/mydb/users", schema)
	if err != nil {
		t.Fatalf("Failed to save schema: %v", err)
	}

	// Verify __has_schema updated
	hasSchema, _ := tree.GetValue("root/mydb/users/__has_schema")
	if hasSchema != "true" {
		t.Error("Expected __has_schema to be true")
	}

	// Retrieve schema
	retrieved, err := tree.GetTableSchema("root/mydb/users")
	if err != nil {
		t.Fatalf("Failed to get schema: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected schema, got nil")
	}

	if retrieved.Version != schema.Version {
		t.Errorf("Expected version %d, got %d", schema.Version, retrieved.Version)
	}

	if len(retrieved.Fields) != len(schema.Fields) {
		t.Errorf("Expected %d fields, got %d", len(schema.Fields), len(retrieved.Fields))
	}

	for key, typ := range schema.Fields {
		if retrieved.Fields[key] != typ {
			t.Errorf("Field %s: expected type %s, got %s", key, typ, retrieved.Fields[key])
		}
	}
}

func TestMergeSchemas(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create initial schema
	existing := &TableSchema{
		Fields: map[string]string{
			"id":   "int",
			"name": "string",
		},
		FlatFields:  []string{"id", "name"},
		Version:     1,
		Created:     "1000000",
		LastUpdated: "1000000",
	}

	// Create new schema with additional field
	new := &TableSchema{
		Fields: map[string]string{
			"id":    "int",
			"name":  "string",
			"email": "string",
		},
		FlatFields: []string{"id", "name", "email"},
		Version:    1,
	}

	merged := tree.MergeSchemas(existing, new)

	if merged.Version != 2 {
		t.Errorf("Expected version 2, got %d", merged.Version)
	}

	if len(merged.Fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(merged.Fields))
	}

	if merged.Fields["email"] != "string" {
		t.Error("Expected email field to be added")
	}

	if merged.Created != existing.Created {
		t.Error("Created timestamp should be preserved")
	}
}

func TestMergeSchemasNoChange(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create schemas with same fields
	existing := &TableSchema{
		Fields:      map[string]string{"id": "int", "name": "string"},
		FlatFields:  []string{"id", "name"},
		Version:     1,
		Created:     "1000000",
		LastUpdated: "1000000",
	}

	new := &TableSchema{
		Fields:     map[string]string{"id": "int", "name": "string"},
		FlatFields: []string{"id", "name"},
		Version:    1,
	}

	merged := tree.MergeSchemas(existing, new)

	// Version should not change
	if merged.Version != 1 {
		t.Errorf("Expected version 1, got %d", merged.Version)
	}
}

func TestValidateRowAgainstSchema(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	schema := &TableSchema{
		Fields: map[string]string{
			"id":   "int",
			"name": "string",
			"age":  "int",
		},
		FlatFields: []string{"id", "name", "age"},
		Version:    1,
	}

	// Valid row
	validRow := models.NewRow(map[string]interface{}{
		"id":   1,
		"name": "John",
		"age":  30,
	})

	err := tree.ValidateRowAgainstSchema(validRow, schema)
	if err != nil {
		t.Errorf("Valid row should pass validation: %v", err)
	}

	// Row with new field (should be allowed - schema evolves)
	rowWithNewField := models.NewRow(map[string]interface{}{
		"id":    2,
		"name":  "Jane",
		"age":   25,
		"email": "jane@example.com",
	})

	err = tree.ValidateRowAgainstSchema(rowWithNewField, schema)
	if err != nil {
		t.Errorf("Row with new field should pass validation: %v", err)
	}
}

func TestValidateRowAgainstSchemaNilSchema(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	row := models.NewRow(map[string]interface{}{
		"id":   1,
		"name": "Test",
	})

	err := tree.ValidateRowAgainstSchema(row, nil)
	if err != nil {
		t.Errorf("Validation with nil schema should pass: %v", err)
	}
}

func TestUpdateSchemaWithNewRow(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// First row - creates schema
	row1 := models.NewRow(map[string]interface{}{
		"id":   1,
		"name": "John",
	})

	schema1, err := tree.UpdateSchemaWithNewRow("root/mydb/users", row1)
	if err != nil {
		t.Fatalf("Failed to update schema: %v", err)
	}

	if schema1.Version != 1 {
		t.Errorf("Expected version 1, got %d", schema1.Version)
	}

	if len(schema1.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(schema1.Fields))
	}

	// Second row with new field
	row2 := models.NewRow(map[string]interface{}{
		"id":    2,
		"name":  "Jane",
		"email": "jane@example.com",
	})

	schema2, err := tree.UpdateSchemaWithNewRow("root/mydb/users", row2)
	if err != nil {
		t.Fatalf("Failed to update schema: %v", err)
	}

	if schema2.Version != 2 {
		t.Errorf("Expected version 2, got %d", schema2.Version)
	}

	if len(schema2.Fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(schema2.Fields))
	}

	if _, exists := schema2.Fields["email"]; !exists {
		t.Error("Expected email field to be added")
	}
}

func TestGetSchemaAsModelsSchema(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Create and save schema
	row := models.NewRow(map[string]interface{}{
		"id":   1,
		"name": "Test",
		"age":  30,
	})
	schema, _ := tree.InferSchemaFromRow(row)
	tree.saveSchemaToStorage("root/mydb/users", schema)

	// Get as models.Schema
	modelSchema, err := tree.GetSchemaAsModelsSchema("root/mydb/users")
	if err != nil {
		t.Fatalf("Failed to get models schema: %v", err)
	}

	if modelSchema == nil {
		t.Fatal("Expected schema, got nil")
	}

	if len(modelSchema) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(modelSchema))
	}

	if modelSchema["id"] != "int" {
		t.Errorf("Expected id type 'int', got %s", modelSchema["id"])
	}
}

// ============================================================================
// Schema Evolution Tests
// ============================================================================

func TestSchemaEvolutionAddField(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert first row
	row1 := models.NewRow(map[string]interface{}{
		"id":   1,
		"name": "John",
	})
	schema1, _ := tree.UpdateSchemaWithNewRow("root/mydb/users", row1)

	// Insert row with new field
	row2 := models.NewRow(map[string]interface{}{
		"id":    2,
		"name":  "Jane",
		"email": "jane@example.com",
		"phone": "555-1234",
	})
	schema2, _ := tree.UpdateSchemaWithNewRow("root/mydb/users", row2)

	if schema2.Version != 2 {
		t.Errorf("Expected version 2, got %d", schema2.Version)
	}

	if len(schema2.Fields) != 4 {
		t.Errorf("Expected 4 fields, got %d", len(schema2.Fields))
	}

	// Original fields should be preserved
	if schema2.Fields["id"] != schema1.Fields["id"] {
		t.Error("Original field type should be preserved")
	}
}

func TestSchemaEvolutionNestedFields(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// First row with nested object
	row1 := models.NewRow(map[string]interface{}{
		"id": 1,
		"user": map[string]interface{}{
			"name": "John",
		},
	})

	schema1, _ := tree.UpdateSchemaWithNewRow("root/mydb/users", row1)

	// Check flattened field
	if _, exists := schema1.Fields["user.name"]; !exists {
		t.Error("Expected flattened field 'user.name'")
	}

	// Second row with deeper nesting
	row2 := models.NewRow(map[string]interface{}{
		"id": 2,
		"user": map[string]interface{}{
			"name": "Jane",
			"address": map[string]interface{}{
				"city": "LA",
			},
		},
	})

	schema2, _ := tree.UpdateSchemaWithNewRow("root/mydb/users", row2)

	if schema2.Version != 2 {
		t.Errorf("Expected version 2, got %d", schema2.Version)
	}

	// Check new nested field
	if _, exists := schema2.Fields["user.address.city"]; !exists {
		t.Error("Expected flattened field 'user.address.city'")
	}
}

// ============================================================================
// Row Count Tests
// ============================================================================

func TestIncrementRowCount(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Increment count
	err := tree.IncrementRowCount("root/mydb/users")
	if err != nil {
		t.Fatalf("Failed to increment row count: %v", err)
	}

	count, err := tree.GetRowCount("root/mydb/users")
	if err != nil {
		t.Fatalf("Failed to get row count: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// Increment again
	tree.IncrementRowCount("root/mydb/users")
	count, _ = tree.GetRowCount("root/mydb/users")

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestDecrementRowCount(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Set initial count
	tree.SetRowCount("root/mydb/users", 5)

	// Decrement count
	err := tree.DecrementRowCount("root/mydb/users")
	if err != nil {
		t.Fatalf("Failed to decrement row count: %v", err)
	}

	count, _ := tree.GetRowCount("root/mydb/users")
	if count != 4 {
		t.Errorf("Expected count 4, got %d", count)
	}
}

func TestDecrementRowCountBelowZero(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Decrement from 0
	tree.DecrementRowCount("root/mydb/users")

	count, _ := tree.GetRowCount("root/mydb/users")
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestGenerateRowID(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	id1 := tree.GenerateRowID()
	time.Sleep(1 * time.Millisecond)
	id2 := tree.GenerateRowID()

	if id1 == "" {
		t.Error("Generated ID should not be empty")
	}

	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}

	if !strings.HasPrefix(id1, "row_") {
		t.Error("Generated ID should have 'row_' prefix")
	}
}

func TestValidateTablePath(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Valid table path
	err := tree.ValidateTablePath("root/mydb/users")
	if err != nil {
		t.Errorf("Valid table path should pass: %v", err)
	}

	// Invalid table path
	err = tree.ValidateTablePath("root/mydb")
	if err == nil {
		t.Error("Expected error for non-table path")
	}
}

func TestValidateDatabasePath(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)

	// Valid database path
	err := tree.ValidateDatabasePath("root/mydb")
	if err != nil {
		t.Errorf("Valid database path should pass: %v", err)
	}

	// Invalid database path
	tree.CreatePath("root/regular")
	err = tree.ValidateDatabasePath("root/regular")
	if err == nil {
		t.Error("Expected error for non-database path")
	}
}

func TestIsSpecialNode(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	if !tree.isSpecialNode("__schema") {
		t.Error("__schema should be special node")
	}

	if !tree.isSpecialNode("__type") {
		t.Error("__type should be special node")
	}

	if tree.isSpecialNode("regular") {
		t.Error("regular should not be special node")
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestDatabaseTableCreationFlow(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create database
	err := tree.CreateDatabase("root/ecommerce", map[string]interface{}{
		"description": "E-commerce database",
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Create multiple tables
	tables := []string{"users", "products", "orders", "categories"}
	for _, tableName := range tables {
		err = tree.CreateTable("root/ecommerce", tableName)
		if err != nil {
			t.Fatalf("Failed to create table %s: %v", tableName, err)
		}
	}

	// Verify all tables exist
	listedTables, err := tree.ListTables("root/ecommerce")
	if err != nil {
		t.Fatalf("Failed to list tables: %v", err)
	}

	if len(listedTables) != 4 {
		t.Errorf("Expected 4 tables, got %d", len(listedTables))
	}

	// Verify database table count
	info, _ := tree.GetDatabaseInfo("root/ecommerce")
	if info.TableCount != 4 {
		t.Errorf("Expected database table count 4, got %d", info.TableCount)
	}
}

func TestMultipleNestedDatabases(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create nested structure
	tree.CreateDatabase("root/company", nil)
	tree.CreateDatabase("root/company/hr", nil)
	tree.CreateDatabase("root/company/sales", nil)

	// Create tables in different databases
	tree.CreateTable("root/company", "departments")
	tree.CreateTable("root/company/hr", "employees")
	tree.CreateTable("root/company/sales", "customers")

	// Verify HR database
	tables, _ := tree.ListTables("root/company/hr")
	if len(tables) != 1 || tables[0] != "employees" {
		t.Error("HR database should have employees table")
	}

	// Verify company database can have both tables and sub-databases
	children, _ := tree.GetNodesInPath("root/company")
	if len(children) < 3 {
		t.Errorf("Expected at least 3 children, got %d", len(children))
	}
}

func TestSchemaVersioning(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Create initial schema
	row1 := models.NewRow(map[string]interface{}{
		"id": 1,
	})
	schema, _ := tree.UpdateSchemaWithNewRow("root/mydb/users", row1)
	initialVersion := schema.Version

	// Add new field
	row2 := models.NewRow(map[string]interface{}{
		"id":   2,
		"name": "Test",
	})
	schema, _ = tree.UpdateSchemaWithNewRow("root/mydb/users", row2)

	if schema.Version != initialVersion+1 {
		t.Errorf("Expected version %d, got %d", initialVersion+1, schema.Version)
	}

	// Same fields - version should not change
	row3 := models.NewRow(map[string]interface{}{
		"id":   3,
		"name": "Test2",
	})
	schema, _ = tree.UpdateSchemaWithNewRow("root/mydb/users", row3)

	if schema.Version != initialVersion+1 {
		t.Errorf("Expected version to remain %d, got %d", initialVersion+1, schema.Version)
	}
}

func TestComplexNestedStructure(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Complex nested structure
	rowData := map[string]interface{}{
		"id": 1,
		"profile": map[string]interface{}{
			"name": "John",
			"contact": map[string]interface{}{
				"email": "john@example.com",
				"phone": map[string]interface{}{
					"mobile": "555-1234",
					"home":   "555-5678",
				},
			},
			"preferences": map[string]interface{}{
				"theme":    "dark",
				"language": "en",
			},
		},
	}

	row := models.NewRow(rowData)
	schema, err := tree.InferSchemaFromRow(row)
	if err != nil {
		t.Fatalf("Failed to infer schema: %v", err)
	}

	// Verify all flattened fields exist
	expectedFields := []string{
		"id",
		"profile.name",
		"profile.contact.email",
		"profile.contact.phone.mobile",
		"profile.contact.phone.home",
		"profile.preferences.theme",
		"profile.preferences.language",
	}

	for _, field := range expectedFields {
		if _, exists := schema.Fields[field]; !exists {
			t.Errorf("Expected field %s to exist in schema", field)
		}
	}
}

func TestTypeCompatibility(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tests := []struct {
		actual     string
		expected   string
		compatible bool
	}{
		{"int", "int", true},
		{"string", "string", true},
		{"int", "string", true}, // int can be stored as string
		{"float", "string", true},
		{"int", "float", true},
		{"float", "int", true},
		{"bool", "string", true},
		{"string", "int", false}, // string cannot be stored as int
	}

	for _, test := range tests {
		result := tree.isTypeCompatible(test.actual, test.expected)
		if result != test.compatible {
			t.Errorf("isTypeCompatible(%s, %s): expected %v, got %v",
				test.actual, test.expected, test.compatible, result)
		}
	}
}

func TestUpdateDatabaseTableCount(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)

	// Initial count
	count, _ := tree.GetValue("root/mydb/__table_count")
	if count != "0" {
		t.Errorf("Expected initial count 0, got %s", count)
	}

	// Create table
	tree.CreateTable("root/mydb", "users")

	// Count should be updated
	count, _ = tree.GetValue("root/mydb/__table_count")
	if count != "1" {
		t.Errorf("Expected count 1, got %s", count)
	}

	// Create another table
	tree.CreateTable("root/mydb", "products")
	count, _ = tree.GetValue("root/mydb/__table_count")
	if count != "2" {
		t.Errorf("Expected count 2, got %s", count)
	}

	// Delete table
	tree.DeleteTable("root/mydb", "users", false)
	count, _ = tree.GetValue("root/mydb/__table_count")
	if count != "1" {
		t.Errorf("Expected count 1 after deletion, got %s", count)
	}
}

// ============================================================================
// Edge Cases and Error Tests
// ============================================================================

func TestCreateDatabaseOnExistingPath(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create database
	tree.CreateDatabase("root/mydb", nil)

	// Try to create again (should not error, just update)
	err := tree.CreateDatabase("root/mydb", nil)
	if err != nil {
		t.Errorf("Creating database on existing path should not error: %v", err)
	}
}

func TestDeleteNonExistentDatabase(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	err := tree.DeleteDatabase("root/nonexistent", false)
	if err == nil {
		t.Error("Expected error when deleting non-existent database")
	}
}

func TestDeleteNonExistentTable(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)

	err := tree.DeleteTable("root/mydb", "nonexistent", false)
	if err == nil {
		t.Error("Expected error when deleting non-existent table")
	}
}

func TestGetInfoOnNonExistentDatabase(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreatePath("root/regular")

	_, err := tree.GetDatabaseInfo("root/regular")
	if err == nil {
		t.Error("Expected error when getting info on non-database")
	}
}

func TestGetInfoOnNonExistentTable(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreatePath("root/mydb/regular")

	_, err := tree.GetTableInfo("root/mydb/regular")
	if err == nil {
		t.Error("Expected error when getting info on non-table")
	}
}

func TestEmptySchemaFields(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Empty row
	emptyRow := models.NewRow(map[string]interface{}{})

	schema, err := tree.InferSchemaFromRow(emptyRow)
	if err != nil {
		t.Fatalf("Failed to infer schema from empty row: %v", err)
	}

	if len(schema.Fields) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(schema.Fields))
	}

	if len(schema.FlatFields) != 0 {
		t.Errorf("Expected 0 flat fields, got %d", len(schema.FlatFields))
	}
}

func TestSchemaHasNewFields(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	existing := &TableSchema{
		Fields: map[string]string{"id": "int", "name": "string"},
	}

	new1 := &TableSchema{
		Fields: map[string]string{"id": "int", "name": "string"},
	}

	new2 := &TableSchema{
		Fields: map[string]string{"id": "int", "name": "string", "email": "string"},
	}

	if tree.schemaHasNewFields(existing, new1) {
		t.Error("Should not detect new fields when schemas match")
	}

	if !tree.schemaHasNewFields(existing, new2) {
		t.Error("Should detect new fields")
	}
}

func TestListTablesEmptyDatabase(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)

	tables, err := tree.ListTables("root/mydb")
	if err != nil {
		t.Fatalf("Failed to list tables: %v", err)
	}

	if len(tables) != 0 {
		t.Errorf("Expected 0 tables, got %d", len(tables))
	}
}

func TestListDatabasesEmptyRoot(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	databases, err := tree.ListDatabases("root")
	if err != nil {
		t.Fatalf("Failed to list databases: %v", err)
	}

	if len(databases) != 0 {
		t.Errorf("Expected 0 databases, got %d", len(databases))
	}
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

func TestConcurrentDatabaseCreation(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	done := make(chan bool)

	for i := 0; i < 5; i++ {
		go func(id int) {
			dbName := fmt.Sprintf("root/db%d", id)
			tree.CreateDatabase(dbName, nil)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify all databases created
	databases, _ := tree.ListDatabases("root")
	if len(databases) != 5 {
		t.Errorf("Expected 5 databases, got %d", len(databases))
	}
}

func TestConcurrentTableCreation(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			tableName := fmt.Sprintf("table%d", id)
			tree.CreateTable("root/mydb", tableName)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all tables created
	tables, _ := tree.ListTables("root/mydb")
	if len(tables) != 10 {
		t.Errorf("Expected 10 tables, got %d", len(tables))
	}
}

func TestConcurrentSchemaUpdates(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	done := make(chan bool)

	// Create initial schema
	initialRow := models.NewRow(map[string]interface{}{"id": 1})
	tree.UpdateSchemaWithNewRow("root/mydb/users", initialRow)

	// Concurrent schema updates with different fields
	for i := 0; i < 5; i++ {
		go func(id int) {
			fieldName := fmt.Sprintf("field%d", id)
			row := models.NewRow(map[string]interface{}{
				"id":      id,
				fieldName: "value",
			})
			tree.UpdateSchemaWithNewRow("root/mydb/users", row)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify schema has all fields
	schema, _ := tree.GetTableSchema("root/mydb/users")
	if schema == nil {
		t.Fatal("Expected schema to exist")
	}

	// Should have id + 5 field fields
	if len(schema.Fields) < 6 {
		t.Errorf("Expected at least 6 fields, got %d", len(schema.Fields))
	}
}
