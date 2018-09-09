package flashes

import (
	"context"
	"goblogengine/appenv"
	"goblogengine/flash"
	"goblogengine/middleware/basehandler"
	"net/http"
)

// Add adds flash messages from the session to the view data.
func Add(fn func(context.Context, appenv.AppEnv, http.ResponseWriter,
	*http.Request) *basehandler.AppError) basehandler.HTTPHandler {
	return func(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter,
		r *http.Request) *basehandler.AppError {
		flashes, err := flash.ReadFlashes(w, r)
		if err != nil {
			return basehandler.AppErrorDefault(err)
		}
		for i := range flashes {
			env.View.AddFlash(flashes[i])
		}

		return fn(ctx, env, w, r)
	}
}
