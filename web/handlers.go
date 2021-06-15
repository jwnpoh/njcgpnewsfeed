package web

import (
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
	http.HandleFunc("/edit", edit)
	http.HandleFunc("/delete", delete)
}

func index(w http.ResponseWriter, r *http.Request) {
    data := *s.Articles
    data = data[0:15]

	err := tpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        return
	}
}

func latest(w http.ResponseWriter, r *http.Request) {
    data := *s.Articles
    data = data[0:15]

	err := tpl.ExecuteTemplate(w, "latest.html", data)
	if err != nil {
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        return
	}
}

func all(w http.ResponseWriter, r *http.Request) {
    data := *s.Articles

	err := tpl.ExecuteTemplate(w, "all.html", data)
	if err != nil {
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        return
	}
}

func search(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    results := Search(q.Get("term"), s.Articles)

    if len(q.Get("term")) == 0 || len(*results) == 0 {
        http.Error(w, "Nothing matched the search term. Try refining your search term.", http.StatusNotFound)
        return
    }

    data := struct{
        Term string
        Articles *db.ArticlesDBByDate
    }{
        Term: q.Get("term"),
        Articles: results ,
    }

	err := tpl.ExecuteTemplate(w, "search.html", data)
	if err != nil {
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        return
	}

}
