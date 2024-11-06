package post

import "strings"

const (
	Connections NYTContentType = "connections"
	Crossword   NYTContentType = "crossword"
	Wordle      NYTContentType = "wordle"
	Strands     NYTContentType = "strands"

	BlueSky  APISource = "bluesky"
	Mastodon APISource = "mastodon"
)

func NYTContentTypeFromString(ct string) NYTContentType {
	// TODO likely this is not a great place to constantly build this map?
	var types = map[string]NYTContentType{
		"connections": Connections,
		"crossword":   Crossword,
		"wordle":      Wordle,
		"strands":     Strands,
	}
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
