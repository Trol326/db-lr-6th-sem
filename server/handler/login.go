package handler

import (
	"database/sql"
	"fmt"
	"log"
	"lr4/database"
	"lr4/utils"
	"net/http"
	"strings"
	"text/template"

	"github.com/gorilla/sessions"
)

type Page struct {
	Date        string
	Username    string
	Displayname string
}

var encryptionKey = "13OtdSecret"
var LoggedUserSession = sessions.NewCookieStore([]byte(encryptionKey))
var sGDisplayName = ""
var ExSignLogin = 0

var logUserPage = utils.ReadTextFile("public/html/login.html")
var logUserTemplate = template.Must(template.New("").Parse(logUserPage))

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	conditionsMap := map[string]interface{}{}
	session, _ := LoggedUserSession.Get(r, "my-user-session")
	if session != nil {
		conditionsMap["Username"] = session.Values["username"]
	}
	if r.FormValue("Login") != "" && r.FormValue("Username") != "" {
		username := r.FormValue("Username")
		password := r.FormValue("Password")
		var sUP = ""
		sUP, sGDisplayName = GetPassDisplayName(fmt.Sprintf("%v", username))
		hashedPasswordFromDatabase := []byte(sUP)
		if !utils.CompareHashAndPassword(hashedPasswordFromDatabase, password) {
			log.Println("Either username or password is wrong")
			conditionsMap["LoginError"] = true
		} else {
			log.Println("Logged in :", username)
			conditionsMap["Username"] = username
			conditionsMap["LoginError"] = false
			session, _ := LoggedUserSession.New(r, "my-user-session")
			session.Values["username"] = username
			ExSignLogin = 1
			err := session.Save(r, w)
			if err != nil {
				log.Println(err)
			}
			http.Redirect(w, r, "/index", http.StatusFound)
		}
	}
	if err := logUserTemplate.Execute(w, conditionsMap); err != nil {
		log.Println(err)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := LoggedUserSession.Get(r, "my-user-session")
	session.Values["username"] = ""
	err := session.Save(r, w)
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func CheckLoginGET(w http.ResponseWriter, r *http.Request) {
	session, err := LoggedUserSession.Get(r, "my-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}
	if session.Values["username"] == "" || ExSignLogin == 0 {
		http.Redirect(w, r, "/logout", http.StatusFound)
	}
}

func CheckLoginPOST(w http.ResponseWriter, r *http.Request) int {
	session, err := LoggedUserSession.Get(r, "my-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}
	if session.Values["username"] == "" || ExSignLogin == 0 {
		return 0
	}
	return 1
}

func GetPassDisplayName(username string) (string, string) {
	db := database.DB
	if db == nil {
		fmt.Print("\nError. No db connection.\n")
		db := database.Connect()
		defer db.Close()
	}
	res, err := db.Query("SELECT pass, displayname FROM users WHERE username =?;", username)
	utils.CheckErr(err, "Не могу выполнить запрос в БД")
	defer res.Close()
	var userPass sql.NullString
	var userDisplayName sql.NullString
	for res.Next() {
		err = res.Scan(&userPass, &userDisplayName)
		utils.CheckErr(err, "Не могу прочесть запись")
		break
	}
	return userPass.String, userDisplayName.String
}

func DateToRus(date string) string {
	date = strings.TrimSpace(date)
	if date == "" {
		return date
	}
	ar := strings.Split(date, "-")
	year, month, day := "", "", ""
	year, month, day = ar[0], ar[1], ar[2]
	sAux := day + "." + month + "." + year
	return sAux
}
