package blueconfig

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sfi2k7/blueconfig/models"
	"github.com/sfi2k7/blueconfig/parser"
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

// ============================================================================
// Row/Record CRUD Tests
// ============================================================================

func TestInsertRow(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup database and table
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Create a row
	row := models.NewRow(map[string]interface{}{
		"name":  "John Doe",
		"age":   30,
		"email": "john@example.com",
	})

	// Insert the row
	rowID, err := tree.InsertRow("root/mydb/users", row)
	if err != nil {
		t.Fatalf("Failed to insert row: %v", err)
	}

	if rowID == "" {
		t.Error("Expected non-empty row ID")
	}

	// Verify row count
	count, _ := tree.GetRowCount("root/mydb/users")
	if count != 1 {
		t.Errorf("Expected row count 1, got %d", count)
	}

	// Verify schema was created
	schema, _ := tree.GetTableSchema("root/mydb/users")
	if schema == nil {
		t.Error("Expected schema to be created")
	}

	if len(schema.Fields) != 3 {
		t.Errorf("Expected 3 fields in schema, got %d", len(schema.Fields))
	}
}

func TestInsertRowWithID(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	row := models.NewRow(map[string]interface{}{
		"name": "Jane Doe",
		"age":  25,
	})

	// Insert with custom ID
	err := tree.InsertRowWithID("root/mydb/users", "user_123", row)
	if err != nil {
		t.Fatalf("Failed to insert row with ID: %v", err)
	}

	// Verify row exists
	exists, _ := tree.RowExists("root/mydb/users", "user_123")
	if !exists {
		t.Error("Expected row to exist")
	}

	// Try to insert duplicate ID
	err = tree.InsertRowWithID("root/mydb/users", "user_123", row)
	if err == nil {
		t.Error("Expected error when inserting duplicate row ID")
	}
}

func TestGetRow(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert a row
	originalRow := models.NewRow(map[string]interface{}{
		"name":  "Alice",
		"age":   28,
		"email": "alice@example.com",
	})

	rowID, _ := tree.InsertRow("root/mydb/users", originalRow)

	// Get the row
	retrievedRow, err := tree.GetRow("root/mydb/users", rowID)
	if err != nil {
		t.Fatalf("Failed to get row: %v", err)
	}

	// Verify fields
	if nameVal, ok := retrievedRow["name"]; ok {
		nameStr := nameVal.AsString()
		if nameStr != "Alice" {
			t.Errorf("Expected name 'Alice', got %v", nameStr)
		}
	} else {
		t.Error("Expected 'name' field to exist")
	}

	if ageVal, ok := retrievedRow["age"]; ok {
		ageStr := ageVal.AsString()
		if ageStr != "28" {
			t.Errorf("Expected age '28', got %v", ageStr)
		}
	} else {
		t.Error("Expected 'age' field to exist")
	}
}

func TestGetRowNotFound(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Try to get non-existent row
	_, err := tree.GetRow("root/mydb/users", "nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent row")
	}
}

func TestUpdateRow(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert initial row
	originalRow := models.NewRow(map[string]interface{}{
		"name": "Bob",
		"age":  30,
	})

	rowID, _ := tree.InsertRow("root/mydb/users", originalRow)

	// Update the row
	updatedRow := models.NewRow(map[string]interface{}{
		"name":  "Bob Smith",
		"age":   31,
		"email": "bob@example.com",
	})

	err := tree.UpdateRow("root/mydb/users", rowID, updatedRow)
	if err != nil {
		t.Fatalf("Failed to update row: %v", err)
	}

	// Verify update
	retrievedRow, _ := tree.GetRow("root/mydb/users", rowID)

	if nameVal, ok := retrievedRow["name"]; ok {
		if nameVal.AsString() != "Bob Smith" {
			t.Errorf("Expected updated name, got %v", nameVal.AsString())
		}
	}

	if emailVal, ok := retrievedRow["email"]; ok {
		if emailVal.AsString() != "bob@example.com" {
			t.Errorf("Expected email field, got %v", emailVal.AsString())
		}
	}
}

func TestUpdateRowFields(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert initial row
	originalRow := models.NewRow(map[string]interface{}{
		"name": "Charlie",
		"age":  25,
		"city": "NYC",
	})

	rowID, _ := tree.InsertRow("root/mydb/users", originalRow)

	// Partial update - only age
	err := tree.UpdateRowFields("root/mydb/users", rowID, map[string]interface{}{
		"age": 26,
	})
	if err != nil {
		t.Fatalf("Failed to update row fields: %v", err)
	}

	// Verify partial update
	retrievedRow, _ := tree.GetRow("root/mydb/users", rowID)

	if ageVal, ok := retrievedRow["age"]; ok {
		if ageVal.AsString() != "26" {
			t.Errorf("Expected age 26, got %v", ageVal.AsString())
		}
	}

	// Name and city should remain unchanged
	if nameVal, ok := retrievedRow["name"]; ok {
		if nameVal.AsString() != "Charlie" {
			t.Errorf("Expected name unchanged, got %v", nameVal.AsString())
		}
	}
}

func TestDeleteRow(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert row
	row := models.NewRow(map[string]interface{}{
		"name": "Dave",
	})

	rowID, _ := tree.InsertRow("root/mydb/users", row)

	// Verify row count before delete
	countBefore, _ := tree.GetRowCount("root/mydb/users")
	if countBefore != 1 {
		t.Errorf("Expected count 1 before delete, got %d", countBefore)
	}

	// Delete row
	err := tree.DeleteRow("root/mydb/users", rowID)
	if err != nil {
		t.Fatalf("Failed to delete row: %v", err)
	}

	// Verify row count after delete
	countAfter, _ := tree.GetRowCount("root/mydb/users")
	if countAfter != 0 {
		t.Errorf("Expected count 0 after delete, got %d", countAfter)
	}

	// Verify row doesn't exist
	exists, _ := tree.RowExists("root/mydb/users", rowID)
	if exists {
		t.Error("Expected row to not exist after deletion")
	}
}

func TestListRows(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert multiple rows
	for i := 0; i < 5; i++ {
		row := models.NewRow(map[string]interface{}{
			"id":   i,
			"name": fmt.Sprintf("User %d", i),
		})
		tree.InsertRow("root/mydb/users", row)
	}

	// List all rows
	rowIDs, err := tree.ListRows("root/mydb/users")
	if err != nil {
		t.Fatalf("Failed to list rows: %v", err)
	}

	if len(rowIDs) != 5 {
		t.Errorf("Expected 5 rows, got %d", len(rowIDs))
	}
}

func TestGetAllRows(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "products")

	// Insert multiple rows
	expectedNames := []string{"Product A", "Product B", "Product C"}
	for _, name := range expectedNames {
		row := models.NewRow(map[string]interface{}{
			"name":  name,
			"price": 99.99,
		})
		tree.InsertRow("root/mydb/products", row)
	}

	// Get all rows
	allRows, err := tree.GetAllRows("root/mydb/products")
	if err != nil {
		t.Fatalf("Failed to get all rows: %v", err)
	}

	if len(allRows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(allRows))
	}

	// Verify each row has expected fields
	for _, row := range allRows {
		if _, ok := row["name"]; !ok {
			t.Error("Expected 'name' field in row")
		}
		if _, ok := row["price"]; !ok {
			t.Error("Expected 'price' field in row")
		}
	}
}

func TestScanRows(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "items")

	// Insert rows
	for i := 0; i < 3; i++ {
		row := models.NewRow(map[string]interface{}{
			"id":   i,
			"name": fmt.Sprintf("Item %d", i),
		})
		tree.InsertRow("root/mydb/items", row)
	}

	// Scan rows
	count := 0
	err := tree.ScanRows("root/mydb/items", func(rowID string, row models.Row) error {
		count++

		// Verify row has expected fields
		if _, ok := row["id"]; !ok {
			t.Error("Expected 'id' field in row")
		}
		if _, ok := row["name"]; !ok {
			t.Error("Expected 'name' field in row")
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to scan rows: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected to scan 3 rows, scanned %d", count)
	}
}

func TestCountRows(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "logs")

	// Insert rows
	for i := 0; i < 10; i++ {
		row := models.NewRow(map[string]interface{}{
			"message": fmt.Sprintf("Log entry %d", i),
		})
		tree.InsertRow("root/mydb/logs", row)
	}

	// Count rows
	count, err := tree.CountRows("root/mydb/logs")
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}

	if count != 10 {
		t.Errorf("Expected count 10, got %d", count)
	}
}

func TestRowExists(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert row with known ID
	row := models.NewRow(map[string]interface{}{
		"name": "Test User",
	})
	tree.InsertRowWithID("root/mydb/users", "known_id", row)

	// Check existence
	exists, _ := tree.RowExists("root/mydb/users", "known_id")
	if !exists {
		t.Error("Expected row to exist")
	}

	notExists, _ := tree.RowExists("root/mydb/users", "unknown_id")
	if notExists {
		t.Error("Expected row to not exist")
	}
}

func TestInsertRowInvalidTable(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	// Don't create table

	row := models.NewRow(map[string]interface{}{
		"name": "Test",
	})

	// Try to insert in non-existent table
	_, err := tree.InsertRow("root/mydb/nonexistent", row)
	if err == nil {
		t.Error("Expected error when inserting into non-existent table")
	}
}

func TestUpdateRowNotFound(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	row := models.NewRow(map[string]interface{}{
		"name": "Test",
	})

	// Try to update non-existent row
	err := tree.UpdateRow("root/mydb/users", "nonexistent", row)
	if err == nil {
		t.Error("Expected error when updating non-existent row")
	}
}

func TestDeleteRowNotFound(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Try to delete non-existent row
	err := tree.DeleteRow("root/mydb/users", "nonexistent")
	if err == nil {
		t.Error("Expected error when deleting non-existent row")
	}
}

