package post

import (
	"strings"
)

const (
	Connections NYTContentType = "connections"
	Crossword   NYTContentType = "crossword"
	Wordle      NYTContentType = "wordle"
	Strands     NYTContentType = "strands"
	Cooking     NYTContentType = "cooking"

	BlueSky  APISource = "bluesky"
	Mastodon APISource = "mastodon"
)

// Where the type can be one of Strands, Connections, Wordle, Crossword, or Cooking
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
		string(Cooking),
		string(Wordle),
		string(Strands),
		string(Connections),
		string(Crossword),
	}
	return hashTags
}

func GetContentType(content string, groupNames []string) NYTContentType {
	for _, name := range groupNames {
		if strings.Contains(strings.ToLower(content), name) {
			return NYTContentType(name)
		}
	}
	return "no name"
}
