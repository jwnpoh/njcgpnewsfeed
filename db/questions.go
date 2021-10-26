// Package db provides functions and types relevant to the backend database for the article feed.
package db

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"

	"google.golang.org/api/sheets/v4"
)

// Question is a struct that represents the question object in each article entry in the ArticlesDBByDate.
type Question struct {
	Year    string
	Number  string
	Wording string
	Count   int
}

// QuestionsDB is a map of questions for quick searching.
type QuestionsDB map[string]Question

// InitQuestionsDB maps a list of questions in a file named by filename and maps them to a questions database.
func InitQuestionsDB(ctx context.Context) (QuestionsDB, error) {
	qnDB := make(map[string]Question)

	srv, err := newSheetsService(ctx)
	if err != nil {
		return qnDB, fmt.Errorf("unable to start Sheets service: %w", err)
	}

	sheetRange := "Questions"
	data, err := getSheetData(srv, sheetRange)
	if err != nil {
		return qnDB, fmt.Errorf("unable to get sheet data: %w", err)
	}

	if len(data.Values) == 0 {
		return qnDB, fmt.Errorf("no data found")
	}

	for _, row := range data.Values {
		year := fmt.Sprintf("%v", row[1])
		number := fmt.Sprintf("%v", row[2])
		wording := fmt.Sprintf("%v", row[3])
		count := fmt.Sprintf("%v", row[4])
		c, err := strconv.Atoi(count)
		if err != nil {
			return qnDB, fmt.Errorf("error processing sheet data")
		}

		qn := Question{year, number, wording, c}

		key := year + " " + number
		qnDB[key] = qn
	}
	return qnDB, nil
}

// QuestionsByArticleCount is an object to rank questions by the number of articles tagged to each question.
type QuestionsByArticleCount []Question

// Implement sort.Sort interface
func (q QuestionsByArticleCount) Len() int           { return len(q) }
func (q QuestionsByArticleCount) Less(i, j int) bool { return q[i].Count < q[j].Count }
func (q QuestionsByArticleCount) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }

// RankQuestionsByArticleCount converts a QuestionsDB into a slice of Article in order to count and rank questions by the numebr of articles tagged to each question.
func RankQuestionsByArticleCount(db QuestionsDB) QuestionsByArticleCount {
	allQns := make(QuestionsByArticleCount, 0)

	for _, v := range db {
		allQns = append(allQns, v)
	}

	sort.Sort(sort.Reverse(allQns))

	return allQns
}

// RemoveArticleQuestions updates the count of articles tagged to the questions of a deleted article and returns an updated QuestionsDB.
func RemoveArticleQuestions(article Article, qnDB QuestionsDB) QuestionsDB {
	questions := article.Questions
	for _, v := range questions {
		key := v.Year + " " + v.Number
		a := qnDB[key]
		a.Count--
		qnDB[key] = a
	}
	return qnDB
}

// BackupQuestions backs up the questions database to a predefined, hard-coded Google Sheet.
func BackupQuestions(ctx context.Context, qnDB QuestionsDB) error {
	srv, err := newSheetsService(ctx)
	if err != nil {
		return fmt.Errorf("unable to start Sheets service: %w", err)
	}

	backupSheetID := os.Getenv("SHEET_ID") // Articles DB new
	backupSheetName := "Questions"

	var valueRange sheets.ValueRange
	valueRange.Values = make([][]interface{}, 0, len(qnDB))

	for k, v := range qnDB {
		record := make([]interface{}, 0, 5)
		record = append(record, k, v.Year, v.Number, v.Wording, v.Count)
		valueRange.Values = append(valueRange.Values, record)
	}

	_, err = srv.Spreadsheets.Values.Update(backupSheetID, backupSheetName, &valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to backup data to backup sheet: %w", err)
	}

	return nil
}

// AppendQuestion appends a question to the initialised QuestionsDB.
func AppendQuestion(ctx context.Context, qn Question) error {
	srv, err := newSheetsService(ctx)
	if err != nil {
		return fmt.Errorf("unable to start Sheets service: %w", err)
	}

	backupSheetID := os.Getenv("SHEET_ID") // Articles DB new
	backupSheetName := "Questions"

	var valueRange sheets.ValueRange
	valueRange.Values = make([][]interface{}, 0, 1)

	record := make([]interface{}, 0, 5)
	key := qn.Year + " " + qn.Number
	record = append(record, key, qn.Year, qn.Number, qn.Wording, qn.Count)
	valueRange.Values = append(valueRange.Values, record)

	_, err = srv.Spreadsheets.Values.Append(backupSheetID, backupSheetName, &valueRange).InsertDataOption("INSERT_ROWS").ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to append question to backup sheet: %w", err)
	}

	return nil
}
