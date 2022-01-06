package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
)

func serveFeeds(w http.ResponseWriter, r *http.Request) {
	page := ""

	r.ParseForm()
	addurl := r.Form.Get("add")
	if addurl != "" {
		var f Feed
		f.Url = addurl
		ok := addUrl(f)
		if !ok {
			page += fmt.Sprintf("<h2>Feed with this URL already exists</h2>\n<hr>\n")
		}
	}

	delhash := r.Form.Get("delete")
	if delhash != "" {
		ok := delHash(delhash)
		if !ok {
			page += fmt.Sprintf("<h2>Feed with this hash does not exist</h2>\n<hr>\n")
		}
	}

	page += fmt.Sprintf("<h1>Feeds [%d]</h1>\n<table>\n", len(feeds))
	for _, f := range feeds {
		page += "<tr>\n"
		hash := fmt.Sprintf("%x", md5.Sum([]byte(f.Url)))
		cache := filepath.Join(cachedir, hash+".rss")
		title := f.Title
		status := "?"
		s, ok := getFeedStatus(hash)
		if ok && s == "..." { // updating
			status = s
		} else if ok && s != "OK" { // error updating (e.g. bad url)
			status = fmt.Sprintf(`<abbr title="%s">ERR</abbr>`, html.EscapeString(s))
		} else {
			if _, err := os.Stat(cache); errors.Is(err, os.ErrNotExist) {
				status = fmt.Sprintf(`<abbr title="Out of Date">OOD</abbr>`)
			} else {
				feed, err := feedFromFile(cache)
				if err != nil { // wrong format (e.g. not xml)
					status = fmt.Sprintf(`<abbr title="%s">ERR</abbr>`, html.EscapeString(err.Error()))
				} else {
					title = fmt.Sprintf(`<a href="/%s">%s</a>`, hash, feed.Title)
					if ok && s == "OK" {
						status = s
					} else {
						status = fmt.Sprintf(`<abbr title="Out of Date">OOD</abbr>`)
					}
				}
			}
		}
		page += fmt.Sprintf("<td>[%s]</td>\n", status)
		page += fmt.Sprintf("<td>%s</td>\n", title)
		page += fmt.Sprintf("<td>%s</td>\n", f.Url)
		page += fmt.Sprintf(`<td><a href="/feeds?delete=%s">DEL</a></td>`+"\n", hash)
		page += "</tr>\n"
	}
	page += "</table>\n"

	serveBase(w, r, page, "Feeds")
}
