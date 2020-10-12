package main

import (
	"database/sql"
	"fmt"
)

func gettargets() [][]string {
	db2, _ := sql.Open("sqlite3", "./db/scrapper.db")
	row2, err3 := db2.Query(`SELECT * FROM targets ORDER BY subject`)
	if err3 != nil {
		panic(err3)
	}
	defer row2.Close()
	var targetlist [][]string
	for row2.Next() {
		target := []string{"", "", ""}
		row2.Scan(&target[0], &target[1], &target[2])
		targetlist = append(targetlist, target)
		fmt.Println(target)

	}
	return targetlist
}
