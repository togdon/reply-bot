package post

import "fmt"

const (
	Connections NYTContentType = "connections"
	Crossword   NYTContentType = "crossword"
	Wordle      NYTContentType = "wordle"
	Strands     NYTContentType = "strands"
	Cooking     NYTContentType = "cooking"

	BlueSky  APISource = "bluesky"
	Mastodon APISource = "mastodon"
)

// Where the type can be one of Strands, Connections, Wordle, Crossword
type NYTContentType string

type APISource string

type Post struct {
	ID      string
	URI     string
	Content string
	Source  APISource
	Type    NYTContentType
}

func GetHashtagsFromTypes() []string {
	hashTags := []string{
		hashtagify(Cooking),
		hashtagify(Wordle),
		hashtagify(Strands),
		hashtagify(Connections),
		hashtagify(Crossword),
	}
	return hashTags
}

func hashtagify(val NYTContentType) string {
	return fmt.Sprintf("#%s", val)
}
