package main

import (
	"fmt"
	"log"
	"net/http"

	"video-organizer/internal/database"
	"video-organizer/internal/handlers"
	"video-organizer/internal/logging"
	"video-organizer/internal/performer"
	"video-organizer/internal/video"
)



func main() {

	logging.SetupLogging()
	database.InitDB()
	performer.UpdateExistingPerformersSchema() // Call the new function here
	performer.AutoAddPerformersFromFolders()
	video.InitializeVideoCache()
	log.Println("Application starting...")

	// Create a file server to serve static files (HTML, CSS, JS) from a "frontend" directory.
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	// Serve .thumbnails directory directly
	http.Handle("/frontend/.thumbnails/", http.StripPrefix("/frontend/.thumbnails/", http.FileServer(http.Dir("./frontend/.thumbnails"))))

	// Handle API requests
	http.HandleFunc("/api/videos", handlers.VideosHandler)
	http.HandleFunc("/api/rename", handlers.RenameHandler)
	http.HandleFunc("/api/chat", handlers.ChatHandler)
	http.HandleFunc("/api/performers", handlers.PerformersHandler)
	http.HandleFunc("/api/performers/", handlers.PerformerDetailsHandler)
	http.HandleFunc("/api/libraries", handlers.LibrariesHandler)
	http.HandleFunc("/api/libraries/", handlers.LibraryDetailsHandler)
	http.HandleFunc("/api/tasks/update-performer-previews", handlers.UpdatePerformerPreviewsHandler)
	http.HandleFunc("/api/tasks/refetch-all-performer-metadata", handlers.RefetchAllPerformerMetadataHandler)
	http.HandleFunc("/api/logs/previous", handlers.PreviousLogsHandler)
	http.HandleFunc("/api/logs/previous/", handlers.PreviousLogFileHandler)
	http.HandleFunc("/api/logs/current", handlers.CurrentLogsHandler)
	http.HandleFunc("/performer-previews/", handlers.PerformerPreviewHandler)

	// Handle video streaming
	http.HandleFunc("/video/", handlers.VideoStreamHandler)

	// Define the port the server will listen on.
	port := "8080"
	fmt.Printf("Starting server at http://localhost:%s\n", port)

	// Start the server and log any errors.
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
