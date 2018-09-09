package datainout_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"goblogengine/datainout"
	"goblogengine/model"
)

var postsText = `The first post in the file
2006-01-02
category1, category2, category with spaces

This post has a date but no time.

* This
* is
* Markdown

^^
The second post in the file
2006-01-02 15:04


This post has a date, time but no categories (note the additional line break).

Another line.
^^
Third post
2006-01-02 20:04


Oneliner
^^
`

var dt1, _ = time.Parse(time.RFC3339, "2006-01-02T00:00:00Z")
var dt2, _ = time.Parse(time.RFC3339, "2006-01-02T15:04:00Z")
var dt3, _ = time.Parse(time.RFC3339, "2006-01-02T20:04:00Z")
var postsSlice = []model.BlogPostVersion{{
	Title: "The first post in the file",
	BodyMarkdown: `This post has a date but no time.

* This
* is
* Markdown
`,
	Categories: []model.Category{
		{Title: "category1"},
		{Title: "category2"},
		{Title: "category with spaces"},
	},
	DatePublished: dt1,
}, {
	Title: "The second post in the file",
	BodyMarkdown: `This post has a date, time but no categories (note the additional line break).

Another line.`,
	DatePublished: dt2,
}, {
	Title:         "Third post",
	BodyMarkdown:  "Oneliner",
	DatePublished: dt3,
}}

func TestParseImportFile(t *testing.T) {
	r := strings.NewReader(postsText)
	have, err := datainout.ParseImportFile(r)
	if err != nil {
		t.Fatalf("ParseImportFile failed: %s", err)
	}
	need := postsSlice

	if len(have) != 3 {
		t.Errorf("Incorrect number of posts")
	}

	if !reflect.DeepEqual(need, have) {
		t.Errorf("Imported data does not match\n******\nNeed:\n%v\n******\nHave:\n%v", need, have)
	}

}

func TestGenerateExportFile(t *testing.T) {
	export, err := datainout.GenerateExport(postsSlice)
	if err != nil {
		t.Fatalf("GenerateExport failed: %s", err)
	}
	have := string(export)
	need := postsText

	if have != need {
		t.Errorf("Exported data does not match\n******\nNeed:\n%v\n******\nHave:\n%v", need, have)
	}
}
