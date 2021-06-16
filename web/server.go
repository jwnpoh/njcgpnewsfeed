package web

import (
	"context"
	"fmt"
	"html/template"
	"log"
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
	TemplateDir string
    Articles *db.ArticlesDBByDate
    Questions db.QuestionsDB
    Ctx context.Context
}

var s Server

// Start takes a Server already initialised with the initial data, parses templates and handlers, and starts ListenAndServe.
func (s *Server) Start() error {
	fmt.Printf(startMsg, s.Port)

	s.parseTemplates()
	s.router()
    err := http.ListenAndServe(":"+s.Port, nil)
	if err != nil {
		return err
	}
	return nil
}

// NewServer initialises the initial data necessary to get going.
func NewServer() *Server {
    ctx := context.Background()
    database := db.NewArticlesDBByDate()
    qnDB, err := db.InitQuestionsDB()
    if err != nil {
        log.Fatal(err)
    }

    if err := database.InitArticlesDB(ctx, qnDB); err != nil {
        log.Fatal(err)
    }

    s.Articles = database
    s.Questions = qnDB
    s.Ctx = ctx
	return &s
}

func (s *Server) parseTemplates() {
	templates := filepath.Join(s.TemplateDir, "*html")
	tpl = template.Must(template.ParseGlob(templates))
}

