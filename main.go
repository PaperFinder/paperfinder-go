//Copyright © 2020 cents02
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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
		panic(fmt.Errorf("CONFIG ERROR: %s", err))
	}
	debug := viper.GetBool("Server.Debug")
	host := viper.GetString("Server.Host")
	port := viper.GetInt("Server.Port")
	//accuracymin := viper.GetInt("Search.accuracy_cutoff")
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
		if strings.ContainsAny(subject, "./") { //Incase of a path traversal attempt
			context.JSON(map[string]string{"Error": "500", "Message": "Invalid Subject " + subject})
			return
		}

		if len(context.Request().Cookies()) > 0 { //If cookies exist
			context.SetCookieKV("last_pref", subject, iris.CookieHTTPOnly(false))
		}
		result := query(question, subject, false)
		context.JSON(result)
	})

	finder.Handle("GET", "/subjects", func(context iris.Context) {
		file, _ := os.Open("_past-papers") //Returs a list of all available past papers
		list, _ := file.Readdirnames(0)
		subjects := strings.Join(list, ",")
		if len(context.Request().Cookies()) > 0 { //If cookies exist
			subjpref := context.GetCookie("last_pref")
			if subjpref != list[0] && subjpref != "none" && subjpref != "" {
				list = append(list, list[0])
				list[0] = subjpref
				subjects = strings.Join(list, ",")
				subjects = strings.ReplaceAll(subjects, ","+subjpref, "") //Instead of going through the whole list to find and remove, lets just do this.
			}

		}

		context.JSON(map[string]string{"Subjects": subjects})
	})

	finder.Handle("GET", "/getcookie", func(context iris.Context) {

		context.SetCookieKV("last_pref", "none", iris.CookieExpires(time.Duration(360)*time.Hour), iris.CookieHTTPOnly(false))
	})

	finder.Run(iris.Addr(host+":"+strconv.Itoa(port)), iris.WithoutServerError(iris.ErrServerClosed))
}
