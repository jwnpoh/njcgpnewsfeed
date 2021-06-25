// Package db provides functions and types relevant to the backend database for the article feed.
package db

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func newSheetsService(ctx context.Context) (*sheets.Service, error) {
	b := os.Getenv("CREDENTIALS")

	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON([]byte(b)))
	if err != nil {
		return srv, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	return srv, nil
}

type sheetData struct {
	ID     string
	Range  string
	Values [][]interface{}
}

func getSheetData(srv *sheets.Service, sheetRange string) (*sheetData, error) {
	var sd sheetData

	sd.ID = os.Getenv("SHEET_ID")
	sd.Range = sheetRange

	resp, err := srv.Spreadsheets.Values.Get(sd.ID, sd.Range).Do()
	if err != nil {
		return &sd, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	sd.Values = resp.Values

	return &sd, nil
}
