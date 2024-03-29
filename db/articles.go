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

// SetTopics is a wrapper around an append function to append multiple topics to the Article struct a.
func (a *Article) SetTopics(topics ...string) {
	for _, j := range topics {
		a.Topics = append(a.Topics, Topic(j))
	}
}

// SetQuestionsNewArticle sets the []Question and increases the count for the question listing in the qnDB for the given Article item.
func (a *Article) SetQuestionsNewArticle(year, number string, qnDB QuestionsDB) (QuestionsDB, error) {
	if _, err := strconv.Atoi(year); err != nil {
		return qnDB, fmt.Errorf("the year input is not a number. try again")
	}
	if _, err := strconv.Atoi(number); err != nil {
		return qnDB, fmt.Errorf("the question number input is not a number. try again")
	}
	key := year + " " + number
	qn := qnDB[key]
	a.Questions = append(a.Questions, qn)
	qnDB[key] = qn
	return qnDB, nil
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

// AddArticleToDB takes a pointer to an article that has already had all its fields populated and adds it to the articles database, sorting by most recent published date.
func (a *Article) AddArticleToDB(db *ArticlesDBByDate, tm TopicsMap, qc QuestionCounter) error {
	for _, v := range a.Topics {
		tm.Increment(v)
	}

	for _, v := range a.Questions {
		qc.Increment(v.Year + " - Q" + v.Number)
	}

	*db = append(*db, *a)
	sort.Sort(sort.Reverse(db))

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
func (db ArticlesDBByDate) Len() int           { return len(db) }
func (db ArticlesDBByDate) Less(i, j int) bool { return db[i].Date < db[j].Date }
func (db ArticlesDBByDate) Swap(i, j int)      { db[i], db[j] = db[j], db[i] }

// NewArticlesDBByDate makes a slice of Articles to initialise the articles database.
func NewArticlesDBByDate() *ArticlesDBByDate {
	db := make(ArticlesDBByDate, 0, 100)

	return &db
}

// EditArticle is a function that the admin can invoke from the live app to edit a specific article.
func (db ArticlesDBByDate) EditArticle(index string, article Article, tm TopicsMap, qc QuestionCounter) error {
	i, err := strconv.Atoi(index)
	if err != nil {
		return fmt.Errorf("unable to parse index of article: %w", err)
	}

	for _, v := range db[i].Topics {
		tm.Decrement(v)
	}

	for _, v := range db[i].Questions {
		qc.Decrement(v.Year + " - Q" + v.Number)
	}

	for _, v := range article.Topics {
		tm.Increment(v)
	}

	for _, v := range article.Questions {
		qc.Increment(v.Year + " - Q" + v.Number)
	}

	db[i] = article
	sort.Sort(sort.Reverse(db))
	return nil
}

// RemoveArticle is a function that the admin can invoke from the live app to remove any offending article.
func (db *ArticlesDBByDate) RemoveArticle(index int, tm TopicsMap, qc QuestionCounter) {
	d := *db

	for _, v := range d[index].Topics {
		tm.Decrement(v)
	}

	for _, v := range d[index].Questions {
		qc.Decrement(v.Year + " - Q" + v.Number)
	}

	copy(d[index:], d[index+1:])
	d[len(d)-1] = Article{}
	*db = d
}

// InitArticlesDB initialises the articles database at first run. Data is downloaded from the incumbent Google Sheets and parsed into the app's data structure. This is meant to be executed only once.
func (db *ArticlesDBByDate) InitArticlesDB(ctx context.Context, qnDB QuestionsDB, tm TopicsMap, qc QuestionCounter) error {
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

	for i, row := range data.Values {
		i++
		if len(row) < 1 {
			err := SendMail("admin@njcgpnewsfeed", "joel_poh_weinan@moe.edu.sg", "jwn.poh@gmail.com", "NJC GP Newsfeed DB Error", fmt.Sprintf("Check Articles DB at row %d", i))
			if err != nil {
				fmt.Printf("Did not send mail - %v", err)
			} else {
				fmt.Println("Sent mail.")
			}
			continue
		}

		a, err := NewArticle()
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		a.Title = fmt.Sprintf("%v", row[0])
		a.URL = fmt.Sprintf("%v", row[1])
		if err := a.SetDate(fmt.Sprintf("%v", row[5])); err != nil {
			return fmt.Errorf("%w", err)
		}

		topics := strings.Split(fmt.Sprintf("%v", row[2]), "\n")
		for _, t := range topics {
			a.SetTopics(t)
			tm.Increment(Topic(t))
		}

		if row[3] == "" {
			*db = append(*db, *a)
			continue
		}

		qnRow := strings.Split(fmt.Sprintf("%v", row[3]), "\n")

		for _, qn := range qnRow {
			fields := strings.Split(qn, " ")
			year := fields[0]
			number := fields[1]
			if err := a.SetQuestions(year, number, qnDB); err != nil {
				return fmt.Errorf("%w", err)
			}
			qc.Increment(year + " - Q" + number)
		}
		*db = append(*db, *a)
	}

	// check questions with zero articles.
	qc.GetZeroArticleQns(qnDB)

	// sort articles by latest date
	sort.Sort(sort.Reverse(db))

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

	// setup values to write to sheet
	var valueRange sheets.ValueRange
	valueRange.Values = make([][]interface{}, 0, len(*database))

	for _, j := range *database {
		sTopics := strings.Builder{}
		for i, k := range j.Topics {
			if i == len(j.Topics)-1 {
				sTopics.WriteString(string(k))
				break
			}
			sTopics.WriteString(string(k) + "\n")
		}

		sQuestions := strings.Builder{}
		sQuestionsKey := strings.Builder{}
		for i, l := range j.Questions {
			if i == len(j.Questions)-1 {
				sQuestions.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number) + " " + l.Wording)
				sQuestionsKey.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number))
				break
			}
			sQuestions.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number) + " " + l.Wording + "\n")
			sQuestionsKey.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number) + "\n")
		}

		record := make([]interface{}, 0, 5)
		record = append(record, j.Title, j.URL, sTopics.String(), sQuestionsKey.String(), sQuestions.String(), j.DisplayDate)
		valueRange.Values = append(valueRange.Values, record)
	}

	// write to sheet
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
		if i == len(article.Topics)-1 {
			sTopics.WriteString(string(k))
			break
		}
		sTopics.WriteString(string(k) + "\n")
	}

	sQuestions := strings.Builder{}
	sQuestionsKey := strings.Builder{}
	for i, l := range article.Questions {
		if i == len(article.Questions)-1 {
			sQuestions.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number) + " " + l.Wording)
			sQuestionsKey.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number))
			break
		}
		sQuestions.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number) + " " + l.Wording + "\n")
		sQuestionsKey.WriteString(fmt.Sprint(l.Year) + " " + fmt.Sprint(l.Number) + "\n")
	}

	record := make([]interface{}, 0, 6)
	record = append(record, article.Title, article.URL, sTopics.String(), sQuestionsKey.String(), sQuestions.String(), article.DisplayDate)
	valueRange.Values = append(valueRange.Values, record)

	_, err = srv.Spreadsheets.Values.Append(backupSheetID, backupSheetName, &valueRange).InsertDataOption("INSERT_ROWS").ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to append article to backup sheet: %w", err)
	}

	return nil
}

