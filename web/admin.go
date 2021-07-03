// Package web contains the server, routing, and handlers logic.
package web

import (
	"bytes"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jwnpoh/njcgpnewsfeed/db"
)

func setCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sessionID",
		Value:    uuid.New().String(),
		MaxAge:   600,
		HttpOnly: true,
	})
}

func checkCookie(w http.ResponseWriter, r *http.Request) bool {
	c, err := r.Cookie("sessionID")
	if err != nil {
		return false
	}

	c.MaxAge = 600
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
		if r.Form.Get("user") == os.Getenv("ADMIN") && r.Form.Get("password") == os.Getenv("PASSWORD") {
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
	r.ParseForm()
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
		if t == "" {
			continue
		}
		t = strings.TrimSpace(t)
		if regex.MatchString(t) {
			year := regexYear.FindString(t)
			number := strings.TrimLeft(regexNumber.FindString(t), "qQ")

			// check if question exists
			key := year + " " + number
			_, ok := s.Questions[key]
			if !ok {
				http.Error(w, fmt.Sprintf("The question for %v Q%v does not exist in the database. Please add the question first then try adding this article again.", year, number), http.StatusNotFound)
				return
			}

			if err := a.SetQuestions(year, number, s.Questions); err != nil {
				http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
			}
		} else {
			a.SetTopics(strings.Title(t))
		}
	}
	a.AddArticleToDB(s.Articles)
	go db.AppendArticle(s.Ctx, a)
	go db.AppendArticleToOld(s.Ctx, a)
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
	r.ParseForm()
	index := r.Form.Get("index")

	s.Articles.RemoveArticle(index)
	go db.BackupArticles(s.Ctx, s.Articles)
}

func addQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()

		year := r.Form.Get("year")
		number := r.Form.Get("number")
		wording := r.Form.Get("wording")

		qn := db.Question{Year: year, Number: number, Wording: wording}
		key := year + " " + number
		s.Questions[key] = qn
		go db.AppendQuestion(s.Ctx, qn)
	}

	err := tpl.ExecuteTemplate(w, "addQuestion.html", nil)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}

func backup(w http.ResponseWriter, r *http.Request) {
	db.BackupArticles(s.Ctx, s.Articles)
	db.BackupQuestions(s.Ctx, s.Questions)
}

func getTitle(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}

	url := string(b)

	title := regexp.MustCompile(`\s?<meta[^p]*property=\"og:title\"\s?content=\"[^\"]*\"\s?\/?>`)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Unable to get response from %s\n", url)
	}
	defer resp.Body.Close()

	b2, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Unable to read response from %s\n", url)
	}

	t := title.Find(b2)

	regexHead := regexp.MustCompile(`\s?<meta[^p]*property=\"og:title\"\s?content=\"`)
	regexTail := regexp.MustCompile(`\"\s?\/?>`)

	head := regexHead.Find(t)
	tail := regexTail.Find(t)

	output := bytes.TrimPrefix(t, head)
	output = bytes.TrimSuffix(output, tail)
	output = bytes.TrimSpace(output)

	titleText := html.UnescapeString(string(output))
	if titleText == "" {
		fmt.Fprint(w, "Could not find title. Please enter manually.")
		return
	}

	fmt.Fprint(w, titleText)
}
