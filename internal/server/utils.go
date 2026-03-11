package server

import (
	"cchoice/internal/utils"
	"net/http"
)

func redirectHX(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Redirect", url)
	w.WriteHeader(http.StatusOK)
}

func redirectHXLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
}
