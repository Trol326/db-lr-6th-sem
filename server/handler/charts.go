package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"lr4/database"
	"lr4/utils"
	"net/http"
	"strings"
	"sync"
)

var arDC = make(map[int]map[string]int64)

func ChartBasicBar(w http.ResponseWriter, r *http.Request) {
	arDC[0] = make(map[string]int64)
	arDC[1] = make(map[string]int64)
	arDC[2] = make(map[string]int64)
	arDC[3] = make(map[string]int64)
	studentid := ""
	activitytype := ""
	sOut := ""
	if r.Method == "GET" {
		fmt.Fprintf(w, "%v", sOut+" GET ")
		return
	} else {
		if CheckLoginPOST(w, r) == 0 {
			fmt.Fprintf(w, "%v", "0####/")
			return
		}
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			fmt.Println("ParseForm() err: ", err)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			utils.CheckErr(err, "Ошибка запроса POST: ChartBasicBar")
		}
		studentid = r.FormValue("studentid")
		studentid = strings.TrimSpace(studentid)
		if studentid == "" {
			studentid = "%"
		}
		activitytype = r.FormValue("activitytype")
		activitytype = strings.TrimSpace(activitytype)
	}
	switch activitytype {
	case "1":
		BasicBarJSON(studentid, 1)
		pagesJson, err := json.Marshal(arDC[1])
		utils.CheckErr(err, "Cannot encode to JSON ")
		fmt.Fprintf(w, "%s", pagesJson)
	case "2":
		BasicBarJSON(studentid, 2)
		pagesJson, err := json.Marshal(arDC[2])
		utils.CheckErr(err, "Cannot encode to JSON ")
		fmt.Fprintf(w, "%s", pagesJson)
	case "3":
		BasicBarJSON(studentid, 3)
		pagesJson, err := json.Marshal(arDC[3])
		utils.CheckErr(err, "Cannot encode to JSON ")
		fmt.Fprintf(w, "%s", pagesJson)
	default:
		var wg sync.WaitGroup
		for i := 1; i <= 3; i++ {
			wg.Add(1)
			go func(i int) {
				BasicBarJSON(studentid, i)
				wg.Done()
			}(i)
		}
		wg.Wait()
		for k := range arDC[1] {
			arDC[0][k] = arDC[1][k] +
				arDC[2][k] + arDC[3][k]
		}
		pagesJson, err := json.Marshal(arDC[0])
		utils.CheckErr(err, "Cannot encode to JSON ")
		fmt.Fprintf(w, "%s", pagesJson)
	}
}
func BasicBarJSON(studentid string, activitytype int) {
	sSQLPoint := ""
	switch activitytype {
	case 1:
		sSQLPoint = "IFNULL((SELECT SUM(studentconference.point) FROM studentconference " +
			"LEFT JOIN conference ON studentconference.conferenceID=conference.id " +
			"WHERE studentconference.studentID=student.id ),0) "
	case 2:
		sSQLPoint = "IFNULL((SELECT SUM(IFNULL(studentproject.point,0)) FROM studentproject " +
			"LEFT JOIN project ON studentproject.projectID=project.id " +
			"WHERE studentproject.studentID=student.id ),0) "
	case 3:
		sSQLPoint = "IFNULL((SELECT SUM(IFNULL(studentpaper.point,0)) FROM studentpaper " +
			"LEFT JOIN paper ON studentpaper.paperID=paper.id " +
			"WHERE studentpaper.studentID=student.id ),0) "
	}
	stmt, err := database.DB.Prepare("SELECT student.fio, " +
		sSQLPoint + "as std_point " +
		"FROM student " +
		"WHERE student.id LIKE ? " +
		"ORDER BY fio " + ";")
	utils.CheckErr(err, "Не могу подготовить запрос к БД")
	defer stmt.Close()
	rows, err := stmt.Query(studentid)
	utils.CheckErr(err, "Не могу выполнить запрос в БД")
	defer rows.Close()
	var fio sql.NullString
	var std_point sql.NullInt64
	for rows.Next() {
		err := rows.Scan(&fio, &std_point)
		utils.CheckErr(err, "Не могу прочесть запись")
		arDC[activitytype][fio.String] = std_point.Int64
	}
}
