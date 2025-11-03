package performer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"video-organizer/internal/config"
	"video-organizer/internal/database"
	"video-organizer/internal/models"
)

func FetchPerformerData(performerName string) ([]byte, error) {
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

func PopulatePerformerFromAPIResponse(performer *models.Performer, rawAPIResponse map[string]interface{}) {
	// Process Aliases
	if aliases, ok := rawAPIResponse["aliases"]; ok {
		if aliasesList, ok := aliases.([]interface{}); ok {
			for _, alias := range aliasesList {
				if aliasMap, ok := alias.(map[string]interface{}); ok {
					for k := range aliasMap {
						performer.Aliases = append(performer.Aliases, k)
					}
				} else if aliasStr, ok := alias.(string); ok {
					performer.Aliases = append(performer.Aliases, aliasStr)
				}
			}
		} else if aliasStr, ok := aliases.(string); ok {
			performer.Aliases = append(performer.Aliases, aliasStr)
		}
	}

	// Process Tags
	if tags, ok := rawAPIResponse["tags"]; ok {
		if tagsList, ok := tags.([]interface{}); ok {
			for _, tag := range tagsList {
				if tagMap, ok := tag.(map[string]interface{}); ok {
					for k := range tagMap {
						performer.Tags = append(performer.Tags, k)
					}
				} else if tagStr, ok := tag.(string); ok {
					performer.Tags = append(performer.Tags, tagStr)
				}
			}
		} else if tagStr, ok := tags.(string); ok {
			performer.Tags = append(performer.Tags, tagStr)
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
}

func UpdateExistingPerformersSchema() {
	log.Println("Updating existing performer schema in database...")

	rows, err := database.GetDB().Query("SELECT name, data FROM performers")
	if err != nil {
		log.Printf("Failed to query existing performers for schema update: %v", err)
		return
	}
	defer rows.Close()

	var performersToUpdate []models.Performer

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

		newPerformer := database.NewDefaultPerformer() // Start with the default template

		newPerformer.Name = name // Keep the existing name

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
		_, err = database.GetDB().Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedJSON), p.Name)
		if err != nil {
			log.Printf("Failed to update performer %s in database: %v", p.Name, err)
		}
	}
	log.Println("Performer schema update completed.")
}

