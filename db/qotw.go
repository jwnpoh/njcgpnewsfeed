package db

import (
	"context"
	"fmt"
	"strconv"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/sheets/v4"
)

type QuestionOfTheWeek struct {
	Question string
	Yes      int
	No       int
  PollID   string
}

type Qotw struct {
	Article Article
	Qn      QuestionOfTheWeek
}

func getQotwSheetData(ctx context.Context) (*sheetData, error) {
	srv, err := newSheetsService(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to start Sheets service: %w", err)
	}

	sheetRange := "QOTW"
	data, err := getSheetData(srv, sheetRange)
	if err != nil {
		return nil, fmt.Errorf("unable to get sheet data: %w", err)
	}

	if len(data.Values) == 0 {
		return nil, fmt.Errorf("no data found")
	}

  return data, nil
}

func GetQotw(ctx context.Context, articles *ArticlesDBByDate) (Qotw, error) {
	var q Qotw

  data, err := getQotwSheetData(ctx)
	if err != nil {
		return q, fmt.Errorf("unable to get sheet data: %w", err)
	}

	row := data.Values[len(data.Values)-1]
	title := row[0]
	question := row[1]
	yes, _ := strconv.Atoi(fmt.Sprintf("%v", row[2]))
	no, _ := strconv.Atoi(fmt.Sprintf("%v", row[3]))
  pollID := fmt.Sprintf("%v", row[4])

	q.Qn.Question = fmt.Sprintf("%v", question)
	q.Qn.Yes = yes
	q.Qn.No = no
  q.Qn.PollID = pollID

	for _, v := range *articles {
		if v.Title == title {
			q.Article = v
		}
	}

	return q, nil
}

func SetQotw(article Article) error {
	return nil
}

func UpdatePoll(selection string, ctx context.Context, articles *ArticlesDBByDate) (QuestionOfTheWeek,error) {
  var q QuestionOfTheWeek
  data, err := getQotwSheetData(ctx)
	if err != nil {
		return q, fmt.Errorf("unable to get sheet data: %w", err)
	}

	row := data.Values[len(data.Values)-1]
	yes, _ := strconv.Atoi(fmt.Sprintf("%v", row[2]))
	no, _ := strconv.Atoi(fmt.Sprintf("%v", row[3]))
  pollID := fmt.Sprintf("%v", row[4])
  
  switch selection {
  case "agree": 
  yes++
  row[2] = fmt.Sprintf("%d", yes)
  case "disagree":
  no++
  row[3] = fmt.Sprintf("%d", no)
  }

  data.Values = data.Values[:len(data.Values)-1]
  data.Values = append(data.Values, row)

	var valueRange sheets.ValueRange
	valueRange.Values = make([][]interface{},0)
  valueRange.Values = append(valueRange.Values, data.Values...)

  // send to Sheets
	srv, err := newSheetsService(ctx)
	if err != nil {
		return q, fmt.Errorf("unable to start Sheets service: %w", err)
	}

	sheetRange := "QOTW"
	_, err = srv.Spreadsheets.Values.Update(data.ID, sheetRange, &valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		return q, fmt.Errorf("unable to update QOTW Sheets data: %w", err)
	}

  q.Yes = yes
  q.No = no
  q.PollID = pollID

  return q, nil
}

func hashPollID(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
