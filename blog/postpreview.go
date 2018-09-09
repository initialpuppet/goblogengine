package blog

import (
	"context"
	"net/http"
	"strconv"

	"goblogengine/appenv"
	"goblogengine/middleware/basehandler"
	"goblogengine/model"

	"goblogengine/external/github.com/gorilla/mux"
)

// AdminPreviewPostVersionGET displays a blog post which has not yet been published.
func AdminPreviewPostVersionGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	vars := mux.Vars(r)

	version, err := strconv.Atoi(vars["version"])
	if err != nil {
		return basehandler.AppErrorf("Invalid version number",
			http.StatusBadRequest, err)
	}

	post, err := model.GetBlogPostVersion(ctx, vars["postslug"], int(version))
	if err != nil {
		return basehandler.AppErrorf("Specified version not found",
			http.StatusNotFound, err)
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
