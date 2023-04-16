package handler

import (
	"database/sql"
	"fmt"
	"lr4/database"
	"lr4/utils"
	"net/http"
	"strconv"
	"strings"
)

func SearchStudent(w http.ResponseWriter, r *http.Request) {
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
			utils.CheckErr(err, "Ошибка запроса POST: SearchStudent")
		}
		searchstr := r.FormValue("searchstr")
		stmt, err := database.DB.Prepare(`SELECT student.id as studentid, 
		student.fio as studentfio, 
		university.name as universityname, 
		faculty.name as facultyname,
		speciality.name as specialitytyname,
		student.contacts as studentcontacts
		FROM student
		LEFT JOIN ufs ON
		student.ufsID=ufs.id
		LEFT JOIN university ON
		ufs.universityID=university.id
		LEFT JOIN faculty ON ufs.facultyID=faculty.id
		LEFT JOIN speciality ON ufs.specialityID=speciality.id
	WHERE student.fio LIKE ? OR student.contacts LIKE ?;`)
		utils.CheckErr(err, "Не могу подготовить запрос к БД (SearchStudent)")
		defer stmt.Close()
		rows, err := stmt.Query(searchstr+"%", searchstr+"%")
		utils.CheckErr(err, "Не могу выполнить запрос в БД (SearchStudent)")
		defer rows.Close()

		var studentid sql.NullInt64
		var studentfio sql.NullString
		var universityname sql.NullString
		var facultyname sql.NullString
		var specialityname sql.NullString
		var studentcontacts sql.NullString
		for rows.Next() {
			err = rows.Scan(&studentid, &studentfio,
				&universityname, &facultyname, &specialityname,
				&studentcontacts)
			utils.CheckErr(err, "Не могу прочесть запись (SearchStudent)")
			sOut += "<tr onDblClick=\"ChooseStudent(" + strconv.FormatInt(studentid.Int64, 10) + ",'" +
				studentfio.String + "'); return false;\"><td>" +
				studentfio.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + universityname.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + facultyname.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + specialityname.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + studentcontacts.String + "</td></tr>\n"
		}
	}
	sOut = `<table class="table table-striped">
	 <thead>
	 <tr>
	 <th scope="col">ФИО</th>
	 <th scope="col">&nbsp;</th>
	 <th scope="col">ВУЗ</th>
	 <th scope="col">&nbsp;</th>
	 <th scope="col">Факультет</th>
	 <th scope="col">&nbsp;</th>
	 <th scope="col">Специальность</th>
	 <th scope="col">&nbsp;</th>
	 <th scope="col">Контакт</th>
	 </tr>
	 </thead>
	 <tbody>` + sOut + "</tbody></table>\n"
	fmt.Fprintf(w, "%v", sOut)
}

func SearchСonference(w http.ResponseWriter, r *http.Request) {
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
			utils.CheckErr(err, "Ошибка запроса POST: SearchСonference")
		}
		studentid := r.FormValue("studentid")
		studentid = strings.TrimSpace(studentid)
		if studentid == "" {
			studentid = "%"
		}
		confname := r.FormValue("confname")
		confcity := r.FormValue("confcity")
		confcity = strings.TrimSpace(confcity)
		if confcity == "" {
			confcity = "%"
		}
		confdatestart := r.FormValue("confdatestart")
		confdatestart = strings.TrimSpace(confdatestart)
		confdateend := r.FormValue("confdateend")
		confdateend = strings.TrimSpace(confdateend)
		if confdateend == "" {
			confdateend = "9999-99-99"
		}
		stmt, err := database.DB.Prepare(`SELECT conference.name as conferencename, 
		city.name as cityname,
		conference.dateStart as conferenceDateStart, 
		conference.dateEnd as conferenceDateEnd 
		FROM conference 
		LEFT JOIN city ON conference.cityID=city.id
		WHERE EXISTS(SELECT * FROM studentConference 
		WHERE studentConference.conferenceID=conference.id 
			AND studentConference.studentID LIKE ? OR ?='%') 
		AND conference.name LIKE ? AND city.id LIKE ? 
		AND((""=? AND "9999-99-99"=?) OR (IFNULL(conference.dateStart,"") BETWEEN ? AND ?) 
		OR (IFNULL(conference.dateEnd,"") BETWEEN ? AND ?) 
		OR (IFNULL(conference.dateStart,"")<=? AND ?<=IFNULL(conference.dateEnd,"")) 
		OR (IFNULL(conference.dateStart,"")<=? AND ?<=IFNULL(conference.dateEnd,"")));`)
		utils.CheckErr(err, "Не могу подготовить запрос к БД (SearchСonference)")
		defer stmt.Close()
		rows, err := stmt.Query(studentid, studentid,
			confname+"%",
			confcity+"%",
			confdatestart, confdateend,
			confdatestart, confdateend,
			confdatestart, confdateend,
			confdatestart, confdatestart,
			confdateend, confdateend)
		utils.CheckErr(err, "Не могу выполнить запрос в БД (SearchСonference)")
		defer rows.Close()
		var conferencename sql.NullString
		var cityname sql.NullString
		var conferencedate_start sql.NullString
		var conferencedate_end sql.NullString
		for rows.Next() {
			err = rows.Scan(&conferencename, &cityname, &conferencedate_start, &conferencedate_end)
			utils.CheckErr(err, "Не могу прочесть запись (SearchСonference)")
			conferencedate_start.String = utils.DateToRus(conferencedate_start.String)
			conferencedate_end.String = utils.DateToRus(conferencedate_end.String)
			sOut += "<tr><td>" + conferencename.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + cityname.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + conferencedate_start.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + conferencedate_end.String + "</td></tr>\n"
		}
	}
	sOut = `<table class="table table-striped">
		 <thead>
		 <tr>
		 <th scope="col">Название конференции</th>
		 <th scope="col">&nbsp;</th>
		 <th scope="col">Город</th>
		 <th scope="col">&nbsp;</th>
		 <th scope="col">Дата начала</th>
		 <th scope="col">&nbsp;</th>
		 <th scope="col">Дата окончания</th>
		 </tr>
		 </thead>
		 <tbody>` + sOut + " </tbody></table>\n"
	fmt.Fprintf(w, "%v", sOut)
}

