package handler

import (
	"database/sql"
	"fmt"
	"lr4/database"
	"lr4/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nguyenthenguyen/docx"
	"github.com/xuri/excelize/v2"
)

func SearchReport(w http.ResponseWriter, r *http.Request) {
	format := ""
	studentid := ""
	confname := ""
	projectname := ""
	papername := ""
	if r.Method == "GET" {
		CheckLoginGET(w, r)
		query := r.URL.Query()
		format = query.Get("format")
		format = strings.TrimSpace(format)
		studentid = query.Get("studentid")
		studentid = strings.TrimSpace(studentid)
		if studentid == "" {
			studentid = "%"
		}
		confname = query.Get("confname")
		confname = strings.TrimSpace(confname)
		if confname == "" {
			confname = "%"
		}
		projectname = query.Get("projectname")
		projectname = strings.TrimSpace(projectname)
		if projectname == "" {
			projectname = "%"
		}
		papername = query.Get("papername")
		papername = strings.TrimSpace(papername)
		if papername == "" {
			papername = "%"
		}
	} else {
		if CheckLoginPOST(w, r) == 0 {
			fmt.Fprintf(w, "%v", "0####/")
			return
		}
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			fmt.Println("ParseForm() err: ", err)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			utils.CheckErr(err, "Ошибка запроса POST: SearchReport")
		}
		format = r.FormValue("format")
		format = strings.TrimSpace(format)
		studentid = r.FormValue("studentid")
		studentid = strings.TrimSpace(studentid)
		if studentid == "" {
			studentid = "%"
		}
		confname = r.FormValue("confname")
		confname = strings.TrimSpace(confname)
		if confname == "" {
			confname = "%"
		}
		projectname = r.FormValue("projectname")
		projectname = strings.TrimSpace(projectname)
		if projectname == "" {
			projectname = "%"
		}
		papername = r.FormValue("papername")
		papername = strings.TrimSpace(papername)
		if papername == "" {
			papername = "%"
		}
	}
	stmt, err := database.DB.Prepare(`SELECT student.fio,
	IFNULL((SELECT SUM(studentconference.point) FROM studentconference 
	LEFT JOIN conference ON studentconference.conferenceID=conference.id 
	WHERE studentconference.studentID=student.id AND conference.name LIKE ? ),0) as conferencePoint, 
	IFNULL((SELECT SUM(IFNULL(studentproject.point,0)) FROM studentproject 
	LEFT JOIN project ON studentproject.projectID=project.id 
	WHERE studentproject.studentID=student.id AND project.name LIKE ? ),0) as projectPoint, 
	IFNULL((SELECT SUM(IFNULL(studentpaper.point,0)) FROM studentpaper 
	LEFT JOIN paper ON studentpaper.paperID=paper.id 
	WHERE studentpaper.studentID=student.id AND paper.name LIKE ? ),0) as paperPoint 
	FROM student 
	WHERE student.id LIKE ? 
	ORDER BY fio ;`)
	utils.CheckErr(err, "Не могу подготовить запрос к БД activity")
	defer stmt.Close()
	rows, err := stmt.Query(
		confname+"%",
		projectname+"%",
		papername+"%",
		studentid)
	utils.CheckErr(err, "Не могу выполнить запрос в БД activity")
	defer rows.Close()
	switch format {
	case "HTML":
		ReportHTML(rows, w)
	case "XLSX":
		ReportXLSX(rows, w)
	case "DOCX":
		ReportDOCX(rows, w)
	default:
		ReportHTML(rows, w)
	}
}
func ReportHTML(rows *sql.Rows, w http.ResponseWriter) {
	sOut := ""
	nItogoConf := 0
	nItogoProject := 0
	nItogoPaper := 0
	nItogoStudent := 0
	var fio sql.NullString
	var conference_point sql.NullString
	var project_point sql.NullString
	var paper_point sql.NullString
	for rows.Next() {
		err := rows.Scan(&fio, &conference_point,
			&project_point, &paper_point)
		utils.CheckErr(err, "Не могу прочесть запись")
		nItogoStudent = 0
		if n, err := strconv.Atoi(conference_point.String); err == nil {
			nItogoConf += n
			nItogoStudent += n
		}
		if n, err := strconv.Atoi(project_point.String); err == nil {
			nItogoProject += n
			nItogoStudent += n
		}
		if n, err := strconv.Atoi(paper_point.String); err == nil {
			nItogoPaper += n
			nItogoStudent += n
		}
		sOut += "<tr><td>" + fio.String + "</td><td>&nbsp;&nbsp;</td>" +
			"<td>" + conference_point.String + "</td><td>&nbsp;&nbsp;</td>" +
			"<td>" + project_point.String + "</td><td>&nbsp;&nbsp;</td>" +
			"<td>" + paper_point.String + "</td><td>&nbsp;&nbsp;</td>" +
			"<td>" + strconv.Itoa(nItogoStudent) + "</td></tr>\n"
	}
	sOut += "<tr><td><b>Итого:</b></td><td>&nbsp;&nbsp;</td>" +
		"<td>" + strconv.Itoa(nItogoConf) + "</td><td>&nbsp;&nbsp;</td>" +
		"<td>" + strconv.Itoa(nItogoProject) + "</td><td>&nbsp;&nbsp;</td>" +
		"<td>" + strconv.Itoa(nItogoPaper) + "</td><td>&nbsp;&nbsp;</td>" +
		"<td>" + strconv.Itoa(nItogoConf+nItogoProject+nItogoPaper) + "</td></tr>\n"
	sOut = "<table class=\"table table-striped\">\n" +
		" <thead>\n" +
		" <tr>\n" +
		" <th scope=\"col\">ФИО</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Конференции</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Проекты</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Статьи</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Все</th>\n" +
		" </tr>\n" +
		" </thead>\n" +
		" <tbody>\n" +
		sOut +
		" </tbody>" +
		"</table>\n"
	fmt.Fprintf(w, "%v", sOut)
}

