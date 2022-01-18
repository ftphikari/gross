package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
)

type Outline struct {
	Title  string `xml:"title,attr,omitempty"`
	Text   string `xml:"text,attr,omitempty"`
	XmlUrl string `xml:"xmlUrl,attr"`
}

type OPML struct {
	XMLName  xml.Name  `xml:"opml"`
	Version  string    `xml:"version,attr"`
	Title    string    `xml:"head>title"`
	Outlines []Outline `xml:"body>outline"`
}

func addOutlines(out []Outline) {
	for _, o := range out {
		var f Feed
		f.Title = "?"
		if len(o.Title) > 0 {
			f.Title = o.Title
		} else if len(o.Text) > 0 {
			f.Title = o.Text
		} else {
			f.Title = "?"
		}
		if len(o.XmlUrl) > 0 {
			f.Url = o.XmlUrl
			addUrl(f)
		}
	}
}

func importOPML(data []byte) error {
	var o OPML
	if err := xml.Unmarshal(data, &o); err != nil {
		return fmt.Errorf("Unable to unmarshal OPML: %s", err)
	}

	addOutlines(o.Outlines)

	err := saveFeed()
	if err != nil {
		log.Println(err)
	}

	return nil
}

func exportOPML(f string) error {
	var o OPML
	o.Version = "2.0"
	o.Title = "OPML export from GRoSS"
	for _, f := range feeds {
		var out Outline
		out.Title = f.Title
		out.Text = f.Title
		out.XmlUrl = f.Url
		o.Outlines = append(o.Outlines, out)
	}

	b, err := xml.Marshal(o)
	if err != nil {
		return fmt.Errorf("Unable to marshal feeds: %s", err)
	}

	err = ioutil.WriteFile(f, b, 0644)
	if err != nil {
		return fmt.Errorf("Unable to save feeds: %s", err)
	}

	return nil
}
