//Copyright © 2020 cents02
package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

func main() {
	//Lets load up the configs before we do anything
	viper.SetConfigName("server_config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("CONFIG ERROR: %s", err))
	}
	debug := viper.GetBool("Server.Debug")
	host := viper.GetString("Server.Host")
	port := viper.GetInt("Server.Port")
	accuracymin := viper.GetInt("Search.accuracy_cutoff")
	finder := iris.New()
	if debug {
		finder.Logger().SetLevel("debug")
	}

	finder.Use(recover.New())
	finder.Use(logger.New())

	finder.HandleDir("/js", "./web/js")
	finder.HandleDir("/css", "./web/css")

	finder.Handle("GET", "/", func(context iris.Context) {
		context.ServeFile("./web/_html-templates/index.html", false)
	})

	finder.Handle("GET", "/finder", func(context iris.Context) {
		log.Println("GET Q:" + context.URLParam("q") + " S: " + context.URLParam("s"))
		if !context.URLParamExists("s") ||
			!context.URLParamExists("q") ||
			context.URLParam("q") == "" ||
			context.URLParam("s") == "" {

			context.Redirect("/", iris.StatusSeeOther)
		}

		subject := context.URLParam("s")
		question := context.URLParam("q")

		dir := path.Join("_past-papers", subject)

		results := "not found"

		bquestion := []byte(strings.ToLower(string([]rune(question)))) // Convert query to bytes for performance

		var papername string
		var qpl string
		var msl string

		backupfound := false

		//backup* are the fallback incase there isn't a perfect match we need to find the closest thing to it
		//backupque is not used yet. It will be introduced along with the paper text extracting
		directfind := false
		var backupbque []byte
		backupacc := 0
		var backuppapername string
		var backupqpl string
		var backupmsl string
		// Goes through each paper in the db
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			ext := strings.Split(info.Name(), ".")
			extt := ext[len(ext)-1]

			//check only the modified papers
			if extt == "pdf" {
				return nil
			}

			bdata, _ := ioutil.ReadFile(path)

			bdata = []byte(strings.ToLower(string(bdata)))
			found := bytes.Contains(bdata, bquestion)
			tmpBackupfound := false
			var tmpBackupbque []byte
			tmpbackupacc := 0
			if !found {
				fmt.Println("BACKUP CHECKING: " + path)
				tmpBackupfound, tmpbackupacc, tmpBackupbque = advsearch(bdata, bquestion)
			} else {
				directfind = true
			}
			fmt.Println("BACKUPACC: " + strconv.Itoa(tmpbackupacc))
			fmt.Println("TMPBACKUPACC: " + strconv.Itoa(tmpbackupacc))
			if found || (tmpBackupfound && tmpbackupacc > backupacc && tmpbackupacc > accuracymin) {
				fmt.Println("SUCCESFUL QUERY IN " + subject)
				results = path

				//query the db
				db, err := sql.Open("sqlite3", "./db/papers.db")
				if err != nil {
					panic(err)
				}
				if debug {
					fmt.Println(path)
				}

				row := db.QueryRow(`SELECT papername,qpl,msl FROM paperinfo WHERE filepath = ?`, "../"+strings.ReplaceAll(path, "\\", "/"))
				db.Close()
				if err := row.Scan(&papername, &qpl, &msl); err != nil {
					fmt.Println("DB ERROR: ", err, "for paper: ", path)
					papername = "NA" //Incase of db error throw NA value
					qpl = "NA"
					msl = "NA"
				}
				if tmpBackupfound {
					backupbque = tmpBackupbque
					backuppapername = papername
					backupqpl = qpl
					backupmsl = msl
					backupacc = tmpbackupacc
					backupfound = true
					if debug {
						fmt.Println(string(backupbque))
					}
				}

			}

			return nil
		})
		if err != nil {
			panic(err)
		}

		//If advanced search was used, we need to state what we found
		if len(question) > 57 {
			question = question[:57] + "..." //Cut off long questions
		}
		if results == "not found" {
			context.JSON(map[string]string{"Query": question, "Found": "False"})
			fmt.Println("FAILED QUERY: ", question)
		} else if backupfound && !directfind {
			context.JSON(map[string]string{"Query": question, "Found": "Partial", "Paper": strings.ReplaceAll(backuppapername, ".pdf", ""), "QPL": backupqpl, "MSL": backupmsl})
		} else {
			context.JSON(map[string]string{"Query": question, "Found": "True", "Paper": strings.ReplaceAll(papername, ".pdf", ""), "QPL": qpl, "MSL": msl})
		}
	})

	finder.Handle("GET", "/subjects", func(context iris.Context) {
		file, _ := os.Open("_past-papers") //Returs a list of all available past papers

		list, _ := file.Readdirnames(0)
		context.JSON(map[string]string{"Subjects": strings.Join(list, ",")})
	})

	finder.Run(iris.Addr(host+":"+strconv.Itoa(port)), iris.WithoutServerError(iris.ErrServerClosed))
}

