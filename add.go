package main

import (
	"fmt"
	"log"
	"sync"
)

var addLock sync.Mutex

func removeFeed(slice []Feed, s int) []Feed {
	return append(slice[:s], slice[s+1:]...)
}

func saveFeed() error {
	err := exportOPML(feedsfile)
	if err != nil {
		return fmt.Errorf("Unable to save the feeds file:", err)
	}
	return nil
}

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
	err := saveFeed()
	if err != nil {
		log.Println(err)
	}
	return true
}

func delHash(hash string) (ok bool) {
	addLock.Lock()
	defer addLock.Unlock()

	// check if exists
	index := -1
	for i, f := range feeds {
		if hash == getHash(f.Url) {
			index = i
			break
		}
	}
	if index < 0 {
		return false
	}

	feeds = removeFeed(feeds, index)
	err := saveFeed()
	if err != nil {
		log.Println(err)
	}
	return true
}
