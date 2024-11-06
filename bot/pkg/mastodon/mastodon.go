package mastodon

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mattn/go-mastodon"
	"github.com/togdon/reply-bot/bot/pkg/environment"
	"github.com/togdon/reply-bot/bot/pkg/gsheets"
	"github.com/togdon/reply-bot/bot/pkg/post"
	"golang.org/x/net/html"
)

const (
	gamesRegex = `(?P<wordle>Wordle\s[1-9],[0-9]{3}\s[X,1-6]\/[1-6])|(?P<connections>Connections\nPuzzle\s\#[1-6]{3}\n[🟨|🟩|🟦|🟪]*\n)|(?P<strands>Strands\s\#[1-9]{3}\n.*\n[🟡,🔵]*)|(?P<crossword>I\ssolved\sthe\s[0-9]{2}\/[0-9]{2}\/[0-9]{4}\sNew\sYork\sTimes(\sMini)?\sCrossword\sin\s)`
)

type Client struct {
	mastodonClient *mastodon.Client
	writeChannel   chan interface{}
	gsheetsClient  *gsheets.Client
}

type config struct {
	server       string
	clientID     string
	clientSecret string
	accessToken  string
}

type Option func(*config) error

func WithConfig(cfg environment.Config) Option {
	return func(c *config) error {
		c.accessToken = cfg.Mastodon.AccessToken
		c.clientID = cfg.Mastodon.ClientID
		c.clientSecret = cfg.Mastodon.ClientSecret
		c.server = cfg.Mastodon.MastodonServer
		return nil
	}
}

