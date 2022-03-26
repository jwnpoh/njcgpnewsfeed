// Package db provides functions and types relevant to the backend database for the article feed.
package db

import (
	"context"
	"fmt"
	"os"
	"sort"

	"google.golang.org/api/sheets/v4"
)

// Question is a struct that represents the question object in each article entry in the ArticlesDBByDate.
type Question struct {
	Year    string
	Number  string
	Wording string
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

		qn := Question{year, number, wording}

		key := year + " " + number
		qnDB[key] = qn
	}
	return qnDB, nil
}

// QuestionCounter is an object to count number of articles per question.
type QuestionCounter map[string]int

// InitQuestionCounter returns an initialised QuestionCounter map.
func InitQuestionCounter() QuestionCounter { return make(map[string]int) }

// Increment increases the count for the question in the QuestionCounter map. Argument must be provided in the format qn.Year + " - Q" + qn.Number.
func (qc QuestionCounter) Increment(key string) {
	qc[key]++
}

// Decrement decreases the count for the question in the QuestionCounter map. Argument must be provided in the format qn.Year + " - Q" + qn.Number.
func (qc QuestionCounter) Decrement(key string) {
	qc[key]--
	if qc[key] < 1 {
		delete(qc, key)
	}
}

// GetZeroArticleQns gets the questions that have zero articles tagged.
func (qc QuestionCounter) GetZeroArticleQns(qnDB QuestionsDB) {
  for _, v := range qnDB {
    qn := v.Year + " - Q" + v.Number
    if _, ok := qc[qn]; !ok {
      qc[qn] = 0
    }
  }
}

// QuestionObject is an object to enable ranking of questions by number of articles tagged.
type QuestionObject struct {
	Key   string
	Value int
}

// QuestionsByArticleCount is an object to rank questions by the number of articles tagged to each question.
type QuestionsByArticleCount []QuestionObject

// Implement sort.Sort interface
func (q QuestionsByArticleCount) Len() int           { return len(q) }
func (q QuestionsByArticleCount) Less(i, j int) bool { return q[i].Value < q[j].Value }
func (q QuestionsByArticleCount) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }

// RankQuestionsByArticleCount converts a QuestionsDB into a slice of Article in order to count and rank questions by the numebr of articles tagged to each question.
func RankQuestionsByArticleCount(counter QuestionCounter) QuestionsByArticleCount {
	qc := make(QuestionsByArticleCount, len(counter))

	var i int
	for k, v := range counter {
		qc[i] = QuestionObject{k, v}
		i++
	}

	sort.Sort(sort.Reverse(qc))

	return qc
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
		record = append(record, k, v.Year, v.Number, v.Wording)
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
	record = append(record, key, qn.Year, qn.Number, qn.Wording)
	valueRange.Values = append(valueRange.Values, record)

	_, err = srv.Spreadsheets.Values.Append(backupSheetID, backupSheetName, &valueRange).InsertDataOption("INSERT_ROWS").ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to append question to backup sheet: %w", err)
	}

	return nil
}
