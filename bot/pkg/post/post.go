package post

const (
	Connections NYTContentType = "connections"
	Crossword   NYTContentType = "crossword"
	Wordle      NYTContentType = "wordle"
	Strands     NYTContentType = "strands"
)

// Where the type can be one of Strands, Connections, Wordle, Crossword
type NYTContentType string

type Post struct {
	ID      string
	URI     string
	Content string
	Type    NYTContentType
}
