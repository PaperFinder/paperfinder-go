package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	finder := iris.New()
	finder.Logger().SetLevel("debug")
	finder.Use(recover.New())
	finder.Use(logger.New())

	finder.HandleDir("/js", "./web/js")
	finder.HandleDir("/css", "./web/css")

	finder.Handle("GET", "/", func(context iris.Context) {
		context.ServeFile("./web/_html-templates/index.html", false)
	})

	finder.Handle("GET", "/finder", func(context iris.Context) {
		if !context.URLParamExists("s") ||
			!context.URLParamExists("q") ||
			context.URLParam("q") == "" ||
			context.URLParam("s") == "" {

			context.Redirect("/", iris.StatusSeeOther)
		}

		subject := context.URLParam("s")
		question := context.URLParam("q")

		dir := path.Join("_past-papers", subject)

		//matcher := search.New(language.English, search.Loose, search.IgnoreCase)
		results := "not found"

		bquestion := []byte(strings.ToLower(question))
		var papername string
		var qpl string
		var msl string
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			ext := strings.Split(info.Name(), ".")
			extt := ext[len(ext)-1]
			if extt == "pdf" {
				return nil
			}

			bdata, _ := ioutil.ReadFile(path)

			bdata = []byte(strings.ToLower(string(bdata)))

			if bytes.Contains(bdata, bquestion) {
				results = path
				/* This is for another time since it doesn't work
				pat := matcher.Compile(bdata)
				st, end := pat.Index(bquestion)
				if st == -1 || end == -1 {
					return nil
				}
				data := string(bdata)
				results += string(data[st:end])
				*/
				db, err := sql.Open("sqlite3", "./db/papers.db")
				if err != nil {
					panic(err)
				}
				fmt.Println(path)
				row := db.QueryRow(`SELECT papername,qpl,msl FROM paperinfo WHERE filepath = ?`, "../"+path)

				if err := row.Scan(&papername, &qpl, &msl); err != nil {
					papername = "NA"
					qpl = "NA"
					msl = "NA"

				}
			}

			return nil
		})
		if err != nil {
			panic(err)
		}

		if results == "" {
			context.JSON(map[string]string{"Query": question, "Found": "False"})
		}
		context.JSON(map[string]string{"Query": question, "Found": "True", "Paper": strings.ReplaceAll(papername, ".pdf", ""), "QPL": qpl, "MSL": msl})
	})
	finder.Handle("GET", "/subjects", func(context iris.Context) {
		file, _ := os.Open("./_past-papers")

		list, _ := file.Readdirnames(0)
		context.JSON(map[string]string{"Subjects": strings.Join(list, ",")})
	})

	finder.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}
