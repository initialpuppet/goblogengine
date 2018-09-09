package model

import (
	"errors"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const imageKind = "Image"

// Image represents data about images uploaded to the blog and saved
// by the csimg package.
type Image struct {
	ID              string
	Name            string
	BlobKey         string
	Filename        string
	Size            string
	ServingURL      string
	CloudStorageURL string
	LocalURL        string
	Added           time.Time
	Author          Author
}

// Save saves the Image to the datastore.
func (i *Image) Save(ctx context.Context) error {
	if i.ID == "" {
		return errors.New("model: image ID cannot be empty")
	}
	k := datastore.NewKey(ctx, imageKind, i.ID, 0, blogRootKey(ctx))
	k, err := datastore.Put(ctx, k, i)
	return err
}

// Delete deletes the Image from the datastore.
func (i *Image) Delete(ctx context.Context) error {
	q := datastore.NewQuery(imageKind).Filter("ID=", i.ID).KeysOnly()
	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		return err
	}
	err = datastore.DeleteMulti(ctx, keys)
	return err
}

// GetImageByID returns image metadata for the supplied ID.
func GetImageByID(ctx context.Context, id string) (*Image, error) {
	if id == "" {
		return nil, errors.New("model: no image ID provided")
	}

	q := datastore.NewQuery(imageKind).Filter("ID=", id).Limit(1)
	var imgs []Image
	_, err := q.GetAll(ctx, &imgs)
	if err != nil {
		return nil, err
	}

	if len(imgs) == 0 {
		return nil, errors.New("model: no image matching supplied ID")
	}

	return &imgs[0], nil

}

// GetAllImage returns metadata for all images in the datastore.
func GetAllImage(ctx context.Context) ([]Image, error) {
	q := datastore.NewQuery(imageKind).Ancestor(blogRootKey(ctx)).Order("-Added")
	var d []Image
	_, err := q.GetAll(ctx, &d)
	return d, err
}

// DeleteAllImage deletes all image metadata from the datastore.
func DeleteAllImage(ctx context.Context) error {
	q := datastore.NewQuery(imageKind).KeysOnly()
	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		return err
	}
	err = datastore.DeleteMulti(ctx, keys)
	return err
}

// GetImageCount returns the number of images.
func GetImageCount(ctx context.Context) (int, error) {
	q := datastore.NewQuery(imageKind).KeysOnly()
	k, err := q.GetAll(ctx, nil)
	return len(k), err
}
