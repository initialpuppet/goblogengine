package blog

import (
	"context"
	"goblogengine/appenv"
	"net/http"

	"goblogengine/middleware/basehandler"
)

// NotFound displays a "Page not found" message and sends appropriate status code
func NotFound(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	return basehandler.AppErrorf("Page not found", http.StatusNotFound, nil)
}
