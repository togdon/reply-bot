package mastodon

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
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
	gamesRegex = `(?P<wordle>Wordle\s[1-9],[0-9]{3}\s[X,1-6]\/[1-6])|(?P<connections>Connections\nPuzzle\s\#[1-6]{3}\n[🟨|🟩|🟦|🟪]*\n)|(?P<strands>.*Strands\s\#[1-9]{3})|(?P<crossword>I\ssolved\sthe\s[0-9]{2}\/[0-9]{2}\/[0-9]{4}\sNew\sYork\sTimes(\sMini)?\sCrossword\sin\s)`
)

type Client struct {
	mastodonClient *mastodon.Client
	writeChannel   chan interface{}
	gsheetsClient  *gsheets.Client
	logger         *slog.Logger
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

func NewClient(logger *slog.Logger, ch chan interface{}, gsheetsClient *gsheets.Client, options ...Option) (*Client, error) {
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
		logger:        logger,
	}, nil
}

func (c *Client) Run(ctx context.Context, cancel context.CancelFunc, errs chan error) error {
	streamCh := make(chan mastodon.Event)

	// stream from public and then iterate to the known supported tags
	// to use the hashtag api
	events, err := c.mastodonClient.StreamingPublic(ctx, false)
	if err != nil {
		c.logger.Error("unable to send a request to mastodon API", "err", err)
		return err
	}
	sendToStream(streamCh, errs, events, err)
	// for _, tag := range post.GetHashtagsFromTypes() {
	// 	ch, err := c.mastodonClient.StreamingHashtag(ctx, tag, false)
	// 	sendToStream(streamCh, errs, ch, err)
	// }

	for {
		select {
		case event := <-streamCh:
			c.logger.Debug("event received", "event", event)
			switch e := event.(type) {
			case *mastodon.UpdateEvent:
				ok, contentType := c.getContentType(e.Status.Content)
				if ok {
					c.logger.Info("Event content", "uri", e.Status.URI, "content", e.Status.Content)
					post, err := createPost(e.Status.URI, e.Status.Content, contentType)
					if err == nil {
						c.writeChannel <- post
						continue
					} else {
						c.logger.Error("Unable to write post", "err", err)
					}

					c.logger.Error("Unable to parse post", "err", err)

				}
			case *mastodon.UpdateEditEvent:
				ok, contentType := c.getContentType(e.Status.Content)
				if ok {
					c.logger.Info("event content", "uri", e.Status.URI, "content", e.Status.Content)
					post, err := createPost(e.Status.URI, e.Status.Content, contentType)
					if err == nil {
						c.writeChannel <- post
						continue
					} else {
						c.logger.Error("Unable to write post", "err", err)
					}

					c.logger.Error("Unable to parse post", "err", err)

				}
			case *mastodon.ErrorEvent:
				c.logger.Error("Unable to handle event", "err", e.Error())
			default:
				// How should we handle this?
				c.logger.Error("Unable to handle event", "err", err)
			}
		case <-ctx.Done():
			c.logger.Info("Context cancelled, shutting down Mastodon client...")
			return nil
		}
	}
}

// sendToStream takes a channel and redirects events to a different channel
// we're doing this because the mastodon API requires us to use different apis but it doesn't allow
// to share a single channel
func sendToStream(streamCh chan mastodon.Event, errs chan error, inCh chan mastodon.Event, err error) {
	if err != nil {
		errs <- err
	}
	go func() {
		for {
			ev := <-inCh
			streamCh <- ev
		}
	}()
}

func (c *Client) Write(ctx context.Context) {

	for {
		select {
		case event := <-c.writeChannel:
			switch e := event.(type) {
			case *post.Post:
				c.logger.Debug("Post received", "post", e)
				err := c.gsheetsClient.AppendRow(*e)
				if err != nil {
					c.logger.Error("unable to write post to gsheet", "err", err)
				}
			default:
				// How should we handle this?
			}
		case <-ctx.Done():
			c.logger.Info("Context cancelled, shutting down Mastodon client...")
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
func (c *Client) getContentType(content string) (bool, post.NYTContentType) {
	var contentType post.NYTContentType
	if content != "" {
		// first, check for NYT URLs
		if parseURLs(findURLs(content)) {
			c.logger.Info("Found NYT Cooking URL")
			return true, post.Cooking
		}

		// next, check for NYT Games shares
		re := regexp.MustCompile(gamesRegex)
		if re.MatchString(content) {
			c.logger.Info("group name", "name", extractContentType(content, re))
			return true, contentType
		}
	}

	return false, contentType
}

func extractContentType(content string, re *regexp.Regexp) post.NYTContentType {
	groupNames := re.SubexpNames()[1:]
	var contentType post.NYTContentType

	for _, match := range re.FindStringSubmatch(content) {
		contentType = post.GetContentType(match, groupNames)
		return contentType
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
