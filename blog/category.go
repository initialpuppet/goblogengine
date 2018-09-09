package blog

import (
	"context"
	"fmt"
	"net/http"

	"goblogengine/flash"
	"goblogengine/model"
	"goblogengine/slug"

	"goblogengine/appenv"
	"goblogengine/middleware/basehandler"
)

type categoryViewModel struct {
	// Entity properties
	Slug  string
	Title string

	// View properties
	URL string
}

type categoryListViewModel struct {
	Categories []categoryViewModel

	// New Entity properties
	Slug  string
	Title string
}

// CategoryListGET displays a list of categories.
func CategoryListGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	var viewModel = new(categoryListViewModel)

	cats, err := model.GetAllCategory(ctx)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	for _, cat := range cats {
		catvm := categoryViewModel{
			Slug:  cat.Slug,
			Title: cat.Title,
			URL:   fmt.Sprintf("/category/%s", cat.Slug),
		}
		viewModel.Categories = append(viewModel.Categories, catvm)
	}

	v := env.View.New("admin/categorylist")
	v.Data = viewModel
	if err := v.Render(ctx, w, r); err != nil {
		return basehandler.AppErrorDefault(err)
	}

	return nil
}

// CategoryListPOST handles a form submission with a new category.
func CategoryListPOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	viewModel := new(categoryListViewModel)

	if err := r.ParseForm(); err != nil {
		return basehandler.AppErrorDefault(err)
	}
	if err := env.FormDecoder.Decode(viewModel, r.Form); err != nil {
		return basehandler.AppErrorDefault(err)
	}

	cat := model.Category{
		Title: viewModel.Title,
		Slug:  slug.Make(viewModel.Title),
	}
	_, err := cat.Save(ctx)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	flash.AddFlash(w, r, "Category added")
	http.Redirect(w, r, "/admin/category/list", http.StatusFound)
	return nil
}

// CategoryDeletePOST deletes a catgory and removes it from all posts.
func CategoryDeletePOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	slug := r.FormValue("Slug")
	if len(slug) == 0 {
		return basehandler.AppErrorf("No Slug provided for delete operation",
			http.StatusBadRequest,
			nil)
	}

	err := model.DeleteCategory(ctx, slug)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	flash.AddFlash(w, r, "Category deleted")
	http.Redirect(w, r, "/admin/category/list", http.StatusFound)
	return nil
}
