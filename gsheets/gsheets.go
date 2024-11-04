package gsheets

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	UrlColumn      = 0
	PostTypeColumn = 1 // News, Games, Cooking
	ResponseColumn = 2 // Whether the post has been responded to
)

// NewSheetsClient initializes a Google Sheets API client using the provided credentials.
func NewSheetsClient(credentialsFile string) (*sheets.Service, error) {
	ctx := context.Background()
	client, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("unable to create Sheets client: %v", err)
	}
	return client, nil
}

// AppendRow adds a new entry to the specified Google Sheet, formatted with URL, Post Type, and Responded checkbox.
func AppendRow(sheetID, sheetName, url, postType string, credentialsFile string) error {
	client, err := NewSheetsClient(credentialsFile)
	if err != nil {
		return fmt.Errorf("failed to initialize Sheets client: %v", err)
	}

	rowData := []interface{}{
		url,
		postType,
		false,
	}

	writeRange := fmt.Sprintf("%s!A:C", sheetName) // Columns A to C

	// Append data to the specified range in the sheet
	_, err = client.Spreadsheets.Values.Append(sheetID, writeRange, &sheets.ValueRange{
		Values: [][]interface{}{rowData},
	}).ValueInputOption("USER_ENTERED").Do()

	if err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}

	log.Println("Row successfully appended.")
	return nil
}