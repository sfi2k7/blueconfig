package blueconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sfi2k7/microweb"
	"go.etcd.io/bbolt"
)

/*

	path examples:
		node1
		node1/node2
		node1/node2/node3
		node1/node2/node3/node4
		node1/node2/node3/node4/node5

		each node has properties:
			node1:
				name, age, place etc


	object {
		path:'path_to_node'
		propname:'prop_name'
		propvalue:'prop_value'
	}

	setNode = node
	setNodeWithProps = node.prop {value}

	getNode (node)
	getNodeWithProps(node)

	getValue(node.prop)
	setValue(node.prop, value)
	deleteProp(node.prop)


*/

// ============================================================================
// Constants
// ============================================================================

const (
	CmdCreatePath = 99
	CmdDeleteNode = 100
)

// ============================================================================
// Types
// ============================================================================

type TreeOptions struct {
	StorageLocationOnDisk string
	Port                  int
	Token                 string
}

type Tree struct {
	db       *bbolt.DB
	diskpath string
	port     int
	token    string
}

type Packet struct {
	Path  string `json:"path"`
	Prop  string `json:"prop"`
	Value string `json:"val"`
}

type response struct {
	Error  string      `json:"error,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

// ============================================================================
// Path Utilities
// ============================================================================

// fixpath normalizes a path to ensure it starts with root and has no double slashes
func fixpath(p string) string {
	// Replace all multiple slashes with single slash
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}
	p = strings.TrimPrefix(p, "/")
	p = strings.TrimSuffix(p, "/")

	if p == "/" || p == "" || p == "root" {
		p = "root"
	} else {
		if !strings.HasPrefix(p, "root/") {
			p = "root/" + p
		}
	}

	return p
}

// parsePath extracts node, property and value from a path based on offset
// prop and value will be extracted based on offset
// for example:
// node1/childnode/prop where prop is prop and offset is 1
// node1/childnode/prop/value where prop is prop and value is value and offset is 2
func parsePath(p string, offset int) (node, prop, value string, err error) {
	p = fixpath(p)

	if offset == 0 {
		return p, "", "", nil
	}

	splitted := strings.Split(p, "/")

	if len(splitted) <= offset {
		return "", "", "", errors.New("invalid offset")
	}

	head := splitted[:len(splitted)-offset]
	if len(splitted)-len(head) < offset {
		return "", "", "", errors.New("invalid offset2")
	}

	node = strings.Join(head, "/")
	tail := splitted[len(splitted)-offset:]
	prop = tail[0]

	if len(tail) > 1 {
		value = tail[1]
	}

	return node, prop, value, nil
}

// sanitizeBucketName removes invalid characters from bucket names
func sanitizeBucketName(s string) string {
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "\"", "_")
	s = strings.ReplaceAll(s, "'", "_")
	s = strings.ReplaceAll(s, "`", "_")
	return s
}

// ============================================================================
// Tree Construction
// ============================================================================

