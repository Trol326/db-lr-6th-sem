package database

import (
	"database/sql"
	"fmt"
	"lr4/utils"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect() *sql.DB {
	var host, user, password = utils.GetDBConfig()
	connData := fmt.Sprintf("%s:%s@tcp(%s:3306)/somedb?charset=utf8", user, password, host)
	fmt.Printf("\nConnecting to DB...\n")
	db, err := sql.Open("mysql", connData)
	utils.CheckErr(err, "Connection error")
	DB = db
	fmt.Printf("Success\n")
	return db
}
