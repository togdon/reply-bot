package post

import "strings"

const (
	Connections NYTContentType = "connections"
	Crossword   NYTContentType = "crossword"
	Wordle      NYTContentType = "wordle"
	Strands     NYTContentType = "strands"
	Cooking     NYTContentType = "cooking"

	BlueSky  APISource = "bluesky"
	Mastodon APISource = "mastodon"
)

var types = map[string]NYTContentType{
	"connections": Connections,
	"crossword":   Crossword,
	"wordle":      Wordle,
	"strands":     Strands,
}

func NYTContentTypeFromString(ct string) NYTContentType {
	return types[strings.ToLower(ct)]
}

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
