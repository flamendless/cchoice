package server

import (
	"cchoice/internal/utils"
	"net/http"
)

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func redirectHX(w http.ResponseWriter, r *http.Request, url string) {
	if isHTMX(r) {
		w.Header().Set("HX-Redirect", url)
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

func redirectHXLogin(w http.ResponseWriter, r *http.Request) {
	if isHTMX(r) {
		w.Header().Set("HX-Redirect", utils.URL("/admin"))
	} else {
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
	}
}