func NewClient(ch chan interface{}, gsheetsClient *gsheets.Client, options ...Option) (*Client, error) {
	var cfg config

	for _, opt := range options {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	return &Client{
		mastodonClient: mastodon.NewClient(
			&mastodon.Config{
				Server:       cfg.server,
				ClientID:     cfg.clientID,
				ClientSecret: cfg.clientSecret,
				AccessToken:  cfg.accessToken,
			}),
		gsheetsClient: gsheetsClient,
		writeChannel:  ch,
	}, nil
}

func (c *Client) Run(ctx context.Context, cancel context.CancelFunc, errs chan error) {
	events, err := c.mastodonClient.StreamingPublic(ctx, false)
	if err != nil {
		errs <- err
	}

	for {
		select {
		case event := <-events:
			switch e := event.(type) {
			case *mastodon.UpdateEvent:
				ok, contentType := parseContent(e.Status.Content)
				if ok {
					fmt.Printf("%v\n%v\n\n", e.Status.URI, e.Status.Content)
					post, err := createPost(e.Status.URI, e.Status.Content, contentType)
					if err == nil {
						c.writeChannel <- post
						continue
					}

					fmt.Printf("Unable to parse post: %v", err)

				}
			case *mastodon.UpdateEditEvent:
				ok, contentType := parseContent(e.Status.Content)
				if ok {
					fmt.Printf("%v\n%v\n\n", e.Status.URI, e.Status.Content)
					post, err := createPost(e.Status.URI, e.Status.Content, contentType)
					if err == nil {
						c.writeChannel <- post
						continue
					}

					fmt.Printf("Unable to parse post: %v", err)

				}
			default:
				// How should we handle this?
			}
		case <-ctx.Done():
			fmt.Println("Context cancelled, shutting down Mastodon client...")
			return
		}
	}
}

func (c *Client) Write(ctx context.Context) {

	for {
		select {
		case event := <-c.writeChannel:
			switch e := event.(type) {
			case *post.Post:
				fmt.Printf("Post received: %v", e)
				err := c.gsheetsClient.AppendRow(*e)
				if err != nil {
					log.Printf("unable to write post to gsheet: %v", err)
				}
			default:
				// How should we handle this?
			}
		case <-ctx.Done():
			fmt.Println("Context cancelled, shutting down Mastodon client...")
			return
		}
	}
}

func createPost(URI string, content string, postType post.NYTContentType) (*post.Post, error) {
	if URI == "" || content == "" {
		return nil, fmt.Errorf("empty content or uri. Content: %s, URI: %s", URI, content)
	}
	post := post.Post{
		ID:      URI,
		URI:     URI,
		Content: content,
		Type:    postType,
		Source:  post.Mastodon,
	}
	return &post, nil
}

// parses the content of a post and returns true if it contains a match for NYT Urls or Games shares
func parseContent(content string) (bool, post.NYTContentType) {
	var contentType post.NYTContentType
	if content != "" {
		// first, check for NYT URLs
		if parseURLs(findURLs(content)) {
			// fmt.Printf("Found NYT Cooking URL: %v\n", content)
			return true, post.Cooking
			// return false
		}

		// next, check for NYT Games shares
		re := regexp.MustCompile(gamesRegex)
		if re.MatchString(content) {
			fmt.Printf("group name %s\n", getContentType(content, re))
			// fmt.Printf("Found NYT Games share: %v\n", content)
			return true, contentType
		}
	}

	return false, contentType
}

func getContentType(content string, re *regexp.Regexp) post.NYTContentType {
	groupNames := re.SubexpNames()[1:]
	var contentType post.NYTContentType
	for matchNum, match := range re.FindAllStringSubmatch(content, -1) {
		for groupIdx, group := range match {
			name := groupNames[groupIdx]
			if name == "" {
				name = "*"
			}
			fmt.Printf("#%d text: '%s', group: '%s'\n", matchNum, group, name)
			contentType = post.NYTContentType(name)
			return contentType
		}
	}

	return contentType
}

// findURLs takes a string of event.Status.Content and returns a string of URLs
// found within the content making sure to exclude any URLs that are associated
// with @mentions or #hashtags
func findURLs(s string) string {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}

	var (
		buf        bytes.Buffer
		extractURL func(node *html.Node, w *bytes.Buffer)
	)

	extractURL = func(node *html.Node, w *bytes.Buffer) {
		if node.Type == html.ElementNode && node.Data == "a" {
			var (
				url   string
				class string
			)

			for _, a := range node.Attr {
				if a.Key == "href" {
					url = a.Val
				}
				if a.Key == "class" {
					class = a.Val
				}
			}

			// only write out URLs if no class is associated with it since those are
			// used to signify @mentions and #hashtags. Note that this still catches
			// quote-toots since they're not technically supported, so they look like
			// regular URLs
			if class == "" {
				w.WriteString(url + "\n")
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extractURL(c, w)
		}
	}

	extractURL(doc, &buf)
	return buf.String()
}

func parseURLs(urls string) bool {
	if urls != "" {
		for _, u := range strings.Split(strings.TrimSuffix(urls, "\n"), "\n") {
			// A loop to unfurl the most common URL shorteners; several of these
			// (e.g., xyz -> trib.al -> real url) are used more than once, or have
			// both an http and https link, we loop until they're unfurled
			unfurlRE := regexp.MustCompile(`(?i)(aje\.io|amzn\.to|api\.follow\.it|bbc\.in|bit\.ly|buff\.ly|cnet\.co|cnn\.it|d\.pr|dlvr\.it|engt\.co|flic\.kr|goo\.gl|ift\.tt|is\.gd|j\.mp|lat\.ms|nbcnews\.to|npi\.li|nyer\.cm|nyti\.ms|on\.ft\.com|on\.msnbc\.com|on\.natgeo\.com|on\.soundcloud\.com|on\.substack\.co|on\.wsj\.com|ow\.ly|pst\.cr|\/redd\.it|reut\.rs|shar\.es|spoti\.fi|st\.news|t\.co|t\.ly|tcrn\.ch|\/ti\.me|tiny\.cc|tinyurl\.com|trib\.al|w\.wiki|wapo\.st|youtu\.be)/`)

			for i := 0; unfurlRE.MatchString(u) && i < 4; i++ {
				u = unfurlURL(u)
			}

			newsRE := regexp.MustCompile(`(?i)cooking\.nytimes\.com`)
			if newsRE.MatchString(u) {
				return true
			}
		}
	}

	return false
}

// unfurlURL takes a URL and returns the final URL after following any redirects
func unfurlURL(s string) string {
	var client = &http.Client{
		Timeout: time.Second * 10,
	}

	res, err := client.Head(s)
	if err != nil {
		if os.IsTimeout(err) {
			// timeout, return nothing
			return ""
		} else {
			// non-timeout err, still return nothing
			return ""
		}
	}

	return res.Request.URL.String()
}
