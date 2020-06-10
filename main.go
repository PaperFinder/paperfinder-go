package main

import (
	"os"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
)

func main() {
	finder := iris.New()
	finder.Logger().SetLevel("debug")
	finder.Use(recover.New())
	finder.Use(logger.New())
	finder.RegisterView(iris.HTML("./_html-templates", ".html"))
	finder.Handle("GET", "/", func(context iris.Context) {
		//In the future probably we will use a db for the subjects rather than scanning the folder each time.
		file, err := os.Open("./_past-papers")
		if err != nil {
			panic(err)
		}
		list, _ := file.Readdirnames(0)
		context.ViewData("subjects", list)
		context.View("index.html")
	})

	finder.Handle("GET", "/finder", func(context iris.Context) {
		if !context.URLParamExists("subject") ||
			!context.URLParamExists("question") ||
			context.URLParam("question") == "" {

			context.Redirect("/", iris.StatusSeeOther)
		}

		subject := context.URLParam("subject")
		question := context.URLParam("question")
		switch subject {
		case "ph":
			context.WriteString("Physics: ")
		case "bio":
			context.WriteString("Biology: ")
		case "chem":
			context.WriteString("Chemistry: ")
		case "pmath":
			context.WriteString("Pure Maths: ")
		}
		context.WriteString(question)
	})

	finder.Run(iris.Addr(":80"), iris.WithoutServerError(iris.ErrServerClosed))
}