// AppendArticleToOld appends a new article added to the web app database to a predefined, hard-coded Google Sheet.
func AppendArticleToOld(ctx context.Context, article *Article) error {
	srv, err := newSheetsService(ctx)
	if err != nil {
		return fmt.Errorf("unable to start Sheets service: %w", err)
	}

	backupSheetID := os.Getenv("OLD_SHEET_ID")
	backupSheetName := "feed"

	var valueRange sheets.ValueRange
	valueRange.Values = make([][]interface{}, 0, 1)

	tags := make([]interface{}, 0)

	sTopics := strings.Builder{}
	for i, k := range article.Topics {
		if i == len(article.Topics)-1 {
			sTopics.WriteString(string(k))
			tags = append(tags, string(k))
			break
		}
		sTopics.WriteString(string(k) + ", ")
		tags = append(tags, string(k))
	}

	sQuestions := strings.Builder{}
	for i, l := range article.Questions {
		if i == len(article.Questions)-1 {
			sQuestions.WriteString(fmt.Sprintf("%s (%s - Q%s)", l.Wording, l.Year, l.Number))
			tags = append(tags, fmt.Sprintf("%s-Q%s", l.Year, l.Number))
			break
		}
		sQuestions.WriteString(fmt.Sprintf("%s (%s - Q%s)<br><br>", l.Wording, l.Year, l.Number))
		tags = append(tags, fmt.Sprintf("%s-Q%s", l.Year, l.Number))
	}

	record := make([]interface{}, 0, 13)
	record = append(record, article.Title, article.URL, sTopics.String(), sQuestions.String(), article.DisplayDate)
	record = append(record, tags...)
	valueRange.Values = append(valueRange.Values, record)

	_, err = srv.Spreadsheets.Values.Append(backupSheetID, backupSheetName, &valueRange).InsertDataOption("INSERT_ROWS").ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("unable to append article to backup sheet: %w", err)
	}

	return nil
}
