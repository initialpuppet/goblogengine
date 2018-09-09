// Package appenv defines data structures for holding environment-wide information
// used by goblogengine.
package appenv

import (
	"sync"

	"github.com/gorilla/sessions"

	"goblogengine/envae"
	"goblogengine/view"

	"github.com/gorilla/schema"
	"google.golang.org/appengine"
)

var env AppEnv

var envMutex sync.RWMutex

// AppEnv defines a structure to store environment-wide information used by the
// application.
type AppEnv struct {
	Config       Config
	View         view.Info
	FormDecoder  *schema.Decoder
	SessionStore *sessions.CookieStore
	User         interface{}

	HostEnv int
}

// Config defines a structure to store global application settings.
type Config struct {
	BlogName       string `envae:"blog_name"`
	BaseDomainName string `envae:"base_domain_name"`

	PostsPerPage      int `envae:"posts_per_page"`
	FeedSize          int `envae:"feed_size"`
	ExcerptCharLength int `envae:"excerpt_char_length"`

	DateFormatForEditing string `envae:"date_format_for_editing"`
	DateFormatShort      string `envae:"date_format_short"`
	DateFormatFull       string `envae:"date_format_full"`

	SessionStoreKey string `envae:"session_store_key"`

	Template view.Template
}

// AppEnv.HostEnv can be either development or live.
const (
	EnvDev  = iota
	EnvLive = iota
)

// Init creates the application settings struct.
func Init() error {
	var e AppEnv

	if appengine.IsDevAppServer() {
		e.HostEnv = EnvDev
	} else {
		e.HostEnv = EnvLive
	}

	// Read the config from AppEngine settings and configure the view engine
	// with default templates
	err := envae.Populate(&e)
	if err != nil {
		return err
	}
	e.View.SetTemplates(e.Config.Template.Root, e.Config.Template.Children)
	e.View.SetDateFormat(e.Config.DateFormatFull)

	e.FormDecoder = schema.NewDecoder()
	e.FormDecoder.IgnoreUnknownKeys(true)
	e.SessionStore = sessions.NewCookieStore([]byte(e.Config.SessionStoreKey))

	setEnv(&e)

	return nil
}

// setEnv safely overwrites the environment information with new data.
func setEnv(e *AppEnv) {
	envMutex.Lock()
	defer envMutex.Unlock()
	env = *e
}

// GetEnv safely returns a copy of the environment information.
func GetEnv() AppEnv {
	envMutex.RLock()
	defer envMutex.RUnlock()
	return env
}
