package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Queries a question through the papers. Query and subjects are strings to be used for the search
// Allow advancedsearch is if the slower more thorough mechanism will be used
func query(question string, subject string, allowadvsearch bool, debug bool) map[string]string {

	dir := path.Join("_past-papers", subject)
	accuracymin := 60
	bquestion := []byte(strings.ToLower(string([]rune(question)))) // Convert query to bytes for performance
	var results string = "not found"
	var papername string = ""
	var qpl string = ""
	var msl string = ""
	var quenum string = ""
	//backupfound := false
	//TODO FIX BUG WHICH ASSIGNS ALL VARS TO NILL
	//backup* are the fallback incase there isn't a perfect match we need to find the closest thing to it
	//backupque is not used yet. It will be introduced along with the paper text extracting
	directfind := false
	var backupbque []byte
	backupacc := 0
	var backuppapername string = ""
	var backupqpl string = ""
	var backupmsl string = ""
	var backupquen string = ""
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

		//If we have gone through all the papers with no luck, lets do it again but thoroughly this time. And lets fine the best match.
		if !found && allowadvsearch {
			tmpBackupfound, tmpbackupacc, tmpBackupbque = advsearch(bdata, bquestion)
		} else if found {
			directfind = true
		}
		if debug {
			fmt.Println("BACKUPACC: " + strconv.Itoa(tmpbackupacc))
			fmt.Println("TMPBACKUPACC: " + strconv.Itoa(tmpbackupacc))
		}
		if found || (tmpBackupfound && tmpbackupacc >= backupacc && tmpbackupacc > accuracymin) {
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

			if tmpBackupfound { //
				backupbque = tmpBackupbque
				backuppapername = papername
				backupqpl = qpl
				backupmsl = msl
				backupacc = tmpbackupacc
				backupquen = findquestion(bdata, tmpBackupbque)
			}
			if found || tmpbackupacc == 100 {
				quenum = findquestion(bdata, bquestion)
				return io.EOF
			}

		}
		return nil
	})
	if err != nil && err != io.EOF {
		panic(err)
	}
	if !directfind && !allowadvsearch {
		return query(question, subject, true, debug)
	}
	//If advanced search was used, we need to state what we found
	if len(question) > 57 {
		question = question[:57] + "..." //Cut off long questions
	}
	if results == "not found" || papername == "NA" { //If there is a db error lets just say that we didn't find it :/
		return map[string]string{"Query": question, "Found": "False"}
		fmt.Println("FAILED QUERY: ", question)
	} else if !directfind {
		return map[string]string{"Query": question, "Found": "Partial", "Paper": strings.ReplaceAll(backuppapername, ".pdf", ""), "QPL": backupqpl, "MSL": backupmsl, "QueN": backupquen}
	} else {
		return map[string]string{"Query": question, "Found": "True", "Paper": strings.ReplaceAll(papername, ".pdf", ""), "QPL": qpl, "MSL": msl, "QueN": quenum}
	}
  
	return map[string]string{"Query": question, "Found": "NA", "NA": strings.ReplaceAll(backuppapername, ".pdf", ""), "QPL": backupqpl, "MSL": backupmsl}

}
func findquestion(bdata []byte, bquestion []byte) string {

	leftind := bytes.Index(bdata, bquestion)
	if leftind < 0 {
		return ""
	}
	questionnum := regexp.MustCompile(`(?im)Total for Question (?P<num>\d+)`)

	answer := questionnum.FindSubmatch(bdata[leftind:])
	if len(answer) > 0 {
		return string(answer[1])
	}
	return ""
}
func advsearch(bdata []byte, bquestion []byte) (bool, int, []byte) {

	if len(bquestion) < 10 { //Being provided with 5 characters, this isn't going to be accurate at all
		return false, 0, bquestion
	}
	nodots := bytes.ReplaceAll(bquestion, []byte("."), []byte(""))
	if bytes.Contains(bdata, nodots) { //Lets attempt to check if full stops stopped us from detecting the paper.
		return true, 100, nodots
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
		return outinFound, outinaccuracy, bquestion[startind:endind]
	}
	/*
		ok you might be wondering what is going on here
		I wonder too as its 3 am but I will try my best to explain his madness
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
	textGap := -((bytes.Index(bdata, bquestion[:leftind]) + leftind - 1) - (bytes.Index(bdata, bquestion[rightind:])))

	bquestionlength := float64(len(bquestion))
	accuracy := int(((bquestionlength - float64(textGap)) / bquestionlength) * 100)
	if accuracy > 90 && accuracy < 100 {
		return true, accuracy, bquestion[rightind:]
	}
	return false, 0, bquestion
}
