package web

import(
	"fmt"
	"regexp"
	"strings"
    "net/http"

    "github.com/google/uuid"
    "github.com/jwnpoh/njcgpnewsfeed/db"

)

func setCookie(w http.ResponseWriter, r *http.Request) {
    http.SetCookie(w, &http.Cookie{
        Name: "sessionID",
        Value: uuid.New().String(), 
        MaxAge: 300,
        HttpOnly: true,
    })
}

func checkCookie(w http.ResponseWriter, r *http.Request) bool {
    c, err := r.Cookie("sessionID")
    if err != nil {
        return false
    }

    c.MaxAge = 300
    return true
}

func admin(w http.ResponseWriter, r *http.Request) {
    if checkCookie(w, r) {
        err := tpl.ExecuteTemplate(w, "dashboard.html", nil)
        if err != nil {
            http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
            return
        }
        return
    }

    if r.Method == "POST" {
        r.ParseForm()
        if r.Form.Get("user") == "admin" && r.Form.Get("password") == "288913" {
                setCookie(w, r)
                err := tpl.ExecuteTemplate(w, "dashboard.html", nil)
                if err != nil {
                    http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
                    return
                }
                return
        } else {
            http.Redirect(w, r, "/admin", http.StatusUnauthorized)
        }
    }

	err := tpl.ExecuteTemplate(w, "admin.html", nil)
	if err != nil {
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        return
	}
}

func form(w http.ResponseWriter, r *http.Request) {
    if !checkCookie(w, r) {
        http.Redirect(w, r, "/admin", http.StatusForbidden)
        return
    }

    if r.Method == "POST" {
        addArticle(w, r)
    }

	err := tpl.ExecuteTemplate(w, "form.html", nil)
	if err != nil {
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        return
	}
}

func addArticle(w http.ResponseWriter, r *http.Request) {
        title := r.Form.Get("title")
        url := r.Form.Get("url")
        date := r.Form.Get("date")
        tags := r.Form.Get("tags")
        xtags := strings.Split(tags, ";")

        a, err := db.NewArticle()
        if err != nil {
            http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
            return
        }

        a.Title = title
        a.URL = url
        if err := a.SetDate(date); err != nil {
            http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
        }
        
        regex := regexp.MustCompile(`^\d{4}\s?-?\s?(q|Q)\d{1,2}$`)
        regexYear := regexp.MustCompile(`^\d{4}`)
        regexNumber := regexp.MustCompile(`(q|Q)\d{1,2}$`)
        for _, t := range xtags {
            t = strings.TrimSpace(t)
            if regex.MatchString(t) {
                year := regexYear.FindString(t)
                number := strings.TrimLeft(regexNumber.FindString(t), "qQ")
                if err := a.SetQuestions(year, number, s.Questions); err != nil {
                    http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
                }
            } else {
                a.SetTopics(strings.Title(t))
            }
        }
        a.AddArticleToDB(s.Articles)
}

func edit(w http.ResponseWriter, r *http.Request) {
    if !checkCookie(w, r) {
        http.Redirect(w, r, "/admin", http.StatusForbidden)
        return
    }

    if r.Method == "POST" {
        editArticle(w, r)
    }

	err := tpl.ExecuteTemplate(w, "edit.html", nil)
	if err != nil {
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        return
	}

}

func editArticle(w http.ResponseWriter, r *http.Request) {

}

func delete(w http.ResponseWriter, r *http.Request) {
    if !checkCookie(w, r) {
        http.Redirect(w, r, "/admin", http.StatusForbidden)
        return
    }

    if r.Method == "POST" {
        deleteArticle(w, r)
    }

    data := *s.Articles
	err := tpl.ExecuteTemplate(w, "delete.html", data)
	if err != nil {
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        return
	}
}

func deleteArticle(w http.ResponseWriter, r *http.Request) {

}