func AutoAddPerformersFromFolders() {
	log.Println("Checking for new performers in folders...")
	entries, err := os.ReadDir(config.PerformerFoldersDir)
	if err != nil {
		log.Printf("Failed to read performer folders directory: %v", err)
		return
	}

	existingPerformers, err := database.GetAllPerformerNames()
	if err != nil {
		log.Printf("Failed to get existing performer names from DB: %v", err)
		return
	}

	existingPerformerMap := make(map[string]bool)
	for _, pName := range existingPerformers {
		existingPerformerMap[pName] = true
	}

	var wg sync.WaitGroup
	newPerformersToInsert := make(chan models.Performer, len(entries)) // Channel to collect new performers

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

			var performer models.Performer
			var existingPerformerJSON string
			err := database.GetDB().QueryRow("SELECT data FROM performers WHERE name = ?", pName).Scan(&existingPerformerJSON)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("Error querying existing performer %s from database: %v", pName, err)
				return
			} else if err == sql.ErrNoRows { // Performer does not exist, initialize with default template
				performer = database.NewDefaultPerformer()
				performer.Name = pName // Ensure name is set
			} else { // Performer exists, unmarshal existing data
				if err := json.Unmarshal([]byte(existingPerformerJSON), &performer); err != nil {
					log.Printf("Error unmarshaling existing performer %s data: %v", pName, err)
					return
				}
			}

			// Fetch performer data from Adultdatalink API
			performerDataBytes, err := FetchPerformerData(pName)
			if err != nil {
				log.Printf("Failed to fetch data for performer %s from Adultdatalink API: %v", pName, err)
			} else {
				var rawAPIResponse map[string]interface{}
				if err := json.Unmarshal(performerDataBytes, &rawAPIResponse); err != nil {
					log.Printf("Failed to unmarshal raw Adultdatalink API response for %s: %v", pName, err)
				} else {
					PopulatePerformerFromAPIResponse(&performer, rawAPIResponse)
					log.Printf("Successfully fetched Adultdatalink API data for %s.", pName)
				}
			}

			// Scan for previews (.mkv and image files)
			performerFolderPath := filepath.Join(config.PerformerFoldersDir, pName)
			performerEntries, err := os.ReadDir(performerFolderPath)
			if err != nil {
				log.Printf("Failed to read performer folder %s for previews: %v", performerFolderPath, err)
			} else {
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
				performer.Previews = previews
				log.Printf("Found %d previews for %s.", len(previews), pName)
			}
			newPerformersToInsert <- performer // Send processed performer to channel
		}(performerName)
	}

	wg.Wait()                    // Wait for all goroutines to finish
	close(newPerformersToInsert) // Close the channel after all goroutines are done

	// Collect all new performers and insert them in a batch
	var performersToBatchInsert []models.Performer
	for p := range newPerformersToInsert {
		performersToBatchInsert = append(performersToBatchInsert, p)
	}

	if len(performersToBatchInsert) > 0 {
		log.Printf("Batch inserting %d new performers into database.", len(performersToBatchInsert))
		tx, err := database.GetDB().Begin()
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

func UpdatePerformerPreviewsTask() {
	log.Println("Starting update performer previews task...")

	rows, err := database.GetDB().Query("SELECT name, data FROM performers")
	if err != nil {
		log.Printf("Failed to query performers for preview update: %v", err)
		return
	}
	defer rows.Close()

	var wg sync.WaitGroup
	performersToUpdate := make(chan models.Performer, 100) // Buffer channel

	for rows.Next() {
		var name string
		var performerJSON string
		if err := rows.Scan(&name, &performerJSON); err != nil {
			log.Printf("Failed to scan performer data for preview update: %v", err)
			continue
		}

		var performer models.Performer
		if err := json.Unmarshal([]byte(performerJSON), &performer); err != nil {
			log.Printf("Failed to unmarshal performer JSON for %s: %v", name, err)
			continue
		}

		wg.Add(1)
		go func(p models.Performer) {
			defer wg.Done()

			performerFolderPath := filepath.Join(config.PerformerFoldersDir, p.Name)
			currentPreviews := []string{}

			performerEntries, err := os.ReadDir(performerFolderPath)
			if err != nil {
				log.Printf("Failed to read performer folder %s for previews: %v", performerFolderPath, err)
			} else {
				for _, entry := range performerEntries {
					ext := strings.ToLower(filepath.Ext(entry.Name()))
					if !entry.IsDir() && ext == ".mkv" {
						relativePath, err := filepath.Rel(config.PerformerFoldersDir, filepath.Join(performerFolderPath, entry.Name()))
						if err != nil {
							log.Printf("Failed to get relative path for preview %s: %v", entry.Name(), err)
							continue
						}
						currentPreviews = append(currentPreviews, filepath.ToSlash(relativePath))
					}
				}
			}

			// Compare current previews with stored previews
			if !CompareStringSlices(p.Previews, currentPreviews) {
				log.Printf("Previews changed for %s. Updating...", p.Name)
				p.Previews = currentPreviews
				// If default preview is no longer valid, clear it
				if p.DefaultPreview != "" && !ContainsString(currentPreviews, p.DefaultPreview) {
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
		_, err = database.GetDB().Exec("UPDATE performers SET data = ? WHERE name = ?", string(updatedJSON), p.Name)
		if err != nil {
			log.Printf("Failed to update performer %s in database: %v", p.Name, err)
		}
	}
	log.Println("Update performer previews task completed.")
}

func CompareStringSlices(a, b []string) bool {
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

func ContainsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
