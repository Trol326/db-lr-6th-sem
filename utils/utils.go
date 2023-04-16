package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func CheckErr(err error, mes string) {
	if err != nil {
		log.Fatalf("%s\nError:\n%s", mes, err)
	}
}

func ReadTextFile(filename string) string {
	content, err := os.ReadFile(filename)
	CheckErr(err, fmt.Sprintf("Can't read file. Filename = %s", filename))
	return string(content)
}

func HashPassword(password string) []byte {
	passwd := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(passwd, 12)
	CheckErr(err, "Can't hash password")
	return hashedPassword
}

func CompareHashAndPassword(hash []byte, password string) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	return err == nil
}

func GetDBConfig() (host string, name string, password string) {
	text := ReadTextFile(filepath.Join("configs", "dbconfig.txt"))
	data := strings.Split(text, ";")
	host, name, password = data[0], data[1], data[2]
	return
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
