// Package db provides functions and types relevant to the backend database for the article feed.
package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Retrieve a token from env then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
    tokenString := os.Getenv("TOKEN")
    r := bytes.NewReader([]byte(tokenString))

	tok := &oauth2.Token{}
    err := json.NewDecoder(r).Decode(tok)
    if err != nil {
        log.Fatalf("unable to decode token - %v", err)
    }

	return config.Client(context.Background(), tok)
}

func newSheetsService(ctx context.Context) (*sheets.Service, error) {
	b := os.Getenv("CREDENTIALS")
	config, err := google.ConfigFromJSON([]byte(b), "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
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

	sd.ID = os.Getenv("SHEET_ID") // Articles DB new
	sd.Range = sheetRange

	// resp, err := srv.Spreadsheets.Values.Get(sd.ID, sd.Range).DateTimeRenderOption("FORMATTED_STRING").Do()
	resp, err := srv.Spreadsheets.Values.Get(sd.ID, sd.Range).Do()
	if err != nil {
		return &sd, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	sd.Values = resp.Values

	return &sd, nil
}

