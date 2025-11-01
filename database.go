package blueconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sfi2k7/blueconfig/models"
)

// ============================================================================
// Constants
// ============================================================================

const (
	TypeDatabase = "database"
	TypeTable    = "table"
	TypeView     = "view"
)

const (
	SchemaNode  = "__schema"
	IndicesNode = "__indices"
	StatsNode   = "__stats"
)

// ============================================================================
// Types
// ============================================================================

// TableSchema represents the dynamic schema of a table
type TableSchema struct {
	Fields      map[string]string `json:"fields"`       // field_name -> type
	FlatFields  []string          `json:"flat_fields"`  // all flattened field names
	Version     int               `json:"version"`      // schema version
	Created     string            `json:"created"`      // timestamp
	LastUpdated string            `json:"last_updated"` // timestamp
}

// DatabaseInfo represents database metadata
type DatabaseInfo struct {
	Path        string
	Type        string
	Created     string
	LastUpdated string
	TableCount  int
	ViewCount   int
}

// TableInfo represents table metadata
type TableInfo struct {
	Path        string
	Name        string
	Type        string
	Created     string
	LastUpdated string
	RowCount    int
	HasSchema   bool
}

// ============================================================================
// Database Operations
// ============================================================================

// IsDatabase checks if a node is a database
func (t *Tree) IsDatabase(path string) (bool, error) {
	nodeType, err := t.GetValue(path + "/__type")
	if err != nil {
		return false, nil
	}
	return nodeType == TypeDatabase, nil
}

// CreateDatabase creates a new database at the specified path
func (t *Tree) CreateDatabase(path string, metadata map[string]interface{}) error {
	// Create the path if it doesn't exist
	err := t.CreatePath(path)
	if err != nil {
		return err
	}

	// Set database type marker
	now := strconv.FormatInt(time.Now().Unix(), 10)
	dbMetadata := map[string]interface{}{
		"__type":        TypeDatabase,
		"__created":     now,
		"__lastupdated": now,
		"__table_count": "0",
		"__view_count":  "0",
	}

	// Merge with custom metadata
	if metadata != nil {
		for key, value := range metadata {
			if !strings.HasPrefix(key, "__") {
				dbMetadata[key] = value
			}
		}
	}

	return t.SetValues(path, dbMetadata)
}

// ListDatabases returns all databases under a given path
func (t *Tree) ListDatabases(rootPath string) ([]string, error) {
	var databases []string

	children, err := t.GetNodesInPath(rootPath)
	if err != nil {
		return databases, nil
	}

	for _, child := range children {
		childPath := rootPath + "/" + child
		if isDB, _ := t.IsDatabase(childPath); isDB {
			databases = append(databases, child)
		}
	}

	return databases, nil
}

// GetDatabaseInfo returns database metadata
func (t *Tree) GetDatabaseInfo(path string) (*DatabaseInfo, error) {
	isDB, err := t.IsDatabase(path)
	if err != nil {
		return nil, err
	}
	if !isDB {
		return nil, errors.New("path is not a database")
	}

	props, err := t.GetAllPropsWithValues(path)
	if err != nil {
		return nil, err
	}

	info := &DatabaseInfo{
		Path:        path,
		Type:        props["__type"],
		Created:     props["__created"],
		LastUpdated: props["__lastupdated"],
	}

	if count, err := strconv.Atoi(props["__table_count"]); err == nil {
		info.TableCount = count
	}
	if count, err := strconv.Atoi(props["__view_count"]); err == nil {
		info.ViewCount = count
	}

	return info, nil
}

// DeleteDatabase deletes a database
func (t *Tree) DeleteDatabase(path string, force bool) error {
	isDB, err := t.IsDatabase(path)
	if err != nil {
		return err
	}
	if !isDB {
		return errors.New("path is not a database")
	}

	// Check if it has children (tables/views) and force is false
	if !force {
		children, err := t.GetNodesInPath(path)
		if err != nil {
			return err
		}

		// Filter out special properties (starting with __)
		hasChildren := false
		for _, child := range children {
			if !strings.HasPrefix(child, "__") {
				hasChildren = true
				break
			}
		}

		if hasChildren {
			return errors.New("database has tables/views, use force=true to delete")
		}
	}

	return t.DeleteNode(path, force)
}

// ============================================================================
// Table Operations
// ============================================================================

// IsTable checks if a node is a table
func (t *Tree) IsTable(path string) (bool, error) {
	nodeType, err := t.GetValue(path + "/__type")
	if err != nil {
		return false, nil
	}
	return nodeType == TypeTable, nil
}

// CreateTable creates a new table within a database (no schema required initially)
func (t *Tree) CreateTable(dbPath, tableName string) error {
	// Verify parent is a database
	isDB, err := t.IsDatabase(dbPath)
	if err != nil {
		return err
	}
	if !isDB {
		return errors.New("parent path is not a database")
	}

	tablePath := dbPath + "/" + tableName

	// Create table with metadata
	now := strconv.FormatInt(time.Now().Unix(), 10)
	metadata := map[string]interface{}{
		"__type":        TypeTable,
		"__name":        tableName,
		"__created":     now,
		"__lastupdated": now,
		"__row_count":   "0",
		"__has_schema":  "false",
	}

	err = t.CreateNodeWithProps(tablePath, metadata)
	if err != nil {
		return err
	}

	// Update database table count
	t.updateDatabaseTableCount(dbPath)

	return nil
}

// ListTables returns all tables in a database
func (t *Tree) ListTables(dbPath string) ([]string, error) {
	isDB, err := t.IsDatabase(dbPath)
	if err != nil {
		return nil, err
	}
	if !isDB {
		return nil, errors.New("path is not a database")
	}

	var tables []string
	children, err := t.GetNodesInPath(dbPath)
	if err != nil {
		return tables, nil
	}

	for _, child := range children {
		// Skip special nodes
		if strings.HasPrefix(child, "__") {
			continue
		}

		childPath := dbPath + "/" + child
		if isTable, _ := t.IsTable(childPath); isTable {
			tables = append(tables, child)
		}
	}

	return tables, nil
}

// GetTableInfo returns table metadata
func (t *Tree) GetTableInfo(path string) (*TableInfo, error) {
	isTable, err := t.IsTable(path)
	if err != nil {
		return nil, err
	}
	if !isTable {
		return nil, errors.New("path is not a table")
	}

	props, err := t.GetAllPropsWithValues(path)
	if err != nil {
		return nil, err
	}

	info := &TableInfo{
		Path:        path,
		Name:        props["__name"],
		Type:        props["__type"],
		Created:     props["__created"],
		LastUpdated: props["__lastupdated"],
	}

	if count, err := strconv.Atoi(props["__row_count"]); err == nil {
		info.RowCount = count
	}

	info.HasSchema = props["__has_schema"] == "true"

	return info, nil
}