func NewOrOpenTree(options TreeOptions) (*Tree, error) {
	os.MkdirAll(filepath.Dir(options.StorageLocationOnDisk), 0600)

	db, err := bbolt.Open(options.StorageLocationOnDisk, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &Tree{
		db:    db,
		port:  options.Port,
		token: options.Token,
	}, nil
}

func (t *Tree) Close() error {
	return t.db.Close()
}

// ============================================================================
// Bucket Access Layer
// ============================================================================

// bucketOperation encapsulates common bucket traversal logic
type bucketOperation struct {
	path   string
	offset int
}

// logBucketTiming logs the time taken for a bucket operation
func logBucketTiming(path string, start time.Time) {
	// fmt.Printf("Bucket(%s) took %s\t", path, time.Since(start))
}

// traverseBuckets navigates to the target bucket in a path
func traverseBuckets(segments []string, rootBucket func(name string) *bbolt.Bucket, nestedBucket func(parent *bbolt.Bucket, name string) *bbolt.Bucket) (*bbolt.Bucket, error) {
	var b *bbolt.Bucket

	for _, segment := range segments {
		if b == nil {
			b = rootBucket(segment)
		} else {
			b = nestedBucket(b, segment)
		}

		if b == nil {
			return nil, errors.New("path does not exist")
		}
	}

	return b, nil
}

// rbucket executes a read-only operation on a bucket at the given path
func (t *Tree) rbucket(path string, offset int, fn func(b *bbolt.Bucket) error) error {
	start := time.Now()
	defer logBucketTiming(path, start)

	path = fixpath(path)
	var err error
	path, _, _, err = parsePath(path, offset)
	if err != nil {
		return err
	}

	segments := strings.Split(path, "/")

	return t.db.View(func(tx *bbolt.Tx) error {
		b, err := traverseBuckets(
			segments,
			func(name string) *bbolt.Bucket {
				return tx.Bucket([]byte(name))
			},
			func(parent *bbolt.Bucket, name string) *bbolt.Bucket {
				return parent.Bucket([]byte(name))
			},
		)
		if err != nil {
			return err
		}
		return fn(b)
	})
}

// rwbucket executes a read-write operation on a bucket at the given path
// Creates buckets if they don't exist
func (t *Tree) rwbucket(path string, fn func(b *bbolt.Bucket) error) error {
	start := time.Now()
	defer logBucketTiming(path, start)

	path = fixpath(path)
	segments := strings.Split(path, "/")

	return t.db.Update(func(tx *bbolt.Tx) error {
		var b *bbolt.Bucket
		var err error

		for _, segment := range segments {
			segment = sanitizeBucketName(segment)

			if b == nil {
				b, err = tx.CreateBucketIfNotExists([]byte(segment))
			} else {
				b, err = b.CreateBucketIfNotExists([]byte(segment))
			}

			if err != nil {
				return err
			}
		}

		return fn(b)
	})
}

// ============================================================================
// Path and Node Operations
// ============================================================================

func (t *Tree) CreatePath(p string) error {
	return t.rwbucket(p, func(b *bbolt.Bucket) error {
		return nil
	})
}

func (t *Tree) DeleteNode(p string, force bool) error {
	p = fixpath(p)
	segments := strings.Split(p, "/")

	nodeToDelete := segments[len(segments)-1]
	parentPath := strings.Join(segments[:len(segments)-1], "/")

	// fmt.Println("going to delete", nodeToDelete)

	if nodeToDelete == "/" || nodeToDelete == "root" || nodeToDelete == "/root" {
		return errors.New("can not delete root node")
	}

	return t.rwbucket(parentPath, func(b *bbolt.Bucket) error {
		innerBucket := b.Bucket([]byte(nodeToDelete))
		if innerBucket == nil {
			return errors.New("node does not exist")
		}

		// Check if bucket has nested buckets
		err := innerBucket.ForEachBucket(func(k []byte) error {
			return errors.New("node has nested nodes - must force to delete")
		})

		if err != nil && !force {
			return err
		}

		return b.DeleteBucket([]byte(nodeToDelete))
	})
}

func (t *Tree) GetNodesInPath(p string) ([]string, error) {
	var nodes []string
	err := t.rbucket(p, 0, func(b *bbolt.Bucket) error {
		return b.ForEachBucket(func(k []byte) error {
			nodes = append(nodes, string(k))
			return nil
		})
	})

	return nodes, err
}

func (t *Tree) GetChildren(p string) ([]string, error) {
	return t.GetNodesInPath(p)
}

// ============================================================================
// Property Operations
// ============================================================================

func (t *Tree) SetValue(p, value string) error {
	nodePath, prop, _, _ := parsePath(p, 1)
	return t.rwbucket(nodePath, func(b *bbolt.Bucket) error {
		return b.Put([]byte(prop), []byte(value))
	})
}

func (t *Tree) SetValues(p string, values map[string]interface{}) error {
	return t.rwbucket(p, func(b *bbolt.Bucket) error {
		for k, v := range values {
			err := b.Put([]byte(k), []byte(fmt.Sprintf("%v", v)))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// CreateNodeWithProps creates a node at the given path and sets its properties in a single transaction
// This is more efficient than calling CreatePath() followed by SetValues()
// Example: CreateNodeWithProps("/users/123", map[string]interface{}{"name": "John", "age": 30})
func (t *Tree) CreateNodeWithProps(p string, properties map[string]interface{}) error {
	return t.rwbucket(p, func(b *bbolt.Bucket) error {
		for k, v := range properties {
			err := b.Put([]byte(k), []byte(fmt.Sprintf("%v", v)))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// SetNodeWithProps is an alias for CreateNodeWithProps for clarity
// Both create the path if it doesn't exist and set the properties
func (t *Tree) SetNodeWithProps(p string, properties map[string]interface{}) error {
	return t.CreateNodeWithProps(p, properties)
}

// BatchCreateNodes creates multiple nodes with their properties in a single transaction
// This is the most efficient way to create many nodes at once
//
//	Example: BatchCreateNodes(map[string]map[string]interface{}{
//	    "/users/1": {"name": "John", "age": 30},
//	    "/users/2": {"name": "Jane", "age": 25},
//	})
func (t *Tree) BatchCreateNodes(nodes map[string]map[string]interface{}) error {
	return t.db.Update(func(tx *bbolt.Tx) error {
		for path, properties := range nodes {
			path = fixpath(path)
			segments := strings.Split(path, "/")

			var b *bbolt.Bucket
			var err error

			// Navigate/create path
			for _, segment := range segments {
				segment = sanitizeBucketName(segment)

				if b == nil {
					b, err = tx.CreateBucketIfNotExists([]byte(segment))
				} else {
					b, err = b.CreateBucketIfNotExists([]byte(segment))
				}

				if err != nil {
					return err
				}
			}

			// Set properties
			for k, v := range properties {
				err := b.Put([]byte(k), []byte(fmt.Sprintf("%v", v)))
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (t *Tree) GetValue(p string) (string, error) {
	_, prop, _, _ := parsePath(p, 1)

	var value string
	err := t.rbucket(p, 1, func(b *bbolt.Bucket) error {
		value = string(b.Get([]byte(prop)))
		return nil
	})
	return value, err
}

func (t *Tree) GetAllProps(p string) ([]string, error) {
	var props []string
	err := t.rbucket(p, 0, func(b *bbolt.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			if v == nil { // Skip buckets, only get key-value pairs
				return nil
			}
			props = append(props, string(k))
			return nil
		})
	})
	return props, err
}

func (t *Tree) GetAllPropsWithValues(p string) (map[string]string, error) {
	props := make(map[string]string)
	err := t.rbucket(p, 0, func(b *bbolt.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			if v == nil { // Skip buckets, only get key-value pairs
				return nil
			}
			props[string(k)] = string(v)
			return nil
		})
	})
	return props, err
}

func (t *Tree) DeleteValue(p, prop string) error {
	return t.rwbucket(p, func(b *bbolt.Bucket) error {
		return b.Delete([]byte(prop))
	})
}

func (t *Tree) HasValue(p, prop string) (bool, error) {
	var hasValue bool
	err := t.rbucket(p, 0, func(b *bbolt.Bucket) error {
		hasValue = b.Get([]byte(prop)) != nil
		return nil
	})
	return hasValue, err
}

// ============================================================================
// Node Scanning Operations
// ============================================================================

type NodeInfo struct {
	Name  string
	Props map[string]string
}

// NodeScanCallback is called for each node during scanning
type NodeScanCallback func(nodeInfo NodeInfo) error

// ScanNodes scans all child nodes and calls the callback for each node with its properties
func (t *Tree) ScanNodes(p string, callback NodeScanCallback) error {
	err := t.rbucket(p, 0, func(b *bbolt.Bucket) error {
		return b.ForEachBucket(func(k []byte) error {
			nodeName := string(k)
			nodeInfo := NodeInfo{
				Name:  nodeName,
				Props: make(map[string]string),
			}

			// Get the nested bucket to read its properties
			childBucket := b.Bucket(k)
			if childBucket != nil {
				childBucket.ForEach(func(propKey, propVal []byte) error {
					if propVal != nil { // Skip nested buckets
						nodeInfo.Props[string(propKey)] = string(propVal)
					}
					return nil
				})
			}

			// Call the callback with the node info
			return callback(nodeInfo)
		})
	})
	return err
}

// ============================================================================
// HTTP Server and Handlers
// ============================================================================

func (t *Tree) Serve() {
	if t.port == 0 {
		panic("http port not set")
	}

	web := microweb.New()
	web.Use(t.authMiddleware)

	// Core API endpoints
	web.Get("/", t.handleGetRequest)
	web.Post("/", t.handlePostRequest)

	// Timeseries group endpoints
	tsGroup := web.Group("/timeseries")
	t.registerTimeseriesRoutes(tsGroup)

	web.Listen(t.port)
}

// authMiddleware validates the authentication token
func (t *Tree) authMiddleware(c *microweb.Context) bool {
	if strings.HasPrefix(c.R.URL.Path, "/public") {
		return true
	}

	if t.token != "" && c.Query("token") != t.token {
		c.Json(response{Error: "Invalid token"})
		return false
	}
	return true
}

// handleGetRequest handles all GET requests
func (t *Tree) handleGetRequest(c *microweb.Context) {
	path := c.R.URL.Path

	if strings.HasSuffix(path, "/props") {
		props, err := t.GetAllProps(strings.TrimSuffix(path, "/props"))
		if err != nil {
			c.Json(response{Error: err.Error()})
			return
		}
		c.Json(response{Result: props})
		return
	}

	if strings.HasSuffix(path, "/values") {
		propsvals, err := t.GetAllPropsWithValues(strings.TrimSuffix(path, "/values"))
		if err != nil {
			c.Json(response{Error: err.Error()})
			return
		}
		c.Json(response{Result: propsvals})
		return
	}

	if strings.HasSuffix(path, "/value") {
		value, err := t.GetValue(strings.TrimSuffix(path, "/value"))
		if err != nil {
			c.Json(response{Error: err.Error()})
			return
		}
		c.Json(response{Result: value})
		return
	}

	nodes, err := t.GetNodesInPath(path)
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}
	c.Json(response{Result: nodes})
}

// handlePostRequest handles all POST requests
func (t *Tree) handlePostRequest(c *microweb.Context) {
	path := c.R.URL.Path

	if strings.HasSuffix(path, "/save") {
		body, err := c.Body()
		if err != nil {
			c.Json(response{Error: err.Error()})
			return
		}

		var m = make(map[string]any)
		err = json.Unmarshal(body, &m)
		if err != nil {
			c.Json(response{Error: err.Error()})
			return
		}

		t.SetValues(strings.TrimSuffix(path, "/save"), m)
		c.Json(response{Result: true})
		return
	}

	if strings.HasSuffix(path, "/set") {
		nodePath, prop, val, err := parsePath(strings.TrimSuffix(path, "/set"), 2)
		fmt.Println("path:", nodePath, "prop:", prop, "val:", val, "err:", err)

		restored, _ := url.JoinPath(nodePath, prop)
		fmt.Println("restored", restored)

		err = t.SetValue(restored, val)
		if err != nil {
			c.Json(response{Error: err.Error()})
			return
		}
		c.Json(response{Result: true})
		return
	}

	if strings.HasSuffix(path, "/create") {
		err := t.CreatePath(strings.TrimSuffix(path, "/create"))
		if err != nil {
			c.Json(response{Error: err.Error()})
			return
		}
		c.Json(response{Result: true})
		return
	}

	c.Json(response{Error: "unknown error"})

	/*
		examples:
			/ create path
			/ set setvalue
			/ save multiple set values
	*/
}

// ============================================================================
// Timeseries HTTP Handlers
// ============================================================================

// registerTimeseriesRoutes registers all timeseries endpoints in a group
func (t *Tree) registerTimeseriesRoutes(g *microweb.Group) {
	// Initialization
	g.Post("/init", t.handleInitTimeseries)

	// Sensor management
	g.Post("/sensors/create", t.handleCreateSensor)
	g.Get("/sensors", t.handleListSensors)
	g.Get("/sensors/{sensor}/info", t.handleGetSensorInfo)
	g.Delete("/sensors/{sensor}", t.handleDeleteSensor)

	// Event management
	g.Post("/sensors/{sensor}/event", t.handlePostEvent)
	g.Get("/sensors/{sensor}/events", t.handleGetAllEvents)
	g.Get("/sensors/{sensor}/latest", t.handleGetLatestValue)
	g.Post("/sensors/{sensor}/cleanup", t.handleDeleteOldEvents)

	// Query
	g.Get("/sensors/{sensor}/data", t.handleGetSensorData)
}

// handleInitTimeseries initializes the timeseries system
func (t *Tree) handleInitTimeseries(c *microweb.Context) {
	err := t.InitTimeseries()
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}
	c.Json(response{Result: true})
}

// handleListSensors returns all sensors
func (t *Tree) handleListSensors(c *microweb.Context) {
	sensors, err := t.ListSensors()
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}
	c.Json(response{Result: sensors})
}

// handleGetSensorInfo returns sensor metadata
func (t *Tree) handleGetSensorInfo(c *microweb.Context) {
	sensor := c.Param("sensor")
	info, err := t.GetSensorInfo(sensor)
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}
	c.Json(response{Result: info})
}

// handleCreateSensor creates a new sensor
func (t *Tree) handleCreateSensor(c *microweb.Context) {
	body, err := c.Body()
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}

	var req struct {
		Name       string `json:"name"`
		SensorType string `json:"sensor_type"`
		Unit       string `json:"unit"`
		Retention  string `json:"retention"`
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}

	err = t.CreateSensor(req.Name, req.SensorType, req.Unit, req.Retention)
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}
	c.Json(response{Result: true})
}

// handlePostEvent posts an event to a sensor
func (t *Tree) handlePostEvent(c *microweb.Context) {
	sensor := c.Param("sensor")
	body, err := c.Body()
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}

	var event Event
	err = json.Unmarshal(body, &event)
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}

	err = t.PostEvent(sensor, event)
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}
	c.Json(response{Result: true})
}

// handleGetAllEvents returns all events for a sensor
func (t *Tree) handleGetAllEvents(c *microweb.Context) {
	sensor := c.Param("sensor")
	events, err := t.GetAllEvents(sensor)
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}
	c.Json(response{Result: events})
}

