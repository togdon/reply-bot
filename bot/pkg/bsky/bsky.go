package bsky

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/togdon/reply-bot/bot/pkg/gsheets"
)

const (
	bskyFeedUrl = "https://public.api.bsky.app/xrpc/app.bsky.feed.getFeed?feed=at://did:plc:ltradugkwaw6yfotr7boceaj/app.bsky.feed.generator/aaapztniwbk46"
)

type Post struct {
	URI         string                 `json:"uri"`
	CID         string                 `json:"cid"`
	Author      map[string]interface{} `json:"author"`
	Record      map[string]interface{} `json:"record"`
	ReplyCount  int                    `json:"replyCount"`
	RepostCount int                    `json:"reposeCount"`
	QuoteCount  int                    `json:"quoteCount"`
}
type FeedItem struct {
	Post Post `json:"post"`
}

type SearchResponse struct {
	Feed []FeedItem `json:"feed"`
}

func FetchPosts(client *gsheets.Client) {

	resp, err := http.Get(bskyFeedUrl)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	var search_hits SearchResponse

	if err := json.Unmarshal(body, &search_hits); err != nil {
		panic(err)
	}

	for _, item := range search_hits.Feed {
		//TODO more specific logic to filter bots / low engagement posts / low follower authors?

		url, err := generateBskyUrl(item.Post)
		if err != nil {
			fmt.Printf("error generating bsky url for uri %v", err)
		}

		fmt.Printf("Associated URL: %s\n", url)

	}

	//TODO write to the google sheet where responses can be generated?
}

func generateBskyUrl(post Post) (string, error) {
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
