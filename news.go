package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/mmcdole/gofeed"
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

	if item.Content == "" && item.Description == "" {
		http.Redirect(w, r, item.Link, http.StatusTemporaryRedirect)
		return
	}

	// check for autoredirect urls
	for _, f := range feeds {
		h := getHash(f.Url)
		if feedhash == h {
			if f.AutoRedirect {
				http.Redirect(w, r, item.Link, http.StatusTemporaryRedirect)
				return
			}
			break
		}
	}

	page += fmt.Sprintf("<h1>%s</h1>\n", item.Title)
	authors := getAuthorsString(item.Authors)
	page += fmt.Sprintf("<p><small>%s", item.PublishedParsed.Format("2006-01-02 15:04"))
	if authors != "" {
		page += fmt.Sprintf(" | %s", authors)
	}
	page += fmt.Sprintf(` | <a href="/%s">%s</a>`, feedhash, feed.Title)
	page += fmt.Sprintf(` | <a href="%s">[Original]</a></small></p>`+"\n", item.Link)
	if item.Content == "" {
		page += fmt.Sprintf("<p>%s</p>\n", item.Description)
	} else {
		page += fmt.Sprintf("<p>%s</p>\n", item.Content)
	}

	serveBase(w, r, page, item.Title)
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
		page += fmt.Sprintf("<h1>[%d] %s</h1>\n", len(news), title)
	}

	early_exit := false
	for i, it := range news {
		if i < from {
			continue
		}

		page += "<blockquote>\n"
		date := it.Item.PublishedParsed.Format("2006-01-02 15:04")
		title := fmt.Sprintf(`<a href="/%s/%s">%s</a>`, it.FeedHash, it.Hash, it.Item.Title)
		authors := getAuthorsString(it.Item.Authors)
		page += fmt.Sprintf("<h3>%s</h3>\n", title)
		page += fmt.Sprintf(`<small>%s`, date)
		if authors != "" {
			page += fmt.Sprintf(` | %s`, authors)
		}
		if name == "" {
			page += fmt.Sprintf(` | <a href="/%s">%s</a>`, it.FeedHash, it.Feed.Title)
		}
		if it.Item.Content != "" && it.Item.Description != "" {
			page += fmt.Sprintf("\n<blockquote>%s</blockquote>\n", it.Item.Description)
		}
		page += fmt.Sprintf(`</small>`)

		page += "</blockquote>\n"

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

	page += fmt.Sprintf(`<h2><a href="/refresh">Refresh</a></h2>` + "\n<hr>\n")

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
