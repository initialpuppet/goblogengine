// Package datainout handles the bulk import and export of blog posts.
// The format and parser used are brittle and designed to fit one very
// specific set of requirements. Use at your own risk.
//
// TODO: Support for Windows line endings
package datainout

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"goblogengine/model"
)

const (
	lineBreak     = "\n"
	postSeparator = "^^"
	dtFormat1     = "2006-01-02"
	dtFormat2     = "2006-01-02 15:04"
)

// Used to indicate import file parsing stages.
const (
	title      = iota
	categories = iota
	date       = iota
	content    = iota
)

// ParseImportFile accepts an io.Reader and parses it according to the blog
// post import specification, returning a slice of BlogPostVersion.
func ParseImportFile(r io.Reader) ([]model.BlogPostVersion, error) {
	var posts []model.BlogPostVersion

	s := bufio.NewScanner(r)
	current := model.BlogPostVersion{}
	stage := title
	linenum := 0
	for s.Scan() {
		linenum++
		t := s.Text()

		switch stage {
		case title:
			if strings.Trim(t, " ") == "" {
				return nil, fmt.Errorf(
					"datainout: line %d: title cannot be blank", linenum)
			}
			current.Title = t
			stage = date
		case date:
			var d time.Time
			d, err := time.Parse(dtFormat1, t)
			if err != nil {
				d, err = time.Parse(dtFormat2, t)
				if err != nil {
					return nil, fmt.Errorf(
						"datainout: line %d: invalid date", linenum)
				}
			}
			current.DatePublished = d
			stage = categories
		case categories:
			cats := strings.Split(t, ",")
			for i := range cats {
				cat := strings.Trim(cats[i], " ")
				if len(cat) > 0 {
					current.Categories = append(current.Categories,
						model.Category{Title: cat})
				}
			}
			stage = content
			s.Scan()
			linenum++
			if s.Text() != "" {
				return nil, fmt.Errorf(
					"datainout: line %d: expected blank line", linenum)
			}
		case content:
			if t == postSeparator {
				current.BodyMarkdown = current.BodyMarkdown[:len(current.BodyMarkdown)-1]
				posts = append(posts, current)
				current = model.BlogPostVersion{}
				stage = title
			} else {
				current.BodyMarkdown = current.BodyMarkdown + t + lineBreak
			}
		}
	}
	// if there is no final separator, add the last post
	// TODO: add tests to verify things work with / without final separator
	if stage == content {
		posts = append(posts, current)
	}

	return posts, nil
}

// GenerateExport accepts a slice of BlogPostVersion and generates text in
// a format suitable to be exported.
func GenerateExport(posts []model.BlogPostVersion) ([]byte, error) {
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].DatePublished.Before(posts[j].DatePublished)
	})

	var buf bytes.Buffer
	for i := range posts {
		buf.WriteString(posts[i].Title)
		buf.WriteString(lineBreak)

		// If an imported post has a publish date with a time of midnight, the
		// time will not be included in the export. It's unlikely to happen for
		// conventionally authored posts as we check nanoseconds as well.
		pubdate := posts[i].DatePublished
		if pubdate.Hour() == 0 &&
			pubdate.Minute() == 0 &&
			pubdate.Second() == 0 &&
			pubdate.Nanosecond() == 0 {
			buf.WriteString(pubdate.Format(dtFormat1))
		} else {
			buf.WriteString(pubdate.Format(dtFormat2))
		}
		buf.WriteString(lineBreak)

		if len(posts[i].Categories) > 0 {
			buf.WriteString(posts[i].Categories[0].Title)
			for _, cat := range posts[i].Categories[1:] {
				buf.WriteString(", " + cat.Title)
			}
		}
		buf.WriteString(lineBreak)

		buf.WriteString(lineBreak)
		buf.WriteString(posts[i].BodyMarkdown)
		buf.WriteString(lineBreak)
		buf.WriteString(postSeparator)
		buf.WriteString(lineBreak)
	}

	e := buf.Bytes()
	return e, nil
}
