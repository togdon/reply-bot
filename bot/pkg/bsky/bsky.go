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

		uri, ok := post["uri"].(string)
		if !ok {
			fmt.Printf("Error for uri %d: could not cast to string", post["uri"])
		}

		// fmt.Println(post)
		fmt.Println(post)

		parts := strings.Split(uri, "/")
		did := parts[2]
		rkey := parts[len(parts)-1]

		// approach here: https://github.com/bluesky-social/atproto/discussions/2523#discussioncomment-9552109
		// uri: at://<DID>/<COLLECTION>/<RKEY>
		// https://bsky.app/profile/<DID>/post/<RKEY>
		// https://bsky.app/profile/did:plc:hpweuxc46j4uqsscqdpcwyog/post/3la5nh6vqln2u
		// NOTE: navigating to this url directly after populating the fields yields: Error: uri must be a valid at-uri

		// another approach: construct it using author handle

		fmt.Printf("Associated URL: https://bsky.app/profile/%s/post/%s\n", did, rkey)
		// postUrl := fmt.Sprintf("https://bsky.app/profile/%s/post/%s\n", did, rkey)

		//TODO write to the google sheet where responses can be generated?

	}
}
