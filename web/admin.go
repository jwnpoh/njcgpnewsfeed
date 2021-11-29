// Package web contains the server, routing, and handlers logic.
package web

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jwnpoh/njcgpnewsfeed/db"
)

type customError struct {
	ErrMsg  string
	HelpMsg string
}

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
	http.SetCookie(w, c)
	return true
}

func login(w http.ResponseWriter, r *http.Request) {
	var Stats struct {
		TotalArticles   int
		AverageArticles int
		TopQuestions    db.QuestionsByArticleCount
		BottomQuestions db.QuestionsByArticleCount
		TopTopics       db.TopicsCount
		BottomTopics    db.TopicsCount
	}

	// get total number of articles in db.
	Stats.TotalArticles = s.Articles.Len()

	// get average number of articles per day.
	Stats.AverageArticles = getAverageNumberOfArticles(Stats.TotalArticles)

	// get top 5 and bottom 5 questions ranked by number of articles tagged.
	qc := db.RankQuestionsByArticleCount(s.QuestionCounter)
	Stats.TopQuestions = qc[:5]
	Stats.BottomQuestions = qc[len(qc)-5:]

	// get top 5 and bottom 5 topics ranked by number of articles tagged.
	tc := db.GetTopicsCount(s.Topics)
	Stats.TopTopics = tc[:5]
	Stats.BottomTopics = tc[len(tc)-5:]

	err := tpl.ExecuteTemplate(w, "dashboard.html", Stats)
	if err != nil {
		msg := customError{
			ErrMsg:  fmt.Sprintf("%v", err),
			HelpMsg: "",
		}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}
}

func admin(w http.ResponseWriter, r *http.Request) {
	if checkCookie(w, r) {
		login(w, r)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		if r.Form.Get("user") == os.Getenv("ADMIN") && r.Form.Get("password") == os.Getenv("PASSWORD") {
			setCookie(w, r)
			login(w, r)
			return
		}
		http.Redirect(w, r, "/admin", http.StatusUnauthorized)
	}

	err := tpl.ExecuteTemplate(w, "admin.html", nil)
	if err != nil {
		msg := customError{
			ErrMsg:  fmt.Sprintf("%v", err),
			HelpMsg: "",
		}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}
}

