package blog

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"goblogengine/appenv"
	"goblogengine/flash"
	"goblogengine/middleware/basehandler"
	"goblogengine/model"
	"goblogengine/slug"
	"goblogengine/taguri"

	"goblogengine/external/github.com/gorilla/mux"
)

type blogPostEditViewModel struct {
	// View data
	VersionCount  int
	Versions      []postVersionListItemViewModel
	AllCategories []string
	NewPost       bool

	// Entity properties
	PostID         string
	Slug           string
	Title          string
	BannerImageURL string
	BodyMarkdown   string
	DatePublished  string
	Published      bool
	Version        int

	// Computed entity properties
	CategoryList string

	// View properties
	PublishImmediately bool
	SelectedVersion    int
	ValidationErrors   map[string]string
}

func (vm *blogPostEditViewModel) addBlogPostVersions(posts []model.BlogPostVersion) {
	for _, p := range posts {
		vmp := postVersionListItemViewModel{
			PostID:        p.PostID,
			Version:       p.Version,
			Title:         p.Title,
			DateCreated:   p.DateCreated,
			DatePublished: p.DatePublished,
			EditURL:       fmt.Sprintf("/admin/post/edit/%s?SelectedVersion=%d", p.Slug, p.Version),
			PreviewURL:    fmt.Sprintf("/admin/post/preview/%s/%d", p.Slug, p.Version),
			PostURL:       fmt.Sprintf("/post/%s", p.Slug),
			Published:     p.Published,
		}

		for i := range p.Categories {
			vmp.Categories = append(vmp.Categories, categoryViewModel{
				Title: p.Categories[i].Title,
				Slug:  p.Categories[i].Slug,
			})
		}

		vm.Versions = append(vm.Versions, vmp)
	}
	vm.VersionCount = len(vm.Versions)
}

func (vm *blogPostEditViewModel) addVersionToEdit(env *appenv.AppEnv, ver *model.BlogPostVersion) {
	vm.PostID = ver.PostID
	vm.Slug = ver.Slug
	vm.Title = ver.Title
	vm.BannerImageURL = ver.BannerImageURL
	vm.BodyMarkdown = ver.BodyMarkdown
	vm.Version = ver.Version
	vm.DatePublished = ver.DatePublished.Format(env.Config.DateFormatForEditing)

	if len(ver.Categories) > 0 {
		catlist := ver.Categories[0].Title
		for _, cat := range ver.Categories[1:] {
			catlist += "," + cat.Title
		}
		vm.CategoryList = catlist
	}
}

func (vm *blogPostEditViewModel) addCategories(ctx context.Context) error {
	cats, err := model.GetAllCategory(ctx)
	if err != nil {
		return err
	}
	for i := range cats {
		vm.AllCategories = append(vm.AllCategories, cats[i].Title)
	}
	return nil
}

// TODO: front and back end validation for the new post form
func (vm *blogPostEditViewModel) validate() bool {
	vm.ValidationErrors = make(map[string]string)

	return true
}

// AdminPostEditGET displays the edit post page
func AdminPostEditGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	viewModel := new(blogPostEditViewModel)
	vars := mux.Vars(r)

	if err := viewModel.addCategories(ctx); err != nil {
		return basehandler.AppErrorDefault(err)
	}

	if postSlug, valid := vars["postslug"]; valid { // editing existing post
		if err := r.ParseForm(); err != nil {
			return basehandler.AppErrorDefault(err)
		}
		if err := env.FormDecoder.Decode(viewModel, r.Form); err != nil {
			return basehandler.AppErrorDefault(err)
		}

		postSlice, err := model.GetBlogPostVersionBySlug(ctx, postSlug)
		if err != nil {
			return basehandler.AppErrorDefault(err)
		}
		if len(postSlice) == 0 {
			return basehandler.AppErrorf("Post not found", http.StatusInternalServerError, nil)
		}
		sort.Slice(postSlice, func(i, j int) bool {
			return postSlice[i].DateCreated.Before(postSlice[j].DateCreated)
		})
		viewModel.addBlogPostVersions(postSlice)

		// Copy the selected version's entity values directly into the view model for editing
		// if no version is selected, use the most recent version
		versionToEdit := new(model.BlogPostVersion)
		if _, valid := r.Form["SelectedVersion"]; valid {
			for i := range viewModel.Versions {
				if postSlice[i].Version == viewModel.SelectedVersion {
					versionToEdit = &postSlice[i]
					break
				}
			}
		} else {
			versionToEdit = &postSlice[len(postSlice)-1]
		}
		viewModel.addVersionToEdit(&env, versionToEdit)

	} else {
		viewModel.NewPost = true
		viewModel.DatePublished = time.Now().Format(env.Config.DateFormatForEditing)
	}

	v := env.View.New("admin/postedit")
	v.Data = viewModel
	if err := v.Render(ctx, w, r); err != nil {
		return basehandler.AppErrorf("", http.StatusInternalServerError, err)
	}

	return nil
}

