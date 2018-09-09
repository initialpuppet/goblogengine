// TODO: Test helper methods instead of building the struct directly
package atomizer_test

import (
	"goblogengine/atomizer"
	"testing"
	"time"
)

var dtf, _ = time.Parse(time.RFC3339, "2005-08-28T12:29:29Z")
var dte1, _ = time.Parse(time.RFC3339, "2005-08-28T03:34:35Z")
var dte2, _ = time.Parse(time.RFC3339, "2005-07-31T12:29:29Z")
var have = atomizer.Feed{
	Title:    "Test Feed",
	Subtitle: "Feed subtitle",
	Updated:  dtf,
	ID:       "urn:uuid:31ab48a0-dc22-48df-8941-fadffd0b81de",
	Links: []atomizer.Link{{
		Rel:  "alternate",
		Type: "text/html",
		Href: "http://example.org/",
	}, {
		Rel:  "self",
		Type: "application/atom+xml",
		Href: "http://example.org/feed.atom",
	}},
	Rights: "Foo righters, and barring any wrongs",
	Entries: []atomizer.Entry{{
		Title: "First Entry",
		Link: atomizer.Link{
			Rel:  "alternate",
			Type: "text/html",
			Href: "http://example.org/2005/04/02/first-entry",
		},
		ID:        "urn:uuid:74bfc20b-bb3c-445e-8929-448d504d2372",
		Updated:   dte1,
		Published: dte1,
		Author: atomizer.Author{
			Name:  "Mr B Foo",
			URI:   "http://example.org",
			Email: "foo@example.com",
		},
		Content: atomizer.Content{
			Type: "html",
			Text: "<p>Preamble to first entry</p><h2>First entry subheading</h2><p>Foo bar</p>",
		}}, {
		Title: "Second Entry",
		Link: atomizer.Link{
			Rel:  "alternate",
			Type: "text/html",
			Href: "http://example.org/2005/04/02/second-entry",
		},
		ID:        "urn:uuid:0ce79e97-6991-43aa-81eb-dcda1ced0ef7",
		Updated:   dte2,
		Published: dte2,
		Author: atomizer.Author{
			Name:  "Mr F Bar",
			URI:   "http://example.org",
			Email: "bar@example.com",
		},
		Content: atomizer.Content{
			Type: "html",
			Text: "<p>Preamble to second entry</p><h2>Second entry subheading</h2><p>Foo bar</p>",
		},
	}},
}

var need = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
 <title>Test Feed</title>
 <subtitle>Feed subtitle</subtitle>
 <updated>2005-08-28T12:29:29Z</updated>
 <id>urn:uuid:31ab48a0-dc22-48df-8941-fadffd0b81de</id>
 <link rel="alternate" type="text/html" href="http://example.org/"></link>
 <link rel="self" type="application/atom+xml" href="http://example.org/feed.atom"></link>
 <rights>Foo righters, and barring any wrongs</rights>
 <generator uri="https://www.github.com/initialpuppet/goblogengine/" version="1.0">GoBlogEngine</generator>
 <entry>
  <title>First Entry</title>
  <link rel="alternate" type="text/html" href="http://example.org/2005/04/02/first-entry"></link>
  <id>urn:uuid:74bfc20b-bb3c-445e-8929-448d504d2372</id>
  <updated>2005-08-28T03:34:35Z</updated>
  <published>2005-08-28T03:34:35Z</published>
  <author>
   <name>Mr B Foo</name>
   <uri>http://example.org</uri>
   <email>foo@example.com</email>
  </author>
  <content type="html">&lt;p&gt;Preamble to first entry&lt;/p&gt;&lt;h2&gt;First entry subheading&lt;/h2&gt;&lt;p&gt;Foo bar&lt;/p&gt;</content>
 </entry>
 <entry>
  <title>Second Entry</title>
  <link rel="alternate" type="text/html" href="http://example.org/2005/04/02/second-entry"></link>
  <id>urn:uuid:0ce79e97-6991-43aa-81eb-dcda1ced0ef7</id>
  <updated>2005-07-31T12:29:29Z</updated>
  <published>2005-07-31T12:29:29Z</published>
  <author>
   <name>Mr F Bar</name>
   <uri>http://example.org</uri>
   <email>bar@example.com</email>
  </author>
  <content type="html">&lt;p&gt;Preamble to second entry&lt;/p&gt;&lt;h2&gt;Second entry subheading&lt;/h2&gt;&lt;p&gt;Foo bar&lt;/p&gt;</content>
 </entry>
</feed>`

func TestToAtom(t *testing.T) {
	output, err := have.ToAtom()
	if err != nil {
		t.Fatal(err)
	}
	have := string(output)

	if have != need {
		t.Errorf("Generated output did not match expected output.\nHAVE\n%s\n\nNEED\n%s", output, need)
	}
}
