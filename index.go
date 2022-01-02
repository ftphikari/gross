package main

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"path/filepath"
)

func serveIndex(w http.ResponseWriter, r *http.Request) {
	page := ""
	page += fmt.Sprintf("<h1>Feeds [%d]</h1>\n<table>\n", len(feeds))
	for _, fu := range feeds {
		page += "<tr>\n"
		hash := fmt.Sprintf("%x", md5.Sum([]byte(fu)))
		title := "?"
		s, ok := getFeedStatus(hash)
		if ok && s == "OK" { // load title
			feed, err := feedFromFile(filepath.Join(cachedir, hash+".rss"))
			if err != nil {
				page += fmt.Sprintf(`<td><abbr title="%s">ERR</abbr></td>`+"\n", err)
				page += fmt.Sprintf("<td>%s</td>\n", title)
				page += fmt.Sprintf("<td>%s</td>\n", fu)
				page += "</tr>\n"
				continue
			}
			title = fmt.Sprintf(`<a href="/news/%s">%s</a>`, hash, feed.Title)
		}
		if ok {
			if s != "..." && s != "OK" {
				s = "ERR" // do not display the whole error message
				// TODO: maybe a tooltip?
			}
			page += fmt.Sprintf("<td>%s</td>\n", s)
		} else {
			page += fmt.Sprintf(`<td><a href="/news/%s?update">UPDATE</a></td>`+"\n", hash)
		}
		page += fmt.Sprintf("<td>%s</td>\n", title)
		page += fmt.Sprintf("<td>%s</td>\n", fu)
		page += "</tr>\n"
	}
	page += "</table>\n"
	serveBase(w, r, page)
}
