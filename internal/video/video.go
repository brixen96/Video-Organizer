package video

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"video-organizer/internal/config"
	"video-organizer/internal/database"
	"video-organizer/internal/models"
)

var VideoCache map[string]models.VideoInfo

func InitializeVideoCache() {
	log.Println("Initializing video cache...")
	VideoCache = make(map[string]models.VideoInfo)

	// Get all performer names from the database
	performerNames, err := database.GetAllPerformerNames()
	if err != nil {
		log.Fatalf("Failed to get performer names for cache initialization: %v", err)
	}

	files, err := os.ReadDir(config.VideoDir)
	if err != nil {
		log.Fatalf("Failed to read video directory for cache initialization: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".mp4" || ext == ".mkv" || ext == ".avi" || ext == ".mov" {
				videoPath := filepath.Join(config.VideoDir, file.Name())
				videoInfo := GenerateVideoInfo(file.Name(), videoPath, performerNames)
				VideoCache[file.Name()] = videoInfo
			}
		}
	}
	log.Printf("Video cache initialized with %d videos.", len(VideoCache))

	// Update performer scene counts in the database
	UpdatePerformerSceneCounts()
}

func GetVideoCache() map[string]models.VideoInfo {
	return VideoCache
}

func GenerateVideoInfo(fileName, videoPath string, performerNames []string) models.VideoInfo {
	thumbnailName := fmt.Sprintf("%s.jpg", strings.TrimSuffix(fileName, filepath.Ext(fileName)))
	thumbnailPath := filepath.Join(config.ThumbnailDir, thumbnailName)

	// Generate thumbnail if it doesn't exist
	if _, err := os.Stat(thumbnailPath); os.IsNotExist(err) {
		cmd := exec.Command("ffmpeg", "-i", videoPath, "-ss", "00:00:01.000", "-vframes", "1", thumbnailPath)
		err := cmd.Run()
		if err != nil {
			log.Println("Failed to generate thumbnail for", fileName, ":", err)
		}
	}

	// Get video metadata using ffprobe
	var metadata models.FFProbeOutput
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", videoPath)
	output, err := cmd.Output()
	if err != nil {
		log.Println("Failed to get metadata for", fileName, ":", err)
	} else {
		if err := json.Unmarshal(output, &metadata); err != nil {
			log.Println("Failed to unmarshal metadata for", fileName, ":", err)
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

	return models.VideoInfo{Name: fileName, Thumbnail: apiThumbnailPath, Metadata: metadata, Performers: identifiedPerformers}
}

func UpdatePerformerSceneCounts() {
	performerSceneCounts := make(map[string]int)

	for _, video := range VideoCache {
		for _, pName := range video.Performers {
			performerSceneCounts[pName]++
		}
	}

	// Update database
	for pName, count := range performerSceneCounts {
		performerJSON := ""
		err := database.GetDB().QueryRow("SELECT data FROM performers WHERE name = ?", pName).Scan(&performerJSON)
		if err == sql.ErrNoRows {
			log.Printf("Performer %s not found in DB, skipping scene count update.", pName)
			continue
		} else if err != nil {
			log.Printf("Error querying performer %s for scene count update: %v", pName, err)
			continue
		}

		var performer models.Performer
		if err := json.Unmarshal([]byte(performerJSON), &performer); err != nil {
			log.Printf("Error unmarshaling performer %s for scene count update: %v", pName, err)
			continue
		}

		performer.SceneCount = count
		updatedPerformerJSON, err := json.Marshal(performer)
		if err != nil {
			log.Printf("Error marshaling updated performer %s for scene count: %v", pName, err)
			continue
		}

		_, err = database.GetDB().Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedPerformerJSON), pName)
		if err != nil {
			log.Printf("Error updating scene count for performer %s: %v", pName, err)
		}
	}
	log.Println("Performer scene counts updated.")
}