func TestRowCRUDWithSchemaEvolution(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert first row with basic fields
	row1 := models.NewRow(map[string]interface{}{
		"name": "User 1",
		"age":  25,
	})
	tree.InsertRow("root/mydb/users", row1)

	// Verify schema version 1
	schema1, _ := tree.GetTableSchema("root/mydb/users")
	if schema1.Version != 1 {
		t.Errorf("Expected schema version 1, got %d", schema1.Version)
	}

	// Insert second row with additional field
	row2 := models.NewRow(map[string]interface{}{
		"name":  "User 2",
		"age":   30,
		"email": "user2@example.com",
	})
	tree.InsertRow("root/mydb/users", row2)

	// Verify schema evolved to version 2
	schema2, _ := tree.GetTableSchema("root/mydb/users")
	if schema2.Version != 2 {
		t.Errorf("Expected schema version 2, got %d", schema2.Version)
	}

	// Verify schema has 3 fields
	if len(schema2.Fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(schema2.Fields))
	}
}

func TestConcurrentRowInserts(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	done := make(chan bool)
	numGoroutines := 20

	// Concurrent inserts
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			row := models.NewRow(map[string]interface{}{
				"id":   id,
				"name": fmt.Sprintf("User %d", id),
			})
			tree.InsertRow("root/mydb/users", row)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all rows inserted
	count, _ := tree.CountRows("root/mydb/users")
	if count != numGoroutines {
		t.Errorf("Expected %d rows, got %d", numGoroutines, count)
	}
}

func TestBulkOperations(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "bulk_test")

	// Bulk insert
	numRows := 100
	rowIDs := make([]string, numRows)

	for i := 0; i < numRows; i++ {
		row := models.NewRow(map[string]interface{}{
			"index": i,
			"data":  fmt.Sprintf("Data %d", i),
		})
		rowID, _ := tree.InsertRow("root/mydb/bulk_test", row)
		rowIDs[i] = rowID
	}

	// Verify count
	count, _ := tree.CountRows("root/mydb/bulk_test")
	if count != numRows {
		t.Errorf("Expected %d rows, got %d", numRows, count)
	}

	// Bulk delete (delete every other row)
	for i := 0; i < numRows; i += 2 {
		tree.DeleteRow("root/mydb/bulk_test", rowIDs[i])
	}

	// Verify count after deletion
	countAfter, _ := tree.CountRows("root/mydb/bulk_test")
	expected := numRows / 2
	if countAfter != expected {
		t.Errorf("Expected %d rows after deletion, got %d", expected, countAfter)
	}
}

// ============================================================================
// Index Tests
// ============================================================================

func TestCreateSingleFieldIndex(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create database and table
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"name":  models.NewValue("Alice"),
		"email": models.NewValue("alice@example.com"),
		"age":   models.NewValue(30),
	})

	// Create single field index
	err := tree.CreateIndex("root/mydb/users", "idx_email", []string{"email"}, false)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Verify index exists
	indexes, _ := tree.ListIndexes("root/mydb/users")
	if len(indexes) != 1 || indexes[0] != "idx_email" {
		t.Errorf("Expected index 'idx_email', got: %v", indexes)
	}

	// Verify index info
	info, err := tree.GetIndexInfo("root/mydb/users", "idx_email")
	if err != nil {
		t.Fatalf("Failed to get index info: %v", err)
	}
	if info.Type != IndexTypeSingle {
		t.Errorf("Expected index type %s, got %s", IndexTypeSingle, info.Type)
	}
	if len(info.Fields) != 1 || info.Fields[0] != "email" {
		t.Errorf("Expected fields [email], got %v", info.Fields)
	}
	if info.EntryCount != 1 {
		t.Errorf("Expected entry count 1, got %d", info.EntryCount)
	}
}

func TestCreateCompositeIndex(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create database and table
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"first_name": models.NewValue("Alice"),
		"last_name":  models.NewValue("Smith"),
		"age":        models.NewValue(30),
	})

	// Create composite index
	err := tree.CreateIndex("root/mydb/users", "idx_name", []string{"first_name", "last_name"}, false)
	if err != nil {
		t.Fatalf("Failed to create composite index: %v", err)
	}

	// Verify index info
	info, err := tree.GetIndexInfo("root/mydb/users", "idx_name")
	if err != nil {
		t.Fatalf("Failed to get index info: %v", err)
	}
	if info.Type != IndexTypeComposite {
		t.Errorf("Expected index type %s, got %s", IndexTypeComposite, info.Type)
	}
	if len(info.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(info.Fields))
	}
}

func TestCreateUniqueIndex(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Create database and table
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"email": models.NewValue("alice@example.com"),
		"name":  models.NewValue("Alice"),
	})

	// Create unique index
	err := tree.CreateIndex("root/mydb/users", "idx_unique_email", []string{"email"}, true)
	if err != nil {
		t.Fatalf("Failed to create unique index: %v", err)
	}

	// Verify index is unique
	info, _ := tree.GetIndexInfo("root/mydb/users", "idx_unique_email")
	if !info.Unique {
		t.Error("Expected index to be unique")
	}

	// Try to insert duplicate - should fail
	_, err = tree.InsertRow("root/mydb/users", models.Row{
		"email": models.NewValue("alice@example.com"),
		"name":  models.NewValue("Alice2"),
	})
	if err == nil {
		t.Error("Expected error on duplicate unique key, got nil")
	}
}

func TestDropIndex(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"email": models.NewValue("alice@example.com"),
	})
	tree.CreateIndex("root/mydb/users", "idx_email", []string{"email"}, false)

	// Drop index
	err := tree.DropIndex("root/mydb/users", "idx_email")
	if err != nil {
		t.Fatalf("Failed to drop index: %v", err)
	}

	// Verify index is gone
	indexes, _ := tree.ListIndexes("root/mydb/users")
	if len(indexes) != 0 {
		t.Errorf("Expected no indexes, got: %v", indexes)
	}

	// Try to drop non-existent index
	err = tree.DropIndex("root/mydb/users", "idx_nonexistent")
	if err == nil {
		t.Error("Expected error when dropping non-existent index")
	}
}

func TestListIndexes(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"email": models.NewValue("alice@example.com"),
		"name":  models.NewValue("Alice"),
		"age":   models.NewValue(30),
	})

	// Initially no indexes
	indexes, _ := tree.ListIndexes("root/mydb/users")
	if len(indexes) != 0 {
		t.Errorf("Expected 0 indexes, got %d", len(indexes))
	}

	// Create multiple indexes
	tree.CreateIndex("root/mydb/users", "idx_email", []string{"email"}, false)
	tree.CreateIndex("root/mydb/users", "idx_name", []string{"name"}, false)
	tree.CreateIndex("root/mydb/users", "idx_age", []string{"age"}, false)

	// List all indexes
	indexes, _ = tree.ListIndexes("root/mydb/users")
	if len(indexes) != 3 {
		t.Errorf("Expected 3 indexes, got %d", len(indexes))
	}
}

func TestRebuildIndex(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert initial data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"email": models.NewValue("alice@example.com"),
	})

	// Create index
	tree.CreateIndex("root/mydb/users", "idx_email", []string{"email"}, false)

	// Insert more data
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"email": models.NewValue("bob@example.com"),
	})
	tree.InsertRowWithID("root/mydb/users", "user3", models.Row{
		"email": models.NewValue("charlie@example.com"),
	})

	// Rebuild index
	err := tree.RebuildIndex("root/mydb/users", "idx_email")
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v", err)
	}

	// Verify entry count
	info, _ := tree.GetIndexInfo("root/mydb/users", "idx_email")
	if info.EntryCount != 3 {
		t.Errorf("Expected entry count 3, got %d", info.EntryCount)
	}
}

func TestIndexMaintenanceOnInsert(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Create index on empty table
	tree.CreateIndex("root/mydb/users", "idx_email", []string{"email"}, false)

	// Insert rows - index should be automatically maintained
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"email": models.NewValue("alice@example.com"),
		"name":  models.NewValue("Alice"),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"email": models.NewValue("bob@example.com"),
		"name":  models.NewValue("Bob"),
	})

	// Verify index was updated
	info, _ := tree.GetIndexInfo("root/mydb/users", "idx_email")
	if info.EntryCount != 2 {
		t.Errorf("Expected entry count 2, got %d", info.EntryCount)
	}

	// Test index lookup
	rowIDs, _ := tree.lookupIndex("root/mydb/users", "idx_email", "alice@example.com")
	if len(rowIDs) != 1 || rowIDs[0] != "user1" {
		t.Errorf("Expected [user1], got %v", rowIDs)
	}
}

func TestIndexMaintenanceOnUpdate(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert and create index
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"email": models.NewValue("alice@example.com"),
		"name":  models.NewValue("Alice"),
	})
	tree.CreateIndex("root/mydb/users", "idx_email", []string{"email"}, false)

	// Old email should be in index
	rowIDs, _ := tree.lookupIndex("root/mydb/users", "idx_email", "alice@example.com")
	if len(rowIDs) != 1 {
		t.Error("Old email not found in index before update")
	}

	// Update row - index should be automatically updated
	tree.UpdateRow("root/mydb/users", "user1", models.Row{
		"email": models.NewValue("alice.new@example.com"),
		"name":  models.NewValue("Alice"),
	})

	// Old email should be removed from index
	rowIDs, _ = tree.lookupIndex("root/mydb/users", "idx_email", "alice@example.com")
	if len(rowIDs) != 0 {
		t.Error("Old email still in index after update")
	}

	// New email should be in index
	rowIDs, _ = tree.lookupIndex("root/mydb/users", "idx_email", "alice.new@example.com")
	if len(rowIDs) != 1 || rowIDs[0] != "user1" {
		t.Errorf("New email not found in index, got: %v", rowIDs)
	}
}

