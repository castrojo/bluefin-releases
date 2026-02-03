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
	AppsTotal            int `json:"appsTotal"`
	AppsWithGitHubRepo   int `json:"appsWithGitHubRepo"`
	AppsWithChangelogs   int `json:"appsWithChangelogs"`
	TotalReleases        int `json:"totalReleases"`
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
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	Summary         string      `json:"summary"`
	Description     string      `json:"description,omitempty"`
	DeveloperName   string      `json:"developerName,omitempty"`
	Icon            string      `json:"icon,omitempty"`
	ProjectLicense  string      `json:"projectLicense,omitempty"`
	Categories      []string    `json:"categories,omitempty"`
	UpdatedAt       string      `json:"updatedAt,omitempty"`
	Version         string      `json:"currentReleaseVersion,omitempty"`
	ReleaseDate     string      `json:"currentReleaseDate,omitempty"`
	FlathubURL      string      `json:"flathubUrl"`
	SourceRepo      *SourceRepo `json:"sourceRepo,omitempty"`
	Releases        []Release   `json:"releases,omitempty"`
	FetchedAt       time.Time   `json:"fetchedAt"`
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

// FlathubApp represents the raw structure from Flathub API feed
type FlathubApp struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	Summary               string   `json:"summary"`
	DeveloperName         string   `json:"developer_name"`
	Icon                  string   `json:"icon"`
	ProjectLicense        string   `json:"project_license"`
	Categories            []string `json:"categories"`
	CurrentReleaseVersion string   `json:"current_release_version"`
	CurrentReleaseDate    string   `json:"current_release_date"`
}

// FlathubAppDetails represents detailed app information from Flathub API
type FlathubAppDetails struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
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
