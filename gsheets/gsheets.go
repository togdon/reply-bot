package gsheets

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// GSheetsClient encapsulates the Sheets service and sheet configuration.
type GSheetsClient struct {
	Service   *sheets.Service
	SheetID   string
	SheetName string
}

// NewGSheetsClient initializes a Google Sheets API client and returns a GSheetsClient instance.
func NewGSheetsClient(credentialsFile, sheetID, sheetName string) (*GSheetsClient, error) {
	ctx := context.Background()
	service, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("unable to create Sheets client: %v", err)
	}

	return &GSheetsClient{
		Service:   service,
		SheetID:   sheetID,
		SheetName: sheetName,
	}, nil
}

// AppendRow adds a new entry to the Google Sheet, formatted with URL, Post Type, and Responded checkbox.
func (c *GSheetsClient) AppendRow(url, postType string) error {
	rowData := []interface{}{
		url,
		postType,
		false,
	}

	writeRange := fmt.Sprintf("%s!A:C", c.SheetName) // Columns A to C

	// Append data to the specified range in the sheet
	_, err := c.Service.Spreadsheets.Values.Append(c.SheetID, writeRange, &sheets.ValueRange{
		Values: [][]interface{}{rowData},
	}).ValueInputOption("USER_ENTERED").Do()

	if err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}

	log.Println("Row successfully appended.")
	return nil
}
