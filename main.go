package main

import (
	"crypto/sha1"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	re      = regexp.MustCompile(`^[a-z0-9]+$`)
	datadir = flag.String("data", "data", "data directory")
	addr    = flag.String("addr", ":8989", "server address")

	//go:embed public
	local embed.FS
)

func main() {
	flag.Parse()

	sub, _ := fs.Sub(local, "public")

	if err := os.MkdirAll(*datadir, 0700); err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", echo.WrapHandler(http.FileServer(http.FS(sub))))
	e.GET("/static/*", echo.WrapHandler(http.FileServer(http.FS(sub))))
	e.POST("/", func(c echo.Context) error {
		text := c.FormValue("text")
		b := sha1.Sum([]byte(text))
		sum := fmt.Sprintf("%x", b[:10])
		if err := ioutil.WriteFile(filepath.Join(*datadir, sum), []byte(text), 0644); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.Redirect(http.StatusFound, "/"+sum)
	})
	e.GET("/:sha1", func(c echo.Context) error {
		p := c.Param("sha1")
		if !re.MatchString(p) {
			return c.String(http.StatusBadRequest, "bad request")
		}
		b, err := ioutil.ReadFile(filepath.Join(*datadir, p))
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.Blob(http.StatusOK, "text/plain; charset=UTF-8", b)
	})
	e.Logger.Fatal(e.Start(*addr))
}
