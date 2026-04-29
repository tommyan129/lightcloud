package main

import (
	//"fmt"
	"lightcloud/db"
	"lightcloud/handler"
	"net/http"
)

func main() {
	db.DBInit()

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer((http.Dir("static")))))

	http.HandleFunc("/login", handler.Login)
	http.HandleFunc("/register", handler.Register)
	http.HandleFunc("/main", handler.MainPage)
	http.HandleFunc("/files", handler.ListFiles)
	http.HandleFunc("/upload", handler.UploadFiles)
	http.HandleFunc("/download", handler.DownloadFiles)
	http.HandleFunc("/share", handler.GetShareLink)
	http.HandleFunc("/share/create", handler.CreateShareLink)
	http.HandleFunc("/share/download", handler.DownloadShareFiles)
	http.HandleFunc("/delete", handler.DeleteFiles)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusFound)
	})
	http.ListenAndServe(":8080", nil)
}
