// Package web contains the server, routing, and handlers logic.
package web

import (
	"fmt"
	"net/http"

	"github.com/jwnpoh/njcgpnewsfeed/db"
)

func (s *Server) router() {
	http.HandleFunc("/", index)
	http.HandleFunc("/latest", latest)
	http.HandleFunc("/all", all)
	http.HandleFunc("/search", search)
	http.HandleFunc("/admin", admin)
	http.HandleFunc("/form", form)
	http.HandleFunc("/delete", delete)
	http.HandleFunc("/edit", edit)
	http.HandleFunc("/editArticle", editArticle)
	http.HandleFunc("/add", addQuestion)
	http.HandleFunc("/backup", backup)
	http.HandleFunc("/getTitle", getTitle)
	http.HandleFunc("/error", errorPage)
}

func index(w http.ResponseWriter, r *http.Request) {
	data := *s.Articles
	data = data[0:12]

	err := tpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		msg := customError{
			ErrMsg:  fmt.Sprintf("%v", err),
			HelpMsg: "",
		}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}
}

func latest(w http.ResponseWriter, r *http.Request) {
	data := *s.Articles
	data = data[0:15]

	err := tpl.ExecuteTemplate(w, "latest.html", data)
	if err != nil {
		msg := customError{
			ErrMsg:  fmt.Sprintf("%v", err),
			HelpMsg: "",
		}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}
}

func all(w http.ResponseWriter, r *http.Request) {
	data := *s.Articles

	err := tpl.ExecuteTemplate(w, "all.html", data)
	if err != nil {
		msg := customError{
			ErrMsg:  fmt.Sprintf("%v", err),
			HelpMsg: "",
		}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	results := db.Search(q.Get("term"), s.Articles)

	if len(q.Get("term")) == 0 || len(*results) == 0 {
		msg := customError{ErrMsg: "Nothing matched the search term.", HelpMsg: "Try refining your search term, or try a different search term."}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}

	data := struct {
		Term     string
		Articles *db.ArticlesDBByDate
	}{
		Term:     q.Get("term"),
		Articles: results,
	}

	err := tpl.ExecuteTemplate(w, "search.html", data)
	if err != nil {
		msg := customError{
			ErrMsg:  fmt.Sprintf("%v", err),
			HelpMsg: "",
		}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}
}

func errorPage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	msg := customError{
		ErrMsg:  q.Get("ErrMsg"),
		HelpMsg: q.Get("HelpMsg"),
	}

	err := tpl.ExecuteTemplate(w, "error.html", msg)
	if err != nil {
		msg := customError{
			ErrMsg:  fmt.Sprintf("%v", err),
			HelpMsg: "",
		}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}
}
