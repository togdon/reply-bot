package gsheets

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/togdon/reply-bot/bot/pkg/post"
)

const (
	SHEET_ID   = "1wD8zsIcn9vUPmL749MFAreXx8cfaYeqRfFoGuSnJ2Lk"
)

// GSheetsClient encapsulates the Sheets service and sheet configuration.
type Client struct {
	Service   *sheets.Service
	SheetID   string
	SheetName string
}

// NewGSheetsClient initializes a Google Sheets API client and returns a GSheetsClient instance.
func NewGSheetsClient(creds []byte, sheetID, sheetName string) (*Client, error) {
	ctx := context.Background()
	service, err := sheets.NewService(ctx, option.WithCredentialsJSON(creds))
	if err != nil {
		return nil, fmt.Errorf("unable to create Sheets client: %v", err)
	}

	return &Client{
		Service:   service,
		SheetID:   sheetID,
		SheetName: sheetName,
	}, nil
}

// AppendRow adds a new entry to the Google Sheet, formatted with URL, Post Type, and Responded checkbox.
func (c *Client) AppendRow(post post.Post) error {
	rowData := []interface{}{
		post.ID,
		post.URI,
		post.Type,
		post.Content,
		post.Source,
		false,
	}

	writeRange := fmt.Sprintf("%s!A:F", c.SheetName) // Columns A to F

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
