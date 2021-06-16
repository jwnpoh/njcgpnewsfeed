package main

import (
	"log"

	"github.com/jwnpoh/njcgpnewsfeed/web"
)

func main() {
    s := web.NewServer()

    s.Port = "8080"
    s.TemplateDir = "html"

    log.Fatal(s.Start())
}
