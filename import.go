package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func serveImport(w http.ResponseWriter, r *http.Request) {
	page := ""

	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		page += fmt.Sprintf("<h2>Unable to read the file: %s</h2>\n", err)
		serveBase(w, r, page, "")
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		page += fmt.Sprintf("<h2>Unable to read the file: %s</h2>\n", err)
		serveBase(w, r, page, "")
		return
	}

	err = importOPML(data)
	if err != nil {
		page += fmt.Sprintf("<h2>Unable to import the file: %s</h2>\n", err)
		serveBase(w, r, page, "")
		return
	}

	http.Redirect(w, r, "/feeds", http.StatusTemporaryRedirect)
}

func serveExport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename=feeds.xml")
	w.Header().Set("Content-Type", "text/x-opml+xml")

	http.ServeFile(w, r, feedsfile)
}
