package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
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
		if ok {
			http.Redirect(w, r, "/feeds", http.StatusTemporaryRedirect)
		} else {
			page += fmt.Sprintf("<h2>Feed with this URL already exists</h2>\n<hr>\n")
		}
	}

	delhash := r.Form.Get("delete")
	if delhash != "" {
		ok := delHash(delhash)
		if ok {
			http.Redirect(w, r, "/feeds", http.StatusTemporaryRedirect)
		} else {
			page += fmt.Sprintf("<h2>Feed with this hash does not exist</h2>\n<hr>\n")
		}
	}

	page += fmt.Sprintf("<h1>[%d] Feeds</h1>\n", len(feeds))

	page += fmt.Sprintf(`<urlmng>
	<form action="/feeds">
	<input type="text" name="add" placeholder="Add URL.." autocomplete="off">
	<button title="Add URL" type="submit"><i class="fas fa-plus"></i></button>
	</form>

	<label for="check" title="Import"><i class="fas fa-download"></i></label>
	</urlmng>

	<input id="check" type="checkbox" name="import">
	<filemng id="filemng">
	<form action="/import" method="post" enctype="multipart/form-data">
	<input type="file" name="file">
	<button title="Upload" type="submit"><i class="fas fa-file-import"></i></button>
	</form>
	</filemng>`)

	for _, f := range feeds {
		page += "<feed>\n"
		hash := fmt.Sprintf("%x", md5.Sum([]byte(f.Url)))
		cache := filepath.Join(cachedir, hash+".rss")
		title := "?"
		if f.Title != "" {
			title = f.Title
		}
		imgurl := "/favicon.ico"
		{
			u, err := url.Parse(f.Url)
			if err == nil {
				imgurl = fmt.Sprintf("https://%s/favicon.ico", u.Hostname())
			}
		}
		status := "?"
		s, ok := getFeedStatus(hash)
		if ok && s == "..." { // updating
			status = `<abbr title="Updating">...</abbr>`
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
					{
						u, err := url.Parse(feed.Link)
						if err == nil {
							imgurl = fmt.Sprintf("https://%s/favicon.ico", u.Hostname())
						}
					}
					title = fmt.Sprintf(`<a href="/%s">%s</a>`, hash, feed.Title)
					if ok && s == "OK" {
						status = s
					} else {
						status = fmt.Sprintf(`<abbr title="Out of Date">OOD</abbr>`)
					}
				}
			}
		}
		page += "<h3>"
		page += fmt.Sprintf(`<img alt src="%s" />`+"\n", imgurl)
		page += title
		page += "</h3>\n"
		page += fmt.Sprintf("<small>\n")
		page += fmt.Sprintf("[%s]\n", status)
		page += fmt.Sprintf(` | <a title="Delete the feed" href="/feeds?delete=%s">DEL</a>`+"\n", hash)
		page += fmt.Sprintf("<p>%s</p>\n", f.Url)
		page += fmt.Sprintf("</small>\n")
		page += "</feed>\n"
	}

	serveBase(w, r, page, "Feeds")
}
