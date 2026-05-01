package main

import (
	"lightcloud/db"
	"lightcloud/handler"
	"net/http"
)

const port = ":8080"

func main() {
	db.DBInit()

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer((http.Dir("static")))))

	http.HandleFunc("/init", handler.Init)
	http.HandleFunc("/login", handler.Login)
	http.HandleFunc("/register", handler.Register)
	http.HandleFunc("/main", handler.MainPage)
	http.HandleFunc("/files", handler.ListFiles)
	http.HandleFunc("/upload", handler.UploadFiles)
	http.HandleFunc("/download", handler.DownloadFiles)
	http.HandleFunc("/share/view", handler.ShareHtmlServe)
	http.HandleFunc("/share", handler.ShareInfo)
	http.HandleFunc("/share/create", handler.CreateShareLink)
	http.HandleFunc("/share/download", handler.DownloadShareFiles)
	http.HandleFunc("/delete", handler.DeleteFiles)
	http.HandleFunc("/logout", handler.Logout)
	http.HandleFunc("/users/search", handler.SearchUsers)
	http.HandleFunc("/share/list", handler.GetMyShareLinks)
	http.HandleFunc("/perm/granted", handler.GetGrantedPerms)
	http.HandleFunc("/perm", handler.UpdatePerm)
	http.HandleFunc("/owner", handler.UpdateOwner)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		existAdmin, err := handler.AdminExists()
		if err != nil {
			http.Error(w, "failed to find db res", http.StatusInternalServerError)
			return
		}
		if !existAdmin {
			http.Redirect(w, r, "/init", http.StatusFound)
		} else {
			http.Redirect(w, r, "/login", http.StatusFound)
		}

	})
	http.ListenAndServe(port, nil)
}