//This function should be moved to a different package when multithreading will be implemented
func advsearch(bdata []byte, bquestion []byte) (bool, int, []byte) {

	if len(bquestion) < 10 { //Being provided with 5 characters, this isn't going to be accurate at all
		return false, 0, bquestion
	}
	if bytes.Contains(bdata, bytes.ReplaceAll(bquestion, []byte("."), []byte(""))) { //Lets attempt to check if full stops stopped us from detecting the paper.
		return true, 100, bquestion
	}

	startind := 1
	endind := len(bquestion)
	outinaccuracy := 0
	outinFound := false
	for endind-startind > int(len(bquestion)/4) { //Attempt an out to inside search
		outinaccuracy = int((float64(endind-startind) / float64(len(bquestion))) * 100) //holy shit I spent an hour trying to fix this cause I didn't put float64. thx golang
		if bytes.Contains(bdata, bquestion[startind:endind]) {
			outinFound = true
			break
		}
		startind++
		if bytes.Contains(bdata, bquestion[startind:endind]) {
			outinFound = true
			break
		}
		endind--
	}
	if outinaccuracy > 60 {
		return outinFound, outinaccuracy, bquestion
	}
	/*
		ok you might be wondering what is going on here
		I wonder too as its 3 am but I will try my best to explaint his madness
		The code below will be run incase the outin fails, which most likely means that the typo is in the middle of the query
		for that reason, we try to find the most matches from the left and right queary
		e.g
		text: This is normal text!
		query: This is nrmal text!
		process:
				   rightind
					v
		[this is n]o[rmal text!]
				 ^
			  leftind
		More than 90% matches so it passes

		example two:

		text: This is abnormal text!
		query: This is nrmal text!
					rightind
					   v
		[this is ]abno[rmal text!]
				^
			leftind
		querygap: 0
		text gap: 4
		Less than 90% matches so it doesnt pass
	*/

	//Nothing works beyond this point
	leftind := 1
	rightind := 1
	for leftind < int(len(bquestion)) { //Attempt an a middleout search
		if bytes.Contains(bdata, bquestion[:leftind+1]) {
			leftind++
		} else {
			break
		}
	}
	rightind = leftind
	for rightind < int(len(bquestion)) {
		if bytes.Contains(bdata, bquestion[rightind+1:]) {
			rightind++
			break
		} else {
			rightind++
		}
	}
	// Here we find the length of the text inbetween our findings
	textGap := (bytes.Index(bdata, bquestion[:leftind]) + leftind) - (bytes.Index(bdata, bquestion[rightind:]))
	if textGap < 0 {
		textGap = -textGap
	}

	bquestionlength := float64(len(bquestion))
	accuracy := int(((bquestionlength - float64(textGap)) / bquestionlength) * 100)
	if accuracy > 90 {
		return true, accuracy, bquestion
	}
	return false, 0, bquestion
}
