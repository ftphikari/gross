package main

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
)

const MaxNews = 50

type News struct {
	Feed     *gofeed.Feed
	FeedHash string
	Item     *gofeed.Item
	Hash     string
}

func getAuthorsString(au []*gofeed.Person) (authors string) {
	for i, a := range au {
		if a.Name != "" {
			authors += a.Name
		}
		if a.Email != "" {
			authors += "<" + a.Email + ">"
		}
		if i != len(au)-1 {
			authors += ", "
		}
	}
	return
}

func getDescription(s string) (desc string) {
	domDocTest := html.NewTokenizer(strings.NewReader(s))
	prevtok := domDocTest.Token()
loopDomTest:
	for {
		tt := domDocTest.Next()
		switch {
		case tt == html.ErrorToken:
			break loopDomTest // End of the document,  done
		case tt == html.StartTagToken:
			prevtok = domDocTest.Token()
		case tt == html.TextToken:
			if prevtok.Data == "script" {
				continue
			}

			txt := strings.TrimSpace(html.UnescapeString(string(domDocTest.Text())))
			if prevtok.Data == "p" || prevtok.Data == "div" {
				return
			} else {
				desc += fmt.Sprintf("%s", txt)
			}
		}
	}

	return
}

func serveNewsItem(w http.ResponseWriter, r *http.Request, feedhash, hash string) {
	page := ""

	s, ok := getFeedStatus(feedhash)
	if ok && s == "..." {
		page += "<h1>Feed of this item is currently updating</h1>"
		serveBase(w, r, page, "")
		return
	}

	var item *gofeed.Item
	var ithash string
	feed, err := feedFromFile(filepath.Join(cachedir, feedhash+".rss"))
	if err != nil {
		page += "<h1>Unable to read the feed file</h1>\n"
		serveBase(w, r, page, "")
		return
	}
	for _, it := range feed.Items {
		ithash = getHash(it.Title + it.Link)
		if hash != ithash {
			continue
		}
		item = it
		break
	}

	if item == nil {
		page += "<h1>Unable to find the feed item</h1>\n"
		serveBase(w, r, page, "")
		return
	}

	http.Redirect(w, r, item.Link, http.StatusTemporaryRedirect)
}

func getNews(name string, news []News, from int) (page string) {
	sort.SliceStable(news, func(i, j int) bool {
		return news[i].Item.PublishedParsed.After(
			*news[j].Item.PublishedParsed,
		)
	})

	{
		title := name
		if name == "" {
			title = "News"
		}
		page += fmt.Sprintf(`<h1>[%d] %s</h1>`+"\n", len(news), title)
	}

	early_exit := false
	for i, it := range news {
		if i < from {
			continue
		}

		page += "<feed>\n"
		title := fmt.Sprintf(`<a href="/%s/%s">%s</a>`, it.FeedHash, it.Hash, it.Item.Title)
		authors := getAuthorsString(it.Item.Authors)
		page += "<h3>"
		u, err := url.Parse(it.Feed.Link)
		if err == nil {
			page += fmt.Sprintf(`<img alt src="https://%s/favicon.ico" />`+"\n", u.Hostname())
		}
		page += title
		page += "</h3>\n"

		date := it.Item.PublishedParsed.Format("2006-01-02 15:04")
		page += fmt.Sprintf("<small>%s", date)
		if authors != "" {
			page += fmt.Sprintf(` | %s`, authors)
		}
		if name == "" {
			page += fmt.Sprintf(` | <a href="/%s">%s</a>`, it.FeedHash, it.Feed.Title)
		}

		/*
			if it.Item.Content != "" && it.Item.Description != "" {
				page += fmt.Sprintf("\n<p>%s</p>\n", it.Item.Description)
			}
		*/
		if it.Item.Description != "" {
			desc := getDescription(it.Item.Description)
			page += fmt.Sprintf("<p>%s</p>\n", desc)
		} else if it.Item.Content != "" {
			desc := getDescription(it.Item.Content)
			page += fmt.Sprintf("<p>%s</p>\n", desc)
		}
		page += fmt.Sprintf("</small>\n")

		page += "</feed>\n"

		if i >= from+MaxNews {
			early_exit = true
			break
		}
	}

	if from >= MaxNews {
		page += fmt.Sprintf(`<a href="?from=%d">Prev</a> `, from-MaxNews)
	}
	if early_exit {
		page += fmt.Sprintf(`<a href="?from=%d">Next</a>`, from+MaxNews)
	}

	return
}

func serveNewsFeed(w http.ResponseWriter, r *http.Request, feedhash string) {
	r.ParseForm()
	from, err := strconv.Atoi(r.Form.Get("from"))
	if err != nil {
		from = 0
	}

	page := ""
	u := ""
	for _, f := range feeds {
		h := getHash(f.Url)
		if h == feedhash {
			u = f.Url
			break
		}
	}
	if u == "" {
		page += fmt.Sprintf("<h1>Feed with hash %s does not exist</h1>\n", feedhash)
		serveBase(w, r, page, "")
		return
	}

	s, ok := getFeedStatus(feedhash)
	if ok && s == "..." {
		page += fmt.Sprintf("<h1>Feed is still updating</h1>\n")
		serveBase(w, r, page, "")
		return
	}

	var news []News
	feed, err := feedFromFile(filepath.Join(cachedir, feedhash+".rss"))
	if err != nil {
		page += fmt.Sprintf("<h1>Unable to load the feed %s</h1>\n", u)
		serveBase(w, r, page, "")
		return
	}
	for _, it := range feed.Items {
		ithash := getHash(it.Title + it.Link)
		news = append(news, News{feed, feedhash, it, ithash})
	}

	if from > len(news)-1 || from < 0 {
		from = 0
	}
	page += getNews(feed.Title, news, from)

	serveBase(w, r, page, feed.Title)
}

func serveNews(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	from, err := strconv.Atoi(r.Form.Get("from"))
	if err != nil {
		from = 0
	}

	page := ""

	// if currently updating
	upd, ok := getFeedStatus("")
	if ok && upd == "..." {
		n := 0
		for _, f := range feeds {
			hash := getHash(f.Url)

			s, ok := getFeedStatus(hash)
			if ok && s == "..." {
				n += 1
				continue
			}
		}
		page += fmt.Sprintf("<h1>Feeds are updating [%d/%d]</h1>\n", len(feeds)-n, len(feeds)-1)

		serveBase(w, r, page, "Updating..")
		return
	}

	// load news
	var news []News
	for _, f := range feeds {
		hash := getHash(f.Url)

		s, ok := getFeedStatus(hash)
		if ok && s != "OK" {
			continue
		}

		feed, err := feedFromFile(filepath.Join(cachedir, hash+".rss"))
		if err != nil {
			continue
		}
		for _, it := range feed.Items {
			ithash := getHash(it.Title + it.Link)
			news = append(news, News{feed, hash, it, ithash})
		}
	}

	if from > len(news)-1 || from < 0 {
		from = 0
	}
	page += getNews("", news, from)

	serveBase(w, r, page, "News")
}
