package models

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// OutputData represents the top-level JSON structure (follows firehose pattern)
type OutputData struct {
	Metadata Metadata `json:"metadata"`
	Apps     []App    `json:"apps"`
}

// Metadata contains build metadata and statistics
type Metadata struct {
	SchemaVersion string      `json:"schemaVersion"`
	GeneratedAt   string      `json:"generatedAt"`
	GeneratedBy   string      `json:"generatedBy"`
	BuildDuration string      `json:"buildDuration"`
	Stats         Stats       `json:"stats"`
	Performance   Performance `json:"performance"`
}

// Stats contains aggregate statistics
type Stats struct {
	AppsTotal          int `json:"appsTotal"`
	AppsWithGitHubRepo int `json:"appsWithGitHubRepo"`
	AppsWithChangelogs int `json:"appsWithChangelogs"`
	TotalReleases      int `json:"totalReleases"`
}

// Performance contains timing breakdown
type Performance struct {
	FlathubFetchDuration string `json:"flathubFetchDuration"`
	DetailsFetchDuration string `json:"detailsFetchDuration"`
	GitHubFetchDuration  string `json:"githubFetchDuration"`
	OutputDuration       string `json:"outputDuration"`
}

// App represents a Flathub application (similar to Release in firehose)
type App struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	Summary           string        `json:"summary"`
	Description       string        `json:"description,omitempty"`
	DeveloperName     string        `json:"developerName,omitempty"`
	Icon              string        `json:"icon,omitempty"`
	ProjectLicense    string        `json:"projectLicense,omitempty"`
	Categories        []string      `json:"categories,omitempty"`
	UpdatedAt         string        `json:"updatedAt,omitempty"`
	Version           string        `json:"currentReleaseVersion,omitempty"`
	ReleaseDate       string        `json:"currentReleaseDate,omitempty"`
	FlathubURL        string        `json:"flathubUrl"`
	SourceRepo        *SourceRepo   `json:"sourceRepo,omitempty"`
	Releases          []Release     `json:"releases,omitempty"`
	FetchedAt         time.Time     `json:"fetchedAt"`
	InstallsLastMonth int           `json:"installsLastMonth,omitempty"`
	FavoritesCount    int           `json:"favoritesCount,omitempty"`
	IsVerified        bool          `json:"isVerified"`
	VerificationInfo  *Verification `json:"verificationInfo,omitempty"`
	AppSet            string        `json:"appSet,omitempty"` // "core" or "dx"
	PackageType       string        `json:"packageType"`      // "flatpak" or "homebrew"
	HomebrewInfo      *HomebrewInfo `json:"homebrewInfo,omitempty"`
}

// HomebrewInfo contains Homebrew-specific package information
type HomebrewInfo struct {
	Formula      string   `json:"formula"`                // Formula name (e.g., "bat", "gh")
	FullName     string   `json:"fullName,omitempty"`     // Full formula name with tap (e.g., "homebrew/core/bat")
	Tap          string   `json:"tap,omitempty"`          // Tap name (e.g., "homebrew/core")
	Homepage     string   `json:"homepage,omitempty"`     // Homepage URL
	Versions     []string `json:"versions,omitempty"`     // Available versions
	Dependencies []string `json:"dependencies,omitempty"` // Package dependencies
	Caveats      string   `json:"caveats,omitempty"`      // Installation caveats/notes
}

// Verification contains app verification details from Flathub
type Verification struct {
	Method        string  `json:"method"` // "website", "login_provider", etc.
	LoginName     *string `json:"loginName,omitempty"`
	LoginProvider *string `json:"loginProvider,omitempty"`
	Website       *string `json:"website,omitempty"`
}

// SourceRepo contains information about the app's source repository
type SourceRepo struct {
	Type  string `json:"type"` // "github", "gitlab", "other"
	URL   string `json:"url"`
	Owner string `json:"owner,omitempty"`
	Repo  string `json:"repo,omitempty"`
}

// Release represents a single release/changelog entry (from GitHub or Flathub)
type Release struct {
	Version     string    `json:"version"`
	Date        time.Time `json:"date"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	URL         string    `json:"url,omitempty"`
	Type        string    `json:"type"` // "github-release", "appstream"
}

// FlathubApp represents the raw structure from Flathub API collection endpoint
type FlathubApp struct {
	AppID                 string        `json:"app_id"`
	Name                  string        `json:"name"`
	Summary               string        `json:"summary"`
	Description           string        `json:"description"`
	DeveloperName         string        `json:"developer_name"`
	Icon                  string        `json:"icon"`
	ProjectLicense        string        `json:"project_license"`
	MainCategories        StringOrArray `json:"main_categories"`
	SubCategories         []string      `json:"sub_categories"`
	UpdatedAt             int64         `json:"updated_at"`
	InstallsLastMonth     int           `json:"installs_last_month"`
	FavoritesCount        int           `json:"favorites_count"`
	VerificationVerified  bool          `json:"verification_verified"`
	VerificationMethod    string        `json:"verification_method"`
	VerificationLoginName *string       `json:"verification_login_name"`
	VerificationWebsite   *string       `json:"verification_website"`
}

// StringOrArray handles JSON fields that can be either string or array
type StringOrArray []string

// UnmarshalJSON implements custom unmarshalling for string or array
func (s *StringOrArray) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = []string{str}
		return nil
	}

	// Try to unmarshal as array
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*s = arr
		return nil
	}

	return fmt.Errorf("field must be string or array")
}

// FlathubCollectionResponse represents the response from /api/v2/collection/recently-updated
type FlathubCollectionResponse struct {
	Hits        []FlathubApp `json:"hits"`
	TotalHits   int          `json:"totalHits"`
	HitsPerPage int          `json:"hitsPerPage"`
	Page        int          `json:"page"`
	TotalPages  int          `json:"totalPages"`
}

// FlathubAppDetails represents detailed app information from Flathub API
type FlathubAppDetails struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	Icon        string                `json:"icon"`
	URLs        map[string]string     `json:"urls"`
	Releases    []FlathubReleaseEntry `json:"releases"`
}

// FlathubReleaseEntry represents a release from Flathub appstream metadata
type FlathubReleaseEntry struct {
	Version     string `json:"version"`
	Date        string `json:"date"`
	Description string `json:"description"`
}

// WriteJSON writes OutputData to a JSON file (pretty-printed)
func (o *OutputData) WriteJSON(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false) // Keep URLs readable

	if err := encoder.Encode(o); err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}

	return nil
}

// FetchResults holds the results of parallel app fetching
type FetchResults struct {
	Apps []App
}
