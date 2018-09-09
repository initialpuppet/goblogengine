// Package atomizer generates an Atom XML feed from data supplied. It does not
// support the full specification.
// See RFC4287: https://tools.ietf.org/html/rfc4287.
// TODO: Category support
package atomizer

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"
)

const xmlns = "http://www.w3.org/2005/Atom"

var g = generator{
	URI:     "https://www.github.com/initialpuppet/goblogengine/",
	Version: "1.0",
	Text:    "GoBlogEngine",
}

// Feed represents an Atom XML feed.
type Feed struct {
	XMLName   xml.Name  `xml:"feed"`
	XMLNS     string    `xml:"xmlns,attr"`
	Title     string    `xml:"title"`
	Subtitle  string    `xml:"subtitle,omitempty"`
	Updated   time.Time `xml:"updated"`
	ID        string    `xml:"id"`
	Links     []Link    `xml:"link"`
	Rights    string    `xml:"rights,omitempty"`
	Generator generator `xml:"generator"`

	Entries []Entry `xml:"entry"`
}

// Generator represents the generator element in the feed.
type generator struct {
	URI     string `xml:"uri,attr"`
	Version string `xml:"version,attr"`
	Text    string `xml:",chardata"`
}

// Link represents a link element in the feed.
type Link struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}

// Entry represents a single entry in the feed.
type Entry struct {
	Title     string    `xml:"title"`
	Link      Link      `xml:"link"`
	ID        string    `xml:"id"`
	Updated   time.Time `xml:"updated"`
	Published time.Time `xml:"published"`
	Author    Author    `xml:"author"`
	Content   Content   `xml:"content"`
}

// Author represents the author in an entry in the feed.
type Author struct {
	Name  string `xml:"name"`
	URI   string `xml:"uri,omitempty"`
	Email string `xml:"email,omitempty"`
}

// Content represents the content in an entry in the feed.
type Content struct {
	Type string `xml:"type,attr"`
	Text string `xml:",chardata"`
}

// NewFeed returns a new copy of Feed configured with initial settings.
func NewFeed(title string, subtitle string, id string,
	rights string, website string, feedurl string) *Feed {
	f := Feed{
		Title:    title,
		Subtitle: subtitle,
		Updated:  time.Now(),
		ID:       id,
		Links: []Link{{
			Rel:  "alternate",
			Type: "text/html",
			Href: website,
		}, {
			Rel:  "self",
			Type: "application/atom+xml",
			Href: feedurl,
		}},
		Rights: rights,
	}

	f.Generator = g
	f.XMLNS = xmlns

	return &f
}

// AddEntry adds an entry to Feed.
func (f *Feed) AddEntry(
	title string,
	url string,
	id string,
	updatedDate time.Time,
	pubDate time.Time,
	authorName string,
	authorURL string,
	authorEmail string,
	contentHTML string) error {

	f.Entries = append(f.Entries, Entry{
		Title: title,
		Link: Link{
			Rel:  "alternate",
			Type: "text/html",
			Href: url,
		},
		ID:        id,
		Updated:   updatedDate,
		Published: pubDate,
		Author: Author{
			Name:  authorName,
			URI:   authorURL,
			Email: authorEmail,
		},
		Content: Content{
			Type: "html",
			Text: contentHTML,
		},
	})

	return nil
}

// ToAtom returns the feed data in XML format meeting the Atom specification.
func (f Feed) ToAtom() ([]byte, error) {
	x, err := xml.MarshalIndent(f, "", " ")
	if err != nil {
		return nil, fmt.Errorf("atomizer: xml marshaling error: %s", err.Error())
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.Write(x)
	return buf.Bytes(), nil
}
