package bsky

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/togdon/reply-bot/bot/pkg/gsheets"
	"github.com/togdon/reply-bot/bot/pkg/post"
)

const (
	bskyFeedUrl = "https://public.api.bsky.app/xrpc/app.bsky.feed.getFeed?feed=at://did:plc:ltradugkwaw6yfotr7boceaj/app.bsky.feed.generator/aaapztniwbk46"
)

type BlueskyPost struct {
	URI         string                 `json:"uri"`
	CID         string                 `json:"cid"`
	Author      map[string]interface{} `json:"author"`
	Record      map[string]interface{} `json:"record"`
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

type BlueskyClient struct {
	PollInterval       int
	FeedsConfigFile    string
	GoogleSheetsClient gsheets.Client
}

func (bskyConf *BlueskyClient) Run() {

	feeds := LoadFeedsFromConfigFile(bskyConf.FeedsConfigFile)

	ticker := time.NewTicker(time.Duration(bskyConf.PollInterval) * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		fmt.Println("polling now")
		for _, feedConf := range feeds {
			FetchPostsFromFeed(feedConf)
		}
	}
}

func LoadFeedsFromConfigFile(feedsConfigfileLoc string) []Feed {

	feedFile, err := os.Open(feedsConfigfileLoc)
	if err != nil {
		fmt.Println("Error opening file:", err)
		panic(err)
	}
	defer feedFile.Close()

	feedRaw, err := io.ReadAll(feedFile)
	if err != nil {
		fmt.Println("Error reading file:", err)
		panic(err)
	}

	var feeds []Feed
	err = json.Unmarshal(feedRaw, &feeds)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		panic(err)
	}

	// Print the data
	fmt.Println(feeds)

	return feeds
}

func FetchPostsFromFeed(feedConfig Feed) {

	resp, err := http.Get(feedConfig.MachineUri)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	var feedResponse FeedResponse

	if err := json.Unmarshal(body, &feedResponse); err != nil {
		panic(err)
	}

	for _, feedItem := range feedResponse.Feed {
		//TODO more specific logic to filter bots / low engagement posts / low follower authors?

		url, err := generateBskyUrl(feedItem.Post)
		if err != nil {
			fmt.Printf("error generating bsky url for uri %v", err)
		}

		fmt.Printf("Associated URL: %s\n", url)

		post, err := createPostFromBskyPost(url, feedItem.Post.Record["text"].(string), contentType)
		if err != nil {
			fmt.Printf("error creating post for uri %s: %v\n", url, err)
		}

		// if err := client.AppendRow(post); err != nil {
		// 	fmt.Printf("error writing to google sheet: %v\n", err)
		// }

		fmt.Printf("Post created: %v\n", post)

		//TODO write to the google sheet where responses can be generated?
	}

}

func generateBskyUrl(post BlueskyPost) (string, error) {
	uri := post.URI
	handle, ok := post.Author["handle"]
	if !ok {
		return "", fmt.Errorf("author handle invalid")
	}

	rkey, err := extractRKey(uri)
	if err != nil {
		return "", fmt.Errorf("failed to extract rkey for post: %w", err)
	}

	return fmt.Sprintf("https://bsky.app/profile/%s/post/%s", handle, rkey), nil

}

func extractRKey(uri string) (string, error) {
	parts := strings.Split(uri, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid uri format")
	}

	return parts[len(parts)-1], nil
}

func createPostFromBskyPost(URI string, content string, postType post.NYTContentType) (post.Post, error) {
	if URI == "" || content == "" {
		return post.Post{}, fmt.Errorf("empty content or uri. Content: %s, URI: %s", URI, content)
	}

	rKey, err := extractRKey(URI)
	if err != nil {
		return post.Post{}, fmt.Errorf("failed to extract rkey for post: %w", err)
	}

	post := post.Post{
		ID:      rKey,
		URI:     URI,
		Content: content,
		Type:    postType,
		Source:  post.BlueSky,
	}
	return post, nil
}
