package installer

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"openriot/logger"
)

// Mirror represents an OpenBSD mirror
type Mirror struct {
	Name   string
	URL    string
	Region string
}

// DefaultMirrors is the list of verified OpenBSD mirrors that support snapshots
var DefaultMirrors = []Mirror{
	{Name: "CDN", URL: "https://cdn.openbsd.org/pub/OpenBSD", Region: "na"},
	{Name: "Sonic", URL: "https://mirrors.sonic.net/pub/OpenBSD", Region: "na"},
	{Name: "Constant", URL: "https://openbsd.mirror.constant.com/pub/OpenBSD", Region: "na"},
	{Name: "Eu", URL: "https://ftp.eu.openbsd.org/pub/OpenBSD", Region: "eu"},
	{Name: "France", URL: "https://ftp.fr.openbsd.org/pub/OpenBSD", Region: "eu"},
	{Name: "Germany", URL: "https://ftp.spline.de/pub/OpenBSD", Region: "eu"},
	{Name: "Japan", URL: "https://ftp.jaist.ac.jp/pub/OpenBSD", Region: "ap"},
	{Name: "Singapore", URL: "https://mirror.freedif.org/pub/OpenBSD", Region: "ap"},
	{Name: "Australia", URL: "https://mirror.aarnet.edu.au/pub/OpenBSD", Region: "au"},
	{Name: "New Zealand", URL: "https://mirror.fsmg.org.nz/pub/OpenBSD", Region: "au"},
}

const installurlPath = "/etc/installurl"
const pingTimeout = 3 * time.Second

// SelectFastestMirror pings all mirrors concurrently and returns the fastest
func SelectFastestMirror() (string, time.Duration, error) {
	type result struct {
		url     string
		latency time.Duration
	}

	results := make(chan result, len(DefaultMirrors))
	var wg sync.WaitGroup

	for _, m := range DefaultMirrors {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			latency, err := pingMirror(url)
			if err == nil {
				results <- result{url: url, latency: latency}
			}
		}(m.URL)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Return first result (fastest to respond)
	var fastest result
	fastestSet := false

	for r := range results {
		if !fastestSet || r.latency < fastest.latency {
			fastest = r
			fastestSet = true
		}
	}

	if !fastestSet {
		return DefaultMirrors[0].URL, 0, fmt.Errorf("no mirrors responded")
	}

	return fastest.url, fastest.latency, nil
}

// pingMirror checks if a mirror is reachable and returns latency
func pingMirror(url string) (time.Duration, error) {
	// Test connectivity by hitting the base URL with a short timeout
	client := &http.Client{
		Timeout: pingTimeout,
	}

	// Use HEAD request to minimize bandwidth
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	// Don't follow redirects
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return latency, nil
}

// WriteInstallurl writes the mirror URL to /etc/installurl and /etc/pkg_add.conf
func WriteInstallurl(mirrorURL string) error {
	if err := os.WriteFile(installurlPath, []byte(mirrorURL+"\n"), 0644); err != nil {
		return err
	}
	// Also write to pkg_add.conf for pkg_add to use
	pkgAddConf := "/etc/pkg_add.conf"
	return os.WriteFile(pkgAddConf, []byte("installpath = "+mirrorURL+"\n"), 0644)
}

// HasInstallurl returns true if /etc/installurl already exists
func HasInstallurl() bool {
	_, err := os.Stat(installurlPath)
	return err == nil
}

// GetInstallurl returns the current mirror URL from /etc/installurl
func GetInstallurl() string {
	data, err := os.ReadFile(installurlPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// hasHardcodedCDN returns true if /etc/pkg_add.conf contains the hardcoded CDN
func hasHardcodedCDN() bool {
	data, err := os.ReadFile("/etc/pkg_add.conf")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "cdn.openbsd.org")
}

// SetupMirror runs mirror detection and writes /etc/installurl if needed
// Called before package installation to ensure fastest mirror is used
func SetupMirror() {
	// Check both installurl and pkg_add.conf
	if HasInstallurl() && !hasHardcodedCDN() {
		return
	}

	logger.Info("Detecting fastest mirror...")

	mirror, latency, err := SelectFastestMirror()
	if err != nil {
		logger.Warn(fmt.Sprintf("Mirror detection failed, using CDN: %v", err))
		mirror = DefaultMirrors[0].URL
		latency = 0
	}

	if err := WriteInstallurl(mirror); err != nil {
		logger.Warn(fmt.Sprintf("Failed to write installurl: %v", err))
		return
	}

	latencyMs := latency.Milliseconds()
	logger.Done(fmt.Sprintf("Using %s (%dms)", mirror, latencyMs))
}
