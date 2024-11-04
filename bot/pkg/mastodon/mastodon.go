package mastodon

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mattn/go-mastodon"
	"github.com/togdon/reply-bot/bot/pkg/environment"
	"golang.org/x/net/html"
)

type Client struct {
	mastodonClient *mastodon.Client
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

func NewClient(options ...Option) (*Client, error) {
	var cfg config

	for _, opt := range options {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	return &Client{
		mastodonClient: mastodon.NewClient(&mastodon.Config{
			Server:       cfg.server,
			ClientID:     cfg.clientID,
			ClientSecret: cfg.clientSecret,
			AccessToken:  cfg.accessToken,
		})}, nil
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
				if parseContent(e.Status.Content) {
					fmt.Printf("%v\n%v\n\n", e.Status.URI, e.Status.Content)
				}
			case *mastodon.UpdateEditEvent:
				if parseContent(e.Status.Content) {
					fmt.Printf("%v\n%v\n\n", e.Status.URI, e.Status.Content)
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

// parses the content of a post and returns true if it contains a match for NYT Urls or Games shares
func parseContent(content string) bool {
	if content != "" {
		// first, check for NYT URLs
		if parseURLs(findURLs(content)) {
			// fmt.Printf("Found NYT URL: %v\n", content)
			return true
		}

		// next, check for NYT Games shares
		contentregex := regexp.MustCompile(`(Wordle\s[1-9],[0-9]{3}\s[X,1-6]\/[1-6])|(Connections\nPuzzle\s\#[1-6]{3}\n[🟨|🟩|🟦|🟪]*\n)|(Strands\s\#[1-9]{3}\n.*\n[🟡,🔵]*)|(I\ssolved\sthe\s[0-9]{2}\/[0-9]{2}\/[0-9]{4}\sNew\sYork\sTimes(\sMini)?\sCrossword\sin\s)`)
		if contentregex.MatchString(content) {
			// fmt.Printf("Found NYT Games share: %v\n", content)
			return true
		}
	}

	return false
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

			newsRE := regexp.MustCompile(`(?i)nytimes\.com`)
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
