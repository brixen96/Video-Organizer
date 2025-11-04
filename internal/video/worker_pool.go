package video

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"video-organizer/internal/appstatus"
	"video-organizer/internal/config"
	"video-organizer/internal/models"
)

// VideoProcessingJob represents a video to be processed
type VideoProcessingJob struct {
	FileName       string
	VideoPath      string
	PerformerNames []string
}

// VideoProcessingResult represents the result of processing a video
type VideoProcessingResult struct {
	VideoInfo models.VideoInfo
	Error     error
}

// WorkerPool manages concurrent video processing
type WorkerPool struct {
	jobs    chan VideoProcessingJob
	results chan VideoProcessingResult
	wg      sync.WaitGroup
	workers int
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		jobs:    make(chan VideoProcessingJob, workers*2), // Buffered channel
		results: make(chan VideoProcessingResult, workers*2),
		workers: workers,
	}
}

// Start launches the worker pool
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker processes video jobs
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for job := range wp.jobs {
		videoInfo := wp.processVideo(job)
		wp.results <- VideoProcessingResult{
			VideoInfo: videoInfo,
			Error:     nil,
		}
	}
}

// processVideo handles the actual video processing
func (wp *WorkerPool) processVideo(job VideoProcessingJob) models.VideoInfo {
	// Generate thumbnail
	thumbnailName := fmt.Sprintf("%s.jpg", filepath.Base(job.FileName[:len(job.FileName)-len(filepath.Ext(job.FileName))]))
	thumbnailPath := filepath.Join(config.ThumbnailDir, thumbnailName)

	generateThumbnail(job.VideoPath, thumbnailPath, job.FileName)

	// Extract metadata
	metadata := extractMetadata(job.VideoPath, job.FileName)

	// Identify performers
	performers := identifyPerformers(job.FileName, job.PerformerNames)

	apiThumbnailPath := filepath.ToSlash(filepath.Join(config.ThumbnailDir, thumbnailName))

	return models.VideoInfo{
		Name:       job.FileName,
		Thumbnail:  apiThumbnailPath,
		Metadata:   metadata,
		Performers: performers,
	}
}

// Submit adds a job to the worker pool
func (wp *WorkerPool) Submit(job VideoProcessingJob) {
	wp.jobs <- job
}

// Close shuts down the worker pool
func (wp *WorkerPool) Close() {
	close(wp.jobs)
	wp.wg.Wait()
	close(wp.results)
}

// Results returns the results channel
func (wp *WorkerPool) Results() <-chan VideoProcessingResult {
	return wp.results
}

// Helper functions extracted from GenerateVideoInfo for reuse

func generateThumbnail(videoPath, thumbnailPath, fileName string) {
	// Check if thumbnail already exists using os.Stat
	if _, err := os.Stat(thumbnailPath); err == nil {
		return // Thumbnail exists
	}

	appstatus.EmitInfo("[Progress]", fmt.Sprintf("Generating thumbnail for %s", fileName))
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-ss", "00:00:01.000", "-vframes", "1", thumbnailPath)
	if err := cmd.Run(); err != nil {
		msg := fmt.Sprintf("Failed to generate thumbnail for %s: %v", fileName, err)
		appstatus.EmitWarning("[Warning]", msg)
	} else {
		appstatus.EmitInfo("[Info]", fmt.Sprintf("Thumbnail generated for %s", fileName))
	}
}

func extractMetadata(videoPath, fileName string) models.FFProbeOutput {
	var metadata models.FFProbeOutput

	appstatus.EmitInfo("[Progress]", fmt.Sprintf("Extracting metadata for %s", fileName))
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", videoPath)
	output, err := cmd.Output()
	if err != nil {
		msg := fmt.Sprintf("Failed to get metadata for %s: %v", fileName, err)
		appstatus.EmitWarning("[Warning]", msg)
		return metadata
	}

	if err := json.Unmarshal(output, &metadata); err != nil {
		msg := fmt.Sprintf("Failed to parse metadata for %s: %v", fileName, err)
		appstatus.EmitWarning("[Warning]", msg)
	} else {
		appstatus.EmitInfo("[Info]", fmt.Sprintf("Metadata extracted for %s", fileName))
	}

	return metadata
}

func identifyPerformers(fileName string, performerNames []string) []string {
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

	return identifiedPerformers
}
