// Package db provides functions and types relevant to the backend database for the article feed.
package db

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Article is a struct representing a single entry in the articlesdb.
type Article struct {
    Title string
    URL string
    Topics []Topic
    Questions []Question
    DisplayDate string
    Date int64
}
 
// SetTopics is a wrapper around an append function to append multiple topics to the Article struct a.
func (a *Article) SetTopics(topics ...string) error {
    for _, j := range topics {
        a.Topics = append(a.Topics, Topic(j))
    }
    return nil
}

func (a  *Article) SetQuestions(year, number int, qnDB *QuestionsDB) error {
    qn := Question{Year:year, Number:number, Wording:qnDB.ListOfQuestions[year][number]}
    a.Questions = append(a.Questions, qn)
    return nil
}

// SetDate is a wrapper around time.Parse to parse a date string to time.Time type, in order to call time.Unix() to return an int64 that makes the article sortable by date.
func (a *Article) SetDate(date string) error {
    cleanDate := strings.ReplaceAll(date, ",", "")
    t, err := time.Parse("Jan 2 2006", cleanDate)
    if err != nil {
        return fmt.Errorf("unable to parse date published - %w", err)
    }

    a.DisplayDate = date
    a.Date = t.Unix()
    return nil
}

// NewArticle returns an Article
func NewArticle() (*Article, error) {
    var a Article

    a.Topics = make([]Topic, 0)
    a.Questions = make([]Question, 0)

    return &a, nil
}

// ArticlesDBByDate is the database of all entries in the articlesdb. Entries are sorted in reverse order of date, with the most recent at index 0.
type ArticlesDBByDate []Article


// Implement sort.Sort interface
func (a ArticlesDBByDate) Len() int {return len(a)}
func (a ArticlesDBByDate) Less(i, j int) bool {return a[i].Date < a[j].Date }
func (a ArticlesDBByDate) Swap(i, j int) {a[i], a[j] = a[j], a[i]}

func InitArticlesDBByDate() (*ArticlesDBByDate) {
    db := make(ArticlesDBByDate, 0, 50)

    return &db
}

func (a *Article) AddArticleToDB(db *ArticlesDBByDate) error {
    *db = append(*db, *a)
    sort.Sort(sort.Reverse(db))

    return nil
}

