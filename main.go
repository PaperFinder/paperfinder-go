package main

import (
	"github.com/kataras/iris/v12"

	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
)

func main() {
	finder := iris.New()
	finder.Logger().SetLevel("debug")
	finder.Use(recover.New())
	finder.Use(logger.New())

	finder.Handle("GET", "/", func(context iris.Context) {
		context.ServeFile("./_html-templates/index.html", false)
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
