package httputil

import (
	"cchoice/internal/metrics"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

func CacheHeaders(
	w http.ResponseWriter,
	r *http.Request,
	fs http.FileSystem,
	path string,
) (bool, http.File, error) {
	var file http.File
	var errFile error

	path = strings.TrimPrefix(path, "static")

	if fs != nil {
		file, errFile = fs.Open(path)
	} else {
		file, errFile = os.Open(path)
	}
	if errFile != nil {
		return false, nil, errFile
	}

	info, err := file.Stat()
	if err != nil {
		return false, nil, err
	}

	etag, err := generateETag(info, r.URL.RawQuery)
	if err != nil {
		return false, nil, err
	}
	lastMod := info.ModTime().UTC().Format(http.TimeFormat)

	setCacheControlHeaders(w, r)
	w.Header().Set("Last-Modified", lastMod)
	w.Header().Set("ETag", etag)

	if match := r.Header.Get("If-None-Match"); match == etag {
		metrics.Cache.HeadersHit()
		w.WriteHeader(http.StatusNotModified)
		return true, file, nil
	}

	if since := r.Header.Get("If-Modified-Since"); since != "" {
		if t, err := time.Parse(http.TimeFormat, since); err == nil {
			if info.ModTime().Before(t.Add(1 * time.Second)) {
				metrics.Cache.HeadersHit()
				w.WriteHeader(http.StatusNotModified)
				return true, file, nil
			}
		}
	}

	metrics.Cache.HeadersMiss()
	return false, file, nil
}

func generateETag(info os.FileInfo, rawQuery string) (string, error) {
	queryParams := parseAndSortQuery(rawQuery)
	h := sha256.New()
	if _, err := fmt.Fprintf(h, "%d-%d-%s", info.Size(), info.ModTime().Unix(), queryParams); err != nil {
		return "", err
	}
	hash := hex.EncodeToString(h.Sum(nil))[:16]
	return fmt.Sprintf(`"f%x-q%s"`, info.ModTime().Unix(), hash), nil
}

func parseAndSortQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	params := strings.Split(rawQuery, "&")
	sort.Strings(params)
	return strings.Join(params, "&")
}

func setCacheControlHeaders(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.Contains(r.URL.Path, "/static/"):
		w.Header().Set("Cache-Control", "public, max-age=86400") // 1 day
	case r.URL.Path == "robots.txt":
		w.Header().Set("Cache-Control", "public, max-age=604800") // 1 week
	default:
		w.Header().Set("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400") // 1 hour, stale 1 day
	}

	if r.URL.RawQuery != "" {
		w.Header().Set("Vary", "Accept, Accept-Encoding")
	}
}

// SetNoCacheHeaders sets headers to prevent caching of error responses
// This is especially important for 404 responses to ensure browsers don't cache
// missing assets, allowing them to be fetched immediately after upload.
func SetNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
