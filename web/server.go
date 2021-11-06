// Package web contains the server, routing, and handlers logic.
package web

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/jwnpoh/njcgpnewsfeed/db"
	"google.golang.org/api/sheets/v4"
)

const startMsg = `
NJC GP News Feed
Author: Joel Poh
 2021 National Junior College

==> Started server, listening on port %v....
==> `

var tpl *template.Template

// Server represents all objects to be initialised in the application.
type Server struct {
	Port        string
	TemplateDir string
	AssetPath   string
	AssetDir    string
	Articles    *db.ArticlesDBByDate
	Questions   db.QuestionsDB
	Topics      db.TopicsCount
	Ctx         context.Context
	Srv         *sheets.Service
}

var s Server

// Start takes a Server already initialised with the initial data, parses templates and handlers, and starts ListenAndServe. Start is to be called by main in package main.
func (s *Server) Start() error {
	log.Printf(startMsg, s.Port)

	// for testing
	for _, v := range s.Topics {
		fmt.Printf("%s:\t%d\n", string(v.Key), v.Value)
	}

	s.parseTemplates()
	s.serveStatic()
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
	qnDB, err := db.InitQuestionsDB(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tm := db.InitTopicsMap()

	if err := database.InitArticlesDB(ctx, qnDB, tm); err != nil {
		log.Fatal(err)
	}

	tc := db.GetTopicsCount(tm)

	s.Articles = database
	s.Questions = qnDB
	s.Topics = tc
	s.Ctx = ctx
	return &s
}

func (s *Server) parseTemplates() {
	templates := filepath.Join(s.TemplateDir, "*html")
	tpl = template.Must(template.ParseGlob(templates))
}

func (s *Server) serveStatic() {
	http.Handle(s.AssetPath, http.StripPrefix(s.AssetPath, http.FileServer(http.Dir(s.AssetDir))))
}
