package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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

		fName := strings.ReplaceAll(info.Name(), ".pdf", "")
		newName := dir + "/" + fName + ".filtered"

		data, err := exec.Command("python", "pdftotext.py", path).Output()

		strdata := string(data)

		fullstopReg := regexp.MustCompile(`\.{2,}`)
		markReg := regexp.MustCompile(`\[\d+\]`)
		markRegRound := regexp.MustCompile(`\(\d+\)`)
		doubleSpace := regexp.MustCompile(`\ {2,}`)

		strdata = strings.ReplaceAll(strdata, ".", " ")
		strdata = strings.ReplaceAll(strdata, "_", "")
		strdata = strings.ReplaceAll(strdata, "\n", " ")
		strdata = strings.ReplaceAll(strdata, "\r", " ")

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
