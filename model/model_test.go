package model

import "testing"
import "google.golang.org/appengine/aetest"

func TestBlogRootKey(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	defer done()
	if err != nil {
		t.Fatalf("Unable to get AppEngine context for testing. Error: %s", err)
	}

	key := blogRootKey(ctx)

	if key.Parent() != nil {
		t.Error("Root key is not root")
	}
}
