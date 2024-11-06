package bsky

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/togdon/reply-bot/bot/pkg/gsheets"
	"github.com/togdon/reply-bot/bot/pkg/post"
)

const (
	bskyFeedUrl     = "https://public.api.bsky.app/xrpc/app.bsky.feed.getFeed?feed=at://did:plc:ltradugkwaw6yfotr7boceaj/app.bsky.feed.generator/aaapztniwbk46"
	pollInterval    = 10000
	feedsConfigFile = "bot/pkg/bsky/feeds.json"
)

type Record struct {
	Text string `json:"text"`
}

type BlueskyPost struct {
	URI         string                 `json:"uri"`
	CID         string                 `json:"cid"`
	Author      map[string]interface{} `json:"author"`
	Record      Record                 `json:"record"`
	ReplyCount  int                    `json:"replyCount"`
	RepostCount int                    `json:"reposeCount"`
	QuoteCount  int                    `json:"quoteCount"`
}
type FeedItem struct {
	Post BlueskyPost `json:"post"`
}

type FeedResponse struct {
	Feed []FeedItem `json:"feed"`
}

type Feed struct {
	Label      string `json:"Label"`
	UiUri      string `json:"UiUri"`
	MachineUri string `json:"MachineUri"`
}

type Client struct {
	PollInterval       int
	FeedsConfigFile    string
	GoogleSheetsClient *gsheets.Client
}

func NewClient(gsheetsClient *gsheets.Client) *Client {
	return &Client{
		PollInterval:       pollInterval,
		FeedsConfigFile:    feedsConfigFile,
		GoogleSheetsClient: gsheetsClient,
	}
}

func (c *Client) Run(errs chan error) {

	feeds, err := c.loadFeedsFromConfigFile(c.FeedsConfigFile)
	if err != nil {
		errs <- err
	}

	ticker := time.NewTicker(time.Duration(c.PollInterval) * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		fmt.Println("polling now")
		for _, feedConf := range feeds {
			err := c.fetchPostsFromFeed(feedConf)
			if err != nil {
				errs <- err
			}
		}
	}
}

func (c *Client) loadFeedsFromConfigFile(feedsConfigfileLoc string) ([]Feed, error) {

	feedFile, err := os.Open(feedsConfigfileLoc)
	if err != nil {
		return nil, &bSkyError{Message: "error opening bsky config file", Err: err}
	}
	defer feedFile.Close()

	feedRaw, err := io.ReadAll(feedFile)
	if err != nil {
		return nil, &bSkyError{Message: "error reading bsky config file", Err: err}
	}

	var feeds []Feed
	err = json.Unmarshal(feedRaw, &feeds)
	if err != nil {
		return nil, &bSkyError{Message: "error unmarshaling bsky feeds", Err: err}
	}

	// Print the data
	fmt.Println(feeds)

	return feeds, nil
}

func (c *Client) fetchPostsFromFeed(feedConfig Feed) error {

	resp, err := http.Get(feedConfig.MachineUri)
	if err != nil {
		return &bSkyError{Message: "error fetching bsky feed", Err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &bSkyError{Message: "error reading bsky feed response", Err: err}
	}

	var feedResponse FeedResponse
	if err := json.Unmarshal(body, &feedResponse); err != nil {
		return &bSkyError{Message: "error unmarshaling bsky feed response", Err: err}
	}

	for _, feedItem := range feedResponse.Feed {
		//TODO more specific logic to filter bots / low engagement posts / low follower authors?

		url, err := generateBskyUrl(feedItem.Post)
		if err != nil {
			log.Printf("error generating bsky url for uri %v", err)
		}

		log.Printf("Associated URL: %s\n", url)

		post, err := createPostFromBskyPost(
			feedItem.Post.CID,
			url,
			feedItem.Post.Record.Text,
			post.NYTContentType(strings.ToLower(feedConfig.Label)),
		)
		if err != nil {
			log.Printf("error creating bsky post for uri %s: %v\n", url, err)
		}

		if err := c.GoogleSheetsClient.AppendRow(post); err != nil {
			log.Printf("error writing bsky post to google sheet: %v\n", err)
		}

		fmt.Printf("bsky post created: %v\n", post)
	}

	return nil
}

func generateBskyUrl(post BlueskyPost) (string, error) {
	uri := post.URI
	handle, ok := post.Author["handle"]
	if !ok {
		return "", &bSkyError{Message: "error generating bsky urls", Err: fmt.Errorf("author handle invalid")}
	}

	rkey, err := extractRKey(uri)
	if err != nil {
		return "", &bSkyError{Message: "error extracting rkey for post", Err: err}
	}

	return fmt.Sprintf("https://bsky.app/profile/%s/post/%s", handle, rkey), nil

}

func extractRKey(uri string) (string, error) {
	parts := strings.Split(uri, "/")
	if len(parts) < 2 {
		return "", &bSkyError{Message: "error extracting rkey for post", Err: fmt.Errorf("invalid uri format")}
	}

	return parts[len(parts)-1], nil
}

func createPostFromBskyPost(CID, URI, content string, postType post.NYTContentType) (post.Post, error) {
	if URI == "" || content == "" {
		return post.Post{}, &bSkyError{Message: "error creating bsky post", Err: fmt.Errorf("empty content or uri. Content: %s, URI: %s", URI, content)}
	}

	post := post.Post{
		ID:      CID,
		URI:     URI,
		Content: content,
		Type:    postType,
		Source:  post.BlueSky,
	}

	return post, nil
}

// Custom blue sky error type
type bSkyError struct {
	Message string
	Err     error
}

func (e *bSkyError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}
