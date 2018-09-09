package model

import (
	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

// blogRootKey returns a *datastore.Key which can be used as the ancestor for
// all entities in the application to ensure strong consistency.
func blogRootKey(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, "Blog", "default", 0, nil)
}

// newIncompleteKeyMulti returns a slice of keys to be used in datastore.*Multi
// operations.
func newIncompleteKeyMulti(ctx context.Context, kind string, parent *datastore.Key, quantity int) []*datastore.Key {
	var keys []*datastore.Key
	for i := 0; i < quantity; i++ {
		k := datastore.NewIncompleteKey(ctx, kind, parent)
		keys = append(keys, k)
	}
	return keys
}
