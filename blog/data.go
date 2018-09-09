package blog

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"goblogengine/appenv"
	"goblogengine/datainout"
	"goblogengine/flash"
	"goblogengine/middleware/basehandler"
	"goblogengine/model"
	"goblogengine/slug"
	"goblogengine/taguri"
)

type importPostsViewModel struct {
	PublishImmediately bool
}

// AdminImportPostsPOST handles a form submission with a text file and imports
// the contents
func AdminImportPostsPOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	publishImmediately := (r.FormValue("PublishImmediately") == "on")

	author, ok := env.User.(*model.Author)
	if !ok {
		return basehandler.AppErrorf("Not logged in",
			http.StatusInternalServerError, nil)
	}

	f, fh, err := r.FormFile("importfile")
	if err == http.ErrMissingFile {
		return basehandler.AppErrorf("Please choose a file to import",
			http.StatusInternalServerError, err)
	}
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	articles, err := datainout.ParseImportFile(f)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	var errs []error
	var count int
	for i := range articles {
		articles[i].Slug = slug.Make(articles[i].Title)

		articles[i].PostID = taguri.Make(articles[i].DatePublished,
			env.Config.BaseDomainName,
			"",
			slug.Make(env.Config.BlogName),
			articles[i].Slug)

		articles[i].Published = publishImmediately
		articles[i].DateCreated = time.Now()
		articles[i].Author = *author

		for j := range articles[i].Categories {
			catslug := slug.Make(articles[i].Categories[j].Title)
			articles[i].Categories[j].Slug = catslug
		}

		_, err := articles[i].Save(ctx, true)
		if err != nil {
			errs = append(errs, err)
		} else {
			count++
		}
	}

	errCount := len(errs)

	alog := fmt.Sprintf(
		"file: %s, posts added: %d, errors: %d, publish: %v",
		fh.Filename,
		count,
		errCount,
		publishImmediately)
	a := model.NewAudit("Import from file", alog, *author)
	a.Save(ctx)

	flashText := fmt.Sprintf("%d posts imported", count)
	if errCount > 0 {
		flashText += fmt.Sprintf(" with %d errors", errCount)
	}
	flash.AddFlash(w, r, flashText)
	http.Redirect(w, r, "/admin/data", http.StatusFound)
	return nil
}

// AdminDataGET displays the data management page.
func AdminDataGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	v := env.View.New("admin/data")
	if err := v.Render(ctx, w, r); err != nil {
		return basehandler.AppErrorDefault(err)
	}
	return nil
}

// AdminExportPostsPOST returns all current posts as a downloadable text file.
func AdminExportPostsPOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	posts, err := model.GetAllBlogPost(ctx)
	if err != nil {
		return basehandler.AppErrorf("Failed getting posts",
			http.StatusInternalServerError, err)
	}

	output, err := datainout.GenerateExport(posts)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment;filename=blogexport.txt")
	w.Write(output)

	return nil
}