func TestIndexMaintenanceOnDelete(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert and create index
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"email": models.NewValue("alice@example.com"),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"email": models.NewValue("bob@example.com"),
	})
	tree.CreateIndex("root/mydb/users", "idx_email", []string{"email"}, false)

	// Verify both entries in index
	info, _ := tree.GetIndexInfo("root/mydb/users", "idx_email")
	if info.EntryCount != 2 {
		t.Errorf("Expected entry count 2, got %d", info.EntryCount)
	}

	// Delete row - index should be automatically cleaned
	tree.DeleteRow("root/mydb/users", "user1")

	// Verify index entry count decreased
	info, _ = tree.GetIndexInfo("root/mydb/users", "idx_email")
	if info.EntryCount != 1 {
		t.Errorf("Expected entry count 1 after delete, got %d", info.EntryCount)
	}

	// Verify deleted email not in index
	rowIDs, _ := tree.lookupIndex("root/mydb/users", "idx_email", "alice@example.com")
	if len(rowIDs) != 0 {
		t.Error("Deleted row still in index")
	}

	// Verify other email still in index
	rowIDs, _ = tree.lookupIndex("root/mydb/users", "idx_email", "bob@example.com")
	if len(rowIDs) != 1 {
		t.Error("Other row not found in index")
	}
}

func TestUniqueIndexViolation(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert initial data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"email": models.NewValue("alice@example.com"),
	})

	// Create unique index
	tree.CreateIndex("root/mydb/users", "idx_email", []string{"email"}, true)

	// Try to insert duplicate - should fail
	_, err := tree.InsertRow("root/mydb/users", models.Row{
		"email": models.NewValue("alice@example.com"),
	})
	if err == nil {
		t.Error("Expected error on duplicate unique key during insert")
	}

	// Try to update to duplicate - should fail
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"email": models.NewValue("bob@example.com"),
	})
	err = tree.UpdateRow("root/mydb/users", "user2", models.Row{
		"email": models.NewValue("alice@example.com"),
	})
	if err == nil {
		t.Error("Expected error on duplicate unique key during update")
	}
}

func TestIndexWithNullValues(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert rows with null values
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"email": models.NewValue("alice@example.com"),
		"phone": models.NewValue(nil), // null value
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"email": models.NewValue("bob@example.com"),
		// phone field missing
	})
	tree.InsertRowWithID("root/mydb/users", "user3", models.Row{
		"email": models.NewValue("charlie@example.com"),
		"phone": models.NewValue("555-1234"),
	})

	// Create index on phone field
	tree.CreateIndex("root/mydb/users", "idx_phone", []string{"phone"}, false)

	// Index should only contain non-null values
	info, _ := tree.GetIndexInfo("root/mydb/users", "idx_phone")
	if info.EntryCount != 1 {
		t.Errorf("Expected entry count 1 (only non-null), got %d", info.EntryCount)
	}

	// Lookup should find the non-null value
	rowIDs, _ := tree.lookupIndex("root/mydb/users", "idx_phone", "555-1234")
	if len(rowIDs) != 1 || rowIDs[0] != "user3" {
		t.Errorf("Expected [user3], got %v", rowIDs)
	}
}

func TestLookupIndex(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"city": models.NewValue("New York"),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"city": models.NewValue("New York"),
	})
	tree.InsertRowWithID("root/mydb/users", "user3", models.Row{
		"city": models.NewValue("Boston"),
	})

	// Create index
	tree.CreateIndex("root/mydb/users", "idx_city", []string{"city"}, false)

	// Lookup New York - should return 2 users
	rowIDs, err := tree.lookupIndex("root/mydb/users", "idx_city", "New York")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if len(rowIDs) != 2 {
		t.Errorf("Expected 2 users in New York, got %d", len(rowIDs))
	}

	// Lookup Boston - should return 1 user
	rowIDs, _ = tree.lookupIndex("root/mydb/users", "idx_city", "Boston")
	if len(rowIDs) != 1 {
		t.Errorf("Expected 1 user in Boston, got %d", len(rowIDs))
	}

	// Lookup non-existent - should return empty
	rowIDs, _ = tree.lookupIndex("root/mydb/users", "idx_city", "Chicago")
	if len(rowIDs) != 0 {
		t.Errorf("Expected 0 users in Chicago, got %d", len(rowIDs))
	}
}

func TestLookupIndexRange(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "products")

	// Insert test data with age values
	tree.InsertRowWithID("root/mydb/products", "prod1", models.Row{
		"price": models.NewValue(10),
	})
	tree.InsertRowWithID("root/mydb/products", "prod2", models.Row{
		"price": models.NewValue(25),
	})
	tree.InsertRowWithID("root/mydb/products", "prod3", models.Row{
		"price": models.NewValue(50),
	})
	tree.InsertRowWithID("root/mydb/products", "prod4", models.Row{
		"price": models.NewValue(75),
	})

	// Create index
	tree.CreateIndex("root/mydb/products", "idx_price", []string{"price"}, false)

	// Range lookup: 20-60
	rowIDs, err := tree.lookupIndexRange("root/mydb/products", "idx_price", "25", "50")
	if err != nil {
		t.Fatalf("Range lookup failed: %v", err)
	}
	// Should include prod2 and prod3
	if len(rowIDs) < 1 {
		t.Errorf("Expected at least 1 result in range, got %d", len(rowIDs))
	}
}

func TestCompositeIndexLookup(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "orders")

	// Insert test data
	tree.InsertRowWithID("root/mydb/orders", "order1", models.Row{
		"customer_id": models.NewValue("cust1"),
		"status":      models.NewValue("pending"),
	})
	tree.InsertRowWithID("root/mydb/orders", "order2", models.Row{
		"customer_id": models.NewValue("cust1"),
		"status":      models.NewValue("completed"),
	})
	tree.InsertRowWithID("root/mydb/orders", "order3", models.Row{
		"customer_id": models.NewValue("cust2"),
		"status":      models.NewValue("pending"),
	})

	// Create composite index
	tree.CreateIndex("root/mydb/orders", "idx_customer_status", []string{"customer_id", "status"}, false)

	// Lookup with composite key: cust1|pending
	rowIDs, err := tree.lookupIndex("root/mydb/orders", "idx_customer_status", "cust1|pending")
	if err != nil {
		t.Fatalf("Composite lookup failed: %v", err)
	}
	if len(rowIDs) != 1 || rowIDs[0] != "order1" {
		t.Errorf("Expected [order1], got %v", rowIDs)
	}
}

func TestCreateIndexOnInvalidTable(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Try to create index on non-existent table
	err := tree.CreateIndex("root/mydb/nonexistent", "idx_test", []string{"field"}, false)
	if err == nil {
		t.Error("Expected error when creating index on non-existent table")
	}
}

func TestCreateDuplicateIndex(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"email": models.NewValue("alice@example.com"),
	})

	// Create index
	tree.CreateIndex("root/mydb/users", "idx_email", []string{"email"}, false)

	// Try to create same index again - should fail
	err := tree.CreateIndex("root/mydb/users", "idx_email", []string{"email"}, false)
	if err == nil {
		t.Error("Expected error when creating duplicate index")
	}
}

// ============================================================================
// Cursor and Selective Field Loading Tests
// ============================================================================

func TestGetRowFields(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert a row with multiple fields
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"name":  models.NewValue("Alice"),
		"email": models.NewValue("alice@example.com"),
		"age":   models.NewValue(30),
		"city":  models.NewValue("NYC"),
	})

	// Load only specific fields
	row, err := tree.GetRowFields("root/mydb/users", "user1", []string{"name", "age"})
	if err != nil {
		t.Fatalf("GetRowFields failed: %v", err)
	}

	// Should have exactly 2 fields
	if len(row) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(row))
	}

	// Should have name and age
	if _, exists := row["name"]; !exists {
		t.Error("Expected 'name' field")
	}
	if _, exists := row["age"]; !exists {
		t.Error("Expected 'age' field")
	}

	// Should NOT have email or city
	if _, exists := row["email"]; exists {
		t.Error("Should not have 'email' field")
	}
	if _, exists := row["city"]; exists {
		t.Error("Should not have 'city' field")
	}

	// Verify values
	if row["name"].AsString() != "Alice" {
		t.Errorf("Expected name 'Alice', got '%s'", row["name"].AsString())
	}
}

func TestScanRowsWithFields(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"name":  models.NewValue("Alice"),
		"email": models.NewValue("alice@example.com"),
		"age":   models.NewValue(30),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"name":  models.NewValue("Bob"),
		"email": models.NewValue("bob@example.com"),
		"age":   models.NewValue(25),
	})

	// Scan with only specific fields
	rowCount := 0
	err := tree.ScanRowsWithFields("root/mydb/users", []string{"name"}, func(rowID string, row models.Row) error {
		rowCount++

		// Should have only 'name' field
		if len(row) != 1 {
			t.Errorf("Expected 1 field, got %d", len(row))
		}

		if _, exists := row["name"]; !exists {
			t.Error("Expected 'name' field")
		}

		return nil
	})

	if err != nil {
		t.Fatalf("ScanRowsWithFields failed: %v", err)
	}

	if rowCount != 2 {
		t.Errorf("Expected 2 rows scanned, got %d", rowCount)
	}
}

func TestQueryCursorBasic(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	rowIDs := []string{"user1", "user2", "user3"}
	for _, rowID := range rowIDs {
		tree.InsertRowWithID("root/mydb/users", rowID, models.Row{
			"name": models.NewValue("User_" + rowID),
		})
	}

	// Create cursor
	cursor := tree.NewQueryCursor("root/mydb/users", rowIDs)

	// Test Count
	if cursor.Count() != 3 {
		t.Errorf("Expected count 3, got %d", cursor.Count())
	}

	// Iterate through cursor
	visitedIDs := []string{}
	for i := 0; i < cursor.Count(); i++ {
		id := cursor.CurrentID()
		visitedIDs = append(visitedIDs, id)

		if i < cursor.Count()-1 && !cursor.Next() {
			t.Error("Next() should return true")
		}
	}

	if len(visitedIDs) != 3 {
		t.Errorf("Expected to visit 3 rows, visited %d", len(visitedIDs))
	}
}

