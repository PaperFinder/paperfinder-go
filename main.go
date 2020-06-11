package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/search"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/sessions"
)

var (
	sescookie = "SESCOOKIE"
	sess      = sessions.New(sessions.Config{Cookie: sescookie})
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

		dir := "_past-papers\\" + subject

		matcher := search.New(language.English, search.Loose, search.IgnoreCase)

		var results string

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			ext := strings.Split(info.Name(), ".")[1]
			if ext == "pdf" {
				return nil
			}

			data, _ := ioutil.ReadFile(path)

			data = []byte(strings.ToLower(string(data)))

			question = strings.ToLower(question)

			pat := matcher.Compile(data)
			st, end := pat.IndexString(question)
			if st == -1 || end == -1 {
				return nil
			}

			results += string(data[st:end])

			return nil
		})
		if err != nil {
			panic(err)
		}
		if results == "" {
			context.WriteString("nothing found")
		}
		context.WriteString(results)
	})
	finder.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}
