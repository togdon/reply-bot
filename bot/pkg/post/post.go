package post

const (
	Games       NYTContentType = "games"
	Connections NYTContentType = "connections"
	Crossword   NYTContentType = "crossword"
	Wordle      NYTContentType = "wordle"
	Strands     NYTContentType = "strands"
)

// Where the type can be one of Strands, Connections, Wordle, Crossword
type NYTContentType string

type Post struct {
	ID      string
	Content string
	Type    NYTContentType
}
