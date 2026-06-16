package handler

import "net/http"

func ShareHtmlServe(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/share.html")
}
