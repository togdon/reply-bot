package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"github.com/mattn/go-mastodon"
	"github.com/togdon/reply-bot/gsheets"
	"golang.org/x/net/html"
)

const (
	SHEET_ID        = "1wD8zsIcn9vUPmL749MFAreXx8cfaYeqRfFoGuSnJ2Lk"
	SHEET_NAME      = "replies"
	credentialsFile = "credentials.json"
)

var config map[string]string

func main() {
	envs, error := GetConfig()

	if error != nil {
		log.Fatalf("Error loading .env or ENV: %v", error)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)

	client := mastodon.NewClient(&mastodon.Config{
		Server:       envs["MASTODON_SERVER"],
		ClientID:     envs["APP_CLIENT_ID"],
		ClientSecret: envs["APP_CLIENT_SECRET"],
		AccessToken:  envs["APP_TOKEN"],
	})

	events, err := client.StreamingPublic(ctx, false)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		<-sc
		cancel()
	}()

	for e := range events {
		switch event := e.(type) {
		case *mastodon.UpdateEvent:
			if parseContent(event.Status.Content) {
				fmt.Printf("%v\n%v\n\n", event.Status.URI, event.Status.Content)
			}
		case *mastodon.UpdateEditEvent:
			if parseContent(event.Status.Content) {
				fmt.Printf("%v\n%v\n\n", event.Status.URI, event.Status.Content)
			}
		}
	}

	// Example call
	// saveToSheet("https://nytimesguild.org/tech/?CKG", "Cooking")
	// saveToSheet("https://nytimesguild.org/tech/?AT", "Athletic")
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
	var buf bytes.Buffer

	var extractURL func(node *html.Node, w *bytes.Buffer)
	extractURL = func(node *html.Node, w *bytes.Buffer) {
		if node.Type == html.ElementNode && node.Data == "a" {

			url := ""
			class := ""

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
			unfurlre := regexp.MustCompile(`(?i)(aje\.io|amzn\.to|api\.follow\.it|bbc\.in|bit\.ly|buff\.ly|cnet\.co|cnn\.it|d\.pr|dlvr\.it|engt\.co|flic\.kr|goo\.gl|ift\.tt|is\.gd|j\.mp|lat\.ms|nbcnews\.to|npi\.li|nyer\.cm|nyti\.ms|on\.ft\.com|on\.msnbc\.com|on\.natgeo\.com|on\.soundcloud\.com|on\.substack\.co|on\.wsj\.com|ow\.ly|pst\.cr|\/redd\.it|reut\.rs|shar\.es|spoti\.fi|st\.news|t\.co|t\.ly|tcrn\.ch|\/ti\.me|tiny\.cc|tinyurl\.com|trib\.al|w\.wiki|wapo\.st|youtu\.be)/`)
			loops := 0
			for unfurlre.MatchString(u) {
				u = unfurlURL(u)
				loops++
				if loops > 3 {
					// Assume that we're stuck in an inescapable loop, break
					break
				}
			}

			newsre := regexp.MustCompile(`(?i)nytimes\.com`)
			if newsre.MatchString(u) {
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

// Appends a row with the given URL and post type to a Google Sheet.
// If the operation fails, it logs a fatal error.
func saveToSheet(url, postType string) {
	err := gsheets.AppendRow(SHEET_ID, SHEET_NAME, url, postType, credentialsFile)
	if err != nil {
		log.Fatalf("Failed to append row: %v", err)
	} else {
		fmt.Println("Post successfully logged in Google Sheets.")
	}
}
