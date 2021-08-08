package main

import (
	"log"
	"os"

	"github.com/jwnpoh/njcgpnewsfeed/web"
)

func main() {
	s := web.NewServer()

	s.Port = os.Getenv("PORT")
	if s.Port == "" {
		s.Port = "8080"
		log.Printf("Defaulting to port %s", s.Port)
	}
	s.TemplateDir = "html"
  s.AssetPath = "/assets/"
  s.AssetDir = "assets"

	log.Fatal(s.Start())
}