func ReportXLSX(rows *sql.Rows, w http.ResponseWriter) {
	nItogoConf := 0
	nItogoProject := 0
	nItogoPaper := 0
	nItogoStudent := 0
	var fio sql.NullString
	var conference_point sql.NullString
	var project_point sql.NullString
	var paper_point sql.NullString
	nLeftCol := 1
	nUpperRow := 1
	sNameCol := ""
	dCoeffX := 1.018
	dCoeffY := 1.0
	xlsx := excelize.NewFile()
	sNameSheet := "Sheet1"
	index, err := xlsx.NewSheet(sNameSheet)
	utils.CheckErr(err, "Can't create new sheet")
	orientation := "landscape"
	err = xlsx.SetPageLayout(
		sNameSheet,
		&excelize.PageLayoutOptions{Orientation: &orientation},
	)
	if err != nil {
		utils.CheckErr(err, "can't set pagemargins")
	}
	bottom, footer, header, left, right, top := 0.2, 0.2, 0.2, 0.2, 0.2, 0.2
	err = xlsx.SetPageMargins(
		sNameSheet,
		&excelize.PageLayoutMarginsOptions{
			Bottom: &bottom,
			Footer: &footer,
			Header: &header,
			Left:   &left,
			Right:  &right,
			Top:    &top,
		},
	)
	if err != nil {
		utils.CheckErr(err, "can't set pagemargins")
	}

	var vBorder = []excelize.Border{
		{Type: "left", Color: "000000", Style: 1},
		{Type: "top", Color: "000000", Style: 1},
		{Type: "bottom", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1},
	}
	styleTNR12, err := xlsx.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Times New Roman",
			Size:   12,
			Bold:   false,
		},
		Alignment: &excelize.Alignment{
			WrapText:   true,
			Vertical:   "center",
			Horizontal: "center",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
		Border: vBorder,
	})
	utils.CheckErr(err, "Can't create style TNR12")

	styleTNR12L, err := xlsx.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Times New Roman",
			Size:   12,
			Bold:   false,
		},
		Alignment: &excelize.Alignment{
			WrapText:   true,
			Vertical:   "center",
			Horizontal: "left",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
		Border: vBorder,
	})
	utils.CheckErr(err, "Can't create style TNR12L")

	styleTNR12B, err := xlsx.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Times New Roman",
			Size:   12,
			Bold:   true,
		},
		Alignment: &excelize.Alignment{
			WrapText:   true,
			Vertical:   "center",
			Horizontal: "center",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
		Border: vBorder,
	})
	utils.CheckErr(err, "Can't create style TNR12B")

	styleTNR12BL, err := xlsx.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Family: "Times New Roman",
			Size:   12,
			Bold:   true,
		},
		Alignment: &excelize.Alignment{
			WrapText:   true,
			Vertical:   "center",
			Horizontal: "left",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
		Border: vBorder,
	})
	utils.CheckErr(err, "Can't create style TNR12BL")
	nMeter := 0
	xlsx.SetRowHeight(sNameSheet, nMeter, 30*dCoeffY)
	xlsx.MergeCell(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter), ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter))
	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter), ExcelXY2Name(nLeftCol, nUpperRow+nMeter), styleTNR12B)
	utils.CheckErr(err, "Can't set style")

	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter), "Отчет")
	nMeter++
	xlsx.SetRowHeight(sNameSheet, nMeter, 25*dCoeffY)
	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter), ExcelXY2Name(nLeftCol, nUpperRow+nMeter), styleTNR12B)
	utils.CheckErr(err, "Can't set style")

	sNameCol = ExcelX2Name(nLeftCol)
	xlsx.SetColWidth(sNameSheet, sNameCol, sNameCol, 35*dCoeffX)
	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter), "ФИО")
	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter), ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter), styleTNR12B)
	utils.CheckErr(err, "Can't set style")

	sNameCol = ExcelX2Name(nLeftCol + 1)
	xlsx.SetColWidth(sNameSheet, sNameCol, sNameCol, 20.5*dCoeffX)
	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter), "Конференции")
	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter), ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter), styleTNR12B)
	utils.CheckErr(err, "Can't set style")

	sNameCol = ExcelX2Name(nLeftCol + 2)
	xlsx.SetColWidth(sNameSheet, sNameCol, sNameCol, 20.5*dCoeffX)
	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter), "Проекты")
	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter), ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter), styleTNR12B)
	utils.CheckErr(err, "Can't set style")

	sNameCol = ExcelX2Name(nLeftCol + 3)
	xlsx.SetColWidth(sNameSheet, sNameCol, sNameCol, 20.5*dCoeffX)
	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter), "Статьи")
	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter), ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter), styleTNR12B)
	utils.CheckErr(err, "Can't set style")

	sNameCol = ExcelX2Name(nLeftCol + 4)
	xlsx.SetColWidth(sNameSheet, sNameCol, sNameCol, 20.5*dCoeffX)
	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter), "Все")

	for rows.Next() {
		nMeter++
		err := rows.Scan(&fio, &conference_point, &project_point, &paper_point)
		utils.CheckErr(err, "Не могу прочесть запись")
		err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter), ExcelXY2Name(nLeftCol, nUpperRow+nMeter), styleTNR12L)
		utils.CheckErr(err, "Can't set style")

		err = xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter), fio.String)
		utils.CheckErr(err, "Can't set value")

		nItogoStudent = 0
		if n, err := strconv.Atoi(conference_point.String); err == nil {
			err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter), ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter), styleTNR12)
			utils.CheckErr(err, "Can't set style")

			xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter), n)
			utils.CheckErr(err, "Can't set value")

			nItogoConf += n
			nItogoStudent += n
		}
		if n, err := strconv.Atoi(project_point.String); err == nil {
			err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter), ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter), styleTNR12)
			utils.CheckErr(err, "Can't set style")

			xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter), n)
			nItogoProject += n
			nItogoStudent += n
		}
		if n, err := strconv.Atoi(paper_point.String); err == nil {
			err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter), ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter), styleTNR12)
			utils.CheckErr(err, "Can't set style")

			xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter), n)
			nItogoPaper += n
			nItogoStudent += n
		}
		err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter), ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter), styleTNR12)
		utils.CheckErr(err, "Can't set style")

		xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter), nItogoStudent)
	}
	nMeter++
	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter), ExcelXY2Name(nLeftCol, nUpperRow+nMeter), styleTNR12BL)
	utils.CheckErr(err, "Can't set style")

	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter), "Итого:")

	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter), ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter), styleTNR12B)
	utils.CheckErr(err, "Can't set style")

	xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter), nItogoConf)

	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter), ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter), styleTNR12B)
	utils.CheckErr(err, "Can't set style")

	xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter), nItogoProject)

	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter), ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter), styleTNR12B)
	utils.CheckErr(err, "Can't set style")

	xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter), nItogoPaper)

	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter), ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter), styleTNR12B)
	utils.CheckErr(err, "Can't set style")

	xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter), nItogoConf+nItogoProject+nItogoPaper)

	xlsx.SetActiveSheet(index)
	file := xlsx
	sDate := time.Now().Format("2006-01-02")
	sDate = utils.DateToRus(sDate)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition",
		"attachment;filename="+sDate+"_Отчет.xlsx")
	w.Header().Set("File-Name", sDate+"_Отчет.xlsx")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	err = file.Write(w)
	if err != nil {
		fmt.Println(err)
	}
}

