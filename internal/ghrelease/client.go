package ghrelease

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client uploads podcast MP3s as GitHub Release assets.
type Client struct {
	token string // GITHUB_TOKEN
	repo  string // "owner/repo" from GITHUB_REPOSITORY
	http  *http.Client
}

type releaseResponse struct {
	ID        int    `json:"id"`
	UploadURL string `json:"upload_url"`
	HTMLURL   string `json:"html_url"`
}

type assetResponse struct {
	BrowserDownloadURL string `json:"browser_download_url"`
}

// New returns a GitHub Release client.
func New(token, repo string) *Client {
	return &Client{token: token, repo: repo, http: &http.Client{Timeout: 120 * time.Second}}
}

// UploadPodcast creates a release and uploads the MP3, returning the public download URL.
func (c *Client) UploadPodcast(mp3 []byte, t time.Time) (string, error) {
	tag := "audio-" + t.Format("20060102-150405")
	name := fmt.Sprintf("Podcast — %02d/%02d/%d", t.Day(), t.Month(), t.Year())

	rel, err := c.createRelease(tag, name)
	if err != nil {
		return "", fmt.Errorf("criar release: %w", err)
	}

	filename := fmt.Sprintf("ada-alan-news-%s.mp3", t.Format("20060102"))
	url, err := c.uploadAsset(rel, filename, mp3)
	if err != nil {
		return "", fmt.Errorf("upload asset: %w", err)
	}
	return url, nil
}

func (c *Client) createRelease(tag, name string) (*releaseResponse, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases", c.repo)

	payload, err := json.Marshal(map[string]any{
		"tag_name":   tag,
		"name":       name,
		"draft":      false,
		"prerelease": false,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GitHub API status %d: %s", resp.StatusCode, string(raw))
	}

	var rel releaseResponse
	if err := json.Unmarshal(raw, &rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func (c *Client) uploadAsset(rel *releaseResponse, filename string, data []byte) (string, error) {
	// upload_url looks like "https://uploads.github.com/repos/owner/repo/releases/123/assets{?name,label}"
	uploadURL := strings.Split(rel.UploadURL, "{")[0] + "?name=" + filename

	req, err := http.NewRequest(http.MethodPost, uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "audio/mpeg")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload status %d: %s", resp.StatusCode, string(raw))
	}

	var a assetResponse
	if err := json.Unmarshal(raw, &a); err != nil {
		return "", err
	}
	return a.BrowserDownloadURL, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
}
