package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Not enough args")
		return
	}
	dir := os.Args[1]
	fmt.Println(dir)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if path == os.Args[1] {
			return nil
		}

		fName := strings.ReplaceAll(info.Name(), ".pdf", "")
		newName := dir + "/" + fName + ".filtered"

		cmd := exec.Command("python", "pdftotext.py", path)

		cmd.Start()
		cmd.Wait()

		data, _ := ioutil.ReadFile(path + ".temp")
		os.Remove(path + ".temp")
		strdata := string(data)

		fullstopReg := regexp.MustCompile(`\.{2,}`)
		markReg := regexp.MustCompile(`\[\d+\]`)
		markRegRound := regexp.MustCompile(`\(\d+\)`)
		doubleSpace := regexp.MustCompile(`\ {2,}`)
		unitrgx := regexp.MustCompile(`/(Unit: [1-9]+)/i`)
		subjectrgx := regexp.MustCompile(`/( .+ Advanced)/i`)
		yearrgx := regexp.MustCompile(`©\d+`) //Pearson is up to date with copyright so why not use it

		strdata = strings.ReplaceAll(strdata, ".", " ")
		strdata = strings.ReplaceAll(strdata, "_", "")
		strdata = strings.ReplaceAll(strdata, "\n", " ")
		strdata = strings.ReplaceAll(strdata, "\r", " ")

		//To improve in the future
		fmt.Println(string(unitrgx.Find([]byte(strdata))))
		fmt.Println(string(subjectrgx.Find([]byte(strdata))))
		fmt.Println(string(yearrgx.Find([]byte(strdata))))
		unit := strings.Split(string(unitrgx.Find([]byte(strdata))), " ")[1]
		subject := strings.Split(string(subjectrgx.Find([]byte(strdata))), " ")[1]
		year := strings.Split(string(yearrgx.Find([]byte(strdata))), "©")[1]
		month := ""
		switch fmonth := newName[:1]; fmonth {
		case "Ja":
			month = "January"
		case "Ju":
			month = "June"
		case "Oc":
			month = "October"
		default:
			month = "NA"
		}
		papername := month + " " + year + " QP - Unit " + unit + " Edexcel " + subject + " A-level.pdf"
		//Lets assume everything is from pmt for now
		qpl := "https://pmt.physicsandmathstutor.com/download/" + subject + "/A-level/Past-Papers/Edexcel-IAL/Unit-" + unit + "/" + papername
		msl := strings.ReplaceAll(qpl, "QP", "MS")
		uunit, err := strconv.Atoi(unit)
		if err != nil {
			uunit = 0
		}
		db, _ := sql.Open("sqlite3", "../db/sqlite-database.db")
		insertpaper := `INSERT INTO papers(ID, filename, subject,unit,qpl,msl) VALUES (,?,?,?,?,?)`

		statement, err := db.Prepare(insertpaper)
		if err != nil {
			log.Fatalln(err.Error())
		}

		_, err = statement.Exec(newName, subject, uunit, qpl, msl)

		if err != nil {
			log.Fatalln(err.Error())
		}
		db.Close()
		strdata = fullstopReg.ReplaceAllLiteralString(strdata, "")
		strdata = markReg.ReplaceAllLiteralString(strdata, "")
		strdata = markRegRound.ReplaceAllLiteralString(strdata, "")
		strdata = doubleSpace.ReplaceAllLiteralString(strdata, " ")

		db.Close()

		err = ioutil.WriteFile(newName, []byte(strdata), 0644)

		return nil
	})
	if err != nil {
		panic(err)
	}
}
