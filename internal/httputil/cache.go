package httputil

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"cchoice/internal/constants"
	"cchoice/internal/metrics"
)

const (
	// cacheControlImmutable is used for fingerprinted assets: the URL changes
	// whenever the file does, so the response can be kept for a year.
	cacheControlImmutable = "public, max-age=31536000, immutable"

	// cacheControlLongLived is used for assets requested without a fingerprint.
	cacheControlLongLived = "public, max-age=2592000, stale-while-revalidate=86400" // 30 days, stale 1 day
)

// immutableAssetExts lists the extensions served out of the static filesystem
// that are generated or replaced wholesale on deploy.
var immutableAssetExts = map[string]struct{}{
	".avif":  {},
	".css":   {},
	".gif":   {},
	".ico":   {},
	".jpeg":  {},
	".jpg":   {},
	".js":    {},
	".json":  {},
	".mjs":   {},
	".otf":   {},
	".png":   {},
	".svg":   {},
	".ttf":   {},
	".webp":  {},
	".woff":  {},
	".woff2": {},
}

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
	isFingerprinted := r.URL.Query().Get(constants.QueryParamAssetVersion) != ""

	switch {
	case isFingerprinted && isImmutableAsset(r.URL.Path):
		// The fingerprint changes whenever the file does, so the response never
		// needs to be revalidated.
		w.Header().Set("Cache-Control", cacheControlImmutable)
	case isImmutableAsset(r.URL.Path):
		w.Header().Set("Cache-Control", cacheControlLongLived)
	case r.URL.Path == "robots.txt":
		w.Header().Set("Cache-Control", "public, max-age=604800") // 1 week
	default:
		w.Header().Set("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400") // 1 hour, stale 1 day
	}

	w.Header().Set("Vary", "Accept-Encoding")
}

// isImmutableAsset reports whether name points at a build artifact whose
// contents are replaced rather than edited in place, which makes it safe to
// cache for a long time.
func isImmutableAsset(name string) bool {
	_, ok := immutableAssetExts[strings.ToLower(path.Ext(name))]
	return ok
}

// SetNoCacheHeaders sets headers to prevent caching of error responses
// This is especially important for 404 responses to ensure browsers don't cache
// missing assets, allowing them to be fetched immediately after upload.
func SetNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
