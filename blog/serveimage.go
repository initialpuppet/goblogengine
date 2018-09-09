package blog

import (
	"context"
	"goblogengine/appenv"
	"goblogengine/csimg"
	"goblogengine/middleware/basehandler"
	"net/http"

	uuid "github.com/satori/go.uuid"

	"goblogengine/external/github.com/gorilla/mux"
)

// ServeImageGET gets an image from Cloud Storage and serves it out.
func ServeImageGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	vars := mux.Vars(r)

	id := vars["imageid"]
	if _, err := uuid.FromString(id); err != nil {
		return basehandler.AppErrorf("Invalid image ID",
			http.StatusBadRequest,
			err)
	}

	img, err := csimg.Read(ctx, id)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(img)

	return nil
}
