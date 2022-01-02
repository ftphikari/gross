package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

type Outline struct {
	Outlines []Outline `xml:"outline",omitempty`
	Text     string    `xml:"text,attr,omitempty"`
	XmlUrl   string    `xml:"xmlUrl,attr"`
}

type OPML struct {
	XMLName  xml.Name  `xml:"opml"`
	Version  string    `xml:"version,attr"`
	Title    string    `xml:"head>title"`
	Outlines []Outline `xml:"body>outline"`
}

func addOutlines(out []Outline) {
	for _, o := range out {
		if len(o.Outlines) > 0 {
			addOutlines(o.Outlines)
			continue
		}
		if len(o.XmlUrl) > 0 {
			addUrl(o.XmlUrl)
		}
	}
}

func importOPML(filname string) error {
	b, err := ioutil.ReadFile(filname)
	if err != nil {
		return fmt.Errorf("Unable to open OPML file: %s", err)
	}

	var o OPML
	if err := xml.Unmarshal(b, &o); err != nil {
		return fmt.Errorf("Unable to unmarshal OPML: %s", err)
	}

	addOutlines(o.Outlines)

	return nil
}

func exportOPML(f string) error {
	var o OPML
	o.Version = "2.0"
	o.Title = "OPML export from srr"
	for _, u := range feeds {
		o.Outlines = append(o.Outlines, Outline{XmlUrl: u})
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
