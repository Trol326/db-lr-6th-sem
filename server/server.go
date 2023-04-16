package server

import (
	"fmt"
	"lr4/database"
	"lr4/server/handler"
	"net/http"
)

func Run() {
	db := database.Connect()
	defer db.Close()

	fmt.Print("\nStarting web server...\n")
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	http.HandleFunc("/", handler.LoginPageHandler)
	http.HandleFunc("/index", handler.Index)
	http.HandleFunc("/logout", handler.LogoutHandler)
	http.HandleFunc("/cityclassifier", handler.CityClassifier)
	http.HandleFunc("/searchstudent", handler.SearchStudent)
	http.HandleFunc("/searchconference", handler.SearchСonference)
	http.HandleFunc("/searchproject", handler.SearchProject)
	http.HandleFunc("/searchpaper", handler.SearchPaper)
	http.HandleFunc("/searchreport", handler.SearchReport)
	http.HandleFunc("/chartbasicbar", handler.ChartBasicBar)

	http.ListenAndServe("localhost:8080", nil)
	fmt.Print("\nShutting down server...\n")
}
