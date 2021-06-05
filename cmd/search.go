// Package cmd implements logic to handle interaction between the view and the db
package cmd

import (
	"fmt"
	"strings"

	"github.com/jwnpoh/njcgpnewsfeed/db"
)

func Search(term string, database *db.ArticlesDBByDate) (*db.ArticlesDBByDate, error) {
    results, err := db.InitArticlesDBByDate()
    if err != nil {
        return results, fmt.Errorf("unable to initialise new slice to store results: %w", err)
    }

    for _, i := range *database {
        if searchTitle(term, i) || searchTopics(term, i) || searchQuestions(term, i) || searchDate(term, i) {
            *results = append(*results, i)
        }
        // } else {
        // return results, fmt.Errorf("search yielded no results")
        // }
    }
    // for _, j := range *database {
    //     for _, k := range j.Questions {
    //         if strings.Contains(strings.ToLower(k.Wording), "how") {
    //             fmt.Printf("%s\n%s\n%s\n%v\n%v\n\n", j.DisplayDate, j.Title, j.URL, j.Topics, j.Questions)
    //         }
    //     }
    // }

    return results, nil
}

func searchTitle(term string, a db.Article) bool {
    return strings.Contains(strings.ToLower(a.Title), strings.ToLower(term))
}

func searchTopics(term string, a db.Article) bool {
    for _, j := range a.Topics {
        return strings.Contains(strings.ToLower(string(j)), strings.ToLower(term))
    }
    return false
}

func searchQuestions(term string, a db.Article) bool {
    for _, j := range a.Questions {
        return strings.Contains(strings.ToLower(j.Wording), strings.ToLower(term)) 
    }
    return false
}

func searchDate(term string, a db.Article) bool {
    return strings.Contains(strings.ToLower(a.DisplayDate), strings.ToLower(term))
}

