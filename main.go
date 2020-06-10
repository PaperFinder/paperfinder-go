package main

import (
	"github.com/kataras/iris/v12"

	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
)

func main() {
	finder := iris.New()
	finder.Logger().SetLevel("info")
	finder.Use(recover.New())
	finder.Use(logger.New())

	finder.Handle("GET", "/", func(context iris.Context) {
		context.ServeFile("./html-templates/index.html", false)
	})

	finder.Handle("GET", "/finder", func(context iris.Context) {
		if !context.URLParamExists("subject") || !context.URLParamExists("question") || context.URLParam("question") == "" {
			context.Redirect("/", iris.StatusSeeOther)
		}
		context.WriteString(context.URLParam("question"))
	})

	finder.Run(iris.Addr(":80"), iris.WithoutServerError(iris.ErrServerClosed))
}
