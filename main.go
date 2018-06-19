package main

import (
	"net/http"
	"github.com/labstack/echo"
	"io"
	"html/template"
	"path/filepath"
	"io/ioutil"
	"strings"
	"os"
	"fmt"
	"bufio"
	"strconv"
	"log"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	fmt.Print("ポート番号を指定して下さい: ")
	// Scannerを使って一行読み
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	port :=scanner.Text()
	i, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Error: 正しい数値を入力して下さい %s", err)
	}
	if i < 0 || i > 65535 {
		log.Fatalf("Error: ポートは0～65535の間で指定して下さい")
	}
	e := echo.New()

	// Templates
	funcMap := template.FuncMap{
		"safehtml": func(text string) template.HTML { return template.HTML(text) },
	}

	t := &Template{
		template.Must(template.New("").Funcs(funcMap).ParseFiles("assets/map.html")),
	}
	e.Renderer = t

	// Routing
	e.Static("/maps", "maps")
	e.Static("/assets", "assets")
	e.GET("/:maps", RenderMap)
	e.GET("/", RenderMap)

	// Start server
	e.Start(":"+port)
}

type Data struct {
	Floors		template.JS
	Maplist		string
	Floornav	string
}

func RenderMap(c echo.Context) error {
	reqmap := c.Param("maps")
	var data Data
	var tempFloors string

	// Get map list from "/maps"
	maps := ls("maps")

	if reqmap == "" {
		reqmap = maps[0]
	}

	// Make maplist
	for _, m := range maps {
		if m == reqmap {
			data.Maplist = data.Maplist + "<option selected value=\"" + m + "\">" + m + "</option>"
		} else {
			data.Maplist = data.Maplist + "<option value=\"" + m + "\">" + m + "</option>"
		}
	}

	// Get floors
	prevDir, _ := filepath.Abs(".")
	os.Chdir("maps/"+reqmap)
	floors := dirwalk("./")
	os.Chdir(prevDir)

	// Make floornav and js
	for _, f := range floors {
		data.Floornav = data.Floornav + "<li>" + fmt.Sprintf("%s", getFileNameWithoutExt(f)) + "</li>"
		tempFloors = tempFloors + "\"/maps/" + reqmap + "/" + f + "\","
	}
	data.Floornav = strings.Replace(data.Floornav, "<li>", "<li class=\"active\">", 1)
	data.Floors = template.JS(strings.TrimRight(tempFloors, ","))

	return c.Render(http.StatusOK, "maps", data)
}

func dirwalk(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}

	return paths
}

func ls(path string) []string {
	fileinfos, _ :=ioutil.ReadDir(path)
	var files []string

	for _,fileinfo := range fileinfos {
		files = append(files, fileinfo.Name())
	}
	return files
}

func getFileNameWithoutExt(path string) string {
	// Fixed with a nice method given by mattn-san
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}