// DeleteTable deletes a table
func (t *Tree) DeleteTable(dbPath, tableName string, force bool) error {
	tablePath := dbPath + "/" + tableName

	isTable, err := t.IsTable(tablePath)
	if err != nil {
		return err
	}
	if !isTable {
		return errors.New("path is not a table")
	}

	// Check if it has rows and force is false
	if !force {
		rowCount, _ := t.GetRowCount(tablePath)
		if rowCount > 0 {
			return errors.New("table has rows, use force=true to delete")
		}
	}

	err = t.DeleteNode(tablePath, force)
	if err != nil {
		return err
	}

	// Update database table count
	t.updateDatabaseTableCount(dbPath)

	return nil
}

// RenameTable renames a table
func (t *Tree) RenameTable(dbPath, oldName, newName string) error {
	oldPath := dbPath + "/" + oldName
	newPath := dbPath + "/" + newName

	// Verify old table exists
	isTable, err := t.IsTable(oldPath)
	if err != nil {
		return err
	}
	if !isTable {
		return errors.New("source path is not a table")
	}

	// Verify new name doesn't exist
	if exists, _ := t.IsTable(newPath); exists {
		return errors.New("table with new name already exists")
	}

	// Get all data from old table
	props, err := t.GetAllPropsWithValues(oldPath)
	if err != nil {
		return err
	}

	// Convert to map[string]interface{}
	propsInterface := make(map[string]interface{})
	for k, v := range props {
		propsInterface[k] = v
	}

	// Create new table
	err = t.CreateNodeWithProps(newPath, propsInterface)
	if err != nil {
		return err
	}

	// Update name property
	err = t.SetValue(newPath+"/__name", newName)
	if err != nil {
		return err
	}

	// TODO: Copy all child nodes (rows, schema, indices, stats)
	// Recursive copy implementation needed for complete rename

	// Delete old table
	err = t.DeleteNode(oldPath, true)
	if err != nil {
		return err
	}

	return nil
}

// ============================================================================
// Dynamic Schema Management
// ============================================================================

// InferSchemaFromRow creates a schema by inferring types from a row
func (t *Tree) InferSchemaFromRow(row models.Row) (*TableSchema, error) {
	schema := &TableSchema{
		Fields:      make(map[string]string),
		FlatFields:  []string{},
		Version:     1,
		Created:     strconv.FormatInt(time.Now().Unix(), 10),
		LastUpdated: strconv.FormatInt(time.Now().Unix(), 10),
	}

	for key, val := range row {
		// Use InferredType() from RowValue
		fieldType := val.InferredType()
		schema.Fields[key] = fieldType
		schema.FlatFields = append(schema.FlatFields, key)
	}

	return schema, nil
}

// GetTableSchema retrieves the current schema for a table
func (t *Tree) GetTableSchema(tablePath string) (*TableSchema, error) {
	schemaPath := tablePath + "/" + SchemaNode

	// Check if schema node exists
	children, err := t.GetNodesInPath(tablePath)
	if err != nil {
		return nil, err
	}

	hasSchema := false
	for _, child := range children {
		if child == SchemaNode {
			hasSchema = true
			break
		}
	}

	if !hasSchema {
		return nil, nil // No schema yet
	}

	// Get schema properties
	props, err := t.GetAllPropsWithValues(schemaPath)
	if err != nil {
		return nil, err
	}

	schema := &TableSchema{
		Created:     props["created"],
		LastUpdated: props["last_updated"],
	}

	// Parse version
	if version, err := strconv.Atoi(props["version"]); err == nil {
		schema.Version = version
	}

	// Parse fields JSON
	if fieldsJSON := props["fields"]; fieldsJSON != "" {
		err = json.Unmarshal([]byte(fieldsJSON), &schema.Fields)
		if err != nil {
			return nil, fmt.Errorf("failed to parse schema fields: %v", err)
		}
	}

	// Parse flat_fields JSON
	if flatFieldsJSON := props["flat_fields"]; flatFieldsJSON != "" {
		err = json.Unmarshal([]byte(flatFieldsJSON), &schema.FlatFields)
		if err != nil {
			return nil, fmt.Errorf("failed to parse flat fields: %v", err)
		}
	}

	return schema, nil
}

// UpdateSchemaWithNewRow updates schema when a new row introduces new fields
func (t *Tree) UpdateSchemaWithNewRow(tablePath string, row models.Row) (*TableSchema, error) {
	existingSchema, err := t.GetTableSchema(tablePath)
	if err != nil {
		return nil, err
	}

	if existingSchema == nil {
		// No schema exists, create new one
		newSchema, err := t.InferSchemaFromRow(row)
		if err != nil {
			return nil, err
		}
		// Save the new schema
		err = t.saveSchemaToStorage(tablePath, newSchema)
		if err != nil {
			return nil, err
		}
		return newSchema, nil
	}

	// Infer schema from new row
	newSchema, err := t.InferSchemaFromRow(row)
	if err != nil {
		return nil, err
	}

	// Merge schemas
	merged := t.MergeSchemas(existingSchema, newSchema)

	// Save if schema changed
	if merged.Version > existingSchema.Version {
		err = t.saveSchemaToStorage(tablePath, merged)
		if err != nil {
			return nil, err
		}
	}

	return merged, nil
}

// MergeSchemas merges two schemas, adding new fields
func (t *Tree) MergeSchemas(existing, new *TableSchema) *TableSchema {
	merged := &TableSchema{
		Fields:      make(map[string]string),
		FlatFields:  []string{},
		Version:     existing.Version,
		Created:     existing.Created,
		LastUpdated: strconv.FormatInt(time.Now().Unix(), 10),
	}

	// Copy existing fields
	for key, typ := range existing.Fields {
		merged.Fields[key] = typ
	}

	// Add new fields
	schemaChanged := false
	for key, typ := range new.Fields {
		if _, exists := merged.Fields[key]; !exists {
			merged.Fields[key] = typ
			schemaChanged = true
		} else {
			// Field exists, check if type is compatible
			// For now, keep existing type (could add type widening logic)
		}
	}

	if schemaChanged {
		merged.Version++
	}

	// Rebuild flat fields list
	for key := range merged.Fields {
		merged.FlatFields = append(merged.FlatFields, key)
	}

	return merged
}

// ValidateRowAgainstSchema validates a row against the table schema
func (t *Tree) ValidateRowAgainstSchema(row models.Row, schema *TableSchema) error {
	if schema == nil {
		return nil // No schema to validate against
	}

	// Check type compatibility for existing fields
	for key, val := range row {
		if expectedType, exists := schema.Fields[key]; exists {
			inferredType := val.InferredType()

			// Allow type compatibility (e.g., int can be stored as string)
			if !t.isTypeCompatible(inferredType, expectedType) {
				return fmt.Errorf("field '%s' type mismatch: expected %s, got %s",
					key, expectedType, inferredType)
			}
		}
		// New fields are allowed (schema will evolve)
	}

	return nil
}

