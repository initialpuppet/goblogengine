package blog

import (
	"context"
	"html/template"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/russross/blackfriday"

	"goblogengine/appenv"
	"goblogengine/middleware/basehandler"
	"goblogengine/model"

	"goblogengine/external/github.com/gorilla/mux"

	"fmt"
)

// postDisplayViewModel represents the user facing data for displaying a blog post.
type postDisplayViewModel struct {
	// Entity properties
	Slug           string
	Title          string
	BannerImageURL string
	BodyMarkdown   string
	DatePublished  string
	Published      bool
	Categories     []categoryViewModel

	// Sub entity properties
	AuthorName string

	// View properties
	URL                      string
	BodyHTML                 template.HTML
	BodyShortHTML            template.HTML
	DatePublishedDisplay     string
	DatePublishedDisplayFull string
	AuthorURL                string
	EditURL                  string
}

func (vm *postDisplayViewModel) fromEntity(p *model.BlogPostVersion, datefFull string, datefShort string, excerptLen int) {
	vm.Slug = p.Slug
	vm.URL = fmt.Sprintf("/post/%s", p.Slug)
	vm.Title = p.Title
	vm.BannerImageURL = p.BannerImageURL
	vm.DatePublished = p.DatePublished.Format(datefFull)
	vm.DatePublishedDisplay = p.DatePublished.Format(datefShort)
	vm.AuthorName = p.Author.DisplayName
	vm.AuthorURL = fmt.Sprintf("/author/%s", p.Author.Slug)
	vm.EditURL = fmt.Sprintf("/admin/post/edit/%s", p.Slug)

	bHTML := template.HTML(blackfriday.MarkdownCommon([]byte(p.BodyMarkdown)))
	vm.BodyHTML = bHTML

	for i := range p.Categories {
		vm.Categories = append(vm.Categories, categoryViewModel{
			Title: p.Categories[i].Title,
			Slug:  p.Categories[i].Slug,
			URL:   "/category/" + p.Categories[i].Slug,
		})
	}

	excerptMarkdown := p.BodyMarkdown
	shortLen := excerptLen
	if len(p.BodyMarkdown) > shortLen {
		for i, w := shortLen, 0; i < len(p.BodyMarkdown); i += w {
			runeValue, width := utf8.DecodeRuneInString(p.BodyMarkdown[i:])
			if i > shortLen && unicode.IsSpace(runeValue) {
				shortLen = i
				break
			}
			w = width
		}
		excerptMarkdown = p.BodyMarkdown[:shortLen]
		excerptMarkdown = strings.TrimRight(excerptMarkdown, " \r\n")
		excerptMarkdown = fmt.Sprintf("%s...", excerptMarkdown)
	}
	eHTML := template.HTML(blackfriday.MarkdownCommon([]byte(excerptMarkdown)))
	vm.BodyShortHTML = eHTML
}

// PostGET displays a single post.
func PostGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	vars := mux.Vars(r)

	post, err := model.GetBlogPostBySlug(ctx, vars["postslug"])
	if err != nil {
		return basehandler.AppErrorf("Post not found", http.StatusNotFound, err)
	}

	viewModel := new(postDisplayViewModel)
	viewModel.fromEntity(
		post,
		env.Config.DateFormatFull,
		env.Config.DateFormatShort,
		env.Config.ExcerptCharLength)

	v := env.View.New("post")
	v.Data = viewModel
	if err := v.Render(ctx, w, r); err != nil {
		return basehandler.AppErrorDefault(err)
	}

	return nil
}