// handleGetLatestValue returns the latest value for a sensor
func (t *Tree) handleGetLatestValue(c *microweb.Context) {
	sensor := c.Param("sensor")
	latest, err := t.GetLatestValue(sensor)
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}
	c.Json(response{Result: latest})
}

// handleGetSensorData queries sensor data with windowing
func (t *Tree) handleGetSensorData(c *microweb.Context) {
	sensor := c.Param("sensor")

	query := SensorQuery{
		WindowType: WindowType(c.Query("window_type")),
		WindowSize: c.Query("window_size"),
		Scale:      c.Query("scale"),
		OutputType: c.Query("output_type"),
	}

	// Set defaults
	if query.WindowType == "" {
		query.WindowType = TumblingWindow
	}
	if query.OutputType == "" {
		query.OutputType = OutputTypeGraph
	}

	result, err := t.GetSensorData(sensor, query)
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}
	c.Json(response{Result: result})
}

// handleDeleteOldEvents deletes old events from a sensor
func (t *Tree) handleDeleteOldEvents(c *microweb.Context) {
	sensor := c.Param("sensor")
	deleted, err := t.DeleteOldEvents(sensor)
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}
	c.Json(response{Result: map[string]int{"deleted": deleted}})
}

// handleDeleteSensor deletes a sensor
func (t *Tree) handleDeleteSensor(c *microweb.Context) {
	sensor := c.Param("sensor")
	err := t.DeleteSensor(sensor)
	if err != nil {
		c.Json(response{Error: err.Error()})
		return
	}
	c.Json(response{Result: true})
}
