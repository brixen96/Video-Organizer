package video

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"video-organizer/internal/appstatus"
	"video-organizer/internal/config"
	"video-organizer/internal/database"
	"video-organizer/internal/models"
)

var (
	VideoCache map[string]models.VideoInfo
	cacheMutex sync.RWMutex
)

func InitializeVideoCache() {
	appstatus.EmitInfo("[Task]", "Starting video cache initialization")

	// OPTIMIZATION: Try loading from disk cache first
	if IsCacheValid(config.VideoDir) {
		cachedVideos, err := LoadCacheFromDisk()
		if err == nil && len(cachedVideos) > 0 {
			cacheMutex.Lock()
			VideoCache = cachedVideos
			cacheMutex.Unlock()

			appstatus.EmitInfo("[Task]", fmt.Sprintf("Video cache loaded from disk with %d videos (instant startup!)", len(VideoCache)))

			// Update performer scene counts in the background
			go UpdatePerformerSceneCounts()
			return
		}
		appstatus.EmitWarning("[Warning]", fmt.Sprintf("Failed to load cache from disk: %v, rebuilding...", err))
	}

	// Cache not valid or doesn't exist, build from scratch
	cacheMutex.Lock()
	VideoCache = make(map[string]models.VideoInfo)
	cacheMutex.Unlock()

	// Get all performer names from the database
	performerNames, err := database.GetAllPerformerNames()
	if err != nil {
		appstatus.EmitError("[Error]", fmt.Sprintf("Failed to get performer names for cache init: %v", err))
		log.Fatalf("Failed to get performer names for cache initialization: %v", err)
	}

	files, err := os.ReadDir(config.VideoDir)
	if err != nil {
		appstatus.EmitError("[Error]", fmt.Sprintf("Failed to read video directory for cache init: %v", err))
		log.Fatalf("Failed to read video directory for cache initialization: %v", err)
	}

	// Count video files and collect jobs
	var videoJobs []VideoProcessingJob
	for _, file := range files {
		if !file.IsDir() {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".mp4" || ext == ".mkv" || ext == ".avi" || ext == ".mov" {
				videoPath := filepath.Join(config.VideoDir, file.Name())
				videoJobs = append(videoJobs, VideoProcessingJob{
					FileName:       file.Name(),
					VideoPath:      videoPath,
					PerformerNames: performerNames,
				})
			}
		}
	}
	totalVideoCount := len(videoJobs)
	appstatus.EmitInfo("[Info]", fmt.Sprintf("Found %d video files to process", totalVideoCount))

	if totalVideoCount == 0 {
		appstatus.EmitInfo("[Task]", "No videos found to process")
		return
	}

	// Create worker pool (number of workers = number of CPU cores, max 8)
	numWorkers := min(8, max(1, totalVideoCount/10)) // Scale workers based on video count
	appstatus.EmitInfo("[Info]", fmt.Sprintf("Starting %d workers for parallel processing", numWorkers))

	pool := NewWorkerPool(numWorkers)
	pool.Start()

	// Submit all jobs
	go func() {
		for _, job := range videoJobs {
			pool.Submit(job)
		}
		pool.Close()
	}()

	// Collect results
	processedCount := 0
	for result := range pool.Results() {
		processedCount++

		cacheMutex.Lock()
		VideoCache[result.VideoInfo.Name] = result.VideoInfo
		cacheMutex.Unlock()

		if processedCount%5 == 0 || processedCount == totalVideoCount {
			appstatus.EmitInfo("[Progress]", fmt.Sprintf("Processed %d/%d videos", processedCount, totalVideoCount))
		}
	}

	appstatus.EmitInfo("[Task]", fmt.Sprintf("Video cache initialized with %d videos", len(VideoCache)))

	// Save cache to disk for next startup
	go func() {
		if err := SaveCacheToDisk(); err != nil {
			log.Printf("Failed to save cache to disk: %v", err)
		}
	}()

	// Update performer scene counts in the database
	UpdatePerformerSceneCounts()
}

func GetVideoCache() map[string]models.VideoInfo {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	// Return a copy to prevent external modifications
	cache := make(map[string]models.VideoInfo, len(VideoCache))
	for k, v := range VideoCache {
		cache[k] = v
	}
	return cache
}

// GetVideoByName retrieves a single video from cache (thread-safe)
func GetVideoByName(name string) (models.VideoInfo, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	video, exists := VideoCache[name]
	return video, exists
}

// UpdateVideoInCache updates a single video in cache (thread-safe)
func UpdateVideoInCache(name string, video models.VideoInfo) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	VideoCache[name] = video
}

// DeleteVideoFromCache removes a video from cache (thread-safe)
func DeleteVideoFromCache(name string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	delete(VideoCache, name)
}

