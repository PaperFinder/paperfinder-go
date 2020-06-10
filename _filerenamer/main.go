package main

import (
	"fmt"
	"os"
	"path/filepath"
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
		name := info.Name()
		words := strings.Split(name, " ")
		if words[2] == "(IAL)" {
			name = words[0][:2] + words[1] + words[3] + "IAL" + ".pdf"
		} else {
			name = words[0][:2] + words[1] + words[2] + ".pdf"
		}

		name = dir + "/" + name

		err = os.Rename(path, name)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