// AdminPostEditPOST handles a post edit form submission
// TODO: form validation improvements
func AdminPostEditPOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	viewModel := new(blogPostEditViewModel)
	viewModel.ValidationErrors = make(map[string]string)

	if err := r.ParseForm(); err != nil {
		return basehandler.AppErrorDefault(err)
	}
	if err := env.FormDecoder.Decode(viewModel, r.Form); err != nil {
		return basehandler.AppErrorDefault(err)
	}

	author, ok := env.User.(*model.Author)
	if !ok {
		return basehandler.AppErrorf("Not logged in",
			http.StatusInternalServerError, nil)
	}

	pubDate, err := time.Parse(env.Config.DateFormatForEditing, viewModel.DatePublished)
	if err != nil {
		viewModel.ValidationErrors["DatePublished"] = "Invalid date"
	}
	if viewModel.PostID == "" {
		viewModel.NewPost = true
		if viewModel.Slug == "" {
			viewModel.Slug = slug.Make(viewModel.Title)
		}
		tag := taguri.Make(pubDate,
			env.Config.BaseDomainName,
			"",
			slug.Make(env.Config.BlogName),
			viewModel.Slug)
		viewModel.PostID = tag
	}
	entry := new(model.BlogPostVersion)
	entry.Slug = viewModel.Slug
	entry.PostID = viewModel.PostID
	entry.Title = viewModel.Title
	entry.BannerImageURL = viewModel.BannerImageURL
	entry.BodyMarkdown = viewModel.BodyMarkdown
	entry.DatePublished = pubDate
	entry.DateCreated = time.Now()
	entry.Published = viewModel.PublishImmediately
	entry.Author = *author
	cats := strings.Split(viewModel.CategoryList, ",")
	for i := range cats {
		c := model.Category{
			Title: cats[i],
			Slug:  slug.Make(cats[i]),
		}
		entry.Categories = append(entry.Categories, c)
	}

	if len(viewModel.ValidationErrors) == 0 {
		_, err := entry.Save(ctx, viewModel.NewPost)
		if err == model.ErrorPostSlugAlreadyExists {
			viewModel.ValidationErrors["Slug"] = "That custom URL is already in use, try another"
		} else if err != nil {
			return basehandler.AppErrorf(
				"Unable to save blog post",
				http.StatusInternalServerError,
				err)
		}
	}

	if len(viewModel.ValidationErrors) > 0 {
		if err := viewModel.addCategories(ctx); err != nil {
			return basehandler.AppErrorDefault(err)
		}

		posts, err := model.GetBlogPostVersionBySlug(ctx, viewModel.Slug)
		if err != nil {
			return basehandler.AppErrorDefault(err)
		}
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].DateCreated.Before(posts[j].DateCreated)
		})
		viewModel.addBlogPostVersions(posts)

		v := env.View.New("admin/postedit")
		v.Data = viewModel
		if err := v.Render(ctx, w, r); err != nil {
			return basehandler.AppErrorDefault(err)
		}
		return nil
	}

	flash.AddFlash(w, r, "Post updated")
	redirectURL := fmt.Sprintf("/admin/post/edit/%s", entry.Slug)
	http.Redirect(w, r, redirectURL, http.StatusFound)

	return nil
}

// AdminPostUnpublishPOST unpublishes a post with a supplied ID.
// TODO: verify that the author owns the post they are unpublishing.
func AdminPostUnpublishPOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	id := r.FormValue("PostID")
	postTitle := r.FormValue("PostTitle")

	// TODO: only accept relative URLs or store in session
	editURL := r.FormValue("ContinueURL")

	err := model.UnpublishBlogPost(ctx, id)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	fmsg := fmt.Sprintf("%s unpublished", postTitle)
	flash.AddFlash(w, r, fmsg)
	http.Redirect(w, r, editURL, http.StatusFound)

	return nil
}

// AdminPostDeletePOST deletes a post with the supplied ID.
func AdminPostDeletePOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	id := r.FormValue("PostID")
	postTitle := r.FormValue("PostTitle")

	err := model.DeleteBlogPost(ctx, id)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	fmsg := fmt.Sprintf("%s deleted", postTitle)
	flash.AddFlash(w, r, fmsg)
	http.Redirect(w, r, "/admin/post/list", http.StatusFound)

	return nil
}

// AdminPostPublishPOST publishes a post with a specified ID and version.
func AdminPostPublishPOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	id := r.FormValue("PostID")
	version := r.FormValue("Version")
	postTitle := r.FormValue("PostTitle")

	versionNum, err := strconv.Atoi(version)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	err = model.PublishBlogPostVersion(ctx, id, versionNum)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	fmsg := fmt.Sprintf("%s published", postTitle)
	flash.AddFlash(w, r, fmsg)
	http.Redirect(w, r, "/admin/post/list", http.StatusFound)

	return nil
}
