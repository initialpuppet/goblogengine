package blog

import (
	"context"
	"net/http"

	"google.golang.org/appengine/log"

	"goblogengine/appenv"
	"goblogengine/csimg"
	"goblogengine/flash"
	"goblogengine/middleware/basehandler"
	"goblogengine/model"
)

// AdminResetPOST handles a submitted form which indicates the user wants to reset all
// data within the application
func AdminResetPOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	var err error
	var errors []error

	err = model.DeleteAllBlogPostVersion(ctx)
	if err != nil {
		errors = append(errors, err)
	}

	err = model.DeleteAllCategory(ctx)
	if err != nil {
		errors = append(errors, err)
	}

	err = model.DeleteAllAuthor(ctx)
	if err != nil {
		errors = append(errors, err)
	}

	err = model.DeleteAllImage(ctx)
	if err != nil {
		errors = append(errors, err)
	}

	_, err = csimg.DeleteAll(ctx)
	if err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		flash.AddFlash(w, r, "Errors occured during the delete operation")
		for i := range errors {
			log.Infof(ctx, "%v", errors[i])
		}
	}

	author, _ := env.User.(*model.Author)
	a := model.NewAudit("Application reset", "", *author)
	a.Save(ctx)

	flash.AddFlash(w, r, "The application has been reset")
	http.Redirect(w, r, "/admin", http.StatusFound)
	return nil
}

// AdminResetGET displays a form which allows the user to confirm a data reset request
func AdminResetGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	v := env.View.New("admin/reset")
	if err := v.Render(ctx, w, r); err != nil {
		return basehandler.AppErrorDefault(err)
	}

	return nil
}
