package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sfi2k7/blueconfig"
)

// ConfigStoreManager manages multiple BlueConfig stores
type ConfigStoreManager struct {
	stores    map[string]*blueconfig.Tree
	storePath string
	mu        sync.RWMutex
}

// StoreInfo represents metadata about a store
type StoreInfo struct {
	Name         string    `json:"name"`
	DisplayName  string    `json:"displayName"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"createdAt"`
	LastAccessed time.Time `json:"lastAccessed"`
	Icon         string    `json:"icon"`
	Color        string    `json:"color"`
	HasPassword  bool      `json:"hasPassword"`
}

// NewConfigStoreManager creates a new store manager
func NewConfigStoreManager(storePath string) (*ConfigStoreManager, error) {
	// Create stores directory if it doesn't exist
	if err := os.MkdirAll(storePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create stores directory: %v", err)
	}

	csm := &ConfigStoreManager{
		stores:    make(map[string]*blueconfig.Tree),
		storePath: storePath,
	}

	// Ensure default store exists
	if err := csm.ensureDefaultStore(); err != nil {
		return nil, fmt.Errorf("failed to create default store: %v", err)
	}

	return csm, nil
}

// ensureDefaultStore creates the default store if it doesn't exist
func (csm *ConfigStoreManager) ensureDefaultStore() error {
	defaultPath := filepath.Join(csm.storePath, "default.db")
	if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
		return csm.Create("default", "Default Store", "Default configuration store", "")
	}
	return nil
}

// Create creates a new store with metadata
func (csm *ConfigStoreManager) Create(name, displayName, description, password string) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	// Validate name
	if name == "" {
		return fmt.Errorf("store name cannot be empty")
	}

	// Check if store already exists
	storePath := filepath.Join(csm.storePath, name+".db")
	if _, err := os.Stat(storePath); err == nil {
		return fmt.Errorf("store '%s' already exists", name)
	}

	// Create the store
	tree, err := blueconfig.NewOrOpenTree(blueconfig.TreeOptions{
		StorageLocationOnDisk: storePath,
	})
	if err != nil {
		return fmt.Errorf("failed to create store: %v", err)
	}

	// Create __storeinfo path
	if err := tree.CreatePath("root/__storeinfo"); err != nil {
		tree.Close()
		return fmt.Errorf("failed to create storeinfo path: %v", err)
	}

	// Set metadata
	now := time.Now().Format(time.RFC3339)

	// Display properties
	if displayName == "" {
		displayName = name
	}
	tree.SetValue("root/__storeinfo/__title", displayName)
	tree.SetValue("root/__storeinfo/__icon", "fa-database")
	tree.SetValue("root/__storeinfo/__color", "#3498db")

	// Password (if provided)
	if password != "" {
		tree.SetValue("root/__storeinfo/__password", password)
	}

	// Regular metadata
	tree.SetValue("root/__storeinfo/name", name)
	tree.SetValue("root/__storeinfo/displayName", displayName)
	tree.SetValue("root/__storeinfo/description", description)
	tree.SetValue("root/__storeinfo/createdAt", now)
	tree.SetValue("root/__storeinfo/lastAccessed", now)

	// Cache the store
	csm.stores[name] = tree

	return nil
}

// Load loads a store from disk or returns cached instance
func (csm *ConfigStoreManager) Load(name string) (*blueconfig.Tree, error) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	// Check cache first
	if tree, exists := csm.stores[name]; exists {
		// Update last accessed
		tree.SetValue("root/__storeinfo/lastAccessed", time.Now().Format(time.RFC3339))
		return tree, nil
	}

	// Load from disk
	storePath := filepath.Join(csm.storePath, name+".db")
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("store '%s' not found", name)
	}

	tree, err := blueconfig.NewOrOpenTree(blueconfig.TreeOptions{
		StorageLocationOnDisk: storePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open store: %v", err)
	}

	// Update last accessed
	tree.SetValue("root/__storeinfo/lastAccessed", time.Now().Format(time.RFC3339))

	// Cache the store
	csm.stores[name] = tree

	return tree, nil
}

