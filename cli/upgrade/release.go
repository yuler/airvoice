package upgrade

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the GitHub repository base URL used for releases.
	DefaultBaseURL = "https://github.com/yuler/airvoice"

	maxChecksumBytes = 1 << 20 // 1 MiB
)

// Client fetches release metadata, checksums, and CLI binaries.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	UserAgent  string
}

// NewClient returns a release client targeting DefaultBaseURL.
func NewClient(userAgent string) *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		UserAgent: userAgent,
	}
}

func (c *Client) baseURL() string {
	if c != nil && strings.TrimSpace(c.BaseURL) != "" {
		return strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	}
	return DefaultBaseURL
}

func (c *Client) httpClient() *http.Client {
	if c != nil && c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 2 * time.Minute}
}

func (c *Client) newRequest(ctx context.Context, method, targetURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, targetURL, nil)
	if err != nil {
		return nil, err
	}
	ua := "airvoice-cli"
	if c != nil && strings.TrimSpace(c.UserAgent) != "" {
		ua = strings.TrimSpace(c.UserAgent)
	}
	req.Header.Set("User-Agent", ua)
	return req, nil
}

// FetchLatestVersion resolves the latest release tag via GitHub's /releases/latest redirect.
func (c *Client) FetchLatestVersion(ctx context.Context) (string, error) {
	base := c.httpClient()
	noRedirect := *base
	noRedirect.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := c.newRequest(ctx, http.MethodGet, c.baseURL()+"/releases/latest")
	if err != nil {
		return "", err
	}
	resp, err := noRedirect.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch latest release: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		return "", fmt.Errorf("unexpected status %d resolving latest release", resp.StatusCode)
	}
	loc := strings.TrimSpace(resp.Header.Get("Location"))
	if loc == "" {
		return "", fmt.Errorf("latest release redirect missing Location header")
	}
	return parseVersionFromTagLocation(loc)
}

func parseVersionFromTagLocation(location string) (string, error) {
	parsed, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("parse redirect location %q: %w", location, err)
	}
	const marker = "/releases/tag/"
	idx := strings.Index(parsed.Path, marker)
	if idx < 0 {
		return "", fmt.Errorf("redirect location %q does not contain %s", location, marker)
	}
	tag := strings.Trim(parsed.Path[idx+len(marker):], "/")
	if slash := strings.IndexByte(tag, '/'); slash >= 0 {
		tag = tag[:slash]
	}
	if _, err := Parse(tag); err != nil {
		return "", fmt.Errorf("invalid release tag %q: %w", tag, err)
	}
	return Normalize(tag), nil
}

// FetchChecksum retrieves the expected SHA-256 hex digest for asset from checksums-cli.txt.
func (c *Client) FetchChecksum(ctx context.Context, version, asset string) (string, error) {
	ver := Normalize(version)
	checksumURL := fmt.Sprintf("%s/releases/download/v%s/checksums-cli.txt", c.baseURL(), ver)
	req, err := c.newRequest(ctx, http.MethodGet, checksumURL)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("download checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksums-cli.txt not available for v%s (HTTP %d)", ver, resp.StatusCode)
	}
	return parseChecksum(io.LimitReader(resp.Body, maxChecksumBytes), asset)
}

func parseChecksum(r io.Reader, asset string) (string, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		sum := strings.ToLower(fields[0])
		name := strings.TrimPrefix(fields[1], "*")
		if name != asset {
			continue
		}
		if len(sum) != 64 {
			return "", fmt.Errorf("invalid SHA-256 digest length for %s", asset)
		}
		if _, err := hex.DecodeString(sum); err != nil {
			return "", fmt.Errorf("invalid SHA-256 digest for %s: %w", asset, err)
		}
		return sum, nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read checksums: %w", err)
	}
	return "", fmt.Errorf("checksum for %s not found in checksums-cli.txt", asset)
}

// DownloadVerified downloads asset into a temporary file in destDir and verifies its SHA-256 digest.
// The temporary file is created before issuing the HTTP request so permission failures fail fast.
func (c *Client) DownloadVerified(ctx context.Context, version, asset, expectedSHA256, destDir string) (string, error) {
	tmp, err := os.CreateTemp(destDir, ".airvoice-upgrade-*.tmp")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	ver := Normalize(version)
	assetURL := fmt.Sprintf("%s/releases/download/v%s/%s", c.baseURL(), ver, asset)
	req, err := c.newRequest(ctx, http.MethodGet, assetURL)
	if err != nil {
		cleanup()
		return "", err
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		cleanup()
		return "", fmt.Errorf("download %s: %w", asset, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		cleanup()
		return "", fmt.Errorf("failed to download %s for v%s (HTTP %d)", asset, ver, resp.StatusCode)
	}

	hasher := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, hasher), resp.Body); err != nil {
		cleanup()
		return "", fmt.Errorf("save %s: %w", asset, err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("close temporary binary: %w", err)
	}

	actual := hex.EncodeToString(hasher.Sum(nil))
	if actual != strings.ToLower(expectedSHA256) {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("checksum mismatch for %s: expected %s, got %s", asset, expectedSHA256, actual)
	}

	return tmpPath, nil
}
