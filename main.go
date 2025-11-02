package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const videoDir = "C:\\Users\\Brix-PC\\Downloads\\Anissa Kate"
const thumbnailDir = "frontend/.thumbnails"
const performerFoldersDir = "frontend/.performers"
const logFile = "app.log"
const oldLogsDir = "old_logs"

var videoCache map[string]VideoInfo

// FFProbe structs

type FFProbeOutput struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

type Stream struct {
	CodecType string `json:"codec_type"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

type Format struct {
	Duration string `json:"duration"`
	Size     string `json:"size"`
}

type RenameRequest struct {
	OldName string `json:"oldName"`
	NewName string `json:"newName"`
}

type ChatMessage struct {
	Message string `json:"message"`
}

		type Performer struct {
			// Fields from Adultdatalink API sample
			Appearance          map[string]interface{} `json:"appearance,omitempty"`
			Performances        map[string]interface{} `json:"performances,omitempty"`
			SocialMedia         map[string]interface{} `json:"social_media,omitempty"`
			PlatformViews       map[string]interface{} `json:"platform_views,omitempty"`
			PlatformVideoCounts map[string]interface{} `json:"platform_video_counts,omitempty"`
			PlatformProfileCounts map[string]interface{} `json:"platform_profile_counts,omitempty"`
			Tags                []string               `json:"tags,omitempty"`
			ExternalLinks       []map[string]string    `json:"external_links,omitempty"`
			Bios                map[string]string      `json:"bios,omitempty"`
			OfficialWebsite     string                 `json:"official_website,omitempty"`
			FeatureDancer       string                 `json:"feature_dancer,omitempty"`
			DateOfBirth         string                 `json:"date_of_birth,omitempty"`
			Age                 string                 `json:"age,omitempty"`
			SexualOrientation   string                 `json:"sexual_orientation,omitempty"`
			AstrologicalSign    string                 `json:"astrological_sign,omitempty"`
			Profession          string                 `json:"profession,omitempty"`
			CareerStatus        string                 `json:"career_status,omitempty"`
			CareerStart         string                 `json:"career_start,omitempty"`
			CareerEnd           string                 `json:"career_end,omitempty"`
			DateOfDeath         string                 `json:"date_of_death,omitempty"`
			PlaceOfBirth        string                 `json:"place_of_birth,omitempty"`
			Nationality         string                 `json:"nationality,omitempty"`
			Rank                string                 `json:"rank,omitempty"`
			Country             string                 `json:"country,omitempty"`
			Avatar              string                 `json:"avatar,omitempty"`
			Subscribers         int                    `json:"subscribers,omitempty"`
			Rating              int                    `json:"rating,omitempty"`
			TotalViews          int                    `json:"total_views,omitempty"`
			TotalVideoCount     int                    `json:"total_video_count,omitempty"`
			TotalPlatformHits   int                    `json:"total_platform_hits,omitempty"`
			Aliases             []string               `json:"aliases,omitempty"`
			ImageURL            string                 `json:"image_url,omitempty"`
	
			// Our top-level fields
			Name       string                 `json:"name"`
			SceneCount int                    `json:"scene_count"`
			Previews   []string               `json:"previews,omitempty"`
			DefaultPreview string             `json:"default_preview,omitempty"`
			Zoo        string                 `json:"zoo,omitempty"`
		}
type VideoInfo struct {
	Name      string        `json:"name"`
	Thumbnail string        `json:"thumbnail"`
	Metadata  FFProbeOutput `json:"metadata"`
	Performers []string     `json:"performers,omitempty"`
}

var db *sql.DB

var defaultPerformer Performer = Performer{
	Appearance: map[string]interface{}{
		"ethnicity":         "Undefined",
		"boobs":             "Undefined",
		"bust":              "Undefined",
		"cup":               "Undefined",
		"bra":               "Undefined",
		"waist":             "Undefined",
		"hip":               "Undefined",
		"butt":              "Undefined",
		"height":            "Undefined",
		"weight":            "Undefined",
		"hair_color":        "Undefined",
		"eye_color":         "Undefined",
		"piercings":         "Undefined",
		"piercing_locations": "Undefined",
		"tattoos":           "Undefined",
		"tattoo_locations":  "Undefined",
		"shoe_size":         "Undefined",
		"body_type":         "Undefined",
		"underarm_hair":     "Undefined",
		"pubic_hair":        "Undefined",
	},
	Performances:        make(map[string]interface{}),
	SocialMedia:         make(map[string]interface{}),
	PlatformViews:       make(map[string]interface{}),
	PlatformVideoCounts: make(map[string]interface{}),
	PlatformProfileCounts: make(map[string]interface{}),
	Tags:                []string{},
	ExternalLinks:       []map[string]string{},
	Bios:                make(map[string]string),
	OfficialWebsite:     "Undefined",
	FeatureDancer:       "Undefined",
	DateOfBirth:         "Undefined",
	Age:                 "Undefined",	
	SexualOrientation:   "Undefined",
	AstrologicalSign:    "Undefined",
	Profession:          "Undefined",
	CareerStatus:        "Undefined",
	CareerStart:         "Undefined",
	CareerEnd:           "Undefined",
	DateOfDeath:         "Undefined",
	PlaceOfBirth:        "Undefined",
	Nationality:         "Undefined",
	Rank:                "Undefined",
	Country:             "Undefined",
	Avatar:              "Undefined",
	Subscribers:         0,
	Rating:              0,
	TotalViews:          0,
	TotalVideoCount:     0,
	TotalPlatformHits:   0,
	Aliases:             []string{},
	ImageURL:            "Undefined",
	SceneCount:          0,
	Previews:            []string{},
	DefaultPreview:      "",
	Zoo:                 "undefined",
}

func setupLogging() {
	// Ensure old_logs directory exists
	if _, err := os.Stat(oldLogsDir); os.IsNotExist(err) {
		os.Mkdir(oldLogsDir, 0755)
	}

	// Rotate previous log file if it exists
	if _, err := os.Stat(logFile); err == nil {
		backupName := filepath.Join(oldLogsDir, fmt.Sprintf("app_%s.log", time.Now().Format("20060102_150405")))
		os.Rename(logFile, backupName)
	}

	// Create new log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func videosHandler(w http.ResponseWriter, r *http.Request) {
	var videos []VideoInfo
	for _, video := range videoCache {
		videos = append(videos, video)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videos)
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./video_organizer.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS performers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		data TEXT
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("Failed to create performers table: %v", err)
	}
	log.Println("Database initialized and performers table ensured.")
}

func updateExistingPerformersSchema() {
	log.Println("Updating existing performer schema in database...")
	rows, err := db.Query("SELECT name, data FROM performers")
	if err != nil {
		log.Printf("Failed to query existing performers for schema update: %v", err)
		return
	}
	defer rows.Close()

	var performersToUpdate []Performer
	for rows.Next() {
		var name string
		var oldJSON string
		if err := rows.Scan(&name, &oldJSON); err != nil {
			log.Printf("Failed to scan performer data for schema update: %v", err)
			continue
		}

		var oldPerformerMap map[string]interface{}
		if err := json.Unmarshal([]byte(oldJSON), &oldPerformerMap); err != nil {
			log.Printf("Failed to unmarshal old performer JSON for %s: %v", name, err)
			continue
		}

		newPerformer := defaultPerformer // Start with the default template
		newPerformer.Name = name         // Keep the existing name

		// Unmarshal old JSON directly into newPerformer to preserve existing fields
		if err := json.Unmarshal([]byte(oldJSON), &newPerformer); err != nil {
			log.Printf("Failed to unmarshal old JSON into newPerformer for %s: %v", name, err)
			continue
		}
		newPerformer.Name = name // Ensure name is preserved after unmarshaling

		performersToUpdate = append(performersToUpdate, newPerformer)
	}

	// Update database with new schema
	for _, p := range performersToUpdate {
		updatedJSON, err := json.Marshal(p)
		if err != nil {
			log.Printf("Failed to marshal updated performer %s to JSON: %v", p.Name, err)
			continue
		}
		_, err = db.Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedJSON), p.Name)
		if err != nil {
			log.Printf("Failed to update performer %s in database: %v", p.Name, err)
		}
	}
	log.Println("Performer schema update completed.")
}

func renameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("Invalid method for renameHandler:", r.Method)
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RenameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid request body for renameHandler:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Rename video file
	oldVideoPath := filepath.Join(videoDir, req.OldName)
	newVideoPath := filepath.Join(videoDir, req.NewName)
	if err := os.Rename(oldVideoPath, newVideoPath); err != nil {
		log.Println("Failed to rename video", req.OldName, ":", err)
		http.Error(w, fmt.Sprintf("Failed to rename video: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println("Successfully renamed video from", req.OldName, "to", req.NewName)

	// Rename thumbnail file
	oldThumbName := fmt.Sprintf("%s.jpg", strings.TrimSuffix(req.OldName, filepath.Ext(req.OldName)))
	newThumbName := fmt.Sprintf("%s.jpg", strings.TrimSuffix(req.NewName, filepath.Ext(req.NewName)))
	oldThumbPath := filepath.Join(thumbnailDir, oldThumbName)
	newThumbPath := filepath.Join(thumbnailDir, newThumbName)
	if _, err := os.Stat(oldThumbPath); err == nil {
		if err := os.Rename(oldThumbPath, newThumbPath); err != nil {
			log.Println("Failed to rename thumbnail for", req.OldName, ":", err)
		}
	}

	// Update cache
	delete(videoCache, req.OldName)

	performerNames, err := getAllPerformerNames()
	if err != nil {
		log.Printf("Failed to get performer names for rename handler: %v", err)
		http.Error(w, "Failed to update performer associations", http.StatusInternalServerError)
		return
	}

	videoCache[req.NewName] = generateVideoInfo(req.NewName, newVideoPath, performerNames)

	// Re-evaluate and update performer scene counts
	updatePerformerSceneCounts()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("Invalid method for chatHandler:", r.Method)
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg ChatMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Println("Invalid request body for chatHandler:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic AI response logic
	reply := "I am a simple AI. I don't have much to say yet."
	if strings.Contains(strings.ToLower(msg.Message), "hello") {
		reply = "Hello there! How can I help you?"
	} else if strings.Contains(strings.ToLower(msg.Message), "naming convention") {
		reply = "A good naming convention is `YYYY-MM-DD_Event-Name.mp4`. This helps with sorting and identification."
	}
	log.Println("Chat message from user:", msg.Message, "; AI reply:", reply)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"reply": reply})
}

func performersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var performer Performer
		if err := json.NewDecoder(r.Body).Decode(&performer); err != nil {
			log.Println("Invalid request body for addPerformer:", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		performerJSON, err := json.Marshal(performer)
		if err != nil {
			log.Println("Failed to marshal performer to JSON:", err)
			http.Error(w, "Failed to process performer data", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO performers(name, data) VALUES(?, ?)", performer.Name, string(performerJSON))
		if err != nil {
			log.Println("Failed to insert performer into database:", err)
			http.Error(w, "Failed to add performer", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Performer added successfully"})

	case http.MethodGet:
		rows, err := db.Query("SELECT data FROM performers")
		if err != nil {
			log.Println("Failed to query performers from database:", err)
			http.Error(w, "Failed to retrieve performers", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var performers []Performer
		for rows.Next() {
			var performerJSON string
			if err := rows.Scan(&performerJSON); err != nil {
				log.Println("Failed to scan performer data:", err)
				continue
			}
			var performer Performer
			if err := json.Unmarshal([]byte(performerJSON), &performer); err != nil {
				log.Println("Failed to unmarshal performer JSON:", err)
				continue
			}
			performers = append(performers, performer)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(performers)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func performerDetailsHandler(w http.ResponseWriter, r *http.Request) {
	performerName := strings.TrimPrefix(r.URL.Path, "/api/performers/")

	// Handle set-default-preview action
	if strings.HasSuffix(performerName, "/set-default-preview") {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		performerName = strings.TrimSuffix(performerName, "/set-default-preview")
		if performerName == "" {
			http.Error(w, "Performer name not specified", http.StatusBadRequest)
			return
		}

		var payload struct {
			PreviewURL string `json:"previewUrl"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var performerJSON string
		err := db.QueryRow("SELECT data FROM performers WHERE name = ?", performerName).Scan(&performerJSON)
		if err == sql.ErrNoRows {
			http.Error(w, "Performer not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Failed to query performer %s from database: %v", performerName, err)
			http.Error(w, "Failed to retrieve performer details", http.StatusInternalServerError)
			return
		}

		var performer Performer
		if err := json.Unmarshal([]byte(performerJSON), &performer); err != nil {
			log.Printf("Failed to unmarshal performer JSON for %s: %v", performerName, err)
			http.Error(w, "Failed to process performer data", http.StatusInternalServerError)
			return
		}

		performer.DefaultPreview = payload.PreviewURL

		updatedPerformerJSON, err := json.Marshal(performer)
		if err != nil {
			log.Printf("Failed to marshal updated performer %s to JSON: %v", performerName, err)
			http.Error(w, "Failed to update performer data", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedPerformerJSON), performerName)
		if err != nil {
			log.Printf("Failed to update performer %s in database: %v", performerName, err)
			http.Error(w, "Failed to save default preview", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Default preview updated successfully"})
		return
	}

	// Handle set-zoo action
	if strings.HasSuffix(performerName, "/set-zoo") {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		performerName = strings.TrimSuffix(performerName, "/set-zoo")
		if performerName == "" {
			http.Error(w, "Performer name not specified", http.StatusBadRequest)
			return
		}

		var payload struct {
			Zoo string `json:"zoo"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var performerJSON string
		err := db.QueryRow("SELECT data FROM performers WHERE name = ?", performerName).Scan(&performerJSON)
		if err == sql.ErrNoRows {
			http.Error(w, "Performer not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Failed to query performer %s from database: %v", performerName, err)
			http.Error(w, "Failed to retrieve performer details", http.StatusInternalServerError)
			return
		}

		var performer Performer
		if err := json.Unmarshal([]byte(performerJSON), &performer); err != nil {
			log.Printf("Failed to unmarshal performer JSON for %s: %v", performerName, err)
			http.Error(w, "Failed to process performer data", http.StatusInternalServerError)
			return
		}

		performer.Zoo = payload.Zoo

		updatedPerformerJSON, err := json.Marshal(performer)
		if err != nil {
			log.Printf("Failed to marshal updated performer %s to JSON: %v", performerName, err)
			http.Error(w, "Failed to update performer data", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedPerformerJSON), performerName)
		if err != nil {
			log.Printf("Failed to update performer %s in database: %v", performerName, err)
			http.Error(w, "Failed to save zoo status", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Zoo status updated successfully"})
		return
	}

	// Existing logic for GET performer details
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if performerName == "" {
		http.Error(w, "Performer name not specified", http.StatusBadRequest)
		return
	}

	var performerJSON string
	err := db.QueryRow("SELECT data FROM performers WHERE name = ?", performerName).Scan(&performerJSON)
	if err == sql.ErrNoRows {
		http.Error(w, "Performer not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println("Failed to query performer from database:", err)
		http.Error(w, "Failed to retrieve performer details", http.StatusInternalServerError)
		return
	}

	var performer Performer
	if err := json.Unmarshal([]byte(performerJSON), &performer); err != nil {
		log.Println("Failed to unmarshal performer JSON:", err)
		http.Error(w, "Failed to process performer data", http.StatusInternalServerError)
		return
	}

	// TODO: Associate scenes with the performer

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(performer)
}

func setDefaultPreviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	performerName := strings.TrimPrefix(r.URL.Path, "/api/performers/")
	performerName = strings.TrimSuffix(performerName, "/set-default-preview")
	if performerName == "" {
		http.Error(w, "Performer name not specified", http.StatusBadRequest)
		return
	}

	var payload struct {
		PreviewURL string `json:"previewUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return	
	}

	var performerJSON string
	err := db.QueryRow("SELECT data FROM performers WHERE name = ?", performerName).Scan(&performerJSON)
	if err == sql.ErrNoRows {
		http.Error(w, "Performer not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Failed to query performer %s from database: %v", performerName, err)
		http.Error(w, "Failed to retrieve performer details", http.StatusInternalServerError)
		return
	}

	var performer Performer
	if err := json.Unmarshal([]byte(performerJSON), &performer); err != nil {
		log.Printf("Failed to unmarshal performer JSON for %s: %v", performerName, err)
		http.Error(w, "Failed to process performer data", http.StatusInternalServerError)
		return
	}

	performer.DefaultPreview = payload.PreviewURL

	updatedPerformerJSON, err := json.Marshal(performer)
	if err != nil {
		log.Printf("Failed to marshal updated performer %s to JSON: %v", performerName, err)
		http.Error(w, "Failed to update performer data", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedPerformerJSON), performerName)
	if err != nil {
		log.Printf("Failed to update performer %s in database: %v", performerName, err)
		http.Error(w, "Failed to save default preview", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Default preview updated successfully"})
}

func updatePerformerPreviewsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go updatePerformerPreviewsTask() // Run the task in a goroutine to avoid blocking the HTTP request

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Update performer previews task started."})
}

func videoStreamHandler(w http.ResponseWriter, r *http.Request) {
	videoName := strings.TrimPrefix(r.URL.Path, "/video/")
	if videoName == "" {
		log.Println("Video name not specified in stream request")
		http.Error(w, "Video name not specified", http.StatusBadRequest)
		return
	}

	videoPath := filepath.Join(videoDir, videoName)

	// Serve the video file
	http.ServeFile(w, r, videoPath)
}

func previousLogsHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(oldLogsDir)
	if err != nil {
		log.Println("Failed to read old logs directory:", err)
		http.Error(w, "Failed to read old logs directory", http.StatusInternalServerError)
		return
	}

	var logFiles []string
	for _, file := range files {
		if !file.IsDir() {
			logFiles = append(logFiles, file.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logFiles)
}

func previousLogFileHandler(w http.ResponseWriter, r *http.Request) {
	fileName := strings.TrimPrefix(r.URL.Path, "/api/logs/previous/")
	if fileName == "" {
		log.Println("Log file name not specified in request")
		http.Error(w, "Log file name not specified", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(oldLogsDir, fileName)

	// Check if the file exists and is within the old_logs directory
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Log file not found: %s", filePath)
		http.Error(w, "Log file not found", http.StatusNotFound)
		return
	}

	// Serve the log file
	http.ServeFile(w, r, filePath)
}

func currentLogsHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the current app.log file
	http.ServeFile(w, r, logFile)
}

func performerPreviewHandler(w http.ResponseWriter, r *http.Request) {
	previewPath := strings.TrimPrefix(r.URL.Path, "/performer-previews/") // Updated prefix
	if previewPath == "" {
		http.Error(w, "Preview path not specified", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(performerFoldersDir, previewPath)

	// Normalize paths for consistent comparison
	normalizedFullPath := filepath.ToSlash(fullPath)
	normalizedPerformerFoldersDir := filepath.ToSlash(performerFoldersDir)

	// Security check: ensure the path is within the performerFoldersDir
	if !strings.HasPrefix(normalizedFullPath, normalizedPerformerFoldersDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, fullPath)
}

func getAllPerformerNames() ([]string, error) {
	rows, err := db.Query("SELECT name FROM performers")
	if err != nil {
		return nil, fmt.Errorf("failed to query performer names: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Println("Failed to scan performer name:", err)
			continue
		}
		names = append(names, name)
	}
	return names, nil
}

func fetchPerformerData(performerName string) ([]byte, error) {
	apiURL := fmt.Sprintf("https://api.adultdatalink.com/pornstar/pornstar-data?name=%s", url.QueryEscape(performerName))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request for Adultdatalink API for %s: %w", performerName, err)
	}
	req.Header.Add("Authorization", "raA8-fkxPODxiwx7WM05wZFy9LBtEmm7g3VGsJ0MjDE") // Add Authorization header

	resp, err := http.DefaultClient.Do(req) // Use DefaultClient to send the request
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from Adultdatalink API for %s: %w", performerName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Adultdatalink API returned non-OK status for %s: %s", performerName, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)

	// log.Printf("DEBUG: Raw Adultdatalink API response for %s: %s", performerName, string(body))

	if err != nil {
		return nil, fmt.Errorf("failed to read Adultdatalink API response body for %s: %w", performerName, err)
	}

	return body, nil
}

func initializeVideoCache() {
	log.Println("Initializing video cache...")
	videoCache = make(map[string]VideoInfo)

	// Get all performer names from the database
	performerNames, err := getAllPerformerNames()
	if err != nil {
		log.Fatalf("Failed to get performer names for cache initialization: %v", err)
	}

	files, err := os.ReadDir(videoDir)
	if err != nil {
		log.Fatalf("Failed to read video directory for cache initialization: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".mp4" || ext == ".mkv" || ext == ".avi" || ext == ".mov" {
				videoPath := filepath.Join(videoDir, file.Name())
				videoInfo := generateVideoInfo(file.Name(), videoPath, performerNames)
				videoCache[file.Name()] = videoInfo
			}
		}
	}
	log.Printf("Video cache initialized with %d videos.", len(videoCache))

	// Update performer scene counts in the database
	updatePerformerSceneCounts()
}

func generateVideoInfo(fileName, videoPath string, performerNames []string) VideoInfo {
		thumbnailName := fmt.Sprintf("%s.jpg", strings.TrimSuffix(fileName, filepath.Ext(fileName)))
		thumbnailPath := filepath.Join(thumbnailDir, thumbnailName)

		// Generate thumbnail if it doesn't exist
		if _, err := os.Stat(thumbnailPath); os.IsNotExist(err) {
			cmd := exec.Command("ffmpeg", "-i", videoPath, "-ss", "00:00:01.000", "-vframes", "1", thumbnailPath)
			err := cmd.Run()
			if err != nil {
				log.Println("Failed to generate thumbnail for", fileName, ":", err)
			}
		}

		// Get video metadata using ffprobe
		var metadata FFProbeOutput
		cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", videoPath)
		output, err := cmd.Output()
		if err != nil {
			log.Println("Failed to get metadata for", fileName, ":", err)
		} else {
			if err := json.Unmarshal(output, &metadata); err != nil {
				log.Println("Failed to unmarshal metadata for", fileName, ":", err)
			}
		}

		apiThumbnailPath := filepath.ToSlash(filepath.Join(".thumbnails", thumbnailName))

		// Identify performers from filename
		var identifiedPerformers []string
		lowerFileName := strings.ToLower(fileName)
		for _, pName := range performerNames {
			if strings.Contains(lowerFileName, strings.ToLower(pName)) {
				identifiedPerformers = append(identifiedPerformers, pName)
			}
		}

				return VideoInfo{Name: fileName, Thumbnail: apiThumbnailPath, Metadata: metadata, Performers: identifiedPerformers}

		}

		

		func updatePerformerSceneCounts() {

			performerSceneCounts := make(map[string]int)

			for _, video := range videoCache {

				for _, pName := range video.Performers {

					performerSceneCounts[pName]++

				}

			}

		

			// Update database

			for pName, count := range performerSceneCounts {

				performerJSON := ""

				err := db.QueryRow("SELECT data FROM performers WHERE name = ?", pName).Scan(&performerJSON)

				if err == sql.ErrNoRows {

					log.Printf("Performer %s not found in DB, skipping scene count update.", pName)

					continue

				} else if err != nil {

					log.Printf("Error querying performer %s for scene count update: %v", pName, err)

					continue

				}

		

				var performer Performer

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

		

				_, err = db.Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedPerformerJSON), pName)

				if err != nil {

					log.Printf("Error updating scene count for performer %s: %v", pName, err)

				}

			}

			log.Println("Performer scene counts updated.")

		}

		
func autoAddPerformersFromFolders() {
	log.Println("Checking for new performers in folders...")
	entries, err := os.ReadDir(performerFoldersDir)
	if err != nil {
		log.Printf("Failed to read performer folders directory: %v", err)
		return
	}

	existingPerformers, err := getAllPerformerNames()
	if err != nil {
		log.Printf("Failed to get existing performer names from DB: %v", err)
		return
	}

	existingPerformerMap := make(map[string]bool)
	for _, pName := range existingPerformers {
		existingPerformerMap[pName] = true
	}

	var wg sync.WaitGroup
	newPerformersToInsert := make(chan Performer, len(entries)) // Channel to collect new performers

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		performerName := entry.Name()
		if _, exists := existingPerformerMap[performerName]; exists {
			continue
		}

		wg.Add(1)
		go func(pName string) {
			defer wg.Done()

			log.Printf("New performer folder found: %s. Processing concurrently.", pName)

					                    newPerformer := Performer{} // Start with an empty Performer struct
										newPerformer.Name = pName        // Set the name
					
										// Check if performer already exists in DB and load their data
										var existingPerformerJSON string
										err := db.QueryRow("SELECT data FROM performers WHERE name = ?", pName).Scan(&existingPerformerJSON)
										if err != nil && err != sql.ErrNoRows {
											log.Printf("Error querying existing performer %s from database: %v", pName, err)
											return
										} else if err == sql.ErrNoRows { // Performer does not exist, initialize with default template
											newPerformer = defaultPerformer
											newPerformer.Name = pName // Ensure name is set
										} else { // Performer exists, unmarshal existing data
											if err := json.Unmarshal([]byte(existingPerformerJSON), &newPerformer); err != nil {
												log.Printf("Error unmarshaling existing performer %s data: %v", pName, err)
												return
											}
										}
					
										// Ensure scene count and previews are initialized if not loaded from DB
										if newPerformer.SceneCount == 0 {
											newPerformer.SceneCount = 0
										}
										if newPerformer.Previews == nil {
											newPerformer.Previews = []string{}
										}
					// Fetch performer data from Adultdatalink API
					performerDataBytes, err := fetchPerformerData(pName)
					if err != nil {
						log.Printf("Failed to fetch data for performer %s from Adultdatalink API: %v", pName, err)
					} else {
						var rawAPIResponse map[string]interface{}
						if err := json.Unmarshal(performerDataBytes, &rawAPIResponse); err != nil {
							log.Printf("Failed to unmarshal raw Adultdatalink API response for %s: %v", pName, err)
						} else {
							// Process Aliases
							if aliases, ok := rawAPIResponse["aliases"]; ok {
								if aliasesList, ok := aliases.([]interface{}); ok {
									var tempAliases []string
									for _, alias := range aliasesList {
										if aliasMap, ok := alias.(map[string]interface{}); ok {
											for k := range aliasMap {
												tempAliases = append(tempAliases, k)
											}
										} else if aliasStr, ok := alias.(string); ok {
											tempAliases = append(tempAliases, aliasStr)
										}
									}
									newPerformer.Aliases = tempAliases
								} else if aliasStr, ok := aliases.(string); ok {
									newPerformer.Aliases = []string{aliasStr}
								}
							}

							// Process Tags
							if tags, ok := rawAPIResponse["tags"]; ok {
								if tagsList, ok := tags.([]interface{}); ok {
									var tempTags []string
									for _, tag := range tagsList {
										if tagMap, ok := tag.(map[string]interface{}); ok {
											for k := range tagMap {
												tempTags = append(tempTags, k)
											}
										} else if tagStr, ok := tag.(string); ok {
											tempTags = append(tempTags, tagStr)
										}
									}
									newPerformer.Tags = tempTags
								} else if tagStr, ok := tags.(string); ok {
									newPerformer.Tags = []string{tagStr}
								}
							}

							// Populate other fields from rawAPIResponse only if they exist
							if imageURL, ok := rawAPIResponse["image_url"].(string); ok && imageURL != "" {
								newPerformer.ImageURL = imageURL
							}
							if appearance, ok := rawAPIResponse["appearance"].(map[string]interface{}); ok && len(appearance) > 0 {
								newPerformer.Appearance = appearance
							}
							if performances, ok := rawAPIResponse["performances"].(map[string]interface{}); ok && len(performances) > 0 {
								newPerformer.Performances = performances
							}
							if socialMedia, ok := rawAPIResponse["social_media"].(map[string]interface{}); ok && len(socialMedia) > 0 {
								newPerformer.SocialMedia = socialMedia
							}
							if platformViews, ok := rawAPIResponse["platform_views"].(map[string]interface{}); ok && len(platformViews) > 0 {
								newPerformer.PlatformViews = platformViews
							}
							if platformVideoCounts, ok := rawAPIResponse["platform_video_counts"].(map[string]interface{}); ok && len(platformVideoCounts) > 0 {
								newPerformer.PlatformVideoCounts = platformVideoCounts
							}
							if platformProfileCounts, ok := rawAPIResponse["platform_profile_counts"].(map[string]interface{}); ok && len(platformProfileCounts) > 0 {
								newPerformer.PlatformProfileCounts = platformProfileCounts
							}
							if externalLinks, ok := rawAPIResponse["external_links"].([]interface{}); ok && len(externalLinks) > 0 {
								// Convert []interface{} to []map[string]string
								var convertedLinks []map[string]string
								for _, link := range externalLinks {
									if linkMap, isMap := link.(map[string]interface{}); isMap {
										convertedLink := make(map[string]string)
										for k, v := range linkMap {
											if strVal, isString := v.(string); isString {
												convertedLink[k] = strVal
											}
										}
										convertedLinks = append(convertedLinks, convertedLink)
									}
								}
								newPerformer.ExternalLinks = convertedLinks
							}
															if bios, ok := rawAPIResponse["bios"]; ok {
																if biosMap, isMap := bios.(map[string]interface{}); isMap {
																	convertedBios := make(map[string]string)
																	for k, v := range biosMap {
																		if strVal, isString := v.(string); isString {
																			convertedBios[k] = strVal
																		}
																	}
																	newPerformer.Bios = convertedBios
																}
															};							if officialWebsite, ok := rawAPIResponse["official_website"].(string); ok && officialWebsite != "" {
								newPerformer.OfficialWebsite = officialWebsite
							}
							if featureDancer, ok := rawAPIResponse["feature_dancer"].(string); ok && featureDancer != "" {
								newPerformer.FeatureDancer = featureDancer
							}
							if dateOfBirth, ok := rawAPIResponse["date_of_birth"].(string); ok && dateOfBirth != "" {
								newPerformer.DateOfBirth = dateOfBirth
							}
							if age, ok := rawAPIResponse["age"].(string); ok && age != "" {
								newPerformer.Age = age
							}
							if sexualOrientation, ok := rawAPIResponse["sexual_orientation"].(string); ok && sexualOrientation != "" {
								newPerformer.SexualOrientation = sexualOrientation
							}
							if astrologicalSign, ok := rawAPIResponse["astrological_sign"].(string); ok && astrologicalSign != "" {
								newPerformer.AstrologicalSign = astrologicalSign
							}
							if profession, ok := rawAPIResponse["profession"].(string); ok && profession != "" {
								newPerformer.Profession = profession
							}
							if careerStatus, ok := rawAPIResponse["career_status"].(string); ok && careerStatus != "" {
								newPerformer.CareerStatus = careerStatus
							}
							if careerStart, ok := rawAPIResponse["career_start"].(string); ok && careerStart != "" {
								newPerformer.CareerStart = careerStart
							}
							if careerEnd, ok := rawAPIResponse["career_end"].(string); ok && careerEnd != "" {
								newPerformer.CareerEnd = careerEnd
							}
							if dateOfDeath, ok := rawAPIResponse["date_of_death"].(string); ok && dateOfDeath != "" {
								newPerformer.DateOfDeath = dateOfDeath
							}
							if placeOfBirth, ok := rawAPIResponse["place_of_birth"].(string); ok && placeOfBirth != "" {
								newPerformer.PlaceOfBirth = placeOfBirth
							}
							if nationality, ok := rawAPIResponse["nationality"].(string); ok && nationality != "" {
								newPerformer.Nationality = nationality
							}
							if rank, ok := rawAPIResponse["rank"].(string); ok && rank != "" {
								newPerformer.Rank = rank
							}
							if country, ok := rawAPIResponse["country"].(string); ok && country != "" {
								newPerformer.Country = country
							}
							if avatar, ok := rawAPIResponse["avatar"].(string); ok && avatar != "" {
								newPerformer.Avatar = avatar
							}
							if subscribers, ok := rawAPIResponse["subscribers"].(float64); ok {
								newPerformer.Subscribers = int(subscribers)
							}
							if rating, ok := rawAPIResponse["rating"].(float64); ok {
								newPerformer.Rating = int(rating)
							}
							if totalViews, ok := rawAPIResponse["total_views"].(float64); ok {
								newPerformer.TotalViews = int(totalViews)
							}
							if totalVideoCount, ok := rawAPIResponse["total_video_count"].(float64); ok {
								newPerformer.TotalVideoCount = int(totalVideoCount)
							}
							if totalPlatformHits, ok := rawAPIResponse["total_platform_hits"].(float64); ok {
								newPerformer.TotalPlatformHits = int(totalPlatformHits)
							}

							log.Printf("Successfully fetched Adultdatalink API data for %s.", pName)
						}
					}

			// Scan for previews (.mkv and image files)
			performerFolderPath := filepath.Join(performerFoldersDir, pName)
			performerEntries, err := os.ReadDir(performerFolderPath)
			if err != nil {
				log.Printf("Failed to read performer folder %s for previews: %v", performerFolderPath, err)
			} else {
				var previews []string
				for _, pEntry := range performerEntries {
					ext := strings.ToLower(filepath.Ext(pEntry.Name()))
					if !pEntry.IsDir() && (ext == ".mkv" || ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif") {
						relativePath, err := filepath.Rel(performerFoldersDir, filepath.Join(performerFolderPath, pEntry.Name()))
						if err != nil {
							log.Printf("Failed to get relative path for preview %s: %v", pEntry.Name(), err)
							continue
						}
						previews = append(previews, filepath.ToSlash(relativePath))
					}
				}
				newPerformer.Previews = previews
				log.Printf("Found %d previews for %s.", len(previews), pName)
			}
			newPerformersToInsert <- newPerformer // Send processed performer to channel
		}(performerName)
	}

	wg.Wait() // Wait for all goroutines to finish
	close(newPerformersToInsert) // Close the channel after all goroutines are done

	// Collect all new performers and insert them in a batch
	var performersToBatchInsert []Performer
	for p := range newPerformersToInsert {
		performersToBatchInsert = append(performersToBatchInsert, p)
	}

	if len(performersToBatchInsert) > 0 {
		log.Printf("Batch inserting %d new performers into database.", len(performersToBatchInsert))
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Failed to begin transaction for batch insert: %v", err)
			return
		}
		stmt, err := tx.Prepare("INSERT INTO performers(name, data) VALUES(?, ?)")
		if err != nil {
			log.Printf("Failed to prepare statement for batch insert: %v", err)
			tx.Rollback()
			return
		}
		defer stmt.Close()

		for _, p := range performersToBatchInsert {
			performerJSON, err := json.Marshal(p)
			if err != nil {
				log.Printf("Failed to marshal performer %s for batch insert: %v", p.Name, err)
				continue
			}
			_, err = stmt.Exec(p.Name, string(performerJSON))
			if err != nil {
				log.Printf("Failed to insert performer %s during batch insert: %v", p.Name, err)
				continue
			}
		}
		err = tx.Commit()
		if err != nil {
			log.Printf("Failed to commit transaction for batch insert: %v", err)
			return
		}
		log.Println("Batch insert of new performers completed.")
	}

	log.Println("Finished checking for new performers.")
}

func fetchPerformerMetadataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	performerName := strings.TrimPrefix(r.URL.Path, "/api/performers/")
	performerName = strings.TrimSuffix(performerName, "/fetch-metadata")
	if performerName == "" {
		http.Error(w, "Performer name not specified", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching metadata for performer: %s", performerName)

	// Fetch performer data from Adultdatalink API
	performerDataBytes, err := fetchPerformerData(performerName)
	if err != nil {
		log.Printf("Failed to fetch data for performer %s from Adultdatalink API: %v", performerName, err)
		http.Error(w, fmt.Sprintf("Failed to fetch metadata: %v", err), http.StatusInternalServerError)
		return
	}

	var rawAPIResponse map[string]interface{}
	if err := json.Unmarshal(performerDataBytes, &rawAPIResponse); err != nil {
		log.Printf("Failed to unmarshal raw Adultdatalink API response for %s: %v", performerName, err)
		http.Error(w, "Failed to process API response", http.StatusInternalServerError)
		return
	}

	newPerformer := defaultPerformer // Start with the default template
	newPerformer.Name = performerName

	// Process Aliases
	if aliases, ok := rawAPIResponse["aliases"]; ok {
		if aliasesList, ok := aliases.([]interface{}); ok {
			for _, alias := range aliasesList {
				if aliasMap, ok := alias.(map[string]interface{}); ok {
					for k := range aliasMap {
						newPerformer.Aliases = append(newPerformer.Aliases, k)
					}
				} else if aliasStr, ok := alias.(string); ok {
					newPerformer.Aliases = append(newPerformer.Aliases, aliasStr)
				}
			}
		} else if aliasStr, ok := aliases.(string); ok {
			newPerformer.Aliases = append(newPerformer.Aliases, aliasStr)
		}
	}

	// Process Tags
	if tags, ok := rawAPIResponse["tags"]; ok {
		if tagsList, ok := tags.([]interface{}); ok {
			for _, tag := range tagsList {
				if tagMap, ok := tag.(map[string]interface{}); ok {
					for k := range tagMap {
						newPerformer.Tags = append(newPerformer.Tags, k)
					}
				} else if tagStr, ok := tag.(string); ok {
					newPerformer.Tags = append(newPerformer.Tags, tagStr)
				}
			}
		} else if tagStr, ok := tags.(string); ok {
			newPerformer.Tags = append(newPerformer.Tags, tagStr)
		}
	}

	// Populate other fields from rawAPIResponse
	if imageURL, ok := rawAPIResponse["image_url"].(string); ok && imageURL != "" {
		newPerformer.ImageURL = imageURL
	}
	if appearance, ok := rawAPIResponse["appearance"].(map[string]interface{}); ok && len(appearance) > 0 {
		newPerformer.Appearance = appearance
	}
	if performances, ok := rawAPIResponse["performances"].(map[string]interface{}); ok && len(performances) > 0 {
		newPerformer.Performances = performances
	}
	if socialMedia, ok := rawAPIResponse["social_media"].(map[string]interface{}); ok && len(socialMedia) > 0 {
		newPerformer.SocialMedia = socialMedia
	}
	if platformViews, ok := rawAPIResponse["platform_views"].(map[string]interface{}); ok && len(platformViews) > 0 {
		newPerformer.PlatformViews = platformViews
	}
	if platformVideoCounts, ok := rawAPIResponse["platform_video_counts"].(map[string]interface{}); ok && len(platformVideoCounts) > 0 {
		newPerformer.PlatformVideoCounts = platformVideoCounts
	}
	if platformProfileCounts, ok := rawAPIResponse["platform_profile_counts"].(map[string]interface{}); ok && len(platformProfileCounts) > 0 {
		newPerformer.PlatformProfileCounts = platformProfileCounts
	}
	if externalLinks, ok := rawAPIResponse["external_links"].([]interface{}); ok && len(externalLinks) > 0 {
		// Convert []interface{} to []map[string]string
		var convertedLinks []map[string]string
		for _, link := range externalLinks {
			if linkMap, isMap := link.(map[string]interface{}); isMap {
				convertedLink := make(map[string]string)
				for k, v := range linkMap {
					if strVal, isString := v.(string); isString {
						convertedLink[k] = strVal
					}
				}
				convertedLinks = append(convertedLinks, convertedLink)
			}
		}
		newPerformer.ExternalLinks = convertedLinks
	}
	if bios, ok := rawAPIResponse["bios"]; ok {
		if biosMap, isMap := bios.(map[string]interface{}); isMap {
			convertedBios := make(map[string]string)
			for k, v := range biosMap {
				if strVal, isString := v.(string); isString {
					convertedBios[k] = strVal
				}
			}
			newPerformer.Bios = convertedBios
		}
	}
	if officialWebsite, ok := rawAPIResponse["official_website"].(string); ok && officialWebsite != "" {
		newPerformer.OfficialWebsite = officialWebsite
	}
	if featureDancer, ok := rawAPIResponse["feature_dancer"].(string); ok && featureDancer != "" {
		newPerformer.FeatureDancer = featureDancer
	}
	if dateOfBirth, ok := rawAPIResponse["date_of_birth"].(string); ok && dateOfBirth != "" {
		newPerformer.DateOfBirth = dateOfBirth
	}
	if age, ok := rawAPIResponse["age"].(string); ok && age != "" {
		newPerformer.Age = age
	}
	if sexualOrientation, ok := rawAPIResponse["sexual_orientation"].(string); ok && sexualOrientation != "" {
		newPerformer.SexualOrientation = sexualOrientation
	}
	if astrologicalSign, ok := rawAPIResponse["astrological_sign"].(string); ok && astrologicalSign != "" {
		newPerformer.AstrologicalSign = astrologicalSign
	}
	if profession, ok := rawAPIResponse["profession"].(string); ok && profession != "" {
		newPerformer.Profession = profession
	}
	if careerStatus, ok := rawAPIResponse["career_status"].(string); ok && careerStatus != "" {
		newPerformer.CareerStatus = careerStatus
	}
	if careerStart, ok := rawAPIResponse["career_start"].(string); ok && careerStart != "" {
		newPerformer.CareerStart = careerStart
	}
	if careerEnd, ok := rawAPIResponse["career_end"].(string); ok && careerEnd != "" {
		newPerformer.CareerEnd = careerEnd
	}
	if dateOfDeath, ok := rawAPIResponse["date_of_death"].(string); ok && dateOfDeath != "" {
		newPerformer.DateOfDeath = dateOfDeath
	}
	if placeOfBirth, ok := rawAPIResponse["place_of_birth"].(string); ok && placeOfBirth != "" {
		newPerformer.PlaceOfBirth = placeOfBirth
	}
	if nationality, ok := rawAPIResponse["nationality"].(string); ok && nationality != "" {
		newPerformer.Nationality = nationality
	}
	if rank, ok := rawAPIResponse["rank"].(string); ok && rank != "" {
		newPerformer.Rank = rank
	}
	if country, ok := rawAPIResponse["country"].(string); ok && country != "" {
		newPerformer.Country = country
	}
	if avatar, ok := rawAPIResponse["avatar"].(string); ok && avatar != "" {
		newPerformer.Avatar = avatar
	}
	if subscribers, ok := rawAPIResponse["subscribers"].(float64); ok {
		newPerformer.Subscribers = int(subscribers)
	}
	if rating, ok := rawAPIResponse["rating"].(float64); ok {
		newPerformer.Rating = int(rating)
	}
	if totalViews, ok := rawAPIResponse["total_views"].(float64); ok {
		newPerformer.TotalViews = int(totalViews)
	}
	if totalVideoCount, ok := rawAPIResponse["total_video_count"].(float64); ok {
		newPerformer.TotalVideoCount = int(totalVideoCount)
	}
	if totalPlatformHits, ok := rawAPIResponse["total_platform_hits"].(float64); ok {
		newPerformer.TotalPlatformHits = int(totalPlatformHits)
	}

	// Update the performer in the database
	updatedPerformerJSON, err := json.Marshal(newPerformer)
	if err != nil {
		log.Printf("Failed to marshal updated performer %s to JSON: %v", performerName, err)
		http.Error(w, "Failed to process performer data", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedPerformerJSON), performerName)
	if err != nil {
		log.Printf("Failed to update performer %s in database: %v", performerName, err)
		http.Error(w, "Failed to update performer data in database", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully fetched and updated metadata for %s.", performerName)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Metadata fetched and updated successfully"})
}

func refetchAllPerformerMetadataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		log.Println("Starting task to refetch all performer metadata...")
		performers, err := getAllPerformerNames()
		if err != nil {
			log.Printf("Failed to get all performer names for metadata refetch: %v", err)
			return
		}

		var wg sync.WaitGroup
		for _, pName := range performers {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				log.Printf("Refetching metadata for %s", name)
				// Fetch performer data from Adultdatalink API
				performerDataBytes, err := fetchPerformerData(name)
				if err != nil {
					log.Printf("Failed to fetch data for performer %s from Adultdatalink API: %v", name, err)
					return
				}

				var rawAPIResponse map[string]interface{}
				if err := json.Unmarshal(performerDataBytes, &rawAPIResponse); err != nil {
					log.Printf("Failed to unmarshal raw Adultdatalink API response for %s: %v", name, err)
					return
				}

				var performer Performer
				var existingPerformerJSON string
				err = db.QueryRow("SELECT data FROM performers WHERE name = ?", name).Scan(&existingPerformerJSON)
				if err != nil {
					log.Printf("Failed to query existing data for performer %s: %v", name, err)
					return
				}
				if err := json.Unmarshal([]byte(existingPerformerJSON), &performer); err != nil {
					log.Printf("Failed to unmarshal existing data for performer %s: %v", name, err)
					return
				}

				// Update performer fields from rawAPIResponse
				// (This part is very similar to fetchPerformerMetadataHandler and autoAddPerformersFromFolders)
				// Process Aliases
	if aliases, ok := rawAPIResponse["aliases"]; ok {
		if aliasesList, ok := aliases.([]interface{}); ok {
			var tempAliases []string
			for _, alias := range aliasesList {
				if aliasMap, ok := alias.(map[string]interface{}); ok {
					for k := range aliasMap {
						tempAliases = append(tempAliases, k)
					}
				} else if aliasStr, ok := alias.(string); ok {
					tempAliases = append(tempAliases, aliasStr)
				}
			}
			performer.Aliases = tempAliases
		} else if aliasStr, ok := aliases.(string); ok {
			performer.Aliases = []string{aliasStr}
		}
	}

	// Process Tags
	if tags, ok := rawAPIResponse["tags"]; ok {
		if tagsList, ok := tags.([]interface{}); ok {
			var tempTags []string
			for _, tag := range tagsList {
				if tagMap, ok := tag.(map[string]interface{}); ok {
					for k := range tagMap {
						tempTags = append(tempTags, k)
					}
				} else if tagStr, ok := tag.(string); ok {
					tempTags = append(tempTags, tagStr)
				}
			}
			performer.Tags = tempTags
		} else if tagStr, ok := tags.(string); ok {
			performer.Tags = []string{tagStr}
		}
	}

	// Populate other fields from rawAPIResponse
	if imageURL, ok := rawAPIResponse["image_url"].(string); ok && imageURL != "" {
		performer.ImageURL = imageURL
	}
	if appearance, ok := rawAPIResponse["appearance"].(map[string]interface{}); ok && len(appearance) > 0 {
		performer.Appearance = appearance
	}
	if performances, ok := rawAPIResponse["performances"].(map[string]interface{}); ok && len(performances) > 0 {
		performer.Performances = performances
	}
	if socialMedia, ok := rawAPIResponse["social_media"].(map[string]interface{}); ok && len(socialMedia) > 0 {
		performer.SocialMedia = socialMedia
	}
	if platformViews, ok := rawAPIResponse["platform_views"].(map[string]interface{}); ok && len(platformViews) > 0 {
		performer.PlatformViews = platformViews
	}
	if platformVideoCounts, ok := rawAPIResponse["platform_video_counts"].(map[string]interface{}); ok && len(platformVideoCounts) > 0 {
		performer.PlatformVideoCounts = platformVideoCounts
	}
	if platformProfileCounts, ok := rawAPIResponse["platform_profile_counts"].(map[string]interface{}); ok && len(platformProfileCounts) > 0 {
		performer.PlatformProfileCounts = platformProfileCounts
	}
	if externalLinks, ok := rawAPIResponse["external_links"].([]interface{}); ok && len(externalLinks) > 0 {
		// Convert []interface{} to []map[string]string
		var convertedLinks []map[string]string
		for _, link := range externalLinks {
			if linkMap, isMap := link.(map[string]interface{}); isMap {
				convertedLink := make(map[string]string)
				for k, v := range linkMap {
					if strVal, isString := v.(string); isString {
						convertedLink[k] = strVal
					}
				}
				convertedLinks = append(convertedLinks, convertedLink)
			}
		}
		performer.ExternalLinks = convertedLinks
	}
	if bios, ok := rawAPIResponse["bios"]; ok {
		if biosMap, isMap := bios.(map[string]interface{}); isMap {
			convertedBios := make(map[string]string)
			for k, v := range biosMap {
				if strVal, isString := v.(string); isString {
					convertedBios[k] = strVal
				}
			}
			performer.Bios = convertedBios
		}
	}
	if officialWebsite, ok := rawAPIResponse["official_website"].(string); ok && officialWebsite != "" {
		performer.OfficialWebsite = officialWebsite
	}
	if featureDancer, ok := rawAPIResponse["feature_dancer"].(string); ok && featureDancer != "" {
		performer.FeatureDancer = featureDancer
	}
	if dateOfBirth, ok := rawAPIResponse["date_of_birth"].(string); ok && dateOfBirth != "" {
		performer.DateOfBirth = dateOfBirth
	}
	if age, ok := rawAPIResponse["age"].(string); ok && age != "" {
		performer.Age = age
	}
	if sexualOrientation, ok := rawAPIResponse["sexual_orientation"].(string); ok && sexualOrientation != "" {
		performer.SexualOrientation = sexualOrientation
	}
	if astrologicalSign, ok := rawAPIResponse["astrological_sign"].(string); ok && astrologicalSign != "" {
		performer.AstrologicalSign = astrologicalSign
	}
	if profession, ok := rawAPIResponse["profession"].(string); ok && profession != "" {
		performer.Profession = profession
	}
	if careerStatus, ok := rawAPIResponse["career_status"].(string); ok && careerStatus != "" {
		performer.CareerStatus = careerStatus
	}
	if careerStart, ok := rawAPIResponse["career_start"].(string); ok && careerStart != "" {
		performer.CareerStart = careerStart
	}
	if careerEnd, ok := rawAPIResponse["career_end"].(string); ok && careerEnd != "" {
		performer.CareerEnd = careerEnd
	}
	if dateOfDeath, ok := rawAPIResponse["date_of_death"].(string); ok && dateOfDeath != "" {
		performer.DateOfDeath = dateOfDeath
	}
	if placeOfBirth, ok := rawAPIResponse["place_of_birth"].(string); ok && placeOfBirth != "" {
		performer.PlaceOfBirth = placeOfBirth
	}
	if nationality, ok := rawAPIResponse["nationality"].(string); ok && nationality != "" {
		performer.Nationality = nationality
	}
	if rank, ok := rawAPIResponse["rank"].(string); ok && rank != "" {
		performer.Rank = rank
	}
	if country, ok := rawAPIResponse["country"].(string); ok && country != "" {
		performer.Country = country
	}
	if avatar, ok := rawAPIResponse["avatar"].(string); ok && avatar != "" {
		performer.Avatar = avatar
	}
	if subscribers, ok := rawAPIResponse["subscribers"].(float64); ok {
		performer.Subscribers = int(subscribers)
	}
	if rating, ok := rawAPIResponse["rating"].(float64); ok {
		performer.Rating = int(rating)
	}
	if totalViews, ok := rawAPIResponse["total_views"].(float64); ok {
		performer.TotalViews = int(totalViews)
	}
	if totalVideoCount, ok := rawAPIResponse["total_video_count"].(float64); ok {
		performer.TotalVideoCount = int(totalVideoCount)
	}
	if totalPlatformHits, ok := rawAPIResponse["total_platform_hits"].(float64); ok {
		performer.TotalPlatformHits = int(totalPlatformHits)
	}


				updatedPerformerJSON, err := json.Marshal(performer)
				if err != nil {
					log.Printf("Failed to marshal updated performer %s to JSON: %v", name, err)
					return
				}

				_, err = db.Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedPerformerJSON), name)
				if err != nil {
					log.Printf("Failed to update performer %s in database: %v", name, err)
					return
				}
				log.Printf("Successfully refetched and updated metadata for %s.", name)
			}(pName)
		}
		wg.Wait()
		log.Println("Finished refetching all performer metadata.")
	}()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task to refetch all performer metadata started."})
}

