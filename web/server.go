package web

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/jwnpoh/njcgpnewsfeed/db"
)

const startMsg = `
NJC GP News Feed
Author: Joel Poh
 2021 National Junior College

==> Started server, listening on port %v....
==> `

var tpl *template.Template

type Server struct {
	Port        string
	AssetPath   string
	AssetDir    string
	TemplateDir string
    Articles *db.ArticlesDBByDate
    Questions db.QuestionsDB
}

var s Server

func (s *Server) Start() error {
	fmt.Printf(startMsg, s.Port)

	s.parseTemplates()
	// s.serveStatic()
	s.router()
    err := http.ListenAndServe(":"+s.Port, nil)
	if err != nil {
		return err
	}
	return nil
}

func NewServer(database *db.ArticlesDBByDate, qnDB db.QuestionsDB) *Server {
    s.Articles = database
    s.Questions = qnDB
	return &s
}

func (s *Server) serveStatic() {
	http.Handle(s.AssetPath, http.StripPrefix(s.AssetPath, http.FileServer(http.Dir(s.AssetDir))))
}

func (s *Server) parseTemplates() {
	templates := filepath.Join(s.TemplateDir, "*html")
	tpl = template.Must(template.ParseGlob(templates))
}

