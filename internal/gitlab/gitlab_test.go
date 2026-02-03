package gitlab

import (
	"context"
	"testing"

	"github.com/castrojo/bluefin-releases/internal/models"
)

func TestFetchGitLabReleases(t *testing.T) {
	tests := []struct {
		name      string
		repoURL   string
		owner     string
		repo      string
		wantError bool
		minCount  int
	}{
		{
			name:      "GNOME File Roller from gitlab.gnome.org",
			repoURL:   "https://gitlab.gnome.org/GNOME/file-roller",
			owner:     "GNOME",
			repo:      "file-roller",
			wantError: false,
			minCount:  0, // May have releases
		},
		{
			name:      "GNOME Firmware from gitlab.gnome.org",
			repoURL:   "https://gitlab.gnome.org/World/gnome-firmware",
			owner:     "World",
			repo:      "gnome-firmware",
			wantError: false,
			minCount:  0, // May have releases
		},
		{
			name:      "Invalid project path",
			repoURL:   "https://gitlab.com/invalid",
			owner:     "",
			repo:      "",
			wantError: true,
			minCount:  0,
		},
	}

	ctx := context.Background()
	token := "" // Test without token (public API)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releases, err := fetchGitLabReleases(ctx, token, tt.repoURL, tt.owner, tt.repo)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(releases) < tt.minCount {
				t.Errorf("Expected at least %d releases, got %d", tt.minCount, len(releases))
			}

			// Validate release structure if we got any
			for _, release := range releases {
				if release.Version == "" {
					t.Errorf("Release has empty version")
				}
				if release.Type != "gitlab-release" {
					t.Errorf("Expected type 'gitlab-release', got '%s'", release.Type)
				}
				if release.Date.IsZero() {
					t.Errorf("Release has zero date")
				}
			}
		})
	}
}

func TestEnrichWithGitLabReleases(t *testing.T) {
	apps := []models.App{
		{
			ID:   "test.app.GitLab",
			Name: "Test GitLab App",
			SourceRepo: &models.SourceRepo{
				Type:  "gitlab",
				URL:   "https://gitlab.gnome.org/World/gnome-firmware",
				Owner: "World",
				Repo:  "gnome-firmware",
			},
			Releases: []models.Release{
				{
					Version: "existing-1.0",
					Type:    "appstream",
				},
			},
		},
		{
			ID:   "test.app.GitHub",
			Name: "Test GitHub App",
			SourceRepo: &models.SourceRepo{
				Type: "github",
				URL:  "https://github.com/test/test",
			},
		},
		{
			ID:         "test.app.NoRepo",
			Name:       "Test No Repo App",
			SourceRepo: nil,
		},
	}

	enriched := EnrichWithGitLabReleases(apps)

	// Check that we didn't lose any apps
	if len(enriched) != len(apps) {
		t.Errorf("Expected %d apps, got %d", len(apps), len(enriched))
	}

	// Find the GitLab app
	var gitlabApp *models.App
	for i := range enriched {
		if enriched[i].ID == "test.app.GitLab" {
			gitlabApp = &enriched[i]
			break
		}
	}

	if gitlabApp == nil {
		t.Fatal("GitLab app not found in enriched apps")
	}

	// The app should still have its original release
	foundOriginal := false
	for _, release := range gitlabApp.Releases {
		if release.Version == "existing-1.0" {
			foundOriginal = true
			break
		}
	}

	if !foundOriginal {
		t.Error("Original release was lost during enrichment")
	}
}
