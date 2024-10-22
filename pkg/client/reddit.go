package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cameronstanley/go-reddit"
	"github.com/gorilla/feeds"
	"github.com/graph-gophers/dataloader"
	"golang.org/x/oauth2"
)

type linkListingChildren struct {
	Kind string      `json:"kind"`
	Data reddit.Link `json:"data"`
}

type linkListingData struct {
	Modhash  string                `json:"modhash"`
	Children []linkListingChildren `json:"children"`
	After    string                `json:"after"`
	Before   interface{}           `json:"before"`
}

type linkListing struct {
	Kind string          `json:"kind"`
	Data linkListingData `json:"data"`
}

type RedditClient struct {
	HttpClient *http.Client
	UserAgent  string
	Token      *oauth2.Token
}

type GetArticleFn = func(client *RedditClient, link *reddit.Link) (*string, error)
type NowFn = func() time.Time

func RssHandler(redditURL string, now NowFn, client *RedditClient, getArticle GetArticleFn, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.Printf("Fetch %s ", r.URL)
	defer timer(r.URL.String())()

	urlPath := strings.Split(r.URL.Path, "?")[0]

	if strings.Contains(urlPath, ".json") {
		log.Println("WARN: Appending .json to url path is deprecated. This will likely to be removed in the future.")
	} else {
		r.URL.Path += ".json"
	}

	url := fmt.Sprintf("%s%s", redditURL, r.URL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	req.Header.Add("User-Agent", client.UserAgent)
	if client.Token != nil {
		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", client.Token.AccessToken))
	}

	q := req.URL.Query()
	q.Add("sr_detail", "1")
	req.URL.RawQuery = q.Encode()

	resp, err := client.HttpClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Printf("ERROR: HTTP: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		http.Error(w, "Subreddit is private.", resp.StatusCode)
		log.Printf("ERROR: Subreddit is private.")
		return
	}

	// strict check - return 404 instead of being redirected by Reddit.
	if match, err := regexp.MatchString(`(?i)^\/r\/[a-z]+\.json$`, resp.Request.URL.Path); err != nil || !match {
		http.Error(w, "Subreddit not found.", http.StatusNotFound)
		log.Println("ERROR: Subreddit not found")
		return
	}

	var result linkListing
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Printf("ERROR: JSON: %s", err.Error())
		return
	}

	// Subreddit about details are returned in each posts when included with "sr_details=1"
	// Attempt to grab them from the first post
	sr_details := result.Data.Children[0].Data.SRDetails

	feed := &feeds.Feed{
		Title:       sr_details.Title,
		Link:        &feeds.Link{Href: fmt.Sprintf("https://www.reddit.com%s", sr_details.URL)},
		Description: sr_details.PublicDescription,
		Image: &feeds.Image{
			Url:   strings.Split(sr_details.CommunityIcon, "?")[0],
			Title: sr_details.Title,
			Link:  fmt.Sprintf("https://www.reddit.com%s", sr_details.URL),
		},
	}

	var limit int
	limitStr, scoreLimit := r.URL.Query()["scoreLimit"]
	if scoreLimit {
		limit, err = strconv.Atoi(limitStr[0])
		if err != nil {
			scoreLimit = false
		}
	}

	var safe bool
	safeStr, hasSafe := r.URL.Query()["safe"]
	if hasSafe {
		safe = strings.ToLower(safeStr[0]) == "true"
	}

	var flair string
	flairStr, hasFlair := r.URL.Query()["flair"]
	if hasFlair {
		flair = flairStr[0]
	}

	loader := articleLoader(client, getArticle)
	var thunks []dataloader.Thunk
	for _, link := range result.Data.Children {
		if hasSafe && safe && (link.Data.Over18 || strings.ToLower(link.Data.LinkFlairText) == "nsfw") {
			continue
		}

		if scoreLimit && limit > link.Data.Score {
			continue
		}

		if hasFlair && flair != "" && link.Data.LinkFlairText != flair {
			continue
		}

		thunks = append(thunks, loader.Load(ctx, dataKey(link.Data)))
	}

	for _, thunk := range thunks {
		val, err := thunk()
		if err != nil {
			continue
		}

		item := val.(*feeds.Item)
		feed.Items = append(feed.Items, item)
	}

	rss, err := feed.ToRss()
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	w.Header().Set("Content-Type", "application/rss+xml")
	w.Header().Set("Cache-Control", "public, maxage=1800")
	io.WriteString(w, rss)
}

func linkToFeed(client *RedditClient, getArticle GetArticleFn, link *reddit.Link) *feeds.Item {
	var content string
	c, _ := getArticle(client, link)
	if c != nil {
		content = *c
	}
	redditUrl := os.Getenv("REDDIT_URL")
	if redditUrl == "" {
		redditUrl = "https://www.reddit.com"
	}
	author := link.Author
	u, err := url.Parse(link.URL)
	if err == nil {
		_ = u.Host
	}
	t := time.Unix(int64(link.CreatedUtc), 0)
	// if item link is to reddit, replace reddit with REDDIT_URL
	itemLink := fmt.Sprintf(`%s%s`, redditUrl, link.Permalink)
	return &feeds.Item{
		Title:       link.Title,
		Link:        &feeds.Link{Href: itemLink},
		Description: link.Selftext,
		Author:      &feeds.Author{Name: author},
		Created:     t,
		Id:          link.ID,
		Content:     content,
	}
}

type dataKey reddit.Link

func (k dataKey) String() string {
	l := reddit.Link(k)
	return l.ID
}

func (k dataKey) Raw() interface{} { return k }

func articleLoader(client *RedditClient, getArticle GetArticleFn) *dataloader.Loader {
	return dataloader.NewBatchedLoader(func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		wg := &sync.WaitGroup{}
		lock := &sync.Mutex{}
		resultMap := make(map[string]*dataloader.Result)

		for _, key := range keys {
			data := reddit.Link(key.(dataKey))
			wg.Add(1)

			go func(lock *sync.Mutex, wg *sync.WaitGroup, l reddit.Link) {
				defer wg.Done()

				item := linkToFeed(client, getArticle, &l)

				lock.Lock()
				defer lock.Unlock()
				resultMap[l.ID] = &dataloader.Result{Data: item}
			}(lock, wg, data)
		}

		wg.Wait()

		var results []*dataloader.Result
		for _, key := range keys {
			data := reddit.Link(key.(dataKey))
			results = append(results, resultMap[data.ID])
		}

		return results
	}, dataloader.WithBatchCapacity(10))
}

func timer(task string) func() {
	start := time.Now()
	return func() {
		log.Printf("OK %s (%vms)", task, time.Since(start).Milliseconds())
	}
}
