package handler

import (
	"database/sql"
	"fmt"
	"log"
	"lr4/database"
	"lr4/utils"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func Index(w http.ResponseWriter, r *http.Request) {
	conditionsMap := map[string]interface{}{}
	session, err := LoggedUserSession.Get(r, "my-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}
	log.Println("Session name : ", session.Name())
	log.Println("Username : ", session.Values["username"])
	if session.Values["username"] == "" || sGDisplayName == "" {
		http.Redirect(w, r, "/logout", http.StatusFound)
	}
	conditionsMap["Username"] = session.Values["username"]
	if r.URL.Path != "/index" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	year, month, day := time.Now().Date()
	curdate := fmt.Sprintf("%02d.%02d.%d", day, month, year)
	username := fmt.Sprintf("%v", session.Values["username"])
	p := &Page{
		Date:        curdate,
		Username:    username,
		Displayname: sGDisplayName,
	}
	t := template.Must(template.ParseFiles("public/html/index.html"))
	t.Execute(w, p)
}

func CityClassifier(w http.ResponseWriter, r *http.Request) {
	sOut := ""
	if r.Method == "GET" {
		fmt.Fprintf(w, "%v", "")
		return
	} else {
		if CheckLoginPOST(w, r) == 0 {
			fmt.Fprintf(w, "%v", "")
			return
		}
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			fmt.Println("ParseForm() err: ", err)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			utils.CheckErr(err, "Ошибка запроса POST: CityClassifier")
		}
		numvariety := r.FormValue("numvariety")
		numvariety = strings.TrimSpace(numvariety)
		sNameField := ""
		if numvariety == "1" {
			sNameField = "name"
		}
		sOut = "<option value=\"\"></option>"
		stmt, err := database.DB.Prepare("SELECT id, " + sNameField + " " +
			"FROM city " + ";")
		utils.CheckErr(err, "Не могу подготовить запрос к БД")
		defer stmt.Close()
		rows, err := stmt.Query()
		utils.CheckErr(err, "Не могу выполнить запрос в БД")
		defer rows.Close()
		var id sql.NullInt64
		var name sql.NullString
		for rows.Next() {
			err = rows.Scan(&id, &name)
			utils.CheckErr(err, "Не могу прочесть запись")
			sId := strconv.FormatInt(id.Int64, 10)
			sOut += "<option value=\"" + sId + "\">" + name.String + "</option>\n"
		}
	}
	fmt.Fprintf(w, "%v", sOut)
}