func SearchProject(w http.ResponseWriter, r *http.Request) {
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
			utils.CheckErr(err, "Ошибка запроса POST: SearchProject")
		}
		studentid := r.FormValue("studentid")
		studentid = strings.TrimSpace(studentid)
		if studentid == "" {
			studentid = "%"
		}
		projectname := r.FormValue("projectname")
		projectdatestart := r.FormValue("projectdatestart")
		projectdatestart = strings.TrimSpace(projectdatestart)
		projectdateend := r.FormValue("projectdateend")
		projectdateend = strings.TrimSpace(projectdateend)
		if projectdateend == "" {
			projectdateend = "9999-99-99"
		}
		projectfio := r.FormValue("projectfio")
		projectfio = strings.TrimSpace(projectfio)
		if projectfio == "" {
			projectfio = "%"
		}
		projectcity := r.FormValue("projectcity")
		projectcity = strings.TrimSpace(projectcity)
		if projectcity == "" {
			projectcity = "%"
		}
		projectorganization := r.FormValue("projectorganization")
		projectorganization = strings.TrimSpace(projectorganization)
		if projectorganization == "" {
			projectorganization = "%"
		}
		projectcontacts := r.FormValue("projectcontacts")
		projectcontacts = strings.TrimSpace(projectcontacts)
		if projectcontacts == "" {
			projectcontacts = "%"
		}
		projectposition := r.FormValue("projectposition")
		projectposition = strings.TrimSpace(projectposition)
		if projectposition == "" {
			projectposition = "%"
		}
		stmt, err := database.DB.Prepare(`SELECT project.name as projectname, 
		project.dateStart as projectDateStart, 
		project.dateEnd as projectDateEnd, 
		projectmanager.fio as projectmanagerfio, 
		city.name as cityname, 
		organization.name as organizationname, 
		projectmanager.contacts as projectmanagercontacts,
		projectmanager.position as projectmanagerposition
		FROM project 
		LEFT JOIN projectmanager ON project.managerID=projectmanager.id
		LEFT JOIN organization ON projectmanager.organizationID=organization.id
		LEFT JOIN city ON organization.cityID=city.id 
		WHERE EXISTS(SELECT * FROM studentproject WHERE studentproject.projectID=project.id AND studentproject.studentID LIKE ? OR 
		?='%') AND project.name LIKE ? AND ((""=? AND "9999-99-99"=?) OR 
		(IFNULL(project.dateStart,"") BETWEEN ? AND ?) OR 
		(IFNULL(project.dateEnd,"") BETWEEN ? AND ?) OR 
		(IFNULL(project.dateStart,"")<=? AND ?<=IFNULL(project.dateEnd,"")) OR 
		(IFNULL(project.dateStart,"")<=? AND ?<=IFNULL(project.dateEnd,""))) AND 
		projectmanager.fio LIKE ? AND 
		city.id LIKE ? AND 
		organization.name LIKE ? AND 
		projectmanager.contacts LIKE ? AND
		projectmanager.position LIKE ?;`)
		utils.CheckErr(err, "Не могу подготовить запрос к БД (SearchProject)")
		defer stmt.Close()
		fmt.Println("projectdatestart=", projectdatestart, "projectdateend=",
			projectdateend)
		rows, err := stmt.Query(studentid, studentid,
			projectname+"%",
			projectdatestart, projectdateend,
			projectdatestart, projectdateend,
			projectdatestart, projectdateend,
			projectdatestart, projectdatestart,
			projectdateend, projectdateend,
			projectfio+"%",
			projectcity+"%",
			projectorganization+"%",
			projectcontacts+"%",
			projectposition+"%")
		utils.CheckErr(err, "Не могу выполнить запрос в БД (SearchProject)")
		defer rows.Close()
		var projectname2 sql.NullString
		var projectdate_start sql.NullString
		var projectdate_end sql.NullString
		var project_managerfio sql.NullString
		var cityname sql.NullString
		var organizationname sql.NullString
		var project_managercontacts sql.NullString
		var project_managerposition sql.NullString
		for rows.Next() {
			err = rows.Scan(&projectname2,
				&projectdate_start, &projectdate_end,
				&project_managerfio, &cityname, &organizationname,
				&project_managercontacts, &project_managerposition)
			utils.CheckErr(err, "Не могу прочесть запись (SearchProject)")
			projectdate_start.String = utils.DateToRus(projectdate_start.String)
			projectdate_end.String = utils.DateToRus(projectdate_end.String)
			sOut += "<tr><td>" + projectname2.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + projectdate_start.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + projectdate_end.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + project_managerfio.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + cityname.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + organizationname.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + project_managercontacts.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + project_managerposition.String + "</td></tr>\n"
		}
	}
	sOut = "<table class=\"table table-striped\">\n" +
		" <thead>\n" +
		" <tr>\n" +
		" <th scope=\"col\">Название проекта</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Дата начала</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\"> Дата окончания </th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Руководитель</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Город</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Организация</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Контакт</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Должность</th>\n" +
		" </tr>\n" +
		" </thead>\n" +
		" <tbody>\n" +
		sOut +
		" </tbody>" +
		"</table>\n"
	fmt.Fprintf(w, "%v", sOut)
}

