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
	sheetID   = "1wD8zsIcn9vUPmL749MFAreXx8cfaYeqRfFoGuSnJ2Lk"
	sheetName = "replies"
	credsFile = "credentials.json"
)

// Client encapsulates the Sheets service and sheet configuration.
type Client struct {
	service   *sheets.Service
	sheetID   string
	sheetName string
}

// NewClient initializes a Google Sheets API client and returns a Client instance.
func NewClient(ctx context.Context) (*Client, error) {
	service, err := sheets.NewService(ctx, option.WithCredentialsFile(credsFile))
	if err != nil {
		return nil, fmt.Errorf("unable to create Sheets client: %v", err)
	}

	return &Client{
		service:   service,
		sheetID:   sheetID,
		sheetName: sheetName,
	}, nil
}

// AppendRow adds a new entry to the Google Sheet, formatted with URL, Post Type, and Responded checkbox.
func (c *Client) AppendRow(post post.Post) error {
	// Append data to the specified range in the sheet
	if _, err := c.service.Spreadsheets.Values.Append(c.sheetID, fmt.Sprintf("%s!A:F", c.sheetName), &sheets.ValueRange{
		Values: [][]interface{}{
			{
				post.ID,
				post.URI,
				post.Type,
				post.Content,
				post.Source,
				false,
			},
		},
	}).ValueInputOption("USER_ENTERED").Do(); err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}

	log.Println("Row successfully appended.")

	return nil
}
