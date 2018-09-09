package blog

import (
	"fmt"
	"goblogengine/appenv"
	"goblogengine/atomizer"
	"goblogengine/model"
	"net/http"
	"sort"

	"github.com/russross/blackfriday"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// AtomGET returns an Atom feed of recent posts. The number of posts returned
// is set in the application configuration.
//
// TODO: Add some error middleware for XML / JSON handlers
// TODO: Cache all the things: https://www.ctrl.blog/entry/feed-caching
func AtomGET(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	env := appenv.GetEnv()

	baseURL := "http://" + env.Config.BaseDomainName
	feedURL := baseURL + "/atom"
	feedID := baseURL

	blogPosts, err := model.GetBlogPostLimit(ctx, 0, env.Config.FeedSize)
	if err != nil {
		log.Errorf(ctx, "Failure getting posts for atom feed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	sort.Slice(blogPosts, func(i, j int) bool {
		return blogPosts[i].DatePublished.After(blogPosts[j].DatePublished)
	})

	f := atomizer.NewFeed(
		env.Config.BlogName,
		"",
		feedID,
		"",
		baseURL,
		feedURL)

	for _, p := range blogPosts {
		f.AddEntry(
			p.Title,
			fmt.Sprintf("%s/post/%s", baseURL, p.Slug),
			p.PostID,
			p.DateCreated,
			p.DatePublished,
			p.Author.DisplayName,
			fmt.Sprintf("%s/author/%s", baseURL, p.Author.Slug),
			"", // no one puts their email on the Internet
			string(blackfriday.MarkdownCommon([]byte(p.BodyMarkdown))))
	}

	a, err := f.ToAtom()
	if err != nil {
		log.Errorf(ctx, "Failure building atom feed: %v", err)
		return
	}

	w.Header().Set("content-type", "application/atom+xml")
	w.Write(a)
}
