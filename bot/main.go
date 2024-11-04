package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"github.com/mattn/go-mastodon"
	"golang.org/x/net/html"
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
			// fmt.Printf("%v: %v\n", event.Status.ID, event.Status.Content)
			parseURLs(findURLs(event.Status.Content))
		case *mastodon.UpdateEditEvent:
			// fmt.Printf("%v: %v\n", event.Status.ID, event.Status.Account.ID)
			parseURLs(findURLs(event.Status.Content))
		}
	}
}

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

func parseURLs(urls string) {
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

			newsre := regexp.MustCompile(`(?i)aljazeera\.com|apnews\.com|arstechnica\.com|axios\.com|bbc\.co\.uk|bbc\.com|bloomberg\.com|cbc\.ca|cnn\.com|economist\.com|gizmodo\.com|huffpost\.com|ign\.com|kotaku\.com|latimes\.com|(c|ms)?nbc(news)?\.com|npr\.org|nytimes\.com|politico\.com|rawstory\.com|reuters\.com|techcrunch\.com|telegraph\.co\.uk|theathletic\.com|theguardian\.com|thehill\.com|theverge\.com|washingtonpost\.com|wired\.com|wsj\.com`)
			if newsre.MatchString(u) {
				up, _ := url.Parse(u)
				queries := removeTrackers(up)

				if len(queries) == 0 {
					// URLs without query strings

					if up.Scheme != "" && up.Host != "" {
						// only output if there's a valid(ish) URL

						// For whatever reason there's a bunch of URLs that are just
						// https://www.bbc.co.uk/news. This skips printing them, since
						// they're just the front page; note that even with the Trackers
						// they're still just links to the front page
						//
						// discu.eu seems to wrap a bunch of upstream news sites, don't
						// print those either since they're always repeated
						if (up.Host != "www.bbc.co.uk" && up.Path != "/news") &&
							(up.Host != "discu.eu") {
							fmt.Printf("%v://%v%v\n", up.Scheme, up.Host, up.Path)
						}
					}
				} else {
					// URLs with query strings

					// Bloomberg seems to have an Anti-DDoS / Subscription paywall that
					// comes up a lot, don't print those
					if up.Host == "www.bloomberg.com" && up.Path == "/tosv2.html" {
						// Do nothing
					} else if up.Host == "arstechnica.com" {
						// ars has a lot of URLs that get posted that look like https://arstechnica.com/?p=1964606
						// but map to https://arstechnica.com/gadgets/2023/09/new-apple-watch-series-9-improves-siri-processing-iphone-finding-and-more/
						ars_unfurl, _ := url.Parse(unfurlURL(u))
						fmt.Printf("%v://%v%v\n", ars_unfurl.Scheme, ars_unfurl.Host, ars_unfurl.Path)
					} else {
						// there's a query string that's *maybe* useful, but probably not... we'll print it for now
						fmt.Printf("%v://%v%v?%v\n", up.Scheme, up.Host, up.Path, queries.Encode())
					}
				}
			}
		}
	}
}

func removeTrackers(u *url.URL) url.Values {
	var queries url.Values

	if u.RawQuery != "" {
		queries, _ = url.ParseQuery(u.RawQuery)
		for k := range queries {
			trackerre := regexp.MustCompile(`(?i)(at|utm)_(bbc_team|brand|campaign|content|format|link_id|link_origin|link_type|medium|name|placement|ptr_name|social-type|source|term)|ab_channel|campaign|cid|cmp|feature|ftag|giftCopy|fbclid|guc|hsenc|hsmi|itid|leadsource|mbid|mkt_tok|mod|origin|partner|pwapi_token|ref|searchResultPosition|smid|smtyp|source|st|taid|tpcc|unlocked_article_code|url|xtor`)
			if trackerre.MatchString(k) {
				queries.Del(k)
			}
		}
	}
	return queries
}

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
