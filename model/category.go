package model

import (
	"errors"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const categoryKind = "Category"

// Category represents a blog category.
type Category struct {
	Slug  string
	Title string
}

// GetAllCategory returns all categories.
func GetAllCategory(ctx context.Context) ([]Category, error) {
	query := datastore.NewQuery("Category").Ancestor(blogRootKey(ctx))

	var cats []Category
	_, err := query.GetAll(ctx, &cats)
	if err != nil {
		return nil, err
	}
	return cats, nil
}

// Save inserts a new category. If the category already exists it
// does nothing.
func (cat *Category) Save(ctx context.Context) (*datastore.Key, error) {
	if cat.Title == "" || cat.Slug == "" {
		return nil, errors.New("Invalid category")
	}

	q := datastore.NewQuery(categoryKind).Filter("Title=", cat.Title).KeysOnly()
	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		k := datastore.NewIncompleteKey(ctx, categoryKind, blogRootKey(ctx))
		k, err := datastore.Put(ctx, k, cat)
		return k, err
	}

	return nil, nil
}

// DeleteCategory deletes a category by Slug
// TODO: remove category from posts as well
func DeleteCategory(ctx context.Context, slug string) error {
	q := datastore.NewQuery(categoryKind).Filter("Slug=", slug).KeysOnly()
	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		err := datastore.DeleteMulti(ctx, keys)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteAllCategory deletes all categories from the datastore.
func DeleteAllCategory(ctx context.Context) error {
	q := datastore.NewQuery(categoryKind).KeysOnly()
	k, err := q.GetAll(ctx, nil)
	if err != nil {
		return err
	}
	err = datastore.DeleteMulti(ctx, k)
	return err
}

// GetCategoryCount returns the number of categories.
func GetCategoryCount(ctx context.Context) (int, error) {
	q := datastore.NewQuery(categoryKind).KeysOnly()
	k, err := q.GetAll(ctx, nil)
	return len(k), err
}
