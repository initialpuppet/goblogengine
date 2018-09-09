package model

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const auditKind = "Audit"

// Audit represents an action by an administrator.
type Audit struct {
	Action  string
	Details string
	When    time.Time
	Author  Author
}

// NewAudit returns a new Audit object populated with the specified values
// and the current date and time.
func NewAudit(action string, details string, author Author) Audit {
	return Audit{
		Action:  action,
		Details: details,
		When:    time.Now(),
		Author:  author,
	}
}

// Save adds the Audit to the datastore.
func (a *Audit) Save(ctx context.Context) (*datastore.Key, error) {
	if a.When.IsZero() {
		a.When = time.Now()
	}
	k := datastore.NewIncompleteKey(ctx, auditKind, blogRootKey(ctx))
	k, err := datastore.Put(ctx, k, a)
	return k, err
}

// String returns a string representation of the audit event.
func (a Audit) String() string {
	txt := a.Action
	if a.Details != "" {
		txt += ": " + a.Details
	}
	return txt
}

// GetAuditTail returns the last 100 audit events.
func GetAuditTail(ctx context.Context) ([]Audit, error) {
	q := datastore.NewQuery(auditKind).
		Order("-When").
		Limit(100)
	var evts []Audit
	_, err := q.GetAll(ctx, &evts)
	return evts, err
}
