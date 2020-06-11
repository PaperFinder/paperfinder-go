package main

import (
	"bytes"
	"io/ioutil"
	"os"

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
	finder.RegisterView(iris.HTML("./_html-templates", ".html"))
	finder.Handle("GET", "/", func(context iris.Context) {
		//In the future probably we will use a db for the subjects rather than scanning the folder each time.
		file, err := os.Open("./_past-papers")
		if err != nil {
			panic(err)
		}
		list, _ := file.Readdirnames(0)
		session := sess.Start(context)
		context.ViewData("subjects", list)
		que := session.Get("que")
		result := session.Get("result")
		if que != nil {
			context.ViewData("que", que)
			if result != nil {
				context.ViewData("result", result)
			} else {
				context.ViewData("result", "no papers")
			}
		}
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
		file, err := os.Open("./_past-papers/" + subject)
		if err != nil {
			panic(err)
		}
		unlist, _ := file.Readdirnames(0)
		//we turn the question in bytes for peak performance
		bquestion := []byte(question)
		for _, unit := range unlist {
			file, _ := os.Open("./_past-papers/" + subject + "/" + unit)

			qplist, _ := file.Readdirnames(0)

			for _, qp := range qplist {
				paper, err := os.Open("./_past-papers/" + subject + "/" + unit + "/" + qp)
				if err != nil {
					panic(err)
				}
				b, err := ioutil.ReadAll(paper)
				if bytes.Contains(b, bquestion) {
					sess.Start(context).set("result", paper.Name())
					sess.Start(context).set("que", question)
					context.Redirect("/")
					return
				}
				if err != nil {
					panic(err)
				}
			}

		}
		sess.Start(context).set("result", nil)
		sess.Start(context).set("que", question)
		context.Redirect("/", iris.StatusSeeOther)
		return

	})

	finder.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}