func updatePerformerPreviewsTask() {
	log.Println("Starting update performer previews task...")
	rows, err := db.Query("SELECT name, data FROM performers")
	if err != nil {
		log.Printf("Failed to query performers for preview update: %v", err)
		return
	}
	defer rows.Close()

	var wg sync.WaitGroup
	performersToUpdate := make(chan Performer, 100) // Buffer channel

	for rows.Next() {
		var name string
		var performerJSON string
		if err := rows.Scan(&name, &performerJSON); err != nil {
			log.Printf("Failed to scan performer data for preview update: %v", err)
			continue
		}

		var performer Performer
		if err := json.Unmarshal([]byte(performerJSON), &performer); err != nil {
			log.Printf("Failed to unmarshal performer JSON for %s: %v", name, err)
			continue
		}

		wg.Add(1)
		go func(p Performer) {
			defer wg.Done()

			performerFolderPath := filepath.Join(performerFoldersDir, p.Name)
			currentPreviews := []string{}

			performerEntries, err := os.ReadDir(performerFolderPath)
			if err != nil {
				log.Printf("Failed to read performer folder %s for previews: %v", performerFolderPath, err)
			} else {
				for _, entry := range performerEntries {
					ext := strings.ToLower(filepath.Ext(entry.Name()))
					if !entry.IsDir() && (ext == ".mkv" || ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif") {
						relativePath, err := filepath.Rel(performerFoldersDir, filepath.Join(performerFolderPath, entry.Name()))
						if err != nil {
							log.Printf("Failed to get relative path for preview %s: %v", entry.Name(), err)
							continue
						}
						currentPreviews = append(currentPreviews, filepath.ToSlash(relativePath))
					}
				}
			}

			// Compare current previews with stored previews
			if !compareStringSlices(p.Previews, currentPreviews) {
				log.Printf("Previews changed for %s. Updating...", p.Name)
				p.Previews = currentPreviews
				// If default preview is no longer valid, clear it
				if p.DefaultPreview != "" && !containsString(currentPreviews, p.DefaultPreview) {
					p.DefaultPreview = ""
				}
				performersToUpdate <- p
			}
		}(performer)
	}

	wg.Wait()
	close(performersToUpdate)

	// Batch update database
	for p := range performersToUpdate {
		updatedJSON, err := json.Marshal(p)
		if err != nil {
			log.Printf("Failed to marshal updated performer %s to JSON: %v", p.Name, err)
			continue
		}
		_, err = db.Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedJSON), p.Name)
		if err != nil {
			log.Printf("Failed to update performer %s in database: %v", p.Name, err)
		}
	}
		log.Println("Update performer previews task completed.")
	}
	
	// Helper function to compare two string slices
	func compareStringSlices(a, b []string) bool {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
		return true
	}
	
	// Helper function to check if a string is in a slice
	func containsString(slice []string, s string) bool {
		for _, v := range slice {
			if v == s {
				return true
			}
		}
		return false
	}
	
	func main() {
	
	setupLogging()
	initDB()
	updateExistingPerformersSchema() // Call the new function here
	autoAddPerformersFromFolders()
	initializeVideoCache()
	log.Println("Application starting...")

	// Create a file server to serve static files (HTML, CSS, JS) from a "frontend" directory.
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	// Handle API requests
	http.HandleFunc("/api/videos", videosHandler)
	http.HandleFunc("/api/rename", renameHandler)
	http.HandleFunc("/api/chat", chatHandler)
	http.HandleFunc("/api/performers", performersHandler)
	http.HandleFunc("/api/performers/", performerDetailsHandler)
	http.HandleFunc("/api/tasks/update-performer-previews", updatePerformerPreviewsHandler)
	http.HandleFunc("/api/tasks/refetch-all-performer-metadata", refetchAllPerformerMetadataHandler)
	http.HandleFunc("/api/logs/previous", previousLogsHandler)
	http.HandleFunc("/api/logs/previous/", previousLogFileHandler)
	http.HandleFunc("/api/logs/current", currentLogsHandler)
	http.HandleFunc("/performer-previews/", performerPreviewHandler)

	// Handle video streaming
	http.HandleFunc("/video/", videoStreamHandler)

	// Define the port the server will listen on.
	port := "8080"
	fmt.Printf("Starting server at http://localhost:%s\n", port)

	// Start the server and log any errors.
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

