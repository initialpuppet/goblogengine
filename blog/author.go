package blog

import (
	"context"
	"net/http"

	"goblogengine/appenv"
	"goblogengine/flash"
	"goblogengine/middleware/basehandler"
	"goblogengine/model"
	"goblogengine/slug"

	"google.golang.org/appengine/user"
)

type authorViewModel struct {
	// Entity properties
	Slug            string
	DisplayName     string
	Email           string
	GoogleAccountID string

	// View properties
	URL     string
	Current bool
}

type authorInsertViewModel struct {
	// Entity properties
	DisplayName string
}

type adminAuthorListViewModel struct {
	Authors []authorViewModel
}

// AdminAuthorListGET displays a list of registered Authors.
func AdminAuthorListGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	authors, err := model.GetAllAuthor(ctx)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	a, _ := env.User.(*model.Author)

	viewModel := new(adminAuthorListViewModel)
	for i := range authors {
		var current bool
		if authors[i].GoogleAccountID == a.GoogleAccountID {
			current = true
		}
		viewModel.Authors = append(viewModel.Authors, authorViewModel{
			DisplayName:     authors[i].DisplayName,
			Email:           authors[i].Email,
			GoogleAccountID: authors[i].GoogleAccountID,
			Current:         current,
			URL:             "/author/" + authors[i].Slug,
		})
	}

	v := env.View.New("admin/authorlist")
	v.Data = viewModel
	if err = v.Render(ctx, w, r); err != nil {
		return basehandler.AppErrorDefault(err)
	}

	return nil
}

// AdminAuthorInsertPOST handles the new author form submission.
func AdminAuthorInsertPOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	u := user.Current(ctx)
	viewModel := new(authorInsertViewModel)

	if err := r.ParseForm(); err != nil {
		return basehandler.AppErrorDefault(err)
	}
	if err := env.FormDecoder.Decode(viewModel, r.PostForm); err != nil {
		return basehandler.AppErrorDefault(err)
	}

	author := model.Author{
		DisplayName:     viewModel.DisplayName,
		Email:           u.Email,
		GoogleAccountID: u.ID,
		Slug:            slug.Make(viewModel.DisplayName),
	}
	_, err := author.Save(ctx)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	a := model.NewAudit("Author registered", "", author)
	a.Save(ctx)

	flash.AddFlash(w, r, "Author registered")
	http.Redirect(w, r, "/admin", http.StatusFound)

	return nil
}

// AdminAuthorInsertGET displays the create author form.
func AdminAuthorInsertGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	v := env.View.New("admin/authorinsert")
	if err := v.Render(ctx, w, r); err != nil {
		return basehandler.AppErrorDefault(err)
	}

	return nil
}
