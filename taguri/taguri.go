// Package taguri generates a Tag URI meeting the specification in RFC4151.
// It does not generate the full range of possible formats, or validate its
// inputs.
//
// https://www.ietf.org/rfc/rfc4151.txt
package taguri

import (
	"fmt"
	"strings"
	"time"
)

// Make accepts the components of the tag and returns the complete tag.
func Make(dt time.Time, authority, frag string, specifics ...string) string {
	dateText := dt.Format("2006-01-02")
	authorityText := strings.ToLower(authority)

	specificsText := ":" + specifics[0]
	for _, s := range specifics[1:] {
		specificsText += ":" + s
	}

	tag := fmt.Sprintf("tag:%s,%s%s", authorityText, dateText, specificsText)

	if len(frag) > 0 {
		tag += "#" + frag
	}

	return tag
}
