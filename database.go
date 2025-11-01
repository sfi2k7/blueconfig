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

	// Check if row exists
	_, err := t.GetNodesInPath(rowPath)
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

	// Check if row exists
	_, err := t.GetNodesInPath(rowPath)
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

	// Delete the row node
	err := t.DeleteNode(rowPath, true)
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
