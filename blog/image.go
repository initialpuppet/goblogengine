package blog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"goblogengine/appenv"
	"goblogengine/csimg"
	"goblogengine/flash"
	"goblogengine/middleware/basehandler"
	"goblogengine/model"
)

// Types of image URL that can be passed to the front-end.
const (
	ImgURLLocal = iota
	ImgURLCloudStorage
	ImgURLServingURL
)

type imageListViewModel struct {
	Images []imageViewModel
}

func (vm *imageListViewModel) addImages(baseURL string, imgs []model.Image, imgHost int) {
	for i := range imgs {
		img := new(imageViewModel)
		img.fromEntity(baseURL, &imgs[i], imgHost)
		vm.Images = append(vm.Images, *img)
	}
}

type imageViewModel struct {
	// Entity properties
	ID   string
	Name string
	Size string

	// Computed properties
	URL      string
	LocalURL string
}

func (vm *imageViewModel) fromEntity(baseURL string, img *model.Image, imgHost int) {
	vm.ID = img.ID
	vm.Name = img.Name
	vm.Size = img.Size

	vm.LocalURL = fmt.Sprintf("%s%s", baseURL, img.LocalURL)

	switch imgHost {
	case ImgURLCloudStorage:
		vm.URL = img.CloudStorageURL
	case ImgURLServingURL:
		vm.URL = img.ServingURL
	case ImgURLLocal:
		vm.URL = img.LocalURL
	}
}

func saveImage(ctx context.Context, img io.Reader, author *model.Author) (*model.Image, error) {
	metadata, err := csimg.Save(ctx, img)
	if err != nil {
		e := fmt.Errorf("error in upload: %v", err)
		return nil, e
	}

	imgdata := model.Image{
		ID:              metadata.ID,
		BlobKey:         metadata.BlobKey,
		Filename:        metadata.Filename,
		Size:            metadata.Size,
		ServingURL:      metadata.ServingURL,
		CloudStorageURL: metadata.CloudStorageURL,
		LocalURL:        fmt.Sprintf("/image/%s", metadata.ID),
		Added:           time.Now(),
		Author:          *author,
	}
	err = imgdata.Save(ctx)
	if err != nil {
		e := fmt.Errorf("failed to save image metadata: %v", err)
		return nil, e
	}
	return &imgdata, nil
}

// AdminImageListGET displays the images held on Cloud Storage.
func AdminImageListGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	viewModel := new(imageListViewModel)
	images, err := model.GetAllImage(ctx)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}
	viewModel.addImages(env.Config.BaseDomainName, images, ImgURLLocal)

	v := env.View.New("admin/imagelist")
	v.Data = viewModel
	if err := v.Render(ctx, w, r); err != nil {
		return basehandler.AppErrorDefault(err)
	}

	return nil
}

// AdminImageListJSGET returns a list of images in Cloud Storage in JSON format.
func AdminImageListJSGET(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	viewModel := new(imageListViewModel)
	images, err := model.GetAllImage(ctx)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}
	viewModel.addImages(env.Config.BaseDomainName, images, ImgURLLocal)

	json, _ := json.Marshal(viewModel)
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	w.Write(json)

	return nil
}

// AdminImageListPOST handles a form submission with multiple images and
// returns a success / failure page.
func AdminImageListPOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	err := r.ParseMultipartForm(32 << 20) // 32 MB
	if err != nil || r.MultipartForm.File == nil {
		// TODO: validation error
		return basehandler.AppErrorf("No image specified or invalid data",
			http.StatusBadRequest, err)
	}

	headers := r.MultipartForm.File["files"]
	if len(headers) == 0 {
		// TODO: validation error
		return basehandler.AppErrorf("No image specified or invalid data",
			http.StatusBadRequest, err)
	}

	author, _ := env.User.(*model.Author)

	count := 0
	for i := range headers {
		file, err := headers[i].Open()
		defer file.Close()
		if err != nil {
			return basehandler.AppErrorDefault(err)
		}

		_, err = saveImage(ctx, file, author)
		if err != nil {
			return basehandler.AppErrorDefault(err)
		}

		count++
	}

	flash.AddFlash(w, r, fmt.Sprintf("%d image(s) added", count))
	http.Redirect(w, r, "/admin/image/list", http.StatusFound)
	return nil
}

// AdminImageUploadJSPOST handles a single image upload and returns JSON data.
func AdminImageUploadJSPOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	file, _, err := r.FormFile("img-input")
	if err != nil {
		// TODO: validation
		return basehandler.AppErrorf("No image specified",
			http.StatusBadRequest, err)
	}

	author, _ := env.User.(*model.Author)

	metadata, err := saveImage(ctx, file, author)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	viewModel := new(imageViewModel)
	viewModel.fromEntity(env.Config.BaseDomainName, metadata, ImgURLLocal)

	json, err := json.Marshal(viewModel)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	w.Write(json)

	return nil
}

// AdminImageUpdateJSPOST updates image metadata and responds with JSON.
func AdminImageUpdateJSPOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	viewModel := new(imageViewModel)
	if err := r.ParseForm(); err != nil {
		return basehandler.AppErrorDefault(err)
	}
	if err := env.FormDecoder.Decode(viewModel, r.Form); err != nil {
		return basehandler.AppErrorDefault(err)
	}

	img, err := model.GetImageByID(ctx, viewModel.ID)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	img.Name = viewModel.Name
	err = img.Save(ctx)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	json, err := json.Marshal(viewModel)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	w.Write(json)

	return nil

}

// AdminImageDeletePOST deletes the specified image from Google Cloud Storage.
func AdminImageDeletePOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	id := r.PostFormValue("id")
	if id == "" {
		// TODO: validation
		return basehandler.AppErrorf("Invalid image delete request",
			http.StatusBadRequest, nil)
	}
	img := model.Image{ID: id}
	err := csimg.Delete(ctx, img.ID)
	img.Delete(ctx)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	flash.AddFlash(w, r, fmt.Sprintf("Image with ID %s deleted", id))
	http.Redirect(w, r, "/admin/image/list", http.StatusFound)
	return nil
}

// AdminImageDeleteAllPOST deletes all blog images from Google Cloud Storage.
func AdminImageDeleteAllPOST(ctx context.Context, env appenv.AppEnv, w http.ResponseWriter, r *http.Request) *basehandler.AppError {
	count, err := csimg.DeleteAll(ctx)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	model.DeleteAllImage(ctx)
	if err != nil {
		return basehandler.AppErrorDefault(err)
	}

	flash.AddFlash(w, r, fmt.Sprintf("%d image(s) deleted", count))
	http.Redirect(w, r, "/admin/image/list", http.StatusFound)
	return nil
}
