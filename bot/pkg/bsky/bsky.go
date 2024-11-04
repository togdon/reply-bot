package bsky

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {
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
		Cursor    string //TODO pagination, no idea how it works yet
		HitsTotal int
		Posts     []map[string]interface{} //rest of the struct is here https://docs.bsky.app/docs/api/app-bsky-feed-search-posts#responses
	}

	var search_hits SearchResponse

	if err := json.Unmarshal(body, &search_hits); err != nil {
		panic(err)
	}

	fmt.Printf("found %d\n", search_hits.HitsTotal)

	for _, post := range search_hits.Posts {
		//TODO more specific logic to filter bots / low engagement posts / low follower authors?

		fmt.Println("URI: ", post["uri"]) //TODO not sure how to go from URI here to a browser loadable url?
		// approach here: https://github.com/bluesky-social/atproto/discussions/2523#discussioncomment-9552109
		// [noah] note: rather than using the DID, we use the author handle which makes the url more concise (in most cases) and readable

		uri, ok := post["uri"].(string)
		if !ok {
			fmt.Printf("Error for uri %d: could not cast to string", post["uri"])
		}

		author, ok := post["author"].(map[string]interface{})
		if !ok {
			fmt.Printf("error casting author to string map")
		}

		handle, ok := author["handle"].(string)
		if !ok {
			fmt.Printf("error casting handle to string")
		}

		parts := strings.Split(uri, "/")
		rkey := parts[len(parts)-1]

		fmt.Printf("Associated URL: https://bsky.app/profile/%s/post/%s\n", handle, rkey)
		// postUrl := fmt.Sprintf("https://bsky.app/profile/%s/post/%s\n", did, rkey)

		//TODO write to the google sheet where responses can be generated?

	}
}
