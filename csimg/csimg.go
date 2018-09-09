// Package csimg saves images in the Google Cloud Storage service, and provides
// a mechanism to retrieve them via the Google CDN.
//
// NOTE: The Go AppEngine development runtime does not appear to support
// ServingURL at present, use the Cloud Storage URL or the raw data locally.
//
// TODO: Check data is JPG format or pass correct extension and content type
// TODO: Tags. Because all libraries need tags.
package csimg

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	uuid "github.com/satori/go.uuid"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/appengine"
	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine/file"
	"google.golang.org/appengine/image"
)

const (
	gcsBaseURL = "https://storage.googleapis.com"
	filePrefix = "csimg"
	fileExt    = ".jpg"
)

// Metadata represents data about a saved image.
type Metadata struct {
	ID              string
	BlobKey         string
	Filename        string
	Size            string
	ServingURL      string
	CloudStorageURL string
}

// Read gets an image from GCS and returns the data as a byte slice.
// TODO: Would returning an io.Reader be useful?
func Read(ctx context.Context, id string) ([]byte, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("csimg: failed to create client: %v", err)
	}
	defer client.Close()

	bucket, err := file.DefaultBucketName(ctx)
	if err != nil {
		return nil, fmt.Errorf("csimg: %v", err)
	}
	bhandle := client.Bucket(bucket)

	filename := getFileName(id)
	r, err := bhandle.Object(filename).NewReader(ctx)
	if err != nil {
		e := fmt.Errorf("csimg: failed to open image %s: %v", filename, err)
		return nil, e
	}

	b, err := ioutil.ReadAll(r)

	return b, err
}

// List retrieves a list of images from GCS and returns their URLs and other
// details.
func List(ctx context.Context) ([]Metadata, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("csimg: failed to create client: %v", err)
	}
	defer client.Close()

	bucket, err := file.DefaultBucketName(ctx)
	if err != nil {
		return nil, fmt.Errorf("csimg: %v", err)
	}
	bhandle := client.Bucket(bucket)

	q := &storage.Query{Prefix: filePrefix}
	iter := bhandle.Objects(ctx, q)

	var metadata []Metadata
	for {
		obj, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			e := fmt.Errorf("csimg: failure iterating: %v", err)
			return nil, e
		}

		csURL := getCloudStorageURL(bucket, obj.Name)
		blobKey, err := getBlobKey(ctx, bucket, obj.Name)
		if err != nil {
			return nil, err
		}
		servingURL, err := getServingURL(ctx, blobKey)
		if err != nil {
			return nil, err
		}

		m := Metadata{
			ID:              "", // we don't have this
			BlobKey:         string(blobKey),
			Filename:        obj.Name,
			Size:            strconv.FormatInt(obj.Size, 10),
			ServingURL:      servingURL,
			CloudStorageURL: csURL,
		}

		metadata = append(metadata, m)
	}

	return metadata, nil
}

// Save accepts an object which satisfies io.Reader and saves it to Google Cloud
// storage with a blog image prefix.
func Save(ctx context.Context, img io.Reader) (*Metadata, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("csimg: failed to create client: %v", err)
	}
	defer client.Close()

	bucket, err := file.DefaultBucketName(ctx)
	if err != nil {
		return nil, fmt.Errorf("csimg: %v", err)
	}
	bhandle := client.Bucket(bucket)

	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("csimg: %v", err)
	}

	id := uuid.String()
	fname := getFileName(id)

	o := bhandle.Object(fname)
	w := o.NewWriter(ctx)
	w.ContentType = "image/jpeg"
	w.CacheControl = "public, max-age=86400"
	w.ACL = []storage.ACLRule{{
		Entity: storage.AllUsers,
		Role:   storage.RoleReader,
	}}

	if _, err := io.Copy(w, img); err != nil {
		return nil, fmt.Errorf("csimg: error copying to bucket: %v", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("csimg: error closing writer: %v", err)
	}

	csURL := getCloudStorageURL(bucket, fname)
	blobKey, err := getBlobKey(ctx, bucket, fname)
	if err != nil {
		return nil, err
	}
	servingURL, err := getServingURL(ctx, blobKey)
	if err != nil {
		return nil, err
	}

	attrs, err := o.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("csimg: error getting attributes: %v", err)
	}

	m := Metadata{
		ID:              id,
		BlobKey:         string(blobKey),
		Filename:        fname,
		Size:            strconv.FormatInt(attrs.Size, 10),
		ServingURL:      servingURL,
		CloudStorageURL: csURL,
	}

	if err != nil {
		return nil, fmt.Errorf("csimg: error saving image metadata")
	}

	return &m, nil
}

// Delete removes an image from Google Cloud Storage and removes the serving URL
// associated with the file's blob key.
func Delete(ctx context.Context, id string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("csimg: failed to create client: %v", err)
	}
	defer client.Close()

	bucket, err := file.DefaultBucketName(ctx)
	if err != nil {
		return fmt.Errorf("csimg: %v", err)
	}
	bhandle := client.Bucket(bucket)
	fname := getFileName(id)

	err = bhandle.Object(fname).Delete(ctx)
	if err != nil {
		return err
	}

	k, err := getBlobKey(ctx, bucket, fname)
	err = image.DeleteServingURL(ctx, k)

	return err
}

// DeleteAll removes all the images from Google Cloud Storage and the
// serving URLs associted with their blob keys.
func DeleteAll(ctx context.Context) (int, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return 0, fmt.Errorf("csimg: failed to create client: %v", err)
	}
	defer client.Close()

	bucket, err := file.DefaultBucketName(ctx)
	if err != nil {
		return 0, fmt.Errorf("csimg: %v", err)
	}
	bhandle := client.Bucket(bucket)

	q := &storage.Query{Prefix: filePrefix}
	iter := bhandle.Objects(ctx, q)

	count := 0
	for {
		objattrs, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("csimg: %v", err)
		}

		blobKey, err := getBlobKey(ctx, bucket, objattrs.Name)
		if err != nil {
			return 0, err
		}
		err = image.DeleteServingURL(ctx, blobKey)
		if err != nil {
			return 0, fmt.Errorf("csimg: unable to delete serving URL")
		}

		obj := bhandle.Object(objattrs.Name)
		obj.Delete(ctx)
		count++
	}

	return count, nil
}

func getBlobKey(ctx context.Context, bucket, name string) (appengine.BlobKey, error) {
	file := fmt.Sprintf("/gs/%s/%s", bucket, name)
	k, err := blobstore.BlobKeyForFile(ctx, file)
	if err != nil {
		e := fmt.Errorf("csimg: failed to get blob key: %v", err)
		return "", e
	}
	return k, nil
}

func getServingURL(ctx context.Context, blobKey appengine.BlobKey) (string, error) {
	opt := &image.ServingURLOptions{
		Secure: true,
		Crop:   false,
	}
	u, err := image.ServingURL(ctx, blobKey, opt)
	if err != nil {
		e := fmt.Errorf("csimg: failed to get serving URL: %v", err)
		return "", e
	}

	url := u.String()
	return url, nil
}

func getCloudStorageURL(bucket, name string) string {
	return fmt.Sprintf("%s/%s/%s", gcsBaseURL, bucket, name)
}

func getFileName(id string) string {
	return filePrefix + id + fileExt
}
