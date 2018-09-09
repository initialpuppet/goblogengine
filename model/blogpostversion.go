package model

import (
	"errors"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const blogPostVersionKind = "BlogPostVersion"

// ErrorPostSlugAlreadyExists is returned when the chosen URL slug is already
// associated with another post in the datastore.
var ErrorPostSlugAlreadyExists = errors.New("model: url slug already in use")

// BlogPostVersion represents a version of a blog post. A single post can have
// many versions, but only one version can be published at a point in time.
type BlogPostVersion struct {
	PostID         string
	Slug           string
	Title          string
	Categories     []Category
	BannerImageURL string
	BodyMarkdown   string `datastore:",noindex"`
	DatePublished  time.Time
	DateCreated    time.Time
	Published      bool
	Author         Author
	Version        int
}

// GetBlogPostBySlug returns a published BlogPostVersion matching the supplied
// URL slug.
func GetBlogPostBySlug(ctx context.Context, slug string) (*BlogPostVersion, error) {
	query := datastore.NewQuery(blogPostVersionKind).
		Filter("Slug=", slug).
		Filter("Published=", true)

	var post = new(BlogPostVersion)
	postlist := query.Run(ctx)
	_, err := postlist.Next(post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

// GetBlogPostVersion returns a BlogPostVersion matching the supplied URL slug
// and version number.
func GetBlogPostVersion(ctx context.Context, slug string, version int) (*BlogPostVersion, error) {
	query := datastore.NewQuery(blogPostVersionKind).
		Ancestor(blogRootKey(ctx)).
		Filter("Slug=", slug).
		Filter("Version=", version)

	var post = new(BlogPostVersion)
	postList := query.Run(ctx)
	_, err := postList.Next(post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

// Save adds the BlogPostVersion to the datastore.
//
// If new is true, an error is returned if there are existing posts with the
// same tag or slug, otherwise a new version is added and the version number
// is incremented.
//
// If ver.Published is true, the inserted version is published and all
// other versions of the post are un-published.
func (ver *BlogPostVersion) Save(ctx context.Context, new bool) (*datastore.Key, error) {
	for i := range ver.Categories {
		_, err := ver.Categories[i].Save(ctx)
		if err != nil {
			log.Errorf(ctx, "model: failed to add category: %v", err)
		}
	}

	var newVersionKey *datastore.Key
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		if new {
			q := datastore.NewQuery(blogPostVersionKind).
				Ancestor(blogRootKey(ctx)).
				Filter("Slug=", ver.Slug).
				KeysOnly()

			k, err := q.GetAll(ctx, nil)
			if err != nil {
				return err
			}
			if len(k) > 0 {
				return ErrorPostSlugAlreadyExists
			}
		} else {
			if ver.Published == true {
				var versions []BlogPostVersion
				query := datastore.NewQuery(blogPostVersionKind).
					Ancestor(blogRootKey(ctx)).
					Filter("Slug=", ver.Slug).
					Filter("Published=", true)

				keys, err := query.GetAll(ctx, &versions)
				if err != nil {
					return err
				}

				for i := range versions {
					versions[i].Published = false
				}

				_, err = datastore.PutMulti(ctx, keys, versions)
				if err != nil {
					return err
				}
			}

			var versions []BlogPostVersion
			query := datastore.NewQuery(blogPostVersionKind).
				Ancestor(blogRootKey(ctx)).
				Filter("Slug=", ver.Slug).
				Order("-Version").
				Limit(1)

			_, err := query.GetAll(ctx, &versions)
			if err != nil {
				return err
			}

			if len(versions) > 0 {
				ver.Version = versions[0].Version + 1
			}

		}

		// Put the new version
		newVersionKey = datastore.NewIncompleteKey(ctx, blogPostVersionKind, blogRootKey(ctx))
		_, err := datastore.Put(ctx, newVersionKey, ver)
		if err != nil {
			return err
		}

		return nil

	}, nil)

	return newVersionKey, err
}

// GetBlogPostLimit returns a slice of BlogPostVersion from offset to limit,
// ordered by most recent first. If limit is < 1 offset is ignored and the
// function returns all available posts.
func GetBlogPostLimit(ctx context.Context, offset int, limit int) ([]BlogPostVersion, error) {
	q := datastore.NewQuery(blogPostVersionKind).
		Filter("Published=", true).
		Order("-DatePublished")

	if limit > 0 {
		q = q.Limit(limit)
		q = q.Offset(offset)
	}

	var posts []BlogPostVersion
	_, err := q.GetAll(ctx, &posts)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// GetAllBlogPost returns a slice of BlogPostVersion representing all distinct
// posts in the datastore. If a version of a post has been published, that
// version is returned, otherwise the most recent draft is returned.
func GetAllBlogPost(ctx context.Context) ([]BlogPostVersion, error) {
	query := datastore.NewQuery(blogPostVersionKind).
		Ancestor(blogRootKey(ctx)).
		Project("PostID").
		Distinct().
		Order("PostID").
		Order("-Published").
		Order("-DateCreated")

	var x []BlogPostVersion
	keys, err := query.GetAll(ctx, &x)
	if err != nil {
		return nil, err
	}

	posts := make([]BlogPostVersion, len(keys))
	err = datastore.GetMulti(ctx, keys, posts)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// GetBlogPostVersionBySlug returns a slice of BlogPostVersion representing
// all versions of a given post.
func GetBlogPostVersionBySlug(ctx context.Context, slug string) ([]BlogPostVersion, error) {
	query := datastore.NewQuery(blogPostVersionKind).
		Ancestor(blogRootKey(ctx)).
		Filter("Slug=", slug)

	var posts []BlogPostVersion
	_, err := query.GetAll(ctx, &posts)

	return posts, err
}

// UnpublishBlogPost unpublishes the currently published version of a blog post.
func UnpublishBlogPost(ctx context.Context, id string) error {
	q := datastore.NewQuery(blogPostVersionKind).
		Filter("PostID=", id).
		Filter("Published=", true)

	var posts []BlogPostVersion
	k, err := q.GetAll(ctx, &posts)
	if err != nil {
		return err
	}

	for i := range posts {
		posts[i].Published = false
	}

	_, err = datastore.PutMulti(ctx, k, posts)
	return err
}

// PublishBlogPostVersion publishes a given BlogPostVersion.
func PublishBlogPostVersion(ctx context.Context, id string, version int) error {
	q := datastore.NewQuery(blogPostVersionKind).
		Filter("PostID=", id)

	var posts []BlogPostVersion
	k, err := q.GetAll(ctx, &posts)
	if err != nil {
		return err
	}

	for i := range posts {
		if posts[i].Version == version {
			posts[i].Published = true
		} else {
			posts[i].Published = false
		}
	}

	_, err = datastore.PutMulti(ctx, k, posts)
	return err
}

// DeleteBlogPost deletes all versions of a blog post.
func DeleteBlogPost(ctx context.Context, id string) error {
	q := datastore.NewQuery(blogPostVersionKind).
		Filter("PostID=", id).
		KeysOnly()

	k, err := q.GetAll(ctx, nil)
	if err != nil {
		return err
	}

	err = datastore.DeleteMulti(ctx, k)
	return err
}

// DeleteAllBlogPostVersion deletes all blog posts.
func DeleteAllBlogPostVersion(ctx context.Context) error {
	q := datastore.NewQuery(blogPostVersionKind).KeysOnly()
	k, err := q.GetAll(ctx, nil)
	datastore.DeleteMulti(ctx, k)
	return err
}

// GetBlogPostVersionCount returns the total number of all BlogPostVersion.
func GetBlogPostVersionCount(ctx context.Context) (int, error) {
	q := datastore.NewQuery(blogPostVersionKind).KeysOnly()
	k, err := q.GetAll(ctx, nil)
	return len(k), err
}

// GetBlogPostCount returns the number of published blog posts.
func GetBlogPostCount(ctx context.Context) (int, error) {
	q := datastore.NewQuery(blogPostVersionKind).
		Filter("Published=", true).
		KeysOnly()
	k, err := q.GetAll(ctx, nil)
	return len(k), err
}

// GetBlogPostDraftCount returns the number of posts which are in a draft state.
func GetBlogPostDraftCount(ctx context.Context) (int, error) {
	query := datastore.NewQuery(blogPostVersionKind).
		Project("PostID").
		Distinct().
		Order("PostID").
		Order("-Published").
		Order("-DateCreated").
		Filter("Published=", false)

	var x []BlogPostVersion
	keys, err := query.GetAll(ctx, &x)

	return len(keys), err
}