func form(w http.ResponseWriter, r *http.Request) {
	if !checkCookie(w, r) {
		http.Redirect(w, r, "/admin", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" {
		addArticle(w, r)
	}

	err := tpl.ExecuteTemplate(w, "form.html", nil)
	if err != nil {
		msg := customError{
			ErrMsg:  fmt.Sprintf("%v", err),
			HelpMsg: "",
		}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}
}

type formData struct {
	title string
	url   string
	date  string
	tags  []string
}

func addArticle(w http.ResponseWriter, r *http.Request) {
	if !checkCookie(w, r) {
		http.Redirect(w, r, "/admin", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	title := r.Form.Get("title")
	url := r.Form.Get("url")
	date := r.Form.Get("date")
	tags := splitTags(r.Form.Get("tags"))

	data := formData{
		title,
		url,
		date,
		tags,
	}

	a, msg, err := formToArticle(data)
	if err != nil {
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}

	a.AddArticleToDB(s.Articles, s.Topics, s.QuestionCounter)
	go db.BackupQuestions(s.Ctx, s.Questions)
	go db.AppendArticle(s.Ctx, a)
	go db.AppendArticleToOld(s.Ctx, a)
}

func delete(w http.ResponseWriter, r *http.Request) {
	if !checkCookie(w, r) {
		http.Redirect(w, r, "/admin", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" {
		deleteArticle(w, r)
	}

	data := *s.Articles
	err := tpl.ExecuteTemplate(w, "delete.html", data)
	if err != nil {
		msg := customError{
			ErrMsg:  fmt.Sprintf("%v", err),
			HelpMsg: "",
		}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}
}

func deleteArticle(w http.ResponseWriter, r *http.Request) {
	if !checkCookie(w, r) {
		http.Redirect(w, r, "/admin", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	index, _ := strconv.Atoi(r.Form.Get("index"))

	s.Articles.RemoveArticle(index, s.Topics, s.QuestionCounter)
	go backup(w, r)
}

func edit(w http.ResponseWriter, r *http.Request) {
	if !checkCookie(w, r) {
		http.Redirect(w, r, "/admin", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		index := r.Form.Get("index")
		http.Redirect(w, r, "/editArticle?index="+index, http.StatusSeeOther)
	}

	data := *s.Articles
	err := tpl.ExecuteTemplate(w, "editList.html", data)
	if err != nil {
		msg := customError{
			ErrMsg:  fmt.Sprintf("%v", err),
			HelpMsg: "",
		}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}
}

func editArticle(w http.ResponseWriter, r *http.Request) {
	if !checkCookie(w, r) {
		http.Redirect(w, r, "/admin", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" {
		editTheArticle(w, r)
		http.Redirect(w, r, "/edit", http.StatusSeeOther)
	}

	r.ParseForm()
	index := r.Form.Get("index")
	i, err := strconv.Atoi(index)
	if err != nil {
		msg := customError{
			ErrMsg:  "Unable to parse index of selected Article to edit.",
			HelpMsg: "Try refreshing the page and try again.",
		}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}

	articles := *s.Articles
	data := struct {
		Index int
		db.Article
	}{
		i,
		articles[i],
	}

	err = tpl.ExecuteTemplate(w, "edit.html", data)
	if err != nil {
		msg := customError{
			ErrMsg:  fmt.Sprintf("%v", err),
			HelpMsg: "",
		}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}
}

func editTheArticle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	index := r.Form.Get("hidden")
	title := r.Form.Get("title")
	url := r.Form.Get("url")
	date := strings.TrimSpace(r.Form.Get("date"))
	tags := splitTags(r.Form.Get("tags"))

	data := formData{
		title,
		url,
		date,
		tags,
	}

	a, msg, err := formToArticle(data)
	if err != nil {
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
		return
	}

	s.Articles.EditArticle(index, *a, s.Topics, s.QuestionCounter)
	go backup(w, r)
}

func addQuestion(w http.ResponseWriter, r *http.Request) {
	if !checkCookie(w, r) {
		http.Redirect(w, r, "/admin", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" {
		if !checkCookie(w, r) {
			http.Redirect(w, r, "/admin", http.StatusUnauthorized)
			return
		}

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
		msg := customError{
			ErrMsg:  fmt.Sprintf("%v", err),
			HelpMsg: "",
		}
		http.Redirect(w, r, "/error?"+fmt.Sprintf("%v=%v&%v=%v", "ErrMsg", msg.ErrMsg, "HelpMsg", msg.HelpMsg), http.StatusSeeOther)
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

func splitTags(tags string) []string {
	tags = strings.TrimSuffix(strings.TrimSpace(tags), ";")
	xtags := strings.Split(tags, ";")

	return xtags
}

func formToArticle(data formData) (*db.Article, customError, error) {
	a, err := db.NewArticle()
	if err != nil {
		msg := customError{ErrMsg: "Unable to initialise new Article.", HelpMsg: "Please try again."}
		return nil, msg, err
	}

	a.Title = data.title
	a.URL = data.url
	if err := a.SetDate(data.date); err != nil {
		msg := customError{ErrMsg: fmt.Sprintf("Unable to parse date %v", data.date), HelpMsg: "Check if the date has been entered correctly"}
		return nil, msg, err
	}

	regex := regexp.MustCompile(`^\d{4}\s?-?\s?(q|Q)\d{1,2}$`)
	regexYear := regexp.MustCompile(`^\d{4}`)
	regexNumber := regexp.MustCompile(`(q|Q)\d{1,2}$`)
	for _, t := range data.tags {
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
				msg := customError{ErrMsg: fmt.Sprintf("The question for %v Q%v does not exist in the database currently.", year, number), HelpMsg: "Please use the Add Question function to add the question to the database first, then try adding the article again."}
				return nil, msg, errors.New("error")
			}

			qnDB, err := a.SetQuestionsNewArticle(year, number, s.Questions)
			if err != nil {
				msg := customError{ErrMsg: fmt.Sprintf("Unable to tag the question %s to the article. Article not created.", key), HelpMsg: "Check if the question tag has been formatted correctly, in the form '2020-Q10'."}
				return nil, msg, err
			}
			s.Questions = qnDB
		} else {
			a.SetTopics(strings.Title(t))
		}
	}
	return a, customError{}, nil
}

func getAverageNumberOfArticles(numOfArticles int) int {
	var average int

	dateOfLaunch, err := time.Parse("Jan 2, 2006", "Jan 15, 2021")
	if err != nil {
		return -1
	}
	timeLive := time.Since(dateOfLaunch)
	daysLive := int(timeLive.Hours()) / 24

	average = numOfArticles / daysLive
	return average
}
