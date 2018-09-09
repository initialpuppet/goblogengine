// Package flash uses Gorilla Sessions to provide flash message functionality.
package flash

import (
	"fmt"
	"goblogengine/appenv"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

const defaultSessionName = "session"

// AddFlash adds a flash to the session
func AddFlash(w http.ResponseWriter, r *http.Request, f string) error {
	env := appenv.GetEnv()
	session, err := env.SessionStore.Get(r, defaultSessionName)
	if err != nil {
		ctx := appengine.NewContext(r)
		log.Warningf(ctx, "Invalid session cookie. Using new session: %s", err)
	}
	session.AddFlash(f)
	err = session.Save(r, w)
	if err != nil {
		return fmt.Errorf("flash: error saving session: %s", err.Error())
	}

	return nil
}

// ReadFlashes returns a slice of strings representing flash messages
func ReadFlashes(w http.ResponseWriter, r *http.Request) ([]string, error) {
	env := appenv.GetEnv()
	session, err := env.SessionStore.Get(r, defaultSessionName)
	if err != nil {
		ctx := appengine.NewContext(r)
		log.Warningf(ctx, "Invalid session cookie. Using new session: %s", err)
	}

	f := session.Flashes()
	var msgs []string
	for i := range f {
		var s = ""
		s, ok := f[i].(string)
		if !ok {
			return nil, fmt.Errorf("flash: error opening session: %s",
				err.Error())
		}
		msgs = append(msgs, s)
	}
	session.Save(r, w)

	return msgs, nil
}
