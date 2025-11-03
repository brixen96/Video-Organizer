package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite driver

	"video-organizer/internal/models"
)

var db *sql.DB

func NewDefaultPerformer() models.Performer {
	return models.Performer{
		Appearance: map[string]interface{}{
			"ethnicity":  "Undefined",
			"boobs":      "Undefined",
			"bust":       "Undefined",
			"cup":        "Undefined",
			"bra":        "Undefined",
			"waist":      "Undefined",
			"hip":        "Undefined",
			"butt":       "Undefined",
			"height":     "Undefined",
			"weight":     "Undefined",
			"hair_color": "Undefined", "eye_color": "Undefined",
			"piercings":          "Undefined",
			"piercing_locations": "Undefined",
			"tattoos":            "Undefined",
			"tattoo_locations":   "Undefined",
			"shoe_size":          "Undefined",
			"body_type":          "Undefined",
			"underarm_hair":      "Undefined",
			"pubic_hair":         "Undefined",
		},
		Performances:          make(map[string]interface{}),
		SocialMedia:           make(map[string]interface{}),
		PlatformViews:         make(map[string]interface{}),
		PlatformVideoCounts:   make(map[string]interface{}),
		PlatformProfileCounts: make(map[string]interface{}),
		Tags:                  []string{},
		ExternalLinks:         []map[string]string{},
		Bios:                  make(map[string]string),
		OfficialWebsite:       "Undefined",
		FeatureDancer:         "Undefined",
		DateOfBirth:           "Undefined",
		Age:                   "Undefined",
		SexualOrientation:     "Undefined", AstrologicalSign: "Undefined",
		Profession:        "Undefined",
		CareerStatus:      "Undefined",
		CareerStart:       "Undefined",
		CareerEnd:         "Undefined",
		DateOfDeath:       "Undefined",
		PlaceOfBirth:      "Undefined",
		Nationality:       "Undefined",
		Rank:              "Undefined",
		Country:           "Undefined",
		Avatar:            "Undefined",
		Subscribers:       0,
		Rating:            0,
		TotalViews:        0,
		TotalVideoCount:   0,
		TotalPlatformHits: 0,
		Aliases:           []string{},
		ImageURL:          "Undefined",
		SceneCount:        0,
		Previews:          []string{},
		DefaultPreview:    "",
		Zoo:               "undefined",
	}
}

func InitDB() {
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

	CREATE TABLE IF NOT EXISTS libraries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		path TEXT NOT NULL,
		is_default INTEGER DEFAULT 0
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
	log.Println("Database initialized and performers table ensured.")
}

func GetDB() *sql.DB {
	return db
}

func GetAllPerformerNames() ([]string, error) {
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
