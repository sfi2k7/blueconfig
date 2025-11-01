# BlueConfig Database Implementation - Summary

## Overview

BlueConfig is a hierarchical tree-based data storage system built on **BoltDB** (bbolt) that provides database-like functionality with a path-based API. The system implements a schema-less, document-oriented database with dynamic schema inference and evolution.

---

## Architecture

### Core Components

1. **Storage Engine**: BoltDB (bbolt) - embedded key-value store
2. **Data Model**: Hierarchical tree structure with path-based addressing
3. **Organization**: Three-tier hierarchy
   - **Databases**: Top-level containers (marked with `__type: "database"`)
   - **Tables**: Data containers within databases (marked with `__type: "table"`)
   - **Rows**: Individual records stored as nested buckets

### File Structure

| File | Lines | Purpose |
|------|-------|---------|
| `database.go` | 786 | Database and table operations, schema management |
| `database_test.go` | 1,473 | Comprehensive test suite (50 test functions) |
| `db.go` | 827 | Core tree operations, BoltDB wrapper, HTTP API |
| `models/object.go` | 600+ | Row modeling, querying, matching |
| `models/value.go` | 700+ | Type system, value handling |

---

## Key Features

### 1. Database Operations

**Create Database**
```go
tree.CreateDatabase("root/mydb", metadata)
```
- Automatically sets metadata: `__type`, `__created`, `__lastupdated`, `__table_count`, `__view_count`
- Supports custom metadata fields
- Path-based addressing

**List/Query Databases**
```go
databases, err := tree.ListDatabases("root")
info, err := tree.GetDatabaseInfo("root/mydb")
```

**Delete Database**
```go
tree.DeleteDatabase("root/mydb", force)
```
- Safety check: prevents deletion if tables exist (unless `force=true`)

### 2. Table Operations

**Create Table**
```go
tree.CreateTable("root/mydb", "users")
```
- Validates parent is a database
- Initializes metadata: `__type`, `__name`, `__row_count`, `__has_schema`
- Updates parent database's table count

**Table Metadata**
- Row count tracking (`IncrementRowCount`, `DecrementRowCount`)
- Schema tracking (`__has_schema` flag)
- Timestamps (`__created`, `__lastupdated`)

**Table Management**
```go
tables, err := tree.ListTables("root/mydb")
info, err := tree.GetTableInfo("root/mydb/users")
tree.DeleteTable("root/mydb", "users", force)
tree.RenameTable("root/mydb", "old_name", "new_name")
```

### 3. Dynamic Schema Management

**Schema Inference**
- Automatically infers field types from data
- Supports nested structures through flattening
- Type detection: `string`, `int`, `float`, `bool`, `null`, `object`, `array`

**Schema Evolution**
```go
// First row creates initial schema
schema, err := tree.InferSchemaFromRow(row)

// Subsequent rows evolve the schema
updated, err := tree.UpdateSchemaWithNewRow(tablePath, newRow)
```

**Schema Merging**
- Adds new fields automatically
- Increments schema version on changes
- Type compatibility checking
- Preserves existing field types

**Schema Storage**
- Stored in `__schema` node under each table
- Contains: `fields` (JSON), `flat_fields` (JSON), `version`, `created`, `last_updated`

**Type Compatibility**
- Exact match allowed
- All types can be stored as `string`
- `int` and `float` are compatible
- Validation errors on incompatible type changes

### 4. Row/Data Model

**Row Structure** (`models/object.go`)
```go
type Row map[string]*RowValue
type Object struct {
    data    Row
    schema  Schema
    methods map[string]MethodFunc
    vars    map[string]any
}
```

**Features**
- Nested data flattening with dot notation (`address.city`)
- Type inference and validation
- Schema-aware value handling
- Query/match capabilities

**Row Operations**
```go
// Create from various sources
obj := NewObject(data)
obj := NewObjectFromMap(mapData)
obj := NewObjectFromStruct(structData)

// Query
val, ok := obj.Get("address.city")
rowType := obj.GetType("field_name")

// Convert
mapData := obj.ToMap()
obj.ToStruct(&targetStruct)
```

### 5. Storage Layer

**Path System**
- All paths start with `root/`
- Hierarchical addressing: `root/mydb/users/row_123`
- Path normalization and sanitization
- Special node prefixes: `__` (metadata)

**Bucket Operations**
- `rwbucket()`: Read-write operations (creates if not exists)
- `rbucket()`: Read-only operations
- Automatic bucket traversal
- Transaction-based operations

