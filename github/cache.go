package github

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheDir = "cache"
	cacheTTL = 24 * time.Hour // 24 hours cache
)

// Cache represents a simple file-based cache
type Cache struct {
	dir string
}

// NewCache creates a new cache instance
func NewCache() (*Cache, error) {
	// Create cache directory if it doesn't exist
	cachePath := filepath.Join(os.TempDir(), cacheDir)
	fmt.Printf("Creating cache at: %s\n", cachePath)
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	
	return &Cache{
		dir: cachePath,
	}, nil
}

// getCacheKey generates a cache key from the URL
func (c *Cache) getCacheKey(url string) string {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

// Get retrieves content from cache if it exists and is not expired
func (c *Cache) Get(url string) ([]byte, bool, error) {
	key := c.getCacheKey(url)
	cacheFile := filepath.Join(c.dir, key)
	
	// Check if file exists
	info, err := os.Stat(cacheFile)
	if os.IsNotExist(err) {
		fmt.Printf("Cache miss for: %s\n", url)
		return nil, false, nil // Not found in cache
	}
	
	if err != nil {
		return nil, false, err
	}
	
	// Check if cache is expired
	if time.Since(info.ModTime()) > cacheTTL {
		// Remove expired cache
		fmt.Printf("Cache expired for: %s\n", url)
		os.Remove(cacheFile)
		return nil, false, nil
	}
	
	// Read cached content
	content, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, false, err
	}
	
	fmt.Printf("Cache hit for: %s\n", url)
	return content, true, nil
}

// Put stores content in cache
func (c *Cache) Put(url string, content []byte) error {
	key := c.getCacheKey(url)
	cacheFile := filepath.Join(c.dir, key)
	
	fmt.Printf("Caching content for: %s\n", url)
	return os.WriteFile(cacheFile, content, 0644)
}

// Clear removes all cached files
func (c *Cache) Clear() error {
	return os.RemoveAll(c.dir)
}