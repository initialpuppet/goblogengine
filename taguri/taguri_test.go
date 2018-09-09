package taguri_test

import (
	"goblogengine/taguri"
	"testing"
	"time"
)

func TestMakeURIWithHostnameOnly(t *testing.T) {
	testdt, _ := time.Parse("2006-01-02", "2017-09-28")

	need := "tag:blog.example.com,2017-09-28:blog:foo#bar"
	have := taguri.Make(testdt, "blog.example.com", "bar", "blog", "foo")

	if need != have {
		t.Errorf("Have %s, need %s", have, need)
	}
}

func TestMakeURIWithEMailAndFragment(t *testing.T) {
	testdt, _ := time.Parse("2006-01-02", "2017-09-28")

	need := "tag:foo@bar.com,2017-09-28:article"
	have := taguri.Make(testdt, "foo@bar.com", "", "article")

	if need != have {
		t.Errorf("Have %s, need %s", have, need)
	}
}
