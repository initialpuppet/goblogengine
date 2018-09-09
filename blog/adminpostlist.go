package blog

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"goblogengine/appenv"
	"goblogengine/middleware/basehandler"
	"goblogengine/model"
)

type adminPostListViewModel struct {
	Posts  []postVersionListItemViewModel
	Drafts []postVersionListItemViewModel
}

type postVersionListItemViewModel struct {
	PostID        string
	Title         string
	DateCreated   time.Time
	DatePublished time.Time
	EditURL       string
	PreviewURL    string
	PostURL       string
	Version       int
	Published     bool
	Categories    []categoryViewModel
}

// AdminPostListGET displays a list of published and unpublished blog posts.
func AdminPostListGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	var viewModel = new(adminPostListViewModel)

	postSlice, err := model.GetAllBlogPost(ctx)
	if err != nil {
		return basehandler.AppErrorf("Unable to retrieve post list",
			http.StatusInternalServerError, err)
	}

	for _, post := range postSlice {
		listItem := postVersionListItemViewModel{
			PostID:        post.PostID,
			Version:       post.Version,
			Title:         post.Title,
			DateCreated:   post.DateCreated,
			DatePublished: post.DatePublished,
			EditURL:       fmt.Sprintf("/admin/post/edit/%s", post.Slug),
			PostURL:       fmt.Sprintf("/post/%s", post.Slug),
			PreviewURL: fmt.Sprintf("/admin/post/preview/%s/%d",
				post.Slug, post.Version),
		}

		if post.Published {
			viewModel.Posts = append(viewModel.Posts, listItem)
		} else {
			viewModel.Drafts = append(viewModel.Drafts, listItem)
		}
	}

	sort.Slice(viewModel.Posts, func(i, j int) bool {
		return viewModel.Posts[i].DatePublished.After(viewModel.Posts[j].DatePublished)
	})

	sort.Slice(viewModel.Drafts, func(i, j int) bool {
		return viewModel.Drafts[i].DateCreated.After(viewModel.Drafts[j].DateCreated)
	})

	v := env.View.New("admin/postlist")
	v.Data = viewModel
	if err := v.Render(ctx, w, r); err != nil {
		return basehandler.AppErrorf("", http.StatusInternalServerError, err)
	}

	return nil
}
