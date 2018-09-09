package slug_test

import (
	"goblogengine/slug"
	"testing"
)

func TestMakeSlug(t *testing.T) {
	input := "blah blah"
	output := slug.Make(input)

	if output != "blah-blah" {
		t.Errorf("String was not URLized, got %s, should be %s", output, input)
	}
}

func TestMakeSlugMulti(t *testing.T) {
	data := []struct {
		input  string
		output string
	}{
		{"blah blah", "blah-blah"},
		{"Blah blah", "blah-blah"},
		{"Blahblah", "blahblah"},
	}

	for _, vals := range data {
		output := slug.Make(vals.input)
		if output != vals.output {
			t.Errorf("Got %s, should be %s", output, vals.output)
		}
	}
}
