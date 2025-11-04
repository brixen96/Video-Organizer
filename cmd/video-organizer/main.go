package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"video-organizer/internal/appstatus"
	"video-organizer/internal/config"
	"video-organizer/internal/database"
	"video-organizer/internal/handlers"
	"video-organizer/internal/logging"
	"video-organizer/internal/performer"
	"video-organizer/internal/video"
)

func main() {
	// Load .env file first
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables or defaults")
	}

	// Initialize logging
	logging.SetupLogging()
	log.Println("=== Video Organizer Starting ===")

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded successfully")
	log.Printf("Server will listen on %s:%s", cfg.Server.Host, cfg.Server.Port)

	// Initialize legacy config variables for backward compatibility
	// TODO: Remove this once all code uses the new Config struct
	config.InitLegacyVars()

	// Initialize database
	database.InitDB()
	log.Println("Database initialized")

	// Run migrations and setup
	performer.UpdateExistingPerformersSchema()
	
	// Auto-scan performers if enabled
	if cfg.Features.EnableAutoScan {
		log.Println("Auto-scan enabled, scanning for new performers...")
		performer.AutoAddPerformersFromFolders()
	}

	// Initialize video cache
	video.InitializeVideoCache()
	log.Println("Video cache initialized")

	// Setup HTTP routes
	setupRoutes()

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server at http://%s", addr)
	log.Println("Press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes() {
	// Serve static files
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	// Serve thumbnails directory
	http.Handle("/frontend/.thumbnails/", 
		http.StripPrefix("/frontend/.thumbnails/", 
			http.FileServer(http.Dir("./frontend/.thumbnails"))))

	// API routes - Videos
	http.HandleFunc("/api/videos", handlers.VideosHandler)
	http.HandleFunc("/api/rename", handlers.RenameHandler)
	http.HandleFunc("/video/", handlers.VideoStreamHandler)

	// API routes - Performers
	http.HandleFunc("/api/performers", handlers.PerformersHandler)
	http.HandleFunc("/api/performers/", handlers.PerformerDetailsHandler)
	http.HandleFunc("/performer-previews/", handlers.PerformerPreviewHandler)

	// API routes - Libraries
	http.HandleFunc("/api/libraries", handlers.LibrariesHandler)
	http.HandleFunc("/api/libraries/", handlers.LibraryDetailsHandler)

	// API routes - Monitoring
	http.HandleFunc("/api/monitor/subscribe", appstatus.SSEHandler)
	http.HandleFunc("/api/monitor/event", handlers.MonitorEmitHandler)
	http.HandleFunc("/api/monitor/events", handlers.MonitorEventsHandler)
	http.HandleFunc("/api/monitor/settings", handlers.MonitorSettingsHandler)

	// API routes - Tasks
	http.HandleFunc("/api/tasks/update-performer-previews", handlers.UpdatePerformerPreviewsHandler)
	http.HandleFunc("/api/tasks/refetch-all-performer-metadata", handlers.RefetchAllPerformerMetadataHandler)

	// API routes - Logs
	http.HandleFunc("/api/logs/previous", handlers.PreviousLogsHandler)
	http.HandleFunc("/api/logs/previous/", handlers.PreviousLogFileHandler)
	http.HandleFunc("/api/logs/current", handlers.CurrentLogsHandler)

	// API routes - Chat (placeholder)
	http.HandleFunc("/api/chat", handlers.ChatHandler)

	log.Println("HTTP routes configured")
}

// gracefulShutdown handles cleanup on application exit
func gracefulShutdown() {
	log.Println("Shutting down gracefully...")
	
	// Close database connections
	if db := database.GetDB(); db != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
	
	log.Println("Shutdown complete")
	os.Exit(0)
}
