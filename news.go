package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mmcdole/gofeed"
)

const MaxNews = 100

type News struct {
	Feed     *gofeed.Feed
	FeedHash string
	Item     *gofeed.Item
	Hash     string
}

func getNews(news []News, from int) (page string) {
	sort.SliceStable(news, func(i, j int) bool {
		return news[i].Item.PublishedParsed.After(
			*news[j].Item.PublishedParsed,
		)
	})

	page += fmt.Sprintf("<h1>News [%d]</h1>\n<table>\n", len(news))

	early_exit := false
	for i, it := range news {
		if i < from {
			continue
		}

		page += "<tr>\n"
		page += fmt.Sprintf("<td>%s</td>\n", it.Item.PublishedParsed.Format("2006-01-02 15:04"))
		title := fmt.Sprintf(`<a href="/news/%s/%s">%s</a>`+"\n", it.FeedHash, it.Hash, it.Item.Title)

		if it.Item.Content != "" && it.Item.Description != "" {
			page += "<td>\n"
			page += fmt.Sprintf("<details>\n<summary>\n%s</summary>\n%s</details>\n", title, it.Item.Description)
			page += "</td>\n"
		} else {
			page += fmt.Sprintf("<td>%s</td>\n", title)
		}
		page += fmt.Sprintf(`<td><a href="/news/%s">%s</a></td>`+"\n", it.FeedHash, it.Feed.Title)
		page += fmt.Sprintf(`<td><a href="%s">[Original]</a></td>`+"\n", it.Item.Link)
		page += "</tr>\n"

		if i >= from+MaxNews {
			early_exit = true
			break
		}
	}
	page += "</table>\n"

	if from >= MaxNews {
		page += fmt.Sprintf(`<a href="?from=%d">Prev</a> `, from-MaxNews)
	}
	if early_exit {
		page += fmt.Sprintf(`<a href="?from=%d">Next</a>`, from+MaxNews)
	}

	return
}

func serveNewsItem(w http.ResponseWriter, r *http.Request, feedhash, hash string) {
	page := ""

	s, ok := getFeedStatus(feedhash)
	if ok && s == "..." {
		page += "<h1>Feed of this item is currently updating</h1>"
		serveBase(w, r, page)
		return
	}

	var item *gofeed.Item
	var ithash string
	feed, err := feedFromFile(filepath.Join(cachedir, feedhash+".rss"))
	if err != nil {
		page += "<h1>Unable to read the feed file</h1>\n"
		serveBase(w, r, page)
		return
	}
	for _, it := range feed.Items {
		ithash = fmt.Sprintf("%x", md5.Sum([]byte(it.Title+it.Link)))
		if hash != ithash {
			continue
		}
		item = it
		break
	}

	page += fmt.Sprintf("<h1>%s</h1>\n", item.Title)
	var authors string
	for i, a := range item.Authors {
		if a.Name != "" {
			authors += a.Name
		}
		if a.Email != "" {
			authors += "<" + a.Email + ">"
		}
		if i != len(item.Authors)-1 {
			authors += ", "
		}
	}
	page += fmt.Sprintf("<p><small>%s", item.PublishedParsed.Format("2006-01-02 15:04"))
	if authors != "" {
		page += fmt.Sprintf(" | %s", authors)
	}
	page += fmt.Sprintf(` | <a href="%s">[Original]</a></small></p>`+"\n", item.Link)
	if item.Content == "" {
		page += fmt.Sprintf("<p>%s</p>", item.Description)
	} else {
		page += fmt.Sprintf("<p>%s</p>\n", item.Content)
	}
	serveBase(w, r, page)
}

func serveNewsFeed(w http.ResponseWriter, r *http.Request, hash string, update bool, from int) {
	page := ""
	u := ""
	for _, fu := range feeds {
		h := fmt.Sprintf("%x", md5.Sum([]byte(fu)))
		if h == hash {
			u = fu
			break
		}
	}
	if u == "" {
		page += fmt.Sprintf("<h1>Feed with hash %s does not exist</h1>\n", hash)
		serveBase(w, r, page)
		return
	}

	s, ok := getFeedStatus(hash)
	if ok && s == "..." {
		page += fmt.Sprintf("<h1>Feed %s is still updating</h1>\n", u)
		serveBase(w, r, page)
		return
	}

	if update {
		updateFeed(u)
		page += fmt.Sprintf("<h1>Updating feed %s</h1>\n", u)
		serveBase(w, r, page)
		return
	}

	var news []News
	for _, fu := range feeds {
		if fu != u {
			continue
		}

		feed, err := feedFromFile(filepath.Join(cachedir, hash+".rss"))
		if err != nil {
			log.Printf("Unable to load feed from file: %s\n", err)
			continue
		}
		for _, it := range feed.Items {
			ithash := fmt.Sprintf("%x", md5.Sum([]byte(it.Title+it.Link)))
			news = append(news, News{feed, hash, it, ithash})
		}
		break
	}

	if from > len(news)-1 || from < 0 {
		from = 0
	}
	page += getNews(news, from)
	serveBase(w, r, page)
}

func serveNews(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	_, update := r.Form["update"]
	from, err := strconv.Atoi(r.Form.Get("from"))
	if err != nil {
		from = 0
	}

	page := ""

	p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	lvl1, lvl2 := path.Split(p)
	if lvl2 != "news" {
		lvl3 := path.Base(lvl1)
		if lvl3 == "news" {
			serveNewsFeed(w, r, lvl2, update, from)
		} else {
			serveNewsItem(w, r, lvl3, lvl2)
		}
		return
	}

	// update news
	if update {
		updateFeed("")
		page += "<h1>Updating all the feeds</h1>\n"
		serveBase(w, r, page)
		return
	}

	// load news
	var news []News
	for _, fu := range feeds {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(fu)))

		s, ok := getFeedStatus(hash)
		if ok && s == "..." {
			log.Printf("trying to get news while feed %s is still updating\n", fu)
			continue
		}

		feed, err := feedFromFile(filepath.Join(cachedir, hash+".rss"))
		if err != nil {
			log.Printf("Unable to load feed from file: %s\n", err)
			continue
		}
		for _, it := range feed.Items {
			ithash := fmt.Sprintf("%x", md5.Sum([]byte(it.Title+it.Link)))
			news = append(news, News{feed, hash, it, ithash})
		}
	}

	if from > len(news)-1 || from < 0 {
		from = 0
	}
	page += getNews(news, from)
	serveBase(w, r, page)

	/*
		reflock.Lock()
		rf := refreshing
		re := referr
		reflock.Unlock()

		if rf {
			lock.Lock()
			ri := refidx
			lf := len(feeds)
			lock.Unlock()

			w.Write([]byte(fmt.Sprintf("[%d/%d]", ri, lf)))
			return
		}

		if len(newsFeed) < 1 || ref == "1" {
			w.Write([]byte("refreshing...\n"))
			go asyncRefreshNews()
			return
		}

		if re != nil {
			w.Write([]byte(fmt.Sprintf("last refresh error: %s\n", re)))
		}

		w.Write([]byte(fmt.Sprintf("last refresh time: %s\n", reftime)))
		for i, it := range newsFeed {
			str := fmt.Sprintf("%s | [%s] %s | %s\n\n", it.Hash, it.Item.PublishedParsed.Format("2006-01-02 15:04"), it.Item.Title, it.Item.Link)
			w.Write([]byte(str))
			if i >= 50 {
				w.Write([]byte("..."))
				return
			}
		}
	*/
}
