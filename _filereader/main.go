package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
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
		strdata, err := readPdf(path) // Read local pdf file
		if err != nil {
			panic(err)
		}

		fName := strings.ReplaceAll(info.Name(), ".pdf", "")

		fullstopReg := regexp.MustCompile(`\.{2,}`)
		markReg := regexp.MustCompile(`\[\d+\]`)
		markRegRound := regexp.MustCompile(`\(\d+\)`)
		doubleSpace := regexp.MustCompile(`\ {2,}`)

		strdata = fullstopReg.ReplaceAllLiteralString(strdata, "")
		strdata = markReg.ReplaceAllLiteralString(strdata, "")
		strdata = markRegRound.ReplaceAllLiteralString(strdata, "")
		strdata = doubleSpace.ReplaceAllLiteralString(strdata, " ")

		strdata = strings.ReplaceAll(strdata, ".", " ")
		strdata = strings.ReplaceAll(strdata, "_", "")
		strdata = strings.ReplaceAll(strdata, "\r\n", " ")
		strdata = strings.ReplaceAll(strdata, "\n", " ")

		err = ioutil.WriteFile(dir+"/"+fName+".filtered", []byte(strdata), 0644)

		return nil
	})
	if err != nil {
		panic(err)
	}
}

func readPdf(path string) (string, error) {
	f, r, err := pdf.Open(path)
	// remember close file
	defer f.Close()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}
