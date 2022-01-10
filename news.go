package main

import (
	"fmt"
	"io/ioutil"
	"log"
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

var displayseen bool

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
	r.ParseForm()
	page := ""

	s, ok := getFeedStatus(feedhash)
	if ok && s == "..." {
		page += "<h1>Feed of this item is currently updating</h1>"
		serveBase(w, r, page, "")
		return
	}

	var item *gofeed.Item
	var ithash string
	feed, err := feedFromFile(filepath.Join(feedscache, feedhash+".rss"))
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

	addSeen(feedhash + "-" + hash)
	err = saveSeen()
	if err != nil {
		log.Println(err)
	}

	if _, see := r.Form["see"]; see {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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
		page += "<h1>"
		title := name
		if name == "" {
			title = "News"
			if displayseen {
				page += `<a active title="Toggle seen" href="/?toggleseen"><i class="fas fa-eye-slash"></i></a> `
			} else {
				page += `<a inactive title="Toggle seen" href="/?toggleseen"><i class="fas fa-eye"></i></a> `
			}
		}
		page += fmt.Sprintf("[%d] %s", len(news), title)
		page += "</h1>\n"
	}

	early_exit := false
	for i, it := range news {
		if i < from {
			continue
		}

		isseen := false
		if _, ok := seen[it.FeedHash+"-"+it.Hash]; ok {
			isseen = true
		}

		if isseen {
			page += "<feed seen>\n"
		} else {
			page += "<feed>\n"
		}
		title := fmt.Sprintf(`<a href="/%s/%s">%s</a>`, it.FeedHash, it.Hash, it.Item.Title)
		authors := getAuthorsString(it.Item.Authors)
		page += "<h3>"
		u, err := url.Parse(it.Feed.Link)
		if err == nil {
			page += fmt.Sprintf(`<img alt src="https://%s/favicon.ico" />`+"\n", u.Hostname())
		}
		page += title
		page += "</h3>\n"

		page += fmt.Sprintf("<small>")
		if name == "" && !isseen {
			page += fmt.Sprintf(`<a title="Mark as seen" href="/%s/%s?see"><i class="fas fa-eye"></i></a>`+"\n", it.FeedHash, it.Hash)
		}
		page += fmt.Sprintf(" %s", it.Item.PublishedParsed.Format("2006-01-02 15:04"))
		if authors != "" {
			page += fmt.Sprintf(` | %s`, authors)
		}
		if name == "" {
			page += fmt.Sprintf(` | <a href="/%s">%s</a>`, it.FeedHash, it.Feed.Title)
		}

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
	feed, err := feedFromFile(filepath.Join(feedscache, feedhash+".rss"))
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

func saveSeen() error {
	var b []byte
	for id := range seen {
		b = append(b, []byte(id+"\n")...)
	}
	err := ioutil.WriteFile(seenfile, b, 0644)
	if err != nil {
		return fmt.Errorf("Unable to save feeds: %s", err)
	}
	return nil
}

func addSeen(seenid string) (ok bool) {
	if _, ok := seen[seenid]; ok {
		return false
	}
	seen[seenid] = true
	return true
}

func serveNews(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	from, err := strconv.Atoi(r.Form.Get("from"))
	if err != nil {
		from = 0
	}

	if _, ts := r.Form["toggleseen"]; ts {
		displayseen = !displayseen
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	_, seeall := r.Form["seeall"]

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

		feed, err := feedFromFile(filepath.Join(feedscache, hash+".rss"))
		if err != nil {
			continue
		}
		for _, it := range feed.Items {
			ithash := getHash(it.Title + it.Link)

			isseen := false
			if _, ok := seen[hash+"-"+ithash]; ok {
				isseen = true
			}

			if !displayseen && isseen {
				continue
			}

			news = append(news, News{feed, hash, it, ithash})
			if seeall {
				addSeen(hash + "-" + ithash)
			}
		}
	}

	if seeall {
		err := saveSeen()
		if err != nil {
			log.Println(err)
		}
	}

	if from > len(news)-1 || from < 0 {
		from = 0
	}
	page += getNews("", news, from)

	serveBase(w, r, page, "News")
}
