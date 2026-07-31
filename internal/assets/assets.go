// Package assets fingerprints the files served from the static filesystem so
// that static asset URLs can be cache-busted per release while the responses
// themselves are cached for a long time by browsers and CDNs.
package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"sync"
	"time"

	"cchoice/cmd/web/static"
	"cchoice/internal/conf"
)

const fingerprintLength = 12

var (
	fingerprints sync.Map

	// bootVersion changes on every process start. Outside production the
	// asset version is tied to the boot instead of to file contents so that
	// a rebuild always serves fresh assets even when nothing was tagged.
	bootVersion = strconv.FormatInt(time.Now().UnixNano(), 36)
)

// Version returns a short, stable token identifying the current contents of the
// static asset at path. Files that cannot be inspected — a missing file, or a
// build without a static filesystem — fall back to the release tag so the token
// still changes on every deploy.
func Version(path string) string {
	if !conf.Conf().IsProd() {
		return bootVersion
	}

	if cached, ok := fingerprints.Load(path); ok {
		return cached.(string)
	}

	version := fingerprint(path)
	fingerprints.Store(path, version)
	return version
}

func fingerprint(path string) string {
	fsys := static.GetFS()
	if fsys == nil {
		return conf.GitTagProd
	}

	info, err := fs.Stat(fsys, normalize(path))
	if err != nil {
		return conf.GitTagProd
	}

	h := sha256.New()
	if _, err := fmt.Fprintf(h, "%d-%d", info.Size(), info.ModTime().UnixNano()); err != nil {
		return conf.GitTagProd
	}
	return hex.EncodeToString(h.Sum(nil))[:fingerprintLength]
}

// normalize turns a request path such as "/static/css/tailwind.css" into a path
// relative to the static filesystem root.
func normalize(path string) string {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, "static/")
	return path
}
