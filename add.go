package main

func addUrl(u string) (ok bool) {
	lock.Lock()
	defer lock.Unlock()

	// check if already exists
	for _, fu := range feeds {
		if fu == u {
			return false
		}
	}

	feeds = append(feeds, u)

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