func ReportDOCX(rows *sql.Rows, w http.ResponseWriter) {
	nItogoConf := 0
	nItogoProject := 0
	nItogoPaper := 0
	var fio sql.NullString
	var conference_point sql.NullString
	var project_point sql.NullString
	var paper_point sql.NullString
	for rows.Next() {
		err := rows.Scan(&fio, &conference_point,
			&project_point, &paper_point)
		utils.CheckErr(err, "Не могу прочесть запись")
		if n, err := strconv.Atoi(conference_point.String); err == nil {
			nItogoConf += n
		}
		if n, err := strconv.Atoi(project_point.String); err == nil {
			nItogoProject += n
		}
		if n, err := strconv.Atoi(paper_point.String); err == nil {
			nItogoPaper += n
		}
	}
	rWord, err := docx.ReadDocxFile("mpage/report.docx")
	utils.CheckErr(err, "Не могу получить доступ к report.docx")
	docx1 := rWord.Editable()
	docx1.Replace("old_1_1", strconv.Itoa(nItogoConf), -1)
	docx1.Replace("old_1_2", strconv.Itoa(nItogoProject), -1)
	docx1.Replace("old_1_3", strconv.Itoa(nItogoPaper), -1)
	file := docx1
	sDate := time.Now().Format("2006-01-02")
	sDate = utils.DateToRus(sDate)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=WhatsApp"+sDate+".docx")
	w.Header().Set("File-Name", "WhatsApp "+sDate+".docx")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	err = file.Write(w)
	if err != nil {
		fmt.Println(err)
	}
	rWord.Close()
}

func ExcelXY2Name(X int, Y int) string {
	c, err := excelize.CoordinatesToCellName(X, Y)
	utils.CheckErr(err, "Ошибка в ExcelXY2Name")
	return c
}

func ExcelX2Name(X int) string {
	c, err := excelize.ColumnNumberToName(X)
	utils.CheckErr(err, "Ошибка в ExcelX2Name")
	return c
}
