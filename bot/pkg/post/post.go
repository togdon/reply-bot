package post

const (
	Connections NYTContentType = "connections"
	Crossword   NYTContentType = "crossword"
	Wordle      NYTContentType = "wordle"
	Strands     NYTContentType = "strands"

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
