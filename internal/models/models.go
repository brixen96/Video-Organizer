package models

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
	Appearance            map[string]interface{} `json:"appearance,omitempty"`
	Performances          map[string]interface{} `json:"performances,omitempty"`
	SocialMedia           map[string]interface{} `json:"social_media,omitempty"`
	PlatformViews         map[string]interface{} `json:"platform_views,omitempty"`
	PlatformVideoCounts   map[string]interface{} `json:"platform_video_counts,omitempty"`
	PlatformProfileCounts map[string]interface{} `json:"platform_profile_counts,omitempty"`
	Tags                  []string               `json:"tags,omitempty"`
	ExternalLinks         []map[string]string    `json:"external_links,omitempty"`
	Bios                  map[string]string      `json:"bios,omitempty"`
	OfficialWebsite       string                 `json:"official_website,omitempty"`
	FeatureDancer         string                 `json:"feature_dancer,omitempty"`
	DateOfBirth           string                 `json:"date_of_birth,omitempty"`
	Age                   string                 `json:"age,omitempty"`
	SexualOrientation     string                 `json:"sexual_orientation,omitempty"`
	AstrologicalSign      string                 `json:"astrological_sign,omitempty"`
	Profession            string                 `json:"profession,omitempty"`
	CareerStatus          string                 `json:"career_status,omitempty"`
	CareerStart           string                 `json:"career_start,omitempty"`
	CareerEnd             string                 `json:"career_end,omitempty"`
	DateOfDeath           string                 `json:"date_of_death,omitempty"`
	PlaceOfBirth          string                 `json:"place_of_birth,omitempty"`
	Nationality           string                 `json:"nationality,omitempty"`
	Rank                  string                 `json:"rank,omitempty"`
	Country               string                 `json:"country,omitempty"`
	Avatar                string                 `json:"avatar,omitempty"`
	Subscribers           int                    `json:"subscribers,omitempty"`
	Rating                int                    `json:"rating,omitempty"`
	TotalViews            int                    `json:"total_views,omitempty"`
	TotalVideoCount       int                    `json:"total_video_count,omitempty"`
	TotalPlatformHits     int                    `json:"total_platform_hits,omitempty"`
	Aliases               []string               `json:"aliases,omitempty"`
	ImageURL              string                 `json:"image_url,omitempty"`

	// Our top-level fields
	Name           string   `json:"name"`
	SceneCount     int      `json:"scene_count"`
	Previews       []string `json:"previews,omitempty"`
	DefaultPreview string   `json:"default_preview,omitempty"`
	Zoo            string   `json:"zoo,omitempty"`
}
type VideoInfo struct {
	Name       string        `json:"name"`
	Thumbnail  string        `json:"thumbnail"`
	Metadata   FFProbeOutput `json:"metadata"`
	Performers []string      `json:"performers,omitempty"`
}

// Library represents a filesystem library configured in the app
type Library struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	IsDefault bool   `json:"isDefault"`
}