func TestQueryCursorWithLimit(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert 10 rows
	rowIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		rowID := fmt.Sprintf("user%d", i)
		rowIDs[i] = rowID
		tree.InsertRowWithID("root/mydb/users", rowID, models.Row{
			"name": models.NewValue(fmt.Sprintf("User %d", i)),
		})
	}

	// Create cursor with limit of 5
	cursor := tree.NewQueryCursor("root/mydb/users", rowIDs).WithLimit(5)

	// Iterate
	count := 0
	for i := 0; i < cursor.Count(); i++ {
		count++
		if !cursor.Next() {
			break
		}
	}

	// Should only iterate 5 times due to limit
	if count != 5 {
		t.Errorf("Expected to iterate 5 times with limit, got %d", count)
	}
}

func TestQueryCursorWithSkip(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert 10 rows
	rowIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		rowID := fmt.Sprintf("user%d", i)
		rowIDs[i] = rowID
		tree.InsertRowWithID("root/mydb/users", rowID, models.Row{
			"idx": models.NewValue(i),
		})
	}

	// Create cursor with skip of 3
	cursor := tree.NewQueryCursor("root/mydb/users", rowIDs).WithSkip(3)

	// First row should be user3 (0-indexed, so skip 0, 1, 2)
	firstID := cursor.CurrentID()
	if firstID != "user3" {
		t.Errorf("Expected first ID 'user3' after skip, got '%s'", firstID)
	}
}

func TestQueryCursorWithPagination(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert 20 rows
	rowIDs := make([]string, 20)
	for i := 0; i < 20; i++ {
		rowID := fmt.Sprintf("user%02d", i)
		rowIDs[i] = rowID
		tree.InsertRowWithID("root/mydb/users", rowID, models.Row{
			"idx": models.NewValue(i),
		})
	}

	// Get page 2 with page size 5 (should be users 5-9)
	cursor := tree.NewQueryCursor("root/mydb/users", rowIDs).WithPagination(2, 5)

	// Should start at user05
	firstID := cursor.CurrentID()
	if firstID != "user05" {
		t.Errorf("Expected page 2 to start at 'user05', got '%s'", firstID)
	}

	// Count iterations
	count := 1 // Already at first position
	for cursor.Next() {
		count++
	}

	if count != 5 {
		t.Errorf("Expected 5 rows in page, got %d", count)
	}
}

func TestSortRowsByFields(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert unsorted data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"name": models.NewValue("Charlie"),
		"age":  models.NewValue(30),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"name": models.NewValue("Alice"),
		"age":  models.NewValue(25),
	})
	tree.InsertRowWithID("root/mydb/users", "user3", models.Row{
		"name": models.NewValue("Bob"),
		"age":  models.NewValue(35),
	})

	rowIDs := []string{"user1", "user2", "user3"}

	// Sort by name ascending
	sortedIDs, err := tree.SortRowsByFields("root/mydb/users", rowIDs, []SortField{
		{FieldName: "name", Direction: SortAsc},
	})

	if err != nil {
		t.Fatalf("SortRowsByFields failed: %v", err)
	}

	// Should be: Alice (user2), Bob (user3), Charlie (user1)
	expected := []string{"user2", "user3", "user1"}
	for i, expectedID := range expected {
		if sortedIDs[i] != expectedID {
			t.Errorf("Position %d: expected %s, got %s", i, expectedID, sortedIDs[i])
		}
	}
}

func TestSortRowsByFieldsDescending(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"age": models.NewValue(30),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"age": models.NewValue(25),
	})
	tree.InsertRowWithID("root/mydb/users", "user3", models.Row{
		"age": models.NewValue(35),
	})

	rowIDs := []string{"user1", "user2", "user3"}

	// Sort by age descending
	sortedIDs, err := tree.SortRowsByFields("root/mydb/users", rowIDs, []SortField{
		{FieldName: "age", Direction: SortDesc},
	})

	if err != nil {
		t.Fatalf("SortRowsByFields failed: %v", err)
	}

	// Should be: 35 (user3), 30 (user1), 25 (user2)
	expected := []string{"user3", "user1", "user2"}
	for i, expectedID := range expected {
		if sortedIDs[i] != expectedID {
			t.Errorf("Position %d: expected %s, got %s", i, expectedID, sortedIDs[i])
		}
	}
}

func TestGetRowIDsOnly(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert data
	for i := 0; i < 5; i++ {
		tree.InsertRowWithID("root/mydb/users", fmt.Sprintf("user%d", i), models.Row{
			"name": models.NewValue(fmt.Sprintf("User %d", i)),
		})
	}

	// Get row IDs only (memory efficient)
	rowIDs, err := tree.GetRowIDsOnly("root/mydb/users")
	if err != nil {
		t.Fatalf("GetRowIDsOnly failed: %v", err)
	}

	if len(rowIDs) != 5 {
		t.Errorf("Expected 5 row IDs, got %d", len(rowIDs))
	}
}

func TestCreateRowCursorsWithFields(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"name":  models.NewValue("Alice"),
		"email": models.NewValue("alice@example.com"),
		"age":   models.NewValue(30),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"name":  models.NewValue("Bob"),
		"email": models.NewValue("bob@example.com"),
		"age":   models.NewValue(25),
	})

	// Create cursors with only name field
	cursors, err := tree.CreateRowCursorsWithFields("root/mydb/users", []string{"user1", "user2"}, []string{"name"})
	if err != nil {
		t.Fatalf("CreateRowCursorsWithFields failed: %v", err)
	}

	if len(cursors) != 2 {
		t.Errorf("Expected 2 cursors, got %d", len(cursors))
	}

	// Each cursor should have only name field
	for _, cursor := range cursors {
		if len(cursor.Fields) != 1 {
			t.Errorf("Expected 1 field in cursor, got %d", len(cursor.Fields))
		}

		if _, exists := cursor.Fields["name"]; !exists {
			t.Error("Expected 'name' field in cursor")
		}
	}
}

// ============================================================================
// Query Planner and Execution Tests
// ============================================================================

func TestFindRowsBasic(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"name": models.NewValue("Alice"),
		"age":  models.NewValue(30),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"name": models.NewValue("Bob"),
		"age":  models.NewValue(25),
	})
	tree.InsertRowWithID("root/mydb/users", "user3", models.Row{
		"name": models.NewValue("Charlie"),
		"age":  models.NewValue(35),
	})

	// Find rows with age > 26
	rows, err := tree.FindRows("root/mydb/users", "age > 26", nil)
	if err != nil {
		t.Fatalf("FindRows failed: %v", err)
	}

	// Should return Alice (30) and Charlie (35)
	if len(rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(rows))
	}
}

func TestFindRowsWithIndex(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"name": models.NewValue("Alice"),
		"city": models.NewValue("NYC"),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"name": models.NewValue("Bob"),
		"city": models.NewValue("LA"),
	})
	tree.InsertRowWithID("root/mydb/users", "user3", models.Row{
		"name": models.NewValue("Charlie"),
		"city": models.NewValue("NYC"),
	})

	// Create index on city
	tree.CreateIndex("root/mydb/users", "idx_city", []string{"city"}, false)

	// Find rows with city == "NYC" (should use index)
	rows, err := tree.FindRows("root/mydb/users", `city == "NYC"`, nil)
	if err != nil {
		t.Fatalf("FindRows failed: %v", err)
	}

	// Should return Alice and Charlie
	if len(rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(rows))
	}

	// Verify the query used index scan
	plan, _ := tree.AnalyzeQuery("root/mydb/users", parser.ParseExprQuery(`city == "NYC"`))
	if plan.Strategy != StrategyIndexScan {
		t.Error("Expected index scan strategy")
	}
}

func TestFindRowsWithLimit(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert 10 rows
	for i := 0; i < 10; i++ {
		tree.InsertRowWithID("root/mydb/users", fmt.Sprintf("user%d", i), models.Row{
			"idx": models.NewValue(i),
		})
	}

	// Find rows with limit of 5
	opts := &QueryOptions{Limit: 5}
	rows, err := tree.FindRows("root/mydb/users", "idx >= 0", opts)
	if err != nil {
		t.Fatalf("FindRows failed: %v", err)
	}

	if len(rows) != 5 {
		t.Errorf("Expected 5 rows due to limit, got %d", len(rows))
	}
}

func TestFindRowsWithSkip(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert 10 rows
	for i := 0; i < 10; i++ {
		tree.InsertRowWithID("root/mydb/users", fmt.Sprintf("user%02d", i), models.Row{
			"idx": models.NewValue(i),
		})
	}

	// Find rows with skip of 3
	opts := &QueryOptions{Skip: 3, Limit: 5}
	rows, err := tree.FindRows("root/mydb/users", "idx >= 0", opts)
	if err != nil {
		t.Fatalf("FindRows failed: %v", err)
	}

	// Should return 5 rows starting from position 3
	if len(rows) != 5 {
		t.Errorf("Expected 5 rows, got %d", len(rows))
	}
}

func TestFindRowsWithSorting(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert unsorted data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"name": models.NewValue("Charlie"),
		"age":  models.NewValue(30),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"name": models.NewValue("Alice"),
		"age":  models.NewValue(25),
	})
	tree.InsertRowWithID("root/mydb/users", "user3", models.Row{
		"name": models.NewValue("Bob"),
		"age":  models.NewValue(35),
	})

	// Find rows sorted by name
	opts := &QueryOptions{
		SortFields: []SortField{{FieldName: "name", Direction: SortAsc}},
	}
	rows, err := tree.FindRows("root/mydb/users", "age > 0", opts)
	if err != nil {
		t.Fatalf("FindRows failed: %v", err)
	}

	// Should be sorted: Alice, Bob, Charlie
	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}

	if rows[0]["name"].AsString() != "Alice" {
		t.Errorf("Expected first row to be Alice, got %s", rows[0]["name"].AsString())
	}
}

