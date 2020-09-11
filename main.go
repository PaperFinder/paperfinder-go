//Copyright © 2020 cents02
package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

func main() {
	//Lets load up the configs before we do anything
	viper.SetConfigName("server_config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("CONFIG ERROR: ", err))
	}
	debug := viper.GetBool("Server.Debug")
	host := viper.GetString("Server.Host")
	port := viper.GetInt("Server.Port")

	finder := iris.New()
	if debug {
		finder.Logger().SetLevel("debug")
	}

	finder.Use(recover.New())
	finder.Use(logger.New())

	finder.HandleDir("/js", "./web/js")
	finder.HandleDir("/css", "./web/css")

	finder.Handle("GET", "/", func(context iris.Context) {
		context.ServeFile("./web/_html-templates/index.html", false)
	})

	finder.Handle("GET", "/finder", func(context iris.Context) {
		log.Println("GET Q:" + context.URLParam("q") + " S: " + context.URLParam("s"))
		if !context.URLParamExists("s") ||
			!context.URLParamExists("q") ||
			context.URLParam("q") == "" ||
			context.URLParam("s") == "" {

			context.Redirect("/", iris.StatusSeeOther)
		}

		subject := context.URLParam("s")
		question := context.URLParam("q")

		dir := path.Join("_past-papers", subject)

		results := "not found"

		bquestion := []byte(strings.ToLower(string([]rune(question)))) // Convert query to bytes for performance

		var papername string
		var qpl string
		var msl string
		// Goes through each paper in the db
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			ext := strings.Split(info.Name(), ".")
			extt := ext[len(ext)-1]

			//check only the modified papers
			if extt == "pdf" {
				return nil
			}

			bdata, _ := ioutil.ReadFile(path)

			bdata = []byte(strings.ToLower(string(bdata)))

			if bytes.Contains(bdata, bquestion) {
				results = path
				//query the db
				db, err := sql.Open("sqlite3", "./db/papers.db")
				if err != nil {
					panic(err)
				}
				if debug {
					fmt.Println(path)
				}
				row := db.QueryRow(`SELECT papername,qpl,msl FROM paperinfo WHERE filepath = ?`, "../"+strings.ReplaceAll(path, "\\", "/"))
				db.Close()
				if err := row.Scan(&papername, &qpl, &msl); err != nil {
					fmt.Println("DB ERROR: ", err, "for paper: ", path)
					papername = "NA" //Incase of db error throw NA value
					qpl = "NA"
					msl = "NA"

				}
			}

			return nil
		})
		if err != nil {
			panic(err)
		}
		if len(question) > 57 {
			question = question[:57] + "..." //Cut off long questions
		}
		if results == "not found" {
			context.JSON(map[string]string{"Query": question, "Found": "False"})
			fmt.Println("FAILED QUERY: ", question)
		} else {
			context.JSON(map[string]string{"Query": question, "Found": "True", "Paper": strings.ReplaceAll(papername, ".pdf", ""), "QPL": qpl, "MSL": msl})
		}
	})

	finder.Handle("GET", "/subjects", func(context iris.Context) {
		file, _ := os.Open("_past-papers") //Returs a list of all available past papers

		list, _ := file.Readdirnames(0)
		context.JSON(map[string]string{"Subjects": strings.Join(list, ",")})
	})

	finder.Run(iris.Addr(host+":"+strconv.Itoa(port)), iris.WithoutServerError(iris.ErrServerClosed))
}
