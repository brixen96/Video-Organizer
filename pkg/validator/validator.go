package validator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Validator provides input validation functions
type Validator struct {
	errors map[string][]string
}

// New creates a new Validator instance
func New() *Validator {
	return &Validator{
		errors: make(map[string][]string),
	}
}

// AddError adds a validation error for a field
func (v *Validator) AddError(field, message string) {
	v.errors[field] = append(v.errors[field], message)
}

// HasErrors returns true if there are any validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors
func (v *Validator) Errors() map[string][]string {
	return v.errors
}

// FirstError returns the first error message, if any
func (v *Validator) FirstError() string {
	for _, messages := range v.errors {
		if len(messages) > 0 {
			return messages[0]
		}
	}
	return ""
}

// Required checks if a string is not empty
func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, fmt.Sprintf("%s is required", field))
	}
}

// MinLength checks if a string meets minimum length
func (v *Validator) MinLength(field, value string, min int) {
	if len(value) < min {
		v.AddError(field, fmt.Sprintf("%s must be at least %d characters", field, min))
	}
}

// MaxLength checks if a string doesn't exceed maximum length
func (v *Validator) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.AddError(field, fmt.Sprintf("%s must not exceed %d characters", field, max))
	}
}

// Pattern checks if a string matches a regex pattern
func (v *Validator) Pattern(field, value, pattern, message string) {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil || !matched {
		v.AddError(field, message)
	}
}

// SafePath validates that a path is safe (no path traversal)
func SafePath(path string) error {
	// Clean the path
	cleaned := filepath.Clean(path)
	
	// Check for path traversal attempts
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("path contains invalid '..' sequence")
	}
	
	// Check for absolute path on Windows (starts with drive letter)
	if len(cleaned) >= 2 && cleaned[1] == ':' {
		return fmt.Errorf("absolute paths are not allowed")
	}
	
	// Check for absolute path on Unix (starts with /)
	if strings.HasPrefix(cleaned, "/") {
		return fmt.Errorf("absolute paths are not allowed")
	}
	
	return nil
}

// SafeName validates that a name is safe for use in filenames
func SafeName(name string) error {
	// Check for empty name
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	
	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("name contains invalid character: %s", char)
		}
	}
	
	// Check for path traversal
	if strings.Contains(name, "..") {
		return fmt.Errorf("name contains invalid '..' sequence")
	}
	
	// Check length
	if len(name) > 255 {
		return fmt.Errorf("name is too long (max 255 characters)")
	}
	
	return nil
}

// PerformerName validates a performer name
func PerformerName(name string) error {
	if err := SafeName(name); err != nil {
		return fmt.Errorf("invalid performer name: %w", err)
	}
	
	// Additional performer-specific validation
	if len(name) < 2 {
		return fmt.Errorf("performer name must be at least 2 characters")
	}
	
	return nil
}

// VideoName validates a video filename
func VideoName(name string) error {
	if err := SafeName(name); err != nil {
		return fmt.Errorf("invalid video name: %w", err)
	}
	
	// Check for valid video extension
	ext := strings.ToLower(filepath.Ext(name))
	validExts := []string{".mp4", ".mkv", ".avi", ".mov", ".webm"}
	valid := false
	for _, validExt := range validExts {
		if ext == validExt {
			valid = true
			break
		}
	}
	
	if !valid {
		return fmt.Errorf("invalid video file extension: %s (allowed: %v)", ext, validExts)
	}
	
	return nil
}

// LibraryPath validates a library path
func LibraryPath(path string) error {
	// Allow absolute paths for libraries
	cleaned := filepath.Clean(path)
	
	// Check for path traversal in relative paths
	if !filepath.IsAbs(cleaned) && strings.Contains(cleaned, "..") {
		return fmt.Errorf("path contains invalid '..' sequence")
	}
	
	// Check length
	if len(cleaned) > 1024 {
		return fmt.Errorf("path is too long (max 1024 characters)")
	}
	
	return nil
}

// SanitizeString removes potentially dangerous characters from a string
func SanitizeString(s string) string {
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")
	
	// Remove control characters except newline and tab
	cleaned := strings.Builder{}
	for _, r := range s {
		if r == '\n' || r == '\t' || r >= 32 {
			cleaned.WriteRune(r)
		}
	}
	
	return strings.TrimSpace(cleaned.String())
}