func TestCountWhere(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"age": models.NewValue(30),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"age": models.NewValue(25),
	})
	tree.InsertRowWithID("root/mydb/users", "user3", models.Row{
		"age": models.NewValue(35),
	})

	// Count users with age > 26
	count, err := tree.CountWhere("root/mydb/users", "age > 26")
	if err != nil {
		t.Fatalf("CountWhere failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestUpdateRowsWhere(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"name":   models.NewValue("Alice"),
		"status": models.NewValue("pending"),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"name":   models.NewValue("Bob"),
		"status": models.NewValue("pending"),
	})
	tree.InsertRowWithID("root/mydb/users", "user3", models.Row{
		"name":   models.NewValue("Charlie"),
		"status": models.NewValue("active"),
	})

	// Update all pending users to active
	count, err := tree.UpdateRowsWhere("root/mydb/users", `status == "pending"`, map[string]interface{}{
		"status": "active",
	})

	if err != nil {
		t.Fatalf("UpdateRowsWhere failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected to update 2 rows, updated %d", count)
	}

	// Verify updates
	row, _ := tree.GetRow("root/mydb/users", "user1")
	if row["status"].AsString() != "active" {
		t.Error("Expected user1 status to be 'active'")
	}
}

func TestDeleteRowsWhere(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"age": models.NewValue(30),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"age": models.NewValue(25),
	})
	tree.InsertRowWithID("root/mydb/users", "user3", models.Row{
		"age": models.NewValue(35),
	})

	// Delete users with age < 30
	count, err := tree.DeleteRowsWhere("root/mydb/users", "age < 30")
	if err != nil {
		t.Fatalf("DeleteRowsWhere failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected to delete 1 row, deleted %d", count)
	}

	// Verify deletion
	remaining, _ := tree.GetRowCount("root/mydb/users")
	if remaining != 2 {
		t.Errorf("Expected 2 remaining rows, got %d", remaining)
	}
}

func TestFirstRow(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"name": models.NewValue("Charlie"),
		"age":  models.NewValue(30),
	})
	tree.InsertRowWithID("root/mydb/users", "user2", models.Row{
		"name": models.NewValue("Alice"),
		"age":  models.NewValue(25),
	})

	// Get first row sorted by name
	row, err := tree.FirstRow("root/mydb/users", "age > 0", []SortField{
		{FieldName: "name", Direction: SortAsc},
	})

	if err != nil {
		t.Fatalf("FirstRow failed: %v", err)
	}

	if row == nil {
		t.Fatal("Expected a row, got nil")
	}

	if row["name"].AsString() != "Alice" {
		t.Errorf("Expected first row to be Alice, got %s", row["name"].AsString())
	}
}

func TestExistsWhere(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"name": models.NewValue("Alice"),
	})

	// Check if Alice exists
	exists, err := tree.ExistsWhere("root/mydb/users", `name == "Alice"`)
	if err != nil {
		t.Fatalf("ExistsWhere failed: %v", err)
	}

	if !exists {
		t.Error("Expected Alice to exist")
	}

	// Check if Bob exists
	exists, _ = tree.ExistsWhere("root/mydb/users", `name == "Bob"`)
	if exists {
		t.Error("Expected Bob not to exist")
	}
}

func TestFindRowsCursor(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	for i := 0; i < 5; i++ {
		tree.InsertRowWithID("root/mydb/users", fmt.Sprintf("user%d", i), models.Row{
			"idx": models.NewValue(i),
		})
	}

	// Get cursor with limit
	opts := &QueryOptions{Limit: 3}
	cursor, err := tree.FindRowsCursor("root/mydb/users", "idx >= 0", opts)
	if err != nil {
		t.Fatalf("FindRowsCursor failed: %v", err)
	}

	// Verify cursor behavior
	if cursor.Count() != 5 {
		t.Errorf("Expected count 5, got %d", cursor.Count())
	}

	// Iterate with limit
	count := 0
	for i := 0; i < cursor.Count(); i++ {
		count++
		if !cursor.Next() {
			break
		}
	}

	// Should stop at 3 due to limit
	if count != 3 {
		t.Errorf("Expected to iterate 3 times with limit, got %d", count)
	}
}

func TestAnalyzeQuery(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	// Setup
	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert test data
	tree.InsertRowWithID("root/mydb/users", "user1", models.Row{
		"name": models.NewValue("Alice"),
		"age":  models.NewValue(30),
	})

	// Create index
	tree.CreateIndex("root/mydb/users", "idx_name", []string{"name"}, false)

	// Test full scan query
	plan1, err := tree.AnalyzeQuery("root/mydb/users", parser.ParseExprQuery("age > 25"))
	if err != nil {
		t.Fatalf("AnalyzeQuery failed: %v", err)
	}

	if plan1.Strategy != StrategyFullScan {
		t.Error("Expected full scan strategy for non-indexed field")
	}

	// Test index scan query
	plan2, err := tree.AnalyzeQuery("root/mydb/users", parser.ParseExprQuery(`name == "Alice"`))
	if err != nil {
		t.Fatalf("AnalyzeQuery failed: %v", err)
	}

	if plan2.Strategy != StrategyIndexScan {
		t.Error("Expected index scan strategy for indexed field")
	}

	if plan2.IndexName != "idx_name" {
		t.Errorf("Expected index name 'idx_name', got '%s'", plan2.IndexName)
	}
}

// ============================================================================
// Transaction Tests
// ============================================================================

func TestTransactionCommit(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Start transaction
	txn, err := tree.BeginTransaction()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer txn.Rollback()

	// Insert rows in transaction
	err = txn.InsertRowWithID("root/mydb/users", "user1", models.NewRow(map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}))
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	err = txn.InsertRowWithID("root/mydb/users", "user2", models.NewRow(map[string]interface{}{
		"name": "Bob",
		"age":  25,
	}))
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	// Commit transaction
	err = txn.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify rows exist
	row1, err := tree.GetRow("root/mydb/users", "user1")
	if err != nil {
		t.Errorf("Row user1 should exist after commit: %v", err)
	}
	if row1["name"].AsString() != "Alice" {
		t.Errorf("Expected name 'Alice', got '%v'", row1["name"].AsString())
	}

	row2, err := tree.GetRow("root/mydb/users", "user2")
	if err != nil {
		t.Errorf("Row user2 should exist after commit: %v", err)
	}
	if row2["name"].AsString() != "Bob" {
		t.Errorf("Expected name 'Bob', got '%v'", row2["name"].AsString())
	}
}

func TestTransactionRollback(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert initial row
	tree.InsertRowWithID("root/mydb/users", "user1", models.NewRow(map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}))

	// Start transaction
	txn, err := tree.BeginTransaction()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Insert row in transaction
	err = txn.InsertRowWithID("root/mydb/users", "user2", models.NewRow(map[string]interface{}{
		"name": "Bob",
		"age":  25,
	}))
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	// Update existing row in transaction
	err = txn.UpdateRowFields("root/mydb/users", "user1", map[string]interface{}{
		"age": 31,
	})
	if err != nil {
		t.Fatalf("Failed to update in transaction: %v", err)
	}

	// Rollback transaction
	err = txn.Rollback()
	if err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	// Verify user2 does NOT exist (was rolled back)
	_, err = tree.GetRow("root/mydb/users", "user2")
	if err == nil {
		t.Error("Row user2 should NOT exist after rollback")
	}

	// Verify user1 still has original age (update was rolled back)
	row1, err := tree.GetRow("root/mydb/users", "user1")
	if err != nil {
		t.Fatalf("Row user1 should still exist: %v", err)
	}
	if row1["age"].AsString() != "30" {
		t.Errorf("Expected age 30 (rollback), got %v", row1["age"].AsString())
	}
}

func TestTransactionUpdate(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert initial rows
	tree.InsertRowWithID("root/mydb/users", "user1", models.NewRow(map[string]interface{}{
		"name":  "Alice",
		"age":   30,
		"email": "alice@example.com",
	}))

	// Start transaction
	txn, err := tree.BeginTransaction()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer txn.Rollback()

	// Update multiple fields
	err = txn.UpdateRowFields("root/mydb/users", "user1", map[string]interface{}{
		"age":  35,
		"city": "NYC",
	})
	if err != nil {
		t.Fatalf("Failed to update in transaction: %v", err)
	}

	// Commit
	err = txn.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Verify updates
	row, err := tree.GetRow("root/mydb/users", "user1")
	if err != nil {
		t.Fatalf("Failed to get row: %v", err)
	}

	if row["age"].AsString() != "35" {
		t.Errorf("Expected age 35, got %v", row["age"].AsString())
	}

	if row["city"].AsString() != "NYC" {
		t.Errorf("Expected city 'NYC', got %v", row["city"].AsString())
	}
}

func TestTransactionDelete(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert rows
	tree.InsertRowWithID("root/mydb/users", "user1", models.NewRow(map[string]interface{}{
		"name": "Alice",
	}))
	tree.InsertRowWithID("root/mydb/users", "user2", models.NewRow(map[string]interface{}{
		"name": "Bob",
	}))
	tree.InsertRowWithID("root/mydb/users", "user3", models.NewRow(map[string]interface{}{
		"name": "Charlie",
	}))

	// Start transaction
	txn, err := tree.BeginTransaction()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer txn.Rollback()

	// Delete user2
	err = txn.DeleteRow("root/mydb/users", "user2")
	if err != nil {
		t.Fatalf("Failed to delete in transaction: %v", err)
	}

	// Commit
	err = txn.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Verify user2 is deleted
	_, err = tree.GetRow("root/mydb/users", "user2")
	if err == nil {
		t.Error("user2 should be deleted")
	}

	// Verify other rows still exist
	_, err = tree.GetRow("root/mydb/users", "user1")
	if err != nil {
		t.Error("user1 should still exist")
	}

	_, err = tree.GetRow("root/mydb/users", "user3")
	if err != nil {
		t.Error("user3 should still exist")
	}
}

