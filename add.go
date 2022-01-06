package main

import (
	"sync"
)

var addLock sync.Mutex

func addUrl(af Feed) (ok bool) {
	addLock.Lock()
	defer addLock.Unlock()

	// check if already exists
	for _, f := range feeds {
		if f.Url == af.Url {
			return false
		}
	}

	feeds = append(feeds, af)

	return true
}

/*
func serveAdd(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	u, ok := r.Form["u"]
	if !ok {
	}
	if u == "" {
		w.Write([]byte("supply url with ?u={}"))
		return
	}
	ok := addUrl(u)
	if ok {
		w.Write([]byte(u + " added"))
	} else {
		w.Write([]byte(u + " already exists"))
	}
}
*/
