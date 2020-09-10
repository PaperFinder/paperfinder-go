package main
//Copyright © 2020 cents02
import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	_ "github.com/gocolly/colly/v2"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Not enough args. Usage: main.go crawl|manual path|link")
		return
	}
	command := os.Args[1]
	if command == "crawl" {
		c := colly.NewCollector(
			colly.AllowedDomains("www.physicsandmathstutor.com"),
		)
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*www.*",
			RandomDelay: 10 * time.Second,
		})
		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			link := e.Attr("href")

			fmt.Printf("Link found: %q -> %s\n", e.Text, link)
			if strings.Contains(link, ".pdf") && strings.Contains(link, "QP") {
				fmt.Printf("Link found: %q -> %s\n", e.Text, link)

				//Here we are filtering out useless stuff
				tempname := strings.Split(link, "/download/")[1]
				fpath := strings.ReplaceAll(tempname, "/Past-Papers/", "/")
				fname := fpath[strings.LastIndex(fpath, "/"):]
				papername := fname[1:]
				fpath = strings.ReplaceAll(fpath, " ", "-")
				subject := strings.Split(tempname, "/")[0]
				unit := strings.Split(tempname, "/")[4]

				words := strings.Split(fname, " ")
				if words[2] == "(IAL)" {
					fname = words[0][1:3] + words[1] + words[3] + "IAL" + ".pdf"
				} else {
					fname = words[0][1:3] + words[1] + words[2] + ".pdf"
				}
				pathdir := "../_past-papers/" + fpath[:strings.LastIndex(fpath, "/")+1]
				fmt.Println("PATHDIR: " + pathdir)
				fpath = "../_past-papers/" + fpath
				err := os.MkdirAll(pathdir, 0700)
				if err != nil {
					panic(err)
				}
				fmt.Println(pathdir)
				out, err := os.Create(fpath)
				if err != nil {
					panic(err)
				}
				defer out.Close()
				resp, err := http.Get(link)
				if err != nil {
					time.Sleep(5 * time.Second)
					resp, err = http.Get(link)
					if err != nil {
						fmt.Printf("CANT FIND: %q\n", fname)
					}
				}
				defer resp.Body.Close()
				_, err = io.Copy(out, resp.Body)
				if err != nil {
					_, err = io.Copy(out, resp.Body)
					if err != nil {
						fmt.Printf("CANT DOWNLOAD: %q\n", fname)
					}
				}
				fmt.Printf("Downloaded: %q\n", fname)
				install(fpath, papername, "", pathdir, fname, link, subject, unit)
				fmt.Printf("Installed: %q\n", fname)
				return
			} else if strings.Contains(link, os.Args[2]) {

				c.Visit(link)
				return
			}

		})
		c.OnRequest(func(r *colly.Request) {
			fmt.Println("Visiting", r.URL.String())
		})
		c.Visit(os.Args[2])

	} else {
		dir := os.Args[2]
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if path == os.Args[2] {
				return nil
			}
			if !(strings.HasSuffix(info.Name(), ".pdf")) {
				return nil
			}
			fName := strings.ReplaceAll(info.Name(), ".pdf", "")
			install(path, "", fName, dir, "", "", "", "")
			return nil
		})
		if err != nil {
			panic(err)
		}

	}
}

// It converts the pdf to text file and removes the . spam
// path: directory leading to the file,
// fpapername: Full paper name
// fName: filtered filename
// dir: folder directory where the file is (to be improved)
// name: Used by the crawler to indicate a full filename, leave it blank when importing manually
// durl: download url found by the crawler
func install(path string, fpapername, fName string, dir string, name string, durl string, dsubject string, dunit string) {
	newName := ""
	if !(fName == "") {
		newName = dir + strings.ReplaceAll(fName, ".pdf", "") + ".filtered"
	} else {
		newName = dir + strings.ReplaceAll(name, ".pdf", "") + ".filtered"
	}
	fmt.Println("PATH: " + path)
	cmd := exec.Command("python3", "pdftotext.py", path)

	fmt.Println(path)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	//cmd.Start()
	//cmd.Wait()

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
	unit := ""
	subject := ""
	qpl := ""
	msl := ""
	if name == "" {
		fmt.Println("NEWNAME: " + newName)
		fmt.Println("unitrgx: " + string(unitrgx.Find(data)))
		fmt.Println("subjectrgx: " + string(subjectrgx.Find(data)))
		fmt.Println("yearrgx: " + string(yearrgx.Find(data)))

		unit = strings.Split(string(unitrgx.Find(data)), " ")[1]
		subject = strings.Split(string(subjectrgx.Find(data)), " ")[0]
		subject = strings.Split(subject, "\n")[0]
		year := strings.Split(string(yearrgx.Find(data)), "©")[1] //TODO change this to unicode code
		month := "NA"
		switch fmonth := fName[:2]; fmonth {
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
		qpl = "https://pmt.physicsandmathstutor.com/download/" + subject + "/A-level/Past-Papers/Edexcel-IAL/Unit-" + unit + "/" + papername
		msl = strings.ReplaceAll(qpl, "QP", "MS")
	} else {

		qpl = durl
		msl = strings.ReplaceAll(qpl, "QP", "MS")
		subject = dsubject
		unit = dunit
	}

	//To improve in the future

	db, err := sql.Open("sqlite3", "../db/papers.db")
	if err != nil {
		panic(err)
	}
	insertpaper := `INSERT INTO paperinfo(ID, filepath, papername, subject,unit,qpl,msl) VALUES (NULL,?,?,?,?,?,?)`
	statement, err := db.Prepare(insertpaper)

	if err != nil {
		log.Fatalln(err.Error())
	}

	_, err = statement.Exec(newName, fpapername, subject, unit, qpl, msl)

	if err != nil {
		fmt.Println("A PAPER WAS ALREADY FOUND IN THE DB.")
	}
	db.Close()
	strdata = fullstopReg.ReplaceAllLiteralString(strdata, "")
	strdata = markReg.ReplaceAllLiteralString(strdata, "")
	strdata = markRegRound.ReplaceAllLiteralString(strdata, "")
	strdata = doubleSpace.ReplaceAllLiteralString(strdata, " ")
	fmt.Println("NEWNAME: " + newName)
	err = ioutil.WriteFile(newName, []byte(strdata), 0644)

	return
}