func TestTransactionMultipleCommits(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	txn, err := tree.BeginTransaction()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	txn.InsertRowWithID("root/mydb/users", "user1", models.NewRow(map[string]interface{}{
		"name": "Alice",
	}))

	err = txn.Commit()
	if err != nil {
		t.Fatalf("First commit failed: %v", err)
	}

	// Second commit should fail
	err = txn.Commit()
	if err == nil {
		t.Error("Second commit should fail")
	}
}

func TestTransactionIsActive(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	txn, _ := tree.BeginTransaction()

	if !txn.IsActive() {
		t.Error("Transaction should be active after begin")
	}

	txn.Commit()

	if txn.IsActive() {
		t.Error("Transaction should not be active after commit")
	}

	txn2, _ := tree.BeginTransaction()

	if !txn2.IsActive() {
		t.Error("Transaction should be active after begin")
	}

	txn2.Rollback()

	if txn2.IsActive() {
		t.Error("Transaction should not be active after rollback")
	}
}

// ============================================================================
// Upsert Tests
// ============================================================================

func TestUpsertInsert(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Upsert on non-existent row should insert
	wasInserted, err := tree.Upsert("root/mydb/users", "user1", models.NewRow(map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}))

	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	if !wasInserted {
		t.Error("Expected insert, got update")
	}

	// Verify row exists
	row, err := tree.GetRow("root/mydb/users", "user1")
	if err != nil {
		t.Fatalf("Row should exist: %v", err)
	}

	if row["name"].AsString() != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", row["name"].AsString())
	}
}

func TestUpsertUpdate(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert initial row
	tree.InsertRowWithID("root/mydb/users", "user1", models.NewRow(map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}))

	// Upsert on existing row should update
	wasInserted, err := tree.Upsert("root/mydb/users", "user1", models.NewRow(map[string]interface{}{
		"name": "Alice Updated",
		"age":  31,
	}))

	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	if wasInserted {
		t.Error("Expected update, got insert")
	}

	// Verify row was updated
	row, err := tree.GetRow("root/mydb/users", "user1")
	if err != nil {
		t.Fatalf("Row should exist: %v", err)
	}

	if row["name"].AsString() != "Alice Updated" {
		t.Errorf("Expected name 'Alice Updated', got %v", row["name"].AsString())
	}

	if row["age"].AsString() != "31" {
		t.Errorf("Expected age 31, got %v", row["age"].AsString())
	}
}

func TestUpsertMany(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert one existing row
	tree.InsertRowWithID("root/mydb/users", "user1", models.NewRow(map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}))

	// Upsert multiple rows (1 update, 2 inserts)
	rows := map[string]models.Row{
		"user1": models.NewRow(map[string]interface{}{
			"name": "Alice Updated",
			"age":  31,
		}),
		"user2": models.NewRow(map[string]interface{}{
			"name": "Bob",
			"age":  25,
		}),
		"user3": models.NewRow(map[string]interface{}{
			"name": "Charlie",
			"age":  35,
		}),
	}

	insertCount, updateCount, err := tree.UpsertMany("root/mydb/users", rows)
	if err != nil {
		t.Fatalf("UpsertMany failed: %v", err)
	}

	if insertCount != 2 {
		t.Errorf("Expected 2 inserts, got %d", insertCount)
	}

	if updateCount != 1 {
		t.Errorf("Expected 1 update, got %d", updateCount)
	}

	// Verify all rows
	row1, _ := tree.GetRow("root/mydb/users", "user1")
	if row1["name"].AsString() != "Alice Updated" {
		t.Error("user1 should be updated")
	}

	row2, _ := tree.GetRow("root/mydb/users", "user2")
	if row2["name"].AsString() != "Bob" {
		t.Error("user2 should be inserted")
	}

	row3, _ := tree.GetRow("root/mydb/users", "user3")
	if row3["name"].AsString() != "Charlie" {
		t.Error("user3 should be inserted")
	}
}

func TestUpsertFields(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Upsert fields on non-existent row (insert)
	wasInserted, err := tree.UpsertFields("root/mydb/users", "user1", map[string]interface{}{
		"name": "Alice",
		"age":  30,
	})

	if err != nil {
		t.Fatalf("UpsertFields failed: %v", err)
	}

	if !wasInserted {
		t.Error("Expected insert")
	}

	// Upsert fields on existing row (update)
	wasInserted, err = tree.UpsertFields("root/mydb/users", "user1", map[string]interface{}{
		"age":  31,
		"city": "NYC",
	})

	if err != nil {
		t.Fatalf("UpsertFields failed: %v", err)
	}

	if wasInserted {
		t.Error("Expected update")
	}

	// Verify fields
	row, _ := tree.GetRow("root/mydb/users", "user1")
	if row["age"].AsString() != "31" {
		t.Errorf("Expected age 31, got %v", row["age"].AsString())
	}

	if row["city"].AsString() != "NYC" {
		t.Errorf("Expected city 'NYC', got %v", row["city"].AsString())
	}
}

// ============================================================================
// Bulk Operations Tests
// ============================================================================

func TestBulkInsert(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	rows := map[string]models.Row{
		"user1": models.NewRow(map[string]interface{}{"name": "Alice", "age": 30}),
		"user2": models.NewRow(map[string]interface{}{"name": "Bob", "age": 25}),
		"user3": models.NewRow(map[string]interface{}{"name": "Charlie", "age": 35}),
	}

	count, err := tree.BulkInsert("root/mydb/users", rows)
	if err != nil {
		t.Fatalf("BulkInsert failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 inserts, got %d", count)
	}

	// Verify all rows exist
	for rowID := range rows {
		_, err := tree.GetRow("root/mydb/users", rowID)
		if err != nil {
			t.Errorf("Row %s should exist", rowID)
		}
	}
}

func TestBulkUpdate(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert initial rows
	tree.InsertRowWithID("root/mydb/users", "user1", models.NewRow(map[string]interface{}{"name": "Alice", "age": 30}))
	tree.InsertRowWithID("root/mydb/users", "user2", models.NewRow(map[string]interface{}{"name": "Bob", "age": 25}))

	// Bulk update
	updates := map[string]models.Row{
		"user1": models.NewRow(map[string]interface{}{"name": "Alice Updated", "age": 31}),
		"user2": models.NewRow(map[string]interface{}{"name": "Bob Updated", "age": 26}),
	}

	count, err := tree.BulkUpdate("root/mydb/users", updates)
	if err != nil {
		t.Fatalf("BulkUpdate failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 updates, got %d", count)
	}

	// Verify updates
	row1, _ := tree.GetRow("root/mydb/users", "user1")
	if row1["name"].AsString() != "Alice Updated" {
		t.Error("user1 name should be updated")
	}

	row2, _ := tree.GetRow("root/mydb/users", "user2")
	if row2["age"].AsString() != "26" {
		t.Error("user2 age should be updated")
	}
}

func TestBulkUpdateFields(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert initial rows
	tree.InsertRowWithID("root/mydb/users", "user1", models.NewRow(map[string]interface{}{"name": "Alice", "age": 30}))
	tree.InsertRowWithID("root/mydb/users", "user2", models.NewRow(map[string]interface{}{"name": "Bob", "age": 25}))

	// Bulk update fields
	updates := map[string]map[string]interface{}{
		"user1": {"age": 31, "city": "NYC"},
		"user2": {"age": 26, "city": "LA"},
	}

	count, err := tree.BulkUpdateFields("root/mydb/users", updates)
	if err != nil {
		t.Fatalf("BulkUpdateFields failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 updates, got %d", count)
	}

	// Verify field updates
	row1, _ := tree.GetRow("root/mydb/users", "user1")
	if row1["city"].AsString() != "NYC" {
		t.Error("user1 city should be NYC")
	}

	row2, _ := tree.GetRow("root/mydb/users", "user2")
	if row2["city"].AsString() != "LA" {
		t.Error("user2 city should be LA")
	}
}

func TestBulkDelete(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert rows
	tree.InsertRowWithID("root/mydb/users", "user1", models.NewRow(map[string]interface{}{"name": "Alice"}))
	tree.InsertRowWithID("root/mydb/users", "user2", models.NewRow(map[string]interface{}{"name": "Bob"}))
	tree.InsertRowWithID("root/mydb/users", "user3", models.NewRow(map[string]interface{}{"name": "Charlie"}))
	tree.InsertRowWithID("root/mydb/users", "user4", models.NewRow(map[string]interface{}{"name": "David"}))

	// Bulk delete
	count, err := tree.BulkDelete("root/mydb/users", []string{"user2", "user3"})
	if err != nil {
		t.Fatalf("BulkDelete failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 deletes, got %d", count)
	}

	// Verify deletions
	_, err = tree.GetRow("root/mydb/users", "user2")
	if err == nil {
		t.Error("user2 should be deleted")
	}

	_, err = tree.GetRow("root/mydb/users", "user3")
	if err == nil {
		t.Error("user3 should be deleted")
	}

	// Verify remaining rows
	_, err = tree.GetRow("root/mydb/users", "user1")
	if err != nil {
		t.Error("user1 should still exist")
	}

	_, err = tree.GetRow("root/mydb/users", "user4")
	if err != nil {
		t.Error("user4 should still exist")
	}
}

