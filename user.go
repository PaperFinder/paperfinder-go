package main

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func trylogin(username string, pass string) bool {
	db, err := sql.Open("sqlite3", "./db/users.db")
	if err != nil {
		panic(err)
	}

	var password string = ""
	var level string = ""
	row := db.QueryRow(`SELECT password,level FROM credentials WHERE username = ?`, username)
	if err := row.Scan(&password, &level); err != nil {
		fmt.Println(err)
		username = ""
	}
	db.Close()
	if username == "" {
		return false
	}

	err2 := bcrypt.CompareHashAndPassword([]byte(password), []byte(pass))
	if err2 != nil {
		return false
	}
	return true
}
