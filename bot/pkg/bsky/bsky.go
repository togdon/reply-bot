package bsky

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/togdon/reply-bot/gsheets"
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

func FetchPosts(client *gsheets.GSheetsClient) {
	BSKY_API_URL := "https://public.api.bsky.app/xrpc"

	BSKY_SEARCH_ENDPOINT := "app.bsky.feed.searchPosts"

	SEARCH_URL := strings.Join([]string{BSKY_API_URL, BSKY_SEARCH_ENDPOINT}, "/")

	search_query := "q=wordle"

	search := strings.Join([]string{SEARCH_URL, search_query}, "?")

	resp, err := http.Get(search)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	type SearchResponse struct {
		Cursor    string `json:"cursor"` //TODO pagination, no idea how it works yet
		HitsTotal int    `json:"hitsTotal"`
		Posts     []Post `json:"posts"` //rest of the struct is here https://docs.bsky.app/docs/api/app-bsky-feed-search-posts#responses
	}

	var search_hits SearchResponse

	if err := json.Unmarshal(body, &search_hits); err != nil {
		panic(err)
	}

	fmt.Printf("found %d\n", search_hits.HitsTotal)

	for _, post := range search_hits.Posts {
		//TODO more specific logic to filter bots / low engagement posts / low follower authors?

		// fmt.Println("URI: ", post["uri"]) //TODO not sure how to go from URI here to a browser loadable url?
		// approach here: https://github.com/bluesky-social/atproto/discussions/2523#discussioncomment-9552109
		// [noah] note: rather than using the DID, we use the author handle which makes the url more concise (in most cases) and readable

		url, err := generateBskyUrl(post)
		if err != nil {
			fmt.Printf("error generating bsky url for uri %v", err)
		}

		fmt.Printf("Associated URL: %s\n", url)

		//TODO write to the google sheet where responses can be generated?

	}
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