func TestBulkUpsert(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert some existing rows
	tree.InsertRowWithID("root/mydb/users", "user1", models.NewRow(map[string]interface{}{"name": "Alice", "age": 30}))
	tree.InsertRowWithID("root/mydb/users", "user2", models.NewRow(map[string]interface{}{"name": "Bob", "age": 25}))

	// Bulk upsert (1 update, 2 inserts)
	rows := map[string]models.Row{
		"user1": models.NewRow(map[string]interface{}{"name": "Alice Updated", "age": 31}),
		"user3": models.NewRow(map[string]interface{}{"name": "Charlie", "age": 35}),
		"user4": models.NewRow(map[string]interface{}{"name": "David", "age": 40}),
	}

	insertCount, updateCount, err := tree.BulkUpsert("root/mydb/users", rows)
	if err != nil {
		t.Fatalf("BulkUpsert failed: %v", err)
	}

	if insertCount != 2 {
		t.Errorf("Expected 2 inserts, got %d", insertCount)
	}

	if updateCount != 1 {
		t.Errorf("Expected 1 update, got %d", updateCount)
	}

	// Verify results
	row1, _ := tree.GetRow("root/mydb/users", "user1")
	if row1["name"].AsString() != "Alice Updated" {
		t.Error("user1 should be updated")
	}

	row3, _ := tree.GetRow("root/mydb/users", "user3")
	if row3["name"].AsString() != "Charlie" {
		t.Error("user3 should be inserted")
	}
}

// ============================================================================
// Range Query Tests
// ============================================================================

func TestFindRowsRange(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "products")

	// Insert test data
	tree.InsertRowWithID("root/mydb/products", "p1", models.NewRow(map[string]interface{}{"name": "A", "price": 10}))
	tree.InsertRowWithID("root/mydb/products", "p2", models.NewRow(map[string]interface{}{"name": "B", "price": 20}))
	tree.InsertRowWithID("root/mydb/products", "p3", models.NewRow(map[string]interface{}{"name": "C", "price": 30}))
	tree.InsertRowWithID("root/mydb/products", "p4", models.NewRow(map[string]interface{}{"name": "D", "price": 40}))

	// Create index on price
	tree.CreateIndex("root/mydb/products", "idx_price", []string{"price"}, false)

	// Find rows with price between 15 and 35
	rows, err := tree.FindRowsRange("root/mydb/products", "idx_price", 15, 35)
	if err != nil {
		t.Fatalf("FindRowsRange failed: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("Expected 2 rows (price 20, 30), got %d", len(rows))
	}
}

func TestFindRowsBetween(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "scores")

	// Insert test data
	for i := 1; i <= 10; i++ {
		tree.InsertRowWithID("root/mydb/scores", fmt.Sprintf("s%d", i), models.NewRow(map[string]interface{}{"value": i * 10}))
	}

	tree.CreateIndex("root/mydb/scores", "idx_value", []string{"value"}, false)

	// Find scores between 30 and 70
	rows, err := tree.FindRowsBetween("root/mydb/scores", "idx_value", 30, 70)
	if err != nil {
		t.Fatalf("FindRowsBetween failed: %v", err)
	}

	if len(rows) != 5 {
		t.Errorf("Expected 5 rows (30, 40, 50, 60, 70), got %d", len(rows))
	}
}

func TestFindRowsGreaterThan(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "items")

	tree.InsertRowWithID("root/mydb/items", "i1", models.NewRow(map[string]interface{}{"qty": 5}))
	tree.InsertRowWithID("root/mydb/items", "i2", models.NewRow(map[string]interface{}{"qty": 10}))
	tree.InsertRowWithID("root/mydb/items", "i3", models.NewRow(map[string]interface{}{"qty": 15}))

	tree.CreateIndex("root/mydb/items", "idx_qty", []string{"qty"}, false)

	rows, err := tree.FindRowsGreaterThan("root/mydb/items", "idx_qty", 7)
	if err != nil {
		t.Fatalf("FindRowsGreaterThan failed: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("Expected 2 rows (10, 15), got %d", len(rows))
	}
}

func TestFindRowsLessThanOrEqual(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "items")

	tree.InsertRowWithID("root/mydb/items", "i1", models.NewRow(map[string]interface{}{"qty": 5}))
	tree.InsertRowWithID("root/mydb/items", "i2", models.NewRow(map[string]interface{}{"qty": 10}))
	tree.InsertRowWithID("root/mydb/items", "i3", models.NewRow(map[string]interface{}{"qty": 15}))

	tree.CreateIndex("root/mydb/items", "idx_qty", []string{"qty"}, false)

	rows, err := tree.FindRowsLessThanOrEqual("root/mydb/items", "idx_qty", 10)
	if err != nil {
		t.Fatalf("FindRowsLessThanOrEqual failed: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("Expected 2 rows (5, 10), got %d", len(rows))
	}
}

// ============================================================================
// Distinct Value Tests
// ============================================================================

func TestGetDistinctValues(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")

	// Insert rows with duplicate cities
	tree.InsertRowWithID("root/mydb/users", "u1", models.NewRow(map[string]interface{}{"city": "NYC"}))
	tree.InsertRowWithID("root/mydb/users", "u2", models.NewRow(map[string]interface{}{"city": "LA"}))
	tree.InsertRowWithID("root/mydb/users", "u3", models.NewRow(map[string]interface{}{"city": "NYC"}))
	tree.InsertRowWithID("root/mydb/users", "u4", models.NewRow(map[string]interface{}{"city": "SF"}))
	tree.InsertRowWithID("root/mydb/users", "u5", models.NewRow(map[string]interface{}{"city": "LA"}))

	values, err := tree.GetDistinctValues("root/mydb/users", "city")
	if err != nil {
		t.Fatalf("GetDistinctValues failed: %v", err)
	}

	if len(values) != 3 {
		t.Errorf("Expected 3 distinct cities (NYC, LA, SF), got %d", len(values))
	}
}

func TestCountDistinct(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "orders")

	// Insert rows
	tree.InsertRowWithID("root/mydb/orders", "o1", models.NewRow(map[string]interface{}{"customer_id": "c1"}))
	tree.InsertRowWithID("root/mydb/orders", "o2", models.NewRow(map[string]interface{}{"customer_id": "c2"}))
	tree.InsertRowWithID("root/mydb/orders", "o3", models.NewRow(map[string]interface{}{"customer_id": "c1"}))
	tree.InsertRowWithID("root/mydb/orders", "o4", models.NewRow(map[string]interface{}{"customer_id": "c3"}))

	count, err := tree.CountDistinct("root/mydb/orders", "customer_id")
	if err != nil {
		t.Fatalf("CountDistinct failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 distinct customers, got %d", count)
	}
}

// ============================================================================
// Atomic Operation Tests
// ============================================================================

func TestIncrementField(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "counters")

	tree.InsertRowWithID("root/mydb/counters", "c1", models.NewRow(map[string]interface{}{"count": 10}))

	// Increment by 5
	newValue, err := tree.IncrementField("root/mydb/counters", "c1", "count", 5)
	if err != nil {
		t.Fatalf("IncrementField failed: %v", err)
	}

	if newValue != 15 {
		t.Errorf("Expected 15, got %d", newValue)
	}

	// Verify value in database
	row, _ := tree.GetRow("root/mydb/counters", "c1")
	count, _ := row["count"].AsInt64()
	if count != 15 {
		t.Errorf("Expected count 15 in DB, got %d", count)
	}
}

func TestDecrementField(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "inventory")

	tree.InsertRowWithID("root/mydb/inventory", "item1", models.NewRow(map[string]interface{}{"stock": 100}))

	newValue, err := tree.DecrementField("root/mydb/inventory", "item1", "stock", 25)
	if err != nil {
		t.Fatalf("DecrementField failed: %v", err)
	}

	if newValue != 75 {
		t.Errorf("Expected 75, got %d", newValue)
	}
}

func TestSetFieldIfNotExists(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "config")

	tree.InsertRowWithID("root/mydb/config", "cfg1", models.NewRow(map[string]interface{}{"key": "timeout"}))

	// Set value on non-existing field
	wasSet, err := tree.SetFieldIfNotExists("root/mydb/config", "cfg1", "value", "30s")
	if err != nil {
		t.Fatalf("SetFieldIfNotExists failed: %v", err)
	}

	if !wasSet {
		t.Error("Expected field to be set")
	}

	// Try to set again (should not overwrite)
	wasSet, err = tree.SetFieldIfNotExists("root/mydb/config", "cfg1", "value", "60s")
	if err != nil {
		t.Fatalf("SetFieldIfNotExists failed: %v", err)
	}

	if wasSet {
		t.Error("Expected field NOT to be set (already exists)")
	}

	// Verify value is still 30s
	row, _ := tree.GetRow("root/mydb/config", "cfg1")
	if row["value"] == nil || row["value"].AsString() != "30s" {
		t.Errorf("Value should still be 30s, got: %v", row["value"])
	}
}

// ============================================================================
// Join Operation Tests
// ============================================================================

func TestInnerJoin(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")
	tree.CreateTable("root/mydb", "orders")

	// Insert users
	tree.InsertRowWithID("root/mydb/users", "u1", models.NewRow(map[string]interface{}{"user_id": "1", "name": "Alice"}))
	tree.InsertRowWithID("root/mydb/users", "u2", models.NewRow(map[string]interface{}{"user_id": "2", "name": "Bob"}))
	tree.InsertRowWithID("root/mydb/users", "u3", models.NewRow(map[string]interface{}{"user_id": "3", "name": "Charlie"}))

	// Insert orders (only for users 1 and 2)
	tree.InsertRowWithID("root/mydb/orders", "o1", models.NewRow(map[string]interface{}{"user_id": "1", "amount": 100}))
	tree.InsertRowWithID("root/mydb/orders", "o2", models.NewRow(map[string]interface{}{"user_id": "1", "amount": 200}))
	tree.InsertRowWithID("root/mydb/orders", "o3", models.NewRow(map[string]interface{}{"user_id": "2", "amount": 150}))

	// Inner join
	results, err := tree.InnerJoin("root/mydb/users", "root/mydb/orders", "user_id", "user_id")
	if err != nil {
		t.Fatalf("InnerJoin failed: %v", err)
	}

	// Should return 3 results (user1 with 2 orders, user2 with 1 order)
	// User3 has no orders, so not included
	if len(results) != 3 {
		t.Errorf("Expected 3 join results, got %d", len(results))
	}

	// Verify one of the results
	found := false
	for _, r := range results {
		if r.LeftRow["name"].AsString() == "Alice" && r.RightRow["amount"].AsString() == "100" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find Alice with order amount 100")
	}
}

