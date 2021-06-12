// Package web implements logic to handle interaction between the view and the db
package web

import (
	"strings"

	"github.com/jwnpoh/njcgpnewsfeed/db"
)

// Search 
func Search(term string, database *db.ArticlesDBByDate) *db.ArticlesDBByDate {
    results := db.InitArticlesDBByDate()

    for _, i := range *database {
        if !searchTitle(term, i) && !searchTopics(term, i) && !searchQuestions(term, i) && !searchDate(term, i) {
            continue
        } else {
            *results = append(*results, i)
        } 
    }

    return results
}

func searchTitle(term string, a db.Article) bool {
    return strings.Contains(strings.ToLower(a.Title), strings.ToLower(term))
}

func searchTopics(term string, a db.Article) bool {
    for _, j := range a.Topics {
        if strings.Contains(strings.ToLower(string(j)), strings.ToLower(term)) {
            return true
        }
    }
    return false
}

func searchQuestions(term string, a db.Article) bool {
    for _, j := range a.Questions {
        if strings.Contains(strings.ToLower(j.Wording), strings.ToLower(term)) {
            return true
        }
    }
    return false
}

func searchDate(term string, a db.Article) bool {
    return strings.Contains(strings.ToLower(a.DisplayDate), strings.ToLower(term))
}

