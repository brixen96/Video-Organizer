package video

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"video-organizer/internal/appstatus"
	"video-organizer/internal/models"
)

const cacheFileName = "video_cache.json"

// CacheMetadata holds metadata about the cached data
type CacheMetadata struct {
	Version      string    `json:"version"`
	CachedAt     time.Time `json:"cached_at"`
	VideoCount   int       `json:"video_count"`
	VideoDir     string    `json:"video_dir"`
	ThumbnailDir string    `json:"thumbnail_dir"`
}

// PersistedCache represents the cache data saved to disk
type PersistedCache struct {
	Metadata CacheMetadata              `json:"metadata"`
	Videos   map[string]models.VideoInfo `json:"videos"`
}

// SaveCacheToDisk persists the video cache to disk
func SaveCacheToDisk() error {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	appstatus.EmitInfo("[Cache]", "Saving video cache to disk")

	cache := PersistedCache{
		Metadata: CacheMetadata{
			Version:      "1.0",
			CachedAt:     time.Now(),
			VideoCount:   len(VideoCache),
			VideoDir:     "", // Will be set from config
			ThumbnailDir: "",
		},
		Videos: VideoCache,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cacheFileName, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	appstatus.EmitInfo("[Cache]", fmt.Sprintf("Saved %d videos to cache file", len(VideoCache)))
	return nil
}

// LoadCacheFromDisk loads the video cache from disk
func LoadCacheFromDisk() (map[string]models.VideoInfo, error) {
	appstatus.EmitInfo("[Cache]", "Attempting to load cache from disk")

	// Check if cache file exists
	if _, err := os.Stat(cacheFileName); os.IsNotExist(err) {
		return nil, fmt.Errorf("cache file does not exist")
	}

	data, err := os.ReadFile(cacheFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache PersistedCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	appstatus.EmitInfo("[Cache]", fmt.Sprintf("Loaded %d videos from cache (cached at %s)",
		cache.Metadata.VideoCount, cache.Metadata.CachedAt.Format(time.RFC3339)))

	return cache.Videos, nil
}

// IsCacheValid checks if the cache is still valid
func IsCacheValid(videoDir string) bool {
	// Check if cache file exists
	if _, err := os.Stat(cacheFileName); os.IsNotExist(err) {
		return false
	}

	// Read cache metadata
	data, err := os.ReadFile(cacheFileName)
	if err != nil {
		log.Printf("Failed to read cache file: %v", err)
		return false
	}

	var cache PersistedCache
	if err := json.Unmarshal(data, &cache); err != nil {
		log.Printf("Failed to unmarshal cache: %v", err)
		return false
	}

	// Check if cache is older than 24 hours
	if time.Since(cache.Metadata.CachedAt) > 24*time.Hour {
		appstatus.EmitInfo("[Cache]", "Cache is older than 24 hours, rebuilding")
		return false
	}

	// Check if video directory has changed
	files, err := os.ReadDir(videoDir)
	if err != nil {
		log.Printf("Failed to read video directory: %v", err)
		return false
	}

	// Count current video files
	currentVideoCount := 0
	for _, file := range files {
		if !file.IsDir() {
			ext := filepath.Ext(file.Name())
			if ext == ".mp4" || ext == ".mkv" || ext == ".avi" || ext == ".mov" {
				currentVideoCount++
			}
		}
	}

	// If video count changed significantly (more than 10%), rebuild cache
	if abs(currentVideoCount-cache.Metadata.VideoCount) > max(1, cache.Metadata.VideoCount/10) {
		appstatus.EmitInfo("[Cache]", fmt.Sprintf("Video count changed significantly (%d -> %d), rebuilding cache",
			cache.Metadata.VideoCount, currentVideoCount))
		return false
	}

	appstatus.EmitInfo("[Cache]", "Cache is valid, using cached data")
	return true
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
