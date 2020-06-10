package main

import (
	"io/ioutil"
	"os"

	"github.com/blevesearch/bleve"
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
		file, err34 := os.Open("./_past-papers/" + subject)
		if err34 != nil {
			panic(err34)
		}
		unlist, err := file.Readdirnames(0)
		if err != nil {
			panic(err)
		}
		mapping := bleve.NewIndexMapping()
		index1, err := bleve.New("./_past-papers/"+subject+"/"+"index.bleve", mapping)
		if err != nil {
			panic(err)
		}
		for _, unit := range unlist {
			file, err4 := os.Open("./_past-papers/" + subject + "/" + unit)
			if err4 != nil {
				panic(err4)
			}
			qplist, err3 := file.Readdirnames(0)
			if err3 != nil {
				panic(err3)
			}
			for _, qp := range qplist {
				paper, err1 := os.Open("./_past-papers/" + subject + "/" + unit + "/" + qp)
				if err1 != nil {
					panic(err1)
				}
				b, err := ioutil.ReadAll(paper)
				if err != nil {
					panic(err)
				}
				index1.Index(paper.Name(), b)
			}

		}
		index1.Close()
		index, err := bleve.Open("./_past-papers/" + subject + "/" + "index.bleve")
		if err != nil {
			panic(err)
		}
		query := bleve.NewMatchQuery(question)
		search := bleve.NewSearchRequest(query)
		searchResults, err := index.Search(search)
		if err != nil {
			panic(err)
		}
		//I have no idea what I'm doing fuck yeah
		context.WriteString(question + "WITH RESULTS OF: " + searchResults.String())
	})

	finder.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}