func GenerateVideoInfo(fileName, videoPath string, performerNames []string) models.VideoInfo {
	appstatus.EmitInfo("[Progress]", fmt.Sprintf("Processing video: %s", fileName))
	thumbnailName := fmt.Sprintf("%s.jpg", strings.TrimSuffix(fileName, filepath.Ext(fileName)))
	thumbnailPath := filepath.Join(config.ThumbnailDir, thumbnailName)

	// Generate thumbnail if it doesn't exist
	if _, err := os.Stat(thumbnailPath); os.IsNotExist(err) {
		appstatus.EmitInfo("[Progress]", fmt.Sprintf("Generating thumbnail for %s", fileName))
		cmd := exec.Command("ffmpeg", "-i", videoPath, "-ss", "00:00:01.000", "-vframes", "1", thumbnailPath)
		if err := cmd.Run(); err != nil {
			msg := fmt.Sprintf("Failed to generate thumbnail for %s: %v", fileName, err)
			appstatus.EmitError("[Error]", msg)
			log.Println(msg)
		} else {
			appstatus.EmitInfo("[Info]", fmt.Sprintf("Thumbnail generated for %s", fileName))
		}
	}

	// Get video metadata using ffprobe
	appstatus.EmitInfo("[Progress]", fmt.Sprintf("Extracting metadata for %s", fileName))
	var metadata models.FFProbeOutput
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", videoPath)
	output, err := cmd.Output()
	if err != nil {
		msg := fmt.Sprintf("Failed to get metadata for %s: %v", fileName, err)
		appstatus.EmitError("[Error]", msg)
		log.Println(msg)
	} else {
		if err := json.Unmarshal(output, &metadata); err != nil {
			msg := fmt.Sprintf("Failed to parse metadata for %s: %v", fileName, err)
			appstatus.EmitError("[Error]", msg)
			log.Println(msg)
		} else {
			appstatus.EmitInfo("[Info]", fmt.Sprintf("Metadata extracted for %s", fileName))
		}
	}

	apiThumbnailPath := filepath.ToSlash(filepath.Join(config.ThumbnailDir, thumbnailName))

	// Identify performers from filename
	var identifiedPerformers []string
	lowerFileName := strings.ToLower(fileName)
	for _, pName := range performerNames {
		if strings.Contains(lowerFileName, strings.ToLower(pName)) {
			identifiedPerformers = append(identifiedPerformers, pName)
		}
	}
	if len(identifiedPerformers) > 0 {
		appstatus.EmitInfo("[Info]", fmt.Sprintf("Found performers in %s: %s", fileName, strings.Join(identifiedPerformers, ", ")))
	}

	return models.VideoInfo{Name: fileName, Thumbnail: apiThumbnailPath, Metadata: metadata, Performers: identifiedPerformers}
}

func UpdatePerformerSceneCounts() {
	appstatus.EmitInfo("[Task]", "Starting performer scene count update")
	performerSceneCounts := make(map[string]int)

	// Count scenes per performer (read lock for cache access)
	cacheMutex.RLock()
	for _, video := range VideoCache {
		for _, pName := range video.Performers {
			performerSceneCounts[pName]++
		}
	}
	cacheMutex.RUnlock()

	appstatus.EmitInfo("[Info]", fmt.Sprintf("Found %d performers with scenes", len(performerSceneCounts)))

	if len(performerSceneCounts) == 0 {
		appstatus.EmitInfo("[Task]", "No performers to update")
		return
	}

	// OPTIMIZED: Batch fetch all performers in one query
	performerNames := make([]string, 0, len(performerSceneCounts))
	for pName := range performerSceneCounts {
		performerNames = append(performerNames, pName)
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(performerNames))
	args := make([]interface{}, len(performerNames))
	for i, name := range performerNames {
		placeholders[i] = "?"
		args[i] = name
	}

	query := fmt.Sprintf("SELECT name, data FROM performers WHERE name IN (%s)", strings.Join(placeholders, ","))
	rows, err := database.GetDB().Query(query, args...)
	if err != nil {
		msg := fmt.Sprintf("Error batch querying performers: %v", err)
		appstatus.EmitError("[Error]", msg)
		log.Printf(msg)
		return
	}
	defer rows.Close()

	// Parse all performers
	performersToUpdate := make(map[string]models.Performer)
	for rows.Next() {
		var name, performerJSON string
		if err := rows.Scan(&name, &performerJSON); err != nil {
			log.Printf("Error scanning performer: %v", err)
			continue
		}

		var performer models.Performer
		if err := json.Unmarshal([]byte(performerJSON), &performer); err != nil {
			log.Printf("Error parsing performer %s: %v", name, err)
			continue
		}

		performer.SceneCount = performerSceneCounts[name]
		performersToUpdate[name] = performer
	}

	appstatus.EmitInfo("[Info]", fmt.Sprintf("Updating %d performers in batch", len(performersToUpdate)))

	// OPTIMIZED: Batch update using transaction
	tx, err := database.GetDB().Begin()
	if err != nil {
		msg := fmt.Sprintf("Error starting transaction: %v", err)
		appstatus.EmitError("[Error]", msg)
		log.Printf(msg)
		return
	}

	stmt, err := tx.Prepare("UPDATE performers SET data = ? WHERE name = ?")
	if err != nil {
		tx.Rollback()
		msg := fmt.Sprintf("Error preparing statement: %v", err)
		appstatus.EmitError("[Error]", msg)
		log.Printf(msg)
		return
	}
	defer stmt.Close()

	updateCount := 0
	totalCount := len(performersToUpdate)
	for name, performer := range performersToUpdate {
		updatedJSON, err := json.Marshal(performer)
		if err != nil {
			log.Printf("Error marshaling performer %s: %v", name, err)
			continue
		}

		_, err = stmt.Exec(string(updatedJSON), name)
		if err != nil {
			log.Printf("Error updating performer %s: %v", name, err)
			continue
		}

		updateCount++
		if updateCount%10 == 0 || updateCount == totalCount {
			appstatus.EmitInfo("[Progress]", fmt.Sprintf("Updated scene counts for %d/%d performers", updateCount, totalCount))
		}
	}

	if err := tx.Commit(); err != nil {
		msg := fmt.Sprintf("Error committing transaction: %v", err)
		appstatus.EmitError("[Error]", msg)
		log.Printf(msg)
		return
	}

	appstatus.EmitInfo("[Task]", fmt.Sprintf("Completed performer scene count update (%d updated)", updateCount))
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
