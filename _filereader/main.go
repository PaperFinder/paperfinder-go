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

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Not enough args")
		return
	}
	dir := os.Args[1]

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if path == os.Args[1] {
			return nil
		}
		if !(strings.HasSuffix(info.Name(), ".pdf")) {
			return nil
		}
		fName := strings.ReplaceAll(info.Name(), ".pdf", "")
		newName := dir + fName + ".filtered"

		cmd := exec.Command("python3", "pdftotext.py", path)
		fmt.Println(path)
		cmd.Start()
		cmd.Wait()

		data, _ := ioutil.ReadFile(path + ".temp")
		os.Remove(path + ".temp")
		strdata := string(data)

		fullstopReg := regexp.MustCompile(`\.{2,}`)
		markReg := regexp.MustCompile(`\[\d+\]`)
		markRegRound := regexp.MustCompile(`\(\d+\)`)
		doubleSpace := regexp.MustCompile(`\ {2,}`)
		unitrgx := regexp.MustCompile(`(Unit? [1-9]+)`)
		subjectrgx := regexp.MustCompile(`\w+\sAdvanced`) //regexp doesn't like spaces apparently
		yearrgx := regexp.MustCompile(`©\d+`)

		strdata = strings.ReplaceAll(strdata, ".", " ")
		strdata = strings.ReplaceAll(strdata, "_", "")
		strdata = strings.ReplaceAll(strdata, "\n", " ")
		strdata = strings.ReplaceAll(strdata, "\r", " ")

		//To improve in the future
		fmt.Println("NEWNAME: " + newName)

		fmt.Println("unitrgx: " + string(unitrgx.Find(data)))
		fmt.Println("subjectrgx: " + string(subjectrgx.Find(data)))
		fmt.Println("yearrgx: " + string(yearrgx.Find(data)))
		unit := strings.Split(string(unitrgx.Find(data)), " ")[1]
		subject := strings.Split(string(subjectrgx.Find(data)), " ")[0]
		subject = strings.Split(subject, "\n")[0]
		year := strings.Split(string(yearrgx.Find(data)), "©")[1] //TODO change this to unicode code
		month := "NA"
		fmt.Println("CHECK 1: " + subject)
		switch fmonth := strings.Split(newName, "/")[4][:2]; fmonth {
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
		fmt.Println("CHECK 2")
		//Lets assume everything is from pmt for now
		qpl := "https://pmt.physicsandmathstutor.com/download/" + subject + "/A-level/Past-Papers/Edexcel-IAL/Unit-" + unit + "/" + papername
		msl := strings.ReplaceAll(qpl, "QP", "MS")
		uunit, err := strconv.Atoi(unit)
		if err != nil {
			uunit = 0
		}
		db, err := sql.Open("sqlite3", "../db/papers.db")
		if err != nil {
			panic(err)
		}
		insertpaper := `INSERT INTO paperinfo(ID, filename, subject,unit,qpl,msl) VALUES (NULL,?,?,?,?,?)`
		fmt.Println("CHECK 3")
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

		err = ioutil.WriteFile(newName, []byte(strdata), 0644)

		return nil
	})
	if err != nil {
		panic(err)
	}
}
