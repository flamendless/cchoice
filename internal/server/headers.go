package server

import (
	"cchoice/internal/metrics"
	"fmt"
	"net/http"
	"os"
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

	lastMod := info.ModTime().UTC().Format(http.TimeFormat)
	etag := fmt.Sprintf(`"%x-%x"`, info.Size(), info.ModTime().Unix())

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
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
