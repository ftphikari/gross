package main

import (
	"crypto/md5"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/mmcdole/gofeed"
	"github.com/nanmu42/gzip"
)

var (
	confdir    string
	cachedir   string
	feedsfile  string
	feeds      []string
	port       int
	lock       sync.Mutex
	fsys       fs.FS
	updFeed    = make(chan string, 1)
	feedstatus = make(map[string]string)
	statuslock sync.RWMutex
	wg         sync.WaitGroup
)

//go:embed site
var site embed.FS

func feedFromFile(filename string) (*gofeed.Feed, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Unable to open feed: %s", err)
	}
	defer f.Close()

	return gofeed.NewParser().Parse(f)
}

func serveBase(w http.ResponseWriter, r *http.Request, page string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data, err := fs.ReadFile(fsys, "base.htm")
	if err != nil {
		log.Println("Unable to read template:", err)
		return
	}

	t, err := template.New("base").Parse(string(data))
	if err != nil {
		log.Println("Unable to parse template:", err)
		return
	}

	st := struct {
		Page string
	}{
		page,
	}
	if err = t.ExecuteTemplate(w, "base", st); err != nil {
		log.Println("Unable to execute template:", err)
		return
	}
}

func serve(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")

	if p == "" {
		serveIndex(w, r)
		return
	}

	if p == "news" || strings.HasPrefix(p, "news/") {
		serveNews(w, r)
		return
	}

	if f, err := fs.Stat(fsys, p); err == nil {
		if f.IsDir() {
			http.FileServer(http.FS(fsys)).ServeHTTP(w, r)
			return
		}

		if strings.HasSuffix(p, ".css") {
			w.Header().Set("Cache-Control", "public, max-age=86400")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=604800")
		}

		http.FileServer(http.FS(fsys)).ServeHTTP(w, r)
		return
	}

	/*
		if p == "search" || strings.HasPrefix(p, "search/") {
			serveSearch(w, r)
			return
		}

		if _, err := fs.Stat(fsys, p + ".tei"); err != nil {
			serve404(w, r)
			log.Println("serve:", p, "not found.", readUserIP(r))
			return
		}

		servePage(w, r, p+".tei")
	*/
}

func setFeedStatus(hash, s string) {
	statuslock.Lock()
	feedstatus[hash] = s
	statuslock.Unlock()
}

func getFeedStatus(hash string) (string, bool) {
	statuslock.RLock()
	s, ok := feedstatus[hash]
	statuslock.RUnlock()
	return s, ok
}

func fetchFeed(u string) {
	// download rss into the file
	var rss []byte
	hash := fmt.Sprintf("%x", md5.Sum([]byte(u)))
	setFeedStatus(hash, "...")

	//time.Sleep(4 * time.Second) // make loading artificially slow to test things

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	resp, err := client.Get(u)
	if err != nil {
		setFeedStatus(hash, fmt.Sprintf("Unable to load Url: %s", err))
		return
	}
	defer resp.Body.Close()

	rss, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		setFeedStatus(hash, fmt.Sprintf("Unable to load Url: %s", err))
		return
	}

	err = ioutil.WriteFile(filepath.Join(cachedir, hash+".rss"), rss, 0644)
	if err != nil {
		setFeedStatus(hash, fmt.Sprintf("Unable to save to cache: %s", err))
		return
	}

	setFeedStatus(hash, "OK")
	return
}

func feedUpdater() {
	for u := range updFeed {
		if u != "" {
			fetchFeed(u)
			continue
		}

		if err := os.RemoveAll(cachedir); err != nil {
			log.Println("Unable to remove cache dir: %s", err)
			continue
		}
		if err := os.MkdirAll(cachedir, 0755); err != nil {
			log.Println("Unable to make cache dir: %s", err)
			continue
		}

		// refresh feeds
		for _, fu := range feeds {
			wg.Add(1)
			go func(u string) {
				fetchFeed(u)
				wg.Done()
			}(fu)
		}

		wg.Wait()
	}
}

func updateFeed(u string) {
	select {
	case updFeed <- u:
	default:
		log.Println("failed to sent an update: channel is full")
	}
}

func main() {
	flag.IntVar(&port, "p", 8080, "port")
	flag.Parse()

	var err error

	fsys, err = fs.Sub(site, "site")
	if err != nil {
		log.Fatal("Unable to load embed site:", err)
	}

	cachedir, err = os.UserCacheDir()
	if err != nil {
		log.Fatal("Unable to get cache dir:", err)
	}
	cachedir = filepath.Join(cachedir, "gross")
	err = os.MkdirAll(cachedir, 0755)
	if err != nil {
		log.Fatal("Unable to make cache dir:", err)
	}

	confdir, err = os.UserConfigDir()
	if err != nil {
		log.Fatal("Unable to get config dir:", err)
	}
	confdir = filepath.Join(confdir, "gross")
	err = os.MkdirAll(confdir, 0755)
	if err != nil {
		log.Fatal("Unable to make config dir:", err)
	}

	feedsfile = filepath.Join(confdir, "feeds.opml")

	err = importOPML(feedsfile)
	if err != nil {
		log.Println("Unable to import OPML from feeds file:", err)
	}

	go feedUpdater()
	updateFeed("")

	http.Handle("/", gzip.DefaultHandler().WrapHandler(http.HandlerFunc(serve)))
	log.Println("Server started on 127.0.0.1:" + strconv.Itoa(port))

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

/*
func serve() {
		case "a":
			fmt.Print("url: ")
			fmt.Scanln(&str)
			if !addUrl(str) {
				log.Println("Url already exists")
				continue
			}
			err := saveFeeds()
			if err != nil {
				log.Println("Unable to save the feeds")
				continue
			}
			fmt.Println("Feed added in", str)
		case "s":
			err := saveFeeds()
			if err != nil {
				log.Println("Unable to save feeds:", err)
				continue
			}
			fmt.Println("ok")
		case "c":
			feeds = nil
	}
}
*/
