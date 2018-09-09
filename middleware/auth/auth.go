package auth

import (
	"context"
	"goblogengine/appenv"
	"goblogengine/middleware/basehandler"
	"goblogengine/model"
	"net/http"

	"google.golang.org/appengine/user"
)

// Require reqires the logged in user to have an associated author and
// redirects them to the Welcome page if they do not.
func Require(fn func(context.Context, appenv.AppEnv, http.ResponseWriter,
	*http.Request) *basehandler.AppError) basehandler.HTTPHandler {
	return func(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter,
		r *http.Request) *basehandler.AppError {
		u := env.User
		if u == nil { // shouldn't happen with existing app.yaml
			return basehandler.AppErrorf("Not logged in",
				http.StatusUnauthorized, nil)
		}
		_, ok := u.(*model.Author)
		if !ok {
			http.Redirect(w, r, "/admin/author/add", http.StatusFound)
		}
		return fn(ctx, env, w, r)
	}
}

// AddInfo adds user info to the environment and the view data.
func AddInfo(fn func(context.Context, appenv.AppEnv, http.ResponseWriter,
	*http.Request) *basehandler.AppError) basehandler.HTTPHandler {
	return func(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter,
		r *http.Request) *basehandler.AppError {
		u := user.Current(ctx)
		if u != nil && u.Admin {
			a, err := model.GetAuthorByEmail(ctx, u.Email)
			if err == model.ErrorNoMatchingAuthor {
				env.User = u
			} else if err != nil {
				return basehandler.AppErrorDefault(err)
			} else {
				env.User = a
				env.View.SetUser(a)
			}
		}
		return fn(ctx, env, w, r)
	}
}
