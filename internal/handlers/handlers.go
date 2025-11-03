package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"video-organizer/internal/appstatus"

	"video-organizer/internal/config"
	"video-organizer/internal/database"
	"video-organizer/internal/models"
	"video-organizer/internal/performer"
	"video-organizer/internal/video"
)

func VideosHandler(w http.ResponseWriter, r *http.Request) {
	appstatus.EmitInfo("videos", "List videos requested")
	var videos []models.VideoInfo
	for _, v := range video.GetVideoCache() {
		videos = append(videos, v)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videos)
}

func RenameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("Invalid method for RenameHandler:", r.Method)
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RenameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid request body for RenameHandler:", err)
		appstatus.EmitError("videos", "Invalid rename request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	appstatus.EmitInfo("videos", fmt.Sprintf("Rename requested: %s -> %s", req.OldName, req.NewName))

	// Rename video file
	oldVideoPath := filepath.Join(config.VideoDir, req.OldName)
	newVideoPath := filepath.Join(config.VideoDir, req.NewName)
	if err := os.Rename(oldVideoPath, newVideoPath); err != nil {
		log.Println("Failed to rename video", req.OldName, ":", err)
		appstatus.EmitError("videos", fmt.Sprintf("Failed to rename %s: %v", req.OldName, err))
		http.Error(w, fmt.Sprintf("Failed to rename video: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println("Successfully renamed video from", req.OldName, "to", req.NewName)
	appstatus.EmitInfo("videos", fmt.Sprintf("Renamed %s -> %s", req.OldName, req.NewName))

	// Rename thumbnail file
	oldThumbName := fmt.Sprintf("%s.jpg", strings.TrimSuffix(req.OldName, filepath.Ext(req.OldName)))
	newThumbName := fmt.Sprintf("%s.jpg", strings.TrimSuffix(req.NewName, filepath.Ext(req.NewName)))
	oldThumbPath := filepath.Join(config.ThumbnailDir, oldThumbName)
	newThumbPath := filepath.Join(config.ThumbnailDir, newThumbName)
	if _, err := os.Stat(oldThumbPath); err == nil {
		if err := os.Rename(oldThumbPath, newThumbPath); err != nil {
			log.Println("Failed to rename thumbnail for", req.OldName, ":", err)
		}
	}

	// Update cache
	delete(video.GetVideoCache(), req.OldName)

	performerNames, err := database.GetAllPerformerNames()
	if err != nil {
		log.Printf("Failed to get performer names for rename handler: %v", err)
		http.Error(w, "Failed to update performer associations", http.StatusInternalServerError)
		return
	}

	video.GetVideoCache()[req.NewName] = video.GenerateVideoInfo(req.NewName, newVideoPath, performerNames)

	// Re-evaluate and update performer scene counts
	video.UpdatePerformerSceneCounts()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("Invalid method for ChatHandler:", r.Method)
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg models.ChatMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Println("Invalid request body for ChatHandler:", err)
		appstatus.EmitError("chat", "Invalid chat request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	appstatus.EmitInfo("chat", fmt.Sprintf("Chat message received: %.80s", msg.Message))

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

func PerformersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var performer models.Performer
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

		_, err = database.GetDB().Exec("INSERT INTO performers(name, data) VALUES(?, ?)", performer.Name, string(performerJSON))
		if err != nil {
			log.Println("Failed to insert performer into database:", err)
			appstatus.EmitError("performers", fmt.Sprintf("Failed to add performer: %v", err))
			http.Error(w, "Failed to add performer", http.StatusInternalServerError)
			return
		}

		appstatus.EmitInfo("performers", fmt.Sprintf("Performer added: %s", performer.Name))
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Performer added successfully"})

	case http.MethodGet:

		rows, err := database.GetDB().Query("SELECT data FROM performers")
		if err != nil {
			log.Println("Failed to query performers from database:", err)
			http.Error(w, "Failed to retrieve performers", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var performers []models.Performer
		for rows.Next() {
			var performerJSON string
			if err := rows.Scan(&performerJSON); err != nil {
				log.Println("Failed to scan performer data:", err)
				continue
			}
			var performer models.Performer
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

// LibrariesHandler handles listing and creating libraries
func LibrariesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := database.GetDB().Query("SELECT id, name, path, is_default FROM libraries ORDER BY id ASC")
		if err != nil {
			log.Printf("Failed to query libraries: %v", err)
			http.Error(w, "Failed to retrieve libraries", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var libs []models.Library
		for rows.Next() {
			var id int
			var name, path string
			var isDefaultInt int
			if err := rows.Scan(&id, &name, &path, &isDefaultInt); err != nil {
				log.Printf("Failed to scan library row: %v", err)
				continue
			}
			libs = append(libs, models.Library{ID: id, Name: name, Path: path, IsDefault: isDefaultInt == 1})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(libs)
		appstatus.EmitInfo("libraries", fmt.Sprintf("Listed libraries (%d)", len(libs)))
		return

	case http.MethodPost:
		var lib models.Library
		if err := json.NewDecoder(r.Body).Decode(&lib); err != nil {
			log.Printf("Invalid request body for LibrariesHandler: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		tx, err := database.GetDB().Begin()
		if err != nil {
			log.Printf("Failed to begin transaction: %v", err)
			http.Error(w, "Failed to create library", http.StatusInternalServerError)
			return
		}
		if lib.IsDefault {
			if _, err := tx.Exec("UPDATE libraries SET is_default = 0"); err != nil {
				tx.Rollback()
				log.Printf("Failed to unset existing defaults: %v", err)
				http.Error(w, "Failed to create library", http.StatusInternalServerError)
				return
			}
		}
		res, err := tx.Exec("INSERT INTO libraries(name, path, is_default) VALUES(?, ?, ?)", lib.Name, lib.Path, boolToInt(lib.IsDefault))
		if err != nil {
			tx.Rollback()
			log.Printf("Failed to insert library: %v", err)
			http.Error(w, "Failed to create library", http.StatusInternalServerError)
			return
		}
		id, _ := res.LastInsertId()
		if err := tx.Commit(); err != nil {
			log.Printf("Failed to commit library insert: %v", err)
			http.Error(w, "Failed to create library", http.StatusInternalServerError)
			return
		}
		lib.ID = int(id)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(lib)
		appstatus.EmitInfo("libraries", fmt.Sprintf("Created library: %s (%d)", lib.Name, lib.ID))
		return

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// LibraryDetailsHandler handles update/delete and set-default actions for a library
func LibraryDetailsHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/libraries/")
	if path == "" {
		http.Error(w, "Library ID not specified", http.StatusBadRequest)
		return
	}

	// set-default action: /api/libraries/{id}/set-default
	if strings.HasSuffix(path, "/set-default") {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := strings.TrimSuffix(path, "/set-default")
		idStr = strings.Trim(idStr, "/")
		id := idStr
		// Start transaction
		tx, err := database.GetDB().Begin()
		if err != nil {
			log.Printf("Failed to begin tx for set-default: %v", err)
			http.Error(w, "Failed to set default", http.StatusInternalServerError)
			return
		}
		if _, err := tx.Exec("UPDATE libraries SET is_default = 0"); err != nil {
			tx.Rollback()
			log.Printf("Failed to unset defaults: %v", err)
			http.Error(w, "Failed to set default", http.StatusInternalServerError)
			return
		}
		if _, err := tx.Exec("UPDATE libraries SET is_default = 1 WHERE id = ?", id); err != nil {
			tx.Rollback()
			log.Printf("Failed to set default for id %s: %v", id, err)
			http.Error(w, "Failed to set default", http.StatusInternalServerError)
			return
		}
		if err := tx.Commit(); err != nil {
			log.Printf("Failed to commit set-default tx: %v", err)
			http.Error(w, "Failed to set default", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Default library updated"})
		appstatus.EmitInfo("libraries", fmt.Sprintf("Set default library id=%s", idStr))
		return
	}

	// Trim any trailing slash
	idStr := strings.Trim(path, "/")

	switch r.Method {
	case http.MethodPut:
		var lib models.Library
		if err := json.NewDecoder(r.Body).Decode(&lib); err != nil {
			log.Printf("Invalid request body for update library: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if lib.IsDefault {
			if _, err := database.GetDB().Exec("UPDATE libraries SET is_default = 0"); err != nil {
				log.Printf("Failed to unset defaults during update: %v", err)
				http.Error(w, "Failed to update library", http.StatusInternalServerError)
				return
			}
		}
		if _, err := database.GetDB().Exec("UPDATE libraries SET name = ?, path = ?, is_default = ? WHERE id = ?", lib.Name, lib.Path, boolToInt(lib.IsDefault), idStr); err != nil {
			log.Printf("Failed to update library %s: %v", idStr, err)
			http.Error(w, "Failed to update library", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Library updated"})
		appstatus.EmitInfo("libraries", fmt.Sprintf("Updated library id=%s", idStr))
		return

	case http.MethodDelete:
		if _, err := database.GetDB().Exec("DELETE FROM libraries WHERE id = ?", idStr); err != nil {
			log.Printf("Failed to delete library %s: %v", idStr, err)
			http.Error(w, "Failed to delete library", http.StatusInternalServerError)
			return
		}
		// Ensure there is a default library if none exists
		var count int
		_ = database.GetDB().QueryRow("SELECT COUNT(*) FROM libraries WHERE is_default = 1").Scan(&count)
		if count == 0 {
			// set first library as default
			_, _ = database.GetDB().Exec("UPDATE libraries SET is_default = 1 WHERE id = (SELECT id FROM libraries ORDER BY id ASC LIMIT 1)")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Library deleted"})
		appstatus.EmitInfo("libraries", fmt.Sprintf("Deleted library id=%s", idStr))
		return

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func PerformerDetailsHandler(w http.ResponseWriter, r *http.Request) {
	var err error
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
		err = database.GetDB().QueryRow("SELECT data FROM performers WHERE name = ?", performerName).Scan(&performerJSON)
		if err == sql.ErrNoRows {
			http.Error(w, "Performer not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Failed to query performer %s from database: %v", performerName, err)
			http.Error(w, "Failed to retrieve performer details", http.StatusInternalServerError)
			return
		}

		var p models.Performer
		if err = json.Unmarshal([]byte(performerJSON), &p); err != nil {
			log.Printf("Failed to unmarshal performer JSON for %s: %v", performerName, err)
			http.Error(w, "Failed to process performer data", http.StatusInternalServerError)
			return
		}

		p.DefaultPreview = payload.PreviewURL
		updatedPerformerJSON, err := json.Marshal(p)
		if err != nil {
			log.Printf("Failed to marshal updated performer %s to JSON: %v", performerName, err)
			http.Error(w, "Failed to update performer data", http.StatusInternalServerError)
			return
		}

		_, err = database.GetDB().Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedPerformerJSON), performerName)
		if err != nil {
			log.Printf("Failed to update performer %s in database: %v", performerName, err)
			http.Error(w, "Failed to save default preview", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Default preview updated successfully"})
		appstatus.EmitInfo("performers", fmt.Sprintf("Set default preview for %s", performerName))
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
		err = database.GetDB().QueryRow("SELECT data FROM performers WHERE name = ?", performerName).Scan(&performerJSON)
		if err == sql.ErrNoRows {
			http.Error(w, "Performer not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Failed to query performer %s from database: %v", performerName, err)
			http.Error(w, "Failed to retrieve performer details", http.StatusInternalServerError)
			return
		}

		var performer models.Performer
		if err = json.Unmarshal([]byte(performerJSON), &performer); err != nil {
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

		_, err = database.GetDB().Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedPerformerJSON), performerName)
		if err != nil {
			log.Printf("ERROR: Failed to update performer %s in database for set-zoo: %v", performerName, err)
			http.Error(w, "Failed to save zoo status", http.StatusInternalServerError)
			return
		}
		log.Printf("DEBUG: Successfully updated zoo status for %s in database.", performerName)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Zoo status updated successfully"})
		appstatus.EmitInfo("performers", fmt.Sprintf("Zoo status updated for %s: %s", performerName, payload.Zoo))
		return
	}

	// Handle reset-metadata action
	if strings.HasSuffix(performerName, "/reset-metadata") {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		performerName = strings.TrimSuffix(performerName, "/reset-metadata")
		if performerName == "" {
			http.Error(w, "Performer name not specified", http.StatusBadRequest)
			return
		}

		var performerJSON string
		err = database.GetDB().QueryRow("SELECT data FROM performers WHERE name = ?", performerName).Scan(&performerJSON)
		if err == sql.ErrNoRows {
			http.Error(w, "Performer not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Failed to query performer %s from database: %v", performerName, err)
			http.Error(w, "Failed to retrieve performer details", http.StatusInternalServerError)
			return
		}

		var existingPerformer models.Performer
		if err := json.Unmarshal([]byte(performerJSON), &existingPerformer); err != nil {
			log.Printf("Failed to unmarshal performer JSON for %s: %v", performerName, err)
			http.Error(w, "Failed to process performer data", http.StatusInternalServerError)
			return
		}

		newPerformer := database.NewDefaultPerformer()
		newPerformer.Name = existingPerformer.Name
		newPerformer.Zoo = existingPerformer.Zoo
		newPerformer.Previews = existingPerformer.Previews
		newPerformer.DefaultPreview = existingPerformer.DefaultPreview

		if newPerformer.Zoo == "true" {
			newPerformer.Profession = "Bestiality"
		}

		updatedPerformerJSON, err := json.Marshal(newPerformer)
		if err != nil {
			log.Printf("Failed to marshal updated performer %s to JSON: %v", performerName, err)
			http.Error(w, "Failed to update performer data", http.StatusInternalServerError)
			return
		}

		_, err = database.GetDB().Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedPerformerJSON), performerName)
		if err != nil {
			log.Printf("Failed to update performer %s in database: %v", performerName, err)
			http.Error(w, "Failed to reset metadata", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Metadata reset successfully"})
		appstatus.EmitInfo("performers", fmt.Sprintf("Reset metadata for %s", performerName))
		return
	}

	// Handle reset-previews action
	if strings.HasSuffix(performerName, "/reset-previews") {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		performerName = strings.TrimSuffix(performerName, "/reset-previews")
		if performerName == "" {
			http.Error(w, "Performer name not specified", http.StatusBadRequest)
			return
		}

		// Scan for previews
		performerFolderPath := filepath.Join(config.PerformerFoldersDir, performerName)
		performerEntries, err := os.ReadDir(performerFolderPath)
		if err != nil {
			log.Printf("Failed to read performer folder %s for previews: %v", performerFolderPath, err)
			http.Error(w, "Failed to read performer folder", http.StatusInternalServerError)
			return
		}

		var previews []string
		for _, pEntry := range performerEntries {
			ext := strings.ToLower(filepath.Ext(pEntry.Name()))
			if !pEntry.IsDir() && ext == ".mkv" {
				relativePath, err := filepath.Rel(config.PerformerFoldersDir, filepath.Join(performerFolderPath, pEntry.Name()))
				if err != nil {
					log.Printf("Failed to get relative path for preview %s: %v", pEntry.Name(), err)
					continue
				}
				previews = append(previews, filepath.ToSlash(relativePath))
			}
		}

		var performerJSON string
		err = database.GetDB().QueryRow("SELECT data FROM performers WHERE name = ?", performerName).Scan(&performerJSON)
		if err == sql.ErrNoRows {
			http.Error(w, "Performer not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Failed to query performer %s from database: %v", performerName, err)
			http.Error(w, "Failed to retrieve performer details", http.StatusInternalServerError)
			return
		}

		var p models.Performer
		if err = json.Unmarshal([]byte(performerJSON), &p); err != nil {
			log.Printf("Failed to unmarshal performer JSON for %s: %v", performerName, err)
			http.Error(w, "Failed to process performer data", http.StatusInternalServerError)
			return
		}

		p.Previews = previews
		if p.DefaultPreview != "" && !performer.ContainsString(previews, p.DefaultPreview) {
			p.DefaultPreview = ""
		}

		updatedPerformerJSON, err := json.Marshal(p)
		if err != nil {
			log.Printf("Failed to marshal updated performer %s to JSON: %v", performerName, err)
			http.Error(w, "Failed to update performer data", http.StatusInternalServerError)
			return
		}

		_, err = database.GetDB().Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedPerformerJSON), performerName)
		if err != nil {
			log.Printf("Failed to update performer %s in database: %v", performerName, err)
			http.Error(w, "Failed to reset previews", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Previews reset successfully"})
		appstatus.EmitInfo("performers", fmt.Sprintf("Reset previews for %s (%d previews)", performerName, len(previews)))
		return
	}

	// Handle fetch-metadata action
	if strings.HasSuffix(performerName, "/fetch-metadata") {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		trimmedPerformerName := strings.TrimSuffix(performerName, "/fetch-metadata")
		if trimmedPerformerName == "" {
			http.Error(w, "Performer name not specified", http.StatusBadRequest)
			return
		}

		log.Printf("Fetching metadata for performer: %s", trimmedPerformerName)

		// Fetch performer data from Adultdatalink API
		performerDataBytes, err := performer.FetchPerformerData(trimmedPerformerName)
		if err != nil {
			log.Printf("Failed to fetch data for performer %s from Adultdatalink API: %v", trimmedPerformerName, err)
			http.Error(w, fmt.Sprintf("Failed to fetch metadata: %v", err), http.StatusInternalServerError)
			return
		}

		var rawAPIResponse map[string]interface{}
		if err := json.Unmarshal(performerDataBytes, &rawAPIResponse); err != nil {
			log.Printf("Failed to unmarshal raw Adultdatalink API response for %s: %v", trimmedPerformerName, err)
			http.Error(w, "Failed to process API response", http.StatusInternalServerError)
			return
		}

		var p models.Performer
		var existingPerformerJSON string
		err = database.GetDB().QueryRow("SELECT data FROM performers WHERE name = ?", trimmedPerformerName).Scan(&existingPerformerJSON)
		if err == sql.ErrNoRows {
			// Performer not found in DB, create a new one
			newPerformer := database.NewDefaultPerformer()
			newPerformer.Name = trimmedPerformerName
			performer.PopulatePerformerFromAPIResponse(&newPerformer, rawAPIResponse)

			performerJSON, err := json.Marshal(newPerformer)
			if err != nil {
				log.Printf("Failed to marshal new performer %s to JSON: %v", trimmedPerformerName, err)
				http.Error(w, "Failed to process performer data", http.StatusInternalServerError)
				return
			}
			_, err = database.GetDB().Exec("INSERT INTO performers(name, data) VALUES(?, ?)", trimmedPerformerName, string(performerJSON))
			if err != nil {
				log.Printf("Failed to insert new performer %s into database: %v", trimmedPerformerName, err)
				http.Error(w, "Failed to add performer", http.StatusInternalServerError)
				return
			}
			log.Printf("Successfully fetched metadata and added new performer %s.", trimmedPerformerName)

		} else if err != nil {
			log.Printf("Failed to query performer %s from database: %v", trimmedPerformerName, err)
			http.Error(w, "Failed to retrieve performer details", http.StatusInternalServerError)
			return
		} else {
			// Performer found, unmarshal existing data and update
			if err = json.Unmarshal([]byte(existingPerformerJSON), &p); err != nil {
				log.Printf("Failed to unmarshal existing performer JSON for %s: %v", trimmedPerformerName, err)
				http.Error(w, "Failed to process performer data", http.StatusInternalServerError)
				return
			}
			performer.PopulatePerformerFromAPIResponse(&p, rawAPIResponse)

			updatedPerformerJSON, err := json.Marshal(p)
			if err != nil {
				log.Printf("Failed to marshal updated performer %s to JSON: %v", trimmedPerformerName, err)
				http.Error(w, "Failed to process performer data", http.StatusInternalServerError)
				return
			}

			_, err = database.GetDB().Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedPerformerJSON), trimmedPerformerName)
			if err != nil {
				log.Printf("Failed to update performer %s in database: %v", trimmedPerformerName, err)
				http.Error(w, "Failed to update performer data in database", http.StatusInternalServerError)
				return
			}
			log.Printf("Successfully fetched and updated metadata for %s.", trimmedPerformerName)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Metadata fetched and updated successfully"})
		appstatus.EmitInfo("metadata", fmt.Sprintf("Fetched metadata for %s", trimmedPerformerName))
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
	err = database.GetDB().QueryRow("SELECT data FROM performers WHERE name = ?", performerName).Scan(&performerJSON)
	if err == sql.ErrNoRows {
		http.Error(w, "Performer not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println("Failed to query performer from database:", err)
		http.Error(w, "Failed to retrieve performer details", http.StatusInternalServerError)
		return
	}

	var p models.Performer
	if err := json.Unmarshal([]byte(performerJSON), &p); err != nil {
		log.Println("Failed to unmarshal performer JSON:", err)
		http.Error(w, "Failed to process performer data", http.StatusInternalServerError)
		return
	}

	// TODO: Associate scenes with the performer

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}
func UpdatePerformerPreviewsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		appstatus.EmitInfo("tasks", "Update performer previews task started")
		// run task and emit basic lifecycle events
		performer.UpdatePerformerPreviewsTask()
		appstatus.EmitInfo("tasks", "Update performer previews task finished")
	}()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Update performer previews task started."})
}

func VideoStreamHandler(w http.ResponseWriter, r *http.Request) {
	videoName := strings.TrimPrefix(r.URL.Path, "/video/")
	if videoName == "" {
		log.Println("Video name not specified in stream request")
		http.Error(w, "Video name not specified", http.StatusBadRequest)
		return
	}

	videoPath := filepath.Join(config.VideoDir, videoName)
	appstatus.EmitInfo("videos", fmt.Sprintf("Streaming video: %s", videoName))

	// Serve the video file
	http.ServeFile(w, r, videoPath)
}

func PreviousLogsHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(config.OldLogsDir)
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
	appstatus.EmitInfo("logs", fmt.Sprintf("Listed previous logs (%d)", len(logFiles)))
}

func PreviousLogFileHandler(w http.ResponseWriter, r *http.Request) {
	fileName := strings.TrimPrefix(r.URL.Path, "/api/logs/previous/")
	if fileName == "" {
		log.Println("Log file name not specified in request")
		http.Error(w, "Log file name not specified", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(config.OldLogsDir, fileName)

	// Check if the file exists and is within the old_logs directory
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Log file not found: %s", filePath)
		http.Error(w, "Log file not found", http.StatusNotFound)
		return
	}

	// Serve the log file
	appstatus.EmitInfo("logs", fmt.Sprintf("Serving previous log file: %s", fileName))
	http.ServeFile(w, r, filePath)
}

func CurrentLogsHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the current app.log file
	appstatus.EmitInfo("logs", "Serving current log file")
	http.ServeFile(w, r, config.LogFile)
}

func PerformerPreviewHandler(w http.ResponseWriter, r *http.Request) {
	previewPath := strings.TrimPrefix(r.URL.Path, "/performer-previews/") // Updated prefix
	if previewPath == "" {
		http.Error(w, "Preview path not specified", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(config.PerformerFoldersDir, previewPath)

	// Normalize paths for consistent comparison
	normalizedFullPath := filepath.ToSlash(fullPath)
	normalizedPerformerFoldersDir := filepath.ToSlash(config.PerformerFoldersDir)

	// Security check: ensure the path is within the performerFoldersDir
	if !strings.HasPrefix(normalizedFullPath, normalizedPerformerFoldersDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	appstatus.EmitInfo("thumbnails", fmt.Sprintf("Serving performer preview: %s", previewPath))
	http.ServeFile(w, r, fullPath)
}

func RefetchAllPerformerMetadataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		appstatus.EmitInfo("metadata", "Refetch all performer metadata task started")
		performers, err := database.GetAllPerformerNames()
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
				performerDataBytes, err := performer.FetchPerformerData(name)
				if err != nil {
					log.Printf("Failed to fetch data for performer %s from Adultdatalink API: %v", name, err)
					return
				}

				var rawAPIResponse map[string]interface{}
				if err := json.Unmarshal(performerDataBytes, &rawAPIResponse); err != nil {
					log.Printf("Failed to unmarshal raw Adultdatalink API response for %s: %v", name, err)
					return
				}

				var performerModel models.Performer
				var existingPerformerJSON string
				err = database.GetDB().QueryRow("SELECT data FROM performers WHERE name = ?", name).Scan(&existingPerformerJSON)
				if err != nil {
					log.Printf("Failed to query existing data for performer %s: %v", name, err)
					return
				}
				if err := json.Unmarshal([]byte(existingPerformerJSON), &performerModel); err != nil {
					log.Printf("Failed to unmarshal existing data for performer %s: %v", name, err)
					return
				}

				performer.PopulatePerformerFromAPIResponse(&performerModel, rawAPIResponse)

				updatedPerformerJSON, err := json.Marshal(performerModel)
				if err != nil {
					log.Printf("Failed to marshal updated performer %s to JSON: %v", name, err)
					return
				}

				_, err = database.GetDB().Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedPerformerJSON), name)
				if err != nil {
					log.Printf("Failed to update performer %s in database: %v", name, err)
					return
				}
				log.Printf("Successfully refetched and updated metadata for %s.", name)
			}(pName)
		}
		wg.Wait()
		appstatus.EmitInfo("metadata", "Refetch all performer metadata task finished")
	}()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task to refetch all performer metadata started."})
}

// MonitorEmitHandler allows posting a simple test event (useful for debugging)
func MonitorEmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Message  string `json:"message"`
		Category string `json:"category"`
		Level    string `json:"level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if strings.ToLower(payload.Level) == "error" {
		appstatus.EmitError(payload.Category, payload.Message)
	} else {
		appstatus.EmitInfo(payload.Category, payload.Message)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// MonitorEventsHandler returns persisted monitor events from the database.
// Supports optional query params: limit (int), since (unix ms), category, level
func MonitorEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	limitStr := q.Get("limit")
	limit := 100
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}
	sinceStr := q.Get("since")
	var whereClauses []string
	var args []interface{}
	if sinceStr != "" {
		if v, err := strconv.ParseInt(sinceStr, 10, 64); err == nil {
			whereClauses = append(whereClauses, "timestamp >= ?")
			args = append(args, v)
		}
	}
	category := q.Get("category")
	if category != "" {
		whereClauses = append(whereClauses, "category = ?")
		args = append(args, category)
	}
	level := q.Get("level")
	if level != "" {
		whereClauses = append(whereClauses, "level = ?")
		args = append(args, level)
	}
	where := ""
	if len(whereClauses) > 0 {
		where = "WHERE " + strings.Join(whereClauses, " AND ")
	}
	query := fmt.Sprintf("SELECT id, type, category, message, level, timestamp FROM monitor_events %s ORDER BY timestamp DESC LIMIT ?", where)
	args = append(args, limit)
	rows, err := database.GetDB().Query(query, args...)
	if err != nil {
		log.Printf("Failed to query monitor events: %v", err)
		http.Error(w, "Failed to retrieve events", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	type evt struct {
		ID        int64  `json:"id"`
		Type      string `json:"type"`
		Category  string `json:"category"`
		Message   string `json:"message"`
		Level     string `json:"level"`
		Timestamp int64  `json:"timestamp"`
	}
	var out []evt
	for rows.Next() {
		var e evt
		if err := rows.Scan(&e.ID, &e.Type, &e.Category, &e.Message, &e.Level, &e.Timestamp); err != nil {
			log.Printf("Failed to scan monitor event row: %v", err)
			continue
		}
		out = append(out, e)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// MonitorSettingsHandler allows reading and writing persisted monitor settings.
// GET returns all settings as a JSON object. POST/PUT accepts a JSON object of key->value pairs to store.
func MonitorSettingsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := database.GetDB().Query("SELECT key, value FROM monitor_settings")
		if err != nil {
			log.Printf("Failed to query monitor settings: %v", err)
			http.Error(w, "Failed to retrieve settings", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		settings := map[string]string{}
		for rows.Next() {
			var k, v string
			if err := rows.Scan(&k, &v); err != nil {
				log.Printf("Failed to scan monitor_settings row: %v", err)
				continue
			}
			settings[k] = v
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(settings)
		return
	case http.MethodPost, http.MethodPut:
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		tx, err := database.GetDB().Begin()
		if err != nil {
			log.Printf("Failed to begin tx for monitor settings: %v", err)
			http.Error(w, "Failed to save settings", http.StatusInternalServerError)
			return
		}
		for k, v := range payload {
			valStr := fmt.Sprintf("%v", v)
			if _, err := tx.Exec("INSERT OR REPLACE INTO monitor_settings(key, value) VALUES(?, ?)", k, valStr); err != nil {
				log.Printf("Failed to upsert monitor setting %s: %v", k, err)
				tx.Rollback()
				http.Error(w, "Failed to save settings", http.StatusInternalServerError)
				return
			}
		}
		if err := tx.Commit(); err != nil {
			log.Printf("Failed to commit monitor settings tx: %v", err)
			http.Error(w, "Failed to save settings", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
