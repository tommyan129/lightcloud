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
	http.HandleFunc("/stream", handler.StreamFile)
	http.HandleFunc("/zip/list", handler.ListZipContents)
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
	http.HandleFunc("/me", handler.GetMe)
	http.HandleFunc("/settings", handler.GetSettings)
	http.HandleFunc("/disk", handler.GetDiskInfo)
	http.HandleFunc("/admin", handler.AdminPage)
	http.HandleFunc("/admin/stats", handler.GetAdminStats)
	http.HandleFunc("/admin/users", handler.GetAdminUsers)
	http.HandleFunc("/admin/users/role", handler.UpdateUserRole)
	http.HandleFunc("/admin/users/delete", handler.DeleteAdminUser)
	http.HandleFunc("/admin/files", handler.GetAdminFiles)
	http.HandleFunc("/admin/files/delete", handler.DeleteAdminFile)
	http.HandleFunc("/admin/shares", handler.GetAdminShares)
	http.HandleFunc("/admin/shares/delete", handler.DeleteAdminShare)
	http.HandleFunc("/admin/sessions", handler.GetAdminSessions)
	http.HandleFunc("/admin/sessions/delete", handler.DeleteAdminSession)
	http.HandleFunc("/folders", handler.ListFolders)
	http.HandleFunc("/folders/create", handler.CreateFolder)
	http.HandleFunc("/folders/delete", handler.DeleteFolder)
	http.HandleFunc("/files/move", handler.MoveFiles)

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