// isTypeCompatible checks if two types are compatible
func (t *Tree) isTypeCompatible(actual, expected string) bool {
	// Exact match
	if actual == expected {
		return true
	}

	// Allow some flexibility
	// int can be stored as string
	if expected == "string" {
		return true
	}

	// int and float are compatible
	if (actual == "int" && expected == "float") || (actual == "float" && expected == "int") {
		return true
	}

	return false
}

// saveSchemaToStorage saves schema to __schema node
func (t *Tree) saveSchemaToStorage(tablePath string, schema *TableSchema) error {
	schemaPath := tablePath + "/" + SchemaNode

	// Create schema node
	err := t.CreatePath(schemaPath)
	if err != nil {
		return err
	}

	// Serialize fields to JSON
	fieldsJSON, err := json.Marshal(schema.Fields)
	if err != nil {
		return err
	}

	flatFieldsJSON, err := json.Marshal(schema.FlatFields)
	if err != nil {
		return err
	}

	// Store schema properties
	schemaProps := map[string]interface{}{
		"version":      strconv.Itoa(schema.Version),
		"created":      schema.Created,
		"last_updated": schema.LastUpdated,
		"fields":       string(fieldsJSON),
		"flat_fields":  string(flatFieldsJSON),
	}

	err = t.SetValues(schemaPath, schemaProps)
	if err != nil {
		return err
	}

	// Update table metadata
	t.SetValue(tablePath+"/__has_schema", "true")
	t.SetValue(tablePath+"/__lastupdated", schema.LastUpdated)

	return nil
}

// GetSchemaAsModelsSchema returns schema in models.Schema format for Object creation
func (t *Tree) GetSchemaAsModelsSchema(tablePath string) (models.Schema, error) {
	schema, err := t.GetTableSchema(tablePath)
	if err != nil {
		return nil, err
	}
	if schema == nil {
		return nil, nil
	}

	// Convert to models.Schema (map[string]string)
	return schema.Fields, nil
}

// ============================================================================
// Row Count and Stats Helpers
// ============================================================================