**Core Operations**
```go
tree.CreatePath(path)
tree.DeleteNode(path, force)
tree.GetNodesInPath(path)
tree.SetValue(path, value)
tree.GetValue(path)
tree.SetValues(path, map[string]interface{})
tree.CreateNodeWithProps(path, properties)
tree.BatchCreateNodes(nodesMap)
```

### 6. HTTP API

**Endpoints** (`db.go:514-647`)
```
GET  /{path}         - List nodes
GET  /{path}/props   - Get property names
GET  /{path}/values  - Get all properties with values
GET  /{path}/value   - Get single value
POST /{path}/create  - Create path
POST /{path}/save    - Save multiple properties
POST /{path}/set     - Set single property
```

**Authentication**
- Token-based (`?token=xxx`)
- Public paths allowed (`/public`)
- Middleware integration

**Timeseries API** (Additional feature)
```
POST /timeseries/init
GET  /timeseries/sensors
POST /timeseries/sensors/create
GET  /timeseries/sensors/{sensor}/data
... (10 endpoints total)
```

---

## Testing Coverage

### Test Suite Statistics
- **Total Test Functions**: 50
- **Lines of Test Code**: 1,473
- **Test Coverage**: Comprehensive

### Test Categories

**Database Operations** (7 tests)
- Create, list, get info, delete
- Metadata validation
- Force deletion handling
- Non-database node handling

**Table Operations** (6 tests)
- Create, list, get info, delete
- Parent validation
- Special node filtering

**Schema Management** (13 tests)
- Inference from rows
- Nested structure handling
- Save/retrieve operations
- Schema merging and versioning
- Evolution with new fields
- Type compatibility

**Row Management** (4 tests)
- Row count increment/decrement
- ID generation
- Path validation

**Integration Tests** (6 tests)
- Complete database creation flow
- Nested database structures
- Complex nested data
- Database table count updates

**Edge Cases** (10 tests)
- Empty schemas and databases
- Non-existent entities
- Concurrent operations (databases, tables, schemas)

**Concurrency** (3 tests)
- Concurrent database creation
- Concurrent table creation
- Concurrent schema updates

---

## Data Flow Examples

### 1. Creating a Database with Tables
```go
// 1. Create database
tree.CreateDatabase("root/myapp", map[string]interface{}{
    "description": "My Application Database",
    "owner": "admin",
})

// 2. Create table
tree.CreateTable("root/myapp", "users")

// 3. Schema auto-generated on first row insert
row := models.Row{
    "name": models.NewRowValue("John"),
    "age": models.NewRowValue(30),
    "address.city": models.NewRowValue("NYC"),
}
tree.UpdateSchemaWithNewRow("root/myapp/users", row)

// 4. Schema evolves with new fields
newRow := models.Row{
    "name": models.NewRowValue("Jane"),
    "age": models.NewRowValue(25),
    "email": models.NewRowValue("jane@example.com"), // New field
}
tree.UpdateSchemaWithNewRow("root/myapp/users", newRow)
```

### 2. Querying Data
```go
// Get database info
info, _ := tree.GetDatabaseInfo("root/myapp")
fmt.Printf("Tables: %d\n", info.TableCount)

// Get table info
tableInfo, _ := tree.GetTableInfo("root/myapp/users")
fmt.Printf("Rows: %d, Has Schema: %v\n", tableInfo.RowCount, tableInfo.HasSchema)

// Get schema
schema, _ := tree.GetTableSchema("root/myapp/users")
fmt.Printf("Fields: %v, Version: %d\n", schema.Fields, schema.Version)
```

---

## Schema System

### TableSchema Structure
```go
type TableSchema struct {
    Fields      map[string]string `json:"fields"`       // field_name -> type
    FlatFields  []string          `json:"flat_fields"`  // all field names
    Version     int               `json:"version"`      // schema version
    Created     string            `json:"created"`      // timestamp
    LastUpdated string            `json:"last_updated"` // timestamp
}
```

### Schema Inference Rules
1. Inspect row values
2. Determine types via `RowValue.InferredType()`
3. Create flat field list (supports nested via dot notation)
4. Store in `__schema` node as JSON

### Schema Evolution Process
1. Fetch existing schema
2. Infer schema from new row
3. Merge schemas (add new fields, preserve existing)
4. Increment version if changed
5. Save updated schema

### Supported Types
- `string`: Text values
- `int`: Integer numbers
- `float`: Floating-point numbers
- `bool`: Boolean values
- `null`: Null values
- `object`: Nested objects (flattened with dot notation)
- `array`: Array values (flattened with index notation)

---

## Special Metadata Nodes

All metadata fields start with `__` prefix:

### Database Metadata
- `__type`: "database"
- `__created`: Unix timestamp
- `__lastupdated`: Unix timestamp
- `__table_count`: Number of tables
- `__view_count`: Number of views (future)

### Table Metadata
- `__type`: "table"
- `__name`: Table name
- `__created`: Unix timestamp
- `__lastupdated`: Unix timestamp
- `__row_count`: Number of rows
- `__has_schema`: "true" or "false"

### Schema Metadata (stored in `__schema` node)
- `version`: Schema version number
- `created`: Creation timestamp
- `last_updated`: Last update timestamp
- `fields`: JSON map of field names to types
- `flat_fields`: JSON array of all field names

---

## Performance Considerations

### Optimization Techniques
1. **Bucket Traversal**: Cached path segments
2. **Batch Operations**: `BatchCreateNodes()` for multiple inserts
3. **Transaction Grouping**: Single transaction for multiple operations
4. **Lazy Schema Loading**: Schema only loaded when needed

### Timing
- Bucket operation timing available via `logBucketTiming()`
- Currently disabled but can be enabled for profiling

---

## TODO Items

From `TODO` file:
```
Registry Based:
    All status posted to .Watches in Registry

Probes:
    - Ping, HTTP Status, Row Count (Mongo/SQL)
    - Date range queries, SystemD status
    - Connection tests, Redis operations
    - ID change tracking

Node Relationships:
    - Graph-like relationships
    - Auto-load node properties

Timeseries Data:
    - Sensors with counter/gauge support
    - Push/Pull operations with windowing
```

---

## Current Limitations & Future Work

### Limitations
1. **No Query Language**: Path-based only, no SQL-like queries
2. **No Indexing**: Linear scans for searches
3. **No Transactions Across Paths**: Each operation is atomic but no multi-path transactions
4. **RenameTable Incomplete**: Doesn't copy child nodes (rows, schema, indices)
5. **No Views**: View type defined but not implemented

### Planned Features
1. **Indices**: `__indices` node structure exists but not implemented
2. **Stats**: `__stats` node structure exists but not implemented
3. **Views**: Type constant defined, implementation pending
4. **Relationships**: Graph-like node relationships
5. **Watches**: Registry-based change notifications
6. **Advanced Queries**: Filtering, sorting, aggregation

---

## Integration Features

### Timeseries Support
- Sensor management (create, list, delete)
- Event storage with timestamps
- Windowing queries (tumbling windows)
- Data retention and cleanup
- Real-time data ingestion

### Mini-App Architecture
- Frontend React app with Redux
- Backend Go handlers
- Recharts visualization
- Bootstrap UI
- Complete implementation (see `MINIAPP_SUMMARY.txt`)

---

## Technical Stack

- **Language**: Go
- **Storage**: BoltDB (bbolt) - embedded key-value store
- **HTTP Framework**: microweb (custom lightweight framework)
- **Testing**: Go standard testing package
- **Frontend** (configadmin): React 19, Redux Toolkit, Recharts, Bootstrap 5

---

## Usage Patterns

### 1. Simple CRUD
```go
// Create
tree.CreateDatabase("root/db", nil)
tree.CreateTable("root/db", "table")

// Read
info, _ := tree.GetTableInfo("root/db/table")
schema, _ := tree.GetTableSchema("root/db/table")

// Update
tree.SetValue("root/db/table/__row_count", "100")

// Delete
tree.DeleteTable("root/db", "table", false)
```

### 2. Schema-Aware Data
```go
obj := models.NewObjectFromMap(map[string]any{
    "name": "Product",
    "price": 99.99,
    "tags": []string{"electronics", "new"},
})

// Access flattened fields
name, _ := obj.Get("name")
tag0, _ := obj.Get("tags.0")
```

### 3. Batch Operations
```go
nodes := map[string]map[string]interface{}{
    "root/db/table/row1": {"id": "1", "name": "Alice"},
    "root/db/table/row2": {"id": "2", "name": "Bob"},
    "root/db/table/row3": {"id": "3", "name": "Charlie"},
}
tree.BatchCreateNodes(nodes)
```

---

## Conclusion

BlueConfig provides a **flexible, schema-less database system** with:
- ✅ Hierarchical organization (databases → tables → rows)
- ✅ Dynamic schema inference and evolution
- ✅ Path-based addressing
- ✅ HTTP API for remote access
- ✅ Comprehensive test coverage (50 tests)
- ✅ Type system with compatibility checking
- ✅ Metadata tracking and validation
- ✅ Concurrent operation support

**Status**: Core database functionality is **complete and tested**. Advanced features (indexing, views, relationships) are planned for future releases.