func SearchPaper(w http.ResponseWriter, r *http.Request) {
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
			utils.CheckErr(err, "Ошибка запроса POST: SearchPaper")
		}
		studentid := r.FormValue("studentid")
		studentid = strings.TrimSpace(studentid)
		if studentid == "" {
			studentid = "%"
		}
		papername := r.FormValue("papername")
		journalname := r.FormValue("journalname")
		publishingname := r.FormValue("publishingname")
		paperdatestart := r.FormValue("paperdatestart")
		paperdatestart = strings.TrimSpace(paperdatestart)
		paperdateend := r.FormValue("paperdateend")
		paperdateend = strings.TrimSpace(paperdateend)
		if paperdateend == "" {
			paperdateend = "9999-99-99"
		}
		stmt, err := database.DB.Prepare(`SELECT paper.name, 
		sciencejournal.name, 
		sciencejournal.publishing, 
		sciencejournal.date 
		FROM paper
		LEFT JOIN sciencejournal ON paper.journalID=sciencejournal.id
		WHERE EXISTS(SELECT * FROM studentpaper WHERE	studentpaper.paperID=paper.id AND studentpaper.studentID LIKE ? OR ?='%') AND 
		paper.name LIKE ? AND 
		sciencejournal.name LIKE ? AND 
		sciencejournal.publishing LIKE ? AND 
		IFNULL(sciencejournal.date,"") BETWEEN ? AND ? ;`)
		utils.CheckErr(err, "Не могу подготовить запрос к БД (SearchPaper)")
		defer stmt.Close()
		fmt.Println("paperdatestart=", paperdatestart, "paperdateend=",
			paperdateend)
		rows, err := stmt.Query(studentid, studentid,
			papername+"%",
			journalname+"%",
			publishingname+"%",
			paperdatestart, paperdateend)
		utils.CheckErr(err, "Не могу выполнить запрос в БД (SearchPaper)")
		defer rows.Close()
		var papername2 sql.NullString
		var journalname2 sql.NullString
		var publishingname2 sql.NullString
		var scientific_journaldate2 sql.NullString
		for rows.Next() {
			err = rows.Scan(&papername2, &journalname2, &publishingname2,
				&scientific_journaldate2)
			utils.CheckErr(err, "Не могу прочесть запись (SearchPaper)")
			scientific_journaldate2.String =
				utils.DateToRus(scientific_journaldate2.String)
			sOut += "<tr><td>" + papername2.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + journalname2.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + publishingname2.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + scientific_journaldate2.String + "</td></tr>\n"
		}
	}
	sOut = "<table class=\"table table-striped\">\n" +
		" <thead>\n" +
		" <tr>\n" +
		" <th scope=\"col\">Название статьи</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Название журнала</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Название издательства</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Дата публикации</th>\n" +
		" </tr>\n" +
		" </thead>\n" +
		" <tbody>\n" +
		sOut +
		" </tbody>" +
		"</table>\n"
	fmt.Fprintf(w, "%v", sOut)
}