// GetRowCount returns the number of rows in a table
func (t *Tree) GetRowCount(tablePath string) (int, error) {
	rowCountStr, err := t.GetValue(tablePath + "/__row_count")
	if err != nil {
		return 0, err
	}

	count, err := strconv.Atoi(rowCountStr)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// IncrementRowCount increments the row count for a table
func (t *Tree) IncrementRowCount(tablePath string) error {
	count, err := t.GetRowCount(tablePath)
	if err != nil {
		count = 0
	}

	count++
	err = t.SetValue(tablePath+"/__row_count", strconv.Itoa(count))
	if err != nil {
		return err
	}

	// Update last updated
	now := strconv.FormatInt(time.Now().Unix(), 10)
	t.SetValue(tablePath+"/__lastupdated", now)

	return nil
}

// DecrementRowCount decrements the row count for a table
func (t *Tree) DecrementRowCount(tablePath string) error {
	count, err := t.GetRowCount(tablePath)
	if err != nil {
		count = 1
	}

	count--
	if count < 0 {
		count = 0
	}

	err = t.SetValue(tablePath+"/__row_count", strconv.Itoa(count))
	if err != nil {
		return err
	}

	// Update last updated
	now := strconv.FormatInt(time.Now().Unix(), 10)
	t.SetValue(tablePath+"/__lastupdated", now)

	return nil
}

// SetRowCount sets the row count for a table
func (t *Tree) SetRowCount(tablePath string, count int) error {
	return t.SetValue(tablePath+"/__row_count", strconv.Itoa(count))
}

// ============================================================================
// Helper Functions
// ============================================================================

// updateDatabaseTableCount updates the table count in a database
func (t *Tree) updateDatabaseTableCount(dbPath string) error {
	tables, err := t.ListTables(dbPath)
	if err != nil {
		return err
	}

	count := len(tables)
	err = t.SetValue(dbPath+"/__table_count", strconv.Itoa(count))
	if err != nil {
		return err
	}

	// Update last updated
	now := strconv.FormatInt(time.Now().Unix(), 10)
	t.SetValue(dbPath+"/__lastupdated", now)

	return nil
}

// GenerateRowID generates a unique row ID
func (t *Tree) GenerateRowID() string {
	// Use timestamp + random component for uniqueness
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("row_%d", timestamp)
}

// ValidateTablePath validates that a path points to a valid table
func (t *Tree) ValidateTablePath(tablePath string) error {
	isTable, err := t.IsTable(tablePath)
	if err != nil {
		return err
	}
	if !isTable {
		return errors.New("path is not a table")
	}
	return nil
}

// ValidateDatabasePath validates that a path points to a valid database
func (t *Tree) ValidateDatabasePath(dbPath string) error {
	isDB, err := t.IsDatabase(dbPath)
	if err != nil {
		return err
	}
	if !isDB {
		return errors.New("path is not a database")
	}
	return nil
}

// isSpecialNode checks if a node name is a special metadata node
func (t *Tree) isSpecialNode(nodeName string) bool {
	return strings.HasPrefix(nodeName, "__")
}

// schemaHasNewFields checks if new schema has fields not in existing schema
func (t *Tree) schemaHasNewFields(existing, new *TableSchema) bool {
	for key := range new.Fields {
		if _, exists := existing.Fields[key]; !exists {
			return true
		}
	}
	return false
}

// ============================================================================
// Row/Record CRUD Operations
// ============================================================================

// InsertRow inserts a new row into a table with automatic schema evolution
// Returns the generated row ID
func (t *Tree) InsertRow(tablePath string, row models.Row) (string, error) {
	// Validate table path
	if err := t.ValidateTablePath(tablePath); err != nil {
		return "", err
	}

	// Generate unique row ID
	rowID := t.GenerateRowID()
	rowPath := tablePath + "/" + rowID

	// Update schema with new row (adds new fields if any)
	_, err := t.UpdateSchemaWithNewRow(tablePath, row)
	if err != nil {
		return "", fmt.Errorf("failed to update schema: %v", err)
	}

	// Validate row against schema
	schema, _ := t.GetTableSchema(tablePath)
	if err := t.ValidateRowAgainstSchema(row, schema); err != nil {
		return "", fmt.Errorf("row validation failed: %v", err)
	}

	// Convert Row to map[string]interface{} for storage
	rowData := make(map[string]interface{})
	for key, val := range row {
		rowData[key] = val.Val()
	}

	// Store the row
	err = t.CreateNodeWithProps(rowPath, rowData)
	if err != nil {
		return "", fmt.Errorf("failed to store row: %v", err)
	}

	// Increment row count
	if err := t.IncrementRowCount(tablePath); err != nil {
		return "", fmt.Errorf("failed to update row count: %v", err)
	}

	// Add to all applicable indexes
	if err := t.addToIndex(tablePath, rowID, row); err != nil {
		// Rollback: remove the row
		t.DeleteNode(rowPath, true)
		t.DecrementRowCount(tablePath)
		return "", fmt.Errorf("failed to update indexes: %v", err)
	}

	return rowID, nil
}

// InsertRowWithID inserts a row with a specific ID (useful for imports or when ID is predetermined)
func (t *Tree) InsertRowWithID(tablePath, rowID string, row models.Row) error {
	// Validate table path
	if err := t.ValidateTablePath(tablePath); err != nil {
		return err
	}

	rowPath := tablePath + "/" + rowID

	// Check if row already exists
	children, err := t.GetNodesInPath(tablePath)
	if err == nil {
		for _, child := range children {
			if child == rowID {
				return fmt.Errorf("row with ID %s already exists", rowID)
			}
		}
	}

	// Update schema with new row
	_, err = t.UpdateSchemaWithNewRow(tablePath, row)
	if err != nil {
		return fmt.Errorf("failed to update schema: %v", err)
	}

	// Validate row against schema
	schema, _ := t.GetTableSchema(tablePath)
	if err := t.ValidateRowAgainstSchema(row, schema); err != nil {
		return fmt.Errorf("row validation failed: %v", err)
	}

	// Convert Row to map[string]interface{}
	rowData := make(map[string]interface{})
	for key, val := range row {
		rowData[key] = val.Val()
	}

	// Store the row
	err = t.CreateNodeWithProps(rowPath, rowData)
	if err != nil {
		return fmt.Errorf("failed to store row: %v", err)
	}

	// Increment row count
	if err := t.IncrementRowCount(tablePath); err != nil {
		return fmt.Errorf("failed to update row count: %v", err)
	}

	// Add to all applicable indexes
	if err := t.addToIndex(tablePath, rowID, row); err != nil {
		// Rollback: remove the row
		t.DeleteNode(rowPath, true)
		t.DecrementRowCount(tablePath)
		return fmt.Errorf("failed to update indexes: %v", err)
	}

	return nil
}

// GetRow retrieves a row by ID
func (t *Tree) GetRow(tablePath, rowID string) (models.Row, error) {
	// Validate table path
	if err := t.ValidateTablePath(tablePath); err != nil {
		return nil, err
	}

	rowPath := tablePath + "/" + rowID

	// Get all properties of the row
	props, err := t.GetAllPropsWithValues(rowPath)
	if err != nil {
		return nil, fmt.Errorf("row not found: %v", err)
	}

	// Get table schema for type information
	schema, _ := t.GetTableSchema(tablePath)

	// Convert to Row
	row := make(models.Row)
	for key, value := range props {
		rowVal := models.NewValue(value)

		// Set schema type if available
		if schema != nil {
			if schemaType, exists := schema.Fields[key]; exists {
				rowVal.SetSchemaType(schemaType)
			}
		}

		row[key] = rowVal
	}

	return row, nil
}

// UpdateRow updates an existing row (replaces all fields)
func (t *Tree) UpdateRow(tablePath, rowID string, row models.Row) error {
	// Validate table path
	if err := t.ValidateTablePath(tablePath); err != nil {
		return err
	}

	rowPath := tablePath + "/" + rowID

	// Get old row for index maintenance
	oldRow, err := t.GetRow(tablePath, rowID)
	if err != nil {
		return fmt.Errorf("row not found: %s", rowID)
	}

	// Update schema with new row (in case new fields are added)
	_, err = t.UpdateSchemaWithNewRow(tablePath, row)
	if err != nil {
		return fmt.Errorf("failed to update schema: %v", err)
	}

	// Validate row against schema
	schema, _ := t.GetTableSchema(tablePath)
	if err := t.ValidateRowAgainstSchema(row, schema); err != nil {
		return fmt.Errorf("row validation failed: %v", err)
	}

	// Convert Row to map[string]interface{}
	rowData := make(map[string]interface{})
	for key, val := range row {
		rowData[key] = val.Val()
	}

	// Update all properties (this will overwrite existing values)
	err = t.SetValues(rowPath, rowData)
	if err != nil {
		return fmt.Errorf("failed to update row: %v", err)
	}

	// Update indexes
	if err := t.updateIndex(tablePath, rowID, oldRow, row); err != nil {
		return fmt.Errorf("failed to update indexes: %v", err)
	}

	// Update last updated timestamp
	now := strconv.FormatInt(time.Now().Unix(), 10)
	t.SetValue(tablePath+"/__lastupdated", now)

	return nil
}

// UpdateRowFields updates specific fields in a row (partial update)
func (t *Tree) UpdateRowFields(tablePath, rowID string, fields map[string]interface{}) error {
	// Validate table path
	if err := t.ValidateTablePath(tablePath); err != nil {
		return err
	}

	rowPath := tablePath + "/" + rowID

	// Get old row for index maintenance
	oldRow, err := t.GetRow(tablePath, rowID)
	if err != nil {
		return fmt.Errorf("row not found: %s", rowID)
	}

	// Create a Row from the fields for schema validation
	row := make(models.Row)
	for key, val := range fields {
		row[key] = models.NewValue(val)
	}

	// Update schema if needed
	_, err = t.UpdateSchemaWithNewRow(tablePath, row)
	if err != nil {
		return fmt.Errorf("failed to update schema: %v", err)
	}

	// Update the fields
	err = t.SetValues(rowPath, fields)
	if err != nil {
		return fmt.Errorf("failed to update fields: %v", err)
	}

	// Get updated row for index maintenance
	newRow, err := t.GetRow(tablePath, rowID)
	if err != nil {
		return fmt.Errorf("failed to retrieve updated row: %v", err)
	}

	// Update indexes
	if err := t.updateIndex(tablePath, rowID, oldRow, newRow); err != nil {
		return fmt.Errorf("failed to update indexes: %v", err)
	}

	// Update last updated timestamp
	now := strconv.FormatInt(time.Now().Unix(), 10)
	t.SetValue(tablePath+"/__lastupdated", now)

	return nil
}

// DeleteRow deletes a row from a table
func (t *Tree) DeleteRow(tablePath, rowID string) error {
	// Validate table path
	if err := t.ValidateTablePath(tablePath); err != nil {
		return err
	}

	rowPath := tablePath + "/" + rowID

	// Get row data before deletion for index maintenance
	row, err := t.GetRow(tablePath, rowID)
	if err != nil {
		return fmt.Errorf("row not found: %s", rowID)
	}

	// Remove from indexes first
	if err := t.removeFromIndex(tablePath, rowID, row); err != nil {
		return fmt.Errorf("failed to update indexes: %v", err)
	}

	// Delete the row node
	err = t.DeleteNode(rowPath, true)
	if err != nil {
		return fmt.Errorf("failed to delete row: %v", err)
	}

	// Decrement row count
	if err := t.DecrementRowCount(tablePath); err != nil {
		return fmt.Errorf("failed to update row count: %v", err)
	}

	return nil
}

// ListRows returns all row IDs in a table
func (t *Tree) ListRows(tablePath string) ([]string, error) {
	// Validate table path
	if err := t.ValidateTablePath(tablePath); err != nil {
		return nil, err
	}

	children, err := t.GetNodesInPath(tablePath)
	if err != nil {
		return nil, err
	}

	// Filter out special nodes (metadata)
	var rows []string
	for _, child := range children {
		if !t.isSpecialNode(child) {
			rows = append(rows, child)
		}
	}

	return rows, nil
}

// GetAllRows retrieves all rows in a table
func (t *Tree) GetAllRows(tablePath string) (map[string]models.Row, error) {
	// Validate table path
	if err := t.ValidateTablePath(tablePath); err != nil {
		return nil, err
	}

	// Get all row IDs
	rowIDs, err := t.ListRows(tablePath)
	if err != nil {
		return nil, err
	}

	// Get table schema once for all rows
	schema, _ := t.GetTableSchema(tablePath)

	// Retrieve each row
	result := make(map[string]models.Row)
	for _, rowID := range rowIDs {
		rowPath := tablePath + "/" + rowID

		props, err := t.GetAllPropsWithValues(rowPath)
		if err != nil {
			continue // Skip rows that can't be read
		}

		// Convert to Row
		row := make(models.Row)
		for key, value := range props {
			rowVal := models.NewValue(value)

			// Set schema type if available
			if schema != nil {
				if schemaType, exists := schema.Fields[key]; exists {
					rowVal.SetSchemaType(schemaType)
				}
			}

			row[key] = rowVal
		}

		result[rowID] = row
	}

	return result, nil
}

// ScanRows scans all rows in a table and calls the callback for each row
// This is more memory-efficient than GetAllRows for large tables
func (t *Tree) ScanRows(tablePath string, callback func(rowID string, row models.Row) error) error {
	// Validate table path
	if err := t.ValidateTablePath(tablePath); err != nil {
		return err
	}

	// Get table schema once
	schema, _ := t.GetTableSchema(tablePath)

	// Use ScanNodes to iterate through child nodes
	return t.ScanNodes(tablePath, func(nodeInfo NodeInfo) error {
		// Skip special metadata nodes
		if t.isSpecialNode(nodeInfo.Name) {
			return nil
		}

		// Convert props to Row
		row := make(models.Row)
		for key, value := range nodeInfo.Props {
			rowVal := models.NewValue(value)

			// Set schema type if available
			if schema != nil {
				if schemaType, exists := schema.Fields[key]; exists {
					rowVal.SetSchemaType(schemaType)
				}
			}

			row[key] = rowVal
		}

		// Call the callback with row ID and row data
		return callback(nodeInfo.Name, row)
	})
}

// CountRows returns the number of rows in a table (using the metadata counter)
func (t *Tree) CountRows(tablePath string) (int, error) {
	return t.GetRowCount(tablePath)
}

// RowExists checks if a row with the given ID exists in a table
func (t *Tree) RowExists(tablePath, rowID string) (bool, error) {
	rowPath := tablePath + "/" + rowID

	// Try to get nodes in the row path
	_, err := t.GetNodesInPath(rowPath)
	if err != nil {
		// If error getting nodes, row doesn't exist or there's another issue
		return false, nil
	}

	return true, nil
}

// ============================================================================
// Index Management
// ============================================================================

// Index constants
const (
	IndexTypeSingle    = "single"
	IndexTypeComposite = "composite"
	IndexEntriesNode   = "_entries" // Bucket for index entries
)

// IndexInfo represents index metadata
type IndexInfo struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`        // "single" or "composite"
	Fields     []string `json:"fields"`      // Indexed field names
	Unique     bool     `json:"unique"`      // Unique constraint
	Created    string   `json:"created"`     // Creation timestamp
	Updated    string   `json:"updated"`     // Last update timestamp
	EntryCount int      `json:"entry_count"` // Number of index entries
}

// CreateIndex creates an index on one or more fields
func (t *Tree) CreateIndex(tablePath, indexName string, fields []string, unique bool) error {
	// Validate table
	if err := t.ValidateTablePath(tablePath); err != nil {
		return err
	}

	if len(fields) == 0 {
		return errors.New("at least one field is required for index")
	}

	// Validate index name
	if indexName == "" || strings.HasPrefix(indexName, "__") {
		return errors.New("invalid index name")
	}

	indexPath := tablePath + "/" + IndicesNode + "/" + indexName

	// Check if index already exists
	children, err := t.GetNodesInPath(tablePath + "/" + IndicesNode)
	if err == nil {
		for _, child := range children {
			if child == indexName {
				return fmt.Errorf("index %s already exists", indexName)
			}
		}
	}

	// Determine index type
	indexType := IndexTypeSingle
	if len(fields) > 1 {
		indexType = IndexTypeComposite
	}

	// Create index metadata
	now := strconv.FormatInt(time.Now().Unix(), 10)
	metadata := map[string]interface{}{
		"__type":        indexType,
		"__unique":      strconv.FormatBool(unique),
		"__created":     now,
		"__updated":     now,
		"__entry_count": "0",
	}

	// Store fields as individual properties
	for i, field := range fields {
		metadata[fmt.Sprintf("__field_%d", i)] = field
	}

	// Create index node with metadata
	err = t.CreateNodeWithProps(indexPath, metadata)
	if err != nil {
		return fmt.Errorf("failed to create index metadata: %v", err)
	}

	// Create entries bucket
	err = t.CreatePath(indexPath + "/" + IndexEntriesNode)
	if err != nil {
		return fmt.Errorf("failed to create index entries: %v", err)
	}

	// Build index from existing data
	err = t.buildIndex(tablePath, indexPath, fields, unique)
	if err != nil {
		// Rollback: delete index on failure
		t.DeleteNode(indexPath, true)
		return fmt.Errorf("failed to build index: %v", err)
	}

	return nil
}

// buildIndex scans table and populates index entries
func (t *Tree) buildIndex(tablePath, indexPath string, fields []string, unique bool) error {
	entriesPath := indexPath + "/" + IndexEntriesNode

	// First pass: collect all index entries in memory
	indexEntries := make(map[string][]string) // indexKey -> []rowIDs

	err := t.ScanRows(tablePath, func(rowID string, row models.Row) error {
		// Extract index key from row
		indexKey, err := t.buildIndexKey(row, fields)
		if err != nil {
			return err
		}

		// Skip if any field is null
		if indexKey == "" {
			return nil
		}

		// For unique indexes, check for duplicates
		if unique {
			if _, exists := indexEntries[indexKey]; exists {
				return fmt.Errorf("duplicate value for unique index: %s", indexKey)
			}
		}

		// Add to in-memory map
		indexEntries[indexKey] = append(indexEntries[indexKey], rowID)
		return nil
	})

	if err != nil {
		return err
	}

	// Second pass: write all index entries to storage
	entryCount := 0
	for indexKey, rowIDs := range indexEntries {
		keyPath := entriesPath + "/" + indexKey

		// Create bucket for this key
		err := t.CreatePath(keyPath)
		if err != nil {
			return err
		}

		// Add all row IDs for this key
		for _, rowID := range rowIDs {
			err = t.SetValue(keyPath+"/"+rowID, "")
			if err != nil {
				return err
			}
			entryCount++
		}
	}

	// Update entry count
	t.SetValue(indexPath+"/__entry_count", strconv.Itoa(entryCount))
	return nil
}

// buildIndexKey creates an index key from row data
// For single field: returns field value as string
// For composite: returns "field1_val|field2_val|field3_val"
func (t *Tree) buildIndexKey(row models.Row, fields []string) (string, error) {
	var keyParts []string

	for _, field := range fields {
		val, exists := row[field]
		if !exists || val.IsNull() {
			// Null values are not indexed
			return "", nil
		}

		// Convert value to string for key
		keyParts = append(keyParts, val.AsString())
	}

	// For single field, return value directly
	if len(keyParts) == 1 {
		return keyParts[0], nil
	}

	// For composite, join with separator
	return strings.Join(keyParts, "|"), nil
}

// DropIndex removes an index
func (t *Tree) DropIndex(tablePath, indexName string) error {
	if err := t.ValidateTablePath(tablePath); err != nil {
		return err
	}

	indexPath := tablePath + "/" + IndicesNode + "/" + indexName

	// Check if index exists
	_, err := t.GetNodesInPath(indexPath)
	if err != nil {
		return fmt.Errorf("index %s does not exist", indexName)
	}

	// Delete index node and all entries
	return t.DeleteNode(indexPath, true)
}

// ListIndexes returns all index names in a table
func (t *Tree) ListIndexes(tablePath string) ([]string, error) {
	if err := t.ValidateTablePath(tablePath); err != nil {
		return nil, err
	}

	indicesPath := tablePath + "/" + IndicesNode

	children, err := t.GetNodesInPath(indicesPath)
	if err != nil {
		// No indices node means no indexes
		return []string{}, nil
	}

	// Filter out special nodes
	var indexes []string
	for _, child := range children {
		if !t.isSpecialNode(child) {
			indexes = append(indexes, child)
		}
	}

	return indexes, nil
}

// GetIndexInfo retrieves index metadata
func (t *Tree) GetIndexInfo(tablePath, indexName string) (*IndexInfo, error) {
	if err := t.ValidateTablePath(tablePath); err != nil {
		return nil, err
	}

	indexPath := tablePath + "/" + IndicesNode + "/" + indexName

	// Get index metadata
	props, err := t.GetAllPropsWithValues(indexPath)
	if err != nil {
		return nil, fmt.Errorf("index %s not found", indexName)
	}

	info := &IndexInfo{
		Name: indexName,
		Type: props["__type"],
	}

	// Parse unique flag
	if uniqueStr := props["__unique"]; uniqueStr != "" {
		info.Unique = uniqueStr == "true"
	}

	// Parse timestamps
	info.Created = props["__created"]
	info.Updated = props["__updated"]

	// Parse entry count
	if countStr := props["__entry_count"]; countStr != "" {
		if count, err := strconv.Atoi(countStr); err == nil {
			info.EntryCount = count
		}
	}

	// Extract fields from __field_N properties
	i := 0
	for {
		fieldKey := fmt.Sprintf("__field_%d", i)
		if field, exists := props[fieldKey]; exists {
			info.Fields = append(info.Fields, field)
			i++
		} else {
			break
		}
	}

	return info, nil
}

// RebuildIndex drops and recreates an index
func (t *Tree) RebuildIndex(tablePath, indexName string) error {
	// Get index info before dropping
	info, err := t.GetIndexInfo(tablePath, indexName)
	if err != nil {
		return err
	}

	// Drop existing index
	err = t.DropIndex(tablePath, indexName)
	if err != nil {
		return err
	}

	// Recreate index
	return t.CreateIndex(tablePath, indexName, info.Fields, info.Unique)
}

// addToIndex adds a row to all applicable indexes
func (t *Tree) addToIndex(tablePath, rowID string, row models.Row) error {
	indexes, err := t.ListIndexes(tablePath)
	if err != nil || len(indexes) == 0 {
		return nil // No indexes, nothing to do
	}

	for _, indexName := range indexes {
		indexPath := tablePath + "/" + IndicesNode + "/" + indexName
		info, err := t.GetIndexInfo(tablePath, indexName)
		if err != nil {
			continue
		}

		// Build index key for this row
		indexKey, err := t.buildIndexKey(row, info.Fields)
		if err != nil || indexKey == "" {
			continue // Skip if key can't be built or is null
		}

		entriesPath := indexPath + "/" + IndexEntriesNode
		keyPath := entriesPath + "/" + indexKey

		// For unique indexes, check for duplicates
		if info.Unique {
			props, err := t.GetAllPropsWithValues(keyPath)
			if err == nil && len(props) > 0 {
				return fmt.Errorf("duplicate value for unique index %s: %s", indexName, indexKey)
			}
		}

		// Create bucket for this key if needed
		t.CreatePath(keyPath)

		// Add row ID to this key
		t.SetValue(keyPath+"/"+rowID, "")

		// Increment entry count
		if countStr, _ := t.GetValue(indexPath + "/__entry_count"); countStr != "" {
			if count, err := strconv.Atoi(countStr); err == nil {
				t.SetValue(indexPath+"/__entry_count", strconv.Itoa(count+1))
			}
		}
	}

	return nil
}

// removeFromIndex removes a row from all applicable indexes
func (t *Tree) removeFromIndex(tablePath, rowID string, row models.Row) error {
	indexes, err := t.ListIndexes(tablePath)
	if err != nil || len(indexes) == 0 {
		return nil
	}

	for _, indexName := range indexes {
		indexPath := tablePath + "/" + IndicesNode + "/" + indexName
		info, err := t.GetIndexInfo(tablePath, indexName)
		if err != nil {
			continue
		}

		// Build index key for this row
		indexKey, err := t.buildIndexKey(row, info.Fields)
		if err != nil || indexKey == "" {
			continue
		}

		entriesPath := indexPath + "/" + IndexEntriesNode
		keyPath := entriesPath + "/" + indexKey

		// Remove row ID from this key
		t.DeleteValue(keyPath, rowID)

		// If this was the last row for this key, remove the key bucket
		props, err := t.GetAllPropsWithValues(keyPath)
		if err == nil && len(props) == 0 {
			t.DeleteNode(keyPath, true)
		}

		// Decrement entry count
		if countStr, _ := t.GetValue(indexPath + "/__entry_count"); countStr != "" {
			if count, err := strconv.Atoi(countStr); err == nil && count > 0 {
				t.SetValue(indexPath+"/__entry_count", strconv.Itoa(count-1))
			}
		}
	}

	return nil
}

// updateIndex updates indexes when a row is modified
func (t *Tree) updateIndex(tablePath, rowID string, oldRow, newRow models.Row) error {
	// Remove old index entries
	if oldRow != nil {
		if err := t.removeFromIndex(tablePath, rowID, oldRow); err != nil {
			return err
		}
	}

	// Add new index entries
	return t.addToIndex(tablePath, rowID, newRow)
}

// lookupIndex finds row IDs matching an index key
func (t *Tree) lookupIndex(tablePath, indexName, indexKey string) ([]string, error) {
	indexPath := tablePath + "/" + IndicesNode + "/" + indexName
	entriesPath := indexPath + "/" + IndexEntriesNode
	keyPath := entriesPath + "/" + indexKey

	// Get all row IDs for this key (stored as properties)
	props, err := t.GetAllPropsWithValues(keyPath)
	if err != nil {
		return []string{}, nil // Key not found, return empty
	}

	// Extract row IDs (property keys)
	var rowIDs []string
	for rowID := range props {
		rowIDs = append(rowIDs, rowID)
	}

	return rowIDs, nil
}

// lookupIndexRange finds row IDs in a range (for composite or range queries)
func (t *Tree) lookupIndexRange(tablePath, indexName, startKey, endKey string) ([]string, error) {
	indexPath := tablePath + "/" + IndicesNode + "/" + indexName
	entriesPath := indexPath + "/" + IndexEntriesNode

	// Get all keys in the index (child buckets of _entries)
	allKeys, err := t.GetNodesInPath(entriesPath)
	if err != nil {
		return []string{}, nil
	}

	var result []string
	for _, key := range allKeys {
		// Check if key is in range
		if (startKey == "" || key >= startKey) && (endKey == "" || key <= endKey) {
			keyPath := entriesPath + "/" + key
			// Get row IDs stored as properties
			props, err := t.GetAllPropsWithValues(keyPath)
			if err == nil {
				for rowID := range props {
					result = append(result, rowID)
				}
			}
		}
	}

	return result, nil
}

// ============================================================================
// Cursor-Based Query Execution (Memory Efficient)
// ============================================================================

// RowCursor represents a lightweight reference to a row with only necessary fields loaded.
// This minimizes memory usage and database lock time during query execution.
type RowCursor struct {
	RowID  string      // The row identifier
	Fields models.Row  // Only the fields needed for query evaluation (partial row)
}

// QueryCursor provides efficient iteration over query results.
// It yields row IDs and allows lazy loading of fields as needed.
// Supports LIMIT, SKIP (offset), and maintains position for pagination.
type QueryCursor struct {
	tablePath string
	rowIDs    []string
	position  int
	tree      *Tree
	limit     int  // Maximum number of rows to return (0 = no limit)
	skip      int  // Number of rows to skip from start (offset)
}

// NewQueryCursor creates a cursor for iterating over row IDs
func (t *Tree) NewQueryCursor(tablePath string, rowIDs []string) *QueryCursor {
	return &QueryCursor{
		tablePath: tablePath,
		rowIDs:    rowIDs,
		position:  0,
		tree:      t,
		limit:     0,
		skip:      0,
	}
}

// WithLimit sets the maximum number of rows to return
func (c *QueryCursor) WithLimit(limit int) *QueryCursor {
	c.limit = limit
	return c
}

// WithSkip sets the number of rows to skip (offset for pagination)
func (c *QueryCursor) WithSkip(skip int) *QueryCursor {
	c.skip = skip
	c.position = skip // Start from the skip position
	return c
}

// WithPagination is a convenience method for LIMIT + SKIP
// Example: cursor.WithPagination(page=2, pageSize=10) -> Skip(10), Limit(10)
func (c *QueryCursor) WithPagination(page, pageSize int) *QueryCursor {
	if page < 1 {
		page = 1
	}
	c.skip = (page - 1) * pageSize
	c.limit = pageSize
	c.position = c.skip
	return c
}

// Next advances the cursor and returns true if there's a next row
// Respects LIMIT constraint if set
func (c *QueryCursor) Next() bool {
	c.position++

	// Check if we've hit the limit
	if c.limit > 0 {
		rowsReturned := c.position - c.skip
		if rowsReturned >= c.limit {
			return false
		}
	}

	return c.position < len(c.rowIDs)
}

// HasNext returns true if there are more rows
// Respects LIMIT constraint if set
func (c *QueryCursor) HasNext() bool {
	nextPos := c.position + 1

	// Check if we've hit the limit
	if c.limit > 0 {
		rowsReturned := nextPos - c.skip
		if rowsReturned >= c.limit {
			return false
		}
	}

	return nextPos < len(c.rowIDs)
}

// CurrentID returns the current row ID
func (c *QueryCursor) CurrentID() string {
	if c.position >= len(c.rowIDs) {
		return ""
	}
	return c.rowIDs[c.position]
}

// LoadFields loads only specific fields for the current row
func (c *QueryCursor) LoadFields(fields []string) (models.Row, error) {
	if c.position >= len(c.rowIDs) {
		return nil, errors.New("cursor out of bounds")
	}
	return c.tree.GetRowFields(c.tablePath, c.rowIDs[c.position], fields)
}

// LoadFullRow loads the complete row (use sparingly)
func (c *QueryCursor) LoadFullRow() (models.Row, error) {
	if c.position >= len(c.rowIDs) {
		return nil, errors.New("cursor out of bounds")
	}
	return c.tree.GetRow(c.tablePath, c.rowIDs[c.position])
}

// Count returns the total number of rows in the cursor
func (c *QueryCursor) Count() int {
	return len(c.rowIDs)
}

// Reset resets the cursor to the beginning
func (c *QueryCursor) Reset() {
	c.position = 0
}

// GetRowFields retrieves only specific fields from a row.
// This is the key function for memory-efficient querying.
// Example: GetRowFields(tablePath, rowID, []string{"name", "age"})
func (t *Tree) GetRowFields(tablePath, rowID string, fields []string) (models.Row, error) {
	// Validate table path
	if err := t.ValidateTablePath(tablePath); err != nil {
		return nil, err
	}

	rowPath := tablePath + "/" + rowID

	// Get table schema for type information
	schema, _ := t.GetTableSchema(tablePath)

	// Build partial row with only requested fields
	row := make(models.Row)

	for _, fieldName := range fields {
		// Get specific field value
		value, err := t.GetValue(rowPath + "/" + fieldName)
		if err != nil {
			// Field doesn't exist - skip or set as null depending on requirement
			row[fieldName] = models.NewValue(nil)
			continue
		}

		rowVal := models.NewValue(value)

		// Set schema type if available
		if schema != nil {
			if schemaType, exists := schema.Fields[fieldName]; exists {
				rowVal.SetSchemaType(schemaType)
			}
		}

		row[fieldName] = rowVal
	}

	return row, nil
}

// ScanRowsWithFields scans rows but only loads specified fields.
// This is memory-efficient for query execution where you only need certain fields.
// The callback receives row ID and a partial row with only the requested fields.
func (t *Tree) ScanRowsWithFields(tablePath string, fields []string, callback func(rowID string, row models.Row) error) error {
	// Validate table path
	if err := t.ValidateTablePath(tablePath); err != nil {
		return err
	}

	// Get table schema once
	schema, _ := t.GetTableSchema(tablePath)

	// Use ScanNodes to iterate through child nodes
	return t.ScanNodes(tablePath, func(nodeInfo NodeInfo) error {
		// Skip special metadata nodes
		if t.isSpecialNode(nodeInfo.Name) {
			return nil
		}

		// Build partial row with only requested fields
		row := make(models.Row)

		for _, fieldName := range fields {
			// Check if field exists in node properties
			if value, exists := nodeInfo.Props[fieldName]; exists {
				rowVal := models.NewValue(value)

				// Set schema type if available
				if schema != nil {
					if schemaType, exists := schema.Fields[fieldName]; exists {
						rowVal.SetSchemaType(schemaType)
					}
				}

				row[fieldName] = rowVal
			} else {
				// Field not present - set as null
				row[fieldName] = models.NewValue(nil)
			}
		}

		// Call the callback with row ID and partial row
		return callback(nodeInfo.Name, row)
	})
}

// ScanRowIDsOnly scans and returns only row IDs without loading any field data.
// This is the most memory-efficient scan - useful for counting or initial filtering.
func (t *Tree) ScanRowIDsOnly(tablePath string, callback func(rowID string) error) error {
	// Validate table path
	if err := t.ValidateTablePath(tablePath); err != nil {
		return err
	}

	// Use ScanNodes but ignore all properties
	return t.ScanNodes(tablePath, func(nodeInfo NodeInfo) error {
		// Skip special metadata nodes
		if t.isSpecialNode(nodeInfo.Name) {
			return nil
		}

		// Call callback with just the row ID
		return callback(nodeInfo.Name)
	})
}

// GetRowIDsOnly returns all row IDs in a table without loading any data.
// Memory-efficient alternative to ListRows when you only need IDs.
func (t *Tree) GetRowIDsOnly(tablePath string) ([]string, error) {
	// Validate table path
	if err := t.ValidateTablePath(tablePath); err != nil {
		return nil, err
	}

	var rowIDs []string
	err := t.ScanRowIDsOnly(tablePath, func(rowID string) error {
		rowIDs = append(rowIDs, rowID)
		return nil
	})

	return rowIDs, err
}

// RowCursorBatch represents a batch of row cursors for efficient processing
type RowCursorBatch struct {
	Cursors   []*RowCursor
	BatchSize int
}

// CreateRowCursorsWithFields creates lightweight cursors for a set of row IDs.
// Only loads the specified fields, keeping memory footprint minimal.
func (t *Tree) CreateRowCursorsWithFields(tablePath string, rowIDs []string, fields []string) ([]*RowCursor, error) {
	cursors := make([]*RowCursor, 0, len(rowIDs))

	for _, rowID := range rowIDs {
		partialRow, err := t.GetRowFields(tablePath, rowID, fields)
		if err != nil {
			// Skip rows that can't be loaded
			continue
		}

		cursors = append(cursors, &RowCursor{
			RowID:  rowID,
			Fields: partialRow,
		})
	}

	return cursors, nil
}

// BatchRowCursors splits row cursors into batches for memory-efficient processing.
// This allows processing large result sets without loading everything into memory.
func BatchRowCursors(cursors []*RowCursor, batchSize int) []RowCursorBatch {
	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	var batches []RowCursorBatch

	for i := 0; i < len(cursors); i += batchSize {
		end := i + batchSize
		if end > len(cursors) {
			end = len(cursors)
		}

		batches = append(batches, RowCursorBatch{
			Cursors:   cursors[i:end],
			BatchSize: end - i,
		})
	}

	return batches
}

// ============================================================================
// Sorting Support for Cursors
// ============================================================================

// SortDirection represents the sort order
type SortDirection int

const (
	SortAsc  SortDirection = 1
	SortDesc SortDirection = -1
)

// SortField represents a field to sort by and its direction
type SortField struct {
	FieldName string
	Direction SortDirection
}

// SortRowsByFields sorts a list of row IDs by specified fields.
// This loads only the sort fields (memory efficient) and sorts in-place.
// Use this before creating a QueryCursor for sorted results.
func (t *Tree) SortRowsByFields(tablePath string, rowIDs []string, sortFields []SortField) ([]string, error) {
	if len(sortFields) == 0 {
		return rowIDs, nil // No sorting needed
	}

	// Extract field names for loading
	fieldNames := make([]string, len(sortFields))
	for i, sf := range sortFields {
		fieldNames[i] = sf.FieldName
	}

	// Create a structure to hold row ID + sort field values
	type sortableRow struct {
		rowID  string
		values models.Row
	}

	// Load sort fields for all rows
	sortableRows := make([]sortableRow, 0, len(rowIDs))
	for _, rowID := range rowIDs {
		values, err := t.GetRowFields(tablePath, rowID, fieldNames)
		if err != nil {
			continue // Skip rows that can't be loaded
		}

		sortableRows = append(sortableRows, sortableRow{
			rowID:  rowID,
			values: values,
		})
	}

	// Sort using custom comparison
	for i := 0; i < len(sortableRows); i++ {
		for j := i + 1; j < len(sortableRows); j++ {
			shouldSwap := false

			// Compare by each sort field in order
			for _, sortField := range sortFields {
				val1 := sortableRows[i].values[sortField.FieldName]
				val2 := sortableRows[j].values[sortField.FieldName]

				// Dereference pointers for comparison
				cmp := compareValues(*val1, *val2)

				if cmp == 0 {
					continue // Equal, check next field
				}

				// Apply sort direction
				if sortField.Direction == SortDesc {
					cmp = -cmp
				}

				if cmp > 0 {
					shouldSwap = true
				}

				break // Decision made, no need to check more fields
			}

			if shouldSwap {
				sortableRows[i], sortableRows[j] = sortableRows[j], sortableRows[i]
			}
		}
	}

	// Extract sorted row IDs
	sortedIDs := make([]string, len(sortableRows))
	for i, sr := range sortableRows {
		sortedIDs[i] = sr.rowID
	}

	return sortedIDs, nil
}

// compareValues compares two RowValue objects
// Returns: -1 if v1 < v2, 0 if equal, 1 if v1 > v2
func compareValues(v1, v2 models.RowValue) int {
	// Handle null values
	if v1.IsNull() && v2.IsNull() {
		return 0
	}
	if v1.IsNull() {
		return -1 // Nulls sort first
	}
	if v2.IsNull() {
		return 1
	}

	// Compare as strings (basic comparison)
	str1 := v1.AsString()
	str2 := v2.AsString()

	if str1 < str2 {
		return -1
	}
	if str1 > str2 {
		return 1
	}
	return 0
}

// NewSortedQueryCursor creates a cursor with pre-sorted row IDs.
// This is more efficient than sorting after cursor creation.
func (t *Tree) NewSortedQueryCursor(tablePath string, rowIDs []string, sortFields []SortField) (*QueryCursor, error) {
	sortedIDs, err := t.SortRowsByFields(tablePath, rowIDs, sortFields)
	if err != nil {
		return nil, err
	}

	return t.NewQueryCursor(tablePath, sortedIDs), nil
}