// List returns all available stores
func (csm *ConfigStoreManager) List() ([]StoreInfo, error) {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	var stores []StoreInfo

	// Read all .db files in stores directory
	files, err := ioutil.ReadDir(csm.storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read stores directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".db") {
			continue
		}

		// Extract store name
		name := strings.TrimSuffix(file.Name(), ".db")

		// Get store info
		storeInfo, err := csm.getStoreInfo(name)
		if err != nil {
			// Add basic info even if we can't read metadata
			stores = append(stores, StoreInfo{
				Name:        name,
				DisplayName: name,
				Icon:        "fa-database",
				Color:       "#6c757d",
			})
			continue
		}

		stores = append(stores, storeInfo)
	}

	return stores, nil
}

// getStoreInfo reads store metadata from __storeinfo
func (csm *ConfigStoreManager) getStoreInfo(name string) (StoreInfo, error) {
	// Try to load from cache first
	var tree *blueconfig.Tree
	var shouldClose bool

	if cachedTree, exists := csm.stores[name]; exists {
		tree = cachedTree
	} else {
		// Open temporarily to read metadata
		storePath := filepath.Join(csm.storePath, name+".db")
		var err error
		tree, err = blueconfig.NewOrOpenTree(blueconfig.TreeOptions{
			StorageLocationOnDisk: storePath,
		})
		if err != nil {
			return StoreInfo{}, err
		}
		shouldClose = true
		defer func() {
			if shouldClose {
				tree.Close()
			}
		}()
	}

	// Read metadata
	info := StoreInfo{
		Name: name,
	}

	// Try to read each field, use defaults if missing
	if displayName, err := tree.GetValue("root/__storeinfo/displayName"); err == nil && displayName != "" {
		info.DisplayName = displayName
	} else {
		info.DisplayName = name
	}

	if description, err := tree.GetValue("root/__storeinfo/description"); err == nil {
		info.Description = description
	}

	if icon, err := tree.GetValue("root/__storeinfo/__icon"); err == nil && icon != "" {
		info.Icon = icon
	} else {
		info.Icon = "fa-database"
	}

	if color, err := tree.GetValue("root/__storeinfo/__color"); err == nil && color != "" {
		info.Color = color
	} else {
		info.Color = "#3498db"
	}

	if createdAtStr, err := tree.GetValue("root/__storeinfo/createdAt"); err == nil && createdAtStr != "" {
		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			info.CreatedAt = t
		}
	}

	if lastAccessedStr, err := tree.GetValue("root/__storeinfo/lastAccessed"); err == nil && lastAccessedStr != "" {
		if t, err := time.Parse(time.RFC3339, lastAccessedStr); err == nil {
			info.LastAccessed = t
		}
	}

	if passwordStr, err := tree.GetValue("root/__storeinfo/__password"); err == nil && passwordStr != "" {
		info.HasPassword = true
	}

	return info, nil
}

// UpdateStoreInfo updates store metadata
func (csm *ConfigStoreManager) UpdateStoreInfo(name string, updates map[string]string) error {
	tree, err := csm.Load(name)
	if err != nil {
		return err
	}

	// Update allowed fields
	allowedFields := map[string]bool{
		"displayName": true,
		"description": true,
		"__icon":      true,
		"__color":     true,
		"__title":     true,
		"__password":  true,
	}

	for key, value := range updates {
		if !allowedFields[key] {
			continue
		}
		fullPath := fmt.Sprintf("root/__storeinfo/%s", key)
		tree.SetValue(fullPath, value)
	}

	return nil
}

// Delete removes a store
func (csm *ConfigStoreManager) Delete(name string) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	// Don't allow deleting default store
	if name == "default" {
		return fmt.Errorf("cannot delete default store")
	}

	// Close if cached
	if tree, exists := csm.stores[name]; exists {
		tree.Close()
		delete(csm.stores, name)
	}

	// Delete file
	storePath := filepath.Join(csm.storePath, name+".db")
	if err := os.Remove(storePath); err != nil {
		return fmt.Errorf("failed to delete store: %v", err)
	}

	return nil
}

// Close closes a specific store
func (csm *ConfigStoreManager) Close(name string) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if tree, exists := csm.stores[name]; exists {
		if err := tree.Close(); err != nil {
			return err
		}
		delete(csm.stores, name)
	}

	return nil
}

// CloseAll closes all open stores
func (csm *ConfigStoreManager) CloseAll() {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	for _, tree := range csm.stores {
		tree.Close()
	}
	csm.stores = make(map[string]*blueconfig.Tree)
}

// ExportStoreList exports store list as JSON (for debugging)
func (csm *ConfigStoreManager) ExportStoreList() (string, error) {
	stores, err := csm.List()
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(stores, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
