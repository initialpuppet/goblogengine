package model

import (
	"errors"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const authorKind = "Author"

// Author represents a person who can write blog posts. Each Author is
// associated with a Google user account.
type Author struct {
	Slug            string
	DisplayName     string
	Email           string
	GoogleAccountID string
}

// ErrorNoMatchingAuthor is returned when no Author entry matching the
// supplied email address can be found in the datastore.
var ErrorNoMatchingAuthor = errors.New("model: no author matching supplied email")

// Save adds the Author to the datastore.
// TODO: check for duplicates.
func (a *Author) Save(ctx context.Context) (*datastore.Key, error) {
	k := datastore.NewIncompleteKey(ctx, authorKind, blogRootKey(ctx))
	k, err := datastore.Put(ctx, k, a)
	return k, err
}

// GetAuthorByEmail returns an Author from the datastore matching the supplied
// email address. Returns ErrorNoMatchingAuthor if there is no matching author.
func GetAuthorByEmail(ctx context.Context, email string) (*Author, error) {
	query := datastore.NewQuery(authorKind).Ancestor(blogRootKey(ctx)).Filter("Email=", email)
	var author = new(Author)
	authorList := query.Run(ctx)
	_, err := authorList.Next(author)
	if err == datastore.Done {
		return nil, ErrorNoMatchingAuthor
	}

	return author, err
}

// GetAllAuthor returns a slice representing all Authors in the datastore.
func GetAllAuthor(ctx context.Context) ([]Author, error) {
	query := datastore.NewQuery(authorKind)
	var authors []Author
	_, err := query.GetAll(ctx, &authors)

	return authors, err
}

// DeleteAllAuthor deletes all Author data.
func DeleteAllAuthor(ctx context.Context) error {
	q := datastore.NewQuery(authorKind).KeysOnly()
	k, err := q.GetAll(ctx, nil)
	if err != nil {
		return err
	}
	err = datastore.DeleteMulti(ctx, k)
	return err
}

// GetAuthorCount returns the number of registered authors.
func GetAuthorCount(ctx context.Context) (int, error) {
	q := datastore.NewQuery(authorKind).KeysOnly()
	k, err := q.GetAll(ctx, nil)
	return len(k), err
}
