// Package db provides functions and types relevant to the backend database for the article feed.
package db

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/sheets/v4"
)

// Article is a struct representing a single entry in the articlesdb.
type Article struct {
	Title       string
	URL         string
	Topics      []Topic
	Questions   []Question
	DisplayDate string
	Date        int64
}

// Define topic-related data structures
type Topic string
type TopicsMap map[Topic]Article

// SetTopics is a wrapper around an append function to append multiple topics to the Article struct a.
func (a *Article) SetTopics(topics ...string) error {
	for _, j := range topics {
		a.Topics = append(a.Topics, Topic(j))
	}
	return nil
}

// SetQuestions sets the []Question for the given Article item.
func (a *Article) SetQuestions(year, number string, qnDB QuestionsDB) error {
	if _, err := strconv.Atoi(year); err != nil {
		return fmt.Errorf("the year input is not a number. try again")
	}
	if _, err := strconv.Atoi(number); err != nil {
		return fmt.Errorf("the question number input is not a number. try again")
	}
	key := year + " " + number
	qn := qnDB[key]
	a.Questions = append(a.Questions, qn)
	return nil
}

// SetDate is a wrapper around time.Parse to parse a date string to time.Time type, in order to call time.Unix() to return an int64 that makes the article sortable by date.
func (a *Article) SetDate(date string) error {
	t, err := time.Parse("Jan 2, 2006", date)
	if err != nil {
		return fmt.Errorf("unable to parse date published - %w", err)
	}

	a.DisplayDate = date
	a.Date = t.Unix()
	return nil
}

// NewArticle returns an Article in order to populate fields for adding to the articles database.
func NewArticle() (*Article, error) {
	var a Article

	a.Topics = make([]Topic, 0)
	a.Questions = make([]Question, 0)

	return &a, nil
}

// ArticlesDBByDate is the database of all entries in the articlesdb. Entries are sorted in reverse order of date, with the most recent at index 0.
type ArticlesDBByDate []Article

// Implement sort.Sort interface
func (a ArticlesDBByDate) Len() int           { return len(a) }
func (a ArticlesDBByDate) Less(i, j int) bool { return a[i].Date < a[j].Date }
func (a ArticlesDBByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// NewArticlesDBByDate makes a slice of Articles to initialise the articles database.
func NewArticlesDBByDate() *ArticlesDBByDate {
	db := make(ArticlesDBByDate, 0, 50)

	return &db
}

// AddArticleToDB takes a pointer to an article that has already had all its fields populated and adds it to the articles database, sorting by most recent published date.
func (a *Article) AddArticleToDB(db *ArticlesDBByDate) error {
	*db = append(*db, *a)
	sort.Sort(sort.Reverse(db))

	return nil
}

// RemoveArticle is a function that the admin can invoke from the live app to remove any offending article.
func (db ArticlesDBByDate) RemoveArticle(index string) {
	j, _ := strconv.Atoi(index)
	copy(db[j:], db[j+1:])
	db[len(db)-1] = Article{}
}

// InitArticlesDB initialises the articles database at first run. Data is downloaded from the incumbent Google Sheets and parsed into the app's data structure. This is meant to be executed only once.
func (database *ArticlesDBByDate) InitArticlesDB(ctx context.Context, qnDB QuestionsDB) error {
	srv, err := newSheetsService(ctx)
	if err != nil {
		return fmt.Errorf("unable to start Sheets service: %w", err)
	}

    sheetRange := "Articles"

	data, err := getSheetData(srv, sheetRange)
	if err != nil {
		return fmt.Errorf("unable to get sheet data: %w", err)
	}

	if len(data.Values) == 0 {
		return fmt.Errorf("no data found")
	}

	for _, row := range data.Values {
		a, err := NewArticle()
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		a.Title = fmt.Sprintf("%v", row[1])
		a.URL = fmt.Sprintf("%v", row[2])
		a.SetDate(fmt.Sprintf("%v", row[0]))

        topics := strings.Split(fmt.Sprintf("%v", row[3]), "\n")
		for _, t := range topics {
				a.SetTopics(t)
		}

        if row[4] == "" {
            *database = append(*database, *a)
            continue
        }

        qnRow := strings.Split(fmt.Sprintf("%v", row[3]), "\n")

        for _, qn := range qnRow {
            fields := strings.Split(qn, " ")
            year := fields[0]
            number := fields[1]
            a.SetQuestions(year, number, qnDB)
        }
        *database = append(*database, *a)
	}

    sort.Sort(sort.Reverse(database))

	return nil
}

// BackupArticles backs up the articles database to a predefined, hard-coded Google Sheet.
func BackupArticles(ctx context.Context, database *ArticlesDBByDate) error {
	srv, err := newSheetsService(ctx)
	if err != nil {
		return fmt.Errorf("unable to start Sheets service: %w", err)
	}

	backupSheetID := os.Getenv("SHEET_ID")
	backupSheetName := "Articles"

	var valueRange sheets.ValueRange
	valueRange.Values = make([][]interface{}, 0, len(*database))

	for _, j := range *database {
		sTopics := strings.Builder{}
		for i, k := range j.Topics {
            if i == len(j.Topics) - 1 {
                sTopics.WriteString(string(k))
                break
            }
            sTopics.WriteString(string(k) + "\n")
        }

		sQuestions := strings.Builder{}
		sQuestionsKey := strings.Builder{}
		for i, l := range j.Questions {
            if i == len(j.Questions) - 1 {
                sQuestions.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number) + " " + l.Wording)
                sQuestionsKey.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number))
                break
            }
			sQuestions.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number) + " " + l.Wording + "\n")
            sQuestionsKey.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number) + "\n")
		}

		record := make([]interface{}, 0, 5)
		record = append(record, j.DisplayDate, j.Title, j.URL, sTopics.String(), sQuestionsKey.String(), sQuestions.String())
		valueRange.Values = append(valueRange.Values, record)
	}

	_, err = srv.Spreadsheets.Values.Update(backupSheetID, backupSheetName, &valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to backup data to backup sheet: %w", err)
	}

	return nil
}

// AppendArticle appends a new article added to the web app database to a predefined, hard-coded Google Sheet.
func AppendArticle(ctx context.Context, article *Article) error {
	srv, err := newSheetsService(ctx)
	if err != nil {
		return fmt.Errorf("unable to start Sheets service: %w", err)
	}

	backupSheetID := os.Getenv("SHEET_ID")
	backupSheetName := "Articles"

	var valueRange sheets.ValueRange
	valueRange.Values = make([][]interface{}, 0, 1)

    sTopics := strings.Builder{}
    for i, k := range article.Topics {
        if i == len(article.Topics) - 1 {
            sTopics.WriteString(string(k))
            break
        }
			sTopics.WriteString(string(k) + "\n")
    }

    sQuestions := strings.Builder{}
    sQuestionsKey := strings.Builder{}
    for i, l := range article.Questions {
        if i == len(article.Questions) - 1 {
            sQuestions.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number) + " " + l.Wording)
            sQuestionsKey.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number))
            break
        }
        sQuestions.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number) + " " + l.Wording + "\n")
        sQuestionsKey.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number) + "\n")
    }

    record := make([]interface{}, 0, 6)
    record = append(record, article.DisplayDate, article.Title, article.URL, sTopics.String(), sQuestionsKey.String(), sQuestions.String())
    valueRange.Values = append(valueRange.Values, record)

	_, err = srv.Spreadsheets.Values.Append(backupSheetID, backupSheetName, &valueRange).InsertDataOption("INSERT_ROWS").ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to append article to backup sheet: %w", err)
	}

	return nil
}
