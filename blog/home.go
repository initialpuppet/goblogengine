package blog

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"

	"goblogengine/appenv"
	"goblogengine/middleware/basehandler"
	"goblogengine/model"

	"goblogengine/external/github.com/gorilla/mux"
)

type homeViewModel struct {
	Posts             []postDisplayViewModel
	PostCount         int
	CurrentPageNumber int
	PageNumbers       []pageNumbersViewModel
	PreviousPageURL   string
	NextPageURL       string

	Categories []categoryViewModel
	Authors    []authorViewModel
}

type pageNumbersViewModel struct {
	PageNumber int
	URL        string
}

type postListViewModel struct {
	Posts []postDisplayViewModel
}

// HomeGET displays a paginated list of blog posts.
func HomeGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	vars := mux.Vars(r)
	var viewModel = new(homeViewModel)

	var pageNum, postOffset, postLimit int
	if val, ok := vars["pagenumber"]; ok {
		p, err := strconv.Atoi(val)
		if err != nil {
			return basehandler.AppErrorf("Invalid page number",
				http.StatusInternalServerError, err)
		}
		pageNum = p
	}
	if pageNum < 1 {
		pageNum = 1
	}

	blogPosts, err := model.GetBlogPostLimit(ctx, 0, -1)
	if err != nil {
		return basehandler.AppErrorf("Failed getting posts",
			http.StatusInternalServerError, err)
	}
	sort.Slice(blogPosts, func(i, j int) bool {
		return blogPosts[i].DatePublished.After(blogPosts[j].DatePublished)
	})
	postCount := len(blogPosts)
	pageCount := int(math.Ceil(float64(postCount) / float64(env.Config.PostsPerPage)))
	postOffset = env.Config.PostsPerPage * (pageNum - 1)
	postLimit = int(math.Min(float64(env.Config.PostsPerPage*pageNum), float64(postCount)))

	for i := 0; i < pageCount; i++ {
		n := i + 1
		viewModel.PageNumbers = append(viewModel.PageNumbers, pageNumbersViewModel{
			PageNumber: n,
			URL:        fmt.Sprintf("/page/%d", n),
		})
	}
	if pageNum > 1 {
		viewModel.PreviousPageURL = fmt.Sprintf("/page/%d", pageNum-1)
	}
	if pageNum < pageCount {
		viewModel.NextPageURL = fmt.Sprintf("/page/%d", pageNum+1)
	}
	viewModel.CurrentPageNumber = pageNum

	for i := postOffset; i < postLimit; i++ {
		p := new(postDisplayViewModel)
		p.fromEntity(
			&blogPosts[i],
			env.Config.DateFormatFull,
			env.Config.DateFormatShort,
			env.Config.ExcerptCharLength)
		viewModel.Posts = append(viewModel.Posts, *p)
	}
	viewModel.PostCount = len(viewModel.Posts)

	authors, err := model.GetAllAuthor(ctx)
	if err != nil {
		return basehandler.AppErrorf("Failed getting authors",
			http.StatusInternalServerError, err)
	}
	for _, author := range authors {
		viewModel.Authors = append(viewModel.Authors, authorViewModel{
			DisplayName: author.DisplayName,
			URL:         fmt.Sprintf("/author/%s", author.Slug),
		})
	}

	categories, err := model.GetAllCategory(ctx)
	if err != nil {
		return basehandler.AppErrorf("Failed getting categories",
			http.StatusInternalServerError, err)
	}
	for _, category := range categories {
		viewModel.Categories = append(viewModel.Categories, categoryViewModel{
			Title: category.Title,
			URL:   fmt.Sprintf("/category/%s", category.Slug),
		})
	}

	v := env.View.New("home")
	v.Data = viewModel
	if err := v.Render(ctx, w, r); err != nil {
		return basehandler.AppErrorDefault(err)
	}

	return nil
}