func TestLeftJoin(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "users")
	tree.CreateTable("root/mydb", "orders")

	// Insert users
	tree.InsertRowWithID("root/mydb/users", "u1", models.NewRow(map[string]interface{}{"user_id": "1", "name": "Alice"}))
	tree.InsertRowWithID("root/mydb/users", "u2", models.NewRow(map[string]interface{}{"user_id": "2", "name": "Bob"}))
	tree.InsertRowWithID("root/mydb/users", "u3", models.NewRow(map[string]interface{}{"user_id": "3", "name": "Charlie"}))

	// Insert orders (only for users 1 and 2)
	tree.InsertRowWithID("root/mydb/orders", "o1", models.NewRow(map[string]interface{}{"user_id": "1", "amount": 100}))
	tree.InsertRowWithID("root/mydb/orders", "o2", models.NewRow(map[string]interface{}{"user_id": "2", "amount": 150}))

	// Left join
	results, err := tree.LeftJoin("root/mydb/users", "root/mydb/orders", "user_id", "user_id")
	if err != nil {
		t.Fatalf("LeftJoin failed: %v", err)
	}

	// Should return 3 results (all users, Charlie with nil RightRow)
	if len(results) != 3 {
		t.Errorf("Expected 3 join results, got %d", len(results))
	}

	// Find Charlie with no orders
	found := false
	for _, r := range results {
		if r.LeftRow["name"].AsString() == "Charlie" && r.RightRow == nil {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected Charlie with nil RightRow (no orders)")
	}
}

// ============================================================================
// Aggregation Tests
// ============================================================================

func TestAggregate(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "sales")

	// Insert test data
	tree.InsertRowWithID("root/mydb/sales", "s1", models.NewRow(map[string]interface{}{"amount": 100}))
	tree.InsertRowWithID("root/mydb/sales", "s2", models.NewRow(map[string]interface{}{"amount": 200}))
	tree.InsertRowWithID("root/mydb/sales", "s3", models.NewRow(map[string]interface{}{"amount": 150}))

	result, err := tree.Aggregate("root/mydb/sales", "amount")
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}

	if result.Count != 3 {
		t.Errorf("Expected count 3, got %d", result.Count)
	}

	if result.Sum != 450 {
		t.Errorf("Expected sum 450, got %f", result.Sum)
	}

	if result.Avg != 150 {
		t.Errorf("Expected avg 150, got %f", result.Avg)
	}
}

func TestSum(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "transactions")

	tree.InsertRowWithID("root/mydb/transactions", "t1", models.NewRow(map[string]interface{}{"amount": 50}))
	tree.InsertRowWithID("root/mydb/transactions", "t2", models.NewRow(map[string]interface{}{"amount": 75}))
	tree.InsertRowWithID("root/mydb/transactions", "t3", models.NewRow(map[string]interface{}{"amount": 25}))

	sum, err := tree.Sum("root/mydb/transactions", "amount")
	if err != nil {
		t.Fatalf("Sum failed: %v", err)
	}

	if sum != 150 {
		t.Errorf("Expected sum 150, got %f", sum)
	}
}

func TestAvg(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "scores")

	tree.InsertRowWithID("root/mydb/scores", "s1", models.NewRow(map[string]interface{}{"score": 80}))
	tree.InsertRowWithID("root/mydb/scores", "s2", models.NewRow(map[string]interface{}{"score": 90}))
	tree.InsertRowWithID("root/mydb/scores", "s3", models.NewRow(map[string]interface{}{"score": 70}))

	avg, err := tree.Avg("root/mydb/scores", "score")
	if err != nil {
		t.Fatalf("Avg failed: %v", err)
	}

	if avg != 80 {
		t.Errorf("Expected avg 80, got %f", avg)
	}
}

func TestMinMax(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "temperatures")

	tree.InsertRowWithID("root/mydb/temperatures", "t1", models.NewRow(map[string]interface{}{"temp": 20}))
	tree.InsertRowWithID("root/mydb/temperatures", "t2", models.NewRow(map[string]interface{}{"temp": 35}))
	tree.InsertRowWithID("root/mydb/temperatures", "t3", models.NewRow(map[string]interface{}{"temp": 15}))

	min, err := tree.Min("root/mydb/temperatures", "temp")
	if err != nil {
		t.Fatalf("Min failed: %v", err)
	}

	max, err := tree.Max("root/mydb/temperatures", "temp")
	if err != nil {
		t.Fatalf("Max failed: %v", err)
	}

	minFloat, _ := models.NewValue(min).AsFloat64()
	maxFloat, _ := models.NewValue(max).AsFloat64()

	if minFloat != 15 {
		t.Errorf("Expected min 15, got %f", minFloat)
	}

	if maxFloat != 35 {
		t.Errorf("Expected max 35, got %f", maxFloat)
	}
}

func TestGroupBy(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "sales")

	// Insert test data
	tree.InsertRowWithID("root/mydb/sales", "s1", models.NewRow(map[string]interface{}{"region": "West", "amount": 100}))
	tree.InsertRowWithID("root/mydb/sales", "s2", models.NewRow(map[string]interface{}{"region": "East", "amount": 200}))
	tree.InsertRowWithID("root/mydb/sales", "s3", models.NewRow(map[string]interface{}{"region": "West", "amount": 150}))
	tree.InsertRowWithID("root/mydb/sales", "s4", models.NewRow(map[string]interface{}{"region": "East", "amount": 250}))

	groups, err := tree.GroupBy("root/mydb/sales", "region", "amount")
	if err != nil {
		t.Fatalf("GroupBy failed: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups (West, East), got %d", len(groups))
	}

	// Check West group
	westGroup, exists := groups["West"]
	if !exists {
		t.Fatal("West group not found")
	}

	if westGroup.Count != 2 {
		t.Errorf("Expected West count 2, got %d", westGroup.Count)
	}

	if westGroup.Sum != 250 {
		t.Errorf("Expected West sum 250, got %f", westGroup.Sum)
	}

	if westGroup.Avg != 125 {
		t.Errorf("Expected West avg 125, got %f", westGroup.Avg)
	}
}

// ============================================================================
// Subquery Tests
// ============================================================================

func TestExistsInSubquery(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "customers")
	tree.CreateTable("root/mydb", "orders")

	// Insert customers
	tree.InsertRowWithID("root/mydb/customers", "c1", models.NewRow(map[string]interface{}{"customer_id": "1", "name": "Alice"}))
	tree.InsertRowWithID("root/mydb/customers", "c2", models.NewRow(map[string]interface{}{"customer_id": "2", "name": "Bob"}))
	tree.InsertRowWithID("root/mydb/customers", "c3", models.NewRow(map[string]interface{}{"customer_id": "3", "name": "Charlie"}))

	// Insert orders (only for customers 1 and 2)
	tree.InsertRowWithID("root/mydb/orders", "o1", models.NewRow(map[string]interface{}{"customer_id": "1"}))
	tree.InsertRowWithID("root/mydb/orders", "o2", models.NewRow(map[string]interface{}{"customer_id": "2"}))

	// Check which customers have orders
	results, err := tree.ExistsInSubquery("root/mydb/customers", "customer_id", "root/mydb/orders", "customer_id")
	if err != nil {
		t.Fatalf("ExistsInSubquery failed: %v", err)
	}

	if !results["c1"] {
		t.Error("Customer c1 should exist in orders")
	}

	if !results["c2"] {
		t.Error("Customer c2 should exist in orders")
	}

	if results["c3"] {
		t.Error("Customer c3 should NOT exist in orders")
	}
}

func TestInSubquery(t *testing.T) {
	tree := setupDatabaseTest(t)
	defer teardownDatabaseTest(t, tree)

	tree.CreateDatabase("root/mydb", nil)
	tree.CreateTable("root/mydb", "products")
	tree.CreateTable("root/mydb", "sold_products")

	// Insert products
	tree.InsertRowWithID("root/mydb/products", "p1", models.NewRow(map[string]interface{}{"product_id": "100", "name": "Widget"}))
	tree.InsertRowWithID("root/mydb/products", "p2", models.NewRow(map[string]interface{}{"product_id": "200", "name": "Gadget"}))
	tree.InsertRowWithID("root/mydb/products", "p3", models.NewRow(map[string]interface{}{"product_id": "300", "name": "Doohickey"}))

	// Insert sold products (only 100 and 300)
	tree.InsertRowWithID("root/mydb/sold_products", "sp1", models.NewRow(map[string]interface{}{"product_id": "100"}))
	tree.InsertRowWithID("root/mydb/sold_products", "sp2", models.NewRow(map[string]interface{}{"product_id": "300"}))

	// Get products that have been sold
	rowIDs, err := tree.InSubquery("root/mydb/products", "product_id", "root/mydb/sold_products", "product_id")
	if err != nil {
		t.Fatalf("InSubquery failed: %v", err)
	}

	if len(rowIDs) != 2 {
		t.Errorf("Expected 2 products sold, got %d", len(rowIDs))
	}

	// Verify p1 and p3 are in results, p2 is not
	hasP1, hasP2, hasP3 := false, false, false
	for _, id := range rowIDs {
		if id == "p1" {
			hasP1 = true
		}
		if id == "p2" {
			hasP2 = true
		}
		if id == "p3" {
			hasP3 = true
		}
	}

	if !hasP1 || !hasP3 {
		t.Error("Expected p1 and p3 in results")
	}

	if hasP2 {
		t.Error("p2 should NOT be in results")
	}
}